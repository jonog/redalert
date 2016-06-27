package checks

import (
	"encoding/json"
	"testing"
)

func testRemoteCommandConfig(cmd, host, port string) []byte {
	json := `
					{
							"name": "Sample RemoteCommand",
							"type": "command",
							"config": {
									"command": "` + cmd + `",
									"ssh_auth_options": {
										"user": "root",
										"password": "root",
										"host": "` + host + `",
										"port": ` + port + `
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

func TestRemoteCommand_ParseAndInitialise(t *testing.T) {
	var config Config
	err := json.Unmarshal(testRemoteCommandConfig("echo hi", "localhost", "22"), &config)
	if err != nil {
		t.Fatalf("error: %#v", err)
	}
	_, err = New(config, testLog())
	if err != nil {
		t.Fatalf("error: %#v", err)
	}
}

func TestRemoteCommand_Check(t *testing.T) {

	container, err := setupContainer("sickp/alpine-sshd")
	if err != nil {
		t.Fatalf("error: %#v", err)
	}
	defer removePostgresContainer(container.ID)

	host, err := getDockerHost()
	if err != nil {
		t.Fatalf("error: %#v, host: %#v", err, host)
	}
	port := container.NetworkSettings.Ports["22/tcp"][0].HostPort

	var config Config
	err = json.Unmarshal(testRemoteCommandConfig("echo hi", host, port), &config)
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

func TestRemoteCommand_Check_MetadataExitStatus(t *testing.T) {

	container, err := setupContainer("sickp/alpine-sshd")
	if err != nil {
		t.Fatalf("error: %#v", err)
	}
	defer removePostgresContainer(container.ID)

	host, err := getDockerHost()
	if err != nil {
		t.Fatalf("error: %#v, host: %#v", err, host)
	}
	port := container.NetworkSettings.Ports["22/tcp"][0].HostPort

	var config Config
	err = json.Unmarshal(testRemoteCommandConfig("exit 111", host, port), &config)
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
		t.Fatalf("expect: %#v, got: %#v", "0", data.Metadata["exit_status"])
	}

}
