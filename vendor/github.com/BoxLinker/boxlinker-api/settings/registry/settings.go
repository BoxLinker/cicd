package registry

import "github.com/urfave/cli"

var (
	BASIC_AUTH_URL string
)

func InitSettings(c *cli.Context){
	BASIC_AUTH_URL = c.String("basic-auth-url")
}
