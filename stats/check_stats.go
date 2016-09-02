package stats

import "github.com/jonog/redalert/utils"

type CheckStats struct {
	LastFailedAt       *occurrence
	LastSuccessfulAt   *occurrence
	LastCheckedAt      *occurrence
	SuccessfulTotal    *counter
	SuccessfulSequence *counter
	FailureTotal       *counter
	FailureSequence    *counter
	CurrentAlertCount  *counter
}

func NewCheckStats() *CheckStats {
	return &CheckStats{
		LastFailedAt:       newOccurrence(),
		LastSuccessfulAt:   newOccurrence(),
		LastCheckedAt:      newOccurrence(),
		SuccessfulTotal:    newCounter(),
		SuccessfulSequence: newCounter(),
		FailureTotal:       newCounter(),
		FailureSequence:    newCounter(),
		CurrentAlertCount:  newCounter(),
	}
}

type CheckStatsPublic struct {
	LastFailedAt       *utils.RFCTime `json:"last_failed_at"`
	LastSuccessfulAt   *utils.RFCTime `json:"last_successful_at"`
	LastCheckedAt      *utils.RFCTime `json:"last_checked_at"`
	SuccessfulTotal    int            `json:"successful_total"`
	SuccessfulSequence int            `json:"successful_sequence"`
	FailureTotal       int            `json:"failure_total"`
	FailureSequence    int            `json:"failure_sequence"`
}

func (c *CheckStats) Export() CheckStatsPublic {
	stats := CheckStatsPublic{
		SuccessfulTotal:    c.SuccessfulTotal.count,
		SuccessfulSequence: c.SuccessfulSequence.count,
		FailureTotal:       c.FailureTotal.count,
		FailureSequence:    c.FailureSequence.count,
	}
	if !c.LastFailedAt.t.IsZero() {
		stats.LastFailedAt = &utils.RFCTime{c.LastFailedAt.t}
	}
	if !c.LastSuccessfulAt.t.IsZero() {
		stats.LastSuccessfulAt = &utils.RFCTime{c.LastSuccessfulAt.t}
	}
	if !c.LastCheckedAt.t.IsZero() {
		stats.LastCheckedAt = &utils.RFCTime{c.LastCheckedAt.t}
	}
	return stats
}
