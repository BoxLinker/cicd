package application

import (
	"io/ioutil"
	"gopkg.in/yaml.v2"
	"fmt"
	"github.com/BoxLinker/boxlinker-api/controller/manager"
	tAuth "github.com/BoxLinker/boxlinker-api/controller/middleware/auth_token"
	userModels "github.com/BoxLinker/boxlinker-api/controller/models/user"
	appModels "github.com/BoxLinker/boxlinker-api/controller/models/application"
	"net/http"
	"github.com/Sirupsen/logrus"
	"github.com/BoxLinker/boxlinker-api"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/codegangsta/negroni"
	"k8s.io/client-go/kubernetes"
)

type Api struct {
	config *Config
	manager manager.ApplicationManager
	clientSet *kubernetes.Clientset
}

type ApiConfig struct {
	Config *Config
	ControllerManager manager.ApplicationManager
	ClientSet *kubernetes.Clientset
}

func NewApi(config ApiConfig) (*Api, error) {
	aApi := &Api{
		config: config.Config,
		manager: config.ControllerManager,
		clientSet: config.ClientSet,
	}
	// check PodConfigure
	if err := aApi.checkPodConfigure(); err != nil {
		return nil, err
	}
	return aApi, nil
}

type Config struct {
	Server struct {
		Addr string `yaml:"addr,omitempty"`
		Debug bool `yaml:"debug"`
	}    `yaml:"server,omitempty"`
	InCluster bool `yaml:"inCluster"`
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
	PodConfigure []struct{
		Memory string `yaml:"memory,omitempty"`
		CPU string `yaml:"cpu,omitempty"`
	} `yaml:"podConfigure,omitempty"`
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

func (a *Api) checkPodConfigure() error {
	configs := a.config.PodConfigure
	podConfigures := make([]*appModels.PodConfigure, 0)
	for _, c := range configs {
		podConfigures = append(podConfigures, &appModels.PodConfigure{
			Memory: c.Memory,
			CPU: c.CPU,
		})
	}
	nums, err := a.manager.SyncPodConfigure(podConfigures)
	if err != nil {
		return err
	}
	logrus.Debugf("Sync PodConfigures %d", nums)
	return nil
}

func (a *Api) Run() error {
	cs := boxlinker.Cors
	// middleware
	apiAuthRequired := tAuth.NewAuthTokenRequired(a.config.Auth.TokenAuthUrl)

	globalMux := http.NewServeMux()

	serviceRouter := mux.NewRouter()
	serviceRouter.HandleFunc("/v1/application/auth/service", a.CreateService).Methods("POST")
	serviceRouter.HandleFunc("/v1/application/auth/service", a.QueryService).Methods("GET")
	serviceRouter.HandleFunc("/v1/application/auth/service/{name}", a.UpdateService).Methods("PUT")
	serviceRouter.HandleFunc("/v1/application/auth/service/{name}", a.DeleteService).Methods("DELETE")
	serviceRouter.HandleFunc("/v1/application/auth/service/{name}/exists", a.IsServiceExist).Methods("GET")

	//volumeRouter := mux.NewRouter()
	serviceRouter.HandleFunc("/v1/application/auth/volume", a.CreateVolume).Methods("POST")
	serviceRouter.HandleFunc("/v1/application/auth/volume", a.QueryVolume).Methods("GET")
	serviceRouter.HandleFunc("/v1/application/auth/volume/{name}", a.DeleteVolume).Methods("DELETE")

	authRouter := negroni.New()
	authRouter.Use(negroni.HandlerFunc(apiAuthRequired.HandlerFuncWithNext))
	authRouter.UseHandler(serviceRouter)
	globalMux.Handle("/v1/application/auth/", authRouter)

	s := &http.Server{
		Addr: a.config.Server.Addr,
		Handler: context.ClearHandler(cs.Handler(globalMux)),
	}

	logrus.Infof("Server run: %s", a.config.Server.Addr)

	return s.ListenAndServe()
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

func int32Ptr(i int32) *int32 { return &i }