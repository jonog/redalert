package checks

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/nu7hatch/gouuid"

	"github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
	"github.com/docker/engine-api/types/container"
	"github.com/docker/engine-api/types/network"
	"github.com/docker/go-connections/nat"
	"golang.org/x/net/context"
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

func getHost() (string, error) {
	dockerHost := os.Getenv("DOCKER_HOST")
	if dockerHost == "" {
		return "127.0.0.1", nil
	}
	u, err := url.Parse(dockerHost)
	if err != nil {
		return "dockerHost: " + dockerHost, err
	}
	host, _, err := net.SplitHostPort(u.Host)
	return host, err
}

func prepareDatabase(address string) error {
	db, err := sql.Open("postgres", "postgres://postgres@"+address+"/postgres?sslmode=disable")
	if err != nil {
		return err
	}
	defer db.Close()

	waitForDBPing(db)

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS emojis(id serial primary key, name text NOT NULL);")
	if err != nil {
		return err
	}

	_, err = db.Exec("INSERT INTO emojis (name) VALUES ('hatched_chick'), ('boom'), ('neckbeard');")
	return err
}

func setupPostgresContainer() (*types.ContainerJSON, error) {

	client, err := client.NewEnvClient()
	if err != nil {
		return nil, err
	}

	emptyMap := make(map[nat.Port]struct{})
	containerConfig := container.Config{
		Image:        "postgres",
		ExposedPorts: emptyMap,
	}
	hostConfig := container.HostConfig{
		PublishAllPorts: true,
	}
	networkConfig := network.NetworkingConfig{}

	u4, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}

	container, err := client.ContainerCreate(context.Background(), &containerConfig, &hostConfig, &networkConfig, "test-container-"+u4.String())
	if err != nil {
		return nil, err
	}
	client.ContainerStart(context.Background(), container.ID, types.ContainerStartOptions{})
	if err != nil {
		return nil, err
	}

	containerData, err := client.ContainerInspect(context.Background(), container.ID)
	if err != nil {
		return nil, err
	}
	return &containerData, nil
}

func removePostgresContainer(containerID string) error {
	client, err := client.NewEnvClient()
	if err != nil {
		return err
	}
	return client.ContainerRemove(context.Background(), containerID, types.ContainerRemoveOptions{Force: true})
}

func waitForDBPing(db *sql.DB) {
	tryDBPing := func() bool {
		err := db.Ping()
		return err == nil
	}
	waitFor(tryDBPing, 1*time.Millisecond, 5*time.Second)
}

func waitForTCP(address string) {
	tryTCP := func() bool {
		conn, _ := net.DialTimeout("tcp", address, 500*time.Millisecond)
		return conn != nil
	}
	waitFor(tryTCP, 500*time.Millisecond, 10*time.Second)
}

func waitFor(predicateFunc func() bool, backoff time.Duration, timeout time.Duration) {
	waitChan := make(chan struct{})
	go func() {
		for {
			if predicateFunc() {
				break
			} else {
				log.Println("Waiting")
				time.Sleep(backoff)
			}
		}
		close(waitChan)
	}()
	select {
	case <-waitChan:
		break
	case <-time.After(timeout):
		log.Println("Timeout")
	}
}
