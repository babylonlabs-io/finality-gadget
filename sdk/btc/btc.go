package btc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	// TODO: for now using public RPC is fine but it can get rate limited. So in the future, we should
	// use dedicated RPC or even consider a config param
	rpcURL = "https://rpc.ankr.com/btc"
	// No need to search from height 0. Any bitcoin height before the first FP registration works
	// 848664 is a block at 2024-06-19
	// https://mempool.space/block/000000000000000000019200f971921bb73e5f16b5098ad71a91849a63964a70
	btcHeightSearchStart uint64 = 848664
)

type rpcRequest struct {
	Method string        `json:"method"`
	Params []interface{} `json:"params"`
	ID     string        `json:"id"`
}

type rpcResponse struct {
	Result json.RawMessage `json:"result"`
	Error  interface{}     `json:"error"`
	ID     string          `json:"id"`
}

func callRPC(method string, params []interface{}) (json.RawMessage, error) {
	requestBody, err := json.Marshal(rpcRequest{Method: method, Params: params, ID: "1"})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", rpcURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response rpcResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	if response.Error != nil {
		return nil, fmt.Errorf("RPC error: %v", response.Error)
	}

	return response.Result, nil
}

// given a timestamp, search the largest block height whose timestamp is less or equal to it
func GetBlockHeightByTimestamp(targetTimestamp uint64) (uint64, error) {
	var currentHeight uint64
	// returns the height of the most-work fully-validated chain.
	result, err := callRPC("getblockcount", []interface{}{})
	if err != nil {
		return 0, err
	}
	if err := json.Unmarshal(result, &currentHeight); err != nil {
		return 0, err
	}

	lowerBound := btcHeightSearchStart
	upperBound := currentHeight

	for lowerBound <= upperBound {
		midHeight := (lowerBound + upperBound) / 2

		// get block hash by height
		result, err := callRPC("getblockhash", []interface{}{midHeight})
		if err != nil {
			return 0, err
		}
		var blockHash string
		if err := json.Unmarshal(result, &blockHash); err != nil {
			return 0, err
		}

		// get block header by hash. the header contains info such as the block time expressed in UNIX epoch time
		result, err = callRPC("getblockheader", []interface{}{blockHash})
		if err != nil {
			return 0, err
		}
		var blockHeader map[string]interface{}
		if err := json.Unmarshal(result, &blockHeader); err != nil {
			return 0, err
		}

		blockTimestamp := uint64(blockHeader["time"].(float64))

		if blockTimestamp < targetTimestamp {
			lowerBound = midHeight + 1
		} else if blockTimestamp > targetTimestamp {
			upperBound = midHeight - 1
		} else {
			return midHeight, nil
		}
	}

	// timestamp is in the future (not in the most-work fully-validated chain)
	// so we cannot determine the height from the timestamp
	if lowerBound > currentHeight {
		return 0, nil
	}

	// timestamp is too old (before we deploy this system)
	// so we cannot determine the height from the timestamp
	if upperBound < btcHeightSearchStart {
		return 0, nil
	}

	return lowerBound - 1, nil
}
