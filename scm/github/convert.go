package github

import (
	"github.com/google/go-github/github"
	"github.com/BoxLinker/cicd/models"
	"strings"
	"fmt"
)

const defaultBranch = "master"

const (
	statusPending = "pending"
	statusSuccess = "success"
	statusFailure = "failure"
	statusError   = "error"
)

const (
	descPending  = "this build is pending"
	descSuccess  = "the build was successful"
	descFailure  = "the build failed"
	descBlocked  = "the build requires approval"
	descDeclined = "the build was rejected"
	descError    = "oops, something went wrong"
)


const (
	headRefs  = "refs/pull/%d/head"  // pull request unmerged
	mergeRefs = "refs/pull/%d/merge" // pull request merged with base
	refspec   = "%s:%s"
)

func convertRepoList(from []github.Repository, u *models.User) []*models.Repo {
	var repos []*models.Repo
	for _, repo := range from {
		repos = append(repos, convertRepo(&repo, u))
	}
	return repos
}

func convertRepo(form *github.Repository, u *models.User) *models.Repo {
	repo := &models.Repo{
		Owner: 		*form.Owner.Login,
		Name: 		*form.Name,
		FullName: 	*form.FullName,
		Link: 		*form.HTMLURL,
		IsPrivate: 	*form.Private,
		Clone: 		*form.CloneURL,
		Avatar: 	*form.Owner.AvatarURL,
		SCM: 		models.RepoGithub,
		Branch: 	defaultBranch,
		UserID: 	u.ID,
	}
	if form.DefaultBranch != nil {
		repo.Branch = *form.DefaultBranch
	}
	return repo
}


// 工具方法，将 github 的 webhook 信息转换成 repo
func convertRepoHook(from *webhook) *models.Repo {
	repo := &models.Repo{
		Owner: 	from.Repo.Owner.Login,
		Name: 	from.Repo.Name,
		FullName: 	from.Repo.FullName,
		Link: 		from.Repo.HTMLURL,
		IsPrivate: 	from.Repo.Private,
		Clone: 		from.Repo.CloneURL,
		Branch: 	from.Repo.DefaultBranch,
		SCM: 		models.RepoGithub,
	}
	if repo.Branch == "" {
		repo.Branch = defaultBranch
	}
	if repo.Owner == "" { // legacy webhooks
		repo.Owner = from.Repo.Owner.Name
	}
	if repo.FullName == "" {
		repo.FullName = repo.Owner + "/" + repo.Name
	}
	return repo
}
// 工具方法，将 github 的 webhook 信息转换成 Build
func convertPushHook(from *webhook) *models.Build {
	build := &models.Build{
		Event: models.EventPush,
		Commit:  from.Head.ID,
		Ref: 	 from.Ref,
		Link: 	 from.Head.URL,
		Branch:  strings.Replace(from.Ref, "refs/heads/", "", -1),
		Message: from.Head.Message,
		Email:   from.Head.Author.Email,
		Avatar:  from.Sender.Avatar,
		Author:  from.Sender.Login,
		Remote:  from.Repo.CloneURL,
		Sender:  from.Sender.Login,
	}
	if len(build.Author) == 0 {
		build.Author = from.Head.Author.Username
	}
	if len(build.Email) == 0 {
		// 怎么搞?
	}
	if strings.HasPrefix(build.Ref, "refs/tags/") {
		build.Event = models.EventTag
		//对于 tag 事件，tag 的 base branch（base_ref）可以作为 build 的 branch
		if strings.HasPrefix(from.BaseRef, "refs/heads/") {
			build.Branch = strings.Replace(from.BaseRef, "refs/heads/", "", -1)
		}
	}
	return build
}

func convertDeployHook(from *webhook) *models.Build {
	build := &models.Build{
		Event: 	models.EventDeploy,
		Commit:  from.Deployment.Sha,
		Link:    from.Deployment.URL,
		Message: from.Deployment.Desc,
		Avatar:  from.Sender.Avatar,
		Author:  from.Sender.Login,
		Ref:     from.Deployment.Ref,
		Branch:  from.Deployment.Ref,
		Deploy:  from.Deployment.Env,
		Sender:  from.Sender.Login,
	}
	// if the ref is a sha or short sha we need to manuallyconstruct the ref.
	if strings.HasPrefix(build.Commit, build.Ref) || build.Commit == build.Ref {
		build.Branch = from.Repo.DefaultBranch
		if build.Branch == "" {
			build.Branch = defaultBranch
		}
		build.Ref = fmt.Sprintf("refs/heads/%s", build.Branch)
	}
	// if the ref is a branch we should make sure it has refs/heads prefix
	if !strings.HasPrefix(build.Ref, "refs/") { // branch or tag
		build.Ref = fmt.Sprintf("refs/heads/%s", build.Branch)
	}
	return build
}

// convertPullHook is a helper function used to extract the Build details
// from a pull request webhook and convert to the common Drone Build structure.
func convertPullHook(from *webhook, merge bool) *models.Build {
	build := &models.Build{
		Event:   models.EventPull,
		Commit:  from.PullRequest.Head.SHA,
		Link:    from.PullRequest.HTMLURL,
		Ref:     fmt.Sprintf(headRefs, from.PullRequest.Number),
		Branch:  from.PullRequest.Base.Ref,
		Message: from.PullRequest.Title,
		Author:  from.PullRequest.User.Login,
		Avatar:  from.PullRequest.User.Avatar,
		Title:   from.PullRequest.Title,
		Sender:  from.Sender.Login,
		Remote:  from.PullRequest.Head.Repo.CloneURL,
		Refspec: fmt.Sprintf(refspec,
			from.PullRequest.Head.Ref,
			from.PullRequest.Base.Ref,
		),
	}
	if merge {
		build.Ref = fmt.Sprintf(mergeRefs, from.PullRequest.Number)
	}
	return build
}


// convertStatus is a helper function used to convert a Drone status to a
// GitHub commit status.
func convertStatus(status string) string {
	switch status {
	case models.StatusPending, models.StatusRunning, models.StatusBlocked:
		return statusPending
	case models.StatusFailure, models.StatusDeclined:
		return statusFailure
	case models.StatusSuccess:
		return statusSuccess
	default:
		return statusError
	}
}

// convertDesc is a helper function used to convert a Drone status to a
// GitHub status description.
func convertDesc(status string) string {
	switch status {
	case models.StatusPending, models.StatusRunning:
		return descPending
	case models.StatusSuccess:
		return descSuccess
	case models.StatusFailure:
		return descFailure
	case models.StatusBlocked:
		return descBlocked
	case models.StatusDeclined:
		return descDeclined
	default:
		return descError
	}
}
