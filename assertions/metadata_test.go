package assertions

import (
	"log"
	"os"
	"testing"

	"github.com/jonog/redalert/data"
)

func testMetadataConfig() Config {
	return Config{
		Source:     "metadata",
		Identifier: "status_code",
		Comparison: "==",
		Target:     "200",
	}
}

func TestMetadata_Equals_AssertSuccess(t *testing.T) {
	a, err := New(testMetadataConfig(), testLog())
	if err != nil {
		t.Fail()
	}
	metadata := map[string]string{"status_code": "200"}
	outcome, err := a.Assert(Options{CheckResponse: data.CheckResponse{Metadata: metadata}})
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

func TestMetadata_Equals_AssertFailure(t *testing.T) {
	a, err := New(testMetadataConfig(), testLog())
	if err != nil {
		t.Fail()
	}
	metadata := map[string]string{"status_code": "500"}
	outcome, err := a.Assert(Options{CheckResponse: data.CheckResponse{Metadata: metadata}})
	if err != nil {
		t.Fail()
	}
	if outcome.Assertion {
		t.Fatalf("expect: %#v, got: %#v", false, outcome.Assertion)
	}
	expectedMessage := "status_code (500) is not equal to 200"
	if outcome.Message != expectedMessage {
		t.Fatalf("expect: %#v, got: %#v", expectedMessage, outcome.Message)
	}
}

func testLog() *log.Logger {
	return log.New(os.Stdout, "", log.Ldate|log.Ltime)
}
