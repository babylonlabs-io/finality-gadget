package types

type FinalityProvider struct {
	Description          FinalityProviderDescription `json:"description"`
	Commission           string                      `json:"commission"`
	BabylonPk            FinaltiyProviderBabylonPk   `json:"babylon_pk"`
	BtcPk                string                      `json:"btc_pk"`
	Pop                  FinalityProviderPop         `json:"pop"`
	MasterPubRand        string                      `json:"master_pub_rand"`
	RegisteredEpoch      string                      `json:"registered_epoch"`
	SlashedBabylonHeight string                      `json:"slashed_babylon_height"`
	SlashedBtcHeight     string                      `json:"slashed_btc_height"`
	ConsumerId           string                      `json:"consumer_id"`
}

type FinalityProviderDescription struct {
	Moniker         string `json:"moniker"`
	Identity        string `json:"identity"`
	Website         string `json:"website"`
	SecurityContact string `json:"security_contact"`
	Details         string `json:"details"`
}

type FinaltiyProviderBabylonPk struct {
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
