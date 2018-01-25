package registry


type Event struct {
	Id 			string 	`json:"id"`
	Timestamp 	string 	`json:"timestamp"`
	Action 		string 	`json:"action"`
	Target 		struct{
		MediaType 		string 		`json:"mediaType"`
		Size 			int64 		`json:"size"`
		Digest 			string 		`json:"digest"`
		Length 			int64 		`json:"length"`
		Repository 		string 		`json:"repository"`
		Url 			string 		`json:"url"`
		Tag 			string 		`json:"tag"`
	} `json:"target"`
	Request 	struct{
		Id 		string 		`json:"id"`
		Addr 	string 		`json:"addr"`
		Host	string 		`json:"host"`
		Method 	string 		`json:"method"`
		UserAgent string 	`json:"useragent"`
	} 	`json:"request"`
	Source 		struct{
		Addr 	string 	`json:"addr"`
		InstanceID string `json:"instanceID"`
	} 	`json:"source"`
}

type EventCallback struct {
	Events [] Event	`json:"events"`
}