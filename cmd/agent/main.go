package main

import (
	"github.com/urfave/cli"
	"github.com/BoxLinker/cicd/version"
)

func main(){
	app := cli.NewApp()
	app.Name = "drone-agent"
	app.Version = version.Version.String()
	app.Usage = "drone agent"
	app.Action = loop
}

