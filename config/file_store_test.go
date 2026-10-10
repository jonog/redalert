package config

import (
	"encoding/json"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/jonog/redalert/checks"
)

func testConfig() []byte {
	json := `
    {
        "checks": [
            {
                "name": "Demo HTTP Status Check",
                "type": "web-ping",
                "config": {
                    "address": "http://httpstat.us/200",
                    "headers": {
                        "X-Api-Key": "ABCD1234"
                    }
                },
                "send_alerts": [
                    "stderr"
                ],
                "backoff": {
                    "interval": 10,
                    "type": "constant"
                },
                "assertions": [
                    {
                        "comparison": "==",
                        "identifier": "status_code",
                        "source": "metadata",
                        "target": "200"
                    }
                ]
            }
        ],
        "notifications": [
            {
              "name": "sms-devops",
              "type": "twilio",
              "config": {
                "account_sid": "XXX",
                "auth_token": "YYY",
                "notification_numbers": "+0987654321",
                "twilio_number": "+1234567890"
              }
            }
        ]
    }`
	return []byte(json)
}

func TestYAMLFileStoreRewriteAndAppend(t *testing.T) {
	path := t.TempDir() + "/config.YML"
	input := `checks:
  - name: Demo
    type: web-ping
    config:
      address: http://localhost
      headers:
        X-Test: value
    assertions:
      - source: text
        comparison: ==
        target: "200"
notifications:
  - name: stderr
    type: stderr
    config: {}
preferences:
  `
	if err := ioutil.WriteFile(path, []byte(input), 0644); err != nil {
		t.Fatal(err)
	}
	store, err := NewFileStore(path)
	if err != nil {
		t.Fatal(err)
	}
	checksBefore, _ := store.Checks()
	notificationsBefore, _ := store.Notifications()
	if checksBefore[0].ID == "" || notificationsBefore[0].ID == "" {
		t.Fatal("IDs were not generated")
	}
	if !strings.Contains(string(mustRead(t, path)), "address: http://localhost") {
		t.Fatal("configuration was not rewritten as YAML")
	}
	if err := store.AppendChecks([]checks.Config{{ID: "added", Name: "Added", Type: "web-ping", Config: json.RawMessage(`{"address":"http://example"}`)}}); err != nil {
		t.Fatal(err)
	}
	reloaded, err := NewFileStore(path)
	if err != nil {
		t.Fatal(err)
	}
	after, _ := reloaded.Checks()
	if len(after) != 2 || after[0].ID != checksBefore[0].ID || after[1].ID != "added" {
		t.Fatalf("reloaded checks = %#v", after)
	}
	if string(after[0].Config) != `{"address":"http://localhost","headers":{"X-Test":"value"}}` {
		t.Fatalf("nested config = %s", after[0].Config)
	}
}

