package db

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"time"

	"github.com/babylonlabs-io/finality-gadget/types"
	bolt "go.etcd.io/bbolt"
	"go.uber.org/zap"
)

type BBoltHandler struct {
	db     *bolt.DB
	logger *zap.Logger
}

var _ IDatabaseHandler = &BBoltHandler{}

const (
	blocksBucket       = "blocks"
	blockHeightsBucket = "block_heights"
	latestBlockBucket  = "latest_block"
	latestBlockKey     = "latest"
)

//////////////////////////////
// CONSTRUCTOR
//////////////////////////////

func NewBBoltHandler(path string, logger *zap.Logger) (*BBoltHandler, error) {
	// 0600 = read/write permission for owner only
	db, err := bolt.Open(path, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		logger.Error("Error opening DB", zap.Error(err))
		return nil, err
	}

	return &BBoltHandler{
		db:     db,
		logger: logger,
	}, nil
}

//////////////////////////////
// METHODS
//////////////////////////////

func (bb *BBoltHandler) CreateInitialSchema() error {
	bb.logger.Info("Initialising DB...")
	return bb.db.Update(func(tx *bolt.Tx) error {
		buckets := []string{blocksBucket, blockHeightsBucket, latestBlockBucket}
		for _, bucket := range buckets {
			if err := bb.tryCreateBucket(tx, bucket); err != nil {
				return err
			}
		}
		return nil
	})
}

func (bb *BBoltHandler) InsertBlock(block *types.Block) error {
	bb.logger.Info("Inserting block to DB", zap.Uint64("block_height", block.BlockHeight))

	// Store mapping number -> block
	err := bb.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		key := bb.itob(block.BlockHeight)
		blockBytes, err := json.Marshal(block)
		if err != nil {
			return err
		}
		return b.Put(key, blockBytes)
	})
	if err != nil {
		bb.logger.Error("Error inserting block", zap.Error(err))
		return err
	}

	// Store mapping hash -> number
	err = bb.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blockHeightsBucket))
		return b.Put([]byte(block.BlockHash), bb.itob(block.BlockHeight))
	})
	if err != nil {
		bb.logger.Error("Error inserting block", zap.Error(err))
		return err
	}

	// Get current latest block
	latestBlock, err := bb.QueryLatestFinalizedBlock()
	if latestBlock == nil {
		latestBlock = &types.Block{BlockHeight: 0}
	}
	if err != nil {
		bb.logger.Error("Error getting latest block", zap.Error(err))
		return err
	}

	// Update latest block if it's the latest
	err = bb.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(latestBlockBucket))
		if err != nil {
			bb.logger.Error("Error getting latest block", zap.Error(err))
			return err
		}
		if latestBlock.BlockHeight < block.BlockHeight {
			return b.Put([]byte(latestBlockKey), bb.itob(block.BlockHeight))
		}
		return nil
	})
	if err != nil {
		bb.logger.Error("Error updating latest block", zap.Error(err))
		return err
	}

	return nil
}

func (bb *BBoltHandler) GetBlockByHeight(height uint64) (*types.Block, error) {
	var block types.Block
	err := bb.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		v := b.Get(bb.itob(height))
		if v == nil {
			return types.ErrBlockNotFound
		}
		return json.Unmarshal(v, &block)
	})
	if err != nil {
		return nil, err
	}
	return &block, nil
}

func (bb *BBoltHandler) GetBlockByHash(hash string) (*types.Block, error) {
	// Fetch block number corresponding to hash
	var blockHeight uint64
	err := bb.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blockHeightsBucket))
		v := b.Get([]byte(hash))
		if v == nil {
			return types.ErrBlockNotFound
		}
		blockHeight = bb.btoi(v)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return bb.GetBlockByHeight(blockHeight)
}

func (bb *BBoltHandler) QueryIsBlockFinalizedByHeight(height uint64) (bool, error) {
	_, err := bb.GetBlockByHeight(height)
	if err != nil {
		if errors.Is(err, types.ErrBlockNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (bb *BBoltHandler) QueryIsBlockFinalizedByHash(hash string) (bool, error) {
	// Fetch block number by hash
	var blockHeight uint64
	err := bb.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blockHeightsBucket))
		res := b.Get([]byte(hash))
		if len(res) == 0 {
			return types.ErrBlockNotFound
		}
		blockHeight = bb.btoi(res)
		return nil
	})
	if err != nil {
		if errors.Is(err, types.ErrBlockNotFound) {
			return false, nil
		}
		return false, err
	}

	return bb.QueryIsBlockFinalizedByHeight(blockHeight)
}

func (bb *BBoltHandler) QueryLatestFinalizedBlock() (*types.Block, error) {
	var latestBlockHeight uint64

	// Fetch latest block height
	err := bb.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(latestBlockBucket))
		v := b.Get([]byte(latestBlockKey))
		if v == nil {
			return types.ErrBlockNotFound
		}
		latestBlockHeight = bb.btoi(v)
		return nil
	})
	if err != nil {
		// If no latest block has been stored yet, return nil
		if errors.Is(err, types.ErrBlockNotFound) {
			return nil, nil
		}
		bb.logger.Error("Error getting latest block", zap.Error(err))
		return nil, err
	}

	// Fetch latest block by height
	return bb.GetBlockByHeight(latestBlockHeight)
}

func (bb *BBoltHandler) Close() error {
	return bb.db.Close()
}

//////////////////////////////
// INTERNAL
//////////////////////////////

func (bb *BBoltHandler) tryCreateBucket(tx *bolt.Tx, bucketName string) error {
	_, err := tx.CreateBucketIfNotExists([]byte(bucketName))
	if err != nil {
		bb.logger.Error("Error creating bucket", zap.Error(err))
	}
	return err
}

func (bb *BBoltHandler) itob(v uint64) []byte {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, v)
	if err != nil {
		bb.logger.Fatal("Error writing to buffer", zap.Error(err))
	}
	return buf.Bytes()
}

func (bb *BBoltHandler) btoi(b []byte) uint64 {
	var v uint64
	buf := bytes.NewReader(b)
	err := binary.Read(buf, binary.BigEndian, &v)
	if err != nil {
		bb.logger.Fatal("Error reading from buffer", zap.Error(err))
	}
	return v
}
