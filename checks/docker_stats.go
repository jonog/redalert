package checks

import (
	"bufio"
	"encoding/json"
	"errors"
	"log"
	"sync"
	"time"

	"golang.org/x/net/context"

	"github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
	"github.com/jonog/redalert/data"
)

func init() {
	Register("docker_stats", NewDockerStats)
}

type DockerStats struct {
	log *log.Logger
}

var DockerStatsMetrics = map[string]MetricInfo{}

type DockerStatsConfig struct {
}

var NewDockerStats = func(config Config, logger *log.Logger) (Checker, error) {
	var commandConfig DockerStatsConfig
	err := json.Unmarshal([]byte(config.Config), &commandConfig)
	if err != nil {
		return nil, err
	}
	return Checker(&DockerStats{logger}), nil
}

func (c *DockerStats) Check() (data.CheckResponse, error) {

	// Note: for each container, this streams 2 messages from the docker stats endpoint
	// this would be a good candidate for a CheckStream() mode which maintains a conn

	response := data.CheckResponse{
		Metrics:  data.Metrics(make(map[string]*float64)),
		Metadata: make(map[string]string),
	}

	client, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	options := types.ContainerListOptions{All: true}
	containers, err := client.ContainerList(context.Background(), options)
	if err != nil {
		return response, err
	}

	allStats := make(chan map[string]float64, len(containers))
	doneChan := make(chan struct{})
	errorChan := make(chan error)

	var wg sync.WaitGroup
	wg.Add(len(containers))
	for _, container := range containers {
		go func(container types.Container) {
			defer wg.Done()
			containerStats, err := c.fetchStats(client, &container) //Name, container.ID)
			if err != nil {
				errorChan <- err
			} else {
				allStats <- containerStats
			}
		}(container)
	}
	go func() {
		defer func() {
			close(doneChan)
			close(allStats)
		}()
		wg.Wait()
	}()

	select {
	case <-doneChan:
		for cStats := range allStats {
			for k, v := range cStats {
				// assignment to avoid ref to iterator pointer
				value := v
				response.Metrics[k] = &value
			}
		}
		return response, nil
	case err := <-errorChan:
		return response, err
	case <-time.After(time.Second * 20):
		return response, errors.New("timeout")
	}
}

func (c *DockerStats) fetchStats(client *client.Client, container *types.Container) (map[string]float64, error) {

	c.log.Println("Fetching stats for container:", container.ID)
	stats := make(map[string]float64)
	containerName, err := getContainerName(container.Names)
	if err != nil {
		return stats, err
	}

	resp, err := client.ContainerStats(context.Background(), container.ID, true)
	if err != nil {
		return stats, err
	}

	dockerStats := []types.Stats{}
	scanner := bufio.NewScanner(resp)
	for scanner.Scan() {
		var latestStats types.Stats
		err = json.Unmarshal(scanner.Bytes(), &latestStats)
		if err != nil {
			break
		}
		dockerStats = append(dockerStats, latestStats)
		if len(dockerStats) == 2 {
			break
		}
	}
	if err != nil {
		return stats, err
	}
	if err := scanner.Err(); err != nil {
		return stats, err
	}
	if len(dockerStats) != 2 {
		return stats, errors.New("docker_stats: unable to retrieve stats")
	}

	return compareStats(containerName, &dockerStats[0], &dockerStats[1]), nil
}

func compareStats(containerName string, s1, s2 *types.Stats) map[string]float64 {
	stats := make(map[string]float64)
	stats[containerName+"_memory_usage"] = float64(s2.MemoryStats.Usage / 1000000.0)
	cpuUsageDelta := float64(s2.CPUStats.CPUUsage.TotalUsage) - float64(s1.CPUStats.CPUUsage.TotalUsage)
	systemCPUUsageDelta := float64(s2.CPUStats.SystemUsage) - float64(s1.CPUStats.SystemUsage)
	stats[containerName+"_cpu_usage_percentage"] = cpuUsageDelta * 100 / systemCPUUsageDelta
	return stats
}

func (c *DockerStats) MetricInfo(metric string) MetricInfo {
	return DockerStatsMetrics[metric]
}

func (c *DockerStats) MessageContext() string {
	return ""
}
