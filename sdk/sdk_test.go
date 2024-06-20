package sdk

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func newE2eClientWithStubContract() *babylonQueryClient {
	stubContractConfig := Config{
		ChainType:    0,
		ContractAddr: "bbn17p9rzwnnfxcjp32un9ug7yhhzgtkhvl9jfksztgw5uh69wac2pgs6spw0g",
	}
	client, err := NewClient(stubContractConfig)
	if err != nil {
		panic(err)
	}
	return client
}

func TestQueryConsumerId(t *testing.T) {
	client := newE2eClientWithStubContract()
	chainId, err := client.queryConsumerId()
	require.Nil(t, err)
	require.Equal(t, "op-stack-l2-12345", chainId)
}

// this uses a stub contract deployed on Osmosis testnet
// TODO: replace with one deployed on Babylon chain
func checkBlockFinalized(client *babylonQueryClient, height uint64, hash string) (bool, error) {
	return client.QueryIsBlockBabylonFinalized(QueryParams{
		BlockHeight:    height,
		BlockHash:      hash,
		BlockTimestamp: uint64(1718332131),
	})
}

// TestSdk is an e2e test to call into a stub CosmWasm contract deployed on Osmosis testnet
func TestSdk(t *testing.T) {
	client := newE2eClientWithStubContract()

	blockHash := "0x3aa074144a25d3ed71c7353a20c579650e0c56a993444c6156d44bb90b932f0d"
	blockHashForked := "forked hash"

	// When the block hash has enough votes
	for i, expected := range []bool{true, true, true, false} {
		finaliezd, err := checkBlockFinalized(client, uint64(i), blockHash)
		require.Nil(t, err)
		require.Equal(t, expected, finaliezd)
	}

	// When the block hash doesn't have enough votes
	for i := range 4 {
		finaliezd, err := checkBlockFinalized(client, uint64(i), blockHashForked)
		require.Nil(t, err)
		require.False(t, finaliezd)
	}
}
