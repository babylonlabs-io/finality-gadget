package types

import "time"

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
	PopBtcSigType              string
	PopBabylonSig              string
	PopBtcSig                  string
	MasterPubRand              string
	RegisteredEpoch            string
	SlashedBabylonHeight       string
	SlashedBtcHeight           string
	ConsumerId                 string
	MsgIndex                   string
}

type EventBTCDelegationStateUpdate struct {
	StakingTxHash string `json:"staking_tx_hash"`
	NewState      string `json:"new_state"`
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

type EventSelectiveSlashing struct {
	StakingTxHash    string `json:"staking_tx_hash"`
	FpBtcPk          []byte `json:"fp_btc_pk"`
	RecoveredFpBtcSk []byte `json:"recovered_fp_btc_sk"`
}

type EventMessage struct {
	Action   string `json:"action"`
	Sender   string `json:"sender"`
	Module   string `json:"module"`
	MsgIndex string `json:"msg_index"`
}
