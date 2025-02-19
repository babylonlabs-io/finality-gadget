package finalitygadget

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/babylonlabs-io/finality-gadget/testutil"
	"github.com/babylonlabs-io/finality-gadget/testutil/mocks"
	"github.com/babylonlabs-io/finality-gadget/types"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

// TODO: add `QueryIsBlockBabylonFinalizedFromBabylon` as test fn once removed from interface

func TestFinalityGadgetDisabled(t *testing.T) {
	ctl := gomock.NewController(t)

	// mock CwClient
	mockCwClient := mocks.NewMockICosmWasmClient(ctl)
	mockCwClient.EXPECT().QueryIsEnabled().Return(false, nil).Times(1)

	mockTestFinalityGadget := &FinalityGadget{
		cwClient:  mockCwClient,
		bbnClient: nil,
		btcClient: nil,
	}

	// check QueryIsBlockBabylonFinalized always returns true when finality gadget is not enabled
	res, err := mockTestFinalityGadget.QueryIsBlockBabylonFinalizedFromBabylon(&types.Block{})
	require.NoError(t, err)
	require.True(t, res)
}

func TestQueryIsBlockBabylonFinalized(t *testing.T) {
	blockWithHashUntrimmed := types.Block{
		BlockHash:      "0xd4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3",
		BlockHeight:    123,
		BlockTimestamp: 12345,
	}

	blockWithHashTrimmed := blockWithHashUntrimmed
	blockWithHashTrimmed.BlockHash = strings.TrimPrefix(blockWithHashUntrimmed.BlockHash, "0x")

	const consumerChainID = "consumer-chain-id"
	const BTCHeight = uint32(111)

	testCases := []struct {
		name                    string
		expectedErr             error
		block                   *types.Block
		allFpPks                []string
		fpPowers                map[string]uint64
		votedProviders          []string
		stakingActivationHeight uint32
		expectResult            bool
	}{
		{
			name:                    "0% votes, expects false",
			block:                   &blockWithHashTrimmed,
			allFpPks:                []string{"pk1", "pk2"},
			fpPowers:                map[string]uint64{"pk1": 100, "pk2": 300},
			votedProviders:          []string{},
			stakingActivationHeight: BTCHeight - 1,
			expectResult:            false,
			expectedErr:             nil,
		},
		{
			name:                    "25% votes, expects false",
			block:                   &blockWithHashTrimmed,
			allFpPks:                []string{"pk1", "pk2"},
			fpPowers:                map[string]uint64{"pk1": 100, "pk2": 300},
			votedProviders:          []string{"pk1"},
			stakingActivationHeight: BTCHeight - 1,
			expectResult:            false,
			expectedErr:             nil,
		},
		{
			name:                    "exact 2/3 votes, expects true",
			block:                   &blockWithHashTrimmed,
			allFpPks:                []string{"pk1", "pk2", "pk3"},
			fpPowers:                map[string]uint64{"pk1": 100, "pk2": 100, "pk3": 100},
			votedProviders:          []string{"pk1", "pk2"},
			stakingActivationHeight: BTCHeight - 1,
			expectResult:            true,
			expectedErr:             nil,
		},
		{
			name:                    "75% votes, expects true",
			block:                   &blockWithHashTrimmed,
			allFpPks:                []string{"pk1", "pk2"},
			fpPowers:                map[string]uint64{"pk1": 100, "pk2": 300},
			votedProviders:          []string{"pk2"},
			stakingActivationHeight: BTCHeight - 1,
			expectResult:            true,
			expectedErr:             nil,
		},
		{
			name:                    "100% votes, expects true",
			block:                   &blockWithHashTrimmed,
			allFpPks:                []string{"pk1", "pk2", "pk3"},
			fpPowers:                map[string]uint64{"pk1": 100, "pk2": 100, "pk3": 100},
			votedProviders:          []string{"pk1", "pk2", "pk3"},
			stakingActivationHeight: BTCHeight - 1,
			expectResult:            true,
			expectedErr:             nil,
		},
		{
			name:                    "untrimmed block hash in input params, 75% votes, expects true",
			block:                   &blockWithHashUntrimmed,
			allFpPks:                []string{"pk1", "pk2", "pk3", "pk4"},
			fpPowers:                map[string]uint64{"pk1": 100, "pk2": 100, "pk3": 100, "pk4": 100},
			votedProviders:          []string{"pk1", "pk2", "pk3"},
			stakingActivationHeight: BTCHeight - 1,
			expectResult:            true,
			expectedErr:             nil,
		},
		{
			name:                    "zero voting power, 100% votes, expects false",
			block:                   &blockWithHashUntrimmed,
			allFpPks:                []string{"pk1", "pk2", "pk3"},
			fpPowers:                map[string]uint64{"pk1": 0, "pk2": 0, "pk3": 0},
			votedProviders:          []string{"pk1", "pk2", "pk3"},
			stakingActivationHeight: BTCHeight - 1,
			expectResult:            false,
			expectedErr:             types.ErrNoFpHasVotingPower,
		},
		{
			name:                    "Btc staking not activated, 100% votes, expects false",
			block:                   &blockWithHashUntrimmed,
			allFpPks:                []string{"pk1", "pk2", "pk3"},
			fpPowers:                map[string]uint64{"pk1": 100, "pk2": 100, "pk3": 100},
			votedProviders:          []string{"pk1", "pk2", "pk3"},
			stakingActivationHeight: BTCHeight + 1,
			expectResult:            false,
			expectedErr:             types.ErrBtcStakingNotActivated,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctl := gomock.NewController(t)
			defer ctl.Finish()

			mockCwClient := mocks.NewMockICosmWasmClient(ctl)
			mockCwClient.EXPECT().QueryIsEnabled().Return(true, nil).Times(1)
			mockCwClient.EXPECT().QueryConsumerId().Return(consumerChainID, nil).Times(1)
			mockBTCClient := mocks.NewMockIBitcoinClient(ctl)
			mockBTCClient.EXPECT().
				GetBlockHeightByTimestamp(tc.block.BlockTimestamp).
				Return(BTCHeight, nil).
				Times(1)

			mockBBNClient := mocks.NewMockIBabylonClient(ctl)
			mockBBNClient.EXPECT().
				QueryAllFpBtcPubKeys(consumerChainID).
				Return(tc.allFpPks, nil).
				Times(1)
			mockBBNClient.EXPECT().
				QueryEarliestActiveDelBtcHeight(tc.allFpPks).
				Return(tc.stakingActivationHeight, nil).
				Times(1)

			if !errors.Is(tc.expectedErr, types.ErrBtcStakingNotActivated) {
				mockBBNClient.EXPECT().
					QueryMultiFpPower(tc.allFpPks, BTCHeight).
					Return(tc.fpPowers, nil).
					Times(1)

				if !errors.Is(tc.expectedErr, types.ErrNoFpHasVotingPower) {
					mockCwClient.EXPECT().
						QueryListOfVotedFinalityProviders(&blockWithHashTrimmed).
						Return(tc.votedProviders, tc.expectedErr).
						Times(1)
				}
			}

			mockFinalityGadget := &FinalityGadget{
				cwClient:  mockCwClient,
				bbnClient: mockBBNClient,
				btcClient: mockBTCClient,
			}

			res, err := mockFinalityGadget.QueryIsBlockBabylonFinalizedFromBabylon(tc.block)
			require.Equal(t, tc.expectResult, res)
			require.Equal(t, tc.expectedErr, err)
		})
	}
}

