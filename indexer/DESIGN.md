# Indexer Design

## DB Schema

SQL tables

1. `table_initial_finality_providers` : table of initial finality providers for a consumer chain
2. `table_initial_delegations` : table of initial delegations to FPs of a consumer chain
3. `table_btc_delegations` : table of identifying information on BTC delegations, such as the FP, staked amount etc, queried from Babylon RPC
4. `table_EventNewFinalityProvider` : table of `EventNewFinalityProvider` events with emitted identifying info
5. `table_EventBTCDelegationStateUpdate` : table of `EventBTCDelegationStateUpdate` events with status changes
6. `table_EventSlashedFinalityProvider` : table of `EventSlashedFinalityProvider` events with status changes

Views

1. `view_finality_providers` : view of all indexed FPs including their status, indexed from `events_EventNewFinalityProvider` and `events_EventPowerDistUpdate`
2. `view_delegations` : view of all delegations including their status, indexed from `events_EventBTCDelegationStateUpdate`

Queries

1. `func_finality_providers_at_block`
2. `func_delegations_at_block`
3. `func_voting_power_dist_at_block`

_TODO: consider if we need seperate queries by btc block vs babylon block vs timestamp_

## Call logic

1. Define env vars / config.toml:
   1. Babylon RPC endpoint
   2. Consumer chain id - `consumer_id`
2. Startup - Populate start info
   1. Fetch and store chain configs
      1. BTC confirmation depth
      2. Checkpoint finalization timeout
      3. Covenant quorum
   2. Fetch all finality providers for a consumer chain with `QueryConsumerFinalityProviders` (passing `consumer_id`), and populate `table_initial_finality_providers`
   3. Loop through each fp and fetch all delegations, storing to `table_initial_delegations`
3. Chain poller: Loop through each block sequentially and parse relevant events, populating tables
   1. Case `EventNewFinalityProvider` : if `consumer_id` matches, store to `table_EventNewFinalityProvider`
   2. Case `EventBTCDelegationStateUpdate`
      1. Store delegation to `table_EventBTCDelegationStateUpdate`
      2. Check `table_btc_delegations` for btc delegation info
      3. If it does not exist, query RPC for it and store it (we want to store all delegations, not just those in our consumer chain, to avoid repeated querying on future delegation state update events)
   3. Case `EventSlashedFinalityProvider`
      1. Store to `table_EventSlashedFinalityProvider`
4. Queries
   1. Get all finality providers at block
      1. Query `table_initial_finality_providers` and `table_EventNewFinalityProvider` for all `block_height <= $block`
      2. Remove any delegations that have been slashed by `EventSlashedFinalityProvider` at block `block_height <= $block` (only consider FPs in our consumer chain)
      3. Check if any FPs are jailed at block `$block`, and if so, remove them from the list of finality providers
      4. Cleanup: Remove any FPs with 0 voting power
   2. Get all delegations at block
      1. Start with `table_initial_delegations` and `table_btc_delegations`
      2. Query `table_EventBTCDelegationStateUpdate` for all `block_height <= $block`, filtering for delegations to FPs in our consumer chain, and apply deltas
      3. Check whether each delegation is still active at given block height based on delegation info and chain configs
      4. Exclude all naturally unbound delegations based on the timelock date
   3. Get voting power distribution at block based on the above

Existing logic

1. Check if CW contract is enabled, and pass through if not
2. Check whether btc staking is activated at height
3. [Indexer] Query all FPs for the consumer chain
4. [Indexer] Get voting powers (uint64) of all FPs (string pk hex)
   1. Query all delegations of each FP
   2. Check if each delegation is active at given height
   3. Sum up active delegations to get voting power
5. Calculate total power
6. Calculate total power of voted FPs

Test strategy

1. Create mock stream of events and expect results
