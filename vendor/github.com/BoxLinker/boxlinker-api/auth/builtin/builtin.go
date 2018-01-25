package builtin

import (
	"github.com/BoxLinker/boxlinker-api/auth"
	"golang.org/x/crypto/bcrypt"
)

type (
	BuiltinAuthenticator struct {
	}
)

func NewAuthenticator() auth.Authenticator {
	return &BuiltinAuthenticator{}
}

func (a BuiltinAuthenticator) IsUpdateSupported() bool {
	return true
}

func (a BuiltinAuthenticator) Name() string {
	return "builtin"
}

func (a BuiltinAuthenticator) Authenticate(username, password, hash string) (bool, error) {
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err == nil {
		return true, nil
	} else {
		return false, nil
	}
}

func (a BuiltinAuthenticator) GenerateToken(uid string, username string, exp ...int64) (string, error) {
	return auth.GenerateToken(uid, username, exp...)
}