func TestQueryBlockRangeBabylonFinalized(t *testing.T) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	l2BlockTime := uint64(2)
	blockA, _ := testutil.RandomL2Block(rng)
	blockB, _ := testutil.GenL2Block(rng, &blockA, l2BlockTime, 1)
	blockC, _ := testutil.GenL2Block(rng, &blockB, l2BlockTime, 1)
	blockD, _ := testutil.GenL2Block(rng, &blockC, l2BlockTime, 300)
	blockE, _ := testutil.GenL2Block(rng, &blockD, l2BlockTime, 1)

	testCases := []struct {
		expErr      error
		expRes      *uint64
		name        string
		queryBlocks []*types.Block
		queryDB     bool
	}{
		{
			name:        "empty query blocks",
			queryBlocks: []*types.Block{},
			expErr:      fmt.Errorf("no blocks provided"),
			expRes:      nil,
			queryDB:     false,
		},
		{
			name:        "single block with finalized",
			queryBlocks: []*types.Block{&blockA},
			expErr:      nil,
			expRes:      &blockA.BlockHeight,
			queryDB:     true,
		},
		{
			name:        "single block with error",
			queryBlocks: []*types.Block{&blockD},
			expErr:      nil,
			expRes:      nil,
			queryDB:     true,
		},
		{
			name:        "non-consecutive blocks",
			queryBlocks: []*types.Block{&blockA, &blockD},
			expErr:      fmt.Errorf("blocks are not consecutive"),
			expRes:      nil,
			queryDB:     false,
		},
		{
			name:        "all consecutive blocks are finalized",
			queryBlocks: []*types.Block{&blockA, &blockB},
			expErr:      nil,
			expRes:      &blockB.BlockHeight,
			queryDB:     true,
		},
		{
			name:        "first two blocks finalized, third has error",
			queryBlocks: []*types.Block{&blockA, &blockB, &blockC},
			expErr:      nil,
			expRes:      nil,
			queryDB:     true,
		},
		{
			name:        "none of the blocks are finalized",
			queryBlocks: []*types.Block{&blockD, &blockE},
			expErr:      nil,
			expRes:      nil,
			queryDB:     true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctl := gomock.NewController(t)
			defer ctl.Finish()

			mockDbHandler := mocks.NewMockIDatabaseHandler(ctl)

			// Setup mock DB responses
			if len(tc.queryBlocks) > 0 && tc.queryDB {
				mockDbHandler.EXPECT().
					QueryEarliestFinalizedBlock().
					Return(&blockA, nil).
					Times(1)
				mockDbHandler.EXPECT().
					QueryLatestFinalizedBlock().
					Return(&blockB, nil).
					Times(1)
			}

			mockFinalityGadget := &FinalityGadget{
				db: mockDbHandler,
			}

			res, err := mockFinalityGadget.QueryBlockRangeBabylonFinalized(tc.queryBlocks)
			require.Equal(t, tc.expRes, res)
			if tc.expErr != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expErr.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestInsertBlock(t *testing.T) {
	block := &types.Block{
		BlockHeight:    1,
		BlockHash:      "0x123",
		BlockTimestamp: 1000,
	}

	// mock db and finality gadget
	ctl := gomock.NewController(t)
	mockDbHandler := mocks.NewMockIDatabaseHandler(ctl)
	// note: the block hash is normalized before passing to the db handler
	blocks := []*types.Block{normalizedBlock(block)}
	mockDbHandler.EXPECT().InsertBlocks(blocks).Return(nil).Times(1)
	mockDbHandler.EXPECT().GetBlockByHeight(block.BlockHeight).Return(block, nil).Times(1)

	mockFinalityGadget := &FinalityGadget{
		db: mockDbHandler,
	}

	// insert block
	err := mockFinalityGadget.insertBlocks(blocks)
	require.NoError(t, err)

	// verify block was inserted
	retrievedBlock, err := mockFinalityGadget.GetBlockByHeight(1)
	require.NoError(t, err)
	require.Equal(t, block.BlockHeight, retrievedBlock.BlockHeight)
	require.Equal(t, block.BlockHash, retrievedBlock.BlockHash)
	require.Equal(t, block.BlockTimestamp, retrievedBlock.BlockTimestamp)
}

func TestBatchInsertBlocks(t *testing.T) {
	// Setup mock controller
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	// Create a larger set of test blocks
	numBlocks := 25 // Testing with 25 blocks
	blocks := make([]*types.Block, numBlocks)
	for i := 0; i < numBlocks; i++ {
		blocks[i] = &types.Block{
			BlockHeight:    uint64(i + 1),
			BlockHash:      fmt.Sprintf("0x%x", i+1000), // unique hash for each block
			BlockTimestamp: uint64(1000 + i*100),        // increasing timestamps
		}
	}

	// Create normalized versions of the blocks
	normalizedBlocks := make([]*types.Block, len(blocks))
	for i, block := range blocks {
		normalizedBlocks[i] = &types.Block{
			BlockHeight:    block.BlockHeight,
			BlockHash:      normalizeBlockHash(block.BlockHash),
			BlockTimestamp: block.BlockTimestamp,
		}
	}

	testCases := []struct {
		name      string
		blocks    []*types.Block
		batchSize uint64
	}{
		{
			name:      "small batch size",
			batchSize: 5,
			blocks:    blocks,
		},
		{
			name:      "medium batch size",
			batchSize: 10,
			blocks:    blocks,
		},
		{
			name:      "large batch size",
			batchSize: 20,
			blocks:    blocks,
		},
		{
			name:      "batch size larger than number of blocks",
			batchSize: 30,
			blocks:    blocks,
		},
		{
			name:      "single block batch",
			batchSize: 1,
			blocks:    blocks[:1], // Test with just one block
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup mock database handler
			mockDbHandler := mocks.NewMockIDatabaseHandler(ctl)

			// Create normalized versions specific to this test case
			tcNormalizedBlocks := make([]*types.Block, len(tc.blocks))
			for i, block := range tc.blocks {
				tcNormalizedBlocks[i] = &types.Block{
					BlockHeight:    block.BlockHeight,
					BlockHash:      normalizeBlockHash(block.BlockHash),
					BlockTimestamp: block.BlockTimestamp,
				}
			}

			// Expect batch insert call with normalized blocks
			if len(tc.blocks) > 0 {
				mockDbHandler.EXPECT().InsertBlocks(tcNormalizedBlocks).Return(nil).Times(1)
			}

			// Setup verification calls for each block
			for i, block := range tc.blocks {
				mockDbHandler.EXPECT().
					GetBlockByHeight(block.BlockHeight).
					Return(tcNormalizedBlocks[i], nil).
					Times(1)
			}

			// Create finality gadget instance with mock DB and specified batch size
			mockFinalityGadget := &FinalityGadget{
				db:        mockDbHandler,
				batchSize: tc.batchSize,
			}

			// Test batch insert
			err := mockFinalityGadget.insertBlocks(tc.blocks)
			require.NoError(t, err)

			// Verify each block was inserted correctly
			for _, block := range tc.blocks {
				retrievedBlock, err := mockFinalityGadget.GetBlockByHeight(block.BlockHeight)
				require.NoError(t, err)
				require.NotNil(t, retrievedBlock)
				require.Equal(t, block.BlockHeight, retrievedBlock.BlockHeight)
				require.Equal(t, normalizeBlockHash(block.BlockHash), retrievedBlock.BlockHash)
				require.Equal(t, block.BlockTimestamp, retrievedBlock.BlockTimestamp)
			}
		})
	}
}

func TestGetBlockByHeight(t *testing.T) {
	block := &types.Block{
		BlockHeight:    1,
		BlockHash:      "0x123",
		BlockTimestamp: 1000,
	}

	// mock db and finality gadget
	ctl := gomock.NewController(t)
	mockDbHandler := mocks.NewMockIDatabaseHandler(ctl)
	// note: the block hash is normalized before passing to the db handler
	blocks := []*types.Block{normalizedBlock(block)}
	mockDbHandler.EXPECT().InsertBlocks(blocks).Return(nil).Times(1)
	mockDbHandler.EXPECT().GetBlockByHeight(block.BlockHeight).Return(block, nil).Times(1)

	mockFinalityGadget := &FinalityGadget{
		db: mockDbHandler,
	}

	// insert block
	err := mockFinalityGadget.insertBlocks(blocks)
	require.NoError(t, err)

	// fetch block by height
	retrievedBlock, err := mockFinalityGadget.GetBlockByHeight(1)
	require.NoError(t, err)
	require.Equal(t, block.BlockHeight, retrievedBlock.BlockHeight)
	require.Equal(t, block.BlockHash, retrievedBlock.BlockHash)
	require.Equal(t, block.BlockTimestamp, retrievedBlock.BlockTimestamp)
}

func TestGetBlockByHeightForNonExistentBlock(t *testing.T) {
	// mock db and finality gadget
	ctl := gomock.NewController(t)
	mockDbHandler := mocks.NewMockIDatabaseHandler(ctl)
	mockDbHandler.EXPECT().GetBlockByHeight(uint64(1)).Return(nil, types.ErrBlockNotFound).Times(1)

	mockFinalityGadget := &FinalityGadget{
		db: mockDbHandler,
	}

	// fetch block by height
	retrievedBlock, err := mockFinalityGadget.GetBlockByHeight(1)
	require.Nil(t, retrievedBlock)
	require.Equal(t, err, types.ErrBlockNotFound)
}

func TestGetBlockByHashWith0xPrefix(t *testing.T) {
	block := &types.Block{
		BlockHeight:    1,
		BlockHash:      "0x123",
		BlockTimestamp: 1000,
	}

	// mock db and finality gadget
	ctl := gomock.NewController(t)
	mockDbHandler := mocks.NewMockIDatabaseHandler(ctl)
	// note: the block hash is normalized before passing to the db handler
	blocks := []*types.Block{normalizedBlock(block)}
	mockDbHandler.EXPECT().InsertBlocks(blocks).Return(nil).Times(1)
	mockDbHandler.EXPECT().GetBlockByHash(normalizeBlockHash(block.BlockHash)).Return(block, nil).Times(2)

	mockFinalityGadget := &FinalityGadget{
		db: mockDbHandler,
	}

	// insert block
	err := mockFinalityGadget.insertBlocks(blocks)
	require.NoError(t, err)

	// fetch block by hash including 0x prefix
	retrievedBlock, err := mockFinalityGadget.GetBlockByHash(block.BlockHash)
	require.NoError(t, err)
	require.Equal(t, block.BlockHeight, retrievedBlock.BlockHeight)
	require.Equal(t, block.BlockHash, retrievedBlock.BlockHash)
	require.Equal(t, block.BlockTimestamp, retrievedBlock.BlockTimestamp)

	// fetch block by hash excluding 0x prefix
	retrievedBlock, err = mockFinalityGadget.GetBlockByHash(strings.TrimPrefix(block.BlockHash, "0x"))
	require.NoError(t, err)
	require.Equal(t, block.BlockHeight, retrievedBlock.BlockHeight)
	require.Equal(t, block.BlockHash, retrievedBlock.BlockHash)
	require.Equal(t, block.BlockTimestamp, retrievedBlock.BlockTimestamp)
}

func TestGetBlockByHashWithout0xPrefix(t *testing.T) {
	block := &types.Block{
		BlockHeight:    1,
		BlockHash:      "123",
		BlockTimestamp: 1000,
	}

	// mock db and finality gadget
	ctl := gomock.NewController(t)
	mockDbHandler := mocks.NewMockIDatabaseHandler(ctl)
	// note: the block hash is normalized before passing to the db handler
	blocks := []*types.Block{normalizedBlock(block)}
	mockDbHandler.EXPECT().InsertBlocks(blocks).Return(nil).Times(1)
	mockDbHandler.EXPECT().GetBlockByHash(normalizeBlockHash(block.BlockHash)).Return(block, nil).Times(2)

	mockFinalityGadget := &FinalityGadget{
		db: mockDbHandler,
	}

	// insert block
	err := mockFinalityGadget.insertBlocks(blocks)
	require.NoError(t, err)

	// fetch block by hash including 0x prefix
	retrievedBlock, err := mockFinalityGadget.GetBlockByHash("0x" + block.BlockHash)
	require.NoError(t, err)
	require.Equal(t, block.BlockHeight, retrievedBlock.BlockHeight)
	require.Equal(t, block.BlockHash, retrievedBlock.BlockHash)
	require.Equal(t, block.BlockTimestamp, retrievedBlock.BlockTimestamp)

	// fetch block by hash excluding 0x prefix
	retrievedBlock, err = mockFinalityGadget.GetBlockByHash(block.BlockHash)
	require.NoError(t, err)
	require.Equal(t, block.BlockHeight, retrievedBlock.BlockHeight)
	require.Equal(t, block.BlockHash, retrievedBlock.BlockHash)
	require.Equal(t, block.BlockTimestamp, retrievedBlock.BlockTimestamp)
}

func TestGetBlockByHashForNonExistentBlock(t *testing.T) {
	// mock db and finality gadget
	ctl := gomock.NewController(t)
	mockDbHandler := mocks.NewMockIDatabaseHandler(ctl)
	mockDbHandler.EXPECT().GetBlockByHash(normalizeBlockHash("123")).Return(nil, types.ErrBlockNotFound).Times(1)

	mockFinalityGadget := &FinalityGadget{
		db: mockDbHandler,
	}

	// fetch block by hash
	retrievedBlock, err := mockFinalityGadget.GetBlockByHash("123")
	require.Nil(t, retrievedBlock)
	require.Equal(t, err, types.ErrBlockNotFound)
}

func TestQueryIsBlockFinalizedByHeight(t *testing.T) {
	blockHeight := uint64(1)

	// mock db and finality gadget
	ctl := gomock.NewController(t)
	mockDbHandler := mocks.NewMockIDatabaseHandler(ctl)
	mockDbHandler.EXPECT().QueryIsBlockFinalizedByHeight(blockHeight).Return(true, nil).Times(1)

	mockFinalityGadget := &FinalityGadget{
		db: mockDbHandler,
	}

	// fetch block status by height
	isFinalized, err := mockFinalityGadget.QueryIsBlockFinalizedByHeight(1)
	require.NoError(t, err)
	require.True(t, isFinalized)
}

func TestQueryIsBlockFinalizedByHeightForNonExistentBlock(t *testing.T) {
	// mock db and finality gadget
	ctl := gomock.NewController(t)
	mockDbHandler := mocks.NewMockIDatabaseHandler(ctl)
	mockDbHandler.EXPECT().QueryIsBlockFinalizedByHeight(uint64(1)).Return(false, types.ErrBlockNotFound).Times(1)

	mockFinalityGadget := &FinalityGadget{
		db: mockDbHandler,
	}

	// fetch block status by height
	isFinalized, err := mockFinalityGadget.QueryIsBlockFinalizedByHeight(1)
	require.False(t, isFinalized)
	require.Equal(t, err, types.ErrBlockNotFound)
}

func TestQueryIsBlockFinalizedByHashWith0xPrefix(t *testing.T) {
	// mock db and finality gadget
	ctl := gomock.NewController(t)
	mockDbHandler := mocks.NewMockIDatabaseHandler(ctl)
	mockDbHandler.EXPECT().QueryIsBlockFinalizedByHash(normalizeBlockHash("0x123")).Return(true, nil).Times(2)

	mockFinalityGadget := &FinalityGadget{
		db: mockDbHandler,
	}

	// fetch block status by hash including 0x prefix
	isFinalized, err := mockFinalityGadget.QueryIsBlockFinalizedByHash("0x123")
	require.NoError(t, err)
	require.True(t, isFinalized)

	// fetch block status by hash excluding 0x prefix
	isFinalized, err = mockFinalityGadget.QueryIsBlockFinalizedByHash("123")
	require.NoError(t, err)
	require.True(t, isFinalized)
}

func TestQueryIsBlockFinalizedByHashWithout0xPrefix(t *testing.T) {
	// mock db and finality gadget
	ctl := gomock.NewController(t)
	mockDbHandler := mocks.NewMockIDatabaseHandler(ctl)
	mockDbHandler.EXPECT().QueryIsBlockFinalizedByHash(normalizeBlockHash("123")).Return(true, nil).Times(2)

	mockFinalityGadget := &FinalityGadget{
		db: mockDbHandler,
	}

	// fetch block status by hash including 0x prefix
	isFinalized, err := mockFinalityGadget.QueryIsBlockFinalizedByHash("0x123")
	require.NoError(t, err)
	require.True(t, isFinalized)

	// fetch block status by hash excluding 0x prefix
	isFinalized, err = mockFinalityGadget.QueryIsBlockFinalizedByHash("123")
	require.NoError(t, err)
	require.True(t, isFinalized)
}

func TestQueryIsBlockFinalizedByHashForNonExistentBlock(t *testing.T) {
	// mock db and finality gadget
	ctl := gomock.NewController(t)
	mockDbHandler := mocks.NewMockIDatabaseHandler(ctl)
	mockDbHandler.EXPECT().QueryIsBlockFinalizedByHash(normalizeBlockHash("123")).Return(false, types.ErrBlockNotFound).Times(1)

	mockFinalityGadget := &FinalityGadget{
		db: mockDbHandler,
	}

	// fetch block status by hash
	isFinalized, err := mockFinalityGadget.QueryIsBlockFinalizedByHash("123")
	require.False(t, isFinalized)
	require.Equal(t, err, types.ErrBlockNotFound)
}

func TestQueryLatestFinalizedBlock(t *testing.T) {
	// define blocks
	first := &types.Block{
		BlockHeight:    1,
		BlockHash:      "0x123",
		BlockTimestamp: 1000,
	}
	second := &types.Block{
		BlockHeight:    2,
		BlockHash:      "0x456",
		BlockTimestamp: 2000,
	}
	normalizedFirst := normalizedBlock(first)
	normalizedSecond := normalizedBlock(second)

	// mock db and finality gadget
	ctl := gomock.NewController(t)
	mockDbHandler := mocks.NewMockIDatabaseHandler(ctl)
	blocks := []*types.Block{normalizedFirst, normalizedSecond}
	mockDbHandler.EXPECT().InsertBlocks(blocks).Return(nil).Times(1)
	mockDbHandler.EXPECT().QueryLatestFinalizedBlock().Return(normalizedSecond, nil).Times(1)

	mockFinalityGadget := &FinalityGadget{
		db: mockDbHandler,
	}

	// insert two blocks
	err := mockFinalityGadget.insertBlocks(blocks)
	require.NoError(t, err)

	// fetch latest block
	latestBlock, err := mockFinalityGadget.QueryLatestFinalizedBlock()
	require.NoError(t, err)
	require.Equal(t, normalizedSecond.BlockHeight, latestBlock.BlockHeight)
	require.Equal(t, normalizedSecond.BlockHash, latestBlock.BlockHash)
	require.Equal(t, normalizedSecond.BlockTimestamp, latestBlock.BlockTimestamp)
}

func TestQueryLatestFinalizedBlockForNonExistentBlock(t *testing.T) {
	// mock db and finality gadget
	ctl := gomock.NewController(t)
	mockDbHandler := mocks.NewMockIDatabaseHandler(ctl)
	mockDbHandler.EXPECT().QueryLatestFinalizedBlock().Return(nil, types.ErrBlockNotFound).Times(1)

	mockFinalityGadget := &FinalityGadget{
		db: mockDbHandler,
	}

	// fetch latest block
	latestBlock, err := mockFinalityGadget.QueryLatestFinalizedBlock()
	require.Nil(t, latestBlock)
	require.Equal(t, err, types.ErrBlockNotFound)
}

func TestQueryBtcStakingActivatedTimestamp(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	mockDbHandler := mocks.NewMockIDatabaseHandler(ctl)
	mockCwClient := mocks.NewMockICosmWasmClient(ctl)
	mockBBNClient := mocks.NewMockIBabylonClient(ctl)
	mockBTCClient := mocks.NewMockIBitcoinClient(ctl)

	mockFinalityGadget := &FinalityGadget{
		db:        mockDbHandler,
		cwClient:  mockCwClient,
		bbnClient: mockBBNClient,
		btcClient: mockBTCClient,
		logger:    zap.NewNop(),
	}

	// Test case 1: Timestamp is already in the database
	mockDbHandler.EXPECT().GetActivatedTimestamp().Return(uint64(1234567890), nil)
	timestamp, err := mockFinalityGadget.QueryBtcStakingActivatedTimestamp()
	require.NoError(t, err)
	require.Equal(t, uint64(1234567890), timestamp)

	// Test case 2: Timestamp is not in the database, need to query from bbnClient
	mockDbHandler.EXPECT().GetActivatedTimestamp().Return(uint64(0), types.ErrActivatedTimestampNotFound)
	mockCwClient.EXPECT().QueryConsumerId().Return("consumer-chain-id", nil)
	mockBBNClient.EXPECT().QueryAllFpBtcPubKeys("consumer-chain-id").Return([]string{"pk1", "pk2"}, nil)
	mockBBNClient.EXPECT().QueryEarliestActiveDelBtcHeight([]string{"pk1", "pk2"}).Return(uint32(100), nil)
	mockBTCClient.EXPECT().GetBlockTimestampByHeight(uint32(100)).Return(uint64(1234567890), nil)

	timestamp, err = mockFinalityGadget.QueryBtcStakingActivatedTimestamp()
	require.NoError(t, err)
	require.Equal(t, uint64(1234567890), timestamp)

	// Test case 3: BTC staking is not activated
	mockDbHandler.EXPECT().GetActivatedTimestamp().Return(uint64(0), types.ErrActivatedTimestampNotFound)
	mockCwClient.EXPECT().QueryConsumerId().Return("consumer-chain-id", nil)
	mockBBNClient.EXPECT().QueryAllFpBtcPubKeys("consumer-chain-id").Return([]string{"pk1", "pk2"}, nil)
	mockBBNClient.EXPECT().QueryEarliestActiveDelBtcHeight([]string{"pk1", "pk2"}).Return(uint32(math.MaxUint32), nil)

	timestamp, err = mockFinalityGadget.QueryBtcStakingActivatedTimestamp()
	require.Equal(t, types.ErrBtcStakingNotActivated, err)
	require.Equal(t, uint64(math.MaxUint64), timestamp)
}

func normalizedBlock(block *types.Block) *types.Block {
	return &types.Block{
		BlockHeight:    block.BlockHeight,
		BlockHash:      normalizeBlockHash(block.BlockHash),
		BlockTimestamp: block.BlockTimestamp,
	}
}