func mustRead(t *testing.T, path string) []byte {
	t.Helper()
	b, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestMalformedYAMLDoesNotOverwrite(t *testing.T) {
	path := t.TempDir() + "/broken.yaml"
	original := []byte("checks: [\n")
	if err := ioutil.WriteFile(path, original, 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := NewFileStore(path); err == nil {
		t.Fatal("expected malformed YAML error")
	}
	if got := mustRead(t, path); string(got) != string(original) {
		t.Fatalf("file changed: %q", got)
	}
}

func TestFileStore_Checks(t *testing.T) {
	err := ioutil.WriteFile("/tmp/test_file_store", testConfig(), 0644)
	if err != nil {
		t.Fatalf("error: %#v", err)
	}
	fs, err := NewFileStore("/tmp/test_file_store")
	if err != nil {
		t.Fatalf("error: %#v", err)
	}
	chks, err := fs.Checks()
	if err != nil {
		t.Fatalf("error: %#v", err)
	}
	if len(chks) != 1 {
		t.Fatalf("checks expected: %#v, got: %#v", 1, len(chks))
	}
	if chks[0].Name != "Demo HTTP Status Check" {
		t.Fatalf("expect: %#v, got: %#v", "Demo HTTP Status Check", chks[0].Name)
	}
	if chks[0].Type != "web-ping" {
		t.Fatalf("expect: %#v, got: %#v", "web-ping", chks[0].Type)
	}
	var wpConfig checks.WebPingerConfig
	err = json.Unmarshal(chks[0].Config, &wpConfig)
	if err != nil {
		t.Fatalf("error: %#v", err)
	}
	if wpConfig.Address != "http://httpstat.us/200" {
		t.Fatalf("expect: %#v, got: %#v", "http://httpstat.us/200", wpConfig.Address)
	}
	if wpConfig.Headers["X-Api-Key"] != "ABCD1234" {
		t.Fatalf("expect: %#v, got: %#v", "ABCD1234", wpConfig.Headers["X-Api-Key"])
	}
	if len(chks[0].SendAlerts) != 1 || chks[0].SendAlerts[0] != "stderr" {
		t.Fatalf("error with send alerts: %#v", chks[0].SendAlerts)
	}
	if chks[0].Backoff.Type != "constant" {
		t.Fatalf("expect: %#v, got: %#v", "constant", chks[0].Backoff.Type)
	}
	if chks[0].Backoff.Interval == nil || *chks[0].Backoff.Interval != 10 {
		t.Fatalf("error with backoff interval: %#v", chks[0].Backoff.Interval)
	}
	if len(chks[0].Assertions) != 1 {
		t.Fatalf("expected assertions: %#v, got: %#v", 1, len(chks[0].Assertions))
	}
	if chks[0].Assertions[0].Comparison != "==" ||
		chks[0].Assertions[0].Identifier != "status_code" ||
		chks[0].Assertions[0].Source != "metadata" ||
		chks[0].Assertions[0].Target != "200" {
		if err != nil {
			t.Fatalf("error with assertions: %#v", chks[0].Assertions[0])
		}
	}
}

func TestFileStore_CheckIDGeneration(t *testing.T) {
	err := ioutil.WriteFile("/tmp/test_file_store", testConfig(), 0644)
	if err != nil {
		t.Fatalf("error: %#v", err)
	}
	fs, err := NewFileStore("/tmp/test_file_store")
	if err != nil {
		t.Fatalf("error: %#v", err)
	}
	chks, err := fs.Checks()
	if err != nil {
		t.Fatalf("error: %#v", err)
	}
	if chks[0].ID == "" {
		t.Fatal("check ID not generated")
	}
}

func TestFileStore_Notifications(t *testing.T) {
	err := ioutil.WriteFile("/tmp/test_file_store", testConfig(), 0644)
	if err != nil {
		t.Fatalf("error: %#v", err)
	}
	fs, err := NewFileStore("/tmp/test_file_store")
	if err != nil {
		t.Fatalf("error: %#v", err)
	}
	n, err := fs.Notifications()
	if err != nil {
		t.Fatalf("error: %#v", err)
	}
	if len(n) != 1 {
		t.Fatalf("notifications expected: %#v, got: %#v", 1, len(n))
	}
	if n[0].Name != "sms-devops" {
		t.Fatalf("expect: %#v, got: %#v", "sms-devops", n[0].Name)
	}
	if n[0].Type != "twilio" {
		t.Fatalf("expect: %#v, got: %#v", "twilio", n[0].Type)
	}
	if n[0].Type != "twilio" {
		t.Fatalf("expect: %#v, got: %#v", "twilio", n[0].Type)
	}
	if n[0].Config["account_sid"] != "XXX" ||
		n[0].Config["auth_token"] != "YYY" ||
		n[0].Config["notification_numbers"] != "+0987654321" ||
		n[0].Config["twilio_number"] != "+1234567890" {
		t.Fatalf("error with notification config: %#v", n[0].Config)
	}
}

func TestFileStore_AppendChecksPreservesOtherValuesAndUpdatesMemory(t *testing.T) {
	path := t.TempDir() + "/config.json"
	if err := ioutil.WriteFile(path, testConfig(), 0644); err != nil {
		t.Fatal(err)
	}
	fs, err := NewFileStore(path)
	if err != nil {
		t.Fatal(err)
	}
	before, _ := fs.Checks()
	if err := fs.AppendChecks([]checks.Config{{ID: "stable-id", Name: "added", Type: "web-ping", Config: json.RawMessage(`{"address":"http://localhost"}`)}}); err != nil {
		t.Fatal(err)
	}
	after, err := fs.Checks()
	if err != nil {
		t.Fatal(err)
	}
	if len(after) != 2 || after[0].ID != before[0].ID || after[1].ID != "stable-id" {
		t.Fatalf("checks after append = %#v", after)
	}
	notifications, _ := fs.Notifications()
	if len(notifications) != 1 || notifications[0].Name != "sms-devops" {
		t.Fatalf("notifications changed: %#v", notifications)
	}
	reloaded, err := NewFileStore(path)
	if err != nil {
		t.Fatal(err)
	}
	persisted, _ := reloaded.Checks()
	if len(persisted) != 2 || persisted[1].ID != "stable-id" {
		t.Fatalf("persisted checks = %#v", persisted)
	}
}
