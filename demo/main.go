package main

import (
	"fmt"

	"github.com/babylonchain/babylon-da-sdk/sdk"
)

func checkBlockFinalized(height uint64, hash string) {
	isFinalized, err := sdk.QueryIsBlockBabylonFinalized(sdk.QueryParams{
		ChainType:      0,
		ContractAddr:   "osmo1zck32had0fpc4fu34ae58zvs3mjd5yrzs70thw027nfqst7edc3sdqak0m",
		BlockHeight:    height,
		BlockHash:      hash,
		BlockTimestamp: uint64(1718332131),
	})
	if err != nil {
		fmt.Printf("error checking block %d: %v\n", height, err)
	} else {
		fmt.Printf("is block %d finalized?: %t\n", height, isFinalized)
	}
}

func main() {
	blockHashWithoutEnoughVotes := "0x3aa074144a25d3ed71c7353a20c579650e0c56a993444c6156d44bb90b932f0d"
	blockHashWithEnoughVotes := "stub hash"

	fmt.Println("=== When the block hash has enough votes: ===")
	for i := range 4 {
		checkBlockFinalized(uint64(i), blockHashWithEnoughVotes)
	}

	fmt.Println("\n=== When the block hash doesn't have enough votes: ===")
	for i := range 4 {
		checkBlockFinalized(uint64(i), blockHashWithoutEnoughVotes)
	}
}
