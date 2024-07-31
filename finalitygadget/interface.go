package finalitygadget

import "github.com/babylonchain/babylon-finality-gadget/types"

type IFinalityGadget interface {
	InsertBlock(block *types.Block) error
	GetBlockStatusByHeight(height uint64) (bool, error)
	GetBlockStatusByHash(hash string) (bool, error)
	GetLatestBlock() (*types.Block, error)
}
