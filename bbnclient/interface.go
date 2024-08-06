package bbnclient

type IBabylonClient interface {
	QueryAllFpBtcPubKeys(consumerId string) ([]string, error)
	QueryFpPower(fpPubkeyHex string, btcHeight uint64) (uint64, error)
	QueryMultiFpPower(fpPubkeyHexList []string, btcHeight uint64) (map[string]uint64, error)
	QueryEarliestActiveDelBtcHeight(fpPubkeyHexList []string) (uint64, error)
}
