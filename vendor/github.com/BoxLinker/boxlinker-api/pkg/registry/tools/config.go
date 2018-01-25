package tools

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
)


type Config struct {
	Server struct {
		Addr string `yaml:"addr,omitempty"`
	}    `yaml:"server,omitempty"`
	Token struct {
		Issuer      string    `yaml:"issuer,omitempty"`
		Expiration  int64    `yaml:"expiration,omitempty"`
		Certificate string    `yaml:"certificate,omitempty"`
		Key         string    `yaml:"key,omitempty"`

		privateKey libtrust.PrivateKey
		publicKey  libtrust.PublicKey
	} `yaml:"token,omitempty"`
	ACL authz.ACL `yaml:"acl,omitempty"`
}

func (c *Config) GenerateToken(issuer, subject, audience string, expiration int64, t, name string,  actions []string) (string, error) {
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
		Issuer:     issuer,
		Subject:    subject,
		Audience:   audience,
		NotBefore:  now - 10,
		IssuedAt:   now,
		Expiration: now + expiration,
		JWTID:      fmt.Sprintf("%d", rand.Int63()),
		Access:     []*token.ResourceActions{},
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
