package registry

type Action struct {
	Id 		string 	`xorm:"pk"`
	ImageID string 	`xorm:"NOT NULL"`
	Action 	string
	Timestamp string
	TargetMediaType string
}