package sdk

import (
	"encoding/json"

	"github.com/babylonchain/babylon-da-sdk/sdk/btc"
	btcstakingtypes "github.com/babylonchain/babylon/x/btcstaking/types"
	sdkquerytypes "github.com/cosmos/cosmos-sdk/types/query"
)

type L2Block struct {
	BlockHeight    uint64 `mapstructure:"block-height"`
	BlockHash      string `mapstructure:"block-hash"`
	BlockTimestamp uint64 `mapstructure:"block-timestamp"`
}

type ContractQueryMsgs struct {
	Config     *contractConfig `json:"config,omitempty"`
	BlockVotes *blockVotes     `json:"block_votes,omitempty"`
	IsEnabled  *isEnabledQuery `json:"is_enabled,omitempty"`
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

type isEnabledQuery struct{}

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

func createBlockVotesQueryData(queryParams *L2Block) ([]byte, error) {
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

func (babylonClient *BabylonQueryClient) queryListOfVotedFinalityProviders(queryParams *L2Block) ([]string, error) {
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

func (babylonClient *BabylonQueryClient) queryFpBtcPubKeys(consumerId string) ([]string, error) {
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

func (babylonClient *BabylonQueryClient) queryConsumerId() (string, error) {
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

func (babylonClient *BabylonQueryClient) queryMultiFpPowerAtHeight(fpPubkeyHexList []string, btcHeight uint64) (map[string]uint64, error) {
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
func (babylonClient *BabylonQueryClient) isDelegationEligible(btcDel *btcstakingtypes.BTCDelegationResponse, btcHeight uint64) (bool, error) {
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

func (babylonClient *BabylonQueryClient) queryFpPower(fpPubkeyHex string, btcHeight uint64) (uint64, error) {
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

func (babylonClient *BabylonQueryClient) QueryIsBlockBabylonFinalized(queryParams *L2Block) (bool, error) {
	// check if the finality gadget is enabled
	// if not, always return true to pass through op derivation pipeline
	isEnabled, err := babylonClient.queryIsEnabled()
	if err != nil {
		return false, err
	}
	if !isEnabled {
		return true, nil
	}

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

	// calculate total voting power
	var totalPower uint64 = 0
	for _, power := range allFpPower {
		totalPower += power
	}

	// no FP has voting power for the consumer chain
	if totalPower == 0 {
		return false, nil
	}

	// get all FPs that voted this (L2 block height, L2 block hash) combination
	votedFpPks, err := babylonClient.queryListOfVotedFinalityProviders(queryParams)
	if err != nil {
		return false, err
	}

	// calculate voted voting power
	var votedPower uint64 = 0
	for _, key := range votedFpPks {
		if power, exists := allFpPower[key]; exists {
			votedPower += power
		}
	}

	// quorom < 2/3
	if votedPower*3 < totalPower*2 {
		return false, nil
	}
	return true, nil
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

func (babylonClient *BabylonQueryClient) queryIsEnabled() (bool, error) {
	queryData, err := createIsEnabledQueryData()
	if err != nil {
		return false, err
	}

	resp, err := babylonClient.querySmartContractState(babylonClient.config.ContractAddr, queryData)
	if err != nil {
		return false, err
	}

	var isEnabled bool
	if err := json.Unmarshal(resp.Data, &isEnabled); err != nil {
		return false, err
	}

	return isEnabled, nil
}
