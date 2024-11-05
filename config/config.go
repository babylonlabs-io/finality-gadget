package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	L2RPCHost         string        `long:"l2-rpc-host" description:"rpc host address of the L2 node"`
	BitcoinRPCHost    string        `long:"bitcoin-rpc-host" description:"rpc host address of the bitcoin node"`
	BitcoinRPCUser    string        `long:"bitcoin-rpc-user" description:"rpc user of the bitcoin node"`
	BitcoinRPCPass    string        `long:"bitcoin-rpc-pass" description:"rpc password of the bitcoin node"`
	FGContractAddress string        `long:"fg-contract-address" description:"BabylonChain op finality gadget contract address"`
	BBNChainID        string        `long:"bbn-chain-id" description:"BabylonChain chain ID"`
	BBNRPCAddress     string        `long:"bbn-rpc-address" description:"BabylonChain chain RPC address"`
	DBFilePath        string        `long:"db-file-path" description:"path to the DB file"`
	GRPCListener      string        `long:"grpc-listener" description:"host:port to listen for gRPC connections"`
	HTTPListener      string        `long:"http-listener" description:"host:port to listen for HTTP connections"`
	BitcoinDisableTLS bool          `long:"bitcoin-disable-tls" description:"disable TLS for RPC connections"`
	PollInterval      time.Duration `long:"retry-interval" description:"interval in seconds to recheck Babylon finality of block"`
	BatchSize         uint64        `long:"batch-size" description:"number of blocks to process in a batch"`
}

func (c *Config) Validate() error {
	// Required fields
	if c.L2RPCHost == "" {
		return fmt.Errorf("l2-rpc-host is required")
	}
	if c.BitcoinRPCHost == "" {
		return fmt.Errorf("bitcoin-rpc-host is required")
	}
	if c.FGContractAddress == "" {
		return fmt.Errorf("fg-contract-address is required")
	}
	if c.BBNChainID == "" {
		return fmt.Errorf("bbn-chain-id is required")
	}
	if c.BBNRPCAddress == "" {
		return fmt.Errorf("bbn-rpc-address is required")
	}
	// TODO: add some default values if missing
	if c.DBFilePath == "" {
		return fmt.Errorf("db-file-path is required")
	}
	if c.GRPCListener == "" {
		return fmt.Errorf("grpc-listener is required")
	}
	if c.HTTPListener == "" {
		return fmt.Errorf("http-listener is required")
	}

	// Numeric validations
	// TODO: add more validations (max batch size, min poll interval)
	if c.PollInterval <= 0 {
		return fmt.Errorf("poll-interval must be positive")
	}
	if c.BatchSize == 0 {
		return fmt.Errorf("batch-size must be greater than 0")
	}

	return nil
}

func Load(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("toml")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}
