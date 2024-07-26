package db

type Block struct {
	BlockHash 	      string  	`json:"block_hash" description:"block hash"`
	BlockHeight   		uint64    `json:"block_height" description:"block height"`
	BlockTimestamp  	uint64 	  `json:"block_timestamp" description:"block timestamp"`
	IsFinalized 			bool    	`json:"is_finalized" description:"babylon block finality status"`
}