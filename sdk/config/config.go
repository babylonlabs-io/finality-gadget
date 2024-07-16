package config

import (
	"fmt"

	"github.com/babylonchain/babylon-finality-gadget/sdk/btcclient"
)

const (
	BabylonLocalnet = "chain-test"
	BabylonDevnet   = "euphrates-0.2.0"
)

// Config defines configuration for the Babylon query client
type Config struct {
	BTCConfig    *btcclient.BTCConfig
	ContractAddr string // CosmWasm contract address
	// TODO: add Config.Validate() to query chain ID (i.e. /status) from RPCAddr and compare
	ChainID string // Chain ID of the Babylon chain (e.g. devnet, testnet, mainnet)
	RPCAddr string // RPC address of the Babylon chain
}

func (config *Config) GetRpcAddr() (string, error) {
	if config.RPCAddr != "" {
		return config.RPCAddr, nil
	}
	return config.getDefaultRpcAddr()
}

func (config *Config) getDefaultRpcAddr() (string, error) {
	switch config.ChainID {
	case "chain-test":
		// for the e2e test
		return "http://127.0.0.1:26657", nil
	case "euphrates-0.2.0":
		return "https://rpc-euphrates.devnet.babylonchain.io/", nil
	// TODO: replace with babylon RPCs when QuerySmartContractStateRequest query is supported
	// TODO: add mainnet RPCs when available
	default:
		return "", fmt.Errorf("unrecognized chain id: %s", config.ChainID)
	}
}
