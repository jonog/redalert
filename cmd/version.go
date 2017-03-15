package cmd

import (
	"fmt"

	"github.com/jonog/redalert/utils"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of Redalert",
	Long:  "Print the version number of Redalert",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Redalert v" + utils.Version() + " sha: " + utils.Build())
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}
