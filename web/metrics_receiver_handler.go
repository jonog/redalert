package web

import (
	"compress/gzip"
	"encoding/json"
	"net/http"

	"github.com/jonog/redalert/checks"
)

type Metric struct {
	Timestamp int
	Metric    string
	Value     float64
	Tags      map[string]string
}

func metricsReceiverHandler(c *appCtx, w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		http.Error(w, "Invalid request method.", 405)
		return
	}

	raw, err := gzip.NewReader(r.Body)
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		return
	}

	var metrics []Metric
	err = json.NewDecoder(raw).Decode(&metrics)
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		return
	}

	for idx := range metrics {

		host, exists := metrics[idx].Tags["host"]
		if !exists || host == "" {
			continue
		}

		_, exists = checks.GlobalSCollector[checks.Host(host)]
		if !exists {
			checks.GlobalSCollector[checks.Host(host)] = make(map[string]*float64)
		}
		checks.GlobalSCollector[checks.Host(host)][metrics[idx].Metric] = &metrics[idx].Value
	}

	w.WriteHeader(204)
	w.Write([]byte(`OK`))
}
