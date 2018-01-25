package cmd

import "github.com/urfave/cli"

var SharedFlags = []cli.Flag{
	cli.BoolFlag{
		Name:   "debug, D",
		Usage:  "enable debug",
		EnvVar: "DEBUG",
	},
	cli.StringFlag{
		Name:   "listen, l",
		Value:  ":8080",
		Usage:  "server listen address",
		EnvVar: "LISTEN",
	},
}

var DBFlags = []cli.Flag{
	cli.StringFlag{
		Name:   "db-user",
		Value:  "root",
		EnvVar: "DB_USER",
	},
	cli.StringFlag{
		Name:   "db-password",
		Value:  "123456",
		EnvVar: "DB_PASSWORD",
	},
	cli.StringFlag{
		Name:   "db-host",
		Value:  "127.0.0.1",
		EnvVar: "DB_HOST",
	},
	cli.StringFlag{
		Name:   "db-port",
		Value:  "3306",
		EnvVar: "DB_PORT",
	},
	cli.StringFlag{
		Name:   "db-name",
		Value:  "boxlinker",
		EnvVar: "DB_NAME",
	},
}

var AMQPFlags = []cli.Flag{
	cli.StringFlag{
		Name: "rabbitmq-uri",
		Value: "amqp://guest:guest@localhost:5672/",
		EnvVar: "RABBITMQ_URI",
	},
	cli.StringFlag{
		Name: "rabbitmq-exchange",
		Value: "test-exchange",
		EnvVar: "RABBITMQ_EXCHANGE",
	},
	cli.StringFlag{
		Name: "rabbitmq-exchange-type",
		Usage: "Exchange type - direct|fanout|topic|x-custom",
		Value: "fanout",
		EnvVar: "RABBITMQ_EXCHANGE_TYPE",
	},
	cli.StringFlag{
		Name: "rabbitmq-queue-name",
		Value: "test-queue-name",
		EnvVar: "RABBITMQ_QUEUE_NAME",
	},
	cli.StringFlag{
		Name: "rabbitmq-consumer-tag",
		Usage: "AMQP consumer tag (should not be blank)",
		Value: "boxlinker-email",
		EnvVar: "RABBITMQ_CONSUMER_TAG",
	},
	cli.StringFlag{
		Name: "rabbitmq-binding-key",
		Value: "boxlinker-email-amqp-binding-key",
		EnvVar: "RABBITMQ_BINDING_KEY",
	},
}