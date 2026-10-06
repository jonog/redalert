package checks

import (
	"database/sql"
	"fmt"
	"net"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
	"github.com/docker/engine-api/types/container"
	"github.com/docker/engine-api/types/network"
	"github.com/docker/go-connections/nat"
	"github.com/nu7hatch/gouuid"
	"golang.org/x/net/context"
)

const sshFixtureImage = "sickp/alpine-sshd@sha256:0f5a58ba5bfc5549a910264f32c337903967bb377d596c91c03611f15b4699ad"

func getDockerHost() (string, error) {
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

func requireIntegration(t *testing.T) {
	t.Helper()
	if os.Getenv("REDALERT_INTEGRATION") != "1" {
		t.Skip("set REDALERT_INTEGRATION=1 to run Docker-backed integration tests")
	}
}

func prepareDatabase(address string) error {
	db, err := sql.Open("postgres", "postgres://postgres@"+address+"/postgres?sslmode=disable")
	if err != nil {
		return err
	}
	defer db.Close()

	if err := waitForDBPing(db); err != nil {
		return err
	}

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS emojis(id serial primary key, name text NOT NULL);")
	if err != nil {
		return err
	}

	_, err = db.Exec("INSERT INTO emojis (name) VALUES ('hatched_chick'), ('boom'), ('neckbeard');")
	return err
}

func setupPostgresContainer() (*types.ContainerJSON, error) {
	if imageName := os.Getenv("POSTGRES_IMAGE"); imageName != "" {
		return setupContainer(imageName)
	}
	return setupContainer("postgres")
}

func setupContainer(image string) (*types.ContainerJSON, error) {

	client, err := client.NewEnvClient()
	if err != nil {
		return nil, err
	}

	emptyMap := make(map[nat.Port]struct{})
	containerConfig := container.Config{
		Image:        image,
		ExposedPorts: emptyMap,
	}
	if image == "postgres" || image == "postgres:9.5" {
		containerConfig.Env = []string{"POSTGRES_HOST_AUTH_METHOD=trust"}
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
	err = client.ContainerStart(context.Background(), container.ID, types.ContainerStartOptions{})
	if err != nil {
		_ = client.ContainerRemove(context.Background(), container.ID, types.ContainerRemoveOptions{Force: true})
		return nil, fmt.Errorf("start fixture container %s: %w", image, err)
	}

	containerData, err := client.ContainerInspect(context.Background(), container.ID)
	if err != nil {
		_ = client.ContainerRemove(context.Background(), container.ID, types.ContainerRemoveOptions{Force: true})
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

func waitForDBPing(db *sql.DB) error {
	tryDBPing := func() bool {
		err := db.Ping()
		return err == nil
	}
	return waitFor(tryDBPing, 100*time.Millisecond, 30*time.Second)
}

func waitForTCP(address string) error {
	tryTCP := func() bool {
		conn, err := net.DialTimeout("tcp", address, 500*time.Millisecond)
		if err != nil {
			return false
		}
		_ = conn.Close()
		return true
	}
	return waitFor(tryTCP, 500*time.Millisecond, 10*time.Second)
}

func waitFor(predicateFunc func() bool, backoff time.Duration, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if predicateFunc() {
			return nil
		}
		time.Sleep(backoff)
	}
	return fmt.Errorf("timed out after %s waiting for fixture readiness", timeout)
}
