package main

import (
	"github.com/urfave/cli"
	"github.com/BoxLinker/cicd/version"
	_ "github.com/joho/godotenv/autoload"
	"fmt"
	"os"
)

func main(){
	app := cli.NewApp()
	app.Name = "drone-server"
	app.Version = version.Version.String()
	app.Usage = "drone server"
	app.Action = server
	app.Flags = flags
	app.Before = before

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
