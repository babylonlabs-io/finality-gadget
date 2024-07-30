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
	Db 						*db.BBoltHandler

	Mutex 				sync.Mutex

	PollInterval 	time.Duration
	startHeight 	uint64
	currHeight 	uint64
}

type BlockInfo struct {
	Hash        string
	Height      uint64
	Timestamp		uint64
}