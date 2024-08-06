package finalitygadget

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	bbnClient "github.com/babylonlabs-io/babylon/client/client"
	bbncfg "github.com/babylonlabs-io/babylon/client/config"
	"github.com/babylonlabs-io/finality-gadget/bbnclient"
	"github.com/babylonlabs-io/finality-gadget/btcclient"
	"github.com/babylonlabs-io/finality-gadget/config"
	"github.com/babylonlabs-io/finality-gadget/cwclient"
	"github.com/babylonlabs-io/finality-gadget/db"
	"github.com/babylonlabs-io/finality-gadget/testutil"
	"github.com/babylonlabs-io/finality-gadget/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.uber.org/zap"
)

var _ types.IFinalityGadget = &FinalityGadget{}

type FinalityGadget struct {
	btcClient types.IBitcoinClient
	bbnClient types.IBabylonClient
	cwClient  types.ICosmWasmClient
	l2Client  *ethclient.Client

	db    *db.BBoltHandler
	mutex sync.Mutex

	pollInterval time.Duration
	startHeight  uint64
	currHeight   uint64
}

func NewFinalityGadget(cfg *config.Config, db *db.BBoltHandler) (*FinalityGadget, error) {
	// Create logger
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}

	// Create bitcoin client
	btcConfig := btcclient.DefaultBTCConfig()
	btcConfig.RPCHost = cfg.BitcoinRPCHost
	var btcClient types.IBitcoinClient
	// Create BTC client
	switch cfg.BBNChainID {
	// TODO: once we set up our own local BTC devnet, we don't need to use this mock BTC client
	case config.BabylonLocalnetChainID:
		btcClient, err = testutil.NewMockBitcoinClient(btcConfig, logger)
	default:
		btcClient, err = btcclient.NewBitcoinClient(btcConfig, logger)
	}
	if err != nil {
		return nil, err
	}

	// Create babylon client
	bbnConfig := bbncfg.DefaultBabylonConfig()
	bbnConfig.RPCAddr = cfg.BBNRPCAddress
	bbnConfig.ChainID = cfg.BBNChainID
	babylonClient, err := bbnClient.New(
		&bbnConfig,
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Babylon client: %w", err)
	}

	// Create cosmwasm client
	cwClient := cwclient.NewClient(babylonClient.QueryClient.RPCClient, cfg.FGContractAddress)

	// Create L2 client
	l2Client, err := ethclient.Dial(cfg.L2RPCHost)
	if err != nil {
		return nil, fmt.Errorf("failed to create OPStack L2 client: %w", err)
	}

	// Create finality gadget
	return &FinalityGadget{
		btcClient:    btcClient,
		bbnClient:    &bbnclient.BabylonClient{QueryClient: babylonClient.QueryClient},
		cwClient:     cwClient,
		l2Client:     l2Client,
		db:           db,
		pollInterval: cfg.PollInterval,
	}, nil
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
				log.Fatalf("error getting new block: %v\n", err)
				continue
			}
			go func() {
				fg.handleBlock(block)
			}()
		}
	}
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
	if localBlock.BlockHeight != 0 && localBlock.BlockHeight >= block.BlockHeight {
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
	log.Printf("Starting finality gadget at block %d...\n", block.BlockHeight)

	// Set the start block and curr finalized block in memory
	fg.startHeight = block.BlockHeight
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
			log.Fatalf("Error checking block %d: %v\n", block.BlockHeight, err)
			return
		}
		if isFinal {
			err = fg.InsertBlock(block)
			if err != nil {
				log.Fatalf("Error storing block %d: %v\n", block.BlockHeight, err)
			}
			return
		}

		// Sleep for `PollInterval` seconds
		time.Sleep(fg.pollInterval * time.Second)
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
