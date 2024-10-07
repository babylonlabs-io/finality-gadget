package indexer

import (
	"encoding/hex"
	"encoding/json"

	"github.com/babylonlabs-io/finality-gadget/types"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"

	bstypes "github.com/babylonlabs-io/babylon/x/btcstaking/types"
	bftypes "github.com/babylonlabs-io/babylon/x/finality/types"
)

type Event struct {
	Index      int              `json:"index"`
	Type       string           `json:"type"`
	Attributes []EventAttribute `json:"attributes"`
}

func (idx *Indexer) ProcessEvent(pgTx pgx.Tx, txInfo *types.TxInfo, evt *Event) error {
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
		// Parse event
		parsed, err := idx.parseEventNewFinalityProvider(evt)
		if err != nil {
			return err
		}
		// If consumer id matches, save event to db
		if parsed.ConsumerId == idx.cfg.BabylonChainId {
			err = idx.db.SaveEventNewFinalityProvider(pgTx, txInfo, evt.Index, parsed)
			if err != nil {
				return err
			}
		}
	case "babylon.btcstaking.v1.EventBTCDelegationStateUpdate":
		// Parse and save event to db
		parsed, err := idx.parseEventBTCDelegationStateUpdate(evt)
		if err != nil {
			return err
		}
		// Note: we store all delegations, including those not to FPs in our consumer chain
		// to avoid an extra query here
		err = idx.db.SaveEventBTCDelegationStateUpdate(pgTx, txInfo, evt.Index, parsed)
		if err != nil {
			return err
		}
		// If btc delegation info not found, query for it and save in db
		del, err := idx.db.GetBTCDelegationInfo(parsed.StakingTxHash)
		if err != nil {
			return err
		}
		if del == nil {
			err = idx.queryAndStoreBTCDelegation(parsed.StakingTxHash)
			if err != nil {
				return err
			}
		}
	case "babylon.finality.v1.EventSlashedFinalityProvider":
		parsed, err := idx.parseEventSlashedFinalityProvider(evt)
		if err != nil {
			return err
		}
		err = idx.db.SaveEventSlashedFinalityProvider(pgTx, txInfo, evt.Index, parsed)
		if err != nil {
			return err
		}
	case "babylon.finality.v1.EventJailedFinalityProvider":
		parsed, err := idx.parseEventJailedFinalityProvider(evt)
		if err != nil {
			return err
		}
		err = idx.db.SaveEventJailedFinalityProvider(pgTx, txInfo, evt.Index, parsed)
		if err != nil {
			return err
		}
	case "babylon.finality.v1.EventUnjailedFinalityProvider":
		parsed, err := idx.parseEventUnjailedFinalityProvider(evt)
		if err != nil {
			return err
		}
		err = idx.db.SaveEventUnjailedFinalityProvider(pgTx, txInfo, evt.Index, parsed)
		if err != nil {
			return err
		}
	}
	return nil
}

func (idx *Indexer) parseEventNewFinalityProvider(evt *Event) (*types.EventNewFinalityProvider, error) {
	idx.logger.Info("Parsing event", zap.String("type", evt.Type))
	var event types.EventNewFinalityProvider
	for _, attr := range evt.Attributes {
		switch attr.Key {
		case "fp":
			var fp bstypes.FinalityProvider
			err := json.Unmarshal([]byte(string(attr.Value)), &fp)
			if err != nil {
				return nil, err
			}
			event.Addr = fp.Addr
			event.DescriptionMoniker = fp.Description.Moniker
			event.DescriptionIdentity = fp.Description.Identity
			event.DescriptionWebsite = fp.Description.Website
			event.DescriptionSecurityContact = fp.Description.SecurityContact
			event.DescriptionDetails = fp.Description.Details
			event.Commission = fp.Commission.BigInt().String()
			event.BtcPk = fp.BtcPk.MarshalHex()
			event.PopBtcSigType = fp.Pop.BtcSigType.String()
			event.PopBtcSig = hex.EncodeToString(fp.Pop.BtcSig)
			event.SlashedBabylonHeight = fp.SlashedBabylonHeight
			event.SlashedBtcHeight = fp.SlashedBtcHeight
			event.Jailed = fp.Jailed
			event.ConsumerId = fp.ConsumerId
			// case "msg_index":
			// 	event.MsgIndex = string(attr.Value)
		}
	}
	return &event, nil
}

