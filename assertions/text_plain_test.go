package assertions

import (
	"testing"

	"github.com/jonog/redalert/data"
)

func testTextPlainConfig() Config {
	return Config{
		Source:     "text/plain",
		Identifier: "",
		Comparison: "==",
		Target:     "HEALTH_CHECK_OK",
	}
}

func TestTextPlain_Equals_AssertSuccess(t *testing.T) {
	a, err := New(testTextPlainConfig(), testLog())
	if err != nil {
		t.Fail()
	}
	response := []byte("HEALTH_CHECK_OK")
	outcome, err := a.Assert(Options{CheckResponse: data.CheckResponse{Response: response}})
	if err != nil {
		t.Fail()
	}
	if !outcome.Assertion {
		t.Fatalf("expect: %#v, got: %#v", true, outcome.Assertion)
	}
	if outcome.Message != "" {
		t.Fatalf("expect: %#v, got: %#v", "", outcome.Message)
	}
}

func TestTextPlain_Equals_AssertFailure(t *testing.T) {
	a, err := New(testTextPlainConfig(), testLog())
	if err != nil {
		t.Fail()
	}
	response := []byte("HEALTH_CHECK_NOT_OK")
	outcome, err := a.Assert(Options{CheckResponse: data.CheckResponse{Response: response}})
	if err != nil {
		t.Fail()
	}
	if outcome.Assertion {
		t.Fatalf("expect: %#v, got: %#v", false, outcome.Assertion)
	}
	expectedMessage := "(HEALTH_CHECK_NOT_OK) is not equal to HEALTH_CHECK_OK"
	if outcome.Message != expectedMessage {
		t.Fatalf("expect: %#v, got: %#v", expectedMessage, outcome.Message)
	}
}
