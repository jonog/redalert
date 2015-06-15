package checks

import (
	"encoding/json"
	"errors"
	"log"
)

func init() {
	Register("scollector", NewSCollector)
}

var GlobalSCollector map[Host]CurrentMetrics = make(map[Host]CurrentMetrics)

// TODO:
// add mutex to handle concurrent read/writes

type Host string
type CurrentMetrics map[string]*float64

type SCollector struct {
	Host string
}

type SCollectorConfig struct {
	Host string `json:"host"`
}

var NewSCollector = func(config Config, logger *log.Logger) (Checker, error) {
	var sCollectorConfig SCollectorConfig
	err := json.Unmarshal([]byte(config.Config), &sCollectorConfig)
	if err != nil {
		return nil, err
	}
	if sCollectorConfig.Host == "" {
		return nil, errors.New("scollector: host to collect stats via scollector cannot be blank")
	}
	return Checker(&SCollector{sCollectorConfig.Host}), nil
}

func (sc *SCollector) Check() (Metrics, error) {

	// Take a snapshot of data streaming into metrics receiver @ /api/put

	_, exists := GlobalSCollector[Host(sc.Host)]
	if !exists {
		GlobalSCollector[Host(sc.Host)] = make(map[string]*float64)
	}

	output := Metrics(make(map[string]*float64))
	for key, val := range GlobalSCollector[Host(sc.Host)] {
		output[key] = val
	}

	return output, nil
}

func (sc *SCollector) MetricInfo(metric string) MetricInfo {
	return MetricInfo{Unit: ""}
}

func (sc *SCollector) MessageContext() string {
	return sc.Host
}
