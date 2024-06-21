package sdk

import (
	"encoding/json"
	"errors"

	"github.com/babylonchain/babylon-da-sdk/sdk/btc"
	sdkquerytypes "github.com/cosmos/cosmos-sdk/types/query"
)

type QueryParams struct {
	BlockHeight    uint64 `mapstructure:"block-height"`
	BlockHash      string `mapstructure:"block-hash"`
	BlockTimestamp uint64 `mapstructure:"block-timestamp"`
}

type ContractQueryMsgs struct {
	Config     *contractConfig `json:"config,omitempty"`
	BlockVotes *blockVotes     `json:"block_votes,omitempty"`
}

type contractConfig struct{}

type contractConfigResponse struct {
	ConsumerId      string `json:"consumer_id"`
	ActivatedHeight uint64 `json:"activated_height"`
}

type blockVotes struct {
	Height uint64 `json:"height"`
	Hash   string `json:"hash"`
}

type blockVotesResponse struct {
	BtcPkHexList []string `json:"fp_pubkey_hex_list"`
}

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

func createBlockVotesQueryData(queryParams QueryParams) ([]byte, error) {
	queryData := ContractQueryMsgs{
		BlockVotes: &blockVotes{
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

func (babylonClient *babylonQueryClient) queryListOfVotedFinalityProviders(queryParams QueryParams) ([]string, error) {
	queryData, err := createBlockVotesQueryData(queryParams)
	if err != nil {
		return nil, err
	}

	resp, err := babylonClient.querySmartContractState(babylonClient.config.ContractAddr, queryData)
	if err != nil {
		return nil, err
	}

	var data blockVotesResponse
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, err
	}

	return data.BtcPkHexList, nil
}

func (babylonClient *babylonQueryClient) queryFpBtcPubKeys(consumerId string) ([]string, error) {
	pagination := &sdkquerytypes.PageRequest{}
	resp, err := babylonClient.bbnClient.QueryClient.QueryConsumerFinalityProviders(consumerId, pagination)
	if err != nil {
		return nil, err
	}

	var pkArr []string

	for _, fp := range resp.FinalityProviders {
		pkArr = append(pkArr, fp.BtcPk.MarshalHex())
	}
	return pkArr, nil
}

func (babylonClient *babylonQueryClient) queryConsumerId() (string, error) {
	queryData, err := createConfigQueryData()
	if err != nil {
		return "", err
	}

	resp, err := babylonClient.querySmartContractState(babylonClient.config.ContractAddr, queryData)
	if err != nil {
		return "", err
	}

	var data contractConfigResponse
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return "", err
	}

	return data.ConsumerId, nil
}

func queryMultiFpPowerAtHeight(fps []string, btcHeight uint64) (map[string]uint64, error) {
	// TODO: implement
	return map[string]uint64{
		"pk1": 12345,
		"pk2": 67890,
		"pk3": 111213,
		"pk4": 141516,
	}, nil
}

func (babylonClient *babylonQueryClient) QueryIsBlockBabylonFinalized(queryParams QueryParams) (bool, error) {
	// get the consumer chain id
	consumerId, err := babylonClient.queryConsumerId()
	if err != nil {
		return false, err
	}

	// get all the FPs pubkey for the consumer chain
	allFpPks, err := babylonClient.queryFpBtcPubKeys(consumerId)
	if err != nil {
		return false, err
	}

	// convert the L2 timestamp to BTC height
	btcblockHeight, err := btc.GetBlockHeightByTimestamp(queryParams.BlockTimestamp)
	if err != nil {
		return false, err
	}

	// get all FPs voting power at this BTC height
	allFpPower, err := queryMultiFpPowerAtHeight(allFpPks, btcblockHeight)
	if err != nil {
		return false, err
	}

	// get all FPs that voted this (L2 block height, L2 block hash) combination
	votedFpPks, err := babylonClient.queryListOfVotedFinalityProviders(queryParams)
	if err != nil {
		return false, err
	}

	// calculate total voting power
	var totalPower uint64 = 0
	for _, power := range allFpPower {
		totalPower += power
	}

	// calculate voted voting power
	var votedPower uint64 = 0
	for _, key := range votedFpPks {
		if power, exists := allFpPower[key]; exists {
			votedPower += power
		}
	}

	quorum := float64(votedPower) / float64(totalPower)
	// TODO: the quorom is hardcode for now. later we can consider make it a config param
	if quorum < 0.667 {
		return false, errors.New("not enough voting power")
	}
	return true, nil
}
