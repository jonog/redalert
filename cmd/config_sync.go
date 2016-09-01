package cmd

import (
	"log"

	"github.com/jonog/redalert/config"
	"github.com/spf13/cobra"
)

// configSyncCmd represents the config-sync command
var configSyncCmd = &cobra.Command{
	Use:   "config-sync",
	Short: "Sync file and database configurations",
	Long:  "Sync file and database configurations",
	Run: func(cmd *cobra.Command, args []string) {

		if !cmd.Flag("config-db").Changed {
			log.Fatal("Please specify a database config")
		}
		configDb := cmd.Flag("config-db").Value.String()
		configFile := cmd.Flag("config-file").Value.String()
		syncConfigFileToDB(configFile, configDb)
	},
}

func init() {
	RootCmd.AddCommand(configSyncCmd)
}

func syncConfigFileToDB(file, db string) {

	fileConfigStore, err := config.NewFileStore(file)
	if err != nil {
		log.Fatal("Missing or invalid: ", file)
	}
	dbConfigStore, err := config.NewDBStore(db)
	if err != nil {
		log.Fatal("Unable to initialise db via :", db, " Error: ", err)
	}

	err = config.Copy(fileConfigStore, dbConfigStore)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("file to db sync complete")
}
