package main

import (
	"log"

	"github.com/jonog/redalert/storage"
)

func syncConfigFileToDB(file, db string) {

	fileConfigStore, err := storage.NewConfigFile(file)
	if err != nil {
		log.Fatal("Missing or invalid config.json")
	}
	dbConfigStore, err := storage.NewConfigDB(db)
	if err != nil {
		log.Fatal("Unable to initialise db via :", db)
	}

	// Load Notifications

	savedNotifications, err := fileConfigStore.Notifications()
	if err != nil {
		log.Fatal(err)
	}

	for _, notificationConfig := range savedNotifications {

		_, err = dbConfigStore.CreateNotificationRecord(notificationConfig)
		if err != nil {
			log.Fatal(err)
		}

	}

	savedChecks, err := fileConfigStore.Checks()
	if err != nil {
		log.Fatal(err)
	}

	for _, checkConfig := range savedChecks {

		_, err = dbConfigStore.CreateCheckRecord(checkConfig)
		if err != nil {
			log.Fatal(err)
		}

	}

	log.Println("file to db sync complete")

}
