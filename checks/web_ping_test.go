package checks

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func testWebPingConfig(address string) []byte {
	json := `
					{
							"name": "Sample WebPing",
							"type": "web-ping",
							"config": {
		             "address":"` + address + `",
		             "headers": {
		               "X-Api-Key": "ABCD1234",
		               "Host": "HostHeader"
		             }
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

func TestWebPing_ParseAndInitialise(t *testing.T) {
	var config Config
	err := json.Unmarshal(testWebPingConfig("http://httpstat.us/200"), &config)
	if err != nil {
		t.Fatalf("error: %#v", err)
	}
	_, err = New(config, testLog())
	if err != nil {
		t.Fatalf("error: %#v", err)
	}
}

func TestWebPing_Check(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "200 OK")
	}))
	defer ts.Close()
	var config Config
	err := json.Unmarshal(testWebPingConfig(ts.URL), &config)
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
	if data.Metadata["status_code"] != "200" {
		t.Fatalf("expect: %#v, got: %#v", "200", data.Metadata["exit_status"])
	}
	if string(data.Response) != "200 OK" {
		t.Fatalf("expect: %#v, got: %#v", "200 OK", string(data.Response))
	}
}

func TestWebPing_Check_Headers(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "%s %s", r.Host, r.Header.Get("X-API-Key"))
	}))
	defer ts.Close()
	var config Config
	err := json.Unmarshal(testWebPingConfig(ts.URL), &config)
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
	if data.Metadata["status_code"] != "200" {
		t.Fatalf("expect: %#v, got: %#v", "200", data.Metadata["exit_status"])
	}
	if string(data.Response) != "HostHeader ABCD1234" {
		t.Fatalf("expect: %#v, got: %#v", "HostHeader ABCD1234", string(data.Response))
	}
}
