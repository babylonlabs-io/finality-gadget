package db

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
)

type PostgresHandler struct {
	conn *pgx.Conn
}

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
	// Load the migration file
	filePath := "db/migrations/0000_initial_tables.sql"

	sqlBytes, err := os.ReadFile(filePath)
	if err != nil {
			return fmt.Errorf("unable to read SQL file: %v", err)
	}
	sql := string(sqlBytes)

	// Execute the SQL statements
	_, err = pg.conn.Exec(ctx, sql)
	if err != nil {
			return fmt.Errorf("unable to execute SQL statements: %v", err)
	}

	return nil
}

func (pg *PostgresHandler) InsertBlock(ctx context.Context, block Block) error {
	fmt.Printf("Inserting block %d into DB\n", block.BlockHeight)
	_, err := pg.conn.Exec(
		ctx, 
		"INSERT INTO blocks (block_height, block_hash, block_timestamp, is_finalized) VALUES ($1, $2, $3, $4)", 
		block.BlockHeight, 
		block.BlockHash, 
		block.BlockTimestamp,
		block.IsFinalized,
	)

	if err != nil {
		return fmt.Errorf("unable to insert block: %v", err)
	}

	return nil
}

func (pg *PostgresHandler) GetBlockStatusByHeight(ctx context.Context, blockHeight uint64) bool {
	query := `SELECT is_finalized FROM blocks WHERE block_height = $1`

	var isFinalized bool
	err := pg.conn.QueryRow(ctx, query, blockHeight).Scan(&isFinalized)
	if err != nil {
		fmt.Printf("GetBlockStatusByHeight: %v\n", err)
		return false
	}

	return isFinalized
}

func (pg *PostgresHandler) GetBlockStatusByHash(ctx context.Context, blockHash string) bool {
	query := `SELECT is_finalized FROM blocks WHERE block_hash = $1`

	var isFinalized bool
	err := pg.conn.QueryRow(ctx, query, blockHash).Scan(&isFinalized) 

	if err != nil {
		fmt.Printf("GetBlockStatusByHash: %v\n", err)
		return false
	}

	return isFinalized
}

func (pg *PostgresHandler) GetLatestConsecutivelyFinalizedBlock(ctx context.Context) (*Block, error) {
	query := `
    SELECT b.block_height, b.block_hash, b.block_timestamp, b.is_finalized
    FROM blocks b
    JOIN blocks prev ON b.block_height = prev.block_height + 1
    WHERE b.is_finalized = TRUE AND prev.is_finalized = TRUE
    ORDER BY b.block_height DESC
    LIMIT 1;
	`

	var block Block
	err := pg.conn.QueryRow(ctx, query).Scan(&block.BlockHeight, &block.BlockHash, &block.BlockTimestamp, &block.IsFinalized)
	if err != nil {
		return nil, fmt.Errorf("unable to get latest consecutively finalized block: %v", err)
	}

	return &block, nil
}

func (pg *PostgresHandler) Close(ctx context.Context) {
	pg.conn.Close(ctx)
}