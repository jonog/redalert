package checks

import (
	"encoding/json"
	"testing"
)

func testTestReportConfig(cmd string) []byte {
	json := `
					{
							"name": "Smoke Tests",
							"type": "test-report",
							"config": {
									"command": "` + cmd + `"
							},
							"send_alerts": [
									"stderr"
							],
							"backoff": {
									"interval": 10,
									"type": "constant"
							}
					}`
	return []byte(json)
}

func TestTestReport_ParseAndInitialise(t *testing.T) {
	var config Config
	err := json.Unmarshal(testTestReportConfig("./scripts/test"), &config)
	if err != nil {
		t.Fatalf("error: %#v", err)
	}
	_, err = New(config, testLog())
	if err != nil {
		t.Fatalf("error: %#v", err)
	}
}

func TestTestReport_Check_PassingTests(t *testing.T) {
	var config Config
	err := json.Unmarshal(testTestReportConfig("cat fixtures/sample_junit_passing.xml"), &config)
	if err != nil {
		t.Fatalf("error: %#v", err)
	}
	checker, err := New(config, testLog())
	if err != nil {
		t.Fatalf("error: %#v", err)
	}
	data, err := checker.Check()
	if err != nil {
		t.Fatalf("error: %#v", err)
	}
	if data.Metadata["status"] != "PASSING" {
		t.Fatalf("expect: %#v, got: %#v", "PASSING", data.Metadata["status"])
	}

	testCount, ok := data.Metrics["test_count"]
	if !ok || testCount == nil {
		t.Fatalf("Expected metric test_count does not exist. metrics: %#v", data.Metrics)
	}
	if *testCount != 9 {
		t.Fatalf("Invalid test_count")
	}

	failureCount, ok := data.Metrics["failure_count"]
	if !ok || failureCount == nil {
		t.Fatalf("Expected metric failure_count does not exist. metrics: %#v", data.Metrics)
	}
	if *failureCount != 0 {
		t.Fatalf("Invalid failure_count")
	}

	passCount, ok := data.Metrics["pass_count"]
	if !ok || passCount == nil {
		t.Fatalf("Expected metric pass_count does not exist. metrics: %#v", data.Metrics)
	}
	if *passCount != 9 {
		t.Fatalf("Invalid pass_count")
	}

	passRate, ok := data.Metrics["pass_rate"]
	if !ok || passRate == nil {
		t.Fatalf("Expected metric pass_rate does not exist. metrics: %#v", data.Metrics)
	}
	if *passRate != 100 {
		t.Fatalf("Invalid pass_rate")
	}

}

func TestTestReport_Check_FailingTests(t *testing.T) {
	var config Config
	err := json.Unmarshal(testTestReportConfig("cat fixtures/sample_junit_failing.xml && exit 1"), &config)
	if err != nil {
		t.Fatalf("error: %#v", err)
	}
	checker, err := New(config, testLog())
	if err != nil {
		t.Fatalf("error: %#v", err)
	}
	data, err := checker.Check()
	if err != nil {
		t.Fatalf("error: %#v", err)
	}
	if data.Metadata["status"] != "FAILING" {
		t.Fatalf("expect: %#v, got: %#v", "FAILING", data.Metadata["status"])
	}

	testCount, ok := data.Metrics["test_count"]
	if !ok || testCount == nil {
		t.Fatalf("Expected metric test_count does not exist. metrics: %#v", data.Metrics)
	}
	if *testCount != 9 {
		t.Fatalf("Invalid test_count")
	}

	failureCount, ok := data.Metrics["failure_count"]
	if !ok || failureCount == nil {
		t.Fatalf("Expected metric failure_count does not exist. metrics: %#v", data.Metrics)
	}
	if *failureCount != 2 {
		t.Fatalf("Invalid failure_count")
	}

	passCount, ok := data.Metrics["pass_count"]
	if !ok || passCount == nil {
		t.Fatalf("Expected metric pass_count does not exist. metrics: %#v", data.Metrics)
	}
	if *passCount != 7 {
		t.Fatalf("Invalid pass_count")
	}

	passRate, ok := data.Metrics["pass_rate"]
	if !ok || passRate == nil {
		t.Fatalf("Expected metric pass_rate does not exist. metrics: %#v", data.Metrics)
	}
	if *passRate != 100*float64(7)/float64(9) {
		t.Fatalf("Invalid pass_rate")
	}

}
