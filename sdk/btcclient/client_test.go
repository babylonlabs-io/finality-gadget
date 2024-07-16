package btcclient

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// TestBtc is an e2e test to call into a stub CosmWasm contract deployed on Osmosis testnet
// TODO: add more tests for some other edge cases
func TestBtcClient(t *testing.T) {
	var blockHeight uint64
	var err error

	// Create logger.
	logger, err := zap.NewProduction()
	require.Nil(t, err)

	// Create BTC client
	btcConfig := DefaultBTCConfig()
	btc, err := NewBTCClient(btcConfig, logger)
	require.Nil(t, err)

	// timestmap between block 848682 and 848683
	blockHeight, err = btc.GetBlockHeightByTimestamp(uint64(1718840690))
	require.Nil(t, err)
	require.Equal(t, uint64(848682), blockHeight)

	// the exact timestamp of block 848682
	blockHeight, err = btc.GetBlockHeightByTimestamp(uint64(1718839311))
	require.Nil(t, err)
	require.Equal(t, uint64(848682), blockHeight)

	// the exact timestamp minus one of block 848682
	blockHeight, err = btc.GetBlockHeightByTimestamp(uint64(1718839310))
	require.Nil(t, err)
	require.Equal(t, uint64(848681), blockHeight)

	// a timestamp in the future i.e. year 2056
	blockHeight, err = btc.GetBlockHeightByTimestamp(uint64(2718840690))
	require.Nil(t, err)
	require.Equal(t, uint64(0), blockHeight)
}
