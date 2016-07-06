package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of Redalert",
	Long:  "Print the version number of Redalert",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Redalert v0.1.0")
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}
