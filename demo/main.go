package main

import (
	"fmt"

	"github.com/babylonchain/babylon-da-sdk/sdk"
)

func checkBlockFinalized(height uint64) {
	isFinalized, _ := sdk.QueryIsBlockBabylonFinalized(sdk.QueryParams{
		ChainType:    0,
		ContractAddr: "osmo1eqetf8qcgjh2xae0v4at0u2zt8n3q008ptasre0p577wxhj9enkqlgmrlv",
		BlockHeight:  height,
	})
	fmt.Println("is finalized?: ", isFinalized)

}

func main() {
	for i := range 10 {
		checkBlockFinalized(uint64(i))
	}
}
