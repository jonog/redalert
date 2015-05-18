package main

import (
	"log"
	"os"

	"github.com/jonog/redalert/core"

	"github.com/jonog/redalert/web"
)

func main() {

	config, err := ReadConfigFile()
	if err != nil {
		panic("Missing or invalid config")
	}

	service := core.NewService()

	// Setup Notifications

	ConfigureStdErr(service)
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
