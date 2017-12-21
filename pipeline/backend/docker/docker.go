package docker

import (
	"github.com/docker/docker/client"
	"github.com/docker/docker/api/types"
	"github.com/BoxLinker/cicd/pipeline/backend"
	"io"
	"context"
	"github.com/docker/docker/api/types/volume"
	"io/ioutil"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/rs/zerolog/log"
)

type engine struct {
	client client.APIClient
}

func New(cli client.APIClient) backend.Engine {
	return &engine{
		client: cli,
	}
}

func NewEnv() (backend.Engine, error) {
	cli, err := client.NewEnvClient()
	if err != nil {
		return nil, err
	}
	return New(cli), nil
}

func (e *engine) Setup(conf *backend.Config) error {
	for _, vol := range conf.Volumes {
		_, err := e.client.VolumeCreate(noContext, volume.VolumesCreateBody{
			Name: 	vol.Name,
			Driver: vol.Driver,
			DriverOpts: vol.DriverOpts,
		})
		if err != nil {
			return err
		}
	}
	for _, network := range conf.Networks {
		_, err := e.client.NetworkCreate(noContext, network.Name, types.NetworkCreate{
			Driver: network.Driver,
			Options: network.DriverOpts,
		})
		if err != nil {
			return err
		}
	}
	return nil
}
func (e *engine) Exec(proc *backend.Step) error {
	ctx := context.Background()

	config := toConfig(proc)
	hostConfig := toHostConfig(proc)

	// image pull 操作 可能会需要授权(username:password)
	pullopts := types.ImagePullOptions{}
	if proc.AuthConfig.Username != "" && proc.AuthConfig.Password != "" {
		pullopts.RegistryAuth, _ = encodeAuthToBase64(proc.AuthConfig)
	}

	//
	if proc.Pull {
		rc, perr := e.client.ImagePull(ctx, config.Image, pullopts)
		if perr == nil {
			io.Copy(ioutil.Discard, rc)
			rc.Close()
		}

		// fix for drone/drone#1917
		if perr != nil && proc.AuthConfig.Password != "" {
			return perr
		}
	}

	_, err := e.client.ContainerCreate(ctx, config, hostConfig, nil, proc.Name)
	if client.IsErrNotFound(err) {
		rc, perr := e.client.ImagePull(ctx, config.Image, pullopts)
		if perr != nil {
			return perr
		}
		io.Copy(ioutil.Discard, rc)
		rc.Close()

		_, err = e.client.ContainerCreate(ctx, config, hostConfig, nil, proc.Name)
	}
	if err != nil {
		return err
	}

	if len(proc.NetworkMode) == 0 {
		for _, net := range proc.Networks {
			err = e.client.NetworkConnect(ctx, net.Name, proc.Name, &network.EndpointSettings{
				Aliases: net.Aliases,
			})
			if err != nil {
				return err
			}
		}
	}

	return e.client.ContainerStart(ctx, proc.Name, startOpts)
}

func (e *engine) Kill(proc *backend.Step) error {
	return e.client.ContainerKill(noContext, proc.Name, "9")
}

func (e *engine) Wait(proc *backend.Step) (*backend.State, error) {
	wait, err := e.client.ContainerWait(noContext, proc.Name, container.WaitConditionNextExit)
	if err != nil {
		// todo
	}
DONE:
	for {
		select {
		case werr := <-err:
			log.Error().Msgf("=> ContainerWait error: %s", werr)
			return nil, werr
		case body := <-wait:
			log.Debug().Msgf("==> ContainerWait result (%+v)", body)
			break DONE
		}
	}

	info, ierr := e.client.ContainerInspect(noContext, proc.Name)
	if ierr != nil {
		return nil, ierr
	}
	log.Debug().Msgf("===> ContainerInspect: (%+v)", info.State)
	if info.State.Running {
		// todo
	}

	return &backend.State{
		Exited: true,
		ExitCode: info.State.ExitCode,
		OOMKilled: info.State.OOMKilled,
	}, nil
}

func (e *engine) Tail(proc *backend.Step) (io.ReadCloser, error) {
	logs, err := e.client.ContainerLogs(noContext, proc.Name, logsOpts)
	if err != nil {
		return nil, err
	}
	rc, wc := io.Pipe()

	go func(){
		stdcopy.StdCopy(wc, wc, logs)
		logs.Close()
		wc.Close()
		rc.Close()
	}()

	return rc, nil
}

func (e *engine) Destroy(conf *backend.Config) error {
	for _, stage := range conf.Stages {
		for _, step := range stage.Steps {
			e.client.ContainerKill(noContext, step.Name, "9")
			e.client.ContainerRemove(noContext, step.Name, removeOpts)
		}
	}
	for _, vol := range conf.Volumes {
		e.client.VolumeRemove(noContext, vol.Name, true)
	}

	for _, net := range conf.Networks {
		e.client.NetworkRemove(noContext, net.Name)
	}
	return nil
}

var (
	noContext = context.Background()

	startOpts = types.ContainerStartOptions{}

	removeOpts = types.ContainerRemoveOptions{
		RemoveVolumes: true,
		RemoveLinks:   false,
		Force:         false,
	}

	logsOpts = types.ContainerLogsOptions{
		Follow:     true,
		ShowStdout: true,
		ShowStderr: true,
		Details:    false,
		Timestamps: false,
	}
)