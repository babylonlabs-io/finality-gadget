package bbnclient

import btcstakingtypes "github.com/babylonchain/babylon/x/btcstaking/types"

// we implemented exact logic as in
// https://github.com/babylonchain/babylon-private/blob/c5a8d317091e2965e20ea56fa10e98d34aaa3547/x/btcstaking/types/btc_delegation.go#L111-L119
func (bbnClient *Client) isDelegationActive(
	btcDel *btcstakingtypes.BTCDelegationResponse,
	btcHeight uint64,
) (bool, error) {
	btccheckpointParams, err := bbnClient.QueryClient.BTCCheckpointParams()
	if err != nil {
		return false, err
	}
	btcstakingParams, err := bbnClient.QueryClient.BTCStakingParams()
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
