package cwclient

import (
	"encoding/json"

	"github.com/babylonlabs-io/finality-gadget/types"
	rpcclient "github.com/cometbft/cometbft/rpc/client"
)

type CosmWasmClient struct {
	rpcclient.Client
	contractAddr string
}

var _ types.ICosmWasmClient = &CosmWasmClient{}

func NewClient(rpcClient rpcclient.Client, contractAddr string) *CosmWasmClient {
	return &CosmWasmClient{
		Client:       rpcClient,
		contractAddr: contractAddr,
	}
}

func (cwClient *CosmWasmClient) QueryListOfVotedFinalityProviders(
	queryParams *types.Block,
) ([]string, error) {
	queryData, err := createBlockVotersQueryData(queryParams)
	if err != nil {
		return nil, err
	}

	resp, err := cwClient.querySmartContractState(queryData)
	if err != nil {
		return nil, err
	}

	votedFpPkHexList := &[]string{}
	if err := json.Unmarshal(resp.Data, votedFpPkHexList); err != nil {
		return nil, err
	}

	return *votedFpPkHexList, nil
}

func (cwClient *CosmWasmClient) QueryConsumerId() (string, error) {
	queryData, err := createConfigQueryData()
	if err != nil {
		return "", err
	}

	resp, err := cwClient.querySmartContractState(queryData)
	if err != nil {
		return "", err
	}

	var data contractConfigResponse
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return "", err
	}

	return data.ConsumerId, nil
}

func (cwClient *CosmWasmClient) QueryIsEnabled() (bool, error) {
	queryData, err := createIsEnabledQueryData()
	if err != nil {
		return false, err
	}

	resp, err := cwClient.querySmartContractState(queryData)
	if err != nil {
		return false, err
	}

	var isEnabled bool
	if err := json.Unmarshal(resp.Data, &isEnabled); err != nil {
		return false, err
	}

	return isEnabled, nil
}
