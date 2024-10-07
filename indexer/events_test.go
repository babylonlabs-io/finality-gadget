package indexer

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	cosmossdk_io_math "cosmossdk.io/math"
	bbntypes "github.com/babylonlabs-io/babylon/types"
	bsctypes "github.com/babylonlabs-io/babylon/x/btcstkconsumer/types"
	ctypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	cfg "github.com/babylonlabs-io/finality-gadget/config"
	"github.com/babylonlabs-io/finality-gadget/db/pg"
	"github.com/babylonlabs-io/finality-gadget/log"
	"github.com/babylonlabs-io/finality-gadget/types"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

///////////////////////////
// SETUP
///////////////////////////

func setup(t *testing.T) (*Indexer, func()) {
	// Create temp test file.
	tempFile, err := os.CreateTemp("", "test-*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tempFile.Close()

	// Create logger.
	logger, err := log.NewRootLogger("console", true)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Define configs
	dbCfg := cfg.DBConfig{
		DBName:     "test",
		DBUsername: "test",
		DBPassword: "test",
		DBDataPath: tempFile.Name(),
		DBPort:     5433,
	}
	bbnCfg := cfg.BBNConfig{
		BabylonChainId:    "test-chain",
		BabylonRPCAddress: "https://rpc-euphrates.devnet.babylonchain.io",
	}

	// Create a new Postgres handler
	db, err := pg.NewPostgresHandler(&dbCfg, logger)
	if err != nil {
		t.Fatalf("Failed to create PostgresHandler: %v", err)
	}

	// Create initial buckets
	err = db.CreateInitialSchema()
	if err != nil {
		t.Fatalf("Failed to create initial schema: %v", err)
	}

	// Create indexer
	idx, err := NewIndexer(&bbnCfg, db, logger)
	if err != nil {
		t.Fatalf("error creating indexer: %v", err)
	}

	// Cleanup function to close DB and remove temp file
	cleanup := func() {
		if err := db.Close(); err != nil {
			t.Fatalf("Failed to close DB: %v", err)
		}
		err := os.RemoveAll(tempFile.Name())
		if err != nil {
			t.Fatalf("Failed to delete DB: %v", err)
		}
	}

	// Setup signal handling for cleanup on interrupt
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-signalChan
		cleanup()
		os.Exit(1) // Exit after cleanup
	}()

	return idx, cleanup
}

///////////////////////////
// TESTS
///////////////////////////

func TestSaveAndQueryInitialFinalityProvider(t *testing.T) {
	idx, cleanup := setup(t)
	defer cleanup()

	// Define initial fps
	mockInitialFp := mockInitialFp(t)
	initialFps := []*bsctypes.FinalityProviderResponse{
		&mockInitialFp,
	}

	// Save initial fps
	err := idx.db.SaveInitialFinalityProviders(initialFps)

	// get initial fps
	fps, err := idx.db.GetInitialFinalityProviders()
	idx.logger.Info("Fetched fps", zap.Any("fps", fps))
	require.NoError(t, err)

	// run checks
	require.Equal(t, 1, len(fps))
	require.Equal(t, mockInitialFp.Description.Moniker, fps[0].DescriptionMoniker)
	require.Equal(t, mockInitialFp.Description.Identity, fps[0].DescriptionIdentity)
	require.Equal(t, mockInitialFp.Description.Website, fps[0].DescriptionWebsite)
	require.Equal(t, mockInitialFp.Description.SecurityContact, fps[0].DescriptionSecurityContact)
	require.Equal(t, mockInitialFp.Description.Details, fps[0].DescriptionDetails)
	require.Equal(t, mockInitialFp.Commission, fps[0].Commission)
	require.Equal(t, mockInitialFp.Addr, fps[0].Addr)
	require.Equal(t, mockInitialFp.BtcPk, fps[0].BtcPk)
	require.Equal(t, mockInitialFp.SlashedBabylonHeight, fps[0].SlashedBabylonHeight)
	require.Equal(t, mockInitialFp.SlashedBtcHeight, fps[0].SlashedBtcHeight)
	require.Equal(t, mockInitialFp.ConsumerId, fps[0].ConsumerId)
}

