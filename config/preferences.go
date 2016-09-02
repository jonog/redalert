package config

type Preferences struct {
	Notifications NotificationPreferences `json:"notifications"`
}

type NotificationPreferences struct {

	// send an alert only after N fails (default to 1)
	FailCountAlertThreshold *int `json:"fail_count_alert_threshold,omitempty"`

	// continue to send fail alerts (default to false)
	RepeatFailAlerts *bool `json:"repeat_fail_alerts,omitempty"`
}

const DefaultFailCountAlertThreshold = 1
const DefaultRepeatFailAlerts = false
