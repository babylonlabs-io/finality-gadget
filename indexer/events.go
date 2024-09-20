package indexer

import (
	"encoding/json"

	"github.com/babylonlabs-io/finality-gadget/types"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

type Event struct {
	Index      int              `json:"index"`
	Type       string           `json:"type"`
	Attributes []EventAttribute `json:"attributes"`
}

func (idx *Indexer) ParseEvent(pgTx pgx.Tx, txInfo *types.TxInfo, evt Event) error {
	// Save event to `events` DB
	// err := idx.db.SaveEvent(pgTx, &types.Event{TxHash: txHash, Name: evt.Type})
	// if err != nil {
	// 	return err
	// }

	// Store event-specific entry to DB
	switch evt.Type {
	// case "message":
	// 	// Loop through attributes
	// 	var attrMap EventMessage
	// 	for _, attr := range evt.Attributes {
	// 		switch attr.Key {
	// 		case "action":
	// 			attrMap.Action = string(attr.Value)
	// 		case "sender":
	// 			attrMap.Sender = string(attr.Value)
	// 		case "module":
	// 			attrMap.Module = string(attr.Value)
	// 		case "msg_index":
	// 			attrMap.MsgIndex = string(attr.Value)
	// 		}
	// 	}
	// 	if attrMap.Action != "" && attrMap.Sender != "" && attrMap.Module != "" && attrMap.MsgIndex != "" {
	// 		fmt.Printf("[message] %v %v %v %v\n", attrMap.Action, attrMap.Sender, attrMap.Module, attrMap.MsgIndex)
	// 		// Save event to DB
	// 		_, err := pgTx.Exec(
	// 			ctx,
	// 			sqlInsertEventMessage,
	// 			attrMap.Action,
	// 			attrMap.Sender,
	// 			attrMap.Module,
	// 			attrMap.MsgIndex,
	// 		)
	// 		if err != nil {
	// 			return err
	// 		}
	// 	}
	// case "tx":
	// 	for _, attr := range evt.Attributes {
	// 		fmt.Printf("[tx] %v: %v\n", attr.Key, string(attr.Value))
	// 	}
	case "babylon.btcstaking.v1.EventNewFinalityProvider":
		parsed, err := idx.ParseEventNewFinalityProvider(evt)
		if err != nil {
			return err
		}
		err = idx.db.SaveEventNewFinalityProvider(pgTx, txInfo, evt.Index, parsed)
		if err != nil {
			return err
		}
	case "babylon.btcstaking.v1.EventBTCDelegationStateUpdate":
		parsed, err := idx.ParseEventBTCDelegationStateUpdate(evt)
		if err != nil {
			return err
		}
		err = idx.db.SaveEventBTCDelegationStateUpdate(pgTx, txInfo, evt.Index, parsed)
		if err != nil {
			return err
		}
	case "babylon.btcstaking.v1.EventSelectiveSlashing":
		parsed, err := idx.ParseEventSelectiveSlashing(evt)
		if err != nil {
			return err
		}
		err = idx.db.SaveEventSelectiveSlashing(pgTx, txInfo, evt.Index, parsed)
		if err != nil {
			return err
		}
	case "babylon.finality.v1.EventSlashedFinalityProvider":
		parsed, err := idx.ParseEventSlashedFinalityProvider(evt)
		if err != nil {
			return err
		}
		err = idx.db.SaveEventSlashedFinalityProvider(pgTx, txInfo, evt.Index, parsed)
		if err != nil {
			return err
		}
	}

	return nil
}

func (idx *Indexer) ParseEventNewFinalityProvider(evt Event) (*types.EventNewFinalityProvider, error) {
	idx.logger.Info("Parsing event", zap.String("type", evt.Type))
	var event types.EventNewFinalityProvider
	for _, attr := range evt.Attributes {
		switch attr.Key {
		case "fp":
			var fp types.FinalityProvider
			err := json.Unmarshal([]byte(string(attr.Value)), &fp)
			if err != nil {
				return nil, err
			}
			event.DescriptionMoniker = fp.Description.Moniker
			event.DescriptionIdentity = fp.Description.Identity
			event.DescriptionWebsite = fp.Description.Website
			event.DescriptionSecurityContact = fp.Description.SecurityContact
			event.DescriptionDetails = fp.Description.Details
			event.Commission = fp.Commission
			event.BabylonPkKey = fp.BabylonPk.Key
			event.BtcPk = fp.BtcPk
			event.PopBtcSigType = fp.Pop.BtcSigType
			event.PopBabylonSig = fp.Pop.BabylonSig
			event.PopBtcSig = fp.Pop.BtcSig
			event.MasterPubRand = fp.MasterPubRand
			event.RegisteredEpoch = fp.RegisteredEpoch
			event.SlashedBabylonHeight = fp.SlashedBabylonHeight
			event.SlashedBtcHeight = fp.SlashedBtcHeight
			event.ConsumerId = fp.ConsumerId
		case "msg_index":
			event.MsgIndex = string(attr.Value)
		}
	}
	return &event, nil
}

func (idx *Indexer) ParseEventBTCDelegationStateUpdate(evt Event) (*types.EventBTCDelegationStateUpdate, error) {
	idx.logger.Info("Parsing event", zap.String("type", evt.Type))
	var event types.EventBTCDelegationStateUpdate
	for _, attr := range evt.Attributes {
		switch attr.Key {
		case "staking_tx_hash":
			event.StakingTxHash = string(attr.Value)
		case "new_state":
			event.NewState = string(attr.Value)
		}
	}
	return &event, nil
}

func (idx *Indexer) ParseEventSelectiveSlashing(evt Event) (*types.EventSelectiveSlashing, error) {
	idx.logger.Info("Parsing event", zap.String("type", evt.Type))
	var event types.EventSelectiveSlashing
	for _, attr := range evt.Attributes {
		switch attr.Key {
		case "evidence":
			var evidence types.SelectiveSlashingEvidence
			err := json.Unmarshal([]byte(string(attr.Value)), &evidence)
			if err != nil {
				return nil, err
			}
			event.StakingTxHash = evidence.StakingTxHash
			event.FpBtcPk = evidence.FpBtcPk
			event.RecoveredFpBtcSk = evidence.RecoveredFpBtcSk
		}
	}
	return &event, nil
}

func (idx *Indexer) ParseEventSlashedFinalityProvider(evt Event) (*types.EventSlashedFinalityProvider, error) {
	idx.logger.Info("Parsing event", zap.String("type", evt.Type))
	var event types.EventSlashedFinalityProvider
	for _, attr := range evt.Attributes {
		switch attr.Key {
		case "evidence":
			var evidence types.Evidence
			err := json.Unmarshal([]byte(string(attr.Value)), &evidence)
			if err != nil {
				return nil, err
			}
			event.FpBtcPk = evidence.FpBtcPk
			event.BlockHeight = evidence.BlockHeight
			event.PubRand = evidence.PubRand
			event.CanonicalAppHash = evidence.CanonicalAppHash
			event.ForkAppHash = evidence.ForkAppHash
			event.CanonicalFinalitySig = evidence.CanonicalFinalitySig
			event.ForkFinalitySig = evidence.ForkFinalitySig
		}
	}
	return &event, nil
}
