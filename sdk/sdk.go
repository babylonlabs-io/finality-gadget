package sdk

import (
	"encoding/json"
	"errors"

	"github.com/babylonchain/babylon-da-sdk/sdk/btc"
	btcstakingtypes "github.com/babylonchain/babylon/x/btcstaking/types"
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

func (babylonClient *babylonQueryClient) queryMultiFpPowerAtHeight(fpPubkeyHexList []string, btcHeight uint64) (map[string]uint64, error) {
	fpPowerMap := make(map[string]uint64)

	for _, fpPubkeyHex := range fpPubkeyHexList {
		fpPower, err := babylonClient.queryFpPower(fpPubkeyHex, btcHeight)
		if err != nil {
			return nil, err
		}
		fpPowerMap[fpPubkeyHex] = fpPower
	}

	return fpPowerMap, nil
}

// we implemented exact logic as in
// https://github.com/babylonchain/babylon-private/blob/c5a8d317091e2965e20ea56fa10e98d34aaa3547/x/btcstaking/types/btc_delegation.go#L111-L119
func (babylonClient *babylonQueryClient) isDelegationEligible(btcDel *btcstakingtypes.BTCDelegationResponse, btcHeight uint64) (bool, error) {
	btccheckpointParams, err := babylonClient.bbnClient.QueryClient.BTCCheckpointParams()
	if err != nil {
		return false, err
	}
	btcstakingParams, err := babylonClient.bbnClient.QueryClient.BTCStakingParams()
	if err != nil {
		return false, err
	}
	wValue := btccheckpointParams.GetParams().CheckpointFinalizationTimeout
	covQuorum := btcstakingParams.GetParams().CovenantQuorum
	ud := btcDel.UndelegationResponse

	if len(ud.GetDelegatorUnbondingSigHex()) > 0 {
		return false, nil
	}

	if btcHeight < btcDel.StartHeight || btcHeight+wValue > btcDel.EndHeight {
		return false, nil
	}

	if uint32(len(btcDel.CovenantSigs)) < covQuorum {
		return false, nil
	}
	if len(ud.CovenantUnbondingSigList) < int(covQuorum) {
		return false, nil
	}
	if len(ud.CovenantSlashingSigs) < int(covQuorum) {
		return false, nil
	}

	return true, nil
}

func (babylonClient *babylonQueryClient) queryFpPower(fpPubkeyHex string, btcHeight uint64) (uint64, error) {
	totalPower := uint64(0)
	pagination := &sdkquerytypes.PageRequest{}
	// queries the BTCStaking module for all delegations of a finality provider
	resp, err := babylonClient.bbnClient.QueryClient.FinalityProviderDelegations(fpPubkeyHex, pagination)
	if err != nil {
		return 0, err
	}
	for {
		// btcDels contains all the queried BTC delegations
		for _, btcDels := range resp.BtcDelegatorDelegations {
			for _, btcDel := range btcDels.Dels {
				// check the eligbility of delegation
				isEligible, err := babylonClient.isDelegationEligible(btcDel, btcHeight)
				if err != nil {
					return 0, err
				}
				if isEligible {
					totalPower += btcDel.TotalSat
				}
			}
		}
		if resp.Pagination == nil || resp.Pagination.NextKey == nil {
			break
		}
		pagination.Key = resp.Pagination.NextKey
	}

	return totalPower, nil
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
	allFpPower, err := babylonClient.queryMultiFpPowerAtHeight(allFpPks, btcblockHeight)
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
