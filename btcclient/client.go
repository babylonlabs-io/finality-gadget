package btcclient

import (
	"fmt"

	"github.com/avast/retry-go/v4"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
	"go.uber.org/zap"
)

type BTCClient struct {
	client *rpcclient.Client
	logger *zap.Logger
	cfg    *BTCConfig
}

func NewBTCClient(cfg *BTCConfig, logger *zap.Logger) (*BTCClient, error) {
	c, err := rpcclient.New(cfg.ToConnConfig(), nil)
	if err != nil {
		return nil, err
	}

	return &BTCClient{
		client: c,
		logger: logger,
		cfg:    cfg,
	}, nil
}

type BlockCountResponse struct {
	count int64
}

func (c *BTCClient) GetBlockCount() (uint64, error) {
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

	return uint64(blockCount.count), nil
}

func (c *BTCClient) GetBlockHashByHeight(height uint64) (*chainhash.Hash, error) {
	callForBlockHash := func() (*chainhash.Hash, error) {
		return c.client.GetBlockHash(int64(height))
	}

	blockHash, err := clientCallWithRetry(callForBlockHash, c.logger, c.cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to get block by height %d: %w", height, err)
	}

	return blockHash, nil
}

func (c *BTCClient) GetBlockHeaderByHash(blockHash *chainhash.Hash) (*wire.BlockHeader, error) {
	callForBlockHeader := func() (*wire.BlockHeader, error) {
		return c.client.GetBlockHeader(blockHash)
	}

	header, err := clientCallWithRetry(callForBlockHeader, c.logger, c.cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to get block header by hash %s: %w", blockHash.String(), err)
	}

	return header, nil
}

func (c *BTCClient) GetBlockHeightByTimestamp(targetTimestamp uint64) (uint64, error) {
	// get the height of the most-work fully-validated chain
	blockHeight, err := c.GetBlockCount()
	if err != nil {
		return 0, err
	}

	lowerBound := uint64(0)
	upperBound := blockHeight

	for lowerBound <= upperBound {
		midHeight := (lowerBound + upperBound) / 2

		// get block hash by height
		blockHash, err := c.GetBlockHashByHeight(midHeight)
		if err != nil {
			return 0, err
		}

		// get block header by hash. the header contains info such as the block time expressed in UNIX epoch time
		blockHeader, err := c.GetBlockHeaderByHash(blockHash)
		if err != nil {
			return 0, err
		}

		blockTimestamp := uint64(blockHeader.Timestamp.Unix())

		if blockTimestamp < targetTimestamp {
			lowerBound = midHeight + 1
		} else if blockTimestamp > targetTimestamp {
			upperBound = midHeight - 1
		} else {
			return midHeight, nil
		}
	}

	// timestamp is in the future (not in the most-work fully-validated chain)
	// so we cannot determine the height from the timestamp
	if lowerBound > blockHeight {
		return 0, nil
	}

	return lowerBound - 1, nil
}

func clientCallWithRetry[T any](
	call retry.RetryableFuncWithData[*T], logger *zap.Logger, cfg *BTCConfig,
) (*T, error) {
	result, err := retry.DoWithData(call, retry.Attempts(cfg.MaxRetryTimes), retry.Delay(cfg.RetryInterval), retry.LastErrorOnly(true),
		retry.OnRetry(func(n uint, err error) {
			logger.Debug(
				"failed to call the RPC client",
				zap.Uint("attempt", n+1),
				zap.Uint("max_attempts", cfg.MaxRetryTimes),
				zap.Error(err),
			)
		}))

	if err != nil {
		return nil, err
	}
	return result, nil
}
