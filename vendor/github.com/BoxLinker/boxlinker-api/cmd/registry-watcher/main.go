package main

import (
	"github.com/urfave/cli"
	"github.com/Sirupsen/logrus"
	api "github.com/BoxLinker/boxlinker-api/api/v1/registry-watcher"
	"os"
	"errors"
	"github.com/BoxLinker/boxlinker-api/controller/manager"
	"github.com/BoxLinker/boxlinker-api/pkg/amqp"
)

var flags = []cli.Flag{
	cli.StringFlag{
		Name: "config-file",
		Value: "./config.yml",
		EnvVar: "CONFIG_FILE",
	},
}

func main(){
	app := cli.NewApp()
	app.Name = "Boxlinker 镜像库监听服务"
	app.Before = func(c *cli.Context) error {
		if c.Bool("debug") {
			logrus.SetLevel(logrus.DebugLevel)
		}
		return nil
	}
	app.Action = action
	app.Flags = flags

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

	amqpProducer := amqp.NewProducer(amqp.ProducerOptions{
		URI: config.Amqp.Host,
		Exchange: config.Amqp.Exchange,
		ExchangeType: config.Amqp.ExchangeType,
		Reliable: config.Amqp.Reliable,
	})

	controllerManager := manager.NewDefaultRegistryWatcherManagerOptions(manager.DefaultRegistryWatcherManagerOptions{
		AmqpProducer: amqpProducer,
	})

	aApi, err := api.NewApi(api.ApiConfig{
		Config: config,
		ControllerManager: controllerManager,
	})

	if err != nil {
		return err
	}

	return aApi.Run()

}