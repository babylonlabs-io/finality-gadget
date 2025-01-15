package finalitygadget

import (
	"context"
	"math/big"

	"github.com/babylonlabs-io/finality-gadget/types"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	eth "github.com/ethereum/go-ethereum/core/types"
)

type IBitcoinClient interface {
	GetBlockCount() (uint32, error)
	GetBlockHashByHeight(height uint32) (*chainhash.Hash, error)
	GetBlockHeaderByHash(blockHash *chainhash.Hash) (*wire.BlockHeader, error)
	GetBlockHeightByTimestamp(targetTimestamp uint64) (uint32, error)
	GetBlockTimestampByHeight(height uint32) (uint64, error)
}

type IBabylonClient interface {
	QueryAllFpBtcPubKeys(consumerId string) ([]string, error)
	QueryFpPower(fpPubkeyHex string, btcHeight uint32) (uint64, error)
	QueryMultiFpPower(fpPubkeyHexList []string, btcHeight uint32) (map[string]uint64, error)
	QueryEarliestActiveDelBtcHeight(fpPubkeyHexList []string) (uint32, error)
}

type ICosmWasmClient interface {
	QueryListOfVotedFinalityProviders(queryParams *types.Block) ([]string, error)
	QueryConsumerId() (string, error)
	QueryIsEnabled() (bool, error)
}

type IEthL2Client interface {
	HeaderByNumber(ctx context.Context, number *big.Int) (*eth.Header, error)
	TransactionReceipt(ctx context.Context, txHash string) (*eth.Receipt, error)
	Close()
}
