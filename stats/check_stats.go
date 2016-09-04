package stats

import "github.com/jonog/redalert/utils"

type CheckStats struct {
	LastFailedAt        *occurrence
	LastSuccessfulAt    *occurrence
	LastCheckedAt       *occurrence
	StateTransitionedAt *occurrence
	SuccessfulTotal     *counter
	SuccessfulSequence  *counter
	FailureTotal        *counter
	FailureSequence     *counter
	CurrentAlertCount   *counter
}

func NewCheckStats() *CheckStats {
	return &CheckStats{
		LastFailedAt:        newOccurrence(),
		LastSuccessfulAt:    newOccurrence(),
		LastCheckedAt:       newOccurrence(),
		StateTransitionedAt: newOccurrence(),
		SuccessfulTotal:     newCounter(),
		SuccessfulSequence:  newCounter(),
		FailureTotal:        newCounter(),
		FailureSequence:     newCounter(),
		CurrentAlertCount:   newCounter(),
	}
}

type CheckStatsPublic struct {
	LastCheckedAt       *utils.RFCTime `json:"last_checked_at"`
	StateTransitionedAt *utils.RFCTime `json:"state_transitioned_at"`
	SuccessfulTotal     int            `json:"successful_total"`
	FailureTotal        int            `json:"failure_total"`
}

func (c *CheckStats) Export() CheckStatsPublic {
	stats := CheckStatsPublic{
		SuccessfulTotal: c.SuccessfulTotal.count,
		FailureTotal:    c.FailureTotal.count,
	}
	if !c.LastCheckedAt.t.IsZero() {
		stats.LastCheckedAt = &utils.RFCTime{c.LastCheckedAt.t}
	}
	if !c.StateTransitionedAt.t.IsZero() {
		stats.StateTransitionedAt = &utils.RFCTime{c.StateTransitionedAt.t}
	}
	return stats
}
