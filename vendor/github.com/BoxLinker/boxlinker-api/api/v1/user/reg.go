package user

import (
	"net/http"
	"github.com/BoxLinker/boxlinker-api"
	userModels "github.com/BoxLinker/boxlinker-api/controller/models/user"
	"regexp"
	"fmt"
	"github.com/BoxLinker/boxlinker-api/auth"
	"github.com/BoxLinker/boxlinker-api/modules/httplib"

	userSettings "github.com/BoxLinker/boxlinker-api/settings/user"
	emailApi "github.com/BoxLinker/boxlinker-api/api/v1/email"
	"encoding/json"
	"time"
	"github.com/Sirupsen/logrus"
)

type RegForm struct {
	Username 	string 	`json:"username"`
	Password 	string 	`json:"password"`
	Email 		string 	`json:"email"`
}

func (f *RegForm) validate() map[string]int {
	m := make(map[string]int)
	if f.Username == "" {
		m["username"] = boxlinker.STATUS_FIELD_REQUIRED
	}

	if f.Password == "" {
		m["password"] = boxlinker.STATUS_FIELD_REQUIRED
	} else if len(f.Password) < 6 {
		m["password"] = boxlinker.STATUS_FIELD_REGEX_FAILED
	}

	if f.Email == "" {
		m["email"] = boxlinker.STATUS_FIELD_REQUIRED
	} else {
		if ok, err := regexp.MatchString("[a-zA-Z0-9_-]+@[a-zA-Z0-9_-]+(\\.[a-zA-Z0-9_-]+)+", f.Email); err != nil {
			fmt.Printf("regexp err: %v", err)
		} else if !ok {
			m["email"] = boxlinker.STATUS_FIELD_REGEX_FAILED
		}
	}
	if len(m) != 0 {
		return m
	}
	return nil
}

func (a *Api) Reg(w http.ResponseWriter, r *http.Request) {
	form := &RegForm{}
	if err := boxlinker.ReadRequestBody(r, form); err != nil {
		boxlinker.Resp(w, boxlinker.STATUS_INTERNAL_SERVER_ERR, nil, err.Error())
		return
	}
	if msg := form.validate(); msg != nil {
		boxlinker.Resp(w, boxlinker.STATUS_FORM_VALIDATE_ERR, msg)
		return
	}
	if found, err := a.manager.IsUserExists(form.Username); err != nil {
		logrus.Errorf("err when IsUserExists(%s): %v", form.Username, err)
		boxlinker.Resp(w, boxlinker.STATUS_INTERNAL_SERVER_ERR, nil, err.Error())
		return
	} else if found {
		boxlinker.Resp(w, boxlinker.STATUS_USER_EXISTS, nil)
		return
	}

	if found, err := a.manager.IsEmailExists(form.Email); err != nil {
		logrus.Errorf("err when IsEmailExists(%s): %v", form.Email, err)
		boxlinker.Resp(w, boxlinker.STATUS_INTERNAL_SERVER_ERR, nil, err.Error())
		return
	} else if found {
		boxlinker.Resp(w, boxlinker.STATUS_EMAIL_EXISTS, nil)
		return
	}


	//pass, err := auth.Hash(form.Password)
	pass, err := auth.Hash(form.Password)
	if err != nil {
		boxlinker.Resp(w, boxlinker.STATUS_INTERNAL_SERVER_ERR, fmt.Errorf("auth hash failed: %v", err))
		return
	}

	u := &userModels.UserToBeConfirmed{
		Name: form.Username,
		Password: string(pass),
		Email: form.Email,
	}

	if err := a.manager.SaveUserToBeConfirmed(u); err != nil {
		boxlinker.Resp(w, 1, nil, err.Error())
		return
	}

	logrus.Debugf("gen verify email token: uid:%s, name:%s", u.Id, u.Name)
	token, err := a.manager.GenerateToken(u.Id, u.Name, time.Now().Add(time.Minute * 15).Unix())

	if err != nil {
		boxlinker.Resp(w, boxlinker.STATUS_INTERNAL_SERVER_ERR, fmt.Errorf("generate token err: %v", err))
		return
	}

	eF := &emailApi.SendForm{
		To: []string{form.Email},
		Subject: "用户注册验证邮件 -- 无需回复",
		Body: 	fmt.Sprintf("<h3>点击下面的链接以完成注册(有效时间 15 分钟)：</h3><br/><a target=\"_blank\" href=\"%s\">%s</a>",
							fmt.Sprintf("%s?confirm_token=%s", userSettings.VERIFY_EMAIL_URI, token),
							"点击这里，验证邮箱",
				),
	}
	logrus.Debugf("send token auth email: %+v", eF)
	b, err := json.Marshal(eF)
	if err != nil {
		boxlinker.Resp(w, boxlinker.STATUS_INTERNAL_SERVER_ERR, fmt.Errorf("email form marshal err: %v", err))
		return
	}
	logrus.Debugf("send email to: %s", a.sendEmailUri)
	resp, err := httplib.Post(a.sendEmailUri).Body(b).Response()
	if err != nil {
		boxlinker.Resp(w, boxlinker.STATUS_INTERNAL_SERVER_ERR, fmt.Errorf("send email err: %v", err))
		return
	}
	status, msg, results, err := boxlinker.ParseResp(resp.Body)

	logrus.Debugf("send email results: %d, %s, %+v, %v", status, msg, results, err)

	// 发送邮件失败，删除 userToBeConfirmed
	if status != boxlinker.STATUS_OK || err != nil {
		if err := a.manager.DeleteUserToBeConfirmed(u.Id); err != nil {
			boxlinker.Resp(w, boxlinker.STATUS_INTERNAL_SERVER_ERR, fmt.Errorf("del userToBeConfirmed err: %v", err))
			return
		}
	}
	if err != nil {
		boxlinker.Resp(w, boxlinker.STATUS_INTERNAL_SERVER_ERR, fmt.Errorf("send email parse body err: %v", err))
		return
	}
	boxlinker.Resp(w, status, nil, msg)
}
