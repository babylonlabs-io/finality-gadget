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

	bbnclient "github.com/babylonlabs-io/babylon/client/client"
	bbncfg "github.com/babylonlabs-io/babylon/client/config"
	fgbbnclient "github.com/babylonlabs-io/finality-gadget/bbnclient"
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
	babylonClient, err := bbnclient.New(
		&bbnConfig,
		logger,
	)
	bbnClient := fgbbnclient.NewBabylonClient(babylonClient.QueryClient)
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
	var btcClient IBitcoinClient
	switch cfg.BitcoinRPCHost {
	case "mock-btc-client":
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

	lastProcessedHeight := uint64(0)
	latestBlock, err := db.QueryLatestFinalizedBlock()
	if err != nil && !errors.Is(err, types.ErrBlockNotFound) {
		return nil, fmt.Errorf("failed to query latest finalized block: %w", err)
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

// QueryBtcStakingActivatedTimestamp retrieves BTC staking activation timestamp from the database
// returns math.MaxUint64, error if any error occurs
func (fg *FinalityGadget) QueryBtcStakingActivatedTimestamp() (uint64, error) {
	// First, try to get the timestamp from the database
	timestamp, err := fg.db.GetActivatedTimestamp()
	if err != nil {
		// If error is not found, try to query it from the bbnClient
		if errors.Is(err, types.ErrActivatedTimestampNotFound) {
			fg.logger.Debug("activation timestamp hasn't been set yet, querying from bbnClient...")
			return fg.queryBtcStakingActivationTimestamp()
		}
		fg.logger.Error("Failed to get activated timestamp from database", zap.Error(err))
		return math.MaxUint64, err
	}
	fg.logger.Debug("BTC staking activated timestamp found in database", zap.Uint64("timestamp", timestamp))
	return timestamp, nil
}

func (fg *FinalityGadget) GetBlockByHeight(height uint64) (*types.Block, error) {
	return fg.db.GetBlockByHeight(height)
}

func (fg *FinalityGadget) GetBlockByHash(hash string) (*types.Block, error) {
	return fg.db.GetBlockByHash(normalizeBlockHash(hash))
}

func (fg *FinalityGadget) QueryTransactionStatus(txHash string) (*types.TransactionInfo, error) {
	if err := validateEVMTxHash(txHash); err != nil {
		return nil, err
	}

	// get block info
	ctx := context.Background()
	txReceipt, err := fg.l2Client.TransactionReceipt(ctx, txHash)
	fg.logger.Debug("Transaction receipt", zap.Uint64("block_number", txReceipt.BlockNumber.Uint64()))
	if err != nil {
		return nil, err
	}
	header, err := fg.l2Client.HeaderByNumber(ctx, txReceipt.BlockNumber)
	fg.logger.Debug("Block info", zap.String("block_hash", header.Hash().Hex()), zap.Uint64("block_timestamp", header.Time))
	if err != nil {
		return nil, err
	}

	// get babylon finalized info
	isBabylonFinalized, err := fg.QueryIsBlockFinalizedByHeight(txReceipt.BlockNumber.Uint64())
	fg.logger.Debug("Babylon finalization status", zap.Bool("is_finalized", isBabylonFinalized))
	if err != nil {
		return nil, err
	}

	// get safe and finalized blocks
	safeBlock, err := fg.l2Client.HeaderByNumber(ctx, big.NewInt(ethrpc.SafeBlockNumber.Int64()))
	fg.logger.Debug("Safe block", zap.Uint64("block_number", safeBlock.Number.Uint64()))
	if err != nil {
		return nil, err
	}
	finalizedBlock, err := fg.l2Client.HeaderByNumber(ctx, big.NewInt(ethrpc.FinalizedBlockNumber.Int64()))
	fg.logger.Debug("Finalized block", zap.Uint64("block_number", finalizedBlock.Number.Uint64()))
	if err != nil {
		return nil, err
	}

	var status types.FinalityStatus
	if finalizedBlock.Number.Uint64() >= header.Number.Uint64() {
		status = types.FinalityStatusFinalized
	} else if isBabylonFinalized {
		status = types.FinalityStatusBitcoinFinalized
	} else if safeBlock.Number.Uint64() >= header.Number.Uint64() {
		status = types.FinalityStatusSafe
	} else {
		status = types.FinalityStatusPending
	}

	fg.logger.Debug("Transaction status", zap.String("block_hash", header.Hash().Hex()), zap.Uint64("block_height", header.Number.Uint64()), zap.String("tx_hash", txHash), zap.String("status", string(status)))

	return &types.TransactionInfo{
		TxHash:           txReceipt.TxHash.Hex(),
		BlockHeight:      header.Number.Uint64(),
		BlockHash:        hex.EncodeToString(header.Hash().Bytes()),
		BlockTimestamp:   header.Time,
		Status:           status,
		BabylonFinalized: isBabylonFinalized || status == types.FinalityStatusFinalized,
	}, nil
}

func (fg *FinalityGadget) QueryChainSyncStatus() (*types.ChainSyncStatus, error) {
	// Query latest block number
	ctx := context.Background()
	latestBlock, err := fg.l2Client.HeaderByNumber(ctx, big.NewInt(ethrpc.LatestBlockNumber.Int64()))
	if err != nil {
		return nil, err
	}

	// Query latest btc finalized block number
	latestBtcFinalizedBlock, err := fg.QueryLatestFinalizedBlock()
	if err != nil {
		return nil, err
	}

	// Query earliest btc finalized block number
	earliestBtcFinalizedBlock, err := fg.db.QueryEarliestFinalizedBlock()
	if err != nil {
		return nil, err
	}

	// Query latest eth finalized block number
	latestEthFinalizedBlock, err := fg.l2Client.HeaderByNumber(ctx, big.NewInt(ethrpc.FinalizedBlockNumber.Int64()))
	if err != nil {
		return nil, err
	}

	return &types.ChainSyncStatus{
		LatestBlockHeight:               latestBlock.Number.Uint64(),
		LatestBtcFinalizedBlockHeight:   latestBtcFinalizedBlock.BlockHeight,
		EarliestBtcFinalizedBlockHeight: earliestBtcFinalizedBlock.BlockHeight,
		LatestEthFinalizedBlockHeight:   latestEthFinalizedBlock.Number.Uint64(),
	}, nil
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
				return fmt.Errorf("error fetching latest finalized L2 block: %w", err)
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

			// at FG startup, this can avoid indexing from blocks that's not activated yet
			// TODO: we can add a flag fullSync
			// true: sync from the first btc finalized block (convertL2BlockHeight(btcStakingActivatedTimestamp))
			// false: sync from the last btc finalized block
			if fg.lastProcessedHeight == 0 {
				fg.lastProcessedHeight = latestFinalizedHeight - 1
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

// Query the BTC staking activation timestamp from bbnClient
// returns math.MaxUint64, ErrBtcStakingNotActivated if the BTC staking is not activated
func (fg *FinalityGadget) queryBtcStakingActivationTimestamp() (uint64, error) {
	allFpPks, err := fg.queryAllFpBtcPubKeys()
	if err != nil {
		return math.MaxUint64, err
	}
	fg.logger.Debug("All consumer FP public keys", zap.Strings("allFpPks", allFpPks))

	earliestDelHeight, err := fg.bbnClient.QueryEarliestActiveDelBtcHeight(allFpPks)
	if err != nil {
		return math.MaxUint64, err
	}
	if earliestDelHeight == math.MaxUint64 {
		return math.MaxUint64, types.ErrBtcStakingNotActivated
	}
	fg.logger.Debug("Earliest active delegation height", zap.Uint64("height", earliestDelHeight))

	btcBlockTimestamp, err := fg.btcClient.GetBlockTimestampByHeight(earliestDelHeight)
	if err != nil {
		return math.MaxUint64, err
	}
	fg.logger.Debug("BTC staking activated at", zap.Uint64("timestamp", btcBlockTimestamp))

	return btcBlockTimestamp, nil
}

// periodically check and update the BTC staking activation timestamp
// Exit the goroutine once we've successfully saved the timestamp
func (fg *FinalityGadget) MonitorBtcStakingActivation(ctx context.Context) {
	ticker := time.NewTicker(fg.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			timestamp, err := fg.queryBtcStakingActivationTimestamp()
			if err != nil {
				if errors.Is(err, types.ErrBtcStakingNotActivated) {
					fg.logger.Debug("BTC staking not yet activated, waiting...")
					continue
				}
				fg.logger.Error("Failed to query BTC staking activation timestamp", zap.Error(err))
				continue
			}

			err = fg.db.SaveActivatedTimestamp(timestamp)
			if err != nil {
				fg.logger.Error("Failed to save activated timestamp to database", zap.Error(err))
				continue
			}
			fg.logger.Debug("Saved BTC staking activated timestamp to database", zap.Uint64("timestamp", timestamp))
			return
		}
	}
}

func normalizeBlockHash(hash string) string {
	return common.HexToHash(hash).Hex()
}

// validateEVMTxHash checks if the given string is a valid EVM transaction hash
func validateEVMTxHash(txHash string) error {
	if len(txHash) != 66 || txHash[:2] != "0x" {
		return fmt.Errorf("invalid EVM transaction hash")
	}
	return nil
}
