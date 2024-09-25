package indexer

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/ignite/cli/v28/ignite/pkg/cosmosclient"

	"github.com/babylonlabs-io/finality-gadget/bbnclient"
	cfg "github.com/babylonlabs-io/finality-gadget/config"
	"github.com/babylonlabs-io/finality-gadget/db"
	ctypes "github.com/cometbft/cometbft/rpc/core/types"
)

const (
	defaultTXsPerPage = 30
)

type Indexer struct {
	bbnClient *bbnclient.BabylonClient
	rpcClient cosmosclient.Client
	db        db.IDatabaseHandler
	logger    *zap.Logger
	cfg       *cfg.Config

	pollInterval    time.Duration
	processedHeight int64
}

func NewIndexer(cfg *cfg.Config, db db.IDatabaseHandler, logger *zap.Logger) (*Indexer, error) {
	// Init Babylon client
	bbnClient, err := bbnclient.NewBabylonClient(cfg.BBNConfig, logger)
	if err != nil {
		logger.Error("Failed to create Babylon client", zap.Error(err))
		return nil, err
	}

	// Init Babylon rpc client
	rpcClient, err := cosmosclient.New(context.Background(), cosmosclient.WithNodeAddress(cfg.BabylonRPCAddress))
	if err != nil {
		logger.Error("Failed to create Babylon client", zap.Error(err))
		return nil, err
	}

	return &Indexer{
		bbnClient: bbnClient,
		rpcClient: rpcClient,
		db:        db,
		logger:    logger,
		cfg:       cfg,
	}, nil
}

func (idx *Indexer) LatestBlockHeight() (int64, error) {
	resp, err := idx.rpcClient.RPC.Status(context.Background())
	if err != nil {
		return 0, err
	}
	return resp.SyncInfo.LatestBlockHeight, nil
}

// Sync fetches all finality providers and delegations at the current block and saves them to the database.
// It begins at the start height and continues until the latest block height.
func (idx *Indexer) Sync() error {
	latestHeight, err := idx.LatestBlockHeight()
	if err != nil {
		return err
	}

	// Get latest fps and save them to db
	fps, err := idx.bbnClient.QueryAllFinalityProviders(idx.cfg.BBNConfig.BabylonChainId)
	if err != nil {
		return err
	}
	err = idx.db.SaveInitialFinalityProviders(fps)
	if err != nil {
		return err
	}

	// Loop through fps, get delegations and save them to db
	for _, fp := range fps {
		dels, err := idx.bbnClient.QueryFPDelegations(fp.BtcPk.MarshalHex())
		if err != nil {
			return err
		}
		err = idx.db.SaveInitialDelegations(dels)
		if err != nil {
			return err
		}
	}

	// Set last processed height.
	idx.processedHeight = latestHeight

	return nil
}

// Poll fetches new blocks at a set interval
func (idx *Indexer) Poll(ctx context.Context) error {
	ticker := time.NewTicker(idx.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			latestHeight, err := idx.LatestBlockHeight()
			if err != nil {
				return err
			}

			if latestHeight <= idx.processedHeight {
				continue
			}

			err = idx.CollectBlocks(idx.processedHeight+1, latestHeight)
			if err != nil {
				return err
			}
		}
	}
}

// Gathers transactions for all blocks starting from a specific height.
// Each group of block transactions is saved sequentially after being collected.
// Adapted from https://github.com/ignite/cli
func (idx *Indexer) CollectBlocks(fromHeight int64, toHeight int64) error {
	ctx := context.Background()
	tc := make(chan []Tx)
	wg, _ := errgroup.WithContext(ctx)

	// Start collecting block transactions.
	// The transactions channel is closed by the client when all transactions
	// are collected or when an error occurs during the collection.
	wg.Go(func() error {
		err := idx.CollectTxs(ctx, fromHeight, toHeight, tc)
		if err != nil {
			idx.logger.Error("Error collecting transactions", zap.Error(err))
			return err
		}
		return nil
	})

	// The transactions for each block are saved in "bulks" so they are not
	// kept in memory. Also, they are saved sequentially to avoid block height
	// gaps that can occur if a group of transactions from a previous block
	// fail to be saved.
	wg.Go(func() error {
		for txs := range tc {
			if err := idx.CollectEvents(ctx, txs); err != nil {
				idx.logger.Error("Error parsing events", zap.Error(err))
				return err
			}
		}
		return nil
	})

	return wg.Wait()
}

