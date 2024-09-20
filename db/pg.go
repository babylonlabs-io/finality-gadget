package db

import (
	"context"
	"encoding/json"
	"math"

	bbntypes "github.com/babylonlabs-io/babylon/x/btcstaking/types"
	bsctypes "github.com/babylonlabs-io/babylon/x/btcstkconsumer/types"
	cfg "github.com/babylonlabs-io/finality-gadget/config"
	"github.com/babylonlabs-io/finality-gadget/types"
	epg "github.com/fergusstrange/embedded-postgres"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

type PostgresHandler struct {
	conn   *pgx.Conn
	pg     *epg.EmbeddedPostgres
	cfg    *cfg.DBConfig
	logger *zap.Logger
}

var _ IDatabaseHandler = &PostgresHandler{}

//////////////////////////////
// CONSTRUCTOR
//////////////////////////////

func NewPostgresHandler(cfg *cfg.DBConfig, logger *zap.Logger) (*PostgresHandler, error) {
	// Create embedded pg instance
	logger.Info("Creating embedded postgres...")
	pgCfg := epg.DefaultConfig().
		Username(cfg.DBUsername).
		Password(cfg.DBPassword).
		Database(cfg.DBName).
		DataPath(cfg.DBDataPath).
		Port(cfg.DBPort)
	postgres := epg.NewDatabase(pgCfg)

	// Start embedded pg
	logger.Info("Starting embedded postgres...")
	err := postgres.Start()
	if err != nil {
		logger.Error("Failed to start embedded postgres", zap.Error(err))
		return nil, err
	}

	// Shutdown embedded pg on error
	defer func() {
		if err != nil {
			logger.Error("Shutting down embedded postgres...")
			if err := postgres.Stop(); err != nil {
				logger.Error("Failed to stop embedded postgres", zap.Error(err))
			}
		}
	}()

	// Start connection to pg
	connString := pgCfg.GetConnectionURL()
	logger.Info("Connecting to postgres db...")
	conn, err := pgx.Connect(context.Background(), connString)
	if err != nil {
		logger.Error("Failed to connect to postgres", zap.Error(err))
		return nil, err
	}

	return &PostgresHandler{
		conn:   conn,
		pg:     postgres,
		cfg:    cfg,
		logger: logger,
	}, nil
}

//////////////////////////////
// METHODS
//////////////////////////////

func (pg *PostgresHandler) CreateInitialSchema() error {
	pg.logger.Info("Initialising DB...")
	_, err := pg.conn.Exec(context.Background(), sqlCreateInitialTables)
	if err != nil {
		pg.logger.Error("Failed to create initial tables", zap.Error(err))
		return err
	}
	return nil
}

func (pg *PostgresHandler) InsertBlock(block *types.Block) error {
	pg.logger.Info("Inserting block to DB", zap.Uint64("block_height", block.BlockHeight))

	_, err := pg.conn.Exec(context.Background(), sqlInsertFinalizedBlock, block.BlockHash, block.BlockHeight, block.BlockTimestamp)
	if err != nil {
		pg.logger.Error("Failed to insert block", zap.Error(err))
		return err
	}
	return nil
}

func (pg *PostgresHandler) GetBlockByHeight(height uint64) (*types.Block, error) {
	row := pg.conn.QueryRow(context.Background(), sqlQueryFinalizedBlockByHeight, height)
	var block types.Block
	err := row.Scan(&block.BlockHash, &block.BlockHeight, &block.BlockTimestamp)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, types.ErrBlockNotFound
		}
		pg.logger.Error("Failed to get finalized block by height", zap.Error(err))
		return nil, err
	}
	return &block, nil
}

func (pg *PostgresHandler) GetBlockByHash(hash string) (*types.Block, error) {
	row := pg.conn.QueryRow(context.Background(), sqlQueryFinalizedBlockByHash, hash)
	var block types.Block
	err := row.Scan(&block.BlockHash, &block.BlockHeight, &block.BlockTimestamp)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, types.ErrBlockNotFound
		}
		pg.logger.Error("Failed to get finalized block by hash", zap.Error(err))
		return nil, err
	}
	return &block, nil
}

func (pg *PostgresHandler) QueryIsBlockFinalizedByHeight(height uint64) (bool, error) {
	row := pg.conn.QueryRow(context.Background(), sqlQueryFinalizedBlockByHeight, height)
	var block types.Block
	err := row.Scan(&block.BlockHash, &block.BlockHeight, &block.BlockTimestamp)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		pg.logger.Error("Failed to query if block is finalized by height", zap.Error(err))
		return false, err
	}
	return true, nil
}

