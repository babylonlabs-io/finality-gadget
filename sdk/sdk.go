package sdk

import (
	"encoding/json"

	bsctypes "github.com/babylonchain/babylon/x/btcstkconsumer/types"
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

func (babylonClient *babylonQueryClient) queryFinalityProviders(consumerId string) ([]*bsctypes.FinalityProviderResponse, error) {
	pagination := &sdkquerytypes.PageRequest{}
	resp, err := babylonClient.bbnClient.QueryClient.QueryConsumerFinalityProviders(consumerId, pagination)
	if err != nil {
		return nil, err
	}
	return resp.FinalityProviders, nil
}

func (babylonClient *babylonQueryClient) QueryIsBlockBabylonFinalized(queryParams QueryParams) (bool, error) {
	votedFps, err := babylonClient.queryListOfVotedFinalityProviders(queryParams)
	if err != nil {
		return false, err
	}

	// TODO: change w real implementation
	var consumerId = ""
	_, err = babylonClient.queryFinalityProviders(consumerId)
	if err != nil {
		return false, err
	}

	// TODO: change w real implementation
	// stub contract: https://www.seiscan.app/atlantic-2/query?contract=sei18fs8atjcxrsypskpk725q2vr8j76q3xwcfle3w2qlna48acmed0sp30xm8
	// stub contract code: https://gist.github.com/bap2pecs/9541adb2ba61e7abb481bf03f863435d
	if len(votedFps) < 3 {
		return false, nil
	}

	return true, nil
}
