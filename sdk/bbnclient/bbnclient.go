package bbnclient

import (
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

func (bbnClient *Client) QueryEarliestActiveDelBtcHeight(fpPkHexList []string) (*uint64, error) {
	var earliestHeight *uint64

	for _, fpPkHex := range fpPkHexList {
		fpHeight, err := bbnClient.QueryFpEarliestActiveDelBtcHeight(fpPkHex)
		if err != nil {
			return nil, err
		}
		if fpHeight != nil && (earliestHeight == nil || *fpHeight < *earliestHeight) {
			earliestHeight = fpHeight
		}
	}

	return earliestHeight, nil
}

func (bbnClient *Client) QueryFpEarliestActiveDelBtcHeight(fpPubkeyHex string) (*uint64, error) {
	pagination := &sdkquerytypes.PageRequest{
		Limit: 100,
	}

	// queries the BTCStaking module for all delegations of a finality provider
	resp, err := bbnClient.QueryClient.FinalityProviderDelegations(fpPubkeyHex, pagination)
	if err != nil {
		return nil, err
	}

	// queries BtcConfirmationDepth, CovenantQuorum, and the latest BTC header
	btccheckpointParams, err := bbnClient.QueryClient.BTCCheckpointParams()
	if err != nil {
		return nil, err
	}

	// get the BTC staking params
	btcstakingParams, err := bbnClient.QueryClient.BTCStakingParams()
	if err != nil {
		return nil, err
	}

	// get the latest BTC header
	btcHeader, err := bbnClient.QueryClient.BTCHeaderChainTip()
	if err != nil {
		return nil, err
	}

	kValue := btccheckpointParams.GetParams().BtcConfirmationDepth
	covQuorum := btcstakingParams.GetParams().CovenantQuorum
	latestBtcHeight := btcHeader.GetHeader().Height

	var earliestBtcHeight *uint64
	for {
		// btcDels contains all the queried BTC delegations
		for _, btcDels := range resp.BtcDelegatorDelegations {
			for _, btcDel := range btcDels.Dels {
				// check whether the delegation is active
				confirmationHeight := btcDel.StartHeight + kValue
				if isActiveBtcDelegation(btcDel, latestBtcHeight, confirmationHeight, covQuorum) {
					if earliestBtcHeight == nil || confirmationHeight < *earliestBtcHeight {
						earliestBtcHeight = &confirmationHeight
					}
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
func isActiveBtcDelegation(btcDel *types.BTCDelegationResponse, latestBtcHeight, confirmationHeight uint64, covQuorum uint32) bool {
	return latestBtcHeight > confirmationHeight &&
		btcDel.EndHeight > confirmationHeight &&
		uint32(len(btcDel.CovenantSigs)) > covQuorum
}
