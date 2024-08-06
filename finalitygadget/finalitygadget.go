package finalitygadget

import (
	"context"
	"encoding/hex"
	"fmt"
	"math"
	"math/big"
	"strings"
	"sync"
	"time"

	bbnClient "github.com/babylonlabs-io/babylon/client/client"
	bbncfg "github.com/babylonlabs-io/babylon/client/config"
	"github.com/babylonlabs-io/finality-gadget/bbnclient"
	"github.com/babylonlabs-io/finality-gadget/btcclient"
	"github.com/babylonlabs-io/finality-gadget/config"
	"github.com/babylonlabs-io/finality-gadget/cwclient"
	"github.com/babylonlabs-io/finality-gadget/db"
	"github.com/babylonlabs-io/finality-gadget/ethl2client"
	"github.com/babylonlabs-io/finality-gadget/testutil/mocks"
	"github.com/babylonlabs-io/finality-gadget/types"
	"github.com/ethereum/go-ethereum/common"
	ethrpc "github.com/ethereum/go-ethereum/rpc"
	"go.uber.org/zap"
)

var _ IFinalityGadget = &FinalityGadget{}

type FinalityGadget struct {
	btcClient btcclient.IBitcoinClient
	bbnClient bbnclient.IBabylonClient
	cwClient  cwclient.ICosmWasmClient
	l2Client  ethl2client.IEthL2Client

	db     db.IDatabaseHandler
	logger *zap.Logger
	mutex  sync.Mutex

	pollInterval time.Duration
	currHeight   uint64
}

//////////////////////////////
// CONSTRUCTOR
//////////////////////////////

func NewFinalityGadget(cfg *config.Config, db db.IDatabaseHandler, logger *zap.Logger) (*FinalityGadget, error) {
	// Create babylon client
	bbnConfig := bbncfg.DefaultBabylonConfig()
	bbnConfig.RPCAddr = cfg.BBNRPCAddress
	bbnConfig.ChainID = cfg.BBNChainID
	babylonClient, err := bbnClient.New(
		&bbnConfig,
		logger,
	)
	bbnClient := bbnclient.NewBabylonClient(babylonClient.QueryClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create Babylon client: %w", err)
	}

	// Create bitcoin client
	btcConfig := btcclient.DefaultBTCConfig()
	btcConfig.RPCHost = cfg.BitcoinRPCHost
	var btcClient btcclient.IBitcoinClient
	switch cfg.BBNChainID {
	// TODO: once we set up our own local BTC devnet, we don't need to use this mock BTC client
	case config.BabylonLocalnetChainID:
		btcClient, err = mocks.NewMockBitcoinClient(btcConfig, logger)
	default:
		btcClient, err = btcclient.NewBitcoinClient(btcConfig, logger)
	}
	if err != nil {
		return nil, err
	}

	// Create cosmwasm client
	cwClient := cwclient.NewCosmWasmClient(babylonClient.QueryClient.RPCClient, cfg.FGContractAddress)

	// Create L2 client
	l2Client, err := ethl2client.NewEthL2Client(cfg.L2RPCHost)
	if err != nil {
		return nil, err
	}

	// Create finality gadget
	return &FinalityGadget{
		btcClient:    btcClient,
		bbnClient:    bbnClient,
		cwClient:     cwClient,
		l2Client:     l2Client,
		db:           db,
		pollInterval: cfg.PollInterval,
		logger:       logger,
	}, nil
}

//////////////////////////////
// METHODS
//////////////////////////////

/* QueryIsBlockBabylonFinalized checks if the given L2 block is finalized by the Babylon finality gadget
 *
 * - if the finality gadget is not enabled, always return true
 * - else, check if the given L2 block is finalized
 * - return true if finalized, false if not finalized, and error if any
 *
 * - to check if the block is finalized, we need to:
 *   - get the consumer chain id
 *   - get all the FPs pubkey for the consumer chain
 *   - convert the L2 block timestamp to BTC height
 *   - get all FPs voting power at this BTC height
 *   - calculate total voting power
 *   - get all FPs that voted this L2 block with the same height and hash
 *   - calculate voted voting power
 *   - check if the voted voting power is more than 2/3 of the total voting power
 */
