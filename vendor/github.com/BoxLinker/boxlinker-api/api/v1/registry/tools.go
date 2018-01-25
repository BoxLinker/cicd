package registry

import (
	"io/ioutil"
	"gopkg.in/yaml.v2"
	"fmt"
	"github.com/docker/libtrust"
	"crypto/tls"
	"crypto/x509"
	"strings"
	"github.com/docker/distribution/registry/auth/token"
	"encoding/json"
	"time"
	"math/rand"
	"encoding/base64"
	"github.com/BoxLinker/boxlinker-api/pkg/registry/authz"
	"sort"
	"github.com/Sirupsen/logrus"
)

type Config struct {
	Server struct {
		Addr string `yaml:"addr,omitempty"`
		Debug bool `yaml:"debug"`
	}    `yaml:"server,omitempty"`
	Token struct {
		Issuer      string    `yaml:"issuer,omitempty"`
		Expiration  int64    `yaml:"expiration,omitempty"`
		Certificate string    `yaml:"certificate,omitempty"`
		Key         string    `yaml:"key,omitempty"`

		privateKey libtrust.PrivateKey
		publicKey  libtrust.PublicKey
	} `yaml:"token,omitempty"`
	DB struct{
		Host string `yaml:"host,omitempty"`
		Port int `yaml:"port,omitempty"`
		User string `yaml:"user,omitempty"`
		Password string `yaml:"password,omitempty"`
		Name string `yaml:"name,omitempty"`
	} `yaml:"db,omitempty"`
	Auth struct{
		TokenAuthUrl string `yaml:"tokenAuthUrl,omitempty"`
		BasicAuthUrl string `yaml:"basicAuthUrl,omitempty"`
	} `yaml:"auth,omitempty"`
	ACL authz.ACL `yaml:"acl,omitempty"`
}

func (c *Config) GenerateToken(ar *authRequest, ares []authzResult) (string, error) {
	logrus.Debugf("GenerateToken: issuer: %s, subject: %s, audience: %s, ares: %+v",
		c.Token.Issuer, ar.Account, ar.Service, ares)
	now := time.Now().Unix()
	_, sigAlg, err := c.Token.privateKey.Sign(strings.NewReader("dummy"), 0)
	if err != nil {
		return "", fmt.Errorf("failed to sign: %s", err)
	}
	header := token.Header{
		Type: "JWT",
		SigningAlg: sigAlg,
		KeyID: c.Token.publicKey.KeyID(),
	}
	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", fmt.Errorf("failed to marshal header: %s", err)
	}
	claims := token.ClaimSet{
		Issuer:     c.Token.Issuer,
		Subject:    ar.Account,
		Audience:   ar.Service,
		NotBefore:  now - 10,
		IssuedAt:   now,
		Expiration: now + c.Token.Expiration,
		JWTID:      fmt.Sprintf("%d", rand.Int63()),
		Access:     []*token.ResourceActions{},
	}
	for _, a := range ares {
		ra := &token.ResourceActions{
			Type:    a.scope.Type,
			Name:    a.scope.Name,
			Actions: a.authorizedActions,
		}
		if ra.Actions == nil {
			ra.Actions = []string{}
		}
		sort.Strings(ra.Actions)
		claims.Access = append(claims.Access, ra)
	}
	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("failed to marshal claims: %s", err)
	}

	payload := fmt.Sprintf("%s%s%s", joseBase64UrlEncode(headerJSON), token.TokenSeparator, joseBase64UrlEncode(claimsJSON))

	sig, sigAlg2, err := c.Token.privateKey.Sign(strings.NewReader(payload), 0)
	if err != nil || sigAlg2 != sigAlg {
		return "", fmt.Errorf("failed to sign token: %s", err)
	}
	logrus.Infof("New token for %s %+v: %s", *ar, ar.Labels, claimsJSON)
	return fmt.Sprintf("%s%s%s", payload, token.TokenSeparator, joseBase64UrlEncode(sig)), nil
}

// Copy-pasted from libtrust where it is private.
func joseBase64UrlEncode(b []byte) string {
	return strings.TrimRight(base64.URLEncoding.EncodeToString(b), "=")
}


func loadCertAndKey(certFile, keyFile string) (pk libtrust.PublicKey, prk libtrust.PrivateKey, err error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return
	}
	x509Cert, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return
	}
	pk, err = libtrust.FromCryptoPublicKey(x509Cert.PublicKey)
	if err != nil {
		return
	}
	prk, err = libtrust.FromCryptoPrivateKey(cert.PrivateKey)
	return
}

func LoadConfig(cPath string) (*Config, error) {
	contents, err := ioutil.ReadFile(cPath)
	if err != nil {
		return nil, err
	}
	c := &Config{}

	if err := yaml.Unmarshal(contents, c); err != nil {
		return nil, fmt.Errorf("load config file err: %s", err)
	}
	//todo validate config file

	if c.Token.Key != "" && c.Token.Certificate != "" {
		c.Token.publicKey, c.Token.privateKey, err = loadCertAndKey(c.Token.Certificate, c.Token.Key)
		if err != nil {
			return nil, fmt.Errorf("failed to load cert and key file: %s", err)
		}
	}
	return c, nil
}

