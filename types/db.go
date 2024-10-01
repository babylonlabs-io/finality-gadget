package types

import (
	"time"
)

type TxInfo struct {
	BlockHeight    int64
	BlockTimestamp time.Time
	TxHash         string
	TxIndex        uint32
}

type EventNewFinalityProvider struct {
	DescriptionMoniker         string
	DescriptionIdentity        string
	DescriptionWebsite         string
	DescriptionSecurityContact string
	DescriptionDetails         string
	Commission                 string
	BabylonPkKey               string
	BtcPk                      string
	// PopBtcSigType              string
	// PopBabylonSig              string
	// PopBtcSig                  string
	// MasterPubRand              string
	// RegisteredEpoch            string
	SlashedBabylonHeight string
	SlashedBtcHeight     string
	ConsumerId           string
	// MsgIndex                   string
}

type EventBTCDelegationStateUpdate struct {
	StakingTxHash string              `json:"staking_tx_hash"`
	NewState      BTCDelegationStatus `json:"new_state"`
}

type EventSlashedFinalityProvider struct {
	FpBtcPk              []byte `json:"fp_btc_pk"`
	BlockHeight          uint64 `json:"block_height"`
	PubRand              []byte `json:"pub_rand"`
	CanonicalAppHash     []byte `json:"canonical_app_hash"`
	ForkAppHash          []byte `json:"fork_app_hash"`
	CanonicalFinalitySig []byte `json:"canonical_finality_sig"`
	ForkFinalitySig      []byte `json:"fork_finality_sig"`
}

type EventJailedFinalityProvider struct {
	PublicKey []byte `json:"public_key"`
}

type EventUnjailedFinalityProvider struct {
	PublicKey []byte `json:"public_key"`
}

type EventMessage struct {
	Action   string `json:"action"`
	Sender   string `json:"sender"`
	Module   string `json:"module"`
	MsgIndex string `json:"msg_index"`
}

// TODO: replaced 'CovenantSigs' by 'NumCovenantSigs', consider if ok
// TODO: omitted UndelegationResponse, consider if ok
type BTCDelegation struct {
	StakerAddr           string   `json:"staker_addr"`
	BtcPk                string   `json:"btc_pk"`
	FpBtcPkList          []string `json:"fp_btc_pk_list"`
	StartHeight          uint64   `json:"start_height"`
	EndHeight            uint64   `json:"end_height"`
	TotalSat             uint64   `json:"total_sat"`
	StakingTxHex         string   `json:"staking_tx_hex"`
	SlashingTxHex        string   `json:"slashing_tx_hex"`
	DelegatorSlashSigHex string   `json:"delegator_slash_sig_hex"`
	// CovenantSigs         []CovenantSignatures `json:"covenant_sigs"`
	NumCovenantSigs  uint32 `json:"num_covenant_sigs"`
	StakingOutputIdx uint32 `json:"staking_output_idx"`
	Active           bool   `json:"active"`
	StatusDesc       string `json:"status_desc"`
	UnbondingTime    uint32 `json:"unbonding_time"`
	// UndelegationResponse *bbntypes.BTCUndelegationResponse `json:"undelegation_response"`
	ParamsVersion uint32 `json:"params_version"`
}

// type CovenantSignatures struct {
// 	CovPk       string   `json:"cov_pk"`
// 	AdaptorSigs []string `json:"adaptor_sigs"`
// }