func (pg *PostgresHandler) QueryIsBlockFinalizedByHash(hash string) (bool, error) {
	row := pg.conn.QueryRow(context.Background(), sqlQueryFinalizedBlockByHash, hash)
	var block types.Block
	err := row.Scan(&block.BlockHash, &block.BlockHeight, &block.BlockTimestamp)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		pg.logger.Error("Failed to query if block is finalized by hash", zap.Error(err))
		return false, err
	}
	return true, nil
}

func (pg *PostgresHandler) QueryLatestFinalizedBlock() (*types.Block, error) {
	row := pg.conn.QueryRow(context.Background(), sqlQueryLatestFinalizedBlock)
	var block types.Block
	err := row.Scan(&block.BlockHash, &block.BlockHeight, &block.BlockTimestamp)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, types.ErrBlockNotFound
		}
		pg.logger.Error("Failed to query latest finalized block", zap.Error(err))
		return nil, err
	}
	return &block, nil
}

func (pg *PostgresHandler) GetActivatedTimestamp() (uint64, error) {
	row := pg.conn.QueryRow(context.Background(), sqlQueryActivatedTimestamp)
	var timestamp uint64
	err := row.Scan(&timestamp)
	if err != nil {
		pg.logger.Error("Failed to get activated timestamp", zap.Error(err))
		return math.MaxUint64, err
	}
	return timestamp, nil
}

func (pg *PostgresHandler) SaveActivatedTimestamp(timestamp uint64) error {
	_, err := pg.conn.Exec(context.Background(), sqlInsertActivatedTimestamp, timestamp)
	if err != nil {
		pg.logger.Error("Failed to save activated timestamp", zap.Error(err))
		return err
	}
	return nil
}

func (pg *PostgresHandler) BeginTx() (pgx.Tx, error) {
	return pg.conn.Begin(context.Background())
}

func (pg *PostgresHandler) CommitTx(tx pgx.Tx) error {
	return tx.Commit(context.Background())
}

func (pg *PostgresHandler) RollbackTx(tx pgx.Tx) error {
	return tx.Rollback(context.Background())
}

// Saves an event to `events` table
// func (pg *PostgresHandler) SaveEvent(tx pgx.Tx, evt *types.Event) error {
// 	_, err := tx.Exec(
// 		context.Background(),
// 		sqlInsertEvent,
// 		evt.TxHash,
// 		evt.Name,
// 	)
// 	if err != nil {
// 		pg.logger.Error("Failed to save event", zap.Error(err))
// 		return err
// 	}
// 	return nil
// }

// Saves initial fps
func (pg *PostgresHandler) SaveInitialFinalityProviders(fps []*bsctypes.FinalityProviderResponse) error {
	pg.logger.Info("Saving initial finality providers...")

	tx, err := pg.BeginTx()
	if err != nil {
		return err
	}

	for _, fp := range fps {
		_, err := pg.conn.Exec(
			context.Background(),
			sqlInsertInitialFinalityProvider,
			fp.Description.Moniker,
			fp.Description.Identity,
			fp.Description.Website,
			fp.Description.SecurityContact,
			fp.Description.Details,
			fp.Commission,
			fp.Addr,
			fp.BtcPk,
			fp.Pop.BtcSigType,
			fp.Pop.BtcSig,
			fp.SlashedBabylonHeight,
			fp.SlashedBtcHeight,
			fp.Height,
			fp.VotingPower,
			fp.ConsumerId,
		)
		if err != nil {
			pg.logger.Error("Failed to save event", zap.Error(err))
			pg.RollbackTx(tx)
			return err
		}
	}

	err = pg.CommitTx(tx)
	if err != nil {
		return err
	}

	return nil
}

func (pg *PostgresHandler) SaveInitialDelegations(dels []*bbntypes.BTCDelegationResponse) error {
	pg.logger.Info("Saving delegations...")

	tx, err := pg.BeginTx()
	if err != nil {
		return err
	}

	for _, del := range dels {
		fpBtcPkList, err := json.Marshal(del.FpBtcPkList)
		if err != nil {
			pg.logger.Error("Failed to marshal fpBtcPkList", zap.Error(err))
			return err
		}
		_, err = pg.conn.Exec(
			context.Background(),
			sqlInsertInitialDelegation,
			del.StakerAddr,
			del.BtcPk,
			fpBtcPkList,
			del.StartHeight,
			del.EndHeight,
			del.TotalSat,
			del.StakingTxHex,
			del.SlashingTxHex,
			del.DelegatorSlashSigHex,
			del.CovenantSigs,
			del.StakingOutputIdx,
			del.Active,
			del.StatusDesc,
			del.UnbondingTime,
			del.ParamsVersion,
		)
		if err != nil {
			// Check if the error is a duplicate entry error
			if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == pgerrcode.UniqueViolation {
				pg.logger.Warn("Duplicate entry, skipping insert", zap.Error(err))
				continue
			}
			pg.logger.Error("Failed to save event", zap.Error(err))
			pg.RollbackTx(tx)
			return err
		}
	}

	err = pg.CommitTx(tx)
	if err != nil {
		return err
	}

	return nil
}

