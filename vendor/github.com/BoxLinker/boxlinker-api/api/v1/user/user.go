package user

import (
	"net/http"
	"github.com/BoxLinker/boxlinker-api"
	"github.com/BoxLinker/boxlinker-api/auth"
)

type (
	ChangePasswordForm struct {
		OldPassword string		`json:"old_password"`
		NewPassword string		`json:"new_password"`
		ConfirmPassword string	`json:"confirm_password"`
	}
)

func (f *ChangePasswordForm) validate() (map[string]int){
	m := make(map[string]int)
	if f.OldPassword == "" {
		m["old_password"] = boxlinker.STATUS_FIELD_REQUIRED
		return m
	}
	if f.NewPassword == "" {
		m["new_password"] = boxlinker.STATUS_FIELD_REQUIRED
		return m
	} else if len(f.NewPassword) < 6 {
		m["new_password"] = boxlinker.STATUS_FIELD_REGEX_FAILED
		return m
	}

	if f.NewPassword != f.ConfirmPassword {
		m["confirm_password"] = boxlinker.STATUS_PASSWORD_CONFIRM_FAILED
		return m
	}

	if f.NewPassword == f.OldPassword {
		m["new_password"] = boxlinker.STATUS_NEW_OLD_PASSWORD_SAME
		return m
	}

	if len(m) == 0 {
		return nil
	}
	return m
}

func (a *Api) GetUsers(w http.ResponseWriter, r *http.Request) {
	users, err := a.manager.GetUsers(boxlinker.ParseHTTPQuery(r))
	if err != nil {
		boxlinker.Resp(w, boxlinker.STATUS_INTERNAL_SERVER_ERR, nil, err.Error())
		return
	}
	var results []map[string]interface{}
	for _, user := range users {
		results = append(results, user.APIJson())
	}
	boxlinker.Resp(w, boxlinker.STATUS_OK, results)
}

func (a *Api) GetUser(w http.ResponseWriter, r *http.Request){
	us := r.Context().Value("user")
	if us == nil {
		boxlinker.Resp(w, boxlinker.STATUS_NOT_FOUND, nil)
		return
	}
	ctx := us.(map[string]interface{})
	if ctx == nil || ctx["uid"] == nil {
		boxlinker.Resp(w, boxlinker.STATUS_NOT_FOUND, nil)
		return
	}
	id := ctx["uid"].(string)
	u := a.manager.GetUserById(id)
	if u == nil {
		boxlinker.Resp(w, boxlinker.STATUS_NOT_FOUND, nil, "not found")
		return
	}
	boxlinker.Resp(w, boxlinker.STATUS_OK, u.APIJson())
}

func (a *Api) ChangePassword(w http.ResponseWriter, r *http.Request) {

	form := &ChangePasswordForm{}

	if err := boxlinker.ReadRequestBody(r, form); err != nil {
		boxlinker.Resp(w, boxlinker.STATUS_INTERNAL_SERVER_ERR, nil, err.Error())
		return
	}
	if validate := form.validate(); validate != nil {
		boxlinker.Resp(w, boxlinker.STATUS_FORM_VALIDATE_ERR, validate)
		return
	}
	ctx := r.Context().Value("user").(map[string]interface{})
	if ctx == nil || ctx["id"] == nil {
		boxlinker.Resp(w, boxlinker.STATUS_NOT_FOUND, nil)
		return
	}
	id := ctx["id"].(string)
	u := a.manager.GetUserById(id)
	if u == nil {
		boxlinker.Resp(w, boxlinker.STATUS_NOT_FOUND, nil, "not found")
		return
	}
	// 验证原始密码正确性
	if ok, err := a.manager.VerifyUsernamePassword(u.Name, form.OldPassword, u.Password); err != nil {
		boxlinker.Resp(w, boxlinker.STATUS_INTERNAL_SERVER_ERR,nil, err.Error())
		return
	} else if !ok {
		boxlinker.Resp(w, boxlinker.STATUS_OLD_PASSWORD_AUTH_FAILED,nil)
		return
	}

	hash, err := auth.Hash(form.NewPassword)
	if err != nil {
		boxlinker.Resp(w, boxlinker.STATUS_INTERNAL_SERVER_ERR, nil, err.Error())
		return
	}
	if success, err := a.manager.UpdatePassword(u.Id, hash); err != nil {
		boxlinker.Resp(w, boxlinker.STATUS_INTERNAL_SERVER_ERR, nil, err.Error())
		return
	} else if !success {
		boxlinker.Resp(w, boxlinker.STATUS_FAILED, nil)
		return
	} else {
		boxlinker.Resp(w, boxlinker.STATUS_OK, nil)
	}
}