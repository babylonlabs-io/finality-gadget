package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/babylonlabs-io/finality-gadget/config"
	"github.com/babylonlabs-io/finality-gadget/db"
	"github.com/babylonlabs-io/finality-gadget/finalitygadget"
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
	db, err := db.NewBBoltHandler(cfg.DBFilePath, logger)
	if err != nil {
		return fmt.Errorf("failed to create DB handler: %w", err)
	}
	defer db.Close()
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

	// Set up channel to listen for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Run finality gadget in a separate goroutine
	go func() {
		if err := fg.ProcessBlocks(context.Background()); err != nil {
			logger.Fatal("Error processing blocks", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	sig := <-sigChan
	if sig == os.Interrupt {
		fmt.Print("\r")
	}

	// Call Close method when interrupt signal is received
	logger.Info("Closing finality gadget server...")
	fg.Close()

	return nil
}
