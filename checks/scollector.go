package checks

import (
	"encoding/json"
	"fmt"
)

// TODO:
// add mutex to handle concurrent read/writes

var GlobalSCollector map[Host]CurrentMetrics = make(map[Host]CurrentMetrics)

type Host string
type CurrentMetrics map[string]float64

type SCollector struct {
	Host string
}

func NewSCollector(host string) *SCollector {
	return &SCollector{host}
}

func (sc *SCollector) Check() (map[string]float64, error) {

	fmt.Println("SCollector Check")

	_, exists := GlobalSCollector[Host(sc.Host)]
	if !exists {
		GlobalSCollector[Host(sc.Host)] = make(map[string]float64)
	}

	jsonB, err := json.MarshalIndent(GlobalSCollector[Host(sc.Host)], "", "\t")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(jsonB))

	output := make(map[string]float64)
	for key, val := range GlobalSCollector[Host(sc.Host)] {
		output[key] = val
	}

	return output, nil
}

func (sc *SCollector) MetricInfo(metric string) MetricInfo {
	return MetricInfo{Unit: ""}
}

func (sc *SCollector) RedAlertMessage() string {
	return "Uhoh fail on " + sc.Host
}

func (sc *SCollector) GreenAlertMessage() string {
	return "Woo-hoo, successful check on " + sc.Host
}
