package mocks

import (
	"github.com/babylonlabs-io/finality-gadget/btcclient"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockBtcClient is the mock implementation of BtcClient.
type MockBitcoinClient struct {
	*btcclient.BitcoinClient
	mock.Mock
}

func NewMockBitcoinClient(cfg *btcclient.BTCConfig, logger *zap.Logger) (*MockBitcoinClient, error) {
	innerClient, err := btcclient.NewBitcoinClient(cfg, logger)

	return &MockBitcoinClient{
		BitcoinClient: innerClient,
	}, err
}

// GetBlockHeightByTimestamp overrides the BTCClient's GetBlockHeightByTimestamp method.
func (c *MockBitcoinClient) GetBlockHeightByTimestamp(targetTimestamp uint64) (uint64, error) {
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
func (c *MockBitcoinClient) GetBlockTimestampByHeight(height uint64) (uint64, error) {
	return 0, nil
}
