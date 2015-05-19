package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/jonog/redalert/checks"
	"github.com/jonog/redalert/core"
	"github.com/jonog/redalert/notifiers"

	"github.com/jonog/redalert/web"
)

func main() {

	config, err := readConfig()
	if err != nil {
		panic("Missing or invalid config")
	}

	service := core.NewService()

	// Setup Notifications

	config.Notifications = append(config.Notifications, notifiers.Config{
		Name: "stderr",
		Type: "stderr",
	})
	for _, notificationConfig := range config.Notifications {
		err = service.RegisterNotifier(notificationConfig)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Setup Checks

	for _, checkConfig := range config.Checks {
		err = service.RegisterCheck(checkConfig)
		if err != nil {
			log.Fatal(err)
		}
	}

	service.Start()

	go web.Run(service, getPort())

	service.KeepRunning()

}

func getPort() string {
	if os.Getenv("RA_PORT") == "" {
		return "8888"
	}
	return os.Getenv("RA_PORT")
}

type Config struct {
	Checks        []checks.Config    `json:"checks"`
	Notifications []notifiers.Config `json:"notifications"`
}

func readConfig() (*Config, error) {
	file, err := ioutil.ReadFile("config.json")
	if err != nil {
		return nil, err
	}
	var config Config
	err = json.Unmarshal(file, &config)
	return &config, err
}
