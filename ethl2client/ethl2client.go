package ethl2client

import (
	"context"
	"fmt"
	"math/big"

	eth "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type EthL2Client struct {
	client *ethclient.Client
}

//////////////////////////////
// CONSTRUCTOR
//////////////////////////////

func NewEthL2Client(rpcHostAddr string) (*EthL2Client, error) {
	l2Client, err := ethclient.Dial(rpcHostAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to create ETH L2 client: %w", err)
	}

	return &EthL2Client{
		client: l2Client,
	}, nil
}

//////////////////////////////
// METHODS
//////////////////////////////

func (c *EthL2Client) HeaderByNumber(ctx context.Context, number *big.Int) (*eth.Header, error) {
	return c.client.HeaderByNumber(ctx, number)
}

func (c *EthL2Client) Close() {
	c.client.Close()
}
