package rolling_update

import (
	"io/ioutil"
	"fmt"
	"gopkg.in/yaml.v2"
	"github.com/BoxLinker/boxlinker-api/controller/manager"
	"k8s.io/client-go/kubernetes"
	"net/http"
	"github.com/gorilla/mux"
	"github.com/codegangsta/negroni"
	"github.com/Sirupsen/logrus"
	"github.com/gorilla/context"
	userModels "github.com/BoxLinker/boxlinker-api/controller/models/user"
	tAuth "github.com/BoxLinker/boxlinker-api/controller/middleware/auth_token"
	"github.com/BoxLinker/boxlinker-api"
)

type Api struct {
	config *Config
	manager manager.RollingUpdateManager
	clientSet *kubernetes.Clientset
}

type ApiConfig struct {
	Config *Config
	ControllerManager manager.RollingUpdateManager
	ClientSet *kubernetes.Clientset
}

type Config struct {
	Server struct {
		Addr string `yaml:"addr,omitempty"`
		Debug bool `yaml:"debug"`
	}    `yaml:"server,omitempty"`
	InCluster bool `yaml:"inCluster"`
	RegistryHost string `yaml:"registryHost"`
	DB struct{
		Host string `yaml:"host,omitempty"`
		Port int `yaml:"port,omitempty"`
		User string `yaml:"user,omitempty"`
		Password string `yaml:"password,omitempty"`
		Name string `yaml:"name,omitempty"`
	} `yaml:"db,omitempty"`
	AMQP struct{
		URI string `yaml:"uri,omitempty"`
		Exchange string `yaml:"exchange,omitempty"`
		ExchangeType string `yaml:"exchangeType,omitempty"`
		QueueName string `yaml:"queueName,omitempty"`
		BindingKey string `yaml:"bindingKey,omitempty"`
	} `yaml:"amqp,omitempty"`
	Auth struct{
		TokenAuthUrl string `yaml:"tokenAuthUrl,omitempty"`
		BasicAuthUrl string `yaml:"basicAuthUrl,omitempty"`
	} `yaml:"auth,omitempty"`
}

func NewApi(config ApiConfig) (*Api, error) {
	aApi := &Api{
		config: config.Config,
		manager: config.ControllerManager,
		clientSet: config.ClientSet,
	}
	return aApi, nil
}


func (a *Api) Run() error {
	cs := boxlinker.Cors
	// middleware
	apiAuthRequired := tAuth.NewAuthTokenRequired(a.config.Auth.TokenAuthUrl)

	globalMux := http.NewServeMux()

	serviceRouter := mux.NewRouter()
	serviceRouter.HandleFunc("/v1/rolling-update/auth/test", a.Test).Methods("GET")

	authRouter := negroni.New()
	authRouter.Use(negroni.HandlerFunc(apiAuthRequired.HandlerFuncWithNext))
	authRouter.UseHandler(serviceRouter)
	globalMux.Handle("/v1/rolling-update/auth/", authRouter)

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


func (a *Api) getUserInfo(r *http.Request) *userModels.User {
	us := r.Context().Value("user")
	if us == nil {
		return nil
	}
	ctx := us.(map[string]interface{})
	if ctx == nil || ctx["uid"] == nil {
		return nil
	}
	return &userModels.User{
		Id: ctx["uid"].(string),
		Name: ctx["username"].(string),
	}
}