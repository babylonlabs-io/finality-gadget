package sdk

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
)

const (
	BabylonTestnet = 0
	BabylonMainnet = 1
)

type QueryParams struct {
	ChainType      int    `mapstructure:"chain-type"`
	ContractAddr   string `mapstructure:"contract-addr"`
	BlockHeight    uint64 `mapstructure:"block-height"`
	BlockHash      string `mapstructure:"block-hash"`
	BlockTimestamp string `mapstructure:"block-timestamp"`
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

type CheckBlockFinalized struct {
	Height    uint64 `json:"height"`
	Hash      string `json:"hash"`
	Timestamp string `json:"timestamp"`
}

type QueryData struct {
	CheckBlockFinalized CheckBlockFinalized `json:"check_block_finalized"`
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

	queryData := QueryData{
		CheckBlockFinalized: CheckBlockFinalized{
			Height:    queryParams.BlockHeight,
			Hash:      queryParams.BlockHash,
			Timestamp: queryParams.BlockTimestamp,
		},
	}
	jsonData, err := json.Marshal(queryData)
	if err != nil {
		log.Fatalf("Error marshaling JSON: %v", err)
	}
	resp, err := queryClient.querySmartContractState(queryParams.ContractAddr, string(jsonData))
	if err != nil {
		fmt.Println("Query error:", err)
	}

	var data QueryIsBlockBabylonFinalizedResponseData
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		fmt.Println("Error unmarshaling data:", err)
	}

	return data.Finalized, nil
}
