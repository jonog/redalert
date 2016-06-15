package assertions

import (
	"testing"

	"github.com/jonog/redalert/data"
)

func testMetricConfig(comparison, target string) Config {
	return Config{
		Source:     "metric",
		Identifier: "latency",
		Comparison: comparison,
		Target:     target,
	}
}

func TestMetric_AssertSuccess(t *testing.T) {
	cases := map[string]struct {
		Actual     float64
		Comparison string
		Target     string
	}{
		"<":    {4, "5", "<"},
		"<= 1": {4, "5", "<="},
		"<= 2": {5, "5", "<="},
		"==":   {5, "5", "=="},
		">= 1": {5, "5", ">="},
		">= 2": {6, "5", ">="},
		">":    {6, "5", ">"},
	}
	for k, tc := range cases {
		if !metricCaseAssertSuccess(k, tc.Actual, tc.Comparison, tc.Target) {
		}
	}
}

func metricCaseAssertSuccess(k string, actual float64, comparison, target string) bool {
	a, err := New(testMetricConfig(comparison, target), testLog())
	if err != nil {
		return false
	}
	metrics := data.Metrics(make(map[string]*float64))
	latency := float64(actual)
	metrics["latency"] = &latency
	outcome, err := a.Assert(Options{CheckResponse: data.CheckResponse{Metrics: metrics}})
	if err != nil {
		return false
	}
	if !outcome.Assertion {
		return false
	}
	if outcome.Message != "" {
		return false
	}
	return true
}
