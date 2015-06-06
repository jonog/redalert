package main

import (
	"log"
	"os"

	"github.com/jonog/redalert/core"
	"github.com/jonog/redalert/notifiers"
	"github.com/jonog/redalert/storage"

	"github.com/jonog/redalert/web"
)

func main() {

	configStore, err := storage.NewConfigFile("config.json")
	if err != nil {
		log.Fatal("Missing or invalid config.json")
	}

	service := core.NewService()

	// Setup StdErr Notifications

	stdErrNotifier, err := notifiers.New(notifiers.Config{
		Name: "stderr",
		Type: "stderr",
	})
	if err != nil {
		log.Fatal(err)
	}

	err = service.RegisterNotifier(stdErrNotifier)
	if err != nil {
		log.Fatal(err)
	}

	// Load Notifications

	savedNotifications, err := configStore.Notifications()
	if err != nil {
		log.Fatal(err)
	}

	for _, notificationConfig := range savedNotifications {

		notifier, err := notifiers.New(notificationConfig)
		if err != nil {
			log.Fatal(err)
		}

		err = service.RegisterNotifier(notifier)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Load Checks

	savedChecks, err := configStore.Checks()
	if err != nil {
		log.Fatal(err)
	}

	for _, checkConfig := range savedChecks {

		check, err := core.NewCheck(checkConfig)
		if err != nil {
			log.Fatal(err)
		}

		err = service.RegisterCheck(check, checkConfig.SendAlerts)
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
