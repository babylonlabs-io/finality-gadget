package verifier

import (
	"sync"
	"time"

	"github.com/babylonchain/babylon-finality-gadget/sdk/client"
	"github.com/babylonchain/babylon-finality-gadget/verifier/db"
	"github.com/ethereum/go-ethereum/ethclient"
)

type Verifier struct {
	SdkClient 		*client.SdkClient
	L2Client 			*ethclient.Client
	Pg 						*db.PostgresHandler

	Mutex 				sync.Mutex

	PollInterval 	time.Duration
	blockHeight 	uint64
}

type Config struct {
	L2RPCHost        		string 					`long:"l2-rpc-host" description:"rpc host address of the L2 node"`
	BitcoinRPCHost      string 					`long:"bitcoin-rpc-host" description:"rpc host address of the bitcoin node"`
	PGConnectionString 	string 					`long:"pg-connection-string" description:"Postgres DB connection string"`
	FGContractAddress 	string 					`long:"fg-contract-address" description:"BabylonChain op finality gadget contract address"`
	BBNChainID 					string 					`long:"bbn-chain-id" description:"BabylonChain chain ID"`
	BBNRPCAddress 			string 					`long:"bbn-rpc-address" description:"BabylonChain chain RPC address"`
	PollInterval				time.Duration		`long:"retry-interval" description:"interval in seconds to recheck Babylon finality of block"`
}

type BlockInfo struct {
	Height      uint64
	Hash        string
	Timestamp		uint64
}