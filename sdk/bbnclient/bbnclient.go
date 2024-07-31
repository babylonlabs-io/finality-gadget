package bbnclient

import (
	"math"

	"github.com/babylonchain/babylon/client/query"
	"github.com/babylonchain/babylon/x/btcstaking/types"
	sdkquerytypes "github.com/cosmos/cosmos-sdk/types/query"
)

type Client struct {
	*query.QueryClient
}

func (bbnClient *Client) QueryAllFpBtcPubKeys(consumerId string) ([]string, error) {
	pagination := &sdkquerytypes.PageRequest{}
	resp, err := bbnClient.QueryClient.QueryConsumerFinalityProviders(consumerId, pagination)
	if err != nil {
		return nil, err
	}

	var pkArr []string

	for _, fp := range resp.FinalityProviders {
		pkArr = append(pkArr, fp.BtcPk.MarshalHex())
	}
	return pkArr, nil
}

func (bbnClient *Client) QueryFpPower(fpPubkeyHex string, btcHeight uint64) (uint64, error) {
	totalPower := uint64(0)
	pagination := &sdkquerytypes.PageRequest{}
	// queries the BTCStaking module for all delegations of a finality provider
	resp, err := bbnClient.QueryClient.FinalityProviderDelegations(fpPubkeyHex, pagination)
	if err != nil {
		return 0, err
	}
	for {
		// btcDels contains all the queried BTC delegations
		for _, btcDels := range resp.BtcDelegatorDelegations {
			for _, btcDel := range btcDels.Dels {
				// check whether the delegation is active
				isActive, err := bbnClient.isDelegationActive(btcDel, btcHeight)
				if err != nil {
					return 0, err
				}
				if isActive {
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

func (bbnClient *Client) QueryMultiFpPower(
	fpPubkeyHexList []string,
	btcHeight uint64,
) (map[string]uint64, error) {
	fpPowerMap := make(map[string]uint64)

	for _, fpPubkeyHex := range fpPubkeyHexList {
		fpPower, err := bbnClient.QueryFpPower(fpPubkeyHex, btcHeight)
		if err != nil {
			return nil, err
		}
		fpPowerMap[fpPubkeyHex] = fpPower
	}

	return fpPowerMap, nil
}

// QueryEarliestActiveDelBtcHeight returns the earliest active BTC staking height
func (bbnClient *Client) QueryEarliestActiveDelBtcHeight(fpPkHexList []string) (uint64, error) {
	allFpEarliestDelBtcHeight := uint64(math.MaxUint64)

	for _, fpPkHex := range fpPkHexList {
		fpEarliestDelBtcHeight, err := bbnClient.QueryFpEarliestActiveDelBtcHeight(fpPkHex)
		if err != nil {
			return math.MaxUint64, err
		}
		if fpEarliestDelBtcHeight < allFpEarliestDelBtcHeight {
			allFpEarliestDelBtcHeight = fpEarliestDelBtcHeight
		}
	}

	return allFpEarliestDelBtcHeight, nil
}

func (bbnClient *Client) QueryFpEarliestActiveDelBtcHeight(fpPubkeyHex string) (uint64, error) {
	pagination := &sdkquerytypes.PageRequest{
		Limit: 100,
	}

	// queries the BTCStaking module for all delegations of a finality provider
	resp, err := bbnClient.QueryClient.FinalityProviderDelegations(fpPubkeyHex, pagination)
	if err != nil {
		return math.MaxUint64, err
	}

	// queries BtcConfirmationDepth, CovenantQuorum, and the latest BTC header
	btccheckpointParams, err := bbnClient.QueryClient.BTCCheckpointParams()
	if err != nil {
		return math.MaxUint64, err
	}

	// get the BTC staking params
	btcstakingParams, err := bbnClient.QueryClient.BTCStakingParams()
	if err != nil {
		return math.MaxUint64, err
	}

	// get the latest BTC header
	btcHeader, err := bbnClient.QueryClient.BTCHeaderChainTip()
	if err != nil {
		return math.MaxUint64, err
	}

	kValue := btccheckpointParams.GetParams().BtcConfirmationDepth
	covQuorum := btcstakingParams.GetParams().CovenantQuorum
	latestBtcHeight := btcHeader.GetHeader().Height

	earliestBtcHeight := uint64(math.MaxUint64)
	for {
		// btcDels contains all the queried BTC delegations
		for _, btcDels := range resp.BtcDelegatorDelegations {
			for _, btcDel := range btcDels.Dels {
				activationHeight := getDelFirstActiveHeight(btcDel, latestBtcHeight, kValue, covQuorum)
				if activationHeight < earliestBtcHeight {
					earliestBtcHeight = activationHeight
				}
			}
		}
		if resp.Pagination == nil || resp.Pagination.NextKey == nil {
			break
		}
		pagination.Key = resp.Pagination.NextKey
	}
	return earliestBtcHeight, nil
}

// The active delegation needs to satisfy:
// 1) the staking tx is k-deep in Bitcoin, i.e., start_height + k
// 2) it receives a quorum number of covenant committee signatures
//
// return math.MaxUint64 if the delegation is not active
//
// Note: the delegation can be unbounded and that's totally fine and shouldn't affect when the chain was activated
func getDelFirstActiveHeight(btcDel *types.BTCDelegationResponse, latestBtcHeight, kValue uint64, covQuorum uint32) uint64 {
	activationHeight := btcDel.StartHeight + kValue
	// not activated yet
	if latestBtcHeight < activationHeight || uint32(len(btcDel.CovenantSigs)) < covQuorum {
		return math.MaxUint64
	}
	return activationHeight
}
