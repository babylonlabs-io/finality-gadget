package main

import (
	"fmt"

	"github.com/babylonchain/babylon-da-sdk/sdk"
)

func checkBlockFinalized(height uint64, hash string) {
	client, err := sdk.NewClient(sdk.Config{
		ChainType:    0,
		ContractAddr: "bbn1ghd753shjuwexxywmgs4xz7x2q732vcnkm6h2pyv9s6ah3hylvrqxxvh0f",
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
	// TODO: now you will see `error checking block 2: not enough voting power`
	// in the future, we will deploy a better stub contract
	checkBlockFinalized(uint64(2), blockHash)
}
