package main

import (
	"github.com/urfave/cli"
	"github.com/BoxLinker/cicd/version"
	_ "github.com/joho/godotenv/autoload"
)

func main(){
	app := cli.NewApp()
	app.Name = "drone-agent"
	app.Version = version.Version.String()
	app.Usage = "drone agent"
	app.Flags = []cli.Flag{
		cli.IntFlag{
			EnvVar: "MAX_PROCS",
			Name:   "max-procs",
			Value:  1,
		},
		cli.StringFlag{
			EnvVar: "SERVER",
			Name: 	"server",
			Value: 	"localhost:9000",
		},
		cli.StringFlag{
			EnvVar: "USERNAME",
			Name:   "username",
			Usage:  "drone auth username",
			Value:  "x-oauth-basic",
		},
		cli.StringFlag{
			EnvVar: "PASSWORD,SECRET",
			Name:   "password",
			Usage:  "drone auth password",
		},
		cli.BoolTFlag{
			EnvVar: "DEBUG",
			Name:   "debug",
			Usage:  "start the agent in debug mode",
		},
	}
	app.Action = loop
}

