package authz

import (
	"regexp"
	"fmt"
	"net"
	"github.com/Sirupsen/logrus"
	"reflect"
	"strings"
	"strconv"
	"path"
	"github.com/BoxLinker/boxlinker-api/pkg/registry/authn"
	"encoding/json"
)

type ACL []ACLEntry

type ACLEntry struct {
	Match   *MatchConditions `yaml:"match"`
	Actions *[]string        `yaml:"actions,flow"`
	Comment *string          `yaml:"comment,omitempty"`
}

type MatchConditions struct {
	Account *string           `yaml:"account,omitempty" json:"account,omitempty"`
	Type    *string           `yaml:"type,omitempty" json:"type,omitempty"`
	Name    *string           `yaml:"name,omitempty" json:"name,omitempty"`
	IP      *string           `yaml:"ip,omitempty" json:"ip,omitempty"`
	Service *string           `yaml:"service,omitempty" json:"service,omitempty"`
	Labels  map[string]string `yaml:"labels,omitempty" json:"labels,omitempty"`
}

func (e ACLEntry) String() string {
	b, _ := json.Marshal(e)
	return string(b)
}

func validatePattern(p string) error {
	if len(p) > 2 && p[0] == '/' && p[len(p)-1] == '/' {
		_, err := regexp.Compile(p[1 : len(p)-1])
		if err != nil {
			return fmt.Errorf("invalid regex pattern: %s", err)
		}
	}
	return nil
}

func parseIPPattern(ipp string) (*net.IPNet, error) {
	ipnet := net.IPNet{}
	ipnet.IP = net.ParseIP(ipp)
	if ipnet.IP != nil {
		if ipnet.IP.To4() != nil {
			ipnet.Mask = net.CIDRMask(32, 32)
		} else {
			ipnet.Mask = net.CIDRMask(128, 128)
		}
		return &ipnet, nil
	} else {
		_, ipnet, err := net.ParseCIDR(ipp)
		if err != nil {
			return nil, err
		}
		return ipnet, nil
	}
}

func validateMatchConditions(mc *MatchConditions) error {
	for _, p := range []*string{mc.Account, mc.Type, mc.Name, mc.Service} {
		if p == nil {
			continue
		}
		err := validatePattern(*p)
		if err != nil {
			return fmt.Errorf("invalid pattern %q: %s", *p, err)
		}
	}
	if mc.IP != nil {
		_, err := parseIPPattern(*mc.IP)
		if err != nil {
			return fmt.Errorf("invalid IP pattern: %s", err)
		}
	}
	for k, v := range mc.Labels {
		err := validatePattern(v)
		if err != nil {
			return fmt.Errorf("invalid match pattern %q for label %s: %s", v, k, err)
		}
	}
	return nil
}

func ValidateACL(acl ACL) error {
	for i, e := range acl {
		err := validateMatchConditions(e.Match)
		if err != nil {
			return fmt.Errorf("entry %d, invalid match conditions: %s", i, err)
		}
	}
	return nil
}
type aclAuthorizer struct {
	acl ACL
}
func NewACLAuthorizer(acl ACL) (Authorizer, error) {
	if err := ValidateACL(acl); err != nil {
		return nil, err
	}
	logrus.Infof("Created ACL Authorizer with %d entries", len(acl))
	return &aclAuthorizer{acl: acl}, nil
}


func (aa *aclAuthorizer) Authorize(ai *AuthRequestInfo) ([]string, error) {
	for _, e := range aa.acl {
		matched := e.Matches(ai)
		if matched {
			logrus.Infof("%s matched %s (Comment: %s)", ai, e, *e.Comment)
			if len(*e.Actions) == 1 && (*e.Actions)[0] == "*" {
				return ai.Actions, nil
			}
			return StringSetIntersection(ai.Actions, *e.Actions), nil
		}
	}
	return nil, NoMatch
}

func (aa *aclAuthorizer) Stop() {
	// Nothing to do.
}

func (aa *aclAuthorizer) Name() string {
	return "static ACL"
}

func getField(i interface{}, name string) (string, bool) {
	s := reflect.Indirect(reflect.ValueOf(i))
	f := reflect.Indirect(s.FieldByName(name))
	if !f.IsValid() {
		return "", false
	}
	return f.String(), true
}
var captureGroupRegex = regexp.MustCompile(`\$\{(.+?):(\d+)\}`)

func matchString(pp *string, s string, vars []string) bool {
	if pp == nil {
		return true
	}
	p := strings.NewReplacer(vars...).Replace(*pp)

	var matched bool
	var err error
	if len(p) > 2 && p[0] == '/' && p[len(p)-1] == '/' {
		matched, err = regexp.Match(p[1:len(p)-1], []byte(s))
	} else {
		matched, err = path.Match(p, s)
	}
	return err == nil && matched
}

func matchIP(ipp *string, ip net.IP) bool {
	if ipp == nil {
		return true
	}
	if ip == nil {
		return false
	}
	ipnet, err := parseIPPattern(*ipp)
	if err != nil { // Can't happen, it supposed to have been validated
		logrus.Errorf("Invalid IP pattern: %s", *ipp)
	}
	return ipnet.Contains(ip)
}

func matchLabels(ml map[string]string, rl authn.Labels, vars []string) bool {
	for label, pattern := range ml {
		labelValues := rl[label]
		matched := false
		for _, lv := range labelValues {
			if matchString(&pattern, lv, vars) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	return true
}

func (mc *MatchConditions) Matches(ai *AuthRequestInfo) bool {
	vars := []string{
		"${account}", regexp.QuoteMeta(ai.Account),
		"${type}", regexp.QuoteMeta(ai.Type),
		"${name}", regexp.QuoteMeta(ai.Name),
		"${service}", regexp.QuoteMeta(ai.Service),
	}

	for _, x := range []string{"Account", "Type", "Name", "Service"} {
		field, _ := getField(mc, x)
		for _, found := range captureGroupRegex.FindAllStringSubmatch(field, -1) {
			key := strings.Title(found[1])
			index, _ := strconv.Atoi(found[2])
			field, has := getField(mc, key)
			if !has {
				logrus.Errorf("No field in '%s' in MatchConditions", key)
				continue
			}
			if len(field) < 2 || field[0] != '/' || field[len(field)-1] != '/' {
				continue
			}
			regex, err := regexp.Compile(field[1 : len(field)-1])
			if err != nil {
				logrus.Errorf("Invalid regex in '%s' of MatchConditions", key)
				continue
			}
			info, has := getField(ai, key)
			if !has {
				logrus.Errorf("No field in '%s' in AuthRequestInfo", key)
				continue
			}
			text := regex.FindStringSubmatch(info)
			if index < 1 || index > len(text) -1 {
				logrus.Errorf("%s: Capture group index out of range", key)
				continue
			}
			vars = append(vars, found[0], text[index])
		}
	}
	return matchString(mc.Account, ai.Account, vars) &&
		matchString(mc.Type, ai.Type, vars) &&
		matchString(mc.Name, ai.Name, vars) &&
		matchString(mc.Service, ai.Service, vars) &&
		matchIP(mc.IP, ai.IP) &&
		matchLabels(mc.Labels, ai.Labels, vars)
}

func (e *ACLEntry) Matches(ai *AuthRequestInfo) bool {
	return e.Match.Matches(ai)
}