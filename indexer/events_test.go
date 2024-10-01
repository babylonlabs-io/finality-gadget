package indexer

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	cfg "github.com/babylonlabs-io/finality-gadget/config"
	"github.com/babylonlabs-io/finality-gadget/db/pg"
	"github.com/babylonlabs-io/finality-gadget/log"
	"github.com/babylonlabs-io/finality-gadget/types"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

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

func TestParseEventNewFinalityProvider(t *testing.T) {
	idx, cleanup := setup(t)
	defer cleanup()

	// Define mock data
	mockFp := mockFp()
	fpBytes, err := json.Marshal(mockFp)
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
						Value: fpBytes,
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
	fps, err := idx.db.GetFinalityProviders(uint64(txInfo.BlockHeight))
	idx.logger.Info("Fetched fps", zap.Any("fps", fps))
	require.NoError(t, err)

	// run checks
	require.Equal(t, 1, len(fps))
	require.Equal(t, mockFp.Description.Moniker, fps[0].DescriptionMoniker)
	require.Equal(t, mockFp.Description.Identity, fps[0].DescriptionIdentity)
	require.Equal(t, mockFp.Description.Website, fps[0].DescriptionWebsite)
	require.Equal(t, mockFp.Description.SecurityContact, fps[0].DescriptionSecurityContact)
	require.Equal(t, mockFp.Description.Details, fps[0].DescriptionDetails)
	require.Equal(t, mockFp.Commission, fps[0].Commission)
	require.Equal(t, mockFp.BabylonPk.Key, fps[0].BabylonPkKey)
	require.Equal(t, mockFp.BtcPk, fps[0].BtcPk)
	require.Equal(t, mockFp.SlashedBabylonHeight, fps[0].SlashedBabylonHeight)
	require.Equal(t, mockFp.SlashedBtcHeight, fps[0].SlashedBtcHeight)
	require.Equal(t, mockFp.ConsumerId, fps[0].ConsumerId)
}

// Internal helpers

func mockFp() types.FinalityProvider {
	return types.FinalityProvider{
		Description:          mockFpDescription(),
		Commission:           "0.1",
		BabylonPk:            mockBabylonPk(),
		BtcPk:                "146e665eb6220a4a9db29f4bdf474af014f73ace48b959a098483774af490cc1",
		SlashedBabylonHeight: "0",
		SlashedBtcHeight:     "0",
		ConsumerId:           "test-chain",
	}
}

func mockFpDescription() types.FinalityProviderDescription {
	return types.FinalityProviderDescription{
		Moniker:         "Test FP",
		Identity:        "test fp",
		Website:         "https://test.com",
		SecurityContact: "test@test.com",
		Details:         "test details",
	}
}

func mockBabylonPk() types.FinalityProviderBabylonPk {
	return types.FinalityProviderBabylonPk{
		Key: "tT3pEG0D5mhEZ2bwiEVhRidQFSpD65ns5VJmD8GPeLA",
	}
}
