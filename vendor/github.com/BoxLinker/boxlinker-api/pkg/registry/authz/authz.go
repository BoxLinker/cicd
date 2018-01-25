package authz

import (
	"errors"
	"net"
	"github.com/BoxLinker/boxlinker-api/pkg/registry/authn"
	"fmt"
	"strings"
)

type Authorizer interface {
	Authorize(ai *AuthRequestInfo) ([]string, error)
	Stop()
	Name() string
}

var NoMatch = errors.New("did not match any rule")

type AuthRequestInfo struct {
	Account string
	Type    string
	Name    string
	Service string
	IP      net.IP
	Actions []string
	Labels  authn.Labels
}

func (ai AuthRequestInfo) String() string {
	return fmt.Sprintf("{%s %s %s %s}", ai.Account, strings.Join(ai.Actions, ","), ai.Type, ai.Name)
}