package main

import (
	"github.com/urfave/cli"
	log "github.com/Sirupsen/logrus"
	"os"
	api "github.com/BoxLinker/boxlinker-api/api/v1/email"
	"github.com/BoxLinker/boxlinker-api/cmd"
)

var flags = []cli.Flag{
	cli.BoolFlag{
		Name: "test",
		EnvVar: "TEST",
	},
	cli.StringFlag{
		Name: "mail-host",
		Value: "smtp.exmail.qq.com:25",
		EnvVar: "MAIL_HOST",
	},
	cli.StringFlag{
		Name: "mail-user",
		Value: "service@boxlinker.com",
		EnvVar: "MAIL_USER",
	},
	cli.StringFlag{
		Name: "mail-user-title",
		Value: "Boxlinker",
		EnvVar: "MAIL_USER_TITLE",
	},
	cli.StringFlag{
		Name: "mail-password",
		Value: "Just4fun",
		EnvVar: "MAIL_PASSWORD",
	},
	cli.StringFlag{
		Name: "mail-type",
		Value: "html",
		EnvVar: "MAIL_TYPE",
	},




}

func main(){
	app := cli.NewApp()
	app.Name = "Boxlinker Email server"
	app.Before = func(c *cli.Context) error {
		if c.Bool("debug") {
			log.SetLevel(log.DebugLevel)
		}
		return nil
	}
	app.Action = action
	app.Flags = append(flags, append(cmd.AMQPFlags, cmd.SharedFlags...)...)

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func action(c *cli.Context) error {

	//notifyMsg := make(chan []byte)

	//amqpConsumer := &amqp.Consumer{
	//	URI: c.String("rabbitmq-uri"),
	//	Exchange: c.String("rabbitmq-exchange"),
	//	ExchangeType: c.String("rabbitmq-exchange-type"),
	//	QueueName: c.String("rabbitmq-queue-name"),
	//	Tag: c.String("rabbitmq-consumer-tag"),
	//	BindingKey: c.String("rabbitmq-binding-key"),
	//	NotifyMsg: notifyMsg,
	//}

	//if err := amqpConsumer.Run(); err != nil {
	//	return err
	//}


	return api.NewApi(api.ApiOptions{
		Listen: c.String("listen"),
		//AMQPConsumer: amqpConsumer,
		EmailOption: api.EmailOption{
			User: c.String("mail-user"),
			UserTitle: c.String("mail-user-title"),
			Host: c.String("mail-host"),
			Password: c.String("mail-password"),
		},
		TestMode: c.Bool("test"),
	}).Run()


}


