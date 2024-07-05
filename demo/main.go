package main

import (
	"errors"
	"fmt"

	"github.com/babylonchain/babylon-da-sdk/btcclient"
	"github.com/babylonchain/babylon-da-sdk/sdk"
)

func checkBlockFinalized(height uint64, hash string) {
	client, err := sdk.NewClient(&sdk.Config{
		ChainType: 0,
		// TODO: avoid using stub contract
		ContractAddr: "bbn1ghd753shjuwexxywmgs4xz7x2q732vcnkm6h2pyv9s6ah3hylvrqxxvh0f",
		BTCConfig: btcclient.DefaultBTCConfig(),
	})

	if err != nil {
		fmt.Printf("error creating client: %v\n", err)
		return
	}

	isFinalized, err := client.QueryIsBlockBabylonFinalized(&sdk.L2Block{
		BlockHeight:    height,
		BlockHash:      hash,
		BlockTimestamp: uint64(1718332131),
	})

	fmt.Println("isFinalized", isFinalized)
	fmt.Println("err", err)

	if err == nil {
		fmt.Printf("is block %d finalized?: %t\n", height, isFinalized)
		return
	}

	if !errors.Is(err, sdk.ErrNoFpHasVotingPower) {
		fmt.Printf("error checking block %d: %v\n", height, err)
	} else {
		fmt.Printf("no FP has voting power at block %d for the consumer chain\n", height)
	}
}

func main() {
	// TODO: this will always return false. we should find a better way to demo it
	checkBlockFinalized(uint64(2), "0x1000000000000000000000000000000000000000000000000000000000000000")
}
