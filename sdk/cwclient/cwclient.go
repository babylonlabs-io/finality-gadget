package cwclient

import (
	"encoding/json"

	rpcclient "github.com/cometbft/cometbft/rpc/client"
)

type Client struct {
	rpcclient.Client
	contractAddr string
}

func NewClient(rpcClient rpcclient.Client, contractAddr string) *Client {
	return &Client{
		Client:       rpcClient,
		contractAddr: contractAddr,
	}
}

func (cwClient *Client) QueryListOfVotedFinalityProviders(
	queryParams *L2Block,
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

func (cwClient *Client) QueryConsumerId() (string, error) {
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

func (cwClient *Client) QueryIsEnabled() (bool, error) {
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
