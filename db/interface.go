package db

import "github.com/babylonlabs-io/finality-gadget/types"

type IDatabaseHandler interface {
	CreateInitialSchema() error
	InsertBlock(block *types.Block) error
	GetBlockByHeight(height uint64) (*types.Block, error)
	GetBlockByHash(hash string) (*types.Block, error)
	QueryIsBlockFinalizedByHeight(height uint64) (bool, error)
	QueryIsBlockFinalizedByHash(hash string) (bool, error)
	QueryLatestFinalizedBlock() (*types.Block, error)
	GetActivatedTimestamp() (uint64, error)
	SaveActivatedTimestamp(timestamp uint64) error
	Close() error
}
