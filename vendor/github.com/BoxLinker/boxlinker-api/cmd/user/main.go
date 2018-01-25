package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/urfave/cli"
	"github.com/BoxLinker/boxlinker-api/controller/manager"
	"github.com/BoxLinker/boxlinker-api/auth/builtin"
	api "github.com/BoxLinker/boxlinker-api/api/v1/user"

	"os"
	settings "github.com/BoxLinker/boxlinker-api/settings/user"
	"fmt"
	"github.com/BoxLinker/boxlinker-api/cmd"
	userModels "github.com/BoxLinker/boxlinker-api/controller/models/user"
	"github.com/BoxLinker/boxlinker-api/controller/models"
)

var (
	flags = []cli.Flag{
		cli.StringFlag{
			Name: "confirm-email-token-secret",
			Value:	"arandomconfirmemailtokensecret",
			EnvVar: "CONFIRM_EMAIL_TOKEN_SECRET",
		},
		cli.StringFlag{
			Name: "send-email-uri",
			Value: "http://localhost:8081/v1/email/send",
			EnvVar: "SEND_EMAIL_URI",
		},
		cli.StringFlag{
			Name: "verify-email-uri",
			Value: "http://localhost:8080/v1/user/auth/confirm_email",
			EnvVar: "VERIFY_EMAIL_URI",
		},

		cli.StringFlag{
			Name:   "admin-name",
			Value:  "admin",
			EnvVar: "ADMIN_NAME",
		},
		cli.StringFlag{
			Name:   "admin-password",
			Value:  "Admin123456",
			EnvVar: "ADMIN_PASSWORD",
		},
		cli.StringFlag{
			Name:   "admin-email",
			Value:  "service@boxlinker.com",
			EnvVar: "ADMIN_EMAIL",
		},
		cli.StringFlag{
			Name: 	"user-password-salt",
			Value:	"arandomuserpasswordsalt",
			EnvVar: "USER_PASSWORD_SALT",
		},
		cli.StringFlag{
			Name:  "cookie-domain",
			Value: "localhost",
			EnvVar: "COOKIE_DOMAIN",
		},


	}
)

func main() {
	app := cli.NewApp()
	app.Name = "Boxlinker 用户服务"
	app.Usage = "Boxlinker 用户服务"
	app.Action = action
	app.Before = func(c *cli.Context) error {
		log.SetLevel(log.DebugLevel)
		if c.Bool("debug") {
			log.SetLevel(log.DebugLevel)
		}
		return nil
	}
	app.Flags = append(flags, append(cmd.DBFlags, cmd.SharedFlags...)...)

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func action(c *cli.Context) error {

	settings.InitSettings(c)


	authenticator := builtin.NewAuthenticator()

	//controllerManager, err := manager.NewManager(manager.ManagerOptions{
	//	Authenticator:	authenticator,
	//	DBUser: 		c.String("db-user"),
	//	DBPassword: 	c.String("db-password"),
	//	DBHost: 		c.String("db-host"),
	//	DBPort: 		c.Int("db-port"),
	//	DBName: 		c.String("db-name"),
	//})
	engine, err := models.NewEngine(models.GetDBOptions(c), userModels.Tables())
	if err != nil {
		return fmt.Errorf("new db engine err: %v", err)
	}
	controllerManager, err := manager.NewUserManager(engine, authenticator)

	if err != nil {
		return fmt.Errorf("New Manager: %s", err.Error())
	}

	if err := controllerManager.CheckAdminUser(); err != nil {
		return fmt.Errorf("CheckAdminUser: %v", err)
	}

	return api.NewApi(api.ApiOptions{
		Listen: c.String("listen"),
		Manager: controllerManager,
		SendEmailUri: c.String("send-email-uri"),
	}).Run()

}
