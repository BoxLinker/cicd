package manager

import (
	"github.com/go-xorm/xorm"
	appModels "github.com/BoxLinker/boxlinker-api/controller/models/application"
	"github.com/Sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	appsv1beta1 "k8s.io/api/apps/v1beta1"
	extv1beta1 "k8s.io/api/extensions/v1beta1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"fmt"
	"github.com/BoxLinker/boxlinker-api/controller/models"
	"k8s.io/apimachinery/pkg/api/resource"
	"github.com/BoxLinker/boxlinker-api"
	"encoding/json"
)

type ApplicationManager interface {
	Manager
	SyncPodConfigure(pcs []*appModels.PodConfigure) (int, error)
	GetServiceByName(namespace, svcName string) (bool, error, *apiv1.Service, *extv1beta1.Ingress, *appsv1beta1.Deployment)

	GetVolumeByName(namespace, name string) (pvc *apiv1.PersistentVolumeClaim, err error)
	DeleteVolume(namespace, name string) error
	CreateVolume(namespace string, volume *models.Volume) (*apiv1.PersistentVolumeClaim, error)
	QueryVolume(namespace string, pc boxlinker.PageConfig) ([]apiv1.PersistentVolumeClaim, error)
}

type DefaultApplicationManager struct {
	DefaultManager
	engine *xorm.Engine
	clientSet *kubernetes.Clientset
}

func (m *DefaultApplicationManager) GetServiceByName(namespace, svcName string) (bool, error, *apiv1.Service, *extv1beta1.Ingress, *appsv1beta1.Deployment) {
	var (
		err error
		svc *apiv1.Service
		ing *extv1beta1.Ingress
		deploy *appsv1beta1.Deployment
	)
	if svc, err = m.clientSet.CoreV1().Services(namespace).Get(svcName, metav1.GetOptions{}); err != nil {
		return false, fmt.Errorf("Service %s/%s not found: %v", namespace, svcName, err), nil, nil, nil
	}
	if ing, err = m.clientSet.ExtensionsV1beta1().Ingresses(namespace).Get(svcName, metav1.GetOptions{}); err != nil {
		return false, fmt.Errorf("Ingress %s/%s not found: %v", namespace, svcName, err), nil, nil, nil
	}
	if deploy, err = m.clientSet.AppsV1beta1().Deployments(namespace).Get(svcName, metav1.GetOptions{}); err != nil {
		return false, fmt.Errorf("Deployment %s/%s not found: %v", namespace, svcName, err), nil, nil, nil
	}
	return true, nil, svc, ing, deploy
}

func (m *DefaultApplicationManager) QueryVolume(namespace string, pc boxlinker.PageConfig) ([]apiv1.PersistentVolumeClaim, error) {
	claims, err := m.clientSet.CoreV1().PersistentVolumeClaims(namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	b, _ := json.MarshalIndent(claims, "", "\t")

	logrus.Debugf("=====---=====")
	logrus.Debugf("%s", b)
	logrus.Debugf("=====---=====")
	results := make([]apiv1.PersistentVolumeClaim, 0)
	var start, end int
	l := len(claims.Items)
	if pc.Offset() >= l {
		start = 0
		end = l
	} else {
		start = pc.Offset()
		if pc.Offset() + pc.Limit() >= l {
			end = l
		} else {
			end = pc.Offset() + pc.Limit()
		}
	}
	logrus.Debugf("start %d, end %d", start, end)
	for _, item := range claims.Items[start:end] {
		logrus.Debugf("==> %s, %s, %s", item.Namespace, item.ObjectMeta.Name, item.Name)
		results = append(results, item)
	}
	return results, nil
}
func (m *DefaultApplicationManager) CreateVolume(namespace string, volume *models.Volume) (*apiv1.PersistentVolumeClaim, error) {
	size, err := resource.ParseQuantity(volume.Size)
	if err != nil {
		return nil, err
	}
	pvc := &apiv1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name: volume.Name,
			Annotations: map[string]string{
				"volume.beta.kubernetes.io/storage-class": "rbd",
			},
		},
		Spec: apiv1.PersistentVolumeClaimSpec{
			AccessModes: []apiv1.PersistentVolumeAccessMode{apiv1.ReadWriteOnce},
			Resources: apiv1.ResourceRequirements{
				Requests: apiv1.ResourceList{
					apiv1.ResourceStorage: size,
				},
			},
		},
	}
	claim, err := m.clientSet.CoreV1().PersistentVolumeClaims(namespace).Create(pvc)
	if err != nil {
		return nil, err
	}
	return claim, nil
}
func (m *DefaultApplicationManager) DeleteVolume(namespace, name string) error {
	return m.clientSet.CoreV1().PersistentVolumeClaims(namespace).Delete(name, &metav1.DeleteOptions{})
}
func (m *DefaultApplicationManager) GetVolumeByName(namespace, name string) (pvc *apiv1.PersistentVolumeClaim, err error) {
	pvc, err = m.clientSet.CoreV1().PersistentVolumeClaims(namespace).Get(name, metav1.GetOptions{})
	return
}
func (m *DefaultApplicationManager) SyncPodConfigure(pcs []*appModels.PodConfigure) (int, error) {
	sess := m.engine.NewSession()
	defer sess.Close()
	i := 0
	for _, pc := range pcs {
		if _, err := sess.Insert(pc); err != nil {
			logrus.Warnf("Sync PodConfigure (%+v) failed (%v)", pc, err)
		} else {
			i++
			logrus.Debugf("Sync PodConfigure (%+v)", pc)
		}
	}
	return i, sess.Commit()
}

func NewApplicationManager(engine *xorm.Engine, clientSet *kubernetes.Clientset) (ApplicationManager, error) {
	return &DefaultApplicationManager{
		engine: engine,
		clientSet: clientSet,
	}, nil
}
