package main

import (
	"fmt"
	"os"

	"github.com/BoxLinker/cicd/version"
	"github.com/Sirupsen/logrus"
	_ "github.com/joho/godotenv/autoload"
	"github.com/urfave/cli"
)

func main() {
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
			Name:   "server",
			Value:  "localhost:9000",
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
		cli.StringFlag{
			EnvVar: "PLATFORM",
			Name:   "platform",
			Value:  "linux/amd64",
		},
	}
	app.Action = loop
	app.Before = func(c *cli.Context) error {
		if c.Bool("debug") {
			logrus.SetLevel(logrus.DebugLevel)
		} else {
			logrus.SetLevel(logrus.InfoLevel)
		}
		return nil
	}
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
