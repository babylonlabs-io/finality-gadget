package testutil

import (
	"math/rand"
	"strings"

	"github.com/babylonlabs-io/finality-gadget/types"
	"github.com/ethereum/go-ethereum/common"
)

func RandomHash(rng *rand.Rand) (out common.Hash) {
	rng.Read(out[:])
	return
}

func RandomL2Block(rng *rand.Rand) (out types.Block, outWithHashTrimmed types.Block) {
	out.BlockHash = RandomHash(rng).String()
	out.BlockHeight = rng.Uint64()
	out.BlockTimestamp = rng.Uint64()
	outWithHashTrimmed = out
	outWithHashTrimmed.BlockHash = strings.TrimPrefix(out.BlockHash, "0x")
	return
}

func GenL2Block(rng *rand.Rand, initBlock *types.Block, l2BlockTime uint64, offset uint64) (out types.Block, outWithHashTrimmed types.Block) {
	out.BlockHash = RandomHash(rng).String()
	out.BlockHeight = initBlock.BlockHeight + offset
	out.BlockTimestamp = initBlock.BlockTimestamp + offset*l2BlockTime
	outWithHashTrimmed = out
	outWithHashTrimmed.BlockHash = strings.TrimPrefix(out.BlockHash, "0x")
	return
}
