package types

import (
	ctypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

type FinalityProvider struct {
	Description          *ctypes.Description        `json:"description"`
	Commission           string                     `json:"commission"`
	BabylonPk            *FinalityProviderBabylonPk `json:"babylon_pk"`
	BtcPk                string                     `json:"btc_pk"`
	SlashedBabylonHeight string                     `json:"slashed_babylon_height"`
	SlashedBtcHeight     string                     `json:"slashed_btc_height"`
	ConsumerId           string                     `json:"consumer_id"`
}

type FinalityProviderBabylonPk struct {
	Key string `json:"key"`
}

type FinalityProviderPop struct {
	BtcSigType string `json:"btc_sig_type"`
	BabylonSig string `json:"babylon_sig"`
	BtcSig     string `json:"btc_sig"`
}

type Evidence struct {
	FpBtcPk              []byte `json:"fp_btc_pk"`
	BlockHeight          uint64 `json:"block_height"`
	PubRand              []byte `json:"pub_rand"`
	CanonicalAppHash     []byte `json:"canonical_app_hash"`
	ForkAppHash          []byte `json:"fork_app_hash"`
	CanonicalFinalitySig []byte `json:"canonical_finality_sig"`
	ForkFinalitySig      []byte `json:"fork_finality_sig"`
}

type BTCDelegationStatus string

const (
	BTCDelegationPending  BTCDelegationStatus = "PENDING"
	BTCDelegationActive   BTCDelegationStatus = "ACTIVE"
	BTCDelegationUnbonded BTCDelegationStatus = "UNBONDED"
)
