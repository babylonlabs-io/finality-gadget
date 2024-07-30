package verifier

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"sync"
	"time"

	"github.com/babylonchain/babylon-finality-gadget/sdk/btcclient"
	"github.com/babylonchain/babylon-finality-gadget/sdk/client"
	sdkconfig "github.com/babylonchain/babylon-finality-gadget/sdk/config"
	"github.com/babylonchain/babylon-finality-gadget/sdk/cwclient"
	"github.com/babylonchain/babylon-finality-gadget/verifier/config"
	"github.com/babylonchain/babylon-finality-gadget/verifier/db"
	"github.com/ethereum/go-ethereum/ethclient"
	ethrpc "github.com/ethereum/go-ethereum/rpc"
)

type Verifier struct {
	SdkClient *client.SdkClient
	L2Client  *ethclient.Client
	Db        *db.BBoltHandler

	Mutex sync.Mutex

	PollInterval time.Duration
	startHeight  uint64
	currHeight   uint64
}

type BlockInfo struct {
	Hash      string
	Height    uint64
	Timestamp uint64
}

func NewVerifier(cfg *config.Config, db *db.BBoltHandler) (*Verifier, error) {
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

	// Create verifier
	return &Verifier{
		SdkClient:    sdkClient,
		L2Client:     l2Client,
		Db:           db,
		PollInterval: cfg.PollInterval,
	}, nil
}

// This function process blocks indefinitely, starting from the last finalized block.
func (vf *Verifier) ProcessBlocks() error {
	// Start service at last finalized block
	err := vf.startService()
	if err != nil {
		return fmt.Errorf("%v", err)
	}

	// Start polling for new blocks at set interval
	ticker := time.NewTicker(vf.PollInterval)
	for range ticker.C {
		block, err := vf.getBlockByNumber(int64(vf.currHeight + 1))
		if err != nil {
			log.Fatalf("error getting new block: %v\n", err)
			continue
		}
		go func() {
			vf.handleBlock(block)
		}()
	}

	return nil
}

// Start service at last finalized block
func (vf *Verifier) startService() error {
	// Query L2 node for last finalized block
	block, err := vf.getLatestFinalizedBlock()
	if err != nil {
		return fmt.Errorf("error getting last finalized block: %v", err)
	}

	// Query local DB for last block processed
	localBlock, err := vf.Db.GetLatestBlock()
	if err != nil {
		return fmt.Errorf("error getting latest block from db: %v", err)
	}

	// if local chain tip is ahead of node, start service at latest block
	if localBlock.BlockHeight != 0 && localBlock.BlockHeight >= block.Height {
		block = &BlockInfo{
			Height:    localBlock.BlockHeight,
			Hash:      localBlock.BlockHash,
			Timestamp: localBlock.BlockTimestamp,
		}
	} else {
		// throw if block height is 0
		if block.Height == 0 {
			return fmt.Errorf("block height 0")
		}

		// Check the block is finalized using sdk client
		isFinal, err := vf.queryIsBlockBabylonFinalized(block)
		// If not finalized, throw error
		if !isFinal {
			return fmt.Errorf("block %d should be finalized according to client but is not", block.Height)
		}
		if err != nil {
			return fmt.Errorf("error checking block %d: %v", block.Height, err)
		}
		// If finalised, store the block in DB and set the last finalized block
		err = vf.insertBlock(block)
		if err != nil {
			return fmt.Errorf("error storing block %d: %v", block.Height, err)
		}
	}

	// Start service at block height
	log.Printf("Starting verifier at block %d...\n", block.Height)

	// Set the start block and curr finalized block in memory
	vf.startHeight = block.Height
	vf.currHeight = block.Height

	return nil
}

// Get last btc finalized block
func (vf *Verifier) getLatestFinalizedBlock() (*BlockInfo, error) {
	return vf.getBlockByNumber(ethrpc.FinalizedBlockNumber.Int64())
}

// Get block by number
func (vf *Verifier) getBlockByNumber(blockNumber int64) (*BlockInfo, error) {
	header, err := vf.L2Client.HeaderByNumber(context.Background(), big.NewInt(blockNumber))
	if err != nil {
		return nil, err
	}
	return &BlockInfo{
		Height:    header.Number.Uint64(),
		Hash:      hex.EncodeToString(header.Hash().Bytes()),
		Timestamp: header.Time,
	}, nil
}

func (vf *Verifier) handleBlock(block *BlockInfo) {
	// while block is not finalized, recheck if block is finalized every `retryInterval` seconds
	// if finalized, store the block in DB and set the last finalized block
	for {
		// Check if block is finalized
		isFinal, err := vf.queryIsBlockBabylonFinalized(block)
		if err != nil {
			log.Fatalf("Error checking block %d: %v\n", block.Height, err)
			return
		}
		if isFinal {
			err = vf.insertBlock(block)
			if err != nil {
				log.Fatalf("Error storing block %d: %v\n", block.Height, err)
			}
			return
		}

		// Sleep for `PollInterval` seconds
		time.Sleep(vf.PollInterval * time.Second)
	}
}

func (vf *Verifier) queryIsBlockBabylonFinalized(block *BlockInfo) (bool, error) {
	return vf.SdkClient.QueryIsBlockBabylonFinalized(cwclient.L2Block{
		BlockHash:      string(block.Hash),
		BlockHeight:    block.Height,
		BlockTimestamp: block.Timestamp,
	})
}

func (vf *Verifier) insertBlock(block *BlockInfo) error {
	// Lock mutex
	vf.Mutex.Lock()
	// Store block in DB
	err := vf.Db.InsertBlock(&db.Block{
		BlockHeight:    block.Height,
		BlockHash:      block.Hash,
		BlockTimestamp: block.Timestamp,
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

func (vf *Verifier) Close() {
	vf.L2Client.Close()
}