func (fg *FinalityGadget) QueryIsBlockBabylonFinalized(block *types.Block) (bool, error) {
	// check if the finality gadget is enabled
	// if not, always return true to pass through op derivation pipeline
	isEnabled, err := fg.cwClient.QueryIsEnabled()
	if err != nil {
		return false, err
	}
	if !isEnabled {
		return true, nil
	}

	// trim prefix 0x for the L2 block hash
	block.BlockHash = strings.TrimPrefix(block.BlockHash, "0x")

	// get all FPs pubkey for the consumer chain
	allFpPks, err := fg.queryAllFpBtcPubKeys()
	if err != nil {
		return false, err
	}

	// convert the L2 timestamp to BTC height
	btcblockHeight, err := fg.btcClient.GetBlockHeightByTimestamp(block.BlockTimestamp)
	if err != nil {
		return false, err
	}

	// check whether the btc staking is actived
	earliestDelHeight, err := fg.bbnClient.QueryEarliestActiveDelBtcHeight(allFpPks)
	if err != nil {
		return false, err
	}
	if btcblockHeight < earliestDelHeight {
		return false, types.ErrBtcStakingNotActivated
	}

	// get all FPs voting power at this BTC height
	allFpPower, err := fg.bbnClient.QueryMultiFpPower(allFpPks, btcblockHeight)
	if err != nil {
		return false, err
	}

	// calculate total voting power
	var totalPower uint64 = 0
	for _, power := range allFpPower {
		totalPower += power
	}

	// no FP has voting power for the consumer chain
	if totalPower == 0 {
		return false, types.ErrNoFpHasVotingPower
	}

	// get all FPs that voted this (L2 block height, L2 block hash) combination
	votedFpPks, err := fg.cwClient.QueryListOfVotedFinalityProviders(block)
	if err != nil {
		return false, err
	}
	if votedFpPks == nil {
		return false, nil
	}
	// calculate voted voting power
	var votedPower uint64 = 0
	for _, key := range votedFpPks {
		if power, exists := allFpPower[key]; exists {
			votedPower += power
		}
	}

	// quorom < 2/3
	if votedPower*3 < totalPower*2 {
		return false, nil
	}
	return true, nil
}

/* QueryBlockRangeBabylonFinalized searches for a row of consecutive finalized blocks in the block range, and returns
 * the last finalized block height
 *
 * Example: if give block range 1-10, and block 1-5 are finalized, and block 6-10 are not finalized, then return 5
 *
 * - if no block in the range is finalized, return (nil, nil)
 * - else, return the height of the last found consecutive finalized block, return error if any
 *
 * Example: if give block range 1-10, and block 1-5 are finalized, and when querying block 6 we meet an error, then
 * return (5, error)
 *
 * Note: caller needs to make sure the given queryBlocks are consecutive (we don't check hashes inside this method)
 * and start from low to high
 */
func (fg *FinalityGadget) QueryBlockRangeBabylonFinalized(
	queryBlocks []*types.Block,
) (*uint64, error) {
	if len(queryBlocks) == 0 {
		return nil, fmt.Errorf("no blocks provided")
	}
	// check if the blocks are consecutive
	for i := 1; i < len(queryBlocks); i++ {
		if queryBlocks[i].BlockHeight != queryBlocks[i-1].BlockHeight+1 {
			return nil, fmt.Errorf("blocks are not consecutive")
		}
	}
	var finalizedBlockHeight *uint64
	for _, block := range queryBlocks {
		isFinalized, err := fg.QueryIsBlockBabylonFinalized(block)
		if err != nil {
			return finalizedBlockHeight, err
		}
		if isFinalized {
			finalizedBlockHeight = &block.BlockHeight
		} else {
			break
		}
	}
	return finalizedBlockHeight, nil
}

