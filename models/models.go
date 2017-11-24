package models

const (
	GITHUB = "github"
	GITLAB = "gitlab"
)

func SCMExists(s string) bool {
	switch s {
	case GITHUB, GITLAB:
		return true
	default:
		return false
	}
}