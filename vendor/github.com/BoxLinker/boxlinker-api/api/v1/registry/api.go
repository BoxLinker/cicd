package registry

import (
	"net/http"
	"github.com/rs/cors"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"fmt"
	"github.com/Sirupsen/logrus"
	"encoding/json"
	"github.com/BoxLinker/boxlinker-api/controller/manager"
	"github.com/BoxLinker/boxlinker-api/pkg/registry/authn"
	tAuth "github.com/BoxLinker/boxlinker-api/controller/middleware/auth_token"
	"strings"
	"sort"
	"github.com/BoxLinker/boxlinker-api/pkg/registry/authz"
	"time"
	"net"
	"github.com/codegangsta/negroni"
)

type Api struct {
	Listen string
	Manager manager.RegistryManager
	Authenticator authn.Authenticator
	Authorizers []authz.Authorizer
	Config *Config
}
type ApiConfig struct {
	Listen string
	Manager manager.RegistryManager
	BasicAuthURL string
	ConfigFilePath string
	Config *Config
}
func NewApi(ac *ApiConfig) (*Api, error) {

	config := ac.Config

	a := &Api{
		Listen: ac.Listen,
		Manager: ac.Manager,
		Config: config,
		Authorizers: []authz.Authorizer{},
	}
	// authenticator
	a.Authenticator = &authn.DefaultAuthenticator{
		BasicAuthURL: a.Config.Auth.BasicAuthUrl,
	}

	// boxlinker ucenter
	//a.Authenticator = &authn.BoxlinkerUCenterAuthenticator{
	//	BasicAuthURL: a.Config.Auth.BasicAuthUrl,
	//}

	//if err := ac.Manager.SaveACL(&registryModels.ACL{
	//	Account: "*",
	//	Name: "library/*",
	//	Actions: "*",
	//}); err != nil {
	//	return nil, err
	//}

	// authorizes
	if config.ACL != nil {
		staticAuthorizer, err := authz.NewACLAuthorizer(config.ACL)
		if err != nil {
			return nil, err
		}
		a.Authorizers = append(a.Authorizers, staticAuthorizer)
	}

	mysqlAuthorizer, err := authz.NewACLMysqlAuthorizer(authz.ACLMysqlConfig{
		Manager: ac.Manager,
		CacheTTL: time.Second * 60,
	})
	if err != nil {
		return nil, err
	}
	a.Authorizers = append(a.Authorizers, mysqlAuthorizer)

	return a, nil
}


type RegistryCallback struct {
	Events [] struct{
		Id 			string 	`json:"id"`
		Timestamp 	string 	`json:"timestamp"`
		Action 		string 	`json:"action"`
		Target 		struct{
			MediaType 		string 		`json:"mediaType"`
			Size 			int64 		`json:"size"`
			Digest 			string 		`json:"digest"`
			Length 			int64 		`json:"length"`
			Repository 		string 		`json:"repository"`
			Url 			string 		`json:"url"`
			Tag 			string 		`json:"tag"`
		} `json:"target"`
		Request 	struct{
			Id 		string 		`json:"id"`
			Addr 	string 		`json:"addr"`
			Host	string 		`json:"host"`
			Method 	string 		`json:"method"`
			UserAgent string 	`json:"useragent"`
		} 	`json:"request"`
		Source 		struct{
			Addr 	string 	`json:"addr"`
			InstanceID string `json:"instanceID"`
		} 	`json:"source"`
	}	`json:"events"`
}

type authScope struct {
	Type string
	Name string
	Actions []string
}

type authzResult struct {
	scope authScope
	authorizedActions []string
}

type authRequest struct {
	RemoteConnAddr string
	RemoteAddr     string
	RemoteIP       net.IP
	User           string
	Password       authn.PasswordString
	Account        string
	Service        string
	Scopes         []authScope
	Labels         authn.Labels
}


func prepareRequest(r *http.Request) (*authRequest, error) {
	ar := &authRequest{}
	user, pass, ok := r.BasicAuth()
	if ok {
		ar.User = user
		ar.Password = authn.PasswordString(pass)
	}
	ar.Account = r.FormValue("account")
	if ar.Account == "" {
		ar.Account = ar.User
	} else if ar.Account != "" && ar.Account != ar.User {
		return nil, fmt.Errorf("user and account are not same (%q and %q)", ar.User, ar.Account)
	}

	ar.Service = r.FormValue("service")

	if err := r.ParseForm(); err != nil {
		return nil, fmt.Errorf("invalid form value: %s", err)
	}
	if r.FormValue("scope") != "" {
		for _, scopeStr := range r.Form["scope"] {
			var scope authScope
			parts := strings.Split(scopeStr, ":")
			switch len(parts) {
			case 3:
				scope = authScope{
					Type: parts[0],
					Name: parts[1],
					Actions: strings.Split(parts[2], ","),
				}
			case 4:
				scope = authScope{
					Type: parts[0],
					Name: parts[1] + ":" + parts[2],
					Actions: strings.Split(parts[3], ","),
				}
			default:
				return nil, fmt.Errorf("invalid scope (%q)", scopeStr)
			}
			sort.Strings(scope.Actions)
			ar.Scopes = append(ar.Scopes, scope)
		}
	}

	return ar, nil
}

