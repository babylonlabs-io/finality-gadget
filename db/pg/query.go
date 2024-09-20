package pg

//////////////////////////
// CREATE TABLES
//////////////////////////

const (
	sqlCreateInitialTables = `
    CREATE TABLE IF NOT EXISTS table_finalized_blocks (
      block_hash TEXT NOT NULL PRIMARY KEY,
      block_height BIGINT NOT NULL,
      block_timestamp BIGINT NOT NULL
    );

    CREATE TABLE IF NOT EXISTS table_activated_timestamp (
      timestamp BIGINT NOT NULL PRIMARY KEY
    );

    CREATE TABLE IF NOT EXISTS table_initial_finality_providers (
      description_moniker TEXT NOT NULL,
      description_identity TEXT NOT NULL,
      description_website TEXT NOT NULL,
      description_security_contact TEXT NOT NULL,
      description_details TEXT NOT NULL,
      commission TEXT NOT NULL,
      addr TEXT NOT NULL,
      btc_pk TEXT NOT NULL PRIMARY KEY,
      pop_btc_sig_type TEXT NOT NULL,
      pop_btc_sig TEXT NOT NULL,
      slashed_babylon_height INT NOT NULL,
      slashed_btc_height INT NOT NULL,
      height INT NOT NULL,
      voting_power INT NOT NULL,
      consumer_id TEXT NOT NULL
    );

    CREATE TABLE IF NOT EXISTS table_initial_delegations (
      staker_addr TEXT NOT NULL,
      btc_pk TEXT NOT NULL PRIMARY KEY,
      fp_btc_pk_list TEXT NOT NULL,
      start_height INT NOT NULL,
      end_height INT NOT NULL,
      total_sat INT NOT NULL,
      staking_tx_hex TEXT NOT NULL,
      slashing_tx_hex TEXT NOT NULL,
      delegator_slash_sig_hex TEXT NOT NULL,
      covenant_sigs TEXT NOT NULL,
      staking_output_idx INT NOT NULL,
      active BOOLEAN NOT NULL,
      status_desc TEXT NOT NULL,
      unbonding_time INT NOT NULL,
      -- undelegation_response omitted
      params_version INT NOT NULL
    );

    CREATE TABLE IF NOT EXISTS events_EventNewFinalityProvider (
      block_height BIGINT NOT NULL,
      block_timestamp TIMESTAMP NOT NULL,
      tx_hash TEXT NOT NULL,
      tx_index SMALLINT,
      event_index SMALLINT,
      description_moniker TEXT NOT NULL,
      description_identity TEXT NOT NULL,
      description_website TEXT NOT NULL,
      description_security_contact TEXT NOT NULL,
      description_details TEXT NOT NULL,
      commission TEXT NOT NULL,
      babylon_pk_key TEXT NOT NULL,
      btc_pk TEXT NOT NULL,
      pop_btc_sig_type TEXT NOT NULL,
      pop_babylon_sig TEXT NOT NULL,
      pop_btc_sig TEXT NOT NULL,
      master_pub_rand TEXT NOT NULL,
      registered_epoch TEXT NOT NULL,
      slashed_babylon_height TEXT NOT NULL,
      slashed_btc_height TEXT NOT NULL,
      consumer_id TEXT NOT NULL,
      msg_index TEXT NOT NULL,
      PRIMARY KEY (tx_hash, event_index)
    );

    CREATE TABLE IF NOT EXISTS events_EventBTCDelegationStateUpdate (
      block_height BIGINT NOT NULL,
      block_timestamp TIMESTAMP NOT NULL,
      tx_hash TEXT NOT NULL,
      tx_index SMALLINT,
      event_index SMALLINT,
      staking_tx_hash TEXT NOT NULL,
      new_state TEXT NOT NULL,
      PRIMARY KEY (tx_hash, event_index)
    );

    CREATE TABLE IF NOT EXISTS events_EventSelectiveSlashing (
      block_height BIGINT NOT NULL,
      block_timestamp TIMESTAMP NOT NULL,
      tx_hash TEXT NOT NULL,
      tx_index SMALLINT,
      event_index SMALLINT,
      staking_tx_hash TEXT NOT NULL,
      fp_btc_pk TEXT NOT NULL,
      recovered_fp_btc_sk TEXT NOT NULL,
      PRIMARY KEY (tx_hash, event_index)
    );

    CREATE TABLE IF NOT EXISTS events_EventMessage (
      block_height BIGINT NOT NULL,
      block_timestamp TIMESTAMP NOT NULL,
      tx_hash TEXT NOT NULL,
      tx_index SMALLINT,
      event_index SMALLINT,
      action TEXT NOT NULL,
      sender TEXT NOT NULL,
      module TEXT NOT NULL,
      msg_index TEXT NOT NULL,
      PRIMARY KEY (tx_hash, event_index)
    );
  `
)

