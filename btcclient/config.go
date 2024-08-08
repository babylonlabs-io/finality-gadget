package btcclient

import (
	"time"

	"github.com/btcsuite/btcd/rpcclient"
)

const (
	// default rpc port of signet is 38332
	defaultBitcoindRpcHost        = "rpc.ankr.com/btc"
	defaultBitcoindRPCUser        = "user"
	defaultBitcoindRPCPass        = "pass"
	defaultBitcoindBlockCacheSize = 20 * 1024 * 1024 // 20 MB
	defaultBlockPollingInterval   = 30 * time.Second
	defaultTxPollingInterval      = 30 * time.Second
	defaultMaxRetryTimes          = 5
	defaultRetryInterval          = 500 * time.Millisecond
	// DefaultTxPollingJitter defines the default TxPollingIntervalJitter
	// to be used for bitcoind backend.
	DefaultTxPollingJitter = 0.5
)

// BTCConfig defines configuration for the Bitcoin client
type BTCConfig struct {
	RPCHost              string        `long:"rpchost" description:"The daemon's rpc listening address."`
	RPCUser              string        `long:"rpcuser" description:"Username for RPC connections."`
	RPCPass              string        `long:"rpcpass" default-mask:"-" description:"Password for RPC connections."`
	PrunedNodeMaxPeers   int           `long:"pruned-node-max-peers" description:"The maximum number of peers staker will choose from the backend node to retrieve pruned blocks from. This only applies to pruned nodes."`
	BlockPollingInterval time.Duration `long:"blockpollinginterval" description:"The interval that will be used to poll bitcoind for new blocks. Only used if rpcpolling is true."`
	TxPollingInterval    time.Duration `long:"txpollinginterval" description:"The interval that will be used to poll bitcoind for new tx. Only used if rpcpolling is true."`
	BlockCacheSize       uint64        `long:"block-cache-size" description:"Size of the Bitcoin blocks cache."`
	MaxRetryTimes        uint          `long:"max-retry-times" description:"The max number of retries to an RPC call in case of failure."`
	RetryInterval        time.Duration `long:"retry-interval" description:"The time interval between each retry."`
}

func DefaultBTCConfig() *BTCConfig {
	return &BTCConfig{
		RPCHost:              defaultBitcoindRpcHost,
		RPCUser:              defaultBitcoindRPCUser,
		RPCPass:              defaultBitcoindRPCPass,
		BlockPollingInterval: defaultBlockPollingInterval,
		TxPollingInterval:    defaultTxPollingInterval,
		BlockCacheSize:       defaultBitcoindBlockCacheSize,
		MaxRetryTimes:        defaultMaxRetryTimes,
		RetryInterval:        defaultRetryInterval,
	}
}

func (cfg *BTCConfig) ToConnConfig() *rpcclient.ConnConfig {
	return &rpcclient.ConnConfig{
		Host:                 cfg.RPCHost,
		User:                 cfg.RPCUser,
		Pass:                 cfg.RPCPass,
		DisableTLS:           false,
		DisableConnectOnNew:  true,
		DisableAutoReconnect: false,
		// we use post mode as it sure it works with either bitcoind or btcwallet
		// we may need to re-consider it later if we need any notifications
		HTTPPostMode: true,
	}
}