// Saves `EventNewFinalityProvider` event
func (pg *PostgresHandler) SaveEventNewFinalityProvider(tx pgx.Tx, txInfo *types.TxInfo, evtIdx int, evt *types.EventNewFinalityProvider) error {
	_, err := tx.Exec(
		context.Background(),
		sqlInsertEventNewFinalityProvider,
		txInfo.BlockHeight,
		txInfo.BlockTimestamp,
		txInfo.TxHash,
		txInfo.TxIndex,
		evtIdx,
		evt.DescriptionMoniker,
		evt.DescriptionIdentity,
		evt.DescriptionWebsite,
		evt.DescriptionSecurityContact,
		evt.DescriptionDetails,
		evt.Commission,
		evt.BabylonPkKey,
		evt.BtcPk,
		evt.PopBtcSigType,
		evt.PopBabylonSig,
		evt.PopBtcSig,
		evt.MasterPubRand,
		evt.RegisteredEpoch,
		evt.SlashedBabylonHeight,
		evt.SlashedBtcHeight,
		evt.ConsumerId,
		evt.MsgIndex,
	)
	if err != nil {
		pg.logger.Error("Failed to save event", zap.Error(err))
		return err
	}
	return nil
}

// Saves `EventNewFinalityProvider` event
func (pg *PostgresHandler) SaveEventBTCDelegationStateUpdate(tx pgx.Tx, txInfo *types.TxInfo, evtIdx int, evt *types.EventBTCDelegationStateUpdate) error {
	_, err := tx.Exec(
		context.Background(),
		sqlInsertEventBTCDelegationStateUpdate,
		txInfo.BlockHeight,
		txInfo.BlockTimestamp,
		txInfo.TxHash,
		txInfo.TxIndex,
		evtIdx,
		evt.StakingTxHash,
		evt.NewState,
	)
	if err != nil {
		pg.logger.Error("Failed to save event", zap.Error(err))
		return err
	}
	return nil
}

func (pg *PostgresHandler) SaveEventSelectiveSlashing(tx pgx.Tx, txInfo *types.TxInfo, evtIdx int, evt *types.EventSelectiveSlashing) error {
	_, err := tx.Exec(
		context.Background(),
		sqlInsertEventSelectiveSlashing,
		txInfo.BlockHeight,
		txInfo.BlockTimestamp,
		txInfo.TxHash,
		txInfo.TxIndex,
		evtIdx,
		evt.StakingTxHash,
		evt.FpBtcPk,
		evt.RecoveredFpBtcSk,
	)
	if err != nil {
		pg.logger.Error("Failed to save event", zap.Error(err))
		return err
	}
	return nil
}

func (pg *PostgresHandler) SaveEventSlashedFinalityProvider(tx pgx.Tx, txInfo *types.TxInfo, evtIdx int, evt *types.EventSlashedFinalityProvider) error {
	_, err := tx.Exec(
		context.Background(),
		sqlInsertEventSlashedFinalityProvider,
		txInfo.BlockHeight,
		txInfo.BlockTimestamp,
		txInfo.TxHash,
		txInfo.TxIndex,
		evtIdx,
		evt.FpBtcPk,
		evt.BlockHeight,
		evt.PubRand,
		evt.CanonicalAppHash,
		evt.ForkAppHash,
		evt.CanonicalFinalitySig,
		evt.ForkFinalitySig,
	)
	if err != nil {
		pg.logger.Error("Failed to save event", zap.Error(err))
		return err
	}
	return nil
}

func (pg *PostgresHandler) Close() error {
	pg.logger.Info("Closing embedded postgres...")
	err := pg.conn.Close(context.Background())
	if err != nil {
		pg.logger.Error("Failed to close connection to postgres", zap.Error(err))
		return err
	}
	err = pg.pg.Stop()
	if err != nil {
		pg.logger.Error("Failed to stop embedded postgres", zap.Error(err))
		return err
	}
	return nil
}
