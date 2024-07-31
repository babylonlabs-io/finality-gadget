package client

import (
	"fmt"

	bbncfg "github.com/babylonchain/babylon/client/config"
	"go.uber.org/zap"

	"github.com/babylonchain/babylon-finality-gadget/sdk/bbnclient"
	"github.com/babylonchain/babylon-finality-gadget/sdk/btcclient"
	sdkconfig "github.com/babylonchain/babylon-finality-gadget/sdk/config"
	"github.com/babylonchain/babylon-finality-gadget/testutil"

	babylonClient "github.com/babylonchain/babylon/client/client"

	"github.com/babylonchain/babylon-finality-gadget/sdk/cwclient"
)

// SdkClient is a client that can only perform queries to a Babylon node
// It only requires the client config to have `rpcAddr`, but not other fields
// such as keyring, chain ID, etc..
type SdkClient struct {
	bbnClient IBabylonClient
	cwClient  ICosmWasmClient
	btcClient IBitcoinClient
}

// NewClient creates a new BabylonFinalityGadgetClient according to the given config
func NewClient(config *sdkconfig.Config) (*SdkClient, error) {
	rpcAddr, err := config.GetRpcAddr()
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
	babylonClient, err := babylonClient.New(
		&bbnConfig,
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Babylon client: %w", err)
	}

	var btcClient IBitcoinClient
	// Create BTC client
	switch config.ChainID {
	// TODO: once we set up our own local BTC devnet, we don't need to use this mock BTC client
	case sdkconfig.BabylonLocalnet:
		btcClient, err = testutil.NewMockBTCClient(config.BTCConfig, logger)
	default:
		btcClient, err = btcclient.NewBTCClient(config.BTCConfig, logger)
	}
	if err != nil {
		return nil, err
	}

	cwClient := cwclient.NewClient(babylonClient.QueryClient.RPCClient, config.ContractAddr)

	return &SdkClient{
		bbnClient: &bbnclient.Client{QueryClient: babylonClient.QueryClient},
		cwClient:  cwClient,
		btcClient: btcClient,
	}, nil
}
