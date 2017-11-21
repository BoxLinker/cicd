package main

import (
	"github.com/urfave/cli"
	"github.com/BoxLinker/cicd/codebase"
	"fmt"
	"github.com/BoxLinker/cicd/codebase/github"
)

func SetupCodeBase(c *cli.Context) (codebase.CodeBase, error) {
	switch {
	case c.Bool("github"):
		return setupGithub(c)
	default:
		return nil, fmt.Errorf("vcs not configured")
	}
}

func setupGithub(c *cli.Context)(codebase.CodeBase, error){
	return github.New(github.Opts{
		HomeHost: 	 c.String("home-host"),
		URL:         c.String("github-server"),
		Context:     c.String("github-context"),
		Client:      c.String("github-client"),
		Secret:      c.String("github-secret"),
		Scopes:      c.StringSlice("github-scope"),
		Username:    c.String("github-git-username"),
		Password:    c.String("github-git-password"),
		PrivateMode: c.Bool("github-private-mode"),
		SkipVerify:  c.Bool("github-skip-verify"),
		MergeRef:    c.BoolT("github-merge-ref"),
	})
}