/* QueryBtcStakingActivatedTimestamp returns the timestamp when the BTC staking is activated
 *
 * - We will check for k deep and covenant quorum to mark a delegation as active
 * - So technically, activated time needs to be max of the following
 *	 - timestamp of Babylon block that BTC delegation receives covenant quorum
 *	 - timestamp of BTC block that BTC delegation's staking tx becomes k-deep
 * - But we don't have a Babylon API to find the earliest Babylon block where BTC delegation gets covenant quorum.
 *   and it's probably not a good idea to add this to Babylon, as this creates more coupling between Babylon and
 *   consumer. So waiting for covenant quorum can be implemented in a clean way only with Babylon side support.
 *   this will be considered as future work
 * - For now, we will use the k-deep BTC block timestamp for the activation timestamp
 * - The time diff issue is not burning. The time diff between pending and active only matters if FP equivocates
 *   during that time period
 *
 * returns math.MaxUint64, ErrBtcStakingNotActivated if the BTC staking is not activated
 */
func (fg *FinalityGadget) QueryBtcStakingActivatedTimestamp() (uint64, error) {
	allFpPks, err := fg.queryAllFpBtcPubKeys()
	if err != nil {
		return math.MaxUint64, err
	}

	// check whether the btc staking is actived
	earliestDelHeight, err := fg.bbnClient.QueryEarliestActiveDelBtcHeight(allFpPks)
	if err != nil {
		return math.MaxUint64, err
	}

	// not activated yet
	if earliestDelHeight == math.MaxUint64 {
		return math.MaxUint64, types.ErrBtcStakingNotActivated
	}

	// get the timestamp of the BTC height
	btcBlockTimestamp, err := fg.btcClient.GetBlockTimestampByHeight(earliestDelHeight)
	if err != nil {
		return math.MaxUint64, err
	}
	return btcBlockTimestamp, nil
}

func (fg *FinalityGadget) GetBlockByHeight(height uint64) (*types.Block, error) {
	return fg.db.GetBlockByHeight(height)
}

func (fg *FinalityGadget) GetBlockByHash(hash string) (*types.Block, error) {
	return fg.db.GetBlockByHash(normalizeBlockHash(hash))
}

func (fg *FinalityGadget) GetBlockStatusByHeight(height uint64) (bool, error) {
	return fg.db.GetBlockStatusByHeight(height)
}

func (fg *FinalityGadget) GetBlockStatusByHash(hash string) (bool, error) {
	return fg.db.GetBlockStatusByHash(normalizeBlockHash(hash))
}

func (fg *FinalityGadget) GetLatestBlock() (*types.Block, error) {
	return fg.db.GetLatestBlock()
}

// This function process blocks indefinitely, starting from the last finalized block.
func (fg *FinalityGadget) ProcessBlocks(ctx context.Context) error {
	// Start service at last finalized block
	err := fg.startService()
	if err != nil {
		return fmt.Errorf("%v", err)
	}

	// Start polling for new blocks at set interval
	ticker := time.NewTicker(fg.pollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			block, err := fg.getBlockByNumber(int64(fg.currHeight + 1))
			if err != nil {
				fg.logger.Fatal("error getting new block", zap.Error(err))
				continue
			}
			go func() {
				fg.handleBlock(block)
			}()
		}
	}
}

func (fg *FinalityGadget) InsertBlock(block *types.Block) error {
	// Lock mutex
	fg.mutex.Lock()
	// Store block in DB
	err := fg.db.InsertBlock(&types.Block{
		BlockHeight:    block.BlockHeight,
		BlockHash:      normalizeBlockHash(block.BlockHash),
		BlockTimestamp: block.BlockTimestamp,
	})
	if err != nil {
		return err
	}
	// Set last finalized block in memory
	fg.currHeight = block.BlockHeight
	// Unlock mutex
	fg.mutex.Unlock()

	return nil
}

func (fg *FinalityGadget) DeleteDB() error {
	return fg.db.DeleteDB()
}

