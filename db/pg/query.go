package pg

//////////////////////////
// CREATE TYPES
//////////////////////////

const (
	sqlCreateTypeBTCDelegationStatus = `
		CREATE TYPE BTCDelegationStatus AS ENUM ('PENDING', 'ACTIVE', 'UNBONDED');
	`
)

//////////////////////////
// CREATE TABLES
//////////////////////////

const (
	sqlCreateInitialTables = `
    CREATE TABLE IF NOT EXISTS table_chain_params (
      k_value INT NOT NULL,
      w_value INT NOT NULL,
      cov_quorum INT NOT NULL
    );

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
      slashed_babylon_height BIGINT NOT NULL,
      slashed_btc_height BIGINT NOT NULL,
      consumer_id TEXT NOT NULL
    );

    CREATE TABLE IF NOT EXISTS table_initial_delegations (
      staker_addr TEXT NOT NULL,
      btc_pk TEXT NOT NULL PRIMARY KEY,
      fp_btc_pk_list TEXT NOT NULL,
      start_height BIGINT NOT NULL,
      end_height BIGINT NOT NULL,
      total_sat BIGINT NOT NULL,
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

    CREATE TABLE IF NOT EXISTS table_btc_delegations (
      staker_addr TEXT NOT NULL,
      btc_pk TEXT NOT NULL PRIMARY KEY,
      fp_btc_pk_list TEXT NOT NULL,
      start_height BIGINT NOT NULL,
      end_height BIGINT NOT NULL,
      total_sat BIGINT NOT NULL,
      staking_tx_hex TEXT NOT NULL,
      slashing_tx_hex TEXT NOT NULL,
      delegator_slash_sig_hex TEXT NOT NULL,
      num_covenant_sigs INT NOT NULL,
      staking_output_idx INT NOT NULL,
      active BOOLEAN NOT NULL,
      status_desc TEXT NOT NULL,
      unbonding_time INT NOT NULL,
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
      slashed_babylon_height BIGINT NOT NULL,
      slashed_btc_height BIGINT NOT NULL,
      consumer_id TEXT NOT NULL,
      PRIMARY KEY (tx_hash, event_index)
    );

    CREATE TABLE IF NOT EXISTS events_EventBTCDelegationStateUpdate (
      block_height BIGINT NOT NULL,
      block_timestamp TIMESTAMP NOT NULL,
      tx_hash TEXT NOT NULL,
      tx_index SMALLINT,
      event_index SMALLINT,
      staking_tx_hash TEXT NOT NULL,
      new_state BTCDelegationStatus NOT NULL,
      PRIMARY KEY (tx_hash, event_index)
    );

    CREATE TABLE IF NOT EXISTS events_EventJailedFinalityProvider (
      block_height BIGINT NOT NULL,
      block_timestamp TIMESTAMP NOT NULL,
      tx_hash TEXT NOT NULL,
      tx_index SMALLINT,
      event_index SMALLINT,
      btc_pk TEXT NOT NULL,
      PRIMARY KEY (tx_hash, event_index)
    );

    CREATE TABLE IF NOT EXISTS events_EventUnjailedFinalityProvider (
      block_height BIGINT NOT NULL,
      block_timestamp TIMESTAMP NOT NULL,
      tx_hash TEXT NOT NULL,
      tx_index SMALLINT,
      event_index SMALLINT,
      btc_pk TEXT NOT NULL,
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
// CREATE FUNCTIONS
//////////////////////////

const (
	sqlCreateFuncVotingPowerDistAtBlock = `
		CREATE OR REPLACE FUNCTION func_voting_power_dist_at_block(_block_height BIGINT)
    RETURNS TABLE (
      fp_btc_pk TEXT,
      voting_power BIGINT
    ) AS $$
    BEGIN
      RETURN QUERY
      WITH
      -- Fetch chain params
      ChainParams AS (
        SELECT k_value, w_value, cov_quorum FROM table_chain_params
      ),
      -- Compile list of FPs belonging to consumer chain
      ConsumerChainFPs AS (
        SELECT btc_pk AS fp_btc_pk FROM table_initial_finality_providers
        UNION ALL
        SELECT btc_pk AS fp_btc_pk FROM events_EventNewFinalityProvider 
        WHERE block_height <= _block_height
      ),
      -- Compile list of slashed FPs
      SlashedFPs AS (
        SELECT btc_pk AS fp_btc_pk FROM events_EventSlashedFinalityProvider
        WHERE block_height <= _block_height
      ),
      -- Compile list of jailed FPs
      JailedFPs AS (
        SELECT j.btc_pk AS fp_btc_pk
        FROM events_EventJailedFinalityProvider j
        WHERE j.block_height <= _block_height
        AND NOT EXISTS (
          SELECT 1
          FROM events_EventUnjailedFinalityProvider u
          WHERE u.btc_pk = j.btc_pk
            AND u.block_height <= _block_height
        )
      ),
      -- Compile list of active FPs
      ActiveFPs AS (
        SELECT fp_btc_pk FROM ConsumerChainFPs
        EXCEPT
        SELECT fp_btc_pk FROM SlashedFPs
        EXCEPT
        SELECT fp_btc_pk FROM JailedFPs
      ),
      -- Compile list of delegation tx hexes
      InitialDelegations AS (
        SELECT 
          id.staking_tx_hex, 
          id.start_height,
          id.end_height,
          id.total_sat, 
          id.num_covenant_sigs,
          id.fp_btc_pk_list,
          'ACTIVE'::BTCDelegationStatus AS state
        FROM table_initial_delegations id
        WHERE id.active
      ),
      -- Compile new delegations
      DelegationUpdates AS (
        SELECT DISTINCT ON (staking_tx_hash) 
          su.staking_tx_hash, 
          bd.start_height,
          bd.end_height,
          bd.total_sat,
          bd.num_covenant_sigs,
          bd.fp_btc_pk_list,
          su.new_state AS state
        FROM events_EventBTCDelegationStateUpdate su
        JOIN table_btc_delegations bd ON su.staking_tx_hash = bd.staking_tx_hash
        WHERE su.block_height <= _block_height
        ORDER BY su.staking_tx_hash, su.block_height DESC
      ),
      -- Compile list of current delegations as of block n
      Delegations AS (
        SELECT
          COALESCE(du.staking_tx_hex, id.staking_tx_hex) AS staking_tx_hex,
          COALESCE(du.start_height, id.start_height) AS start_height,
          COALESCE(du.end_height, id.end_height) AS end_height,
          COALESCE(du.total_sat, id.total_sat) AS total_sat,
          COALESCE(du.num_covenant_sigs, id.num_covenant_sigs) AS num_covenant_sigs,
          COALESCE(du.fp_btc_pk_list, id.fp_btc_pk_list) AS fp_btc_pk_list,
          COALESCE(du.state, id.state) AS state
        FROM InitialDelegations id
        LEFT JOIN DelegationUpdates du ON id.staking_tx_hash = du.staking_tx_hash
        WHERE COALESCE(du.state, id.state) = 'ACTIVE'::BTCDelegationStatus
      ),
      -- Filter for active FPs and delegations
      ActiveDelegations AS (
        SELECT d.*
        FROM Delegations d
        JOIN ActiveFPs af ON d.fp_btc_pk_list @> ARRAY[af.fp_btc_pk]
        JOIN ChainParams cp ON true
        WHERE _block_height >= d.start_height + cp.k_value
          AND _block_height + cp.w_value <= d.end_height
          AND d.num_covenant_sigs >= cp.cov_quorum
      ),
      -- Return list of FPs and their voting power
      VotingPowerDist AS (
        SELECT 
          af.fp_btc_pk,
          SUM(d.total_sat) AS voting_power
        FROM ActiveDelegations d
        JOIN ActiveFPs af ON d.fp_btc_pk_list @> ARRAY[af.fp_btc_pk]
        GROUP BY af.fp_btc_pk
      )
      SELECT * FROM VotingPowerDist;
    END;
    $$ LANGUAGE plpgsql;
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
	sqlInsertChainParams = `
		INSERT INTO table_chain_params (k_value, w_value, cov_quorum) VALUES ($1, $2, $3)
	`
	// TODO: consider if we can remove any fields from below if not needed
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
      slashed_babylon_height,
      slashed_btc_height,
      consumer_id
    ) VALUES (
      $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
    )
  `
	// TODO: consider if we can remove any fields from below if not needed
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
      slashed_babylon_height,
      slashed_btc_height,
      consumer_id
    ) VALUES (
      $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16
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
	sqlInsertEventJailedFinalityProvider = `
		INSERT INTO events_EventJailedFinalityProvider (
      block_height,
      block_timestamp,
      tx_hash,
      tx_index,
      event_index,
      btc_pk
    ) VALUES ($1, $2, $3, $4, $5, $6)
	`
	sqlInsertEventUnjailedFinalityProvider = `
		INSERT INTO events_EventUnjailedFinalityProvider (
      block_height,
      block_timestamp,
      tx_hash,
      tx_index,
      event_index,
      btc_pk
    ) VALUES ($1, $2, $3, $4, $5, $6)
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
	// TODO: consider if we need to add more fields to below from `BTCDelegationResponse`
	sqlInsertBTCDelegationInfo = `
    INSERT INTO table_btc_delegations (
      staker_addr,
      btc_pk,
      fp_btc_pk_list,
      start_height,
      end_height,
      total_sat,
      staking_tx_hex,
      slashing_tx_hex,
      delegator_slash_sig_hex,
      num_covenant_sigs,
      staking_output_idx,
      active,
      status_desc,
      unbonding_time,
      params_version
    ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
  `
)