func (idx *Indexer) parseEventBTCDelegationStateUpdate(evt *Event) (*types.EventBTCDelegationStateUpdate, error) {
	idx.logger.Info("Parsing event", zap.String("type", evt.Type))
	var event types.EventBTCDelegationStateUpdate
	for _, attr := range evt.Attributes {
		switch attr.Key {
		case "staking_tx_hash":
			event.StakingTxHash = string(attr.Value)
		case "new_state":
			var newState bstypes.BTCDelegationStatus
			err := json.Unmarshal([]byte(string(attr.Value)), &newState)
			if err != nil {
				return nil, err
			}
			event.NewState = newState
		}
	}
	return &event, nil
}

func (idx *Indexer) queryAndStoreBTCDelegation(stakingTxHash string) error {
	// Query for btc delegation info
	res, err := idx.bbnClient.QueryBTCDelegation(stakingTxHash)
	if err != nil {
		return err
	}
	// Save btc delegation info to db
	fpBtcPkList := make([]string, len(res.FpBtcPkList))
	for i, fpBtcPk := range res.FpBtcPkList {
		fpBtcPkList[i] = fpBtcPk.MarshalHex()
	}
	err = idx.db.SaveBTCDelegationInfo(&types.BTCDelegation{
		StakerAddr:       res.StakerAddr,
		BtcPk:            res.BtcPk.MarshalHex(),
		FpBtcPkList:      fpBtcPkList,
		StartHeight:      res.StartHeight,
		EndHeight:        res.EndHeight,
		TotalSat:         res.TotalSat,
		StakingTxHex:     res.StakingTxHex,
		SlashingTxHex:    res.SlashingTxHex,
		NumCovenantSigs:  uint32(len(res.CovenantSigs)),
		StakingOutputIdx: res.StakingOutputIdx,
		Active:           res.Active,
		StatusDesc:       res.StatusDesc,
		UnbondingTime:    res.UnbondingTime,
		ParamsVersion:    res.ParamsVersion,
	})
	if err != nil {
		return err
	}
	return nil
}

func (idx *Indexer) parseEventSlashedFinalityProvider(evt *Event) (*types.EventSlashedFinalityProvider, error) {
	idx.logger.Info("Parsing event", zap.String("type", evt.Type))
	var event types.EventSlashedFinalityProvider
	for _, attr := range evt.Attributes {
		switch attr.Key {
		case "evidence":
			var evidence bftypes.Evidence
			err := json.Unmarshal([]byte(string(attr.Value)), &evidence)
			if err != nil {
				return nil, err
			}
			event.FpBtcPk = evidence.FpBtcPk.MarshalHex()
			event.BlockHeight = evidence.BlockHeight
			event.PubRand = evidence.PubRand.MarshalHex()
			event.CanonicalAppHash = hex.EncodeToString(evidence.CanonicalAppHash)
			event.ForkAppHash = hex.EncodeToString(evidence.ForkAppHash)
			event.CanonicalFinalitySig = evidence.CanonicalFinalitySig.ToHexStr()
			event.ForkFinalitySig = evidence.ForkFinalitySig.ToHexStr()
		}
	}
	return &event, nil
}

func (idx *Indexer) parseEventJailedFinalityProvider(evt *Event) (*types.EventJailedFinalityProvider, error) {
	idx.logger.Info("Parsing event", zap.String("type", evt.Type))
	var event types.EventJailedFinalityProvider
	for _, attr := range evt.Attributes {
		switch attr.Key {
		case "public_key":
			event.PublicKey = attr.Value
		}
	}
	return &event, nil
}

func (idx *Indexer) parseEventUnjailedFinalityProvider(evt *Event) (*types.EventUnjailedFinalityProvider, error) {
	idx.logger.Info("Parsing event", zap.String("type", evt.Type))
	var event types.EventUnjailedFinalityProvider
	for _, attr := range evt.Attributes {
		switch attr.Key {
		case "public_key":
			event.PublicKey = attr.Value
		}
	}
	return &event, nil
}
