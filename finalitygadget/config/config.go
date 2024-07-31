package config

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	L2RPCHost         string        `long:"l2-rpc-host" description:"rpc host address of the L2 node"`
	BitcoinRPCHost    string        `long:"bitcoin-rpc-host" description:"rpc host address of the bitcoin node"`
	FGContractAddress string        `long:"fg-contract-address" description:"BabylonChain op finality gadget contract address"`
	BBNChainID        string        `long:"bbn-chain-id" description:"BabylonChain chain ID"`
	BBNRPCAddress     string        `long:"bbn-rpc-address" description:"BabylonChain chain RPC address"`
	DBFilePath        string        `long:"db-file-path" description:"path to the DB file"`
	GRPCServerPort    string        `long:"grpc-server-port" description:"port of the gRPC server"`
	PollInterval      time.Duration `long:"retry-interval" description:"interval in seconds to recheck Babylon finality of block"`
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

	return &config, nil
}