func (fg *FinalityGadget) Close() {
	fg.l2Client.Close()
}

//////////////////////////////
// INTERNAL
//////////////////////////////

func (fg *FinalityGadget) queryAllFpBtcPubKeys() ([]string, error) {
	// get the consumer chain id
	consumerId, err := fg.cwClient.QueryConsumerId()
	if err != nil {
		return nil, err
	}

	// get all the FPs pubkey for the consumer chain
	allFpPks, err := fg.bbnClient.QueryAllFpBtcPubKeys(consumerId)
	if err != nil {
		return nil, err
	}
	return allFpPks, nil
}

// Get last btc finalized block
func (fg *FinalityGadget) getLatestFinalizedBlock() (*types.Block, error) {
	return fg.getBlockByNumber(ethrpc.FinalizedBlockNumber.Int64())
}

// Get block by number
func (fg *FinalityGadget) getBlockByNumber(blockNumber int64) (*types.Block, error) {
	header, err := fg.l2Client.HeaderByNumber(context.Background(), big.NewInt(blockNumber))
	if err != nil {
		return nil, err
	}
	return &types.Block{
		BlockHeight:    header.Number.Uint64(),
		BlockHash:      hex.EncodeToString(header.Hash().Bytes()),
		BlockTimestamp: header.Time,
	}, nil
}

// Start service at last finalized block
func (fg *FinalityGadget) startService() error {
	// Query L2 node for last finalized block
	block, err := fg.getLatestFinalizedBlock()
	if err != nil {
		return fmt.Errorf("error getting last finalized block: %v", err)
	}

	// Query local DB for last block processed
	localBlock, err := fg.db.GetLatestBlock()
	if err != nil {
		return fmt.Errorf("error getting latest block from db: %v", err)
	}

	// if local chain tip is ahead of node, start service at latest block
	if localBlock != nil && localBlock.BlockHeight >= block.BlockHeight {
		block = &types.Block{
			BlockHeight:    localBlock.BlockHeight,
			BlockHash:      localBlock.BlockHash,
			BlockTimestamp: localBlock.BlockTimestamp,
		}
	} else {
		// throw if block height is 0
		if block.BlockHeight == 0 {
			return fmt.Errorf("block height 0")
		}

		// Check the block is finalized using sdk client
		isFinal, err := fg.QueryIsBlockBabylonFinalized(block)
		// If not finalized, throw error
		if !isFinal {
			return fmt.Errorf("block %d should be finalized according to client but is not", block.BlockHeight)
		}
		if err != nil {
			return fmt.Errorf("error checking block %d: %v", block.BlockHeight, err)
		}
		// If finalised, store the block in DB and set the last finalized block
		err = fg.InsertBlock(block)
		if err != nil {
			return fmt.Errorf("error storing block %d: %v", block.BlockHeight, err)
		}
	}

	// Start service at block height
	fg.logger.Info("Starting finality gadget at block", zap.Uint64("block_height", block.BlockHeight))

	// Set the curr finalized block in memory
	fg.currHeight = block.BlockHeight

	return nil
}

func (fg *FinalityGadget) handleBlock(block *types.Block) {
	// while block is not finalized, recheck if block is finalized every `retryInterval` seconds
	// if finalized, store the block in DB and set the last finalized block
	for {
		// Check if block is finalized
		isFinal, err := fg.QueryIsBlockBabylonFinalized(block)
		if err != nil {
			fg.logger.Fatal("Error checking block", zap.Uint64("block_height", block.BlockHeight), zap.Error(err))
			return
		}
		if isFinal {
			err = fg.InsertBlock(block)
			if err != nil {
				fg.logger.Fatal("Error storing block", zap.Uint64("block_height", block.BlockHeight), zap.Error(err))
			}
			return
		}

		// Sleep for `PollInterval` seconds
		time.Sleep(fg.pollInterval * time.Second)
	}
}

func normalizeBlockHash(hash string) string {
	return common.HexToHash(hash).Hex()
}
