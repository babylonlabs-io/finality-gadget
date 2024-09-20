package types

type FinalityProvider struct {
	Description struct {
		Moniker         string `json:"moniker"`
		Identity        string `json:"identity"`
		Website         string `json:"website"`
		SecurityContact string `json:"security_contact"`
		Details         string `json:"details"`
	} `json:"description"`
	Commission string `json:"commission"`
	BabylonPk  struct {
		Key string `json:"key"`
	} `json:"babylon_pk"`
	BtcPk string `json:"btc_pk"`
	Pop   struct {
		BtcSigType string `json:"btc_sig_type"`
		BabylonSig string `json:"babylon_sig"`
		BtcSig     string `json:"btc_sig"`
	} `json:"pop"`
	MasterPubRand        string `json:"master_pub_rand"`
	RegisteredEpoch      string `json:"registered_epoch"`
	SlashedBabylonHeight string `json:"slashed_babylon_height"`
	SlashedBtcHeight     string `json:"slashed_btc_height"`
	ConsumerId           string `json:"consumer_id"`
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

type SelectiveSlashingEvidence struct {
	StakingTxHash    string `json:"staking_tx_hash"`
	FpBtcPk          []byte `json:"fp_btc_pk"`
	RecoveredFpBtcSk []byte `json:"recovered_fp_btc_sk"`
}
