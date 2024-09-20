package pg

import (
	"math"
	"os"
	"os/signal"
	"syscall"
	"testing"

	cfg "github.com/babylonlabs-io/finality-gadget/config"
	"github.com/babylonlabs-io/finality-gadget/log"
	"github.com/babylonlabs-io/finality-gadget/types"
	"github.com/stretchr/testify/assert"
)

func setupDB(t *testing.T) (*PostgresHandler, func()) {
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

	// Create a new Postgres handler
	db, err := NewPostgresHandler(&cfg.DBConfig{
		DBName:     "test",
		DBUsername: "test",
		DBPassword: "test",
		DBDataPath: tempFile.Name(),
		DBPort:     5433,
	}, logger)
	if err != nil {
		t.Fatalf("Failed to create PostgresHandler: %v", err)
	}

	// Create initial buckets
	err = db.CreateInitialSchema()
	if err != nil {
		t.Fatalf("Failed to create initial schema: %v", err)
	}

	// Cleanup function to close DB and remove temp file
	cleanup := func() {
		if err := db.Close(); err != nil {
			t.Fatalf("Failed to close DB: %v", err)
		}
		err := os.RemoveAll(tempFile.Name())
		if err != nil {
			t.Fatalf("Failed to delete DB: %v", err)
		}
	}

	// Setup signal handling for cleanup on interrupt
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-signalChan
		cleanup()
		os.Exit(1) // Exit after cleanup
	}()

	return db, cleanup
}

func TestInsertBlock(t *testing.T) {
	db, cleanup := setupDB(t)
	defer cleanup()

	block := &types.Block{
		BlockHeight:    1,
		BlockHash:      "0x123",
		BlockTimestamp: 1000,
	}

	err := db.InsertBlock(block)
	assert.NoError(t, err)

	// Verify block was inserted
	retrievedBlock, err := db.GetBlockByHeight(block.BlockHeight)
	assert.NoError(t, err)
	assert.Equal(t, block.BlockHeight, retrievedBlock.BlockHeight)
	assert.Equal(t, block.BlockHash, retrievedBlock.BlockHash)
	assert.Equal(t, block.BlockTimestamp, retrievedBlock.BlockTimestamp)
}

func TestGetBlockByHeight(t *testing.T) {
	db, cleanup := setupDB(t)
	defer cleanup()

	// Insert a block
	block := &types.Block{
		BlockHeight:    1,
		BlockHash:      "0x123",
		BlockTimestamp: 1000,
	}

	err := db.InsertBlock(block)
	assert.NoError(t, err)

	// Retrieve block by height
	retrievedBlock, err := db.GetBlockByHeight(block.BlockHeight)
	assert.NoError(t, err)
	assert.Equal(t, block.BlockHeight, retrievedBlock.BlockHeight)
	assert.Equal(t, block.BlockHash, retrievedBlock.BlockHash)
	assert.Equal(t, block.BlockTimestamp, retrievedBlock.BlockTimestamp)
}

func TestGetBlockByHeightForNonExistentBlock(t *testing.T) {
	db, cleanup := setupDB(t)
	defer cleanup()

	block, err := db.GetBlockByHeight(1)
	assert.Nil(t, block)
	assert.Equal(t, types.ErrBlockNotFound, err)
}

func TestGetBlockByHash(t *testing.T) {
	db, cleanup := setupDB(t)
	defer cleanup()

	// Insert a block
	block := &types.Block{
		BlockHeight:    1,
		BlockHash:      "0x123",
		BlockTimestamp: 1000,
	}

	err := db.InsertBlock(block)
	assert.NoError(t, err)

	// Retrieve block by hash
	retrievedBlock, err := db.GetBlockByHash(block.BlockHash)
	assert.NoError(t, err)
	assert.Equal(t, block.BlockHeight, retrievedBlock.BlockHeight)
	assert.Equal(t, block.BlockHash, retrievedBlock.BlockHash)
}

func TestGetBlockByHashForNonExistentBlock(t *testing.T) {
	db, cleanup := setupDB(t)
	defer cleanup()

	block, err := db.GetBlockByHash("0x123")
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
	assert.Equal(t, types.ErrBlockNotFound, err)
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