func (a *Api) authorizeScope(ai *authz.AuthRequestInfo) ([]string, error) {

	for i, authorizer := range a.Authorizers {
		result, err := authorizer.Authorize(ai)
		logrus.Infof("Authz %s %s -> %s, %v", authorizer.Name(), *ai, result, err)
		if err != nil {
			if err == authz.NoMatch {
				continue
			}
			err = fmt.Errorf("authz #%d returned error: %s", i+1, err)
			logrus.Errorf("%s: %s", *ai, err)
			return nil, err
		}
		return result, nil
	}
	logrus.Warningf("%s did not match any authz rule", *ai)
	return nil, nil
}

func (a *Api) authorize(ar *authRequest) ([]authzResult, error) {
	ares := []authzResult{}
	for _, scope := range ar.Scopes {
		ai := &authz.AuthRequestInfo{
			Account: ar.Account,
			Type: scope.Type,
			Name: scope.Name,
			Service: ar.Service,
			Actions: scope.Actions,
		}
		actions, err := a.authorizeScope(ai)
		if err != nil {
			return nil, err
		}
		ares = append(ares, authzResult{scope: scope, authorizedActions: actions})
	}
	return ares, nil
}

// POST 	/v1/registry/auth
func (a *Api) DoRegistryAuth(w http.ResponseWriter, r *http.Request){
	ar, err := prepareRequest(r)
	ares := []authzResult{}
	if err != nil {
		http.Error(w, fmt.Sprintf("Bad Request: %s", err), http.StatusBadRequest)
		return
	}
	logrus.Debugf("Auth request: %+v", ar)
	authResult, _, err := a.Authenticator.Authenticate(ar.User, ar.Password)
	if err != nil {
		http.Error(w, fmt.Sprintf("Authentication failed (%s)", err), http.StatusUnauthorized)
		return
	}
	if !authResult {
		logrus.Warnf("Auth failed (user:%s)", ar.User)
		w.Header()["WWW-Authenticate"] = []string{fmt.Sprintf(`Basic realm="%s"`, a.Config.Token.Issuer)}
		http.Error(w, "Auth Failed.", http.StatusUnauthorized)
		return
	}
	// authorize based on scopes
	if len(ar.Scopes) > 0 {
		ares, err = a.authorize(ar)
		if err != nil {
			http.Error(w, fmt.Sprintf("Authorization failed (%s)", err), http.StatusInternalServerError)
			return
		}
	} else {
		// Authentication-only request ("docker login"), pass through.
	}

	logrus.Debugf("Ares: (%+v)", ares)
	//t := &a.Config.Token
	token, err := a.Config.GenerateToken(ar, ares)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to generate token (%s)", err), http.StatusInternalServerError)
		return
	}

	result, _ := json.Marshal(&map[string]string{"token": token})
	logrus.Debugf("generate token: %s", string(result))
	w.Header().Set("Content-Type", "application/json")
	w.Write(result)
}


func (a * Api) Run() error {
	cs := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "DELETE", "PUT", "OPTIONS"},
		AllowedHeaders: []string{"Origin", "Content-Type", "Accept", "token", "X-Requested-With", "X-Access-Token"},
	})
	// middleware
	apiAuthRequired := tAuth.NewAuthTokenRequired(a.Config.Auth.TokenAuthUrl)
	// boxlinker token auth
	//apiAuthRequired := tAuth.NewBoxlinkerAuthTokenRequired(a.Config.Auth.TokenAuthUrl)

	globalMux := http.NewServeMux()

	eventRouter := mux.NewRouter()
	eventRouter.HandleFunc("/v1/registry/callback/auth", a.DoRegistryAuth).Methods("GET")
	eventRouter.HandleFunc("/v1/registry/callback/event", a.RegistryEvent).Methods("POST")
	globalMux.Handle("/v1/registry/callback/", eventRouter)

	imageRouter := mux.NewRouter()
	imageRouter.HandleFunc("/v1/registry/auth/image/list", a.QueryImages).Methods("GET")
	imageRouter.HandleFunc("/v1/registry/auth/image/exists", a.ImageExists).Methods("GET")
	imageRouter.HandleFunc("/v1/registry/auth/image/new", a.SaveImage).Methods("POST")
	imageRouter.HandleFunc("/v1/registry/auth/image/{id}", a.GetImage).Methods("GET")
	imageRouter.HandleFunc("/v1/registry/auth/image/{id}/description", a.UpdateImageDescription).Methods("PUT")
	imageRouter.HandleFunc("/v1/registry/auth/image/{id}/html_doc", a.UpdateImageHtmlDoc).Methods("PUT")
	imageRouter.HandleFunc("/v1/registry/auth/image/{id}/privilege", a.UpdateImagePrivilege).Methods("PUT")
	imageRouter.HandleFunc("/v1/registry/auth/image/{id}", a.DeleteImage).Methods("DELETE")
	imageAuthRouter := negroni.New()
	imageAuthRouter.Use(negroni.HandlerFunc(apiAuthRequired.HandlerFuncWithNext))
	imageAuthRouter.UseHandler(imageRouter)
	globalMux.Handle("/v1/registry/auth/", imageAuthRouter)

	imagePubRouter := mux.NewRouter()
	imagePubRouter.HandleFunc("/v1/registry/pub/image/list", a.QueryPubImages).Methods("GET")
	globalMux.Handle("/v1/registry/pub/", imagePubRouter)


	s := &http.Server{
		Addr: a.Listen,
		Handler: context.ClearHandler(cs.Handler(globalMux)),
	}

	logrus.Infof("Server run: %s", a.Listen)

	return s.ListenAndServe()
}