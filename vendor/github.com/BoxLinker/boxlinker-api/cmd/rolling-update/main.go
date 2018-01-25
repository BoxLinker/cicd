package main

import (
	"github.com/urfave/cli"
	"github.com/Sirupsen/logrus"
	"os"
	"fmt"
	api "github.com/BoxLinker/boxlinker-api/api/v1/rolling-update"
	"github.com/BoxLinker/boxlinker-api/controller/manager"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"flag"
	"path/filepath"
	"k8s.io/client-go/tools/clientcmd"
	"github.com/BoxLinker/boxlinker-api/controller/models"
	"errors"
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
	app.Name = "Boxlinker 滚动更新服务"
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

	var (
		clientSet *kubernetes.Clientset
	)

	if config.InCluster {
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

	info, err := clientSet.ServerVersion()
	if err != nil {
		return err
	}
	logrus.Infof("connect to api-server (%+v)", info)
	// KUBE_CGROUP_DRIVER="--cgroup-driver=systemd"
	// KUBE_STORAGE_DRIVER="--storage-driver=etcd2"
	engine, err := models.NewEngine(models.DBOptions{
		User: config.DB.User,
		Password: config.DB.Password,
		Name: config.DB.Name,
		Host: config.DB.Host,
		Port: config.DB.Port,
	}, models.RollingUpdateTables())
	if err != nil {
		return fmt.Errorf("new db engine err: %v", err)
	}

	ac := config.AMQP
	controllerManager, err := manager.NewDefaultRollingUpdateManager(&manager.DefaultRollingUpdateConfig{
		RegistryHost: config.RegistryHost,
	},engine, clientSet, &amqp.ConsumerConfig{
		URI: ac.URI,
		ExchangeType: ac.ExchangeType,
		Exchange: ac.Exchange,
		QueueName: ac.QueueName,
		BindingKey: ac.BindingKey,
	})
	if err != nil {
		return fmt.Errorf("new controller manager err: %v", err)
	}

	appApi, err := api.NewApi(api.ApiConfig{
		Config: config,
		ControllerManager: controllerManager,
		ClientSet: clientSet,
	})

	if err != nil {
		return err
	}
	if err := appApi.Run(); err != nil {
		return err
	}
	return nil
}


func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}