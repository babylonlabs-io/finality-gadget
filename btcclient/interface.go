package btcclient

import (
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
)

type IBitcoinClient interface {
	GetBlockCount() (uint64, error)
	GetBlockHashByHeight(height uint64) (*chainhash.Hash, error)
	GetBlockHeaderByHash(blockHash *chainhash.Hash) (*wire.BlockHeader, error)
	GetBlockHeightByTimestamp(targetTimestamp uint64) (uint64, error)
	GetBlockTimestampByHeight(height uint64) (uint64, error)
}
