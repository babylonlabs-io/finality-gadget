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
	bstypes "github.com/babylonlabs-io/babylon/x/btcstaking/types"
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
	require.NoError(t, err)

	// get initial fps
	fps, err := idx.db.GetInitialFinalityProviders()
	idx.logger.Info("Fetched fps", zap.Any("fps", fps))
	require.NoError(t, err)

	// run checks
	require.Equal(t, 1, len(fps))
	checkInitialFpsEqual(t, mockInitialFp, fps[0])
}

func TestSaveAndQueryNewFinalityProvider(t *testing.T) {
	idx, cleanup := setup(t)
	defer cleanup()

	// Define mock data
	mockNewFp := mockNewFp(t)
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
	require.Equal(t, mockNewFp, fps[0])
}

func TestSaveAndQueryFinalityProviderAtHeight(t *testing.T) {
	idx, cleanup := setup(t)
	defer cleanup()

	// Define mock data
	mockInitialFp := mockInitialFp(t)
	initialFps := []*bsctypes.FinalityProviderResponse{
		&mockInitialFp,
	}
	mockNewFp := mockNewFp(t)
	newFpBytes, err := json.Marshal(mockNewFp)
	require.NoError(t, err)

	txInfo := types.TxInfo{
		BlockHeight:    5,
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

	// Save initial fps
	err = idx.db.SaveInitialFinalityProviders(initialFps)
	require.NoError(t, err)
	err = idx.ProcessMockEvents(context.Background(), evts)
	require.NoError(t, err)

	// get fps at height 1
	fps1, err := idx.db.GetFinalityProvidersAtHeight(1)
	require.NoError(t, err)

	// get fps at height 5
	fps5, err := idx.db.GetFinalityProvidersAtHeight(5)
	require.NoError(t, err)

	// run checks
	require.Equal(t, 1, len(fps1))
	require.Equal(t, mockInitialFp, fps1[0])
	require.Equal(t, 2, len(fps5))
	require.Equal(t, mockInitialFp, fps5[0])
	require.Equal(t, mockNewFp, fps5[1])
}

///////////////////////////
// INTERNAL
///////////////////////////

func mockInitialFp(t *testing.T) bsctypes.FinalityProviderResponse {
	description := mockDescription()
	commission := cosmossdk_io_math.NewIntWithDecimal(1, 1).ToLegacyDec()
	btcPk := mockBtcPk(t, "146e665eb6220a4a9db29f4bdf474af014f73ace48b959a098483774af490cc1")
	pop := mockPop(t)

	return bsctypes.FinalityProviderResponse{
		Description:          description,
		Commission:           &commission,
		Addr:                 "tT3pEG0D5mhEZ2bwiEVhRidQFSpD65ns5VJmD8GPeLA",
		BtcPk:                btcPk,
		Pop:                  pop,
		SlashedBabylonHeight: 0,
		SlashedBtcHeight:     0,
		Height:               1000,
		VotingPower:          150000,
		ConsumerId:           "test-chain",
	}
}

func mockNewFp(t *testing.T) bstypes.FinalityProvider {
	description := mockDescription()
	commission := cosmossdk_io_math.NewIntWithDecimal(1, 1).ToLegacyDec()
	btcPk := mockBtcPk(t, "146e665eb6220a4a9db29f4bdf474af014f73ace48b959a098483774af490cc1")
	pop := mockPop(t)

	return bstypes.FinalityProvider{
		Addr:                 "tT3pEG0D5mhEZ2bwiEVhRidQFSpD65ns5VJmD8GPeLA",
		Description:          description,
		Commission:           &commission,
		BtcPk:                btcPk,
		Pop:                  pop,
		SlashedBabylonHeight: 0,
		SlashedBtcHeight:     0,
		Jailed:               false,
		ConsumerId:           "test-chain",
	}
}

func mockDescription() *ctypes.Description {
	return &ctypes.Description{
		Moniker:         "Test FP",
		Identity:        "test fp",
		Website:         "https://test.com",
		SecurityContact: "test@test.com",
		Details:         "test details",
	}
}

func mockBtcPk(t *testing.T, hex string) *bbntypes.BIP340PubKey {
	btcPk, err := bbntypes.NewBIP340PubKeyFromHex(hex)
	require.NoError(t, err)
	return btcPk
}

func mockPop(t *testing.T) *bstypes.ProofOfPossessionBTC {
	pop := bstypes.ProofOfPossessionBTC{
		BtcSigType: 0,
		BtcSig:     []byte("test sig"),
	}
	return &pop
}

func checkInitialFpsEqual(t *testing.T, fp1 bsctypes.FinalityProviderResponse, fp2 *types.FinalityProvider) {
	require.Equal(t, fp1.Addr, fp2.Addr)
	require.Equal(t, fp1.Description.Moniker, fp2.DescriptionMoniker)
	require.Equal(t, fp1.Description.Identity, fp2.DescriptionIdentity)
	require.Equal(t, fp1.Description.Website, fp2.DescriptionWebsite)
	require.Equal(t, fp1.Description.SecurityContact, fp2.DescriptionSecurityContact)
	require.Equal(t, fp1.Description.Details, fp2.DescriptionDetails)
	require.Equal(t, fp1.Commission.BigInt().String(), fp2.Commission)
	require.Equal(t, fp1.BtcPk.MarshalHex(), fp2.BtcPk)
	require.Equal(t, fp1.SlashedBabylonHeight, fp2.SlashedBabylonHeight)
	require.Equal(t, fp1.SlashedBtcHeight, fp2.SlashedBtcHeight)
	require.Equal(t, fp1.ConsumerId, fp2.ConsumerId)
}
