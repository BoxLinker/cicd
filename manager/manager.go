package manager

import (
	"k8s.io/client-go/kubernetes"
	"fmt"
	"flag"
	"path/filepath"
	"k8s.io/client-go/tools/clientcmd"
	"github.com/Sirupsen/logrus"
	"k8s.io/client-go/rest"
	"os"
	"github.com/BoxLinker/cicd/store"
	"github.com/BoxLinker/cicd/scm"
	"github.com/BoxLinker/cicd/models"
	"github.com/BoxLinker/cicd/logging"
	"github.com/BoxLinker/cicd/queue"
	"github.com/BoxLinker/cicd/pubsub"
)

type DefaultManager struct {
	dataStore store.Store
	clientSet *kubernetes.Clientset
	scmMap map[string]scm.SCM
	logs logging.Log
	queue queue.Queue
	pubsub pubsub.Publisher
}

type Options struct {
	KubernetesInCluster bool
	Driver string
	DataSource string
	Store store.Store
	Logs logging.Log
	Queue queue.Queue
	Pubsub pubsub.Publisher
	SCMMap map[string]scm.SCM
}

func New(opts *Options) (*DefaultManager, error) {
	var (
		clientSet *kubernetes.Clientset
		err error
		config *rest.Config
	)
	if opts.KubernetesInCluster {
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
		clientSet, err = kubernetes.NewForConfig(config)
		if err != nil {
			return nil, fmt.Errorf("connect to incluster k8s error: %v", err)
		}
	} else {
		var kubeconfig *string
		if home := homeDir(); home != "" {
			kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
		} else {
			kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
		}
		flag.Parse()

		// use the current context in kubeconfig
		config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
		if err != nil {
			panic(err.Error())
		}
		logrus.Infof("kubeconfig (%+v)", config)
		// create the clientset
		clientSet, err = kubernetes.NewForConfig(config)
		if err != nil {
			return nil, fmt.Errorf("connect to k8s error: %v", err)
		}
	}

	return &DefaultManager{
		dataStore: opts.Store,
		clientSet: clientSet,
		scmMap: opts.SCMMap,
		logs: opts.Logs,
		queue: opts.Queue,
		pubsub: opts.Pubsub,
	}, nil

}

func (m *DefaultManager) GetSCM(scm string) scm.SCM {
	return m.scmMap[scm]
}

func (m *DefaultManager) ConfigStore() models.ConfigStore {
	return m.dataStore
}

func (m *DefaultManager) Logs() logging.Log {
	return m.logs
}
func (m *DefaultManager) Pubsub() pubsub.Publisher {
	return m.pubsub
}

func (m *DefaultManager) Queue() queue.Queue {
	return m.queue
}

func (m *DefaultManager) Store() store.Store {
	return m.dataStore
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}