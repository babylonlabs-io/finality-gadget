package sdk

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
)

/*
 * The e2e tests here use:
 *
 * stub contract code: https://gist.github.com/bap2pecs/9541adb2ba61e7abb481bf03f863435d
 * example stub instance: https://www.seiscan.app/atlantic-2/query?contract=sei18fs8atjcxrsypskpk725q2vr8j76q3xwcfle3w2qlna48acmed0sp30xm8
 */

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
func queryListOfVotedFinalityProvidersHelper(client *babylonQueryClient, height uint64, hash string) ([]string, error) {
	return client.queryListOfVotedFinalityProviders(QueryParams{
		BlockHeight:    height,
		BlockHash:      hash,
		BlockTimestamp: uint64(1718332131),
	})
}

func TestQueryListOfVotedFinalityProviders(t *testing.T) {
	client := newE2eClientWithStubContract()

	blockHash := "0x3aa074144a25d3ed71c7353a20c579650e0c56a993444c6156d44bb90b932f0d"
	blockHashForked := "forked hash"

	// When the block hash has enough votes
	fps, err := queryListOfVotedFinalityProvidersHelper(client, uint64(2), blockHash)
	sort.Strings(fps)
	require.Nil(t, err)
	require.Equal(t, []string{"pk1", "pk2", "pk3"}, fps)

	// When the block hash is a fork
	fps, err = queryListOfVotedFinalityProvidersHelper(client, uint64(2), blockHashForked)
	require.Nil(t, err)
	require.Equal(t, []string{"pk-that-voted-for-a-fork"}, fps)

	// When the block is not voted
	fps, err = queryListOfVotedFinalityProvidersHelper(client, uint64(3), "latest L2 block hash")
	require.Nil(t, err)
	require.Equal(t, []string{}, fps)

}
