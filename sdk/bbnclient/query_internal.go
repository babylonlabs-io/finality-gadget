package bbnclient

import btcstakingtypes "github.com/babylonchain/babylon/x/btcstaking/types"

// we implemented exact logic as in GetStatus
// https://github.com/babylonchain/babylon-private/blob/c5a8d317091e2965e20ea56fa10e98d34aaa3547/x/btcstaking/types/btc_delegation.go#L88-L109
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
	kValue := btccheckpointParams.GetParams().BtcConfirmationDepth
	wValue := btccheckpointParams.GetParams().CheckpointFinalizationTimeout
	covQuorum := btcstakingParams.GetParams().CovenantQuorum
	ud := btcDel.UndelegationResponse

	if len(ud.GetDelegatorUnbondingSigHex()) > 0 {
		return false, nil
	}

	// k is not involved in the `GetStatus` logic as Babylon will accept a BTC delegation request
	// only when staking tx is k-deep on BTC.
	//
	// But the msg handler performs both checks 1) ensure staking tx is k-deep, and 2) ensure the
	// staking tx's timelock has at least w BTC blocks left.
	// (https://github.com/babylonchain/babylon-private/blob/d64ddc97d1c8b9f695b814b7b1b92ce133f2547b/x/btcstaking/keeper/msg_server.go#L266-L278)
	//
	// So after the msg handler accepts BTC delegation the 1st check is no longer needed
	// the k-value check is added per
	//
	// So in our case, we need to check both to ensure the delegation is active
	if btcHeight < btcDel.StartHeight+kValue || btcHeight+wValue > btcDel.EndHeight {
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
