package sdk

import (
	"time"

	"context"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	rpcclient "github.com/cometbft/cometbft/rpc/client"
	sdkclient "github.com/cosmos/cosmos-sdk/client"
)

// babylonClientConfig defines configuration for the Babylon query client
type babylonClientConfig struct {
	rpcAddr string `mapstructure:"rpc-addr"`
}

// babylonQueryClient is a client that can only perform queries to a Babylon node
// It only requires the client config to have `rpcAddr`, but not other fields
// such as keyring, chain ID, etc..
type babylonQueryClient struct {
	rpcClient rpcclient.Client
}

// newBabylonQueryClient creates a new babylonQueryClient according to the given config
func newBabylonQueryClient(queryConfig babylonClientConfig) (*babylonQueryClient, error) {
	rpcClient, err := sdkclient.NewClientFromNode(queryConfig.rpcAddr)
	if err != nil {
		return nil, err
	}

	return &babylonQueryClient{
		rpcClient: rpcClient,
	}, nil
}

// querySmartContractState queries the smart contract state given the contract address and query data
func (babylonClient *babylonQueryClient) querySmartContractState(contractAddress string, queryData []byte) (*wasmtypes.QuerySmartContractStateResponse, error) {
	// hardcode the timeout to 20 seconds. We can expose it to the params once needed
	timeout := 20 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	sdkClientCtx := sdkclient.Context{Client: babylonClient.rpcClient}
	wasmQueryClient := wasmtypes.NewQueryClient(sdkClientCtx)

	req := &wasmtypes.QuerySmartContractStateRequest{
		Address:   contractAddress,
		QueryData: queryData,
	}
	return wasmQueryClient.SmartContractState(ctx, req)
}
