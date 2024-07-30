package main

import (
	"fmt"
	"log"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/spf13/cobra"

	"github.com/babylonchain/babylon-finality-gadget/verifier/config"
	"github.com/babylonchain/babylon-finality-gadget/verifier/db"
	"github.com/babylonchain/babylon-finality-gadget/verifier/server"
	"github.com/babylonchain/babylon-finality-gadget/verifier/verifier"
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

	// Start server
	go func() {
		if _, err := server.StartServer(&server.ServerConfig{
			Port:   cfg.ServerPort,
			Db: 		db,
		}); err != nil {
			log.Fatalf("Error starting server: %v\n", err)
		}
	}()

	// Start processing blocks
	return vrf.ProcessBlocks()
}