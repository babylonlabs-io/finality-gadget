package sdk

import (
	"fmt"
	"time"

	"context"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	bbnclient "github.com/babylonchain/babylon/client/client"
	bbncfg "github.com/babylonchain/babylon/client/config"
	rpcclient "github.com/cometbft/cometbft/rpc/client"
	sdkclient "github.com/cosmos/cosmos-sdk/client"
	"go.uber.org/zap"
)

const (
	BabylonTestnet = 0
	BabylonMainnet = 1
)

// Config defines configuration for the Babylon query client
type Config struct {
	ChainType    int    `mapstructure:"chain-type"`
	ContractAddr string `mapstructure:"contract-addr"`
}

// TODO: replace with babylon RPCs when QuerySmartContractStateRequest query is supported
func (config Config) getRpcAddr() (string, error) {
	switch config.ChainType {
	case BabylonTestnet:
		return "https://sei-testnet-2-rpc.brocha.in", nil
	case BabylonMainnet:
		return "https://rpc.testnet.osmosis.zone:443", nil
	default:
		return "", fmt.Errorf("unrecognized chain type: %d", config.ChainType)
	}
}

// babylonQueryClient is a client that can only perform queries to a Babylon node
// It only requires the client config to have `rpcAddr`, but not other fields
// such as keyring, chain ID, etc..
type babylonQueryClient struct {
	// TODO: remove rpcClient after the Babylon testnet supports cw
	rpcClient rpcclient.Client
	bbnClient *bbnclient.Client
	config    *Config
}

// NewClient creates a new babylonQueryClient according to the given config
func NewClient(config Config) (*babylonQueryClient, error) {
	rpcAddr, err := config.getRpcAddr()
	if err != nil {
		return nil, err
	}

	rpcClient, err := sdkclient.NewClientFromNode(rpcAddr)
	if err != nil {
		return nil, err
	}

	bbnConfig := bbncfg.DefaultBabylonConfig()
	// TODO: replace it with config.getRpcAddr() after the Babylon testnet supports cw
	bbnConfig.RPCAddr = "https://rpc-euphrates.devnet.babylonchain.io/"

	logger, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}

	// Note: We can just ignore the below info which is printed by bbnclient.New
	// service injective.evm.v1beta1.Msg does not have cosmos.msg.v1.service proto annotation
	bbnClient, err := bbnclient.New(
		&bbnConfig,
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Babylon client: %w", err)
	}

	return &babylonQueryClient{
		rpcClient: rpcClient,
		bbnClient: bbnClient,
		config:    &config,
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
