package testutil

import (
	"github.com/babylonchain/babylon-finality-gadget/sdk/btcclient"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockBtcClient is the mock implementation of BtcClient.
type MockBtcClient struct {
	*btcclient.BTCClient
	mock.Mock
}

func NewMockBTCClient(cfg *btcclient.BTCConfig, logger *zap.Logger) (*MockBtcClient, error) {
	innerClient, err := btcclient.NewBTCClient(cfg, logger)

	return &MockBtcClient{
		BTCClient: innerClient,
	}, err
}

// GetBlockHeightByTimestamp overrides the BTCClient's GetBlockHeightByTimestamp method.
func (c *MockBtcClient) GetBlockHeightByTimestamp(targetTimestamp uint64) (uint64, error) {
	// by default, kValue is 6 and wValue is 20
	//
	// in out test, the two mock delegations has:
	// - btcDel.StartHeight: is 1, btcDel.EndHeight: 101
	// - btcDel.StartHeight: is 8, btcDel.EndHeight: 108
	//
	// so to avoid `QueryIsBlockBabylonFinalized` returning `ErrBtcStakingNotActivated` in test,
	// we should return a height that's between [8+6, 101-20]
	//
	// so here we choose a number around the middle
	return 50, nil
}

// this is used to determine when the BTC staking is activated. return 0 to
// simulate that the BTC staking is always activated
func (c *MockBtcClient) GetBlockTimestampByHeight(height uint64) (uint64, error) {
	return 0, nil
}
