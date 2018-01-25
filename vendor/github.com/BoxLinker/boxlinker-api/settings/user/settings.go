package user

import (
	"github.com/urfave/cli"
	log "github.com/Sirupsen/logrus"
)

var (
	ADMIN_NAME string
	ADMIN_PASSWORD string
	ADMIN_EMAIL string
	TOKEN_KEY string

	COOKIE_DOMAIN string

	USER_PASSWORD_SALT string

	CONFIRM_EMAIL_TOKEN_SECRET string

	// 验证邮件里的验证链接前缀
	VERIFY_EMAIL_URI string
)

func paramRequired(key,name string) {
	if name == "" {
		log.Fatalf("param '%s' required!", key)
	}
}


func InitSettings(c *cli.Context){
	TOKEN_KEY = c.String("token-key")

	ADMIN_NAME = c.String("admin-name")
	paramRequired("admin-name", ADMIN_NAME)

	ADMIN_PASSWORD = c.String("admin-password")
	paramRequired("admin-password", ADMIN_PASSWORD)

	ADMIN_EMAIL = c.String("admin-email")
	paramRequired("admin-email", ADMIN_EMAIL)

	USER_PASSWORD_SALT = c.String("user-password-salt")
	paramRequired("user-password-salt", USER_PASSWORD_SALT)

	COOKIE_DOMAIN = c.String("cookie-domain")

	CONFIRM_EMAIL_TOKEN_SECRET = c.String("confirm-email-token-secret")

	VERIFY_EMAIL_URI = c.String("verify-email-uri")
}

