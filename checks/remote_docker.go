package checks

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/jonog/redalert/data"
	"github.com/jonog/redalert/utils"

	"golang.org/x/crypto/ssh"
)

func init() {
	Register("remote-docker", NewDockerRemoteDocker)
}

type RemoteDocker struct {
	User     string
	Password string
	Host     string
	Key      string
	Tool     string
	log      *log.Logger
}

type RemoteDockerConfig struct {
	User     string `json:"user"`
	Password string `json:"password"`
	Host     string `json:"host"`
	Key      string `json:"key"`
	Tool     string `json:"tool"`
}

var NewDockerRemoteDocker = func(config Config, logger *log.Logger) (Checker, error) {

	var remoteDockerConfig RemoteDockerConfig
	err := json.Unmarshal([]byte(config.Config), &remoteDockerConfig)
	if err != nil {
		return nil, err
	}

	tool := utils.StringDefault(remoteDockerConfig.Tool, "nc")
	if !utils.FindStringInArray(tool, []string{"nc", "socat"}) {
		return nil, errors.New("checks: unknown tool in remote docker config")
	}

	if remoteDockerConfig.User == "" {
		return nil, errors.New("remote-docker: user cannot be blank")
	}

	if remoteDockerConfig.Host == "" {
		return nil, errors.New("remote-docker: host cannot be blank")
	}

	return Checker(&RemoteDocker{
		User:     remoteDockerConfig.User,
		Password: remoteDockerConfig.Password,
		Host:     remoteDockerConfig.Host,
		Key:      remoteDockerConfig.Key,
		Tool:     tool,
		log:      logger,
	}), nil
}

func parseAndUnmarshal(raw string, data interface{}) error {

	httpRawSplit := strings.Split(raw, "\n\r\n")
	if len(httpRawSplit) < 2 {
		return errors.New("invalid format")
	}

	jsonStr := httpRawSplit[1]
	return json.Unmarshal([]byte(jsonStr), data)
}

func (r *RemoteDocker) dockerAPISocketAccess() string {
	if r.Tool == "nc" {
		return "nc -U /var/run/docker.sock"
	}
	if r.Tool == "socat" {
		return "socat - UNIX-CONNECT:/var/run/docker.sock"
	}
	return ""
}

func (r *RemoteDocker) dockerAPIStreamSocketAccess() string {
	if r.Tool == "nc" {
		return "nc -U /var/run/docker.sock"
	}
	if r.Tool == "socat" {
		return "socat -t 2 - UNIX-CONNECT:/var/run/docker.sock"
	}
	return ""
}

