package httplib

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/rs/cors"
)

func ParsePageConfig(r *http.Request) PageConfig {
	q := r.URL.Query()
	currentPageS := q.Get("currentPage")

	currentPage, err := strconv.ParseInt(currentPageS, 10, 0)
	if err != nil || currentPage <= 0 {
		currentPage = 1
	}

	pageCountS := q.Get("pageCount")

	pageCount, err := strconv.ParseInt(pageCountS, 10, 0)
	if err != nil || pageCount <= 0 {
		pageCount = 10
	}
	pc := PageConfig{}
	pc.CurrentPage = int(currentPage)
	pc.PageCount = int(pageCount)
	return pc
}

func ReadRequestBody(r *http.Request, bean interface{}) error {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, bean); err != nil {
		return err
	}
	return nil
}

func GetQueryParam(r *http.Request, name string) string {
	valArr := r.URL.Query()[name]
	if len(valArr) == 1 {
		return valArr[0]
	}
	return ""
}

var Cors = cors.New(cors.Options{
	AllowedOrigins: []string{"*"},
	AllowedMethods: []string{"GET", "POST", "DELETE", "PUT", "OPTIONS"},
	AllowedHeaders: []string{"Origin", "Content-Type", "Accept", "token", "X-Requested-With", "X-Access-Token"},
})
