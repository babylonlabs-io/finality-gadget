package sdk

import (
	"encoding/json"
	"fmt"
	"strconv"
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
	BlockTimestamp uint64 `mapstructure:"block-timestamp"`
}

// TODO: replace with babylon RPCs when QuerySmartContractStateRequest query is supported
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

type checkBlockFinalizedData struct {
	Height    uint64 `json:"height"`
	Hash      string `json:"hash"`
	Timestamp string `json:"timestamp"`
}

type queryIsBlockBabylonFinalizedData struct {
	CheckBlockFinalized checkBlockFinalizedData `json:"check_block_finalized"`
}

type queryIsBlockBabylonFinalizedResponseData struct {
	Finalized bool `json:"finalized"`
}

func createQueryData(queryParams QueryParams) ([]byte, error) {
	queryData := queryIsBlockBabylonFinalizedData{
		CheckBlockFinalized: checkBlockFinalizedData{
			Height:    queryParams.BlockHeight,
			Hash:      queryParams.BlockHash,
			Timestamp: strconv.FormatUint(queryParams.BlockTimestamp, 10),
		},
	}
	data, err := json.Marshal(queryData)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func QueryIsBlockBabylonFinalized(queryParams QueryParams) (bool, error) {
	rpcAddr, err := queryParams.getRpcAddr()
	if err != nil {
		return false, err
	}

	queryClient, err := newBabylonQueryClient(babylonQueryConfig{
		rpcAddr: rpcAddr,
		// hardcode the timeout to 20 seconds. We can expose it to the params once needed
		timeout: 20 * time.Second,
	})
	if err != nil {
		return false, err
	}

	queryData, err := createQueryData(queryParams)
	if err != nil {
		return false, err
	}

	resp, err := queryClient.querySmartContractState(queryParams.ContractAddr, queryData)
	if err != nil {
		return false, err
	}

	var data queryIsBlockBabylonFinalizedResponseData
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return false, err
	}

	return data.Finalized, nil
}
