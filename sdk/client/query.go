package client

import (
	"fmt"
	"math"
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

	// get all FPs pubkey for the consumer chain
	allFpPks, err := sdkClient.queryAllFpBtcPubKeys()
	if err != nil {
		return false, err
	}

	// convert the L2 timestamp to BTC height
	btcblockHeight, err := sdkClient.btcClient.GetBlockHeightByTimestamp(queryParams.BlockTimestamp)
	if err != nil {
		return false, err
	}

	// check whether the btc staking is actived
	earliestDelHeight, err := sdkClient.bbnClient.QueryEarliestActiveDelBtcHeight(allFpPks)
	if err != nil {
		return false, err
	}
	if btcblockHeight < earliestDelHeight {
		return false, ErrBtcStakingNotActivated
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

/* QueryBtcStakingActivatedTimestamp returns the timestamp when the BTC staking is activated
 *
 * - We will check for k deep and covenant quorum to mark a delegation as active
 * - So technically, activated time needs to be max of the following
 *	 - timestamp of Babylon block that BTC delegation receives covenant quorum
 *	 - timestamp of BTC block that BTC delegation's staking tx becomes k-deep
 * - But we don't have a Babylon API to find the earliest Babylon block where BTC delegation gets covenant quorum.
 *   and it's probably not a good idea to add this to Babylon, as this creates more coupling between Babylon and
 *   consumer. So waiting for covenant quorum can be implemented in a clean way only with Babylon side support.
 *   this will be considered as future work
 * - For now, we will use the k-deep BTC block timestamp for the activation timestamp
 * - The time diff issue is not burning. The time diff between pending and active only matters if FP equivocates
 *   during that time period
 *
 * returns math.MaxUint64, ErrBtcStakingNotActivated if the BTC staking is not activated
 */
func (sdkClient *SdkClient) QueryBtcStakingActivatedTimestamp() (uint64, error) {
	allFpPks, err := sdkClient.queryAllFpBtcPubKeys()
	if err != nil {
		return math.MaxUint64, err
	}

	// check whether the btc staking is actived
	earliestDelHeight, err := sdkClient.bbnClient.QueryEarliestActiveDelBtcHeight(allFpPks)
	if err != nil {
		return math.MaxUint64, err
	}

	// not activated yet
	if earliestDelHeight == math.MaxUint64 {
		return math.MaxUint64, ErrBtcStakingNotActivated
	}

	// get the timestamp of the BTC height
	btcBlockTimestamp, err := sdkClient.btcClient.GetBlockTimestampByHeight(earliestDelHeight)
	if err != nil {
		return math.MaxUint64, err
	}
	return btcBlockTimestamp, nil
}

func (sdkClient *SdkClient) queryAllFpBtcPubKeys() ([]string, error) {
	// get the consumer chain id
	consumerId, err := sdkClient.cwClient.QueryConsumerId()
	if err != nil {
		return nil, err
	}

	// get all the FPs pubkey for the consumer chain
	allFpPks, err := sdkClient.bbnClient.QueryAllFpBtcPubKeys(consumerId)
	if err != nil {
		return nil, err
	}
	return allFpPks, nil
}
