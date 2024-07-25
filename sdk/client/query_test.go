package client

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/babylonchain/babylon-finality-gadget/sdk/cwclient"
	"github.com/babylonchain/babylon-finality-gadget/testutil/mocks"
	"github.com/babylonchain/babylon-finality-gadget/testutils"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestFinalityGadgetDisabled(t *testing.T) {
	ctl := gomock.NewController(t)

	// mock CwClient
	mockCwClient := mocks.NewMockICosmWasmClient(ctl)
	mockCwClient.EXPECT().QueryIsEnabled().Return(false, nil).Times(1)

	mockSdkClient := &SdkClient{
		cwClient:  mockCwClient,
		bbnClient: nil,
		btcClient: nil,
	}

	// check QueryIsBlockBabylonFinalized always returns true when finality gadget is not enabled
	res, err := mockSdkClient.QueryIsBlockBabylonFinalized(cwclient.L2Block{})
	require.NoError(t, err)
	require.True(t, res)
}

func TestQueryIsBlockBabylonFinalized(t *testing.T) {
	blockWithHashUntrimmed := cwclient.L2Block{
		BlockHash:      "0xd4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3",
		BlockHeight:    123,
		BlockTimestamp: 12345,
	}

	blockWithHashTrimmed := blockWithHashUntrimmed
	blockWithHashTrimmed.BlockHash = strings.TrimPrefix(blockWithHashUntrimmed.BlockHash, "0x")

	const consumerChainID = "consumer-chain-id"
	const BTCHeight = uint64(111)

	testCases := []struct {
		name           string
		expectedErr    error
		queryParams    *cwclient.L2Block
		allFpPks       []string
		fpPowers       map[string]uint64
		votedProviders []string
		expectResult   bool
	}{
		{
			name:           "0% votes, expects false",
			queryParams:    &blockWithHashTrimmed,
			allFpPks:       []string{"pk1", "pk2"},
			fpPowers:       map[string]uint64{"pk1": 100, "pk2": 300},
			votedProviders: []string{},
			expectResult:   false,
			expectedErr:    nil,
		},
		{
			name:           "25% votes, expects false",
			queryParams:    &blockWithHashTrimmed,
			allFpPks:       []string{"pk1", "pk2"},
			fpPowers:       map[string]uint64{"pk1": 100, "pk2": 300},
			votedProviders: []string{"pk1"},
			expectResult:   false,
			expectedErr:    nil,
		},
		{
			name:           "exact 2/3 votes, expects true",
			queryParams:    &blockWithHashTrimmed,
			allFpPks:       []string{"pk1", "pk2", "pk3"},
			fpPowers:       map[string]uint64{"pk1": 100, "pk2": 100, "pk3": 100},
			votedProviders: []string{"pk1", "pk2"},
			expectResult:   true,
			expectedErr:    nil,
		},
		{
			name:           "75% votes, expects true",
			queryParams:    &blockWithHashTrimmed,
			allFpPks:       []string{"pk1", "pk2"},
			fpPowers:       map[string]uint64{"pk1": 100, "pk2": 300},
			votedProviders: []string{"pk2"},
			expectResult:   true,
			expectedErr:    nil,
		},
		{
			name:           "100% votes, expects true",
			queryParams:    &blockWithHashTrimmed,
			allFpPks:       []string{"pk1", "pk2", "pk3"},
			fpPowers:       map[string]uint64{"pk1": 100, "pk2": 100, "pk3": 100},
			votedProviders: []string{"pk1", "pk2", "pk3"},
			expectResult:   true,
			expectedErr:    nil,
		},
		{
			name:           "untrimmed block hash in input params, 75% votes, expects true",
			queryParams:    &blockWithHashUntrimmed,
			allFpPks:       []string{"pk1", "pk2", "pk3", "pk4"},
			fpPowers:       map[string]uint64{"pk1": 100, "pk2": 100, "pk3": 100, "pk4": 100},
			votedProviders: []string{"pk1", "pk2", "pk3"},
			expectResult:   true,
		},
		{
			name:           "zero voting power, 100% votes, expects false",
			queryParams:    &blockWithHashUntrimmed,
			allFpPks:       []string{"pk1", "pk2", "pk3"},
			fpPowers:       map[string]uint64{"pk1": 0, "pk2": 0, "pk3": 0},
			votedProviders: []string{"pk1", "pk2", "pk3"},
			expectResult:   false,
			expectedErr:    ErrNoFpHasVotingPower,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctl := gomock.NewController(t)
			defer ctl.Finish()

			mockCwClient := mocks.NewMockICosmWasmClient(ctl)
			mockCwClient.EXPECT().QueryIsEnabled().Return(true, nil).Times(1)
			mockCwClient.EXPECT().QueryConsumerId().Return(consumerChainID, nil).Times(1)
			if tc.expectedErr != ErrNoFpHasVotingPower {
				mockCwClient.EXPECT().
					QueryListOfVotedFinalityProviders(&blockWithHashTrimmed).
					Return(tc.votedProviders, nil).
					Times(1)
			}

			mockBTCClient := mocks.NewMockIBitcoinClient(ctl)
			mockBTCClient.EXPECT().
				GetBlockHeightByTimestamp(tc.queryParams.BlockTimestamp).
				Return(BTCHeight, nil).
				Times(1)

			mockBBNClient := mocks.NewMockIBabylonClient(ctl)
			mockBBNClient.EXPECT().
				QueryAllFpBtcPubKeys(consumerChainID).
				Return(tc.allFpPks, nil).
				Times(1)
			mockBBNClient.EXPECT().
				QueryMultiFpPower(tc.allFpPks, BTCHeight).
				Return(tc.fpPowers, nil).
				Times(1)

			mockSdkClient := &SdkClient{
				cwClient:  mockCwClient,
				bbnClient: mockBBNClient,
				btcClient: mockBTCClient,
			}

			res, err := mockSdkClient.QueryIsBlockBabylonFinalized(*tc.queryParams)
			require.Equal(t, tc.expectResult, res)
			require.Equal(t, tc.expectedErr, err)
		})
	}
}

