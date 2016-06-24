package checks

import (
	"encoding/json"
	"testing"
)

func testDockerStatsConfig() []byte {
	json := `
			{
					"name": "Sample DockerStats",
					"type": "docker_stats",
					"config": {},
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

func TestDockerStats_ParseAndInitialise(t *testing.T) {
	var config Config
	err := json.Unmarshal(testDockerStatsConfig(), &config)
	if err != nil {
		t.Fatalf("error: %#v", err)
	}
	_, err = New(config, testLog())
	if err != nil {
		t.Fatalf("error: %#v", err)
	}
}

func TestDockerStats_Check(t *testing.T) {

	// use postgres image as a test container

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

	var config Config
	err = json.Unmarshal(testDockerStatsConfig(), &config)
	if err != nil {
		t.Fatalf("error: %#v", err)
	}
	checker, err := New(config, testLog())
	if err != nil {
		t.Fatalf("error: %#v", err)
	}

	waitForTCP(host + ":" + port)

	checkData, err := checker.Check()
	if err != nil {
		t.Fatalf("error: %#v", err)
	}

	name, err := getContainerName([]string{container.Name})
	if err != nil {
		t.Fatalf("error: %#v", err)
	}

	expectedStats := []string{"memory_usage", "cpu_usage_percentage"}
	for _, stat := range expectedStats {
		if _, exists := checkData.Metrics[name+"_"+stat]; !exists {
			t.Fatalf("missing %s stat for container %s", name, stat)
		}
	}
}
