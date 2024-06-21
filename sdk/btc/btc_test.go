package btc

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestBtc is an e2e test to call into a stub CosmWasm contract deployed on Osmosis testnet
// TODO: add more tests for some other edge cases
func TestBtc(t *testing.T) {
	var blockHeight uint64
	var err error

	// timestmap between block 848682 and 848683
	blockHeight, err = GetBlockHeightByTimestamp(uint64(1718840690))
	require.Nil(t, err)
	require.Equal(t, uint64(848682), blockHeight)

	// the exact timestamp of block 848682
	blockHeight, err = GetBlockHeightByTimestamp(uint64(1718839311))
	require.Nil(t, err)
	require.Equal(t, uint64(848682), blockHeight)

	// the exact timestamp minus one of block 848682
	blockHeight, err = GetBlockHeightByTimestamp(uint64(1718839310))
	require.Nil(t, err)
	require.Equal(t, uint64(848681), blockHeight)

	// a timestamp in the future i.e. year 2056
	blockHeight, err = GetBlockHeightByTimestamp(uint64(2718840690))
	require.Nil(t, err)
	require.Equal(t, uint64(0), blockHeight)

	// a timestamp in the past i.e. year 2014
	blockHeight, err = GetBlockHeightByTimestamp(uint64(1418840690))
	require.Nil(t, err)
	require.Equal(t, uint64(0), blockHeight)

}
