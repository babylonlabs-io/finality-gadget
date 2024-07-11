CREATE TABLE IF NOT EXISTS blocks (
  block_height INTEGER PRIMARY KEY,
  block_hash TEXT NOT NULL,
  block_timestamp INTEGER NOT NULL,
  is_finalized BOOLEAN NOT NULL
);
CREATE INDEX IF NOT EXISTS blocks_hash_idx ON blocks (block_hash, block_timestamp);