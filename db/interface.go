package db

import "github.com/babylonlabs-io/finality-gadget/types"

type IDatabaseHandler interface {
	CreateInitialSchema() error
	InsertBlock(block *types.Block) error
	GetBlockByHeight(height uint64) (*types.Block, error)
	GetBlockByHash(hash string) (*types.Block, error)
	GetBlockStatusByHeight(height uint64) (bool, error)
	GetBlockStatusByHash(hash string) (bool, error)
	GetLatestBlock() (*types.Block, error)
	DeleteDB() error
	Close() error
}
