package btcclient

import (
	"fmt"
	"math"

	"github.com/avast/retry-go/v4"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
	"go.uber.org/zap"
)

type BitcoinClient struct {
	client *rpcclient.Client
	logger *zap.Logger
	cfg    *BTCConfig
}

//////////////////////////////
// CONSTRUCTOR
//////////////////////////////

func NewBitcoinClient(cfg *BTCConfig, logger *zap.Logger) (*BitcoinClient, error) {
	c, err := rpcclient.New(cfg.ToConnConfig(), nil)
	if err != nil {
		return nil, err
	}

	return &BitcoinClient{
		client: c,
		logger: logger,
		cfg:    cfg,
	}, nil
}

//////////////////////////////
// METHODS
//////////////////////////////

type BlockCountResponse struct {
	count int64
}

func (c *BitcoinClient) GetBlockCount() (uint64, error) {
	callForBlockCount := func() (*BlockCountResponse, error) {
		count, err := c.client.GetBlockCount()
		if err != nil {
			return nil, err
		}

		return &BlockCountResponse{count: count}, nil
	}

	blockCount, err := clientCallWithRetry(callForBlockCount, c.logger, c.cfg)
	if err != nil {
		return 0, fmt.Errorf("failed to get block count: %w", err)
	}

	if blockCount.count < 0 {
		return 0, fmt.Errorf("unexpected negative block count: %d", blockCount.count)
	}

	return uint64(blockCount.count), nil
}

func (c *BitcoinClient) GetBlockHashByHeight(height uint64) (*chainhash.Hash, error) {
	callForBlockHash := func() (*chainhash.Hash, error) {
		if height > math.MaxInt64 {
			return nil, fmt.Errorf("block height %d exceeds maximum int64 value", height)
		}
		return c.client.GetBlockHash(int64(height))
	}

	blockHash, err := clientCallWithRetry(callForBlockHash, c.logger, c.cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to get block by height %d: %w", height, err)
	}

	return blockHash, nil
}

func (c *BitcoinClient) GetBlockHeaderByHash(blockHash *chainhash.Hash) (*wire.BlockHeader, error) {
	callForBlockHeader := func() (*wire.BlockHeader, error) {
		return c.client.GetBlockHeader(blockHash)
	}

	header, err := clientCallWithRetry(callForBlockHeader, c.logger, c.cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to get block header by hash %s: %w", blockHash.String(), err)
	}

	return header, nil
}

func (c *BitcoinClient) GetBlockHeightByTimestamp(targetTimestamp uint64) (uint64, error) {
	// get the height of the most-work fully-validated chain
	blockHeight, err := c.GetBlockCount()
	if err != nil {
		return 0, err
	}

	lowerBound := uint64(0)
	upperBound := blockHeight

	for lowerBound <= upperBound {
		midHeight := (lowerBound + upperBound) / 2

		blockTimestamp, err := c.GetBlockTimestampByHeight(midHeight)
		if err != nil {
			return 0, err
		}

		if blockTimestamp < targetTimestamp {
			lowerBound = midHeight + 1
		} else if blockTimestamp > targetTimestamp {
			upperBound = midHeight - 1
		} else {
			return midHeight, nil
		}
	}

	// timestamp is in the future (not in the most-work fully-validated chain)
	// we return the max uint64 to indicate this
	if lowerBound > blockHeight {
		return math.MaxUint64, nil
	}

	return lowerBound - 1, nil
}

func (c *BitcoinClient) GetBlockTimestampByHeight(height uint64) (uint64, error) {
	// get block hash by height
	blockHash, err := c.GetBlockHashByHeight(height)
	if err != nil {
		return 0, err
	}

	// get block header by hash. the header contains info such as the block time expressed in UNIX epoch time
	blockHeader, err := c.GetBlockHeaderByHash(blockHash)
	if err != nil {
		return 0, err
	}

	timestamp := blockHeader.Timestamp.Unix()
	if timestamp < 0 {
		return 0, fmt.Errorf("negative timestamp encountered: %d", timestamp)
	}
	return uint64(timestamp), nil
}

func (c *BitcoinClient) GetBlock(hash *chainhash.Hash) (*wire.MsgBlock, error) {
	callForBlock := func() (*wire.MsgBlock, error) {
		return c.client.GetBlock(hash)
	}

	block, err := clientCallWithRetry(callForBlock, c.logger, c.cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to get block by hash %s: %w", hash.String(), err)
	}

	return block, nil
}

//////////////////////////////
// INTERNAL
//////////////////////////////

func clientCallWithRetry[T any](
	call retry.RetryableFuncWithData[*T], logger *zap.Logger, cfg *BTCConfig,
) (*T, error) {
	result, err := retry.DoWithData(
		call,
		retry.Attempts(cfg.MaxRetryTimes),
		retry.Delay(cfg.RetryInterval),
		retry.LastErrorOnly(true),
		retry.OnRetry(func(n uint, err error) {
			logger.Debug(
				"failed to call the RPC client",
				zap.Uint("attempt", n+1),
				zap.Uint("max_attempts", cfg.MaxRetryTimes),
				zap.Error(err),
			)
		}),
	)

	if err != nil {
		return nil, err
	}
	return result, nil
}
