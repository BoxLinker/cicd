package compiler

import (
	"fmt"

	"github.com/BoxLinker/cicd/pipeline/backend"
	"github.com/BoxLinker/cicd/pipeline/frontend"
	"github.com/BoxLinker/cicd/pipeline/frontend/yaml"
)

type Secret struct {
	Name  string
	Value string
	Match []string
}

type ResourceLimit struct {
	MemSwapLimit int64
	MemLimit     int64
	ShmSize      int64
	CPUQuota     int64
	CPUShares    int64
	CPUSet       string
}

type Registry struct {
	Hostname string
	Username string
	Password string
	Email    string
	Token    string
}

// Compiler compiles the yaml
type Compiler struct {
	local      bool
	escalated  []string
	prefix     string
	volumes    []string
	networks   []string
	env        map[string]string
	base       string
	path       string
	metadata   frontend.Metadata
	registries []Registry
	secrets    map[string]Secret
	cacher     Cacher
	reslimit   ResourceLimit
}

// New creates a new Compiler with options.
func New(opts ...Option) *Compiler {
	compiler := &Compiler{
		env:     map[string]string{},
		secrets: map[string]Secret{},
	}
	for _, opt := range opts {
		opt(compiler)
	}
	return compiler
}

func (c *Compiler) Compile(conf *yaml.Config) *backend.Config {
	config := new(backend.Config)

	// create a default volume
	config.Volumes = append(config.Volumes, &backend.Volume{
		Name:   fmt.Sprintf("%s_default", c.prefix),
		Driver: "local",
	})

	// create a default network
	config.Networks = append(config.Networks, &backend.Network{
		Name:   fmt.Sprintf("%s_default", c.prefix),
		Driver: "bridge",
	})

	//
	if len(conf.Workspace.Base) != 0 {
		c.base = conf.Workspace.Base
	}
	if len(conf.Workspace.Path) != 0 {
		c.path = conf.Workspace.Path
	}

	// add default clone step
	if c.local == false && len(conf.Clone.Containers) == 0 {
		container := &yaml.Container{
			Name:  "clone",
			Image: "index.boxlinker.com/boxlinker/cicd-plugins-git:latest",
			Vargs: map[string]interface{}{"depth": "0"},
			AuthConfig: yaml.AuthConfig{
				Username: "boxlinker",
				Password: "QAZwsx123",
			},
		}
		switch c.metadata.Sys.Arch {
		case "linux/arm":
			container.Image = "plugins/git:linux-arm"
		case "linux/arm64":
			container.Image = "plugins/git:linux-arm64"
		}
		name := fmt.Sprintf("%s_clone", c.prefix)
		step := c.createProcess(name, container, "clone")

		stage := new(backend.Stage)
		stage.Name = name
		stage.Alias = "clone"
		stage.Steps = append(stage.Steps, step)

		config.Stages = append(config.Stages, stage)
	} else if c.local == false {
		for i, container := range conf.Clone.Containers {
			if !container.Constraints.Match(c.metadata) {
				continue
			}
			stage := new(backend.Stage)
			stage.Name = fmt.Sprintf("%s_clone_%v", c.prefix, i)
			stage.Alias = container.Name

			name := fmt.Sprintf("%s_clone_%d", c.prefix, i)
			step := c.createProcess(name, container, "clone")
			stage.Steps = append(stage.Steps, step)

			config.Stages = append(config.Stages, stage)
		}
	}

	c.setupCache(conf, config)

	// add services steps
	if len(conf.Services.Containers) != 0 {
		stage := new(backend.Stage)
		stage.Name = fmt.Sprintf("%s_services", c.prefix)
		stage.Alias = "services"

		for i, container := range conf.Services.Containers {
			if !container.Constraints.Match(c.metadata) {
				continue
			}

			name := fmt.Sprintf("%s_services_%d", c.prefix, i)
			step := c.createProcess(name, container, "services")
			stage.Steps = append(stage.Steps, step)
		}
		config.Stages = append(config.Stages, stage)
	}

	// add pipeline steps. 1 pipeline step per stage, at the moment
	var stage *backend.Stage
	var group string
	for i, container := range conf.Pipeline.Containers {
		// skip if local and should not run local
		if c.local && !container.Constraints.Local.Bool() {
			continue
		}

		if !container.Constraints.Match(c.metadata) {
			continue
		}

		if stage == nil || group != container.Group || container.Group == "" {
			group = container.Group

			stage = new(backend.Stage)
			stage.Name = fmt.Sprintf("%s_stage_%v", c.prefix, i)
			stage.Alias = container.Name
			config.Stages = append(config.Stages, stage)
		}

		name := fmt.Sprintf("%s_step_%d", c.prefix, i)
		step := c.createProcess(name, container, "pipeline")
		stage.Steps = append(stage.Steps, step)
	}

	// publish := conf.Publish
	// if publish {
	// 	container := &yaml.Container{
	// 		Name:  "publish",
	// 		Image: "index.boxlinker.com/boxlinker/cicd-plugins-docker:latest",
	// 		Vargs: map[string]interface{}{
	// 			"repo":     fmt.Sprintf("index.boxlinker.com/%s", strings.ToLower(c.metadata.Repo.Name)),
	// 			"auto_tag": true,
	// 		},
	// 		Privileged: true,
	// 		AuthConfig: yaml.AuthConfig{
	// 			Username: "boxlinker",
	// 			Password: "QAZwsx123",
	// 		},
	// 	}
	// 	stagePublish := new(backend.Stage)
	// 	stagePublish.Name = fmt.Sprintf("%s_stage_publish", c.prefix)
	// 	stagePublish.Alias = container.Name
	// 	config.Stages = append(config.Stages, stagePublish)

	// 	name := fmt.Sprintf("%s_step_publish", c.prefix)
	// 	step := c.createProcess(name, container, "pipeline")
	// 	stagePublish.Steps = append(stagePublish.Steps, step)
	// }

	c.setupCacheRebuild(conf, config)

	return config

}

func (c *Compiler) setupCache(conf *yaml.Config, ir *backend.Config) {
	if c.local || len(conf.Cache) == 0 || c.cacher == nil {
		return
	}

	container := c.cacher.Restore(c.metadata.Repo.Name, c.metadata.Curr.Commit.Branch, conf.Cache)
	name := fmt.Sprintf("%s_restore_cache", c.prefix)
	step := c.createProcess(name, container, "cache")

	stage := new(backend.Stage)
	stage.Name = name
	stage.Alias = "restore_cache"
	stage.Steps = append(stage.Steps, step)

	ir.Stages = append(ir.Stages, stage)
}

func (c *Compiler) setupCacheRebuild(conf *yaml.Config, ir *backend.Config) {
	if c.local || len(conf.Cache) == 0 || c.metadata.Curr.Event != "push" || c.cacher == nil {
		return
	}
	container := c.cacher.Rebuild(c.metadata.Repo.Name, c.metadata.Curr.Commit.Branch, conf.Cache)

	name := fmt.Sprintf("%s_rebuild_cache", c.prefix)
	step := c.createProcess(name, container, "cache")

	stage := new(backend.Stage)
	stage.Name = name
	stage.Alias = "rebuild_cache"
	stage.Steps = append(stage.Steps, step)

	ir.Stages = append(ir.Stages, stage)
}
