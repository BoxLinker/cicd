package server

import (
	"github.com/urfave/cli"
	"github.com/Sirupsen/logrus"
	cicdServer "github.com/BoxLinker/cicd/server"
	"os"
	"github.com/BoxLinker/cicd/manager"
	"golang.org/x/sync/errgroup"
	"net"
	"google.golang.org/grpc"
	"context"
	"google.golang.org/grpc/metadata"
	oldcontext "golang.org/x/net/context"
	"errors"
	"github.com/BoxLinker/cicd/store/datastore"
	"github.com/BoxLinker/cicd/pipeline/rpc/proto"
	"github.com/BoxLinker/cicd/logging"
	"github.com/BoxLinker/cicd/pubsub"
	"github.com/BoxLinker/cicd/models"
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
	cli.BoolFlag{
		EnvVar: "KUBERNETES_IN_CLUSTER",
		Name: "kubernetes-in-cluster",
		Usage: "whether connect to kubernetes by in cluster mode",
	},
	cli.StringFlag{
		EnvVar: "HOME_HOST",
		Name: "home-host",
		Usage: "boxlinker home page host",
	},
	cli.StringFlag{
		EnvVar: "DATABASE_DRIVER",
		Name: "database-driver",
	},
	cli.StringFlag{
		EnvVar: "DATABASE_DATASOURCE",
		Name: "database-datasource",
	},
	cli.StringFlag{
		EnvVar: "TOKEN_AUTH_URL",
		Name: "token-auth-url",
	},
	cli.BoolFlag{
		EnvVar: "GITHUB",
		Name: "github",
	},
	cli.StringFlag{
		EnvVar: "GITHUB_SERVER",
		Name: "github-server",
	},
	cli.StringFlag{
		EnvVar: "GITHUB_CLIENT",
		Name: "github-client",
	},
	cli.StringFlag{
		EnvVar: "GITHUB_SECRET",
		Name: "github-secret",
	},
	cli.StringSliceFlag{
		EnvVar: "GITHUB_SCOPE",
		Name: "github-scope",
	},
	cli.StringFlag{
		EnvVar: "AGENT_SECRET",
		Name: "agent-secret",
	},
}

func server(c *cli.Context) error {

	var (
		err error
		//dbEngine *xorm.Engine
		//clientSet *kubernetes.Clientset
	)

	if c.Bool("debug") {
		logrus.Infof("Debug enabled")
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.WarnLevel)
	}

	// connect to db
	//dbType := c.String("db-type")
	//switch dbType {
	//case "mysql":
	//	dbEngine, err = models.NewEngine(models.GetDBOptions(c), models.Tables())
	//	if err != nil {
	//		return err
	//	}
	//	break
	//default:
	//	return fmt.Errorf("unknow db type %s", dbType)
	//}
	//// connect to k8s api server
	//if c.Bool("kubernetes-in-cluster") {
	//	config, err := rest.InClusterConfig()
	//	if err != nil {
	//		return err
	//	}
	//	clientSet, err = kubernetes.NewForConfig(config)
	//	if err != nil {
	//		return fmt.Errorf("connect to incluster k8s error: %v", err)
	//	}
	//} else {
	//	var kubeconfig *string
	//	if home := homeDir(); home != "" {
	//		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	//	} else {
	//		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	//	}
	//	flag.Parse()
	//
	//	// use the current context in kubeconfig
	//	k8sConfig, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	//	if err != nil {
	//		panic(err.Error())
	//	}
	//	logrus.Infof("kubeconfig (%+v)", k8sConfig)
	//	// create the clientset
	//	clientSet, err = kubernetes.NewForConfig(k8sConfig)
	//	if err != nil {
	//		return fmt.Errorf("connect to k8s error: %v", err)
	//	}
	//}
	//controllerManager := new(manager.DefaultManager)
	//controllerManager.ClientSet = clientSet
	//controllerManager.DBEngine = dbEngine

	scmMap, err := SetupCodeBase(c)
	if err != nil {
		return err
	}

	dataStore := datastore.New(c.String("database-driver"), c.String("database-datasource"))

	logs := logging.New()
	queues := setupQueue(c, dataStore)
	pubsubs := pubsub.New()

	controllerManager, err := manager.New(&manager.Options{
		Store: dataStore,
		KubernetesInCluster: c.Bool("kubernetes-in-cluster"),
		SCMMap: scmMap,
		Logs: logs,
		Queue: queues,
		Pubsub: pubsubs,
	})

	if err != nil {
		return err
	}

	var g errgroup.Group

	g.Go(func ()error{
		cs := new(cicdServer.Server)
		//cs.CodeBase = cb
		cs.Manager = controllerManager
		cs.Listen = c.String("listen")
		cs.Config = cicdServer.Config{
			TokenAuthURL: c.String("token-auth-url"),
			HomeHost: c.String("home-host"),
		}

		return cs.Run()
	})

	g.Go(func() error {
		lis, err := net.Listen("tcp", ":9000")
		if err != nil {
			logrus.Error(err)
			return err
		}

		auther := &authorizer{
			password: c.String("agent-secret"),
		}

		s := grpc.NewServer(
			grpc.StreamInterceptor(auther.streamInterceptor),
			grpc.UnaryInterceptor(auther.unaryInterceptor),
		)

		ss := new(cicdServer.BoxCIServer)
		ss.Queue = queues
		ss.Logger = logs
		ss.Pubsub = pubsubs
		ss.Store = dataStore
		ss.SCM = scmMap[models.GITHUB] // todo scm 类型应该是根据请求参数来设置

		proto.RegisterBoxCIServer(s, ss)

		err = s.Serve(lis)
		if err != nil {
			logrus.Error(err)
			return err
		}
		return nil
	})

	return g.Wait()
}

type authorizer struct {
	username string
	password string
}

func (a *authorizer) streamInterceptor(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	if err := a.authorize(stream.Context()); err != nil {
		return err
	}
	return handler(srv, stream)
}

// func(ctx context.Context, req interface{}, info *UnaryServerInfo, handler UnaryHandler) (resp interface{}, err error)
func (a *authorizer) unaryInterceptor(ctx oldcontext.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	if err := a.authorize(ctx); err != nil {
		return nil, err
	}
	return handler(ctx, req)
}

func (a *authorizer) authorize(ctx context.Context) error {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if len(md["password"]) > 0 && md["password"][0] == a.password {
			return nil
		}
		return errors.New("invalid agent token")
	}
	return errors.New("missing agent token")
}

func before(c *cli.Context) error { return nil }

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}