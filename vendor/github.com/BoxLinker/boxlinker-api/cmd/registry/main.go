package main

import (
	"github.com/urfave/cli"
	"github.com/Sirupsen/logrus"
	"os"
	cmd "github.com/BoxLinker/boxlinker-api/cmd"
	registryModels "github.com/BoxLinker/boxlinker-api/controller/models/registry"
	api "github.com/BoxLinker/boxlinker-api/api/v1/registry"
	"fmt"
	"github.com/BoxLinker/boxlinker-api/controller/models"
	"github.com/BoxLinker/boxlinker-api/controller/manager"
	"errors"
)

var flags = []cli.Flag{
	cli.StringFlag{
		Name: "basic-auth-url",
		Value: "http://localhost:8080/v1/user/auth/basicAuth",
		EnvVar: "BASIC_AUTH_URL",
	},

	cli.StringFlag{
		Name: "config-file",
		Value: "./auth_config.yml",
		EnvVar: "CONFIG_FILE",
	},
}

func main(){
	app := cli.NewApp()
	app.Name = "Boxlinker Registry server"
	app.Before = func(c *cli.Context) error {
		if c.Bool("debug") {
			logrus.SetLevel(logrus.DebugLevel)
		}
		return nil
	}
	app.Action = action
	app.Flags = append(flags, append(cmd.DBFlags, cmd.SharedFlags...)...)

	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}


func action(c *cli.Context) error {

	configFilePath := c.String("config-file")
	if len(configFilePath) == 0 {
		return errors.New("no config file provided")
	}

	config, err := api.LoadConfig(configFilePath)
	if err != nil {
		return err
	}

	if config.Server.Debug {
		logrus.SetLevel(logrus.DebugLevel)
	}
	basicAuthURL := c.String("basic-auth-url")
	if len(basicAuthURL) == 0 {
		return errors.New("basic-auth-url is required")
	}

	engine, err := models.NewEngine(models.DBOptions{
		User: config.DB.User,
		Password: config.DB.Password,
		Name: config.DB.Name,
		Host: config.DB.Host,
		Port: config.DB.Port,
	}, registryModels.Tables())
	if err != nil {
		return fmt.Errorf("new db engine err: %v", err)
	}

	controllerManager, err := manager.NewRegistryManager(engine)
	if err != nil {
		return fmt.Errorf("new controller manager err: %v", err)
	}


	a, err := api.NewApi(&api.ApiConfig{
		Listen: c.String("listen"),
		Manager: controllerManager,
		ConfigFilePath: configFilePath,
		BasicAuthURL: basicAuthURL,
		Config: config,
	})
	if err != nil {
		return err
	}

	return fmt.Errorf("run api err: %v", a.Run())
}