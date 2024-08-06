package ethl2client

import (
	"context"
	"math/big"

	eth "github.com/ethereum/go-ethereum/core/types"
)

type IEthL2Client interface {
	HeaderByNumber(ctx context.Context, number *big.Int) (*eth.Header, error)
	Close()
}