func (r *RemoteDocker) Check() (data.CheckResponse, error) {

	response := data.CheckResponse{
		Metrics: data.Metrics(make(map[string]*float64)),
	}

	auth := NewSSHAuthenticator(r.log, SSHAuthOptions{Password: r.Password, Key: r.Key})
	if len(auth.auths) == 0 {
		return response, errors.New("remote-docker: no SSH authentication methods available")
	}
	defer auth.Cleanup()

	client, err := ssh.Dial("tcp", r.Host+":"+"22", &ssh.ClientConfig{
		User: r.User,
		Auth: auth.auths,
	})
	if err != nil {
		return response, fmt.Errorf("remote-docker: error dialing ssh. err: %v", err)
	}
	defer client.Close()

	sshOutput, err := runCommandStrOutput(client, `echo -e "GET /containers/json HTTP/1.0\r\n" | `+r.dockerAPISocketAccess())
	if err != nil {
		return response, fmt.Errorf("remote-docker: error getting container list. err: %v", err)
	}

	if len(sshOutput) == 0 {
		r.log.Println("ERROR: cannot get list of containers from docker remote API")
		return response, errors.New("remote-docker: no data obtained when retrieving container list")
	}

	var containers []Container
	err = parseAndUnmarshal(sshOutput, &containers)
	if err != nil {
		return response, errors.New("remote-docker: unable to parse container list")
	}

	for _, c := range containers {

		cmd := `<<<'GET /containers/` + c.Id + `/stats HTTP/1.0'$'\r'$'\n' ` + r.dockerAPIStreamSocketAccess() + ` | head -6 | tail -2`

		sshOutput, err := runCommandStrOutput(client, cmd)
		if err != nil {
			r.log.Println("ERROR: unable to successfully ssh to obtain container stats", err)
			continue
		}

		if len(sshOutput) == 0 {
			r.log.Println("ERROR: cannot get container stats from docker remote API")
			continue
		}

		readings := strings.Split(sshOutput, "\n")
		if len(readings) < 2 {
			r.log.Println("ERROR: two readings were not obtained from docker remote API")
			continue
		}

		var containerStats1 ContainerStats
		err = json.Unmarshal([]byte(readings[0]), &containerStats1)
		if err != nil {
			r.log.Println("ERROR: unmarshalling container stats json (1st reading)", err)
			continue
		}
		var containerStats2 ContainerStats
		err = json.Unmarshal([]byte(readings[1]), &containerStats2)
		if err != nil {
			r.log.Println("ERROR: unmarshalling container stats json (2nd reading)", err)
			continue
		}

		containerName, err := getContainerName(c.Names)
		if err != nil {
			r.log.Println("ERROR: establishing container name", err)
			continue
		}

		// TODO: collect all the metrics
		containerMemory := float64(containerStats2.MemoryStats.Usage / 1000000.0)
		response.Metrics[containerName+".memory"] = &containerMemory

		cpuUsageDelta := float64(containerStats2.CpuStats.CpuUsage.TotalUsage) - float64(containerStats1.CpuStats.CpuUsage.TotalUsage)
		systemCpuUsageDelta := float64(containerStats2.CpuStats.SystemCpuUsage) - float64(containerStats1.CpuStats.SystemCpuUsage)
		cpuUsagePercent := cpuUsageDelta * 100 / systemCpuUsageDelta

		response.Metrics[containerName+".cpu"] = &cpuUsagePercent

	}

	containerCount := float64(len(containers))
	response.Metrics["container_count"] = &containerCount

	return response, nil
}

func (r *RemoteDocker) MetricInfo(metric string) MetricInfo {
	return MetricInfo{Unit: ""}
}

func (r *RemoteDocker) MessageContext() string {
	return "docker host - " + r.Host
}

type Container struct {
	Command string
	Created int
	Id      string
	Image   string
	Names   []string
	Ports   []PortConfig
	Status  string
}

type PortConfig struct {
	IP          string
	PrivatePort int
	PublicPort  int
	Type        string
}

type ContainerStats struct {
	Read    string `json:"read"`
	Network struct {
		RxDropped int `json:"rx_dropped"`
		RxBytes   int `json:"rx_bytes"`
		RxErrors  int `json:"rx_errors"`
		TxPackets int `json:"tx_packets"`
		TxDropped int `json:"tx_dropped"`
		RxPackets int `json:"rx_packets"`
		TxErrors  int `json:"tx_errors"`
		TxBytes   int `json:"tx_bytes"`
	} `json:"network"`
	MemoryStats struct {
		Stats struct {
			TotalRss int `json:"total_rss"`
			// TODO: add additional mem stats
		} `json:"stats"`
		MaxUsage int `json:"max_usage"`
		Usage    int `json:"usage"`
		Failcnt  int `json:"failcnt"`
		Limit    int `json:"limit"`
	} `json:"memory_stats"`
	CpuStats struct {
		CpuUsage struct {
			PercpuUsage       []int `json:"percpu_usage"`
			UsageInUsermode   int   `json:"usage_in_usermode"`
			TotalUsage        int   `json:"total_usage"`
			UsageInKernelmode int   `json:"usage_in_kernelmode"`
		} `json:"cpu_usage"`
		SystemCpuUsage int `json:"system_cpu_usage"`
	} `json:"cpu_stats"`
}
