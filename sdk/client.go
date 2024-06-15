package sdk

import (
	"time"

	"context"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	rpcclient "github.com/cometbft/cometbft/rpc/client"
	sdkclient "github.com/cosmos/cosmos-sdk/client"
)

// babylonQueryConfig defines configuration for the Babylon query client
type babylonQueryConfig struct {
	rpcAddr string        `mapstructure:"rpc-addr"`
	timeout time.Duration `mapstructure:"timeout"`
}

// babylonQueryClient is a client that can only perform queries to a Babylon node
// It only requires the query config to have `Timeout` and `RPCAddr`, but not other fields
// such as keyring, chain ID, etc..
type babylonQueryClient struct {
	rpcClient rpcclient.Client
	timeout   time.Duration
}

// newBabylonQueryClient creates a new babylonQueryClient according to the given config
func newBabylonQueryClient(queryConfig babylonQueryConfig) (*babylonQueryClient, error) {
	rpcClient, err := sdkclient.NewClientFromNode(queryConfig.rpcAddr)
	if err != nil {
		return nil, err
	}

	return &babylonQueryClient{
		rpcClient: rpcClient,
		timeout:   queryConfig.timeout,
	}, nil
}

// querySmartContractState queries the smart contract state given the contract address and query data
func (babylonClient *babylonQueryClient) querySmartContractState(contractAddress string, queryData []byte) (*wasmtypes.QuerySmartContractStateResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), babylonClient.timeout)
	defer cancel()

	sdkClientCtx := sdkclient.Context{Client: babylonClient.rpcClient}
	wasmQueryClient := wasmtypes.NewQueryClient(sdkClientCtx)

	req := &wasmtypes.QuerySmartContractStateRequest{
		Address:   contractAddress,
		QueryData: queryData,
	}
	return wasmQueryClient.SmartContractState(ctx, req)
}
