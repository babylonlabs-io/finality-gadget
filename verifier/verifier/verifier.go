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

	// Create verifier
	return &Verifier{
		SdkClient: sdkClient,
		L2Client: l2Client,
		Pg: pg,
		PollInterval: cfg.PollInterval,
	}, nil
}

func (vf *Verifier) ProcessBlocks(ctx context.Context) error {
	// Query L2 node for last finalized block
	block, err := vf.getLatestFinalizedBlock(ctx)
	if err != nil {
		return fmt.Errorf("error getting last finalized block: %v", err)
	}

	// If this block already exists in our DB, skip the finality check
	exists := vf.Pg.GetBlockStatusByHeight(ctx, block.Height)
	if err != nil {
		return fmt.Errorf("error checking block %d: %v", block.Height, err)
	}
	if !exists {
		// Check the block is finalized using sdk client
		isFinal, err := vf.queryIsBlockBabylonFinalized(ctx, block)
		// If not finalized, throw error
		if !isFinal {
			return fmt.Errorf("block %d should be finalized according to client but is not", block.Height)
		}
		if err != nil {
			return fmt.Errorf("error checking block %d: %v", block.Height, err)
		}
		// If finalised, store the block in DB and set the last finalized block
		err = vf.insertBlock(ctx, block)
		if err != nil {
			return fmt.Errorf("error storing block %d: %v", block.Height, err)
		}
	}

	// Start service at block height
	fmt.Printf("starting service at block %d\n", block.Height)

	// Store the last finalized block height in memory
	vf.blockHeight = block.Height

	// Start polling for new blocks at set interval
	ticker := time.NewTicker(vf.PollInterval)
	for range ticker.C {
		fmt.Printf("polling for new block at height %d\n", vf.blockHeight + 1)
		block, err := vf.getBlockByNumber(ctx, int64(vf.blockHeight + 1))
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
	vf.blockHeight = block.Height
	// Unlock mutex
	vf.Mutex.Unlock()

	return nil
}