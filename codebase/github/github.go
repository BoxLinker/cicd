package github

import (
	"github.com/BoxLinker/cicd/codebase"
	"net/url"
	"net"
	"strings"
	"net/http"
	"github.com/BoxLinker/cicd/models"
	"golang.org/x/oauth2"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/cabernety/gopkg/httplib"
	"golang.org/x/net/context"
	"crypto/tls"
	"github.com/google/go-github/github"
)

const (
	defaultURL = "https://github.com"     // Default GitHub URL
	defaultAPI = "https://api.github.com" // Default GitHub API URL
)

// Opts defines configuration options.
type Opts struct {
	HomeHost 	string
	URL         string   // GitHub server url.
	Context     string   // Context to display in status check
	Client      string   // GitHub oauth client id.
	Secret      string   // GitHub oauth client secret.
	Scopes      []string // GitHub oauth scopes
	Username    string   // Optional machine account username.
	Password    string   // Optional machine account password.
	PrivateMode bool     // GitHub is running in private mode.
	SkipVerify  bool     // Skip ssl verification.
	MergeRef    bool     // Clone pull requests using the merge ref.
}

func New(opts Opts) (codebase.CodeBase, error) {
	url_, err := url.Parse(opts.URL)
	if err != nil {
		return nil, err
	}
	host, _, err := net.SplitHostPort(url_.Host)
	if err == nil {
		url_.Host = host
	}
	cb := &client{
		HomeHost: 	 opts.HomeHost,
		API:         defaultAPI,
		URL:         defaultURL,
		Context:     opts.Context,
		Client:      opts.Client,
		Secret:      opts.Secret,
		Scopes:      opts.Scopes,
		PrivateMode: opts.PrivateMode,
		SkipVerify:  opts.SkipVerify,
		MergeRef:    opts.MergeRef,
		Machine:     url_.Host,
		Username:    opts.Username,
		Password:    opts.Password,
	}

	if opts.URL != defaultAPI {
		cb.URL = strings.TrimSuffix(opts.URL, "/")
		cb.API = cb.URL + "/api/v3/"
	}

	logrus.Debugf("client: (%+v)", cb)


	// Hack to enable oauth2 access in older GHE
	//oauth2.RegisterBrokenAuthHeaderProvider(cb.URL)
	return cb, nil
}

type client struct {
	HomeHost 	string
	URL         string
	Context     string
	API         string
	Client      string
	Secret      string
	Scopes      []string
	Machine     string
	Username    string
	Password    string
	PrivateMode bool
	SkipVerify  bool
	MergeRef    bool
}

func (c *client) Authorize(w http.ResponseWriter, r *http.Request, stateParam string) (*models.CodeBaseUser, error) {

	oauth2Config := &oauth2.Config{
		ClientID: c.Client,
		ClientSecret: c.Secret,
		Scopes: c.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL: fmt.Sprintf("%s/login/oauth/authorize", c.URL),
			TokenURL: fmt.Sprintf("%s/login/oauth/access_token", c.URL),
		},
		RedirectURL: fmt.Sprintf("%s://%s/v1/cicd/authorize", httplib.GetScheme(r), httplib.GetHost(r)),
	}

	if err := r.FormValue("error"); err != "" {
		return nil, &codebase.AuthError{
			Err: err,
			Description: r.FormValue("error_description"),
			URI:         r.FormValue("error_uri"),
		}
	}

	code := r.FormValue("code")
	state := r.FormValue("state")
	if len(code) == 0 || len(state) == 0 {
		http.Redirect(w, r, oauth2Config.AuthCodeURL(stateParam), http.StatusSeeOther)
		return nil, nil
	}

	logrus.Debugf("github authorize \ncode(%s) \nstate(%s)", code, state)
	token, err := oauth2Config.Exchange(c.newContext(), code)
	if err != nil {
		return nil, err
	}

	client := c.newClientToken(token.AccessToken)
	user, _, err := client.Users.Get("")
	if err != nil {
		return nil, err
	}
	return &models.CodeBaseUser{
		Login: *user.Login,
		Token: state,
		AccessToken: token.AccessToken,
		Kind: "github",
	}, nil
}

func (c *client) newContext() context.Context {
	if !c.SkipVerify {
		return oauth2.NoContext
	}
	return context.WithValue(nil, oauth2.HTTPClient, &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	})
}

func (c *client) newClientToken(token string) *github.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(oauth2.NoContext, ts)
	if c.SkipVerify {
		tc.Transport.(*oauth2.Transport).Base = &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
	}
	githubClient := github.NewClient(tc)
	githubClient.BaseURL, _ = url.Parse(c.API)
	return githubClient
}