package client

import (
	"fmt"

	bbncfg "github.com/babylonchain/babylon/client/config"
	"go.uber.org/zap"

	"github.com/babylonchain/babylon-finality-gadget/sdk/btcclient"
	sdkconfig "github.com/babylonchain/babylon-finality-gadget/sdk/config"

	babylonClient "github.com/babylonchain/babylon/client/client"

	"github.com/babylonchain/babylon-finality-gadget/sdk/bbnclient"
	"github.com/babylonchain/babylon-finality-gadget/sdk/cwclient"
	"github.com/babylonchain/babylon-finality-gadget/testutils"
)

type BtcClient interface {
	GetBlockHeightByTimestamp(targetTimestamp uint64) (uint64, error)
}

// SdkClient is a client that can only perform queries to a Babylon node
// It only requires the client config to have `rpcAddr`, but not other fields
// such as keyring, chain ID, etc..
type SdkClient struct {
	config    *sdkconfig.Config
	bbnClient *bbnclient.Client
	cwClient  *cwclient.Client
	btcClient BtcClient
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

	var btcClient BtcClient
	// Create BTC client
	switch config.ChainID {
	case sdkconfig.BabylonLocalnet:
		btcClient, err = testutils.NewMockBTCClient(config.BTCConfig, logger)
	default:
		btcClient, err = btcclient.NewBTCClient(config.BTCConfig, logger)
	}
	if err != nil {
		return nil, err
	}

	cwClient := cwclient.NewClient(babylonClient.QueryClient.RPCClient, config.ContractAddr)

	return &SdkClient{
		config:    config,
		bbnClient: &bbnclient.Client{QueryClient: babylonClient.QueryClient},
		cwClient:  cwClient,
		btcClient: btcClient,
	}, nil
}
