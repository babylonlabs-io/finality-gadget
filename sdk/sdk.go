package sdk

import (
	"encoding/json"
	"strconv"
)

type QueryParams struct {
	BlockHeight    uint64 `mapstructure:"block-height"`
	BlockHash      string `mapstructure:"block-hash"`
	BlockTimestamp uint64 `mapstructure:"block-timestamp"`
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

func (babylonClient *babylonQueryClient) QueryIsBlockBabylonFinalized(queryParams QueryParams) (bool, error) {
	queryData, err := createQueryData(queryParams)
	if err != nil {
		return false, err
	}

	resp, err := babylonClient.querySmartContractState(babylonClient.config.ContractAddr, queryData)
	if err != nil {
		return false, err
	}

	var data queryIsBlockBabylonFinalizedResponseData
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return false, err
	}

	return data.Finalized, nil
}