// Collects transactions from multiple consecutive blocks.
// Transactions from a single block are send to the channel only if all transactions
// from that block are collected successfully.
// Blocks are traversed sequentially starting from a height until the latest block height
// available at the moment this method is called.
// The channel might contain the transactions collected successfully up until that point
// when an error is returned.
// Adapted from https://github.com/ignite/cli
func (idx *Indexer) CollectTxs(ctx context.Context, fromHeight int64, toHeight int64, tc chan<- []Tx) error {
	defer close(tc)

	if fromHeight == 0 {
		fromHeight = 1
	}

	for height := fromHeight; height <= toHeight; height++ {
		idx.logger.Info("Processing height", zap.Int64("height", height))
		txs, err := idx.GetBlockTxs(ctx, height)
		if err != nil {
			idx.logger.Error("Error getting block transactions", zap.Error(err))
			return err
		}

		// Ignore blocks without transactions
		if txs == nil {
			continue
		}

		// Make sure that collection finishes if the context
		// is done when the transactions channel is full
		select {
		case <-ctx.Done():
			return ctx.Err()
		case tc <- txs:
		}
	}

	return nil
}

// GetBlockTxs returns the transactions in a block.
// The list of transactions can be empty if there are no transactions in the block
// at the moment this method is called.
// Tendermint might index a limited number of block so trying to fetch transactions
// from a block that is not indexed would return an error.
func (idx *Indexer) GetBlockTxs(ctx context.Context, height int64) (txs []Tx, err error) {
	if height == 0 {
		return nil, fmt.Errorf("height must be greater than 0")
	}

	r, err := idx.rpcClient.RPC.Block(ctx, &height)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch block %d: %w", height, err)
	}

	query := createTxSearchByHeightQuery(height)

	// Fetch transactions for first page and total page count
	txs = make([]Tx, 0)
	rawTxs, totalCount, err := idx.getBlockTxsByPage(ctx, 1, query, defaultTXsPerPage)
	if err != nil {
		return nil, err
	}
	for _, tx := range rawTxs {
		txs = append(txs, Tx{
			BlockTime: r.Block.Time,
			Raw:       tx,
		})
	}

	// Fetch transactions for remaining pages in parallel
	totalPages := (totalCount + defaultTXsPerPage - 1) / defaultTXsPerPage
	txsChan := make(chan []*ctypes.ResultTx, totalPages)

	g, gCtx := errgroup.WithContext(ctx)
	for i := 2; i <= totalPages; i++ {
		page := i
		g.Go(func() error {
			pageTxs, _, err := idx.getBlockTxsByPage(gCtx, page, query, defaultTXsPerPage)
			if err != nil {
				return err
			}
			select {
			case txsChan <- pageTxs:
			case <-gCtx.Done():
				return gCtx.Err()
			}
			return nil
		})
	}

	// Wait for all pages to be fetched
	go func() {
		g.Wait()
		close(txsChan)
	}()

	// Collect transactions from all pages
	for i := 2; i <= totalPages; i++ {
		select {
		case rawTxs := <-txsChan:
			for _, tx := range rawTxs {
				txs = append(txs, Tx{
					BlockTime: r.Block.Time,
					Raw:       tx,
				})
			}
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	// Check errors
	if err := g.Wait(); err != nil {
		return nil, err
	}

	// Set latest height in memory
	idx.processedHeight = height

	return txs, nil
}

func (idx *Indexer) getBlockTxsByPage(ctx context.Context, page int, query string, perPage int) ([]*ctypes.ResultTx, int, error) {
	res, err := idx.rpcClient.RPC.TxSearch(ctx, query, false, &page, &perPage, "asc")
	if err != nil {
		return nil, 0, err
	}

	return res.Txs, res.TotalCount, nil
}

func (idx *Indexer) CollectEvents(ctx context.Context, txs []Tx) error {
	// Start atomic insert tx
	pgTx, err := idx.db.BeginTx()
	if err != nil {
		return fmt.Errorf("unable to start DB tx: %v", err)
	}

	defer pgTx.Rollback(ctx)

	// Loop through txs
	for _, tx := range txs {
		txInfo := tx.GetTxInfo()
		events, err := tx.GetEvents()
		if err != nil {
			return err
		}
		// Loop through events
		for _, evt := range events {
			// Parse event and handle
			err := idx.ParseEvent(pgTx, txInfo, evt)
			if err != nil {
				return err
			}
		}
	}

	return pgTx.Commit(ctx)
}

func createTxSearchByHeightQuery(height int64) string {
	params := url.Values{}
	params.Set("tx.height", strconv.FormatInt(height, 10))
	return params.Encode()
}
