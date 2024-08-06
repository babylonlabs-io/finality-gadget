package finalitygadget

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"sync"
	"time"

	"github.com/babylonlabs-io/finality-gadget/db"
	"github.com/babylonlabs-io/finality-gadget/finalitygadget/config"
	"github.com/babylonlabs-io/finality-gadget/sdk/btcclient"
	"github.com/babylonlabs-io/finality-gadget/sdk/client"
	sdkconfig "github.com/babylonlabs-io/finality-gadget/sdk/config"
	"github.com/babylonlabs-io/finality-gadget/sdk/cwclient"
	"github.com/babylonlabs-io/finality-gadget/types"
	"github.com/ethereum/go-ethereum/ethclient"
	ethrpc "github.com/ethereum/go-ethereum/rpc"
)

var _ IFinalityGadget = &FinalityGadget{}

type FinalityGadget struct {
	SdkClient *client.SdkClient
	L2Client  *ethclient.Client
	Db        *db.BBoltHandler

	Mutex sync.Mutex

	PollInterval time.Duration
	startHeight  uint64
	currHeight   uint64
}

func NewFinalityGadget(cfg *config.Config, db *db.BBoltHandler) (*FinalityGadget, error) {
	// Create finality gadget client
	btcConfig := btcclient.DefaultBTCConfig()
	btcConfig.RPCHost = cfg.BitcoinRPCHost
	sdkClient, err := client.NewClient(&sdkconfig.Config{
		BTCConfig:    btcConfig,
		ContractAddr: cfg.FGContractAddress,
		ChainID:      cfg.BBNChainID,
		RPCAddr:      cfg.BBNRPCAddress,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating client: %w", err)
	}

	// Create L2 client
	l2Client, err := ethclient.Dial(cfg.L2RPCHost)
	if err != nil {
		return nil, fmt.Errorf("failed to create OPStack L2 client: %w", err)
	}

	// Create finality gadget
	return &FinalityGadget{
		SdkClient:    sdkClient,
		L2Client:     l2Client,
		Db:           db,
		PollInterval: cfg.PollInterval,
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
	ticker := time.NewTicker(fg.PollInterval)
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
	localBlock, err := fg.Db.GetLatestBlock()
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
		isFinal, err := fg.queryIsBlockBabylonFinalized(block)
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

// Get last btc finalized block
func (fg *FinalityGadget) getLatestFinalizedBlock() (*types.Block, error) {
	return fg.getBlockByNumber(ethrpc.FinalizedBlockNumber.Int64())
}

// Get block by number
func (fg *FinalityGadget) getBlockByNumber(blockNumber int64) (*types.Block, error) {
	header, err := fg.L2Client.HeaderByNumber(context.Background(), big.NewInt(blockNumber))
	if err != nil {
		return nil, err
	}
	return &types.Block{
		BlockHeight:    header.Number.Uint64(),
		BlockHash:      hex.EncodeToString(header.Hash().Bytes()),
		BlockTimestamp: header.Time,
	}, nil
}

func (fg *FinalityGadget) handleBlock(block *types.Block) {
	// while block is not finalized, recheck if block is finalized every `retryInterval` seconds
	// if finalized, store the block in DB and set the last finalized block
	for {
		// Check if block is finalized
		isFinal, err := fg.queryIsBlockBabylonFinalized(block)
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
		time.Sleep(fg.PollInterval * time.Second)
	}
}

func (fg *FinalityGadget) queryIsBlockBabylonFinalized(block *types.Block) (bool, error) {
	return fg.SdkClient.QueryIsBlockBabylonFinalized(cwclient.L2Block{
		BlockHash:      string(block.BlockHash),
		BlockHeight:    block.BlockHeight,
		BlockTimestamp: block.BlockTimestamp,
	})
}

func (fg *FinalityGadget) InsertBlock(block *types.Block) error {
	// Lock mutex
	fg.Mutex.Lock()
	// Store block in DB
	err := fg.Db.InsertBlock(&types.Block{
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
	fg.Mutex.Unlock()

	return nil
}

func (fg *FinalityGadget) GetBlockStatusByHeight(height uint64) (bool, error) {
	return fg.Db.GetBlockStatusByHeight(height)
}

func (fg *FinalityGadget) GetBlockStatusByHash(hash string) (bool, error) {
	return fg.Db.GetBlockStatusByHash(normalizeBlockHash(hash))
}

func (fg *FinalityGadget) GetLatestBlock() (*types.Block, error) {
	return fg.Db.GetLatestBlock()
}

func (fg *FinalityGadget) DeleteDB() error {
	return fg.Db.DeleteDB()
}

func (fg *FinalityGadget) Close() {
	fg.L2Client.Close()
}
