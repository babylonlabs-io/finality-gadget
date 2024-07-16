package client

import (
	"github.com/babylonchain/babylon-finality-gadget/sdk/cwclient"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
)

type IBabylonClient interface {
	QueryAllFpBtcPubKeys(consumerId string) ([]string, error)
	QueryFpPower(fpPubkeyHex string, btcHeight uint64) (uint64, error)
	QueryMultiFpPower(fpPubkeyHexList []string, btcHeight uint64) (map[string]uint64, error)
}

type IBitcoinClient interface {
	GetBlockCount() (uint64, error)
	GetBlockHashByHeight(height uint64) (*chainhash.Hash, error)
	GetBlockHeaderByHash(blockHash *chainhash.Hash) (*wire.BlockHeader, error)
	GetBlockHeightByTimestamp(targetTimestamp uint64) (uint64, error)
}

type ICosmWasmClient interface {
	QueryListOfVotedFinalityProviders(queryParams *cwclient.L2Block) ([]string, error)
	QueryConsumerId() (string, error)
	QueryIsEnabled() (bool, error)
}
