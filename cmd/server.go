package main

import (
	"github.com/urfave/cli"
	"github.com/Sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"fmt"
	"k8s.io/client-go/rest"
	"flag"
	"path/filepath"
	"k8s.io/client-go/tools/clientcmd"
	cicdServer "github.com/BoxLinker/cicd/server"
	"os"
	"github.com/BoxLinker/cicd/models"
	"github.com/go-xorm/xorm"
)

var flags = []cli.Flag{
	cli.BoolFlag{
		EnvVar: "DEBUG",
		Name:   "debug",
		Usage:  "start the server in debug mode",
	},
	cli.StringFlag{
		EnvVar: "LISTEN",
		Name: "listen",
		Usage: "http listen adress",
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
	cli.BoolFlag{
		EnvVar: "KUBERNETES_IN_CLUSTER",
		Name: "kubernetes-in-cluster",
		Usage: "whether connect to kubernetes by in cluster mode",
	},
	cli.StringFlag{
		EnvVar: "DB_TYPE",
		Name: "db-type",
		Value: "mysql",
		Usage: "what db you used to connect",
	},
	cli.StringFlag{
		EnvVar: "DB_USER",
		Name: "db-user",
	},
	cli.StringFlag{
		EnvVar: "DB_PASSWORD",
		Name: "db-password",
	},
	cli.StringFlag{
		EnvVar: "DB_NAME",
		Name: "db-name",
	},
	cli.StringFlag{
		EnvVar: "DB_HOST",
		Name: "db-host",
	},
	cli.IntFlag{
		EnvVar: "DB_PORT",
		Name: "db-port",
	},
	cli.StringFlag{
		EnvVar: "TOKEN_AUTH_URL",
		Name: "token-auth-url",
	},
}

func server(c *cli.Context) error {

	var (
		err error
		dbEngine *xorm.Engine
		clientSet *kubernetes.Clientset
	)

	if c.Bool("debug") {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.WarnLevel)
	}

	// connect to db
	dbType := c.String("db-type")
	switch dbType {
	case "mysql":
		dbEngine, err = models.NewEngine(models.GetDBOptions(c), models.Tables())
		if err != nil {
			return err
		}
		break
	default:
		return fmt.Errorf("unknow db type %s", dbType)
	}
	// connect to k8s api server
	if c.Bool("kubernetes-in-cluster") {
		config, err := rest.InClusterConfig()
		if err != nil {
			return err
		}
		clientSet, err = kubernetes.NewForConfig(config)
		if err != nil {
			return fmt.Errorf("connect to incluster k8s error: %v", err)
		}
	} else {
		var kubeconfig *string
		if home := homeDir(); home != "" {
			kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
		} else {
			kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
		}
		flag.Parse()

		// use the current context in kubeconfig
		k8sConfig, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
		if err != nil {
			panic(err.Error())
		}
		logrus.Infof("kubeconfig (%+v)", k8sConfig)
		// create the clientset
		clientSet, err = kubernetes.NewForConfig(k8sConfig)
		if err != nil {
			return fmt.Errorf("connect to k8s error: %v", err)
		}
	}

	// setup codebase
	cb, err := SetupCodeBase(c)
	if err != nil {
		return err
	}

	cs := new(cicdServer.Server)
	cs.CodeBase = cb
	cs.ClientSet = clientSet
	cs.DBEngine = dbEngine
	cs.Listen = c.String("listen")

	return cs.Run(c)
}

func before(c *cli.Context) error { return nil }

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}