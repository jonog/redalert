package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// RootCmd configures the CLI
var RootCmd = &cobra.Command{
	Use:   "redalert",
	Short: "Monitor infrastructure and trigger alerts",
	Long: `For monitoring your infrastructure and sending notifications if stuff is not ok.

CHECKS:
*  website monitoring & latency measurement (check type: web-ping)
*  server metrics from local machine (check type: scollector)
*  Docker container metrics (check type: docker-stats)
*  Docker container metrics from remote host via SSH (check type: remote-docker)
*  Postgres counts/stats via SQL queries (check type: postgres)
*  TCP connectivity monitoring & latency measurement (check type: tcp)
*  execute local commands & capture output (check type: command)
*  execute remote commands via SSH & capture output (check type: remote-command)

ALERTS:
*  email (gmail)
*  SMS (twilio)
*  Slack (slack)
*  unix stream (stderr)
`,
}

// Execute parses the required flags and commands for the CLI
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

var cfgFile string
var cfgDb string
var port int

func init() {
	RootCmd.PersistentFlags().StringVarP(&cfgFile, "config-file", "f", "config.json", "config file")
	RootCmd.PersistentFlags().StringVarP(&cfgDb, "config-db", "d", "", "config database url")
	RootCmd.PersistentFlags().IntVarP(&port, "port", "p", 8888, "port to run server")
}
