package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/spf13/cobra"

	"github.com/babylonchain/babylon-finality-gadget/db"
	"github.com/babylonchain/babylon-finality-gadget/server"
	"github.com/babylonchain/babylon-finality-gadget/verifier"
	"github.com/babylonchain/babylon-finality-gadget/verifier/config"
)

const (
	cfgFlag = "cfg"
)

// CommandStart returns the start command of fpd daemon.
func CommandStart() *cobra.Command {
	var cmd = &cobra.Command{
		Use:     "start",
		Short:   "Start the finality gadget verifier daemon",
		Long:    `Start the finality gadget verifier daemon. Note that a config.toml file is required to start the daemon.`,
		Example: `vrf start --cfg config.toml`,
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

	// Init local DB for storing and querying blocks
	db, err := db.NewBBoltHandler(cfg.DBFilePath)
	if err != nil {
		return fmt.Errorf("failed to create DB handler: %w", err)
	}
	defer db.Close()
	err = db.TryCreateInitialBuckets()
	if err != nil {
		return fmt.Errorf("create initial buckets error: %w", err)
	}

	// Create verifier
	vrf, err := verifier.NewVerifier(cfg, db)
	if err != nil {
		return fmt.Errorf("error creating verifier: %v", err)
	}
	defer vrf.Close()

	// Start server
	s, err := server.Start(&server.ServerConfig{
		Port: cfg.ServerPort,
		Db:   db,
	})
	if err != nil {
		log.Fatalf("Error starting server: %v\n", err)
	}

	// Set up channel to listen for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Run verifier in a separate goroutine
	go func() {
		if err := vrf.ProcessBlocks(); err != nil {
			log.Fatalf("Error processing blocks: %v\n", err)
		}
	}()

	// Wait for interrupt signal
	sig := <-sigChan
	if sig == os.Interrupt {
		fmt.Print("\r")
	}

	// Call Stop method when interrupt signal is received
	if err := s.Stop(); err != nil {
		log.Fatalf("Error stopping server: %v\n", err)
		return err
	}
	return nil
}
