package db

type Block struct {
	BlockHeight   		uint64    `json:"block_height" description:"block height"`
	BlockHash 	      string  	`json:"block_hash" description:"block hash"`
	BlockTimestamp  	uint64 	  `json:"block_timestamp" description:"block timestamp"`
	IsFinalized 			bool    	`json:"is_finalized" description:"babylon block finality status"`
}