package registry_watcher

import (
	"io/ioutil"
	"fmt"
	"gopkg.in/yaml.v2"
	"github.com/BoxLinker/boxlinker-api/controller/manager"
	"net/http"
	"github.com/gorilla/mux"
	"github.com/codegangsta/negroni"
	"github.com/Sirupsen/logrus"
	"github.com/gorilla/context"
	tAuth "github.com/BoxLinker/boxlinker-api/controller/middleware/auth_amqp"
	"github.com/BoxLinker/boxlinker-api"
)

type Api struct {
	config *Config
	manager manager.RegistryWatcherManager
}

type ApiConfig struct {
	Config *Config
	ControllerManager manager.RegistryWatcherManager
}

type Config struct {
	Server struct {
		Addr string `yaml:"addr,omitempty"`
		Debug bool `yaml:"debug"`
	}    `yaml:"server,omitempty"`
	Auth struct{
		RegistryAuthorization string `yaml:"registryAuthorization,omitempty"`
	} `yaml:"auth,omitempty"`
	Amqp struct{
		Host string `yaml:"host,omitempty"`
		Exchange string `yaml:"exchange,omitempty"`
		ExchangeType string `yaml:"exchangeType,omitempty"`
		Reliable bool `yaml:"reliable"`
	}
}

func NewApi(config ApiConfig) (*Api, error) {
	aApi := &Api{
		config: config.Config,
		manager: config.ControllerManager,
	}
	return aApi, nil
}

func (a *Api) Run() error {
	cs := boxlinker.Cors
	// middleware
	apiAuthRequired := tAuth.NewAuthAmqpRequired(a.config.Auth.RegistryAuthorization)

	globalMux := http.NewServeMux()

	serviceRouter := mux.NewRouter()
	serviceRouter.HandleFunc("/v1/registry-watcher/event", a.RegistryEvent).Methods("POST")

	authRouter := negroni.New()
	authRouter.Use(negroni.HandlerFunc(apiAuthRequired.HandlerFuncWithNext))
	authRouter.UseHandler(serviceRouter)
	globalMux.Handle("/v1/registry-watcher/", authRouter)

	s := &http.Server{
		Addr: a.config.Server.Addr,
		Handler: context.ClearHandler(cs.Handler(globalMux)),
	}

	logrus.Infof("Server run: %s", a.config.Server.Addr)

	return s.ListenAndServe()
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

	return c, nil
}