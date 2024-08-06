package db

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/babylonlabs-io/finality-gadget/types"
	bolt "go.etcd.io/bbolt"
)

type BBoltHandler struct {
	db *bolt.DB
}

const (
	blocksBucket       = "blocks"
	blockHeightsBucket = "block_heights"
	latestBlockBucket  = "latest_block"
	latestBlockKey     = "latest"
)

var (
	ErrBlockNotFound = errors.New("block not found")
)

func NewBBoltHandler(path string) (*BBoltHandler, error) {
	// 0600 = read/write permission for owner only
	db, err := bolt.Open(path, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Fatalf("Error opening DB: %v\n", err)
		return nil, err
	}

	return &BBoltHandler{
		db: db,
	}, nil
}

func (bb *BBoltHandler) TryCreateInitialBuckets() error {
	log.Printf("Initialising DB...")
	return bb.db.Update(func(tx *bolt.Tx) error {
		buckets := []string{blocksBucket, blockHeightsBucket, latestBlockBucket}
		for _, bucket := range buckets {
			if err := tryCreateBucket(tx, bucket); err != nil {
				return err
			}
		}
		return nil
	})
}

func tryCreateBucket(tx *bolt.Tx, bucketName string) error {
	_, err := tx.CreateBucketIfNotExists([]byte(bucketName))
	if err != nil {
		log.Fatalf("Error creating bucket: %v\n", err)
	}
	return err
}

func (bb *BBoltHandler) InsertBlock(block *types.Block) error {
	log.Printf("Inserting block %d to DB...\n", block.BlockHeight)

	// Store mapping number -> block
	err := bb.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		key := itob(block.BlockHeight)
		blockBytes, err := json.Marshal(block)
		if err != nil {
			return err
		}
		return b.Put(key, blockBytes)
	})
	if err != nil {
		log.Fatalf("Error inserting block: %v\n", err)
		return err
	}

	// Store mapping hash -> number
	err = bb.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blockHeightsBucket))
		return b.Put([]byte(block.BlockHash), itob(block.BlockHeight))
	})
	if err != nil {
		log.Fatalf("Error inserting block: %v\n", err)
		return err
	}

	// Get current latest block
	latestBlock, err := bb.GetLatestBlock()
	if err != nil {
		log.Fatalf("Error getting latest block: %v\n", err)
		return err
	}

	// Update latest block if it's the latest
	err = bb.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(latestBlockBucket))
		if err != nil {
			log.Fatalf("Error getting latest block: %v\n", err)
			return err
		}
		if latestBlock.BlockHeight < block.BlockHeight {
			return b.Put([]byte(latestBlockKey), itob(block.BlockHeight))
		}
		return nil
	})
	if err != nil {
		log.Fatalf("Error updating latest block: %v\n", err)
		return err
	}

	return nil
}

func (bb *BBoltHandler) GetBlockByHeight(height uint64) (*types.Block, error) {
	var block types.Block
	err := bb.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		v := b.Get(itob(height))
		if v == nil {
			return ErrBlockNotFound
		}
		return json.Unmarshal(v, &block)
	})
	if err != nil {
		return nil, err
	}
	return &block, nil
}

func (bb *BBoltHandler) GetBlockStatusByHeight(height uint64) (bool, error) {
	_, err := bb.GetBlockByHeight(height)
	if err != nil {
		if errors.Is(err, ErrBlockNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (bb *BBoltHandler) GetBlockStatusByHash(hash string) (bool, error) {
	// Fetch block number by hash
	var blockHeight uint64
	err := bb.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blockHeightsBucket))
		res := b.Get([]byte(hash))
		if len(res) == 0 {
			return ErrBlockNotFound
		}
		blockHeight = btoi(res)
		return nil
	})
	if err != nil {
		if errors.Is(err, ErrBlockNotFound) {
			return false, nil
		}
		return false, err
	}

	return bb.GetBlockStatusByHeight(blockHeight)
}

func (bb *BBoltHandler) GetLatestBlock() (*types.Block, error) {
	var latestBlockHeight uint64

	// Fetch latest block height
	err := bb.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(latestBlockBucket))
		v := b.Get([]byte(latestBlockKey))
		if v == nil {
			return ErrBlockNotFound
		}
		latestBlockHeight = btoi(v)
		return nil
	})
	if err != nil {
		// If no latest block has been stored yet, return empty block (block 0)
		if errors.Is(err, ErrBlockNotFound) {
			return &types.Block{
				BlockHeight:    0,
				BlockHash:      "",
				BlockTimestamp: 0,
			}, nil
		}
		log.Fatalf("Error getting latest block: %v\n", err)
		return nil, err
	}

	// Fetch latest block by height
	return bb.GetBlockByHeight(latestBlockHeight)
}

func (bb *BBoltHandler) DeleteDB() error {
	absPath, err := filepath.Abs(bb.db.Path())
	if err != nil {
		log.Fatalf("failed to get db absolute path: %v\n", err)
		return fmt.Errorf("failed to get db absolute path: %w", err)
	}

	return os.Remove(absPath)
}

func (bb *BBoltHandler) Close() error {
	return bb.db.Close()
}

func itob(v uint64) []byte {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, v)
	if err != nil {
		log.Fatalf("Error writing to buffer: %v\n", err)
	}
	return buf.Bytes()
}

func btoi(b []byte) uint64 {
	var v uint64
	buf := bytes.NewReader(b)
	err := binary.Read(buf, binary.BigEndian, &v)
	if err != nil {
		log.Fatalf("Error reading from buffer: %v\n", err)
	}
	return v
}