//////////////////////////
// QUERIES
//////////////////////////

const (
	sqlQueryFinalityProvidersAtHeight = `
		SELECT 
      description_moniker, 
      description_identity, 
      description_website, 
      description_security_contact, 
      description_details, 
      commission, 
      addr AS babylon_pk, 
      btc_pk,
      slashed_babylon_height,
      slashed_btc_height,
      consumer_id
    FROM table_initial_finality_providers

    UNION ALL
    
    SELECT 
      description_moniker, 
      description_identity, 
      description_website, 
      description_security_contact, 
      description_details, 
      commission, 
      babylon_pk_key AS babylon_pk,
      btc_pk,
      slashed_babylon_height,
      slashed_btc_height,
      consumer_id
    FROM events_EventNewFinalityProvider
    WHERE block_height <= $1
	`
	sqlQueryInitialFinalityProviders = `
    SELECT 
      description_moniker, 
      description_identity, 
      description_website, 
      description_security_contact, 
      description_details, 
      commission, 
      addr, 
      btc_pk,
      slashed_babylon_height,
      slashed_btc_height,
      consumer_id
    FROM table_initial_finality_providers
  `
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
	sqlQueryBTCDelegationInfo = `
		SELECT 
      staker_addr, 
      btc_pk, 
      fp_btc_pk_list, 
      start_height, 
      end_height, 
      total_sat, 
      staking_tx_hex, 
      slashing_tx_hex, 
      delegator_slash_sig_hex, 
      num_covenant_sigs, 
      staking_output_idx, 
      active, 
      status_desc, 
      unbonding_time, 
      params_version 
    FROM table_btc_delegations WHERE staking_tx_hash = $1
	`
	sqlQueryVotingPowerDistAtBlock = `
		SELECT btc_pk, voting_power FROM func_voting_power_dist_at_block($1)
	`
)
