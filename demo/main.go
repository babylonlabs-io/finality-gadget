package main

import (
	"fmt"

	"github.com/babylonchain/babylon-da-sdk/sdk"
)

func checkBlockFinalized(height uint64, hash string) {
	client, err := sdk.NewClient(sdk.Config{
		ChainType:    0,
		ContractAddr: "sei18fs8atjcxrsypskpk725q2vr8j76q3xwcfle3w2qlna48acmed0sp30xm8",
	})

	if err != nil {
		fmt.Printf("error creating client: %v\n", err)
		return
	}

	isFinalized, err := client.QueryIsBlockBabylonFinalized(sdk.QueryParams{
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
	blockHash := "0x3aa074144a25d3ed71c7353a20c579650e0c56a993444c6156d44bb90b932f0d"
	blockHashForked := "forked hash"

	fmt.Println("=== When the block hash has enough votes: ===")
	for i := range 4 {
		checkBlockFinalized(uint64(i), blockHash)
	}

	fmt.Println("\n=== When the block hash doesn't have enough votes: ===")
	for i := range 4 {
		checkBlockFinalized(uint64(i), blockHashForked)
	}
}