//////////////////////////
// INSERT ENTRIES
//////////////////////////

const (
	sqlInsertFinalizedBlock = `
		INSERT INTO table_finalized_blocks (block_hash, block_height, block_timestamp) VALUES ($1, $2, $3)
	`
	sqlInsertActivatedTimestamp = `
		INSERT INTO table_activated_timestamp (timestamp) VALUES ($1)
	`
	// sqlInsertEvent = `
	// 	INSERT INTO events (tx_hash, name) VALUES ($1, $2)
	// `
	sqlInsertInitialFinalityProvider = `
    INSERT INTO table_initial_finality_providers (
      description_moniker,
      description_identity,
      description_website,
      description_security_contact,
      description_details,
      commission,
      addr,
      btc_pk,
      pop_btc_sig_type,
      pop_btc_sig,
      slashed_babylon_height
      slashed_btc_height,
      height,
      voting_power,
      consumer_id
    ) VALUES (
      $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
    )
  `
	sqlInsertInitialDelegation = `
    INSERT INTO table_initial_delegations (
      staker_addr,
      btc_pk,
      fp_btc_pk_list,
      start_height,
      end_height,
      total_sat,
      staking_tx_hex,
      slashing_tx_hex,
      delegator_slash_sig_hex,
      covenant_sigs,
      staking_output_idx,
      active,
      status_desc,
      unbonding_time,
      params_version
    ) VALUES (
      $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
    )
  `
	sqlInsertEventNewFinalityProvider = `
		INSERT INTO events_EventNewFinalityProvider (
      block_height,
      block_timestamp,
      tx_hash,
      tx_index,
      event_index,
      description_moniker,
      description_identity,
      description_website,
      description_security_contact,
      description_details,
      commission,
      babylon_pk_key,
      btc_pk,
      pop_btc_sig_type,
      pop_babylon_sig,
      pop_btc_sig,
      master_pub_rand,
      registered_epoch,
      slashed_babylon_height,
      slashed_btc_height,
      consumer_id,
      msg_index
    ) VALUES (
      $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22
    )
	`
	sqlInsertEventBTCDelegationStateUpdate = `
		INSERT INTO events_EventBTCDelegationStateUpdate (
      block_height,
      block_timestamp,
      tx_hash,
      tx_index,
      event_index,
      staking_tx_hash, 
      new_state
    ) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	sqlInsertEventSelectiveSlashing = `
		INSERT INTO events_EventSelectiveSlashing (
      block_height,
      block_timestamp,
      tx_hash,
      tx_index,
      event_index,
      staking_tx_hash, 
      fp_btc_pk, 
      recovered_fp_btc_sk
    ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	sqlInsertEventSlashedFinalityProvider = `
		INSERT INTO events_EventSlashedFinalityProvider (
      block_height,
      block_timestamp,
      tx_hash,
      tx_index,
      event_index,
      fp_btc_pk, 
      block_height, 
      pub_rand, 
      canonical_app_hash, 
      fork_app_hash, 
      canonical_finality_sig, 
      fork_finality_sig
    ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	sqlInsertEventMessage = `
    INSERT INTO events_EventMessage (
      block_height,
      block_timestamp,
      tx_hash,
      tx_index,
      event_index,  
      action, 
      sender, 
      module, 
      msg_index
    ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
  `
)

//////////////////////////
// QUERIES
//////////////////////////

const (
	sqlQueryFinalizedBlockByHeight = `
		SELECT block_hash, block_height, block_timestamp FROM table_finalized_blocks WHERE block_height = $1
	`
	sqlQueryFinalizedBlockByHash = `
		SELECT block_hash, block_height, block_timestamp FROM table_finalized_blocks WHERE block_hash = $1
	`
	sqlQueryLatestFinalizedBlock = `
		SELECT block_hash, block_height, block_timestamp FROM table_finalized_blocks ORDER BY block_height DESC LIMIT 1
	`
	sqlQueryActivatedTimestamp = `
		SELECT timestamp FROM table_activated_timestamp ORDER BY timestamp DESC LIMIT 1
	`
)
