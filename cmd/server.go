package cmd

import (
	"fmt"
	"log"

	"github.com/jonog/redalert/config"
	"github.com/jonog/redalert/core"
	"github.com/jonog/redalert/notifiers"
	"github.com/jonog/redalert/rpc"
	"github.com/jonog/redalert/storage"
	"github.com/jonog/redalert/web"
	"github.com/spf13/cobra"
)

type serverConfig struct {
	webPort      int
	disableBrand bool
	rpcPort      int
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Run checks and server stats",
	Long:  "Run checks and server stats",
	Run: func(cmd *cobra.Command, args []string) {
		if cmd.Flag("config-db").Changed && cmd.Flag("config-file").Changed {
			log.Fatal("Please specify only one config source")
		}
		var configStore config.Store
		var err error
		if cmd.Flag("config-db").Changed {
			log.Println("Config via db")
			configDb := cmd.Flag("config-db").Value.String()
			configStore, err = config.NewDBStore(configDb)
			if err != nil {
				log.Fatal("Unable to initialise db via :", configDb, " Error: ", err)
			}
		} else {
			log.Println("Config via file")
			configFile := cmd.Flag("config-file").Value.String()
			configStore, err = config.NewFileStore(configFile)
			if err != nil {
				log.Fatal("Missing or invalid format: ", configFile)
			}
		}
		runServer(configStore, serverConfig{
			webPort:      webPort,
			disableBrand: disableBrand,
			rpcPort:      rpcPort,
		})
	},
}

func init() {
	RootCmd.AddCommand(serverCmd)
}

func runServer(configStore config.Store, config serverConfig) {
	// Event Storage
	const MaxEventsStored = 100

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

		check, err := core.NewCheck(checkConfig, storage.NewMemoryList(MaxEventsStored))
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

	go web.Run(service, config.webPort, config.disableBrand)
	go rpc.Run(service, config.rpcPort)
	fmt.Println(`
____ ____ ___  ____ _    ____ ____ ___
|--< |=== |__> |--| |___ |=== |--<  |

`)
	fmt.Println("Web Running on port ", config.webPort)
	fmt.Println("RPC Running on port ", config.rpcPort)

	service.KeepRunning()
}
