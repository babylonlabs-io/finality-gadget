package client

import (
	"fmt"
	"strings"

	"github.com/babylonchain/babylon-finality-gadget/sdk/cwclient"
)

/* QueryIsBlockBabylonFinalized checks if the given L2 block is finalized by the Babylon finality gadget
 *
 * - if the finality gadget is not enabled, always return true
 * - else, check if the given L2 block is finalized
 * - return true if finalized, false if not finalized, and error if any
 *
 * - to check if the block is finalized, we need to:
 *   - get the consumer chain id
 *   - get all the FPs pubkey for the consumer chain
 *   - convert the L2 block timestamp to BTC height
 *   - get all FPs voting power at this BTC height
 *   - calculate total voting power
 *   - get all FPs that voted this L2 block with the same height and hash
 *   - calculate voted voting power
 *   - check if the voted voting power is more than 2/3 of the total voting power
 */
func (sdkClient *SdkClient) QueryIsBlockBabylonFinalized(
	queryParams cwclient.L2Block,
) (bool, error) {
	// check if the finality gadget is enabled
	// if not, always return true to pass through op derivation pipeline
	isEnabled, err := sdkClient.cwClient.QueryIsEnabled()
	if err != nil {
		return false, err
	}
	if !isEnabled {
		return true, nil
	}

	// trim prefix 0x for the L2 block hash
	queryParams.BlockHash = strings.TrimPrefix(queryParams.BlockHash, "0x")

	// get the consumer chain id
	consumerId, err := sdkClient.cwClient.QueryConsumerId()
	if err != nil {
		return false, err
	}

	// get all the FPs pubkey for the consumer chain
	allFpPks, err := sdkClient.bbnClient.QueryAllFpBtcPubKeys(consumerId)
	if err != nil {
		return false, err
	}

	// convert the L2 timestamp to BTC height
	btcblockHeight, err := sdkClient.btcClient.GetBlockHeightByTimestamp(queryParams.BlockTimestamp)
	if err != nil {
		return false, err
	}

	// get all FPs voting power at this BTC height
	allFpPower, err := sdkClient.bbnClient.QueryMultiFpPower(allFpPks, btcblockHeight)
	if err != nil {
		return false, err
	}

	// calculate total voting power
	var totalPower uint64 = 0
	for _, power := range allFpPower {
		totalPower += power
	}

	// no FP has voting power for the consumer chain
	if totalPower == 0 {
		return false, ErrNoFpHasVotingPower
	}

	// get all FPs that voted this (L2 block height, L2 block hash) combination
	votedFpPks, err := sdkClient.cwClient.QueryListOfVotedFinalityProviders(&queryParams)
	if err != nil {
		return false, err
	}
	if votedFpPks == nil {
		return false, nil
	}
	// calculate voted voting power
	var votedPower uint64 = 0
	for _, key := range votedFpPks {
		if power, exists := allFpPower[key]; exists {
			votedPower += power
		}
	}

	// quorom < 2/3
	if votedPower*3 < totalPower*2 {
		return false, nil
	}
	return true, nil
}

/* QueryBlockRangeBabylonFinalized searches for a row of consecutive finalized blocks in the block range, and returns
 * the last finalized block height
 *
 * Example: if give block range 1-10, and block 1-5 are finalized, and block 6-10 are not finalized, then return 5
 *
 * - if no block in the range is finalized, return (nil, nil)
 * - else, return the height of the last found consecutive finalized block, return error if any
 *
 * Example: if give block range 1-10, and block 1-5 are finalized, and when querying block 6 we meet an error, then
 * return (5, error)
 *
 * Note: caller needs to make sure the given queryBlocks are consecutive (we don't check hashes inside this method)
 * and start from low to high
 */
func (sdkClient *SdkClient) QueryBlockRangeBabylonFinalized(
	queryBlocks []*cwclient.L2Block,
) (*uint64, error) {
	if len(queryBlocks) == 0 {
		return nil, fmt.Errorf("no blocks provided")
	}
	// check if the blocks are consecutive
	for i := 1; i < len(queryBlocks); i++ {
		if queryBlocks[i].BlockHeight != queryBlocks[i-1].BlockHeight+1 {
			return nil, fmt.Errorf("blocks are not consecutive")
		}
	}
	var finalizedBlockHeight *uint64
	for _, block := range queryBlocks {
		isFinalized, err := sdkClient.QueryIsBlockBabylonFinalized(*block)
		if err != nil {
			return finalizedBlockHeight, err
		}
		if isFinalized {
			finalizedBlockHeight = &block.BlockHeight
		} else {
			break
		}
	}
	return finalizedBlockHeight, nil
}
