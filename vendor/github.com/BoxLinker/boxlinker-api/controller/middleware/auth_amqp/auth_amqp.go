package auth_amqp

import "net/http"

type authAmqpRequired struct {
	authorization string
}

func NewAuthAmqpRequired(authorization string) *authAmqpRequired {
	return &authAmqpRequired{
		authorization: authorization,
	}
}

func (a *authAmqpRequired) HandlerFuncWithNext(w http.ResponseWriter, r *http.Request, next http.HandlerFunc){
	authorization := r.Header.Get("Authorization")
	if a.authorization == authorization && next != nil {
		next(w, r)
	} else {
		http.Error(w, "", http.StatusUnauthorized)
	}
}