package sdk

import (
	"encoding/json"
	"fmt"
	"time"
)

const (
	BabylonTestnet = 0
	BabylonMainnet = 1
)

type QueryParams struct {
	ChainType    int    `mapstructure:"chain-type"`
	ContractAddr string `mapstructure:"contract-addr"`
	BlockHeight  uint64 `mapstructure:"block-height"`
}

// TODO: replace with babylon RPC
func (queryParams QueryParams) getRpcAddr() (string, error) {
	switch queryParams.ChainType {
	case BabylonTestnet:
		return "https://rpc.testnet.osmosis.zone:443", nil
	case BabylonMainnet:
		return "https://rpc.testnet.osmosis.zone:443", nil
	default:
		return "", fmt.Errorf("unrecognized chain type: %d", queryParams.ChainType)
	}
}

type QueryIsBlockBabylonFinalizedResponseData struct {
	Finalized bool `json:"finalized"`
}

func QueryIsBlockBabylonFinalized(queryParams QueryParams) (bool, error) {
	rpcAddr, err := queryParams.getRpcAddr()
	if err != nil {
		return false, err
	}

	queryClient, _ := newBabylonQueryClient(BabylonQueryConfig{
		RPCAddr: rpcAddr,
		// hardcode the timeout to 20 seconds. We can expose it to the params once needed
		Timeout: 20 * time.Second,
	})

	// TODO: is there a better way to define the json schema?
	queryData := fmt.Sprintf(`{
        "check_block_finalized": {"height": %d}
    }`, queryParams.BlockHeight)
	resp, _ := queryClient.querySmartContractState(queryParams.ContractAddr, queryData)

	var data QueryIsBlockBabylonFinalizedResponseData
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		fmt.Println("Error unmarshaling data:", err)
	}

	return data.Finalized, nil
}
