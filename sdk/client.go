package sdk

import (
	"fmt"
	"time"

	"context"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	bbnclient "github.com/babylonchain/babylon/client/client"
	bbncfg "github.com/babylonchain/babylon/client/config"
	sdkclient "github.com/cosmos/cosmos-sdk/client"
	"go.uber.org/zap"

	"github.com/babylonchain/babylon-finality-gadget/btcclient"
	"github.com/babylonchain/babylon-finality-gadget/testutils"
)

type BtcClient interface {
	GetBlockHeightByTimestamp(targetTimestamp uint64) (uint64, error)
}

const (
	BabylonLocalnet = -1
	BabylonTestnet  = 0
	BabylonMainnet  = 1
)

// Config defines configuration for the Babylon query client
type Config struct {
	BTCConfig    *btcclient.BTCConfig `mapstructure:"btc-config"`
	ContractAddr string               `mapstructure:"contract-addr"`
	ChainType    int                  `mapstructure:"chain-type"`
}

func (config *Config) getRpcAddr() (string, error) {
	switch config.ChainType {
	case BabylonLocalnet:
		// only for the e2e test
		return "http://127.0.0.1:26657", nil
	case BabylonTestnet:
		return "https://rpc-euphrates.devnet.babylonchain.io/", nil
	// TODO: replace with babylon RPCs when QuerySmartContractStateRequest query is supported
	case BabylonMainnet:
		return "https://rpc-euphrates.devnet.babylonchain.io/", nil
	default:
		return "", fmt.Errorf("unrecognized chain type: %d", config.ChainType)
	}
}

// BabylonFinalityGadgetClient is a client that can only perform queries to a Babylon node
// It only requires the client config to have `rpcAddr`, but not other fields
// such as keyring, chain ID, etc..
type BabylonFinalityGadgetClient struct {
	config    *Config
	bbnClient *bbnclient.Client
	BtcClient BtcClient
}

// NewClient creates a new BabylonFinalityGadgetClient according to the given config
func NewClient(config *Config) (*BabylonFinalityGadgetClient, error) {
	rpcAddr, err := config.getRpcAddr()
	if err != nil {
		return nil, err
	}

	bbnConfig := bbncfg.DefaultBabylonConfig()
	bbnConfig.RPCAddr = rpcAddr

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

	var btcClient BtcClient
	// Create BTC client
	switch config.ChainType {
	case BabylonLocalnet:
		btcClient, err = testutils.NewMockBTCClient(config.BTCConfig, logger)
	default:
		btcClient, err = btcclient.NewBTCClient(config.BTCConfig, logger)
	}
	if err != nil {
		return nil, err
	}
	return &BabylonFinalityGadgetClient{
		bbnClient: bbnClient,
		config:    config,
		BtcClient: btcClient,
	}, nil
}

// querySmartContractState queries the smart contract state given the contract address and query data
func (babylonClient *BabylonFinalityGadgetClient) querySmartContractState(contractAddress string, queryData []byte) (*wasmtypes.QuerySmartContractStateResponse, error) {
	// hardcode the timeout to 20 seconds. We can expose it to the params once needed
	timeout := 20 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	sdkClientCtx := sdkclient.Context{Client: babylonClient.bbnClient.RPCClient}
	wasmQueryClient := wasmtypes.NewQueryClient(sdkClientCtx)

	req := &wasmtypes.QuerySmartContractStateRequest{
		Address:   contractAddress,
		QueryData: queryData,
	}
	return wasmQueryClient.SmartContractState(ctx, req)
}
