package cwclient

import "github.com/babylonlabs-io/finality-gadget/types"

type ICosmWasmClient interface {
	QueryListOfVotedFinalityProviders(queryParams *types.Block) ([]string, error)
	QueryConsumerId() (string, error)
	QueryIsEnabled() (bool, error)
}