func TestSaveAndQueryNewFinalityProvider(t *testing.T) {
	idx, cleanup := setup(t)
	defer cleanup()

	// Define mock data
	mockNewFp := mockNewFp()
	newFpBytes, err := json.Marshal(mockNewFp)
	require.NoError(t, err)

	txInfo := types.TxInfo{
		BlockHeight:    1,
		BlockTimestamp: time.Unix(1000, 0),
		TxHash:         "0x123",
		TxIndex:        1,
	}
	evts := []EventWithTxInfo{
		{
			TxInfo: &txInfo,
			Event: &Event{
				Type:  "babylon.btcstaking.v1.EventNewFinalityProvider",
				Index: 1,
				Attributes: []EventAttribute{
					{
						Key:   "fp",
						Value: newFpBytes,
					},
					{
						Key:   "msg_index",
						Value: []byte("1"),
					},
				},
			},
		},
	}

	// Process mock events.
	ctx := context.Background()
	err = idx.ProcessMockEvents(ctx, evts)
	require.NoError(t, err)

	// get fps
	fps, err := idx.db.GetFinalityProvidersAtHeight(uint64(txInfo.BlockHeight))
	idx.logger.Info("Fetched fps", zap.Any("fps", fps))
	require.NoError(t, err)

	// run checks
	require.Equal(t, 1, len(fps))
	require.Equal(t, mockNewFp.Description.Moniker, fps[0].DescriptionMoniker)
	require.Equal(t, mockNewFp.Description.Identity, fps[0].DescriptionIdentity)
	require.Equal(t, mockNewFp.Description.Website, fps[0].DescriptionWebsite)
	require.Equal(t, mockNewFp.Description.SecurityContact, fps[0].DescriptionSecurityContact)
	require.Equal(t, mockNewFp.Description.Details, fps[0].DescriptionDetails)
	require.Equal(t, mockNewFp.Commission, fps[0].Commission)
	require.Equal(t, mockNewFp.BabylonPk.Key, fps[0].BabylonPkKey)
	require.Equal(t, mockNewFp.BtcPk, fps[0].BtcPk)
	require.Equal(t, mockNewFp.SlashedBabylonHeight, fps[0].SlashedBabylonHeight)
	require.Equal(t, mockNewFp.SlashedBtcHeight, fps[0].SlashedBtcHeight)
	require.Equal(t, mockNewFp.ConsumerId, fps[0].ConsumerId)
}

///////////////////////////
// INTERNAL
///////////////////////////

func mockInitialFp(t *testing.T) bsctypes.FinalityProviderResponse {
	description := &ctypes.Description{
		Moniker:         "Test FP",
		Identity:        "test fp",
		Website:         "https://test.com",
		SecurityContact: "test@test.com",
		Details:         "test details",
	}
	commission := cosmossdk_io_math.NewIntWithDecimal(1, 1).ToLegacyDec()
	btcPk, err := bbntypes.NewBIP340PubKeyFromHex("146e665eb6220a4a9db29f4bdf474af014f73ace48b959a098483774af490cc1")
	require.NoError(t, err)

	return bsctypes.FinalityProviderResponse{
		Description:          description,
		Commission:           &commission,
		Addr:                 "tT3pEG0D5mhEZ2bwiEVhRidQFSpD65ns5VJmD8GPeLA",
		BtcPk:                btcPk,
		Pop:                  nil,
		SlashedBabylonHeight: 0,
		SlashedBtcHeight:     0,
		Height:               1000,
		VotingPower:          150000,
		ConsumerId:           "test-chain",
	}
}

func mockNewFp() types.FinalityProvider {
	description := &ctypes.Description{
		Moniker:         "Test FP",
		Identity:        "test fp",
		Website:         "https://test.com",
		SecurityContact: "test@test.com",
		Details:         "test details",
	}
	babylonPk := &types.FinalityProviderBabylonPk{
		Key: "tT3pEG0D5mhEZ2bwiEVhRidQFSpD65ns5VJmD8GPeLA",
	}
	return types.FinalityProvider{
		Description:          description,
		Commission:           "0.1",
		BabylonPk:            babylonPk,
		BtcPk:                "146e665eb6220a4a9db29f4bdf474af014f73ace48b959a098483774af490cc1",
		SlashedBabylonHeight: "0",
		SlashedBtcHeight:     "0",
		ConsumerId:           "test-chain",
	}
}
