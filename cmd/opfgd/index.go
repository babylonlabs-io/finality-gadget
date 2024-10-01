package main

import (
	"context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/babylonlabs-io/finality-gadget/config"
	"github.com/babylonlabs-io/finality-gadget/db/pg"
	"github.com/babylonlabs-io/finality-gadget/indexer"
	"github.com/babylonlabs-io/finality-gadget/log"
	sig "github.com/lightningnetwork/lnd/signal"
)

// CommandIndex returns the index command of fpd daemon.
func CommandIndex() *cobra.Command {
	var cmd = &cobra.Command{
		Use:     "index",
		Short:   "Start the op finality gadget indexer daemon",
		Long:    `Start the op finality gadget indexer daemon. Note that a config.toml file is required to start the indexer daemon.`,
		Example: `opfgd index --cfg config.toml`,
		Args:    cobra.NoArgs,
		RunE:    runEWithClientCtx(runIndexCmd),
	}
	return cmd
}

func runIndexCmd(ctx client.Context, cmd *cobra.Command, args []string) error {
	// Parse configs
	cfgPath, err := cmd.Flags().GetString(cfgFlag)
	if err != nil {
		return err
	}
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Create logger
	logger, err := log.NewRootLogger("console", true)
	if err != nil {
		return fmt.Errorf("failed to create logger: %w", err)
	}

	// Init local DB for storing and querying blocks
	db, err := pg.NewPostgresHandler(&cfg.DBConfig, logger)
	if err != nil {
		return fmt.Errorf("failed to create DB handler: %w", err)
	}

	// Create initial db schema
	err = db.CreateInitialSchema()
	if err != nil {
		return fmt.Errorf("create initial buckets error: %w", err)
	}

	// Create indexer
	idx, err := indexer.NewIndexer(&cfg.BBNConfig, db, logger)
	if err != nil {
		logger.Fatal("Error creating fp indexer", zap.Error(err))
		return err
	}

	// Create a cancellable context
	// Start grpc server
	// Hook interceptor for os signals.
	shutdownInterceptor, err := sig.Intercept()
	if err != nil {
		return err
	}

	// Index events in a separate goroutine
	go func() {
		// On startup, sync all fps and delegations as of latest height
		err := idx.Sync()
		if err != nil {
			logger.Fatal("Error syncing blocks", zap.Error(err))
		}

		// Once synced, start polling for new blocks
		err = idx.Poll(context.Background())
		if err != nil {
			logger.Fatal("Error polling blocks", zap.Error(err))
		}
	}()

	// Wait for shutdown signal
	<-shutdownInterceptor.ShutdownChannel()

	// Call Close method when interrupt signal is received
	logger.Info("Shutting down database...")
	db.Close()

	return nil
}
