package sdk

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// this uses a stub contract deployed on Osmosis testnet
// TODO: replace with one deployed on Babylon chain
func checkBlockFinalized(height uint64, hash string) (bool, error) {
	return QueryIsBlockBabylonFinalized(QueryParams{
		ChainType:      0,
		ContractAddr:   "osmo1zck32had0fpc4fu34ae58zvs3mjd5yrzs70thw027nfqst7edc3sdqak0m",
		BlockHeight:    height,
		BlockHash:      hash,
		BlockTimestamp: uint64(1718332131),
	})
}

// TestSdk is an e2e test to call into a stub CosmWasm contract deployed on Osmosis testnet
func TestSdk(t *testing.T) {
	blockHashWithoutEnoughVotes := "0x3aa074144a25d3ed71c7353a20c579650e0c56a993444c6156d44bb90b932f0d"
	blockHashWithEnoughVotes := "stub hash"

	// When the block hash has enoguh votes
	for i, expected := range []bool{true, true, true, false} {
		finaliezd, err := checkBlockFinalized(uint64(i), blockHashWithEnoughVotes)
		require.Nil(t, err)
		require.Equal(t, expected, finaliezd)
	}

	// When the block hash doesn't have enoguh votes
	for i := range 4 {
		finaliezd, err := checkBlockFinalized(uint64(i), blockHashWithoutEnoughVotes)
		require.Nil(t, err)
		require.False(t, finaliezd)
	}
}
