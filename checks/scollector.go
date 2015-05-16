package checks

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

	// Take a snapshot of data streaming into metrics receiver @ /api/put

	_, exists := GlobalSCollector[Host(sc.Host)]
	if !exists {
		GlobalSCollector[Host(sc.Host)] = make(map[string]float64)
	}

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
