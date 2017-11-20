package main

import (
	"github.com/urfave/cli"
	"github.com/Sirupsen/logrus"
	"github.com/cabernety/gopkg/amqp"
)

var flags = []cli.Flag{
	cli.BoolFlag{
		EnvVar: "DEBUG",
		Name:   "debug",
		Usage:  "start the server in debug mode",
	},
	cli.StringFlag{
		EnvVar: "RABBITMQ_HOST",
		Name: "rabbitmq-host",
		Usage: "the host connect to rabbitmq",
	},
	cli.StringFlag{
		EnvVar: "RABBITMQ_EXCHANGE",
		Name: "rabbitmq-exchange",
		Usage: "the rabbitmq exchange connect to rabbitmq",
	},
	cli.StringFlag{
		EnvVar: "RABBITMQ_EXCHANGE_TYPE",
		Name: "rabbitmq-exchange-type",
		Usage: "the rabbitmq exchange type connect to rabbitmq",
	},
	cli.StringFlag{
		EnvVar: "RABBITMQ_QUEUE_NAME",
		Name: "rabbitmq-queue-name",
		Usage: "the rabbitmq queue name connect to rabbitmq",
	},
	cli.StringFlag{
		EnvVar: "RABBITMQ_BINDING_KEY",
		Name: "rabbitmq-binding-key",
		Usage: "the rabbitmq binding key connect to rabbitmq",
	},

}

func server(c *cli.Context) error {
	if c.Bool("debug") {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.WarnLevel)
	}

	mqMsgCh := make(chan []byte)
	// run mq consumer
	mqConsumer, err := amqp.NewConsumer(&amqp.ConsumerConfig{
		URI: c.String("rabbitmq-host"),
		Exchange: c.String("rabbitmq-exchange"),
		ExchangeType: c.String("rabbitmq-exchange-type"),
		QueueName: c.String("rabbitmq-queue-name"),
		BindingKey: c.String("rabbitmq-binding-key"),
		NotifyMsg: mqMsgCh,
	})
	if err != nil {
		return err
	}
	go mqConsumer.Run()



	return nil
}

func before(c *cli.Context) error { return nil }
