package finalitygadget

import (
	"context"
	"encoding/hex"
	"errors"
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
	"github.com/babylonlabs-io/finality-gadget/types"
	"github.com/ethereum/go-ethereum/common"
	ethrpc "github.com/ethereum/go-ethereum/rpc"
	"go.uber.org/zap"
)

var _ IFinalityGadget = &FinalityGadget{}

type FinalityGadget struct {
	btcClient IBitcoinClient
	bbnClient IBabylonClient
	cwClient  ICosmWasmClient
	l2Client  IEthL2Client

	db     db.IDatabaseHandler
	logger *zap.Logger
	mutex  sync.Mutex

	pollInterval        time.Duration
	lastProcessedHeight uint64
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
	if cfg.BitcoinRPCUser != "" && cfg.BitcoinRPCPass != "" {
		btcConfig.RPCUser = cfg.BitcoinRPCUser
		btcConfig.RPCPass = cfg.BitcoinRPCPass
	}
	if cfg.BitcoinDisableTLS {
		btcConfig.DisableTLS = true
	}
	btcClient, err := btcclient.NewBitcoinClient(btcConfig, logger)
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

	lastProcessedHeight := uint64(0)
	latestBlock, err := db.QueryLatestFinalizedBlock()
	if err != nil {
		return nil, err
	}
	if latestBlock != nil {
		lastProcessedHeight = latestBlock.BlockHeight
	}

	// Create finality gadget
	return &FinalityGadget{
		btcClient:           btcClient,
		bbnClient:           bbnClient,
		cwClient:            cwClient,
		l2Client:            l2Client,
		db:                  db,
		pollInterval:        cfg.PollInterval,
		lastProcessedHeight: lastProcessedHeight,
		logger:              logger,
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
	// handle case where no block is finalized
	if finalizedBlockHeight == nil {
		return nil, nil
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
	fg.logger.Debug("All consumer FP public keys", zap.Strings("allFpPks", allFpPks))
	// check whether the btc staking is actived
	earliestDelHeight, err := fg.bbnClient.QueryEarliestActiveDelBtcHeight(allFpPks)
	// not activated yet
	if earliestDelHeight == math.MaxUint64 {
		return math.MaxUint64, types.ErrBtcStakingNotActivated
	}
	if err != nil {
		return math.MaxUint64, err
	}
	fg.logger.Debug("Earliest active delegation height", zap.Uint64("height", earliestDelHeight))

	// get the timestamp of the BTC height
	btcBlockTimestamp, err := fg.btcClient.GetBlockTimestampByHeight(earliestDelHeight)
	if err != nil {
		return math.MaxUint64, err
	}
	fg.logger.Debug("BTC staking activated at", zap.Uint64("timestamp", btcBlockTimestamp))
	return btcBlockTimestamp, nil
}

func (fg *FinalityGadget) GetBlockByHeight(height uint64) (*types.Block, error) {
	return fg.db.GetBlockByHeight(height)
}

func (fg *FinalityGadget) GetBlockByHash(hash string) (*types.Block, error) {
	return fg.db.GetBlockByHash(normalizeBlockHash(hash))
}

func (fg *FinalityGadget) QueryIsBlockFinalizedByHeight(height uint64) (bool, error) {
	return fg.db.QueryIsBlockFinalizedByHeight(height)
}

func (fg *FinalityGadget) QueryIsBlockFinalizedByHash(hash string) (bool, error) {
	return fg.db.QueryIsBlockFinalizedByHash(normalizeBlockHash(hash))
}

func (fg *FinalityGadget) QueryLatestFinalizedBlock() (*types.Block, error) {
	return fg.db.QueryLatestFinalizedBlock()
}

// This function process blocks indefinitely, starting from the last finalized block.
func (fg *FinalityGadget) ProcessBlocks(ctx context.Context) error {
	// Start polling for new blocks at set interval
	ticker := time.NewTicker(fg.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			latestFinalizedBlock, err := fg.l2Client.HeaderByNumber(ctx, big.NewInt(ethrpc.FinalizedBlockNumber.Int64()))
			if err != nil {
				// log the error and continue
				fg.logger.Error("Error fetching latest finalized L2 block", zap.Error(err))
				continue
			}

			latestFinalizedHeight := latestFinalizedBlock.Number.Uint64()
			latestFinalizedBlockTime := latestFinalizedBlock.Time

			// get the BTC staking activation timestamp
			btcStakingActivatedTimestamp, err := fg.QueryBtcStakingActivatedTimestamp()
			if err != nil {
				if errors.Is(err, types.ErrBtcStakingNotActivated) {
					fg.logger.Info("BTC staking not yet activated, waiting...")
					continue
				}
				return fmt.Errorf("error querying BTC staking activation timestamp: %w", err)
			}

			// only process blocks after the btc staking is activated
			if latestFinalizedBlockTime < btcStakingActivatedTimestamp {
				fg.logger.Info("Skipping block before BTC staking activation", zap.Uint64("block_height", latestFinalizedHeight))
				fg.lastProcessedHeight = latestFinalizedHeight
				continue
			}

			// if the latest finalized block is greater than the last processed block, process it
			if latestFinalizedHeight > fg.lastProcessedHeight {
				fg.logger.Info("Processing block", zap.Uint64("block_height", latestFinalizedHeight))
				if err := fg.handleBlock(ctx, latestFinalizedHeight); err != nil {
					return fmt.Errorf("error processing block %d: %w", latestFinalizedHeight, err)
				}
			}
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

	// Unlock mutex
	fg.mutex.Unlock()

	return nil
}

func (fg *FinalityGadget) Close() {
	fg.l2Client.Close()
	fg.db.Close()
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

// Get block by number
func (fg *FinalityGadget) queryBlockByHeight(blockNumber int64) (*types.Block, error) {
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

func (fg *FinalityGadget) handleBlock(ctx context.Context, latestFinalizedHeight uint64) error {
	for height := fg.lastProcessedHeight + 1; height <= latestFinalizedHeight; height++ {
		select {
		case <-ctx.Done():
			return nil
		default:
			block, err := fg.queryBlockByHeight(int64(height))
			if err != nil {
				return fmt.Errorf("error getting block at height %d: %w", height, err)
			}

			// Check the block is babylon finalized using sdk client
			isFinal, err := fg.QueryIsBlockBabylonFinalized(block)
			if err != nil && !errors.Is(err, types.ErrBtcStakingNotActivated) {
				return fmt.Errorf("error checking block %d: %v", block.BlockHeight, err)
			}
			// If not finalized, throw error
			if !isFinal {
				return fmt.Errorf("block %d should be finalized according to client but is not", block.BlockHeight)
			}

			// If finalised, store the block in DB
			err = fg.InsertBlock(block)
			if err != nil {
				return fmt.Errorf("error storing block %d: %v", block.BlockHeight, err)
			}
			fg.lastProcessedHeight = block.BlockHeight
			fg.logger.Info("Inserted new finalized block", zap.Uint64("block_height", block.BlockHeight))
		}
	}

	return nil
}

func normalizeBlockHash(hash string) string {
	return common.HexToHash(hash).Hex()
}
