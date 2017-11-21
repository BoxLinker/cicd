package github

import (
	"github.com/BoxLinker/cicd/codebase"
	"net/url"
	"net"
	"strings"
	"golang.org/x/oauth2"
	"net/http"
	"github.com/BoxLinker/cicd/models"
)

const (
	defaultURL = "https://github.com"     // Default GitHub URL
	defaultAPI = "https://api.github.com" // Default GitHub API URL
)

// Opts defines configuration options.
type Opts struct {
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

	// Hack to enable oauth2 access in older GHE
	oauth2.RegisterBrokenAuthHeaderProvider(cb.URL)
	return cb, nil
}

type client struct {
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

func (c *client) Authorize(w http.ResponseWriter, r *http.Request) (*models.User, error) {
	return &models.User{
		Token: "",
	}, nil
}