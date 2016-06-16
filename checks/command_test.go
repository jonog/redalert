package checks

import (
	"encoding/json"
	"log"
	"os"
	"testing"
)

func testCommandConfig(cmd string) []byte {
	json := `
					{
							"name": "Sample Command",
							"type": "command",
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

func TestCommand_ParseAndInitialise(t *testing.T) {
	var config Config
	err := json.Unmarshal(testCommandConfig("echo hi"), &config)
	if err != nil {
		t.Fatalf("error: %#v", err)
	}
	_, err = New(config, testLog())
	if err != nil {
		t.Fatalf("error: %#v", err)
	}
}

func TestCommand_Check(t *testing.T) {
	var config Config
	err := json.Unmarshal(testCommandConfig("echo hi"), &config)
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
	if data.Metadata["exit_status"] != "0" {
		t.Fatalf("expect: %#v, got: %#v", "0", data.Metadata["exit_status"])
	}
	if string(data.Response) != "hi\n" {
		t.Fatalf("expect: %#v, got: %#v", "hi", string(data.Response))
	}
}

func TestCommand_Check_MetadataExitStatus(t *testing.T) {
	var config Config
	err := json.Unmarshal(testCommandConfig("exit 111"), &config)
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
	if data.Metadata["exit_status"] != "111" {
		t.Fatalf("expect: %#v, got: %#v", "111", data.Metadata["exit_status"])
	}
}

func testLog() *log.Logger {
	return log.New(os.Stdout, "", log.Ldate|log.Ltime)
}
