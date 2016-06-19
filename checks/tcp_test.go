package checks

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func testTCPConfig(host string, port string) []byte {
	json := `
					{
							"name": "Sample TCP",
							"type": "tcp",
							"config": {
		             "host":"` + host + `",
		             "port": ` + port + `
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

func TestTCP_ParseAndInitialise(t *testing.T) {
	var config Config
	err := json.Unmarshal(testTCPConfig("httpstat.us", "80"), &config)
	if err != nil {
		t.Fatalf("error: %#v", err)
	}
	_, err = New(config, testLog())
	if err != nil {
		t.Fatalf("error: %#v", err)
	}
}

func TestTCP_Check(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "200 OK")
	}))
	defer ts.Close()
	url, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatalf("error: %#v", err)
	}
	hostPath := strings.Split(url.Host, ":")
	var config Config
	err = json.Unmarshal(testTCPConfig(hostPath[0], hostPath[1]), &config)
	if err != nil {
		t.Fatalf("error: %#v", err)
	}
	checker, err := New(config, testLog())
	if err != nil {
		t.Fatalf("error: %#v", err)
	}
	_, err = checker.Check()
	if err != nil {
		t.Fatalf("error: %#v", err)
	}
}

func TestTCP_Check_Fail(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "200 OK")
	}))
	url, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatalf("error: %#v", err)
	}
	hostPath := strings.Split(url.Host, ":")

	// stop server
	ts.Close()

	var config Config
	err = json.Unmarshal(testTCPConfig(hostPath[0], hostPath[1]), &config)
	if err != nil {
		t.Fatalf("error: %#v", err)
	}
	checker, err := New(config, testLog())
	if err != nil {
		t.Fatalf("error: %#v", err)
	}
	_, err = checker.Check()
	if err == nil {
		t.Fatal("error expected, check did not return error")
	}
}
