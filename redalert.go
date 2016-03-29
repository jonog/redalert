package main // import "github.com/jonog/redalert"

import (
	"flag"
	"log"
	"os"

	"github.com/jonog/redalert/core"
	"github.com/jonog/redalert/notifiers"
	"github.com/jonog/redalert/storage"

	"github.com/jonog/redalert/web"
)

func main() {

	var (
		configStorageType = flag.String("config", "file", "choice of config store: [file, db]")
		configFilename    = flag.String("config_file", "config.json", "path to json config")
		configDbURL       = flag.String("config_db", "postgres://user:pass@host/db_name", "connection url for config db")
	)
	flag.Parse()

	// `sync` command
	if len(flag.Args()) == 1 && flag.Args()[0] == "sync" {
		syncConfigFileToDB(*configFilename, *configDbURL)
		return
	}

	if *configStorageType != "file" && *configStorageType != "db" {
		log.Fatal("Invalid config store option.")
	}

	var configStore storage.ConfigStorage
	var err error

	switch *configStorageType {
	case "file":
		log.Println("Config via file")
		configStore, err = storage.NewConfigFile(*configFilename)
		if err != nil {
			log.Fatal("Missing or invalid config.json")
		}
	case "db":
		log.Println("Config via db")
		configStore, err = storage.NewConfigDB(*configDbURL)
		if err != nil {
			log.Fatal("Unable to initialise db via :", *configDbURL)
		}
	default:
		log.Fatal("Invalid config storage option: ", *configStorageType)
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
	var checkIdx int

	savedChecks, err := configStore.Checks()
	if err != nil {
		log.Fatal(err)
	}

	for _, checkConfig := range savedChecks {

		check, err := core.NewCheck(checkConfig)
		if err != nil {
			log.Fatal(err)
		}

		err = service.RegisterCheck(check, checkConfig.SendAlerts, checkIdx)
		if err != nil {
			log.Fatal(err)
		}
		checkIdx++
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