func TestQueryBlockRangeBabylonFinalized(t *testing.T) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	l2BlockTime := uint64(2)
	blockA, blockAWithHashTrimmed := testutils.RandomL2Block(rng)
	blockB, blockBWithHashTrimmed := testutils.GenL2Block(rng, &blockA, l2BlockTime, 1)
	blockC, blockCWithHashTrimmed := testutils.GenL2Block(rng, &blockB, l2BlockTime, 1)
	blockD, blockDWithHashTrimmed := testutils.GenL2Block(rng, &blockC, l2BlockTime, 300) // 10 minutes later
	blockE, blockEWithHashTrimmed := testutils.GenL2Block(rng, &blockD, l2BlockTime, 1)
	blockF, blockFWithHashTrimmed := testutils.GenL2Block(rng, &blockE, l2BlockTime, 300)
	blockG, blockGWithHashTrimmed := testutils.GenL2Block(rng, &blockF, l2BlockTime, 1)

	testCases := []struct {
		name         string
		expectedErr  error
		expectResult *uint64
		queryBlocks  []*cwclient.L2Block
	}{
		{"empty query blocks", fmt.Errorf("no blocks provided"), nil, []*cwclient.L2Block{}},
		{"single block with finalized", nil, &blockA.BlockHeight, []*cwclient.L2Block{&blockA}},
		{"single block with error", fmt.Errorf("RPC rate limit error"), nil, []*cwclient.L2Block{&blockD}},
		{"non-consecutive blocks", fmt.Errorf("blocks are not consecutive"), nil, []*cwclient.L2Block{&blockA, &blockD}},
		{"the first two blocks are finalized and the last block has error", fmt.Errorf("RPC rate limit error"), &blockB.BlockHeight, []*cwclient.L2Block{&blockA, &blockB, &blockC}},
		{"all consecutive blocks are finalized", nil, &blockB.BlockHeight, []*cwclient.L2Block{&blockA, &blockB}},
		{"none of the block is finalized and the first block has error", fmt.Errorf("RPC rate limit error"), nil, []*cwclient.L2Block{&blockD, &blockE}},
		{"none of the block is finalized and the second block has error", nil, nil, []*cwclient.L2Block{&blockF, &blockG}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctl := gomock.NewController(t)
			defer ctl.Finish()

			mockCwClient := mocks.NewMockICosmWasmClient(ctl)
			mockBTCClient := mocks.NewMockIBitcoinClient(ctl)
			mockBBNClient := mocks.NewMockIBabylonClient(ctl)
			mockSdkClient := &SdkClient{
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

			mockBBNClient.EXPECT().QueryAllFpBtcPubKeys("consumer-chain-id").Return([]string{"pk1", "pk2", "pk3"}, nil).AnyTimes()
			mockBBNClient.EXPECT().QueryMultiFpPower([]string{"pk1", "pk2", "pk3"}, gomock.Any()).Return(map[string]uint64{"pk1": 100, "pk2": 200, "pk3": 300}, nil).AnyTimes()

			res, err := mockSdkClient.QueryBlockRangeBabylonFinalized(tc.queryBlocks)
			require.Equal(t, tc.expectResult, res)
			require.Equal(t, tc.expectedErr, err)
		})
	}
}
