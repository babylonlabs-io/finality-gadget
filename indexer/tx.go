package indexer

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/babylonlabs-io/finality-gadget/types"
	ctypes "github.com/cometbft/cometbft/rpc/core/types"
)

type Tx struct {
	// BlockTime returns the time of the block that contains the transaction.
	BlockTime time.Time

	// Raw contains the transaction as returned by the Tendermint API.
	Raw *ctypes.ResultTx
}

// EventAttribute defines a transaction event attribute.
type EventAttribute struct {
	Key   string `json:"key"`
	Value []byte `json:"value"`
}

func (t Tx) GetTxInfo() *types.TxInfo {
	txInfo := &types.TxInfo{
		BlockHeight:    t.Raw.Height,
		BlockTimestamp: t.BlockTime,
		TxHash:         t.Raw.Hash.String(),
		TxIndex:        t.Raw.Index,
	}
	return txInfo
}

func (t Tx) GetEvents() (events []Event, err error) {
	for idx, e := range t.Raw.TxResult.Events {
		evt := Event{Type: e.Type, Index: idx}

		for _, a := range e.Attributes {
			// Make sure that the attribute value is a valid JSON encoded string.
			// Tendermint event attribute values contain JSON encoded values without quotes
			// so string values need to be encoded to be quoted and saved as valid JSONB.
			v, err := formatAttributeValue([]byte(a.Value))
			if err != nil {
				return nil, fmt.Errorf("error encoding event attr '%s.%s': %w", e.Type, a.Key, err)
			}

			evt.Attributes = append(evt.Attributes, EventAttribute{
				Key:   a.Key,
				Value: v,
			})
		}

		events = append(events, evt)
	}

	return events, nil
}

func formatAttributeValue(v []byte) ([]byte, error) {
	if json.Valid(v) {
		return v, nil
	}

	// Encode all string or invalid values
	return json.Marshal(string(v))
}
