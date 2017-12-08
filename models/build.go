package models

type Build struct {
	ID 			int64 		`json:"id"        meddler:"build_id,pk"`
	RepoID 		int64 		`json:"-"         meddler:"build_repo_id"`
	ConfigID 	int64 		`json:"-"         meddler:"build_config_id"`
	Number 		int 		`json:"number"    meddler:"build_number"`
	Parent    int     `json:"parent"        meddler:"build_parent"`

	Event 		string 		`json:"event"     meddler:"build_event"`
	Status 		string 		`json:"status"    meddler:"build_status"`
	Error     string  `json:"error"         meddler:"build_error"`
	Created 	int64 		`json:"created_at" meddler:"build_created"`
	Started 	int64 		`json:"started_at" meddler:"build_started"`
	Finished 	int64 		`json:"finished_at" meddler:"build_finished"`
	Enqueued  int64   `json:"enqueued_at"   meddler:"build_enqueued"`
	Commit 		string 		`json:"commit"      meddler:"build_commit"`
	Branch 		string 		`json:"branch"      meddler:"build_branch"`
	Ref       string  `json:"ref"           meddler:"build_ref"`
	Sender 		string 		`json:"sender"      meddler:"build_sender"`
	Author 		string 		`json:"author"      meddler:"build_author"`
	Avatar 		string 		`json:"avatar"      meddler:"build_avatar"`
	Email 		string 		`json:"email"      meddler:"build_email"`
	Remote 		string 		`json:"remote"        meddler:"build_remote"`
	Title     string  `json:"title"         meddler:"build_title"`
	Message   string  `json:"message"       meddler:"build_message"`
	Link      string  `json:"link_url"      meddler:"build_link"`
	Deploy    string  `json:"deploy_to"     meddler:"build_deploy"`
	Refspec   string  `json:"refspec"       meddler:"build_refspec"`
	Verified  bool    `json:"verified"      meddler:"build_verified"` // deprecate

	Procs     []*Proc `json:"procs,omitempty" meddler:"-"`

}

// Trim trims string values that would otherwise exceed
// the database column sizes and fail to insert.
func (b *Build) Trim() {
	if len(b.Title) > 1000 {
		b.Title = b.Title[:1000]
	}
	if len(b.Message) > 2000 {
		b.Message = b.Message[:2000]
	}
}
