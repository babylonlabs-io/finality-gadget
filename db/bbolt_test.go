package db

import (
	"math"
	"os"
	"testing"

	"github.com/babylonlabs-io/finality-gadget/log"
	"github.com/babylonlabs-io/finality-gadget/types"
	"github.com/stretchr/testify/assert"
)

func setupDB(t *testing.T) (*BBoltHandler, func()) {
	// Create temp test file.
	tempFile, err := os.CreateTemp("", "test-*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tempFile.Close()

	// Create logger.
	logger, err := log.NewRootLogger("console", true)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Create a new BBoltHandler
	db, err := NewBBoltHandler(tempFile.Name(), logger)
	if err != nil {
		t.Fatalf("Failed to create BBoltHandler: %v", err)
	}

	// Create initial buckets
	err = db.CreateInitialSchema()
	if err != nil {
		t.Fatalf("Failed to create initial buckets: %v", err)
	}

	// Cleanup function to close DB and remove temp file
	cleanup := func() {
		err := os.Remove(tempFile.Name())
		if err != nil {
			t.Fatalf("Failed to delete DB: %v", err)
		}
		db.Close()
	}

	return db, cleanup
}

func TestInsertBlock(t *testing.T) {
	handler, cleanup := setupDB(t)
	defer cleanup()

	block := &types.Block{
		BlockHeight:    1,
		BlockHash:      "0x123",
		BlockTimestamp: 1000,
	}

	err := handler.InsertBlock(block)
	assert.NoError(t, err)

	// Verify block was inserted
	retrievedBlock, err := handler.GetBlockByHeight(block.BlockHeight)
	assert.NoError(t, err)
	assert.Equal(t, block.BlockHeight, retrievedBlock.BlockHeight)
	assert.Equal(t, block.BlockHash, retrievedBlock.BlockHash)
	assert.Equal(t, block.BlockTimestamp, retrievedBlock.BlockTimestamp)
}

func TestGetBlockByHeight(t *testing.T) {
	handler, cleanup := setupDB(t)
	defer cleanup()

	// Insert a block
	block := &types.Block{
		BlockHeight:    1,
		BlockHash:      "0x123",
		BlockTimestamp: 1000,
	}
	err := handler.InsertBlock(block)
	assert.NoError(t, err)

	// Retrieve block by height
	retrievedBlock, err := handler.GetBlockByHeight(block.BlockHeight)
	assert.NoError(t, err)
	assert.Equal(t, block.BlockHeight, retrievedBlock.BlockHeight)
	assert.Equal(t, block.BlockHash, retrievedBlock.BlockHash)
	assert.Equal(t, block.BlockTimestamp, retrievedBlock.BlockTimestamp)
}

func TestGetBlockByHeightForNonExistentBlock(t *testing.T) {
	handler, cleanup := setupDB(t)
	defer cleanup()

	block, err := handler.GetBlockByHeight(1)
	assert.Nil(t, block)
	assert.Equal(t, types.ErrBlockNotFound, err)
}

func TestGetBlockByHash(t *testing.T) {
	handler, cleanup := setupDB(t)
	defer cleanup()

	// Insert a block
	block := &types.Block{
		BlockHeight:    1,
		BlockHash:      "0x123",
		BlockTimestamp: 1000,
	}
	err := handler.InsertBlock(block)
	assert.NoError(t, err)

	// Retrieve block by hash
	retrievedBlock, err := handler.GetBlockByHash(block.BlockHash)
	assert.NoError(t, err)
	assert.Equal(t, block.BlockHeight, retrievedBlock.BlockHeight)
	assert.Equal(t, block.BlockHash, retrievedBlock.BlockHash)
	assert.Equal(t, block.BlockTimestamp, retrievedBlock.BlockTimestamp)
}

func TestGetBlockByHashForNonExistentBlock(t *testing.T) {
	handler, cleanup := setupDB(t)
	defer cleanup()

	block, err := handler.GetBlockByHash("0x123")
	assert.Nil(t, block)
	assert.Equal(t, types.ErrBlockNotFound, err)
}

func TestQueryIsBlockFinalizedByHeight(t *testing.T) {
	handler, cleanup := setupDB(t)
	defer cleanup()

	// Insert a block
	block := &types.Block{
		BlockHeight:    1,
		BlockHash:      "0x123",
		BlockTimestamp: 1000,
	}
	err := handler.InsertBlock(block)
	assert.NoError(t, err)

	// Retrieve block status by height
	isFinalized, err := handler.QueryIsBlockFinalizedByHeight(block.BlockHeight)
	assert.NoError(t, err)
	assert.Equal(t, isFinalized, true)
}

func TestQueryIsBlockFinalizedByHeightForNonExistentBlock(t *testing.T) {
	handler, cleanup := setupDB(t)
	defer cleanup()

	isFinalized, err := handler.QueryIsBlockFinalizedByHeight(1)
	assert.NoError(t, err)
	assert.Equal(t, isFinalized, false)
}

func TestQueryIsBlockFinalizedByHash(t *testing.T) {
	handler, cleanup := setupDB(t)
	defer cleanup()

	// Insert a block
	block := &types.Block{
		BlockHeight:    1,
		BlockHash:      "0x123",
		BlockTimestamp: 1000,
	}
	err := handler.InsertBlock(block)
	assert.NoError(t, err)

	// Retrieve block status by hash
	isFinalized, err := handler.QueryIsBlockFinalizedByHash(block.BlockHash)
	assert.NoError(t, err)
	assert.Equal(t, isFinalized, true)
}

func TestQueryIsBlockFinalizedByHashForNonExistentBlock(t *testing.T) {
	handler, cleanup := setupDB(t)
	defer cleanup()

	isFinalized, err := handler.QueryIsBlockFinalizedByHash("0x123")
	assert.NoError(t, err)
	assert.Equal(t, isFinalized, false)
}

func TestQueryEarliestFinalizedBlock(t *testing.T) {
	handler, cleanup := setupDB(t)
	defer cleanup()

	// Insert two blocks
	first := &types.Block{
		BlockHeight:    1,
		BlockHash:      "0x123",
		BlockTimestamp: 1000,
	}
	second := &types.Block{
		BlockHeight:    2,
		BlockHash:      "0x456",
		BlockTimestamp: 1050,
	}
	third := &types.Block{
		BlockHeight:    3,
		BlockHash:      "0x789",
		BlockTimestamp: 1100,
	}
	err := handler.InsertBlock(first)
	assert.NoError(t, err)
	err = handler.InsertBlock(second)
	assert.NoError(t, err)
	err = handler.InsertBlock(third)
	assert.NoError(t, err)

	// Query earliest consecutively finalized block
	earliestBlock, err := handler.QueryEarliestFinalizedBlock()
	assert.NoError(t, err)
	assert.Equal(t, earliestBlock.BlockHeight, first.BlockHeight)
	assert.Equal(t, earliestBlock.BlockHash, first.BlockHash)
	assert.Equal(t, earliestBlock.BlockTimestamp, first.BlockTimestamp)
}

func TestQueryLatestFinalizedBlock(t *testing.T) {
	handler, cleanup := setupDB(t)
	defer cleanup()

	// Insert two blocks
	first := &types.Block{
		BlockHeight:    1,
		BlockHash:      "0x123",
		BlockTimestamp: 1000,
	}
	second := &types.Block{
		BlockHeight:    2,
		BlockHash:      "0x456",
		BlockTimestamp: 1050,
	}
	err := handler.InsertBlock(first)
	assert.NoError(t, err)
	err = handler.InsertBlock(second)
	assert.NoError(t, err)

	// Retrieve latest block
	latestBlock, err := handler.QueryLatestFinalizedBlock()
	assert.NoError(t, err)
	assert.Equal(t, latestBlock.BlockHeight, second.BlockHeight)
	assert.Equal(t, latestBlock.BlockHash, second.BlockHash)
	assert.Equal(t, latestBlock.BlockTimestamp, second.BlockTimestamp)
}

func TestQueryLatestFinalizedBlockNonExistent(t *testing.T) {
	handler, cleanup := setupDB(t)
	defer cleanup()

	latestBlock, err := handler.QueryLatestFinalizedBlock()
	assert.Nil(t, latestBlock)
	assert.NoError(t, err)
}

func TestGetActivatedTimestamp(t *testing.T) {
	handler, cleanup := setupDB(t)
	defer cleanup()

	// Test when timestamp is not set
	timestamp, err := handler.GetActivatedTimestamp()
	assert.Equal(t, uint64(math.MaxUint64), timestamp)
	assert.Equal(t, types.ErrActivatedTimestampNotFound, err)

	// Set timestamp
	expectedTimestamp := uint64(1234567890)
	err = handler.SaveActivatedTimestamp(expectedTimestamp)
	assert.NoError(t, err)

	// Test when timestamp is set
	timestamp, err = handler.GetActivatedTimestamp()
	assert.NoError(t, err)
	assert.Equal(t, expectedTimestamp, timestamp)
}

func TestSaveActivatedTimestamp(t *testing.T) {
	handler, cleanup := setupDB(t)
	defer cleanup()

	// Set timestamp
	expectedTimestamp := uint64(1234567890)
	err := handler.SaveActivatedTimestamp(expectedTimestamp)
	assert.NoError(t, err)

	// Verify timestamp was saved
	timestamp, err := handler.GetActivatedTimestamp()
	assert.NoError(t, err)
	assert.Equal(t, expectedTimestamp, timestamp)
}
