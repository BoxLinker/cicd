package manager

import (
	"github.com/go-xorm/xorm"
	"k8s.io/client-go/kubernetes"
)

type DefaultManager struct {
	ClientSet *kubernetes.Clientset
	DBEngine *xorm.Engine
}

