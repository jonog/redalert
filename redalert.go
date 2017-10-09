package main // import "github.com/jonog/redalert"

import (
	"github.com/jonog/redalert/cmd"
	"github.com/jonog/redalert/utils"
)

var (
	version string
	commit  string
)

func main() {
	utils.RegisterVersionAndBuild(version, commit)
	cmd.Execute()
}
