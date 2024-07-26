package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type PostgresHandler struct {
	conn *pgx.Conn
}

const (
	createTablesSql = `
		CREATE TABLE IF NOT EXISTS blocks (
			block_height INTEGER PRIMARY KEY,
			block_hash TEXT NOT NULL,
			block_timestamp INTEGER NOT NULL,
			is_finalized BOOLEAN NOT NULL
		);
		CREATE INDEX IF NOT EXISTS blocks_hash_idx ON blocks (block_hash, block_timestamp);
	`
	insertBlockSql = `
		INSERT INTO blocks (block_height, block_hash, block_timestamp, is_finalized) VALUES ($1, $2, $3, $4)
	`
	getBlockStatusByHeightSql = `
		SELECT is_finalized FROM blocks WHERE block_height = $1
	`
	getBlockStatusByHashSql = `
		SELECT is_finalized FROM blocks WHERE block_hash = $1
	`
	getLatestBlockSql = `
		SELECT block_height, block_hash, block_timestamp, is_finalized FROM blocks ORDER BY block_height DESC LIMIT 1
	`
	getLatestConsecutivelyFinalizedBlockSql = `
    SELECT b.block_height, b.block_hash, b.block_timestamp, b.is_finalized
    FROM blocks b
    JOIN blocks prev ON b.block_height = prev.block_height + 1
    WHERE b.is_finalized = TRUE AND prev.is_finalized = TRUE
    ORDER BY b.block_height DESC
    LIMIT 1;
	`
)

func NewPostgresHandler(ctx context.Context, connString string) (*PostgresHandler, error) {
	conn, err := pgx.Connect(ctx, connString)
	if err != nil {
		return nil, err
	}

	return &PostgresHandler{
		conn: conn,
	}, nil
}

func (pg *PostgresHandler) TryCreateInitialTables(ctx context.Context) error {
	// Execute the SQL statements
	_, err := pg.conn.Exec(ctx, createTablesSql)
	if err != nil {
			return fmt.Errorf("unable to execute SQL statements: %v", err)
	}

	return nil
}

func (pg *PostgresHandler) InsertBlock(ctx context.Context, block Block) error {
	fmt.Printf("Inserting block %d into DB\n", block.BlockHeight)

	// Start atomic insert tx
	tx, err := pg.conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("unable to start DB tx: %v", err)
	}
	
	// Insert block
	_, err = tx.Exec(
		ctx, 
		insertBlockSql, 
		block.BlockHeight, 
		block.BlockHash, 
		block.BlockTimestamp,
		block.IsFinalized,
	)
	// Rollback tx if error
	if err != nil {
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
			return fmt.Errorf("unable to insert block: %v, unable to rollback transaction: %v", err, rollbackErr)
		}
		return fmt.Errorf("unable to insert block: %v", err)
	}

	// Commit tx if no error
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("unable to commit transaction: %v", err)
	}

	return nil
}

func (pg *PostgresHandler) GetBlockStatusByHeight(ctx context.Context, blockHeight uint64) bool {
	var isFinalized bool
	err := pg.conn.QueryRow(ctx, getBlockStatusByHeightSql, blockHeight).Scan(&isFinalized)
	if err != nil {
		fmt.Printf("GetBlockStatusByHeight: %v\n", err)
		return false
	}

	return isFinalized
}

func (pg *PostgresHandler) GetBlockStatusByHash(ctx context.Context, blockHash string) bool {
	var isFinalized bool
	err := pg.conn.QueryRow(ctx, getBlockStatusByHashSql, blockHash).Scan(&isFinalized) 

	if err != nil {
		fmt.Printf("GetBlockStatusByHash: %v\n", err)
		return false
	}

	return isFinalized
}

func (pg *PostgresHandler) GetLatestBlock(ctx context.Context) (*Block, error) {
	var block Block
	err := pg.conn.QueryRow(ctx, getLatestBlockSql).Scan(&block.BlockHeight, &block.BlockHash, &block.BlockTimestamp, &block.IsFinalized)
	if err != nil {
		return nil, fmt.Errorf("unable to get latest block: %v", err)
	}

	return &block, nil
}

func (pg *PostgresHandler) GetLatestConsecutivelyFinalizedBlock(ctx context.Context) (*Block, error) {
	var block Block
	err := pg.conn.QueryRow(ctx, getLatestConsecutivelyFinalizedBlockSql).Scan(&block.BlockHeight, &block.BlockHash, &block.BlockTimestamp, &block.IsFinalized)
	if err != nil {
		return nil, fmt.Errorf("unable to get latest consecutively finalized block: %v", err)
	}

	return &block, nil
}

func (pg *PostgresHandler) Close(ctx context.Context) {
	pg.conn.Close(ctx)
}