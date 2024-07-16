package client

import (
	"testing"

	"github.com/babylonchain/babylon-finality-gadget/sdk/cwclient"
	"github.com/babylonchain/babylon-finality-gadget/testutil/mocks"
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
	queryParams := cwclient.L2Block{
		BlockHash:      "d4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3",
		BlockHeight:    123,
		BlockTimestamp: 12345,
	}

	const consumerChainID = "consumer-chain-id"
	const BTCHeight = uint64(111)

	testCases := []struct {
		name           string
		fpPowers       map[string]uint64
		allFpPks       []string
		votedProviders []string
		expectResult   bool
	}{
		{
			name:           "25% votes, expects false",
			allFpPks:       []string{"pk1", "pk2"},
			votedProviders: []string{"pk1"},
			fpPowers:       map[string]uint64{"pk1": 100, "pk2": 300},
			expectResult:   false,
		},
		{
			name:           "75% votes, expects true",
			allFpPks:       []string{"pk1", "pk2"},
			votedProviders: []string{"pk2"},
			fpPowers:       map[string]uint64{"pk1": 100, "pk2": 300},
			expectResult:   true,
		},
		{
			name:           "exact 2/3 votes, expects true",
			allFpPks:       []string{"pk1", "pk2", "pk3"},
			votedProviders: []string{"pk1", "pk2"},
			fpPowers:       map[string]uint64{"pk1": 100, "pk2": 100, "pk3": 100},
			expectResult:   true,
		},
		{
			name:           "everyone has 100 voting power, 3 of them votes, return true",
			allFpPks:       []string{"pk1", "pk2", "pk3", "pk4"},
			votedProviders: []string{"pk1", "pk2", "pk3"},
			fpPowers:       map[string]uint64{"pk1": 100, "pk2": 100, "pk3": 100, "pk4": 100},
			expectResult:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctl := gomock.NewController(t)
			defer ctl.Finish()

			mockCwClient := mocks.NewMockICosmWasmClient(ctl)
			mockCwClient.EXPECT().QueryIsEnabled().Return(true, nil).Times(1)
			mockCwClient.EXPECT().QueryConsumerId().Return(consumerChainID, nil).Times(1)
			mockCwClient.EXPECT().
				QueryListOfVotedFinalityProviders(&queryParams).
				Return(tc.votedProviders, nil).
				Times(1)

			mockBTCClient := mocks.NewMockIBitcoinClient(ctl)
			mockBTCClient.EXPECT().
				GetBlockHeightByTimestamp(queryParams.BlockTimestamp).
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

			res, err := mockSdkClient.QueryIsBlockBabylonFinalized(queryParams)
			require.NoError(t, err)
			require.Equal(t, tc.expectResult, res)
		})
	}
}
