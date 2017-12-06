package server

import (
	"github.com/urfave/cli"
	"github.com/BoxLinker/cicd/scm"
	"github.com/BoxLinker/cicd/scm/github"
	"github.com/BoxLinker/cicd/models"
	"github.com/BoxLinker/cicd/store"
	"github.com/BoxLinker/cicd/queue"
)

func setupQueue(c *cli.Context, s store.Store) queue.Queue {
	return models.WithTaskStore(queue.New(), s)
}

func SetupCodeBase(c *cli.Context) (map[string]scm.SCM, error) {
	m := map[string]scm.SCM{}
	var err error

	if c.Bool(models.GITHUB) {
		m[models.GITHUB], err = setupGithub(c)
		if err != nil {
			return nil, err
		}
	}
	return m, nil
}

func setupGithub(c *cli.Context)(scm.SCM, error){
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