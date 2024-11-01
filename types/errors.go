package types

import "errors"

var (
	ErrBlockNotFound              = errors.New("block not found")
	ErrInvalidBlockRange          = errors.New("invalid block range")
	ErrNoFpHasVotingPower         = errors.New("no FP has voting power for the consumer chain")
	ErrBtcStakingNotActivated     = errors.New("BTC staking is not activated for the consumer chain")
	ErrActivatedTimestampNotFound = errors.New("BTC staking activated timestamp not found")
)
