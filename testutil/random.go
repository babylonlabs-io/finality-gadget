package testutil

import (
	"math/rand"
	"strings"

	"github.com/babylonchain/babylon-finality-gadget/sdk/cwclient"
	"github.com/ethereum/go-ethereum/common"
)

func RandomHash(rng *rand.Rand) (out common.Hash) {
	rng.Read(out[:])
	return
}

func RandomL2Block(rng *rand.Rand) (out cwclient.L2Block, outWithHashTrimmed cwclient.L2Block) {
	out.BlockHash = RandomHash(rng).String()
	out.BlockHeight = rng.Uint64()
	out.BlockTimestamp = rng.Uint64()
	outWithHashTrimmed = out
	outWithHashTrimmed.BlockHash = strings.TrimPrefix(out.BlockHash, "0x")
	return
}

func GenL2Block(rng *rand.Rand, initBlock *cwclient.L2Block, l2BlockTime uint64, offset uint64) (out cwclient.L2Block, outWithHashTrimmed cwclient.L2Block) {
	out.BlockHash = RandomHash(rng).String()
	out.BlockHeight = initBlock.BlockHeight + offset
	out.BlockTimestamp = initBlock.BlockTimestamp + offset*l2BlockTime
	outWithHashTrimmed = out
	outWithHashTrimmed.BlockHash = strings.TrimPrefix(out.BlockHash, "0x")
	return
}
