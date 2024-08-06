package types

import (
	"context"
	"math/big"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	eth "github.com/ethereum/go-ethereum/core/types"
)

type IBabylonClient interface {
	QueryAllFpBtcPubKeys(consumerId string) ([]string, error)
	QueryFpPower(fpPubkeyHex string, btcHeight uint64) (uint64, error)
	QueryMultiFpPower(fpPubkeyHexList []string, btcHeight uint64) (map[string]uint64, error)
	QueryEarliestActiveDelBtcHeight(fpPubkeyHexList []string) (uint64, error)
}

type IBitcoinClient interface {
	GetBlockCount() (uint64, error)
	GetBlockHashByHeight(height uint64) (*chainhash.Hash, error)
	GetBlockHeaderByHash(blockHash *chainhash.Hash) (*wire.BlockHeader, error)
	GetBlockHeightByTimestamp(targetTimestamp uint64) (uint64, error)
	GetBlockTimestampByHeight(height uint64) (uint64, error)
}

type ICosmWasmClient interface {
	QueryListOfVotedFinalityProviders(queryParams *Block) ([]string, error)
	QueryConsumerId() (string, error)
	QueryIsEnabled() (bool, error)
}

type IEthL2Client interface {
	HeaderByNumber(ctx context.Context, number *big.Int) (*eth.Header, error)
	Close()
}
