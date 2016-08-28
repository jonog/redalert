package stats

type CheckStats struct {
	LastFailedAt       *occurrence
	LastSuccessfulAt   *occurrence
	LastCheckedAt      *occurrence
	SuccessfulTotal    *counter
	SuccessfulSequence *counter
	FailureTotal       *counter
	FailureSequence    *counter
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
	}
}
