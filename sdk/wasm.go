package sdk

import (
	"context"
	"strconv"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdkclient "github.com/cosmos/cosmos-sdk/client"
	grpctypes "github.com/cosmos/cosmos-sdk/types/grpc"
	"google.golang.org/grpc/metadata"
)

// getBabylonQueryContext returns a context that includes the height and uses the timeout from the config
// (adapted from https://github.com/strangelove-ventures/lens/blob/v0.5.4/client/query/query_options.go#L29-L36)
// TODO: understand the func
func (babylonClient *babylonQueryClient) getBabylonQueryContext() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), babylonClient.timeout)
	height := 0
	strHeight := strconv.Itoa(height)
	ctx = metadata.AppendToOutgoingContext(ctx, grpctypes.GRPCBlockHeightHeader, strHeight)
	return ctx, cancel
}

func (babylonClient *babylonQueryClient) querySmartContractState(contractAddress string, queryData string) (*wasmtypes.QuerySmartContractStateResponse, error) {
	ctx, cancel := babylonClient.getBabylonQueryContext()
	defer cancel()

	sdkClientCtx := sdkclient.Context{Client: babylonClient.rpcClient}
	wasmQueryClient := wasmtypes.NewQueryClient(sdkClientCtx)

	req := &wasmtypes.QuerySmartContractStateRequest{
		Address:   contractAddress,
		QueryData: wasmtypes.RawContractMessage(queryData),
	}
	return wasmQueryClient.SmartContractState(ctx, req)
}
