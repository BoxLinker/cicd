package main

import (
	"github.com/urfave/cli"
	"github.com/BoxLinker/cicd/pipeline/rpc"
	"os"
	"github.com/Sirupsen/logrus"
	"google.golang.org/grpc"
	oldcontext "golang.org/x/net/context"
	"github.com/tevino/abool"
	"google.golang.org/grpc/metadata"
	"context"
	"github.com/BoxLinker/cicd/signal"
	"sync"
)

func loop(c *cli.Context) error {
	filter := rpc.Filter{
		Labels: map[string]string{
			"platform": c.String("platform"),
		},
		Expr: c.String("boxci-filter"),
	}

	hostname := c.String("hostname")
	if len(hostname) == 0 {
		hostname, _ = os.Hostname()
	}

	if c.BoolT("debug") {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.WarnLevel)
	}


	conn, err := grpc.Dial(
		c.String("server"),
		grpc.WithInsecure(),
		grpc.WithPerRPCCredentials(&credentials{
			username: c.String("username"),
			password: c.String("password"),
		}),
	)

	if err != nil {
		return err
	}

	defer conn.Close()

	client := rpc.NewGrpcClient(conn)

	sigterm := abool.New()

	ctx := metadata.NewOutgoingContext(
		context.Background(),
		metadata.Pairs("hostname", hostname),
	)
	ctx = signal.WithContextFunc(ctx, func(){
		println("ctrl+c received, terminating process")
		sigterm.Set()
	})

	var wg sync.WaitGroup
	parallel := c.Int("max-procs")
	wg.Add(parallel)

	for i := 0; i < parallel; i++ {
		go func(){
			defer wg.Done()
			for {
				if sigterm.IsSet() {
					return
				}
				r := runner{
					client: client,
					filter: filter,
					hostname: hostname,
				}
				if err := r.run(ctx); err != nil {
					logrus.Errorf("pipeline done with error")
					return
				}
			}
		}()
	}

	wg.Wait()
	return nil
}

// NOTE we need to limit the size of the logs and files that we upload.
// The maximum grpc payload size is 4194304. So until we implement streaming
// for uploads, we need to set these limits below the maximum.
const (
	maxLogsUpload = 2000000 // this is per step
	maxFileUpload = 1000000
)

type runner struct {
	client   rpc.Peer
	filter   rpc.Filter
	hostname string
}

func (r *runner) run(ctx context.Context) error {
	return nil
}


type credentials struct {
	username string
	password string
}

func (c *credentials) GetRequestMetadata(oldcontext.Context, ...string) (map[string]string, error) {
	return map[string]string{
		"username": c.username,
		"password": c.password,
	}, nil
}

func (c *credentials) RequireTransportSecurity() bool {
	return false
}
