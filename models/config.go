package models

type ConfigStore interface {
	ConfigLoad(int64) (*Config, error)
	ConfigFind(*Repo, string) (*Config, error)
	ConfigFindApproved(*Config) (bool, error)
	ConfigCreate(*Config) error
}

type Config struct {
	ID 		int64 	`json:"-" meddler:"config_id,pk"`
	RepoID 	int64 	`json:"-" meddler:"config_repo_id"`
	Data 	string 	`json:"data" meddler:"config_data"`
	Hash 	string 	`json:"hash" meddler:"config_hash"`
}