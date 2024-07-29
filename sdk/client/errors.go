package client

import "fmt"

var (
	ErrNoFpHasVotingPower     = fmt.Errorf("no FP has voting power for the consumer chain")
	ErrBtcStakingNotActivated = fmt.Errorf("BTC staking is not activated for the consumer chain")
)
