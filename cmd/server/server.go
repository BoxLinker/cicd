package main

import (
	"context"
	"errors"
	"net"
	"os"

	"github.com/BoxLinker/cicd/logging"
	"github.com/BoxLinker/cicd/manager"
	"github.com/BoxLinker/cicd/models"
	"github.com/BoxLinker/cicd/pipeline/rpc/proto"
	"github.com/BoxLinker/cicd/pubsub"
	cicdServer "github.com/BoxLinker/cicd/server"
	"github.com/BoxLinker/cicd/store/datastore"
	"github.com/Sirupsen/logrus"
	"github.com/urfave/cli"
	oldcontext "golang.org/x/net/context"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var flags = []cli.Flag{
	cli.BoolFlag{
		EnvVar: "DEBUG",
		Name:   "debug",
		Usage:  "start the server in debug mode",
	},
	cli.StringFlag{
		EnvVar: "LISTEN",
		Name:   "listen",
		Usage:  "http listen address",
	},
	cli.StringFlag{
		EnvVar: "TCP_LISTEN",
		Name:   "tcp-listen",
		Usage:  "tcp listen address",
		Value:  ":9000",
	},
	cli.BoolFlag{
		EnvVar: "KUBERNETES_IN_CLUSTER",
		Name:   "kubernetes-in-cluster",
		Usage:  "whether connect to kubernetes by in cluster mode",
	},
	cli.StringFlag{
		EnvVar: "HOME_HOST",
		Name:   "home-host",
		Usage:  "boxlinker home page host",
	},
	cli.StringFlag{
		EnvVar: "DATABASE_DRIVER",
		Name:   "database-driver",
	},
	cli.StringFlag{
		EnvVar: "DATABASE_DATASOURCE",
		Name:   "database-datasource",
	},
	cli.StringFlag{
		EnvVar: "TOKEN_AUTH_URL",
		Name:   "token-auth-url",
	},
	cli.BoolFlag{
		EnvVar: "GITHUB",
		Name:   "github",
	},
	cli.StringFlag{
		EnvVar: "GITHUB_SERVER",
		Name:   "github-server",
	},
	cli.StringFlag{
		EnvVar: "GITHUB_CLIENT",
		Name:   "github-client",
	},
	cli.StringFlag{
		EnvVar: "GITHUB_SECRET",
		Name:   "github-secret",
	},
	cli.StringSliceFlag{
		EnvVar: "GITHUB_SCOPE",
		Name:   "github-scope",
	},
	cli.StringFlag{
		EnvVar: "AGENT_SECRET",
		Name:   "agent-secret",
	},
	cli.StringFlag{
		EnvVar: "REPO_CONFIG",
		Name:   "repo-config",
		Usage:  "file path for the drone config",
		Value:  ".boxci.yml",
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

	scmMap, err := SetupCodeBase(c)
	if err != nil {
		return err
	}

	dataStore := datastore.New(c.String("database-driver"), c.String("database-datasource"))

	logs := logging.New()
	queues := setupQueue(c, dataStore)
	pubsubs := pubsub.New()

	controllerManager, err := manager.New(&manager.Options{
		Store:               dataStore,
		KubernetesInCluster: c.Bool("kubernetes-in-cluster"),
		SCMMap:              scmMap,
		Logs:                logs,
		Queue:               queues,
		Pubsub:              pubsubs,
	})

	if err != nil {
		return err
	}

	var g errgroup.Group

	g.Go(func() error {
		cs := new(cicdServer.Server)
		//cs.CodeBase = cb
		cs.Manager = controllerManager
		cs.Listen = c.String("listen")
		cs.Config = cicdServer.Config{
			TokenAuthURL: c.String("token-auth-url"),
			HomeHost:     c.String("home-host"),
			RepoConfig:   c.String("repo-config"),
		}

		return cs.Run()
	})

	g.Go(func() error {
		lis, err := net.Listen("tcp", c.String("tcp-listen"))
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

		logrus.Infof("grpc server listen on: %s", c.String("tcp-listen"))
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
