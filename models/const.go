package models

const (
	RepoGithub = "github"
)

const (
	EventPush   = "push"
	EventPull   = "pull_request"
	EventTag    = "tag"
	EventDeploy = "deployment"
)

const (
	StatusSkipped  = "skipped"
	StatusPending  = "pending"
	StatusRunning  = "running"
	StatusSuccess  = "success"
	StatusFailure  = "failure"
	StatusKilled   = "killed"
	StatusError    = "error"
	StatusBlocked  = "blocked"
	StatusDeclined = "declined"
)

const (
	VisibilityPublic   = "public"
	VisibilityPrivate  = "private"
	VisibilityInternal = "internal"
)