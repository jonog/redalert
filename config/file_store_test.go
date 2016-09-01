package config

import (
	"encoding/json"
	"io/ioutil"
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
