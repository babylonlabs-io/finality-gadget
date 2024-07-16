package client

import "github.com/babylonchain/babylon-finality-gadget/sdk/cwclient"

type ISdkClient interface {
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
	QueryIsBlockBabylonFinalized(queryParams *cwclient.L2Block) (bool, error)

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
	QueryBlockRangeBabylonFinalized(queryBlocks []*cwclient.L2Block) (*uint64, error)
}
