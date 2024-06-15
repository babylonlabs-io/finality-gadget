package sdk

import (
	"time"

	rpcclient "github.com/cometbft/cometbft/rpc/client"
	sdkclient "github.com/cosmos/cosmos-sdk/client"
)

// BabylonQueryConfig defines configuration for the Babylon query client
type BabylonQueryConfig struct {
	RPCAddr string        `mapstructure:"rpc-addr"`
	Timeout time.Duration `mapstructure:"timeout"`
}

// babylonQueryClient is a client that can only perform queries to a Babylon node
// It only requires the query config to have `Timeout` and `RPCAddr`, but not other fields
// such as keyring, chain ID, etc..
type babylonQueryClient struct {
	RPCClient rpcclient.Client
	timeout   time.Duration
}

// newBabylonQueryClient creates a new babylonQueryClient according to the given config
func newBabylonQueryClient(queryConfig BabylonQueryConfig) (*babylonQueryClient, error) {
	rpcClient, err := sdkclient.NewClientFromNode(queryConfig.RPCAddr)
	if err != nil {
		return nil, err
	}

	return &babylonQueryClient{
		RPCClient: rpcClient,
		timeout:   queryConfig.Timeout,
	}, nil
}
