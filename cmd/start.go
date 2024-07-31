package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/spf13/cobra"

	rpcclient "github.com/babylonchain/babylon-finality-gadget/client"
	"github.com/babylonchain/babylon-finality-gadget/db"
	"github.com/babylonchain/babylon-finality-gadget/finalitygadget"
	"github.com/babylonchain/babylon-finality-gadget/finalitygadget/config"
	"github.com/babylonchain/babylon-finality-gadget/server"
	sig "github.com/lightningnetwork/lnd/signal"
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

	// Create finality gadget
	fg, err := finalitygadget.NewFinalityGadget(cfg, db)
	if err != nil {
		log.Fatalf("Error creating finality gadget: %v\n", err)
		return fmt.Errorf("error creating finality gadget: %v", err)
	}

	// Start grpc server
	// Hook interceptor for os signals.
	shutdownInterceptor, err := sig.Intercept()
	if err != nil {
		return err
	}
	srv := server.NewFinalityGadgetServer(cfg, db, fg, shutdownInterceptor)
	go func() {
		err = srv.RunUntilShutdown()
		if err != nil {
			log.Fatalf("Finality gadget server error: %v\n", err)
		}
	}()

	// Create grpc client
	hostAddr := "localhost:" + cfg.GRPCServerPort
	client, err := rpcclient.NewFinalityGadgetGrpcClient(db, hostAddr)
	if err != nil {
		log.Fatalf("Error creating grpc client: %v\n", err)
		return fmt.Errorf("error creating grpc client: %v", err)
	}

	// Set up channel to listen for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Run verifier in a separate goroutine
	go func() {
		if err := fg.ProcessBlocks(); err != nil {
			log.Fatalf("Error processing blocks: %v\n", err)
		}
	}()

	// Wait for interrupt signal
	sig := <-sigChan
	if sig == os.Interrupt {
		fmt.Print("\r")
	}

	// Call Close method when interrupt signal is received
	log.Printf("Closing finality gadget server...")
	if err := client.Close(); err != nil {
		log.Fatalf("Error stopping grpc client: %v\n", err)
		return err
	}
	fg.Close()

	return nil
}
