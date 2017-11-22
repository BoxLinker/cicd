package github

import (
	"github.com/google/go-github/github"
	"github.com/BoxLinker/cicd/models"
)

const defaultBranch = "master"

func convertRepoList(from []github.Repository, u *models.CodeBaseUser) []*models.Repo {
	var repos []*models.Repo
	for _, repo := range from {
		repos = append(repos, convertRepo(&repo, u))
	}
	return repos
}

func convertRepo(form *github.Repository, u *models.CodeBaseUser) *models.Repo {
	repo := &models.Repo{
		UserID: u.ID,
		Name: *form.Name,
		FullName: *form.FullName,
		Owner: *form.Owner.Login,
		Link: *form.HTMLURL,
		Clone: *form.CloneURL,
		Kind: models.RepoGithub,
		Branch: defaultBranch,
		IsPrivate: *form.Private,
	}
	if form.DefaultBranch != nil {
		repo.Branch = *form.DefaultBranch
	}
	return repo
}