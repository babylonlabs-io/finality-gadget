package finalitygadget

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"sync"
	"time"

	"github.com/babylonchain/babylon-finality-gadget/db"
	"github.com/babylonchain/babylon-finality-gadget/finalitygadget/config"
	"github.com/babylonchain/babylon-finality-gadget/sdk/btcclient"
	"github.com/babylonchain/babylon-finality-gadget/sdk/client"
	sdkconfig "github.com/babylonchain/babylon-finality-gadget/sdk/config"
	"github.com/babylonchain/babylon-finality-gadget/sdk/cwclient"
	"github.com/babylonchain/babylon-finality-gadget/types"
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

	// Create verifier
	return &FinalityGadget{
		SdkClient:    sdkClient,
		L2Client:     l2Client,
		Db:           db,
		PollInterval: cfg.PollInterval,
	}, nil
}

// This function process blocks indefinitely, starting from the last finalized block.
func (s *FinalityGadget) ProcessBlocks() error {
	// Start service at last finalized block
	err := s.startService()
	if err != nil {
		return fmt.Errorf("%v", err)
	}

	// Start polling for new blocks at set interval
	ticker := time.NewTicker(s.PollInterval)
	for range ticker.C {
		block, err := s.getBlockByNumber(int64(s.currHeight + 1))
		if err != nil {
			log.Fatalf("error getting new block: %v\n", err)
			continue
		}
		go func() {
			s.handleBlock(block)
		}()
	}

	return nil
}

// Start service at last finalized block
func (s *FinalityGadget) startService() error {
	// Query L2 node for last finalized block
	block, err := s.getLatestFinalizedBlock()
	if err != nil {
		return fmt.Errorf("error getting last finalized block: %v", err)
	}

	// Query local DB for last block processed
	localBlock, err := s.Db.GetLatestBlock()
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
		isFinal, err := s.queryIsBlockBabylonFinalized(block)
		// If not finalized, throw error
		if !isFinal {
			return fmt.Errorf("block %d should be finalized according to client but is not", block.BlockHeight)
		}
		if err != nil {
			return fmt.Errorf("error checking block %d: %v", block.BlockHeight, err)
		}
		// If finalised, store the block in DB and set the last finalized block
		err = s.InsertBlock(block)
		if err != nil {
			return fmt.Errorf("error storing block %d: %v", block.BlockHeight, err)
		}
	}

	// Start service at block height
	log.Printf("Starting verifier at block %d...\n", block.BlockHeight)

	// Set the start block and curr finalized block in memory
	s.startHeight = block.BlockHeight
	s.currHeight = block.BlockHeight

	return nil
}

// Get last btc finalized block
func (s *FinalityGadget) getLatestFinalizedBlock() (*types.Block, error) {
	return s.getBlockByNumber(ethrpc.FinalizedBlockNumber.Int64())
}

// Get block by number
func (s *FinalityGadget) getBlockByNumber(blockNumber int64) (*types.Block, error) {
	header, err := s.L2Client.HeaderByNumber(context.Background(), big.NewInt(blockNumber))
	if err != nil {
		return nil, err
	}
	return &types.Block{
		BlockHeight:    header.Number.Uint64(),
		BlockHash:      hex.EncodeToString(header.Hash().Bytes()),
		BlockTimestamp: header.Time,
	}, nil
}

func (s *FinalityGadget) handleBlock(block *types.Block) {
	// while block is not finalized, recheck if block is finalized every `retryInterval` seconds
	// if finalized, store the block in DB and set the last finalized block
	for {
		// Check if block is finalized
		isFinal, err := s.queryIsBlockBabylonFinalized(block)
		if err != nil {
			log.Fatalf("Error checking block %d: %v\n", block.BlockHeight, err)
			return
		}
		if isFinal {
			err = s.InsertBlock(block)
			if err != nil {
				log.Fatalf("Error storing block %d: %v\n", block.BlockHeight, err)
			}
			return
		}

		// Sleep for `PollInterval` seconds
		time.Sleep(s.PollInterval * time.Second)
	}
}

func (vf *FinalityGadget) queryIsBlockBabylonFinalized(block *types.Block) (bool, error) {
	return vf.SdkClient.QueryIsBlockBabylonFinalized(cwclient.L2Block{
		BlockHash:      string(block.BlockHash),
		BlockHeight:    block.BlockHeight,
		BlockTimestamp: block.BlockTimestamp,
	})
}

func (s *FinalityGadget) InsertBlock(block *types.Block) error {
	// Lock mutex
	s.Mutex.Lock()
	// Store block in DB
	err := s.Db.InsertBlock(&types.Block{
		BlockHeight:    block.BlockHeight,
		BlockHash:      block.BlockHash,
		BlockTimestamp: block.BlockTimestamp,
	})
	if err != nil {
		return err
	}
	// Set last finalized block in memory
	s.currHeight = block.BlockHeight
	// Unlock mutex
	s.Mutex.Unlock()

	return nil
}

func (s *FinalityGadget) GetBlockStatusByHeight(height uint64) (bool, error) {
	return s.Db.GetBlockStatusByHeight(height)
}

func (s *FinalityGadget) GetBlockStatusByHash(hash string) (bool, error) {
	return s.Db.GetBlockStatusByHash(hash)
}

func (s *FinalityGadget) GetLatestBlock() (*types.Block, error) {
	return s.Db.GetLatestBlock()
}

func (s *FinalityGadget) Close() {
	s.L2Client.Close()
}
