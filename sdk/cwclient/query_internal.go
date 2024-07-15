package cwclient

import (
	"context"
	"encoding/json"
	"time"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	cosmosclient "github.com/cosmos/cosmos-sdk/client"
)

// hardcode the timeout to 20 seconds. We can expose it to the params once needed
const DefaultTimeout = 20 * time.Second

func createBlockVotersQueryData(queryParams *L2Block) ([]byte, error) {
	queryData := ContractQueryMsgs{
		BlockVoters: &blockVotersQuery{
			Height: queryParams.BlockHeight,
			Hash:   queryParams.BlockHash,
		},
	}
	data, err := json.Marshal(queryData)
	if err != nil {
		return nil, err
	}
	return data, nil
}

type contractConfigResponse struct {
	ConsumerId      string `json:"consumer_id"`
	ActivatedHeight uint64 `json:"activated_height"`
}
type ContractQueryMsgs struct {
	Config      *contractConfig   `json:"config,omitempty"`
	BlockVoters *blockVotersQuery `json:"block_voters,omitempty"`
	IsEnabled   *isEnabledQuery   `json:"is_enabled,omitempty"`
}

type blockVotersQuery struct {
	Hash   string `json:"hash"`
	Height uint64 `json:"height"`
}

type isEnabledQuery struct{}

type contractConfig struct{}

func createConfigQueryData() ([]byte, error) {
	queryData := ContractQueryMsgs{
		Config: &contractConfig{},
	}
	data, err := json.Marshal(queryData)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func createIsEnabledQueryData() ([]byte, error) {
	queryData := ContractQueryMsgs{
		IsEnabled: &isEnabledQuery{},
	}
	data, err := json.Marshal(queryData)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// querySmartContractState queries the smart contract state given the contract address and query data
func (cwClient *Client) querySmartContractState(
	queryData []byte,
) (*wasmtypes.QuerySmartContractStateResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancel()

	sdkClientCtx := cosmosclient.Context{Client: cwClient.Client}
	wasmQueryClient := wasmtypes.NewQueryClient(sdkClientCtx)

	req := &wasmtypes.QuerySmartContractStateRequest{
		Address:   cwClient.contractAddr,
		QueryData: queryData,
	}
	return wasmQueryClient.SmartContractState(ctx, req)
}
