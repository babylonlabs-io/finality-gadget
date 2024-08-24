package finalitygadget

import "github.com/babylonlabs-io/finality-gadget/types"

type IFinalityGadget interface {
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
	QueryIsBlockBabylonFinalized(block *types.Block) (bool, error)

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
	QueryBlockRangeBabylonFinalized(queryBlocks []*types.Block) (*uint64, error)

	/* QueryBtcStakingActivatedTimestamp returns the timestamp when the BTC staking is activated
	 *
	 * - We will check for k deep and covenant quorum to mark a delegation as active
	 * - So technically, activated time needs to be max of the following
	 *	 - timestamp of Babylon block that BTC delegation receives covenant quorum
	 *	 - timestamp of BTC block that BTC delegation's staking tx becomes k-deep
	 * - But we don't have a Babylon API to find the earliest Babylon block where BTC delegation gets covenant quorum.
	 *   and it's probably not a good idea to add this to Babylon, as this creates more coupling between Babylon and
	 *   consumer. So waiting for covenant quorum can be implemented in a clean way only with Babylon side support.
	 *   This will be considered as future work
	 * - For now, we will use the k-deep BTC block timestamp for the activation timestamp
	 * - The time diff issue is not burning. The time diff between pending and active only matters if FP equivocates
	 *   during that time period
	 *
	 * returns math.MaxUint64, ErrBtcStakingNotActivated if the BTC staking is not activated
	 */
	QueryBtcStakingActivatedTimestamp() (uint64, error)

	// InsertBlock inserts a btc finalized block into the local db
	InsertBlock(block *types.Block) error

	// GetBlockByHeight returns the btc finalized block at given height by querying the local db
	GetBlockByHeight(height uint64) (*types.Block, error)

	// GetBlockByHash returns the btc finalized block at given hash by querying the local db
	GetBlockByHash(hash string) (*types.Block, error)

	// QueryIsBlockFinalizedByHeight returns the btc finalization status of a block at given height by querying the local db
	QueryIsBlockFinalizedByHeight(height uint64) (bool, error)

	// QueryIsBlockFinalizedByHash returns the btc finalization status of a block at given hash by querying the local db
	QueryIsBlockFinalizedByHash(hash string) (bool, error)

	// QueryLatestFinalizedBlock returns the latest finalized block by querying the local db
	QueryLatestFinalizedBlock() (*types.Block, error)

	// QueryTransactionStatus returns the finality status of a transaction
	QueryTransactionStatus(txHash string) (*types.TransactionInfo, error)
}
