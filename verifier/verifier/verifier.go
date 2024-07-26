package verifier

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	"github.com/babylonchain/babylon-finality-gadget/sdk/btcclient"
	"github.com/babylonchain/babylon-finality-gadget/sdk/client"
	sdkconfig "github.com/babylonchain/babylon-finality-gadget/sdk/config"
	"github.com/babylonchain/babylon-finality-gadget/sdk/cwclient"
	"github.com/babylonchain/babylon-finality-gadget/verifier/db"
	"github.com/babylonchain/babylon-finality-gadget/verifier/server"
	"github.com/ethereum/go-ethereum/ethclient"
	ethrpc "github.com/ethereum/go-ethereum/rpc"
)

func NewVerifier(ctx context.Context, cfg *Config) (*Verifier, error) {
	// Create finality gadget client
	btcConfig := btcclient.DefaultBTCConfig()
	btcConfig.RPCHost = cfg.BitcoinRPCHost
	sdkClient, err := client.NewClient(&sdkconfig.Config{
		BTCConfig:    btcConfig,
		ContractAddr: cfg.FGContractAddress,
		ChainID: cfg.BBNChainID,
		RPCAddr: cfg.BBNRPCAddress,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating client: %w", err)
	}

	// Create L2 client
	l2Client, err := ethclient.Dial(cfg.L2RPCHost)
	if err != nil {
		return nil, fmt.Errorf("failed to create OPStack L2 client: %w", err)
	}

	// Create local DB for storing and querying blocks
	pg, err := db.NewPostgresHandler(ctx, cfg.PGConnectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres handler: %w", err)
	}
	err = pg.TryCreateInitialTables(ctx)
	if err != nil {
		return nil, fmt.Errorf("create initial tables error: %w", err)
	}

	// Start server
	go server.StartServer(ctx, &server.ServerConfig{
		Port:       cfg.ServerPort,
		ConnString: cfg.PGConnectionString,
	})

	// Create verifier
	return &Verifier{
		SdkClient: sdkClient,
		L2Client: l2Client,
		Pg: pg,
		PollInterval: cfg.PollInterval,
	}, nil
}

// This function process blocks indefinitely, starting from the last finalized block.
func (vf *Verifier) ProcessBlocks(ctx context.Context) error {
	return vf.ProcessNBlocks(ctx, -1)
}

// This function process n blocks, starting from the last finalized block.
// n is the number of blocks to process, pass in -1 to process blocks indefinitely.
// Passing in non-negative n is useful for testing.
func (vf *Verifier) ProcessNBlocks(ctx context.Context, n int) error {
	// Start service at last finalized block
	err := vf.startService(ctx)
	if err != nil {
		return fmt.Errorf("%v", err)
	}

	// Start polling for new blocks at set interval
	ticker := time.NewTicker(vf.PollInterval)
	for range ticker.C {
		// If n is positive and we have processed n blocks, stop the process
		if n > 0 && vf.currHeight == vf.startHeight + uint64(n - 1) {
			break
		}
		fmt.Printf("polling for new block at height %d\n", vf.currHeight + 1)
		block, err := vf.getBlockByNumber(ctx, int64(vf.currHeight + 1))
		if err != nil {
			fmt.Printf("error getting new block: %v\n", err)
			continue
		}
		go func() {
			vf.handleBlock(ctx, block)
		}()
	}

	return nil
}

// Start service at last finalized block
func (vf *Verifier) startService(ctx context.Context) error {
	// Query L2 node for last finalized block
	block, err := vf.getLatestFinalizedBlock(ctx)
	if err != nil {
		return fmt.Errorf("error getting last finalized block: %v", err)
	}
	fmt.Printf("[ProcessBlocks] last finalized block: %d\n", block.Height)

	// Query local DB for last block processed
	localBlock, err := vf.Pg.GetLatestBlock(ctx)
	if err != nil {
		return fmt.Errorf("error getting latest block from db: %v", err)
	}

	// if local chain tip is ahead of node, start service at latest block
	if localBlock.BlockHeight >= block.Height {
		block = &BlockInfo{
			Height: 		localBlock.BlockHeight,
			Hash:   		localBlock.BlockHash,
			Timestamp: 	localBlock.BlockTimestamp,
		}
	} else {
		// If block height is 0, replace it with block height 1
		if block.Height == 0 {
			block, err = vf.getBlockByNumber(ctx, 1)
			if err != nil {
				return fmt.Errorf("error getting block 1: %v", err)
			}
		}
	
		// Check the block is finalized using sdk client
		isFinal, err := vf.queryIsBlockBabylonFinalized(ctx, block)
		fmt.Printf("[ProcessBlocks] is block %d finalized: %v\n", block.Height, isFinal)
		// If not finalized, throw error
		if !isFinal {
			return fmt.Errorf("block %d should be finalized according to client but is not", block.Height)
		}
		if err != nil {
			return fmt.Errorf("error checking block %d: %v", block.Height, err)
		}
		// If finalised, store the block in DB and set the last finalized block
		fmt.Printf("[ProcessBlocks] storing block %d\n", block.Height)
		err = vf.insertBlock(ctx, block)
		if err != nil {
			return fmt.Errorf("error storing block %d: %v", block.Height, err)
		}
	}

	// Start service at block height
	fmt.Printf("starting service at block %d\n", block.Height)

	// Set the start block and curr finalized block in memory
	vf.startHeight = block.Height
	vf.currHeight = block.Height

	return nil
}

// Get last btc finalized block
func (vf *Verifier) getLatestFinalizedBlock(ctx context.Context) (*BlockInfo, error) {
	return vf.getBlockByNumber(ctx, ethrpc.FinalizedBlockNumber.Int64())
}

// Get block by number
func (vf *Verifier) getBlockByNumber(ctx context.Context, blockNumber int64) (*BlockInfo, error) {
	header, err := vf.L2Client.HeaderByNumber(ctx, big.NewInt(blockNumber))
	if err != nil {
		return nil, err
	}
	return &BlockInfo{
		Height: 		header.Number.Uint64(),
		Hash:   		hex.EncodeToString(header.Hash().Bytes()),
		Timestamp: 	header.Time,
	}, nil
}

func (vf *Verifier) handleBlock(ctx context.Context, block *BlockInfo) {
	// while block is not finalized, recheck if block is finalized every `retryInterval` seconds
	// if finalized, store the block in DB and set the last finalized block
	for {
		fmt.Printf("checking finality of block %d\n", block.Height)
		// Check if block is finalized
		isFinal, err := vf.queryIsBlockBabylonFinalized(ctx, block)
		if err != nil {
			fmt.Printf("error checking block %d: %v\n", block.Height, err)
			return
		}
		if isFinal {
			err = vf.insertBlock(ctx, block)
			if err != nil {
				fmt.Printf("error storing block %d: %v\n", block.Height, err)
			}
			return
		}

		// Sleep for `PollInterval` seconds
		time.Sleep(vf.PollInterval * time.Second)
	}
}

func (vf *Verifier) queryIsBlockBabylonFinalized(ctx context.Context, block *BlockInfo) (bool, error) {
	return vf.SdkClient.QueryIsBlockBabylonFinalized(cwclient.L2Block{
		BlockHash: 				string(block.Hash),
		BlockHeight: 			block.Height,
		BlockTimestamp: 	block.Timestamp,
	})
}

func (vf *Verifier) insertBlock(ctx context.Context, block *BlockInfo) error {
	// Lock mutex
	vf.Mutex.Lock()
	// Store block in DB
	err := vf.Pg.InsertBlock(ctx, db.Block{
		BlockHeight: 			block.Height,
		BlockHash:   			block.Hash,
		BlockTimestamp:   block.Timestamp,
		IsFinalized: 			true,
	})
	if err != nil {
		return err
	}
	// Set last finalized block in memory
	vf.currHeight = block.Height
	// Unlock mutex
	vf.Mutex.Unlock()

	return nil
}