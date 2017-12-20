package github

import (
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
	"github.com/BoxLinker/cicd/scm"
	"regexp"
	"strconv"
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

func New(opts Opts) (scm.SCM, error) {
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

func (c *client) Status(u *models.User, r *models.Repo, b *models.Build, link string) error {
	client := c.newClientToken(u.AccessToken)
	switch b.Event {
	case "deployment":
		return deploymentStatus(client, r, b, link)
	default:
		return repoStatus(client, r, b, link, c.Context)
	}
}

func repoStatus(client *github.Client, r *models.Repo, b *models.Build, link, ctx string) error {
	context := ctx
	switch b.Event {
	case models.EventPull:
		context += "/pr"
	default:
		if len(b.Event) > 0 {
			context += "/" + b.Event
		}
	}

	data := github.RepoStatus{
		Context: 		github.String(context),
		State: 			github.String(convertStatus(b.Status)),
		Description: 	github.String(convertDesc(b.Status)),
		TargetURL: 		github.String(link),
	}
	logrus.Debugf("SCM github CreateStatus owner(%s) repo(%s) commit(%s) data(%+v)", r.Owner, r.Name, b.Commit, &data)
	_, _, err := client.Repositories.CreateStatus(r.Owner, r.Name, b.Commit, &data)
	return err
}

var reDeploy = regexp.MustCompile(".+/deployments/(\\d+)")

func deploymentStatus(client *github.Client, r *models.Repo, b *models.Build, link string) error {
	matches := reDeploy.FindStringSubmatch(b.Link)
	if len(matches) != 2 {
		return nil
	}
	id, _ := strconv.Atoi(matches[1])

	data := github.DeploymentStatusRequest{
		State: 		github.String(convertStatus(b.Status)),
		Description: github.String(convertDesc(b.Status)),
		TargetURL: 	github.String(link),
	}

	_, _, err := client.Repositories.CreateDeploymentStatus(r.Owner, r.Name, id, &data)
	return err
}

func (c *client) File(u *models.User, r *models.Repo, b *models.Build, f string) ([]byte, error) {
	return c.FileRef(u, r, b.Commit, f)
}

func (c *client) FileRef(u *models.User, r *models.Repo, ref, f string) ([]byte, error) {
	client := c.newClientToken(u.AccessToken)
	opts := new(github.RepositoryContentGetOptions)
	opts.Ref = ref
	data, _, _, err := client.Repositories.GetContents(r.Owner, r.Name, f, opts)
	if err != nil {
		return nil, err
	}
	return data.Decode()
}

func (c *client) Repo(u *models.User, owner, repoName string) (*models.Repo, error) {
	client := c.newClientToken(u.AccessToken)
	repo, _, err := client.Repositories.Get(owner, repoName)
	if err != nil {
		return nil, err
	}
	return convertRepo(repo, u), nil
}

func (c *client) Repos(u *models.User) ([]*models.Repo, error) {
	client := c.newClientToken(u.AccessToken)
	opts := new(github.RepositoryListOptions)
	opts.PerPage = 100
	opts.Page = 1

	var repos []*models.Repo
	if opts.Page > 0 {
		list, resp, err := client.Repositories.List("", opts)
		if err != nil {
			return nil, err
		}
		repos = append(repos, convertRepoList(list, u)...)
		opts.Page = resp.NextPage
	}
	return repos, nil
}

func (c *client) Authorize(w http.ResponseWriter, r *http.Request, stateParam string) (*models.User, error) {

	oauth2Config := &oauth2.Config{
		ClientID: c.Client,
		ClientSecret: c.Secret,
		Scopes: c.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL: fmt.Sprintf("%s/login/oauth/authorize", c.URL),
			TokenURL: fmt.Sprintf("%s/login/oauth/access_token", c.URL),
		},
		RedirectURL: fmt.Sprintf("%s://%s/v1/cicd/authorize/github", httplib.GetScheme(r), httplib.GetHost(r)),
	}

	if err := r.FormValue("error"); err != "" {
		return nil, &scm.AuthError{
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
	return &models.User{
		Login: *user.Login,
		Email: *user.Email,
		Token: state,
		AccessToken: token.AccessToken,
		SCM: "github",
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

func (c *client) Hook(r *http.Request) (*models.Repo, *models.Build, error) {
	return parseHook(r, c.MergeRef)
}

func (c *client) Activate(u *models.User, r *models.Repo, link string) error {
	if err := c.Deactivate(u, r, link); err != nil {
		return err
	}
	client := c.newClientToken(u.AccessToken)
	hook := &github.Hook{
		Name: github.String("web"),
		Events: []string{
			"push",
			"pull_request",
			"deployment",
		},
		Config: map[string]interface{}{
			"url": 		link,
			"content_type": "form",
		},
	}
	_, _, err := client.Repositories.CreateHook(r.Owner, r.Name, hook)
	return err
}

func (c *client) Deactivate(u *models.User, r *models.Repo, link string) error {
	client := c.newClientToken(u.AccessToken)
	hooks, _, err := client.Repositories.ListHooks(r.Owner, r.Name, nil)
	if err != nil {
		return err
	}
	match := matchHooks(hooks, link)
	if match == nil {
		return nil
	}
	_, err = client.Repositories.DeleteHook(r.Owner, r.Name, *match.ID)
	return err
}

func matchHooks(hooks []github.Hook, rawurl string) *github.Hook {
	link, err := url.Parse(rawurl)
	if err != nil {
		return nil
	}
	for _, hook := range hooks {
		if hook.ID == nil {
			continue
		}
		v, ok := hook.Config["url"]
		if !ok {
			continue
		}
		s, ok := v.(string)
		if !ok {
			continue
		}
		hookurl, err := url.Parse(s)
		if err == nil && hookurl.Host == link.Host {
			return &hook
		}
	}
	return nil
}