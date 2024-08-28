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

func TestFinalityGadgetDisabled(t *testing.T) {
	ctl := gomock.NewController(t)

	// mock CwClient
	mockCwClient := mocks.NewMockICosmWasmClient(ctl)
	mockCwClient.EXPECT().QueryIsEnabled().Return(false, nil).Times(1)

	mockFinalityGadget := &FinalityGadget{
		cwClient:  mockCwClient,
		bbnClient: nil,
		btcClient: nil,
	}

	// check QueryIsBlockBabylonFinalized always returns true when finality gadget is not enabled
	res, err := mockFinalityGadget.QueryIsBlockBabylonFinalized(&types.Block{})
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
	const BTCHeight = uint64(111)

	testCases := []struct {
		name                    string
		expectedErr             error
		block                   *types.Block
		allFpPks                []string
		fpPowers                map[string]uint64
		votedProviders          []string
		stakingActivationHeight uint64
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

			res, err := mockFinalityGadget.QueryIsBlockBabylonFinalized(tc.block)
			require.Equal(t, tc.expectResult, res)
			require.Equal(t, tc.expectedErr, err)
		})
	}
}

func TestQueryBlockRangeBabylonFinalized(t *testing.T) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	l2BlockTime := uint64(2)
	blockA, blockAWithHashTrimmed := testutil.RandomL2Block(rng)
	blockB, blockBWithHashTrimmed := testutil.GenL2Block(rng, &blockA, l2BlockTime, 1)
	blockC, blockCWithHashTrimmed := testutil.GenL2Block(rng, &blockB, l2BlockTime, 1)
	blockD, blockDWithHashTrimmed := testutil.GenL2Block(rng, &blockC, l2BlockTime, 300) // 10 minutes later
	blockE, blockEWithHashTrimmed := testutil.GenL2Block(rng, &blockD, l2BlockTime, 1)
	blockF, blockFWithHashTrimmed := testutil.GenL2Block(rng, &blockE, l2BlockTime, 300)
	blockG, blockGWithHashTrimmed := testutil.GenL2Block(rng, &blockF, l2BlockTime, 1)

	testCases := []struct {
		name         string
		expectedErr  error
		expectResult *uint64
		queryBlocks  []*types.Block
	}{
		{"empty query blocks", fmt.Errorf("no blocks provided"), nil, []*types.Block{}},
		{"single block with finalized", nil, &blockA.BlockHeight, []*types.Block{&blockA}},
		{"single block with error", fmt.Errorf("RPC rate limit error"), nil, []*types.Block{&blockD}},
		{"non-consecutive blocks", fmt.Errorf("blocks are not consecutive"), nil, []*types.Block{&blockA, &blockD}},
		{"the first two blocks are finalized and the last block has error", fmt.Errorf("RPC rate limit error"), &blockB.BlockHeight, []*types.Block{&blockA, &blockB, &blockC}},
		{"all consecutive blocks are finalized", nil, &blockB.BlockHeight, []*types.Block{&blockA, &blockB}},
		{"none of the block is finalized and the first block has error", fmt.Errorf("RPC rate limit error"), nil, []*types.Block{&blockD, &blockE}},
		{"none of the block is finalized and the second block has error", nil, nil, []*types.Block{&blockF, &blockG}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctl := gomock.NewController(t)
			defer ctl.Finish()

			mockCwClient := mocks.NewMockICosmWasmClient(ctl)
			mockBTCClient := mocks.NewMockIBitcoinClient(ctl)
			mockBBNClient := mocks.NewMockIBabylonClient(ctl)
			mockFinalityGadget := &FinalityGadget{
				cwClient:  mockCwClient,
				bbnClient: mockBBNClient,
				btcClient: mockBTCClient,
			}

			mockCwClient.EXPECT().QueryIsEnabled().Return(true, nil).AnyTimes()
			mockCwClient.EXPECT().QueryConsumerId().Return("consumer-chain-id", nil).AnyTimes()
			mockCwClient.EXPECT().QueryListOfVotedFinalityProviders(&blockAWithHashTrimmed).Return([]string{"pk1", "pk2", "pk3"}, nil).AnyTimes()
			mockCwClient.EXPECT().QueryListOfVotedFinalityProviders(&blockBWithHashTrimmed).Return([]string{"pk1", "pk2", "pk3"}, nil).AnyTimes()
			mockCwClient.EXPECT().QueryListOfVotedFinalityProviders(&blockCWithHashTrimmed).Return([]string{"pk1", "pk2", "pk3"}, nil).AnyTimes()
			mockCwClient.EXPECT().QueryListOfVotedFinalityProviders(&blockDWithHashTrimmed).Return([]string{"pk3"}, nil).AnyTimes()
			mockCwClient.EXPECT().QueryListOfVotedFinalityProviders(&blockEWithHashTrimmed).Return([]string{"pk1"}, nil).AnyTimes()
			mockCwClient.EXPECT().QueryListOfVotedFinalityProviders(&blockFWithHashTrimmed).Return([]string{"pk2"}, nil).AnyTimes()
			mockCwClient.EXPECT().QueryListOfVotedFinalityProviders(&blockGWithHashTrimmed).Return([]string{"pk3"}, nil).AnyTimes()

			mockBTCClient.EXPECT().GetBlockHeightByTimestamp(blockA.BlockTimestamp).Return(uint64(111), nil).AnyTimes()
			mockBTCClient.EXPECT().GetBlockHeightByTimestamp(blockB.BlockTimestamp).Return(uint64(111), nil).AnyTimes()
			mockBTCClient.EXPECT().GetBlockHeightByTimestamp(blockC.BlockTimestamp).Return(uint64(111), fmt.Errorf("RPC rate limit error")).AnyTimes()
			mockBTCClient.EXPECT().GetBlockHeightByTimestamp(blockD.BlockTimestamp).Return(uint64(112), fmt.Errorf("RPC rate limit error")).AnyTimes()
			mockBTCClient.EXPECT().GetBlockHeightByTimestamp(blockE.BlockTimestamp).Return(uint64(112), nil).AnyTimes()
			mockBTCClient.EXPECT().GetBlockHeightByTimestamp(blockF.BlockTimestamp).Return(uint64(113), nil).AnyTimes()
			mockBTCClient.EXPECT().GetBlockHeightByTimestamp(blockG.BlockTimestamp).Return(uint64(113), fmt.Errorf("RPC rate limit error")).AnyTimes()

			mockBBNClient.EXPECT().QueryEarliestActiveDelBtcHeight(gomock.Any()).Return(uint64(1), nil).AnyTimes()
			mockBBNClient.EXPECT().QueryAllFpBtcPubKeys("consumer-chain-id").Return([]string{"pk1", "pk2", "pk3"}, nil).AnyTimes()
			mockBBNClient.EXPECT().QueryMultiFpPower([]string{"pk1", "pk2", "pk3"}, gomock.Any()).Return(map[string]uint64{"pk1": 100, "pk2": 200, "pk3": 300}, nil).AnyTimes()

			res, err := mockFinalityGadget.QueryBlockRangeBabylonFinalized(tc.queryBlocks)
			require.Equal(t, tc.expectResult, res)
			require.Equal(t, tc.expectedErr, err)
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
	mockDbHandler.EXPECT().InsertBlock(normalizedBlock(block)).Return(nil).Times(1)
	mockDbHandler.EXPECT().GetBlockByHeight(block.BlockHeight).Return(block, nil).Times(1)

	mockFinalityGadget := &FinalityGadget{
		db: mockDbHandler,
	}

	// insert block
	err := mockFinalityGadget.InsertBlock(block)
	require.NoError(t, err)

	// verify block was inserted
	retrievedBlock, err := mockFinalityGadget.GetBlockByHeight(1)
	require.NoError(t, err)
	require.Equal(t, block.BlockHeight, retrievedBlock.BlockHeight)
	require.Equal(t, block.BlockHash, retrievedBlock.BlockHash)
	require.Equal(t, block.BlockTimestamp, retrievedBlock.BlockTimestamp)
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
	mockDbHandler.EXPECT().InsertBlock(normalizedBlock(block)).Return(nil).Times(1)
	mockDbHandler.EXPECT().GetBlockByHeight(block.BlockHeight).Return(block, nil).Times(1)

	mockFinalityGadget := &FinalityGadget{
		db: mockDbHandler,
	}

	// insert block
	err := mockFinalityGadget.InsertBlock(block)
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
	mockDbHandler.EXPECT().InsertBlock(normalizedBlock(block)).Return(nil).Times(1)
	mockDbHandler.EXPECT().GetBlockByHash(normalizeBlockHash(block.BlockHash)).Return(block, nil).Times(2)

	mockFinalityGadget := &FinalityGadget{
		db: mockDbHandler,
	}

	// insert block
	err := mockFinalityGadget.InsertBlock(block)
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
	mockDbHandler.EXPECT().InsertBlock(normalizedBlock(block)).Return(nil).Times(1)
	mockDbHandler.EXPECT().GetBlockByHash(normalizeBlockHash(block.BlockHash)).Return(block, nil).Times(2)

	mockFinalityGadget := &FinalityGadget{
		db: mockDbHandler,
	}

	// insert block
	err := mockFinalityGadget.InsertBlock(block)
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
	mockDbHandler.EXPECT().InsertBlock(normalizedFirst).Return(nil).Times(1)
	mockDbHandler.EXPECT().InsertBlock(normalizedSecond).Return(nil).Times(1)
	mockDbHandler.EXPECT().QueryLatestFinalizedBlock().Return(normalizedSecond, nil).Times(1)

	mockFinalityGadget := &FinalityGadget{
		db: mockDbHandler,
	}

	// insert two blocks
	err := mockFinalityGadget.InsertBlock(first)
	require.NoError(t, err)
	err = mockFinalityGadget.InsertBlock(second)
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
	mockBBNClient.EXPECT().QueryEarliestActiveDelBtcHeight([]string{"pk1", "pk2"}).Return(uint64(100), nil)
	mockBTCClient.EXPECT().GetBlockTimestampByHeight(uint64(100)).Return(uint64(1234567890), nil)

	timestamp, err = mockFinalityGadget.QueryBtcStakingActivatedTimestamp()
	require.NoError(t, err)
	require.Equal(t, uint64(1234567890), timestamp)

	// Test case 3: BTC staking is not activated
	mockDbHandler.EXPECT().GetActivatedTimestamp().Return(uint64(0), types.ErrActivatedTimestampNotFound)
	mockCwClient.EXPECT().QueryConsumerId().Return("consumer-chain-id", nil)
	mockBBNClient.EXPECT().QueryAllFpBtcPubKeys("consumer-chain-id").Return([]string{"pk1", "pk2"}, nil)
	mockBBNClient.EXPECT().QueryEarliestActiveDelBtcHeight([]string{"pk1", "pk2"}).Return(uint64(math.MaxUint64), nil)

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
