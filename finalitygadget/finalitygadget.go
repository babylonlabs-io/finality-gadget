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
	batchSize           uint64
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
		batchSize:           cfg.BatchSize,
		lastProcessedHeight: lastProcessedHeight,
		logger:              logger,
	}, nil
}

//////////////////////////////
// METHODS
//////////////////////////////

// TODO: make this method internal once fully tested. External services should query the database instead.
/* QueryIsBlockBabylonFinalizedFromBabylon checks if the given L2 block is finalized by querying the Babylon node
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
func (fg *FinalityGadget) QueryIsBlockBabylonFinalizedFromBabylon(block *types.Block) (bool, error) {
	if block == nil {
		return false, fmt.Errorf("block is nil")
	}

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

// QueryIsBlockBabylonFinalized queries the finality status of a given block height from the internal db
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

	// convert the L2 timestamp to BTC height
	btcblockHeight, err := fg.btcClient.GetBlockHeightByTimestamp(block.BlockTimestamp)
	if err != nil {
		return false, err
	}

	// check whether the btc staking is activated
	btcStakingActivatedTimestamp, err := fg.QueryBtcStakingActivatedTimestamp()
	if err != nil {
		return false, err
	}
	if btcblockHeight < btcStakingActivatedTimestamp {
		return false, types.ErrBtcStakingNotActivated
	}

	// query the finality status of the block from internal db
	return fg.db.QueryIsBlockFinalizedByHeight(block.BlockHeight)
}

/* QueryBlockRangeBabylonFinalized searches the internal db and returns the last consecutively finalized block in the block range
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

	// query the earliest and latest finalized block from internal db
	earliestFinalizedBlock, err := fg.db.QueryEarliestFinalizedBlock()
	if err != nil {
		return nil, err
	}
	if earliestFinalizedBlock == nil {
		fg.logger.Error("No earliest finalized block found")
		return nil, fmt.Errorf("no earliest finalized block found")
	}
	latestFinalizedBlock, err := fg.QueryLatestFinalizedBlock()
	if err != nil {
		return nil, err
	}
	if latestFinalizedBlock == nil {
		fg.logger.Error("No latest finalized block found")
		return nil, fmt.Errorf("no latest finalized block found")
	}

	// blocks inserted to the db must be consecutive, so we can simply perform a range
	// check against the first and last blocks

	// block range starts before earliest finalized block, or ends after latest finalized block,
	// then no blocks are consecutively finalized
	if queryBlocks[0].BlockHeight < earliestFinalizedBlock.BlockHeight ||
		queryBlocks[len(queryBlocks)-1].BlockHeight > latestFinalizedBlock.BlockHeight {
		return nil, nil
	}

	// both block ranges are consecutive, simply return the latest common block
	finalizedBlockHeight := latestFinalizedBlock.BlockHeight
	if queryBlocks[len(queryBlocks)-1].BlockHeight < latestFinalizedBlock.BlockHeight {
		finalizedBlockHeight = queryBlocks[len(queryBlocks)-1].BlockHeight
	}

	return &finalizedBlockHeight, nil
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
	if err != nil || txReceipt == nil {
		return nil, err
	}
	fg.logger.Debug("Transaction receipt", zap.Uint64("block_number", txReceipt.BlockNumber.Uint64()))
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

// This function is run once at startup and starts the FG from the last finalized block.
// Note that starting FG before the chain has ETH finalized its first block will cause a panic.
// The intended startup order for new chains is:
//  1. Start the OP chain with `babylonFinalityGadgetRpc` in rollup configs set to an empty string
//  2. Wait for the chain to finalize its first block
//  3. Integrate FG with it disabled on CW contract
//  3. Restart OP chain after setting `babylonFinalityGadgetRpc`
//  4. Enable FG on CW contract (for network with multiple nodes, enable after majority of nodes upgrade)
func (fg *FinalityGadget) Startup(ctx context.Context) error {
	fg.logger.Info("Starting up finality gadget...")
	// Start polling for new blocks at set interval
	ticker := time.NewTicker(fg.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			// query rpc for latest eth finalized block
			// at this point, FG is disabled so the derivation pipeline passes through
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

			// throw error if btc staking activated before the first block was finalized (see startup order above)
			if latestFinalizedHeight == 0 && btcStakingActivatedTimestamp < latestFinalizedBlockTime {
				return fmt.Errorf("BTC staking activated before the first finalized block")
			}

			// skip blocks before btc staking is activated
			if latestFinalizedBlockTime < btcStakingActivatedTimestamp {
				fg.logger.Info("Skipping block before BTC staking activation", zap.Uint64("block_height", latestFinalizedHeight))
				continue
			}

			// otherwise, startup the FG at latest finalized block (taking the later of the db and rpc values)
			latestFinalizedBlockDb, err := fg.QueryLatestFinalizedBlock()
			if err != nil {
				return fmt.Errorf("error fetching latest finalized block from db: %w", err)
			}
			fg.lastProcessedHeight = latestFinalizedHeight - 1
			if latestFinalizedBlockDb != nil && latestFinalizedBlockDb.BlockHeight > latestFinalizedHeight {
				fg.lastProcessedHeight = latestFinalizedBlockDb.BlockHeight
			}
			fg.logger.Info("Starting finality gadget from block", zap.Uint64("block_height", fg.lastProcessedHeight+1))

			return nil
		}
	}
}

// This function process blocks indefinitely, starting from the last finalized block.
func (fg *FinalityGadget) ProcessBlocks(ctx context.Context) error {
	fg.logger.Info("Processing blocks...")
	// Start polling for new blocks at set interval
	ticker := time.NewTicker(fg.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			fg.logger.Debug("Exiting block processing loop...")
			return nil
		case <-ticker.C:
			fg.logger.Debug("Processing new blocks...")
			// get latest block
			latestBlock, err := fg.l2Client.HeaderByNumber(ctx, big.NewInt(ethrpc.LatestBlockNumber.Int64()))
			if err != nil {
				return fmt.Errorf("error fetching latest L2 block: %w", err)
			}
			fg.logger.Debug("Received latest block", zap.Uint64("block_height", latestBlock.Number.Uint64()))

			// if the last processed block is less than the latest block, process all intervening blocks
			if fg.lastProcessedHeight < latestBlock.Number.Uint64() {
				fg.logger.Info("Processing new blocks", zap.Uint64("start_height", fg.lastProcessedHeight+1), zap.Uint64("end_height", latestBlock.Number.Uint64()))
				if err := fg.processBlocksTillHeight(ctx, latestBlock.Number.Uint64()); err != nil {
					return fmt.Errorf("error processing block %d: %w", latestBlock.Number.Uint64(), err)
				}
			}
		}
	}
}

func (fg *FinalityGadget) insertBlocks(blocks []*types.Block) error {
	// Lock mutex
	fg.mutex.Lock()
	defer fg.mutex.Unlock()

	// Normalize block hashes
	normalizedBlocks := make([]*types.Block, len(blocks))
	for i, block := range blocks {
		normalizedBlocks[i] = &types.Block{
			BlockHeight:    block.BlockHeight,
			BlockHash:      normalizeBlockHash(block.BlockHash),
			BlockTimestamp: block.BlockTimestamp,
		}
	}

	// Store blocks in DB
	if err := fg.db.InsertBlocks(normalizedBlocks); err != nil {
		return fmt.Errorf("failed to batch insert blocks: %w", err)
	}

	return nil
}

func (fg *FinalityGadget) Close() {
	fg.l2Client.Close()

	if err := fg.db.Close(); err != nil {
		fg.logger.Error("Error closing database", zap.Error(err))
	}

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

// Process blocks in batches of size `fg.batchSize` until the latest height
func (fg *FinalityGadget) processBlocksTillHeight(ctx context.Context, latestHeight uint64) error {
	fg.logger.Debug("Processing blocks till height", zap.Uint64("height", latestHeight))
	for batchStartHeight := fg.lastProcessedHeight + 1; batchStartHeight <= latestHeight; {
		select {
		case <-ctx.Done():
			fg.logger.Debug("Exiting block processing loop...")
			return nil
		default:
			// Calculate batch start and end heights
			batchEndHeight := batchStartHeight + fg.batchSize - 1
			if batchEndHeight > latestHeight {
				batchEndHeight = latestHeight
			}
			fg.logger.Info("Processing batch of blocks", zap.Uint64("batch_start_height", batchStartHeight), zap.Uint64("batch_end_height", batchEndHeight))

			// Create batch of blocks to check in parallel
			results := make(chan *types.Block, batchEndHeight-batchStartHeight+1)
			errors := make(chan error, batchEndHeight-batchStartHeight+1)
			var wg sync.WaitGroup

			// Query batch in parallel
			for height := batchStartHeight; height <= batchEndHeight; height++ {
				wg.Add(1)
				go func(h uint64) {
					defer wg.Done()
					block, err := fg.processHeight(h)
					if block != nil && err == nil {
						fg.logger.Debug("Processed block", zap.Uint64("block_height", h), zap.String("block_hash", block.BlockHash), zap.Uint64("batch_start_height", batchStartHeight), zap.Uint64("batch_end_height", batchEndHeight))
					}
					results <- block
					errors <- err
				}(height)
			}

			// Close results channel once all goroutines complete
			go func() {
				wg.Wait()
				fg.logger.Debug("Closing channels for batch", zap.Uint64("batch_start_height", batchStartHeight), zap.Uint64("batch_end_height", batchEndHeight))
				close(results)
				close(errors)
			}()

			// Extract and handle error (if any)
			for err := range errors {
				if err != nil {
					return err
				}
			}

			// Extract blocks and find last consecutive finalized block.
			// As channels are async, blocks will NOT be ordered by height
			// We use a map to first sort blocks by height, then extract the last
			// consecutively finalized block.
			sortedBlocks := make(map[uint64]*types.Block)
			for block := range results {
				if block != nil {
					sortedBlocks[block.BlockHeight] = block
				}
			}

			var finalizedBlocks []*types.Block
			var lastFinalizedHeight uint64
			for i := batchStartHeight; i <= batchEndHeight; i++ {
				if block, ok := sortedBlocks[i]; ok {
					if block != nil {
						finalizedBlocks = append(finalizedBlocks, block)
						lastFinalizedHeight = block.BlockHeight
					}
				} else {
					break
				}
			}
			fg.logger.Debug("Last finalized block in batch", zap.Uint64("block_height", lastFinalizedHeight), zap.Uint64("batch_start_height", batchStartHeight), zap.Uint64("batch_end_height", batchEndHeight))

			// If no blocks were finalized, wait for next poll
			if lastFinalizedHeight < batchStartHeight || len(finalizedBlocks) == 0 {
				fg.logger.Debug("No blocks finalized, waiting for next poll", zap.Uint64("batch_start_height", batchStartHeight), zap.Uint64("batch_end_height", batchEndHeight))
				return nil
			}

			// Batch insert all consecutive finalized blocks
			fg.logger.Debug("Inserting finalized blocks", zap.Uint64("start_height", finalizedBlocks[0].BlockHeight), zap.Uint64("end_height", finalizedBlocks[len(finalizedBlocks)-1].BlockHeight))
			if err := fg.insertBlocks(finalizedBlocks); err != nil {
				return fmt.Errorf("error storing blocks: %w", err)
			}
			fg.lastProcessedHeight = lastFinalizedHeight

			// Update start height for next batch
			batchStartHeight = lastFinalizedHeight + 1
		}
	}

	return nil
}

func (fg *FinalityGadget) processHeight(height uint64) (*types.Block, error) {
	fg.logger.Debug("Processing block", zap.Uint64("block_height", height))
	// Fetch block from rpc
	if height > math.MaxInt64 {
		fg.logger.Debug("Block height exceeds maximum int64 value", zap.Uint64("block_height", height))
		return nil, fmt.Errorf("block height %d exceeds maximum int64 value", height)
	}
	block, err := fg.queryBlockByHeight(int64(height))
	if err != nil || block == nil {
		fg.logger.Error("Error fetching block", zap.Uint64("block_height", height), zap.Error(err))
		return nil, fmt.Errorf("error getting block at height %d: %w", height, err)
	}
	fg.logger.Debug("Fetched block", zap.Uint64("block_height", height), zap.String("block_hash", block.BlockHash))

	// Check finalization
	isFinalized, err := fg.QueryIsBlockBabylonFinalizedFromBabylon(block)
	if err != nil {
		fg.logger.Error("Error checking if block is finalized from babylon", zap.Uint64("block_height", height), zap.Error(err))
		return nil, fmt.Errorf("error checking is block %d finalized from babylon: %w", height, err)
	}
	fg.logger.Debug("Fetched block finality status", zap.Uint64("block_height", height), zap.Bool("is_finalized", isFinalized))

	if !isFinalized {
		fg.logger.Debug("Block not finalized", zap.Uint64("block_height", height))
		return nil, nil
	}

	fg.logger.Debug("Block finalized", zap.Uint64("block_height", height))
	return block, nil
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
	if earliestDelHeight == math.MaxUint32 {
		return math.MaxUint64, types.ErrBtcStakingNotActivated
	}
	fg.logger.Debug("Earliest active delegation height", zap.Uint32("height", earliestDelHeight))

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
