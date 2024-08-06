package finalitygadget

import "github.com/ethereum/go-ethereum/common"

func normalizeBlockHash(hash string) string {
	return common.HexToHash(hash).Hex()
}
