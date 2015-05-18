package checks

type Config struct {
	Name       string   `json:"name"`
	Type       string   `json:"type"`
	Interval   int      `json:"interval"`
	SendAlerts []string `json:"send_alerts"`

	// used for web-ping
	Address string `json:"address"`

	// used for scollector
	Host string `json:"host"`
}

type MetricInfo struct {
	Unit string
}

type Metrics map[string]float64
