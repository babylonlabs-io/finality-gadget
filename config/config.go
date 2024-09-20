package config

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	FGConfig  `mapstructure:"fg"`
	L2Config  `mapstructure:"l2"`
	BTCConfig `mapstructure:"btc"`
	BBNConfig `mapstructure:"bbn"`
	DBConfig  `mapstructure:"db"`
}

type FGConfig struct {
	GRPCListener         string        `long:"grpc-listener" description:"host:port to listen for gRPC connections"`
	HTTPListener         string        `long:"http-listener" description:"host:port to listen for HTTP connections"`
	StartHeight          int64         `long:"start-height" description:"Start height for indexing"`
	IndexingPollInterval time.Duration `long:"indexing-poll-interval" description:"Poll interval for indexing"`
	BlockPollInterval    time.Duration `long:"block-poll-interval" description:"Block polling interval"`
}

type L2Config struct {
	L2RPCHost string `long:"l2-rpc-host" description:"rpc host address of the L2 node"`
}

type BTCConfig struct {
	BitcoinRPCHost    string `long:"btc-rpc-host" description:"rpc host address of the bitcoin node"`
	BitcoinRPCUser    string `long:"btc-rpc-user" description:"rpc user of the bitcoin node"`
	BitcoinRPCPass    string `long:"btc-rpc-pass" description:"rpc password of the bitcoin node"`
	BitcoinDisableTLS bool   `long:"btc-disable-tls" description:"disable TLS for RPC connections"`
}

type BBNConfig struct {
	BabylonRPCAddress string `long:"bbn-rpc-addr" description:"Babylon RPC address"`
	BabylonChainId    string `long:"bbn-chain-id" description:"Babylon consumer chain ID"`
	FGContractAddress string `long:"fg-contract-address" description:"Babylon OP finality gadget contract address"`
}

type DBConfig struct {
	DBUsername      string `long:"db-username" description:"DB username"`
	DBPassword      string `long:"db-password" description:"DB password"`
	DBName          string `long:"db-name" description:"DB name"`
	DBDataPath      string `long:"db-data-path" description:"path to the DB file"`
	DBPort          uint32 `long:"db-port" description:"DB port"`
	DBBBoltFilePath string `long:"db-bbolt-file-path" description:"path to the DB bbolt file"`
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
