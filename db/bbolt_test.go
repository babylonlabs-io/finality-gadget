package db

import (
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

func TestGetLatestBlock(t *testing.T) {
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
	latestBlock, err := handler.GetLatestBlock()
	assert.NoError(t, err)
	assert.Equal(t, latestBlock.BlockHeight, second.BlockHeight)
	assert.Equal(t, latestBlock.BlockHash, second.BlockHash)
	assert.Equal(t, latestBlock.BlockTimestamp, second.BlockTimestamp)
}

func TestGetLatestBlockNonExistent(t *testing.T) {
	handler, cleanup := setupDB(t)
	defer cleanup()

	latestBlock, err := handler.GetLatestBlock()
	assert.Nil(t, latestBlock)
	assert.NoError(t, err)
}
