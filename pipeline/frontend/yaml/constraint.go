package yaml

import (
	"github.com/BoxLinker/cicd/pipeline/frontend/yaml/types"
	"path/filepath"
	libcompose "github.com/docker/libcompose/yaml"
)

type (
	Constraints struct {
		Ref         Constraint
		Repo        Constraint
		Instance    Constraint
		Platform    Constraint
		Environment Constraint
		Event       Constraint
		Branch      Constraint
		Status      Constraint
		Matrix      ConstraintMap
		Local       types.BoolTrue
	}

	Constraint struct {
		Include []string
		Exclude []string
	}

	ConstraintMap struct {
		Include map[string]string
		Exclude map[string]string
	}
)


// Match returns true if the string matches the include patterns and does not
// match any of the exclude patterns.
func (c *Constraint) Match(v string) bool {
	if c.Excludes(v) {
		return false
	}
	if c.Includes(v) {
		return true
	}
	if len(c.Include) == 0 {
		return true
	}
	return false
}

// Includes returns true if the string matches the include patterns.
func (c *Constraint) Includes(v string) bool {
	for _, pattern := range c.Include {
		if ok, _ := filepath.Match(pattern, v); ok {
			return true
		}
	}
	return false
}

// Excludes returns true if the string matches the exclude patterns.
func (c *Constraint) Excludes(v string) bool {
	for _, pattern := range c.Exclude {
		if ok, _ := filepath.Match(pattern, v); ok {
			return true
		}
	}
	return false
}

// UnmarshalYAML unmarshals the constraint.
func (c *Constraint) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var out1 = struct {
		Include libcompose.Stringorslice
		Exclude libcompose.Stringorslice
	}{}

	var out2 libcompose.Stringorslice

	unmarshal(&out1)
	unmarshal(&out2)

	c.Exclude = out1.Exclude
	c.Include = append(
		out1.Include,
		out2...,
	)
	return nil
}
