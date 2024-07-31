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
	// has to be a small number so when FP e2e tests use it, the test can finish quickly
	// if it's too large, it will result in unbounding of the delegation
	return 10, nil
}

// this is used to determine when the BTC staking is activated. return 0 to
// simulate that the BTC staking is always activated
func (c *MockBtcClient) GetBlockTimestampByHeight(height uint64) (uint64, error) {
	return 0, nil
}
