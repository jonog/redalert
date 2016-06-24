package checks

import (
	"encoding/json"
	"fmt"
	"testing"

	_ "github.com/lib/pq"
)

func testPostgresConfig(host, port string) []byte {
	json := `
			{
					"name": "Sample Postgres",
					"type": "postgres",
					"config": {
							"connection_url": "postgres://postgres@` + host + ":" + port + `/postgres?sslmode=disable",
							"metric_queries": [
									{
											"metric": "emoji_count",
											"query": "select count(*) from emojis"
									}
							]
					},
					"send_alerts": [
							"stderr"
					],
					"backoff": {
							"interval": 120,
							"type": "linear"
					}
			}`
	return []byte(json)
}

func TestPostgres_ParseAndInitialise(t *testing.T) {
	var config Config
	err := json.Unmarshal(testPostgresConfig("localhost", "5432"), &config)
	if err != nil {
		t.Fatalf("error: %#v", err)
	}
	_, err = New(config, testLog())
	if err != nil {
		t.Fatalf("error: %#v", err)
	}
}

func TestPostgres_Check(t *testing.T) {

	container, err := setupPostgresContainer()
	if err != nil {
		t.Fatalf("error: %#v", err)
	}
	defer removePostgresContainer(container.ID)

	host, err := getHost()
	if err != nil {
		t.Fatalf("error: %#v, host: %#v", err, host)
	}
	port := container.NetworkSettings.Ports["5432/tcp"][0].HostPort

	fmt.Println("host", host)
	fmt.Println("port", port)

	var config Config
	err = json.Unmarshal(testPostgresConfig(host, port), &config)
	if err != nil {
		t.Fatalf("error: %#v", err)
	}
	checker, err := New(config, testLog())
	if err != nil {
		t.Fatalf("error: %#v", err)
	}

	waitForTCP(host + ":" + port)

	prepareDatabase(host + ":" + port)
	if err != nil {
		t.Fatalf("error: %#v", err)
	}

	checkData, err := checker.Check()
	if err != nil {
		t.Fatalf("error: %#v", err)
	}

	count, ok := checkData.Metrics["emoji_count"]
	if !ok || count == nil {
		t.Fatalf("Expected metric emoji_count does not exist. metrics: %#v", checkData.Metrics)
	}

	if *count != 3 {
		t.Fatalf("Invalid count")
	}

}
