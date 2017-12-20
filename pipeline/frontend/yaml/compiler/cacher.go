package compiler

import (
	"strings"
	"github.com/BoxLinker/cicd/pipeline/frontend/yaml"
	"path"
	libcompose "github.com/docker/libcompose/yaml"
)

// Cacher defines a compiler transform that can be used
// to implement default caching for a repository.
type Cacher interface {
	Restore(repo, branch string, mounts []string) *yaml.Container
	Rebuild(repo, branch string, mounts []string) *yaml.Container
}

type volumeCacher struct {
	base string
}

func (c *volumeCacher) Restore(repo, branch string, mounts []string) *yaml.Container {
	return &yaml.Container{
		Name:  "rebuild_cache",
		Image: "plugins/volume-cache:1.0.0",
		Vargs: map[string]interface{}{
			"mount":       mounts,
			"path":        "/cache",
			"restore":     true,
			"file":        strings.Replace(branch, "/", "_", -1) + ".tar",
			"fallback_to": "master.tar",
		},
		Volumes: libcompose.Volumes{
			Volumes: []*libcompose.Volume{
				{
					Source:      path.Join(c.base, repo),
					Destination: "/cache",
					// TODO add access mode
				},
			},
		},
	}
}

func (c *volumeCacher) Rebuild(repo, branch string, mounts []string) *yaml.Container {
	return &yaml.Container{
		Name:  "rebuild_cache",
		Image: "plugins/volume-cache:1.0.0",
		Vargs: map[string]interface{}{
			"mount":   mounts,
			"path":    "/cache",
			"rebuild": true,
			"flush":   true,
			"file":    strings.Replace(branch, "/", "_", -1) + ".tar",
		},
		Volumes: libcompose.Volumes{
			Volumes: []*libcompose.Volume{
				{
					Source:      path.Join(c.base, repo),
					Destination: "/cache",
					// TODO add access mode
				},
			},
		},
	}
}

