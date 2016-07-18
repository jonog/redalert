package assertions

import (
	"testing"

	"github.com/jonog/redalert/data"
)

var jsonSuccessTests = []struct {
	config   Config
	response []byte
}{
	{Config{
		Source:     "json",
		Identifier: "status",
		Comparison: "==",
		Target:     "HEALTH_CHECK_OK",
	}, []byte(`{"status": "HEALTH_CHECK_OK"}`)},
	{Config{
		Source:     "json",
		Identifier: "status.service",
		Comparison: "==",
		Target:     "HEALTH_CHECK_OK",
	}, []byte(`{"status": {"service": "HEALTH_CHECK_OK"}}`)},
	{Config{
		Source:     "json",
		Identifier: "status.1",
		Comparison: "==",
		Target:     "OK",
	}, []byte(`{"status": ["NOT_OK", "OK"]}`)},
	{Config{
		Source:     "json",
		Identifier: "services.1.status",
		Comparison: "==",
		Target:     "OK",
	}, []byte(`{"services": [{"status": "NOT_OK"},{"status": "OK"}]}`)},
}

func TestJSON_SuccessScenarios(t *testing.T) {
	for _, tt := range jsonSuccessTests {
		a, err := New(tt.config, testLog())
		if err != nil {
			t.Fail()
		}
		response := []byte(tt.response)
		outcome, err := a.Assert(Options{CheckResponse: data.CheckResponse{Response: response}})
		if err != nil {
			t.Fail()
		}
		if !outcome.Assertion {
			t.Fatalf("expect: %#v, got: %#v", true, outcome.Assertion)
		}
	}
}

var jsonFailureTests = []struct {
	config   Config
	response []byte
}{
	{Config{
		Source:     "json",
		Identifier: "status",
		Comparison: "==",
		Target:     "HEALTH_CHECK_OK",
	}, []byte(`{"status": "HEALTH_CHECK_NOT_OK"}`)},
}

func TestJSON_FailureScenarios(t *testing.T) {
	for _, tt := range jsonFailureTests {
		a, err := New(tt.config, testLog())
		if err != nil {
			t.Fail()
		}
		response := []byte(tt.response)
		outcome, err := a.Assert(Options{CheckResponse: data.CheckResponse{Response: response}})
		if err != nil {
			t.Fail()
		}
		if outcome.Assertion {
			t.Fatalf("expect: %#v, got: %#v", false, outcome.Assertion)
		}

	}
}

var jsonErrorTests = []struct {
	config   Config
	response []byte
}{
	{Config{
		Source:     "json",
		Identifier: "doesnotexist",
		Comparison: "==",
		Target:     "HEALTH_CHECK",
	}, []byte(`{"status": "HEALTH_CHECK"}`)},
}

func TestJSON_ErrorScenarios(t *testing.T) {
	for _, tt := range jsonErrorTests {
		a, err := New(tt.config, testLog())
		if err != nil {
			t.Fail()
		}
		response := []byte(tt.response)
		_, err = a.Assert(Options{CheckResponse: data.CheckResponse{Response: response}})
		if err == nil {
			t.Fail()
		}
	}
}
