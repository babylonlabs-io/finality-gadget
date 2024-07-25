package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/babylonchain/babylon-finality-gadget/verifier/verifier"
	"github.com/joho/godotenv"
)
func main() {
	ctx := context.Background()

	// Load .env file
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("error getting current working directory: %v\n", err)
	}
	envFile := filepath.Join(cwd, ".env")
	fmt.Printf("loading env file: %s\n", envFile)
	err = godotenv.Load(envFile)
	if err != nil {
		fmt.Printf("error loading .env file: %v\n", err)
	}
	
	// Import and parse env vars
	L2_RPC_HOST := os.Getenv("L2_RPC_HOST")
	BITCOIN_RPC_HOST := os.Getenv("BITCOIN_RPC_HOST")
	PG_CONNECTION_STRING := os.Getenv("PG_CONNECTION_STRING")
	FG_CONTRACT_ADDRESS := os.Getenv("FG_CONTRACT_ADDRESS")
	BBN_CHAIN_ID := os.Getenv("BBN_CHAIN_ID")
	BBN_RPC_ADDRESS := os.Getenv("BBN_RPC_ADDRESS")
	POLL_INTERVAL_IN_SECS := os.Getenv("POLL_INTERVAL_IN_SECS")

	pollInterval, err := strconv.Atoi(POLL_INTERVAL_IN_SECS)
	if err != nil {
		fmt.Printf("error converting poll interval to int: %v\n", err)
	}

	// Create verifier
	vf, err := verifier.NewVerifier(ctx, &verifier.Config{
		L2RPCHost: L2_RPC_HOST,
		BitcoinRPCHost: BITCOIN_RPC_HOST,
		PGConnectionString: PG_CONNECTION_STRING,
		FGContractAddress:FG_CONTRACT_ADDRESS,
		BBNChainID: BBN_CHAIN_ID,
		BBNRPCAddress: BBN_RPC_ADDRESS,
		PollInterval: time.Second * time.Duration(pollInterval),
	})
	if err != nil {
		fmt.Printf("error creating verifier: %v\n", err)
	}

	// Start processing blocks
	err = vf.ProcessBlocks(ctx)
	if err != nil {
		fmt.Println(err)
	}
}