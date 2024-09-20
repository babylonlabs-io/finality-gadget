package main

import (
	"context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/babylonlabs-io/finality-gadget/config"
	"github.com/babylonlabs-io/finality-gadget/db/pg"
	"github.com/babylonlabs-io/finality-gadget/finalitygadget"
	"github.com/babylonlabs-io/finality-gadget/indexer"
	"github.com/babylonlabs-io/finality-gadget/log"
	"github.com/babylonlabs-io/finality-gadget/server"
	sig "github.com/lightningnetwork/lnd/signal"
)

const (
	cfgFlag = "cfg"
)

// CommandStart returns the start command of fpd daemon.
func CommandStart() *cobra.Command {
	var cmd = &cobra.Command{
		Use:     "start",
		Short:   "Start the op finality gadget daemon",
		Long:    `Start the op finality gadget daemon. Note that a config.toml file is required to start the daemon.`,
		Example: `opfgd start --cfg config.toml`,
		Args:    cobra.NoArgs,
		RunE:    runEWithClientCtx(runStartCmd),
	}
	return cmd
}

func runStartCmd(ctx client.Context, cmd *cobra.Command, args []string) error {
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

	// Create finality gadget
	fg, err := finalitygadget.NewFinalityGadget(cfg, db, logger)
	if err != nil {
		logger.Fatal("Error creating finality gadget", zap.Error(err))
		return fmt.Errorf("error creating finality gadget: %v", err)
	}

	// Create indexer
	idx, err := indexer.NewIndexer(cfg, db, logger)
	if err != nil {
		logger.Fatal("Error creating fp indexer", zap.Error(err))
		return err
	}

	// Create a cancellable context
	fgCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// Start monitoring BTC staking activation
	go fg.MonitorBtcStakingActivation(fgCtx)

	// Start grpc server
	// Hook interceptor for os signals.
	shutdownInterceptor, err := sig.Intercept()
	if err != nil {
		return err
	}
	srv := server.NewFinalityGadgetServer(cfg, db, fg, shutdownInterceptor, logger)
	go func() {
		err = srv.RunUntilShutdown()
		if err != nil {
			logger.Fatal("Finality gadget server error", zap.Error(err))
		}
	}()

	// Run finality gadget in a separate goroutine
	go func() {
		if err := fg.ProcessBlocks(fgCtx); err != nil {
			logger.Fatal("Error processing blocks", zap.Error(err))
		}
	}()

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
	logger.Info("Closing finality gadget server...")
	fg.Close()

	logger.Info("Shutting down database...")
	db.Close()

	return nil
}
