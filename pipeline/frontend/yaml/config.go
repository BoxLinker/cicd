package yaml

import (
	libcompose "github.com/docker/libcompose/yaml"
	"gopkg.in/yaml.v2"
)

type (
	Config struct {
		Cache 		libcompose.Stringorslice
		Platform 	string
		Branches 	Constraint
	}
)

func ParseBytes(b []byte) (*Config, error) {
	out := new(Config)
	err := yaml.Unmarshal(b, out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func ParseString(s string) (*Config, error) {
	return ParseBytes(
		[]byte(s),
	)
}