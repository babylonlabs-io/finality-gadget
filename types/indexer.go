package types

type FinalityProvider struct {
	Addr                       string `json:"addr"`
	DescriptionMoniker         string `json:"description_moniker"`
	DescriptionIdentity        string `json:"description_identity"`
	DescriptionWebsite         string `json:"description_website"`
	DescriptionSecurityContact string `json:"description_security_contact"`
	DescriptionDetails         string `json:"description_details"`
	Commission                 string `json:"commission"`
	BtcPk                      string `json:"btc_pk"`
	SlashedBabylonHeight       uint64 `json:"slashed_babylon_height"`
	SlashedBtcHeight           uint64 `json:"slashed_btc_height"`
	ConsumerId                 string `json:"consumer_id"`
}
