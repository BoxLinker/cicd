package manager

import (
	mAuth "github.com/BoxLinker/boxlinker-api/auth"
	"errors"
)

type Manager interface {
	VerifyAuthToken(token string) (map[string]interface{}, error)


}

type DefaultManager struct {
}




func (m DefaultManager) VerifyAuthToken(token string) (map[string]interface{}, error) {
	errFailed := errors.New("Token 解析失败")
	success, data, err := mAuth.AuthToken(token)
	if err != nil {
		return nil, err
	}
	_username := data["username"]
	if _username == nil {
		return nil, errFailed
	}
	_uid := data["uid"]
	if _uid == nil {
		return nil, errFailed
	}
	//username := _username.(string)
	//u := m.GetUserByName(username)
	//if u == nil {
	//	return nil, fmt.Errorf("未找到用户: %s", username)
	//}
	//
	if !success {
		return nil, errFailed
	}

	return map[string]interface{}{
		"uid": _uid.(string),
		"username": _username.(string),
	}, nil
}
