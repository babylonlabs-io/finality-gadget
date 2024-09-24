# Indexer Design

## DB Schema

SQL tables

1. `table_initial_finality_providers` : table of initial finality providers for a consumer chain
2. `table_initial_delegations` : table of initial delegations to FPs of a consumer chain
3. `table_btc_delegations` : table of identifying information on BTC delegations, such as the FP, staked amount etc, queried from Babylon RPC
4. `table_EventNewFinalityProvider` : table of `EventNewFinalityProvider` events with emitted identifying info
5. `table_EventBTCDelegationStateUpdate` : table of `EventBTCDelegationStateUpdate` events with status changes
6. `table_EventSelectiveSlashing` : table of `EventSelectiveSlashing` events with status changes
7. `table_EventSlashedFinalityProvider` : table of `EventSlashedFinalityProvider` events with status changes

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
   1. Fetch all finality providers for a consumer chain with `QueryConsumerFinalityProviders` (passing `consumer_id`), and populate `table_initial_finality_providers`
   2. Loop through each fp and fetch all delegations, storing to `table_initial_delegations`
3. Chain poller: Loop through each block sequentially and parse relevant events, populating tables
   1. Case `EventNewFinalityProvider` : if `consumer_id` matches, store to `table_EventNewFinalityProvider`
   2. Case `EventBTCDelegationStateUpdate`
      1. Check `table_btc_delegations` for btc delegation info
      2. If it does not exist, query RPC for it and store it (we want to store all delegations, not just those in our consumer chain, to avoid repeated querying on future delegation state update events)
      3. If delegation is made to FP in our consumer chain, then store to `table_EventBTCDelegationStateUpdate`
   3. Case `EventSelectiveSlashing`
      1. Check if FP exists in our table
      2. If so, store to `table_EventSelectiveSlashing`
   4. Case `EventSlashedFinalityProvider`
      1. Check if FP exists in our table
      2. If so, store to `table_EventSlashedFinalityProvider`
4. Queries
   1. Get all finality providers at block
      1. Query `table_initial_finality_providers` and `table_EventNewFinalityProvider` for all `block_height <= $block`
      2. Deduct voting power of FPs based on `events_EventPowerDistUpdate` at block `block_height <= $block` and remove any FPs with 0 voting power
   2. Get all delegations at block
      1. Start with `table_initial_delegations` and `table_btc_delegations`
      2. Query `table_EventBTCDelegationStateUpdate` for all `block_height <= $block` and apply deltas
      3. Exclude all naturally unbound delegations based on the timelock date
   3. Check if any FPs are jailed, and if so, remove them from the list of finality providers
   4. Get voting power distribution at block based on the above
   5. Calculate whether block is finalized based on the above
