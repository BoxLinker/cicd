package models

type SCMType string
const (
	GITHUB SCMType = "github"
	GITLAB SCMType = "gitlab"
)

func (s SCMType) Exists() bool {
	switch s {
	case GITHUB, GITLAB:
		return true
	default:
		return false
	}
}