package boxlinker

import (
	"net/http"
	"encoding/json"
	"io/ioutil"
	"io"
)

type respResults struct {
	status int
	msg string
	results interface{}
}

func ParseResp(body io.ReadCloser) (int, string, interface{}, error){
	b, err := ioutil.ReadAll(body)
	if err != nil {
		return -1, "", nil, err
	}
	re := &respResults{}
	if err := json.Unmarshal(b, re); err != nil {
		return -1, "", nil, err
	}
	return re.status, re.msg, re.results, nil
}

func Resp(w http.ResponseWriter, status int, results interface{}, msg ...string){
	var (
		err error
		outS []byte
	)

	msgS := ""
	if len(msg) == 1 {
		msgS = msg[0]
	}

	outS, err = json.Marshal(map[string]interface{}{
		"status": status,
		"msg": msgS,
		"results": results,
	})
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.WriteHeader(200)
	w.Write(outS)
}
