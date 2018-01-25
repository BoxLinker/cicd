package user

import (
	"net/http"
	"github.com/BoxLinker/boxlinker-api"
	log "github.com/Sirupsen/logrus"
	settings "github.com/BoxLinker/boxlinker-api/settings/user"
	"time"
	"golang.org/x/crypto/bcrypt"
)

type LoginForm struct {
	Username string		`json:"username"`
	Password string		`json:"password"`
}

func (f *LoginForm) validate() string {
	if f.Username == "" {
		return "您还没有填写用户名"
	}
	if f.Password == "" {
		return "您还没有填写用户名"
	}
	return ""
}

func (a *Api) BasicAuth(w http.ResponseWriter, r *http.Request) {
	user, pass, ok := r.BasicAuth()
	if !ok {
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	log.Debugf("user: %s, pass: %s", user, pass)
	u := a.manager.GetUserByName(user)
	if u == nil {
		log.Debugf("user %s not found", user)
		http.Error(w, "", http.StatusNotFound)
		return
	}
	log.Debugf("user found (%+v)", u)

	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(pass)); err != nil {
		http.Error(w, "", http.StatusUnauthorized)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func (a *Api) Login(w http.ResponseWriter, r *http.Request){
	form := &LoginForm{}
	if err := boxlinker.ReadRequestBody(r, form); err != nil {
		boxlinker.Resp(w, boxlinker.STATUS_INTERNAL_SERVER_ERR, nil, err.Error())
		return
	}
	log.Debugf("form: %v", form)
	if msg := form.validate(); msg != "" {
		boxlinker.Resp(w, 1, nil, msg)
		return
	}

	u := a.manager.GetUserByName(form.Username)
	if u == nil {
		boxlinker.Resp(w, boxlinker.STATUS_NOT_FOUND, nil, "user not found")
		return
	}
	success, err := a.manager.VerifyUsernamePassword(form.Username, form.Password, u.Password)
	if err != nil {
		boxlinker.Resp(w, 1, nil, err.Error())
		return
	}
	if !success {
		boxlinker.Resp(w, 1, nil, "failed")
		return
	}
	token, err := a.manager.GenerateToken(u.Id, u.Name)
	if err != nil {
		boxlinker.Resp(w, 1, nil, err.Error())
		return
	}
	cookie := &http.Cookie{
		Name:"X-Access-Token",
		Value: token,
		Expires: time.Now().Add(30*24*time.Hour),
		Domain: settings.COOKIE_DOMAIN,
	}
	http.SetCookie(w, cookie)
	boxlinker.Resp(w, 0, map[string]string{
		"X-Access-Token": token,
	})
}




