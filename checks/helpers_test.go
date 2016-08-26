package checks

import (
	"database/sql"
	"log"
	"net"
	"net/url"
	"os"
	"time"

	"github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
	"github.com/docker/engine-api/types/container"
	"github.com/docker/engine-api/types/network"
	"github.com/docker/go-connections/nat"
	"github.com/nu7hatch/gouuid"
	"golang.org/x/net/context"
)

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
