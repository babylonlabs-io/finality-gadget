package main

import (
	"fmt"

	"github.com/babylonchain/babylon-da-sdk/sdk"
)

func checkBlockFinalized(height uint64) {
	isFinalized, _ := sdk.QueryIsBlockBabylonFinalized(sdk.QueryParams{
		ChainType:      0,
		ContractAddr:   "osmo1zck32had0fpc4fu34ae58zvs3mjd5yrzs70thw027nfqst7edc3sdqak0m",
		BlockHeight:    height,
		BlockHash:      "fake hash",
		BlockTimestamp: "1718332131",
	})
	fmt.Println("is finalized?: ", isFinalized)

}

func main() {
	for i := range 4 {
		checkBlockFinalized(uint64(i))
	}
}
