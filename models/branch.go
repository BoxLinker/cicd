package models

type Branch struct {
	ID     int64  `json:"branch_id"             meddler:"branch_id,pk"`
	Name   string `json:"name"                   meddler:"name"`
	RepoID int64  `json:"repo_id"  meddler:"branch_repo_id"`
}
