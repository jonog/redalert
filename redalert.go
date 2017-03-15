package main // import "github.com/jonog/redalert"

import (
	"github.com/jonog/redalert/cmd"
	"github.com/jonog/redalert/utils"
)

var (
	Version string
	Build   string
)

func main() {
	utils.RegisterVersionAndBuild(Version, Build)
	cmd.Execute()
}
