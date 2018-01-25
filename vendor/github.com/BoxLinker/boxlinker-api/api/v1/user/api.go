package user

import (
	"net/http"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/BoxLinker/boxlinker-api/controller/manager"
	mAuth "github.com/BoxLinker/boxlinker-api/controller/middleware/auth"
	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/negroni"
)

type ApiOptions struct {
	Listen string
	Manager manager.UserManager
	SendEmailUri string
}

type Api struct {
	listen string
	manager manager.UserManager
	sendEmailUri string
}

func NewApi(config ApiOptions) *Api {
	return &Api{
		listen: config.Listen,
		manager: config.Manager,
		sendEmailUri: config.SendEmailUri,
	}
}
// get 	/v1/user/auth/token
// post /v1/user/auth/login
// post	/v1/user/auth/reg
// get	/v1/user/auth/confirm_email?confirm_token=
// put	/v1/user/account/list
// put	/v1/user/account/:id/changepassword
// get	/v1/user/account/:id
func (a *Api) Run() error {
	cs := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "DELETE", "PUT", "OPTIONS"},
		AllowedHeaders: []string{"Origin", "Content-Type", "Accept", "token", "X-Requested-With", "X-Access-Token"},
	})
	// middleware
	apiAuthRequired := mAuth.NewAuthRequired(a.manager)

	globalMux := http.NewServeMux()

	loginRegRouter := mux.NewRouter()
	loginRegRouter.HandleFunc("/v1/user/auth/basicAuth", a.BasicAuth).Methods("GET")
	loginRegRouter.HandleFunc("/v1/user/auth/login", a.Login).Methods("POST")
	loginRegRouter.HandleFunc("/v1/user/auth/reg", a.Reg).Methods("POST")
	loginRegRouter.HandleFunc("/v1/user/auth/confirm_email", a.ConfirmEmail).Methods("GET")
	globalMux.Handle("/v1/user/auth/", loginRegRouter)

	accountRouter := mux.NewRouter()
	accountRouter.HandleFunc("/v1/user/account/authToken", a.AuthToken).Methods("GET")
	accountRouter.HandleFunc("/v1/user/account/list", a.GetUsers).Methods("GET")
	accountRouter.HandleFunc("/v1/user/account/changepassword", a.ChangePassword).Methods("PUT")
	accountRouter.HandleFunc("/v1/user/account/userinfo", a.GetUser).Methods("GET")
	accountAuthRouter := negroni.New()
	accountAuthRouter.Use(negroni.HandlerFunc(apiAuthRequired.HandlerFuncWithNext))
	accountAuthRouter.UseHandler(accountRouter)
	globalMux.Handle("/v1/user/account/", accountAuthRouter)

	s := &http.Server{
		Addr:    a.listen,
		Handler: context.ClearHandler(cs.Handler(globalMux)),
	}

	log.Infof("Server listen on: %s", a.listen)

	return s.ListenAndServe()
}
