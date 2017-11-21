package middleware

import "net/http"

type Middleware interface {
	HandlerFuncWithNext(w http.ResponseWriter, r *http.Request, next http.HandlerFunc)
}
