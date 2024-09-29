package btcclient

import (
	"math"
	"testing"

	"github.com/babylonlabs-io/finality-gadget/log"
	"github.com/stretchr/testify/require"
)

// TODO: 1) not rely on mainnet RPC; 2) add more tests for some other edge cases
func TestBtcClient(t *testing.T) {
	var blockHeight uint64
	var err error

	// Create logger.
	logger, err := log.NewRootLogger("console", true)
	require.Nil(t, err)

	// Create BTC client
	btcConfig := DefaultBTCConfig()
	btc, err := NewBitcoinClient(btcConfig, logger)
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
	require.Equal(t, uint64(math.MaxUint64), blockHeight)
}
