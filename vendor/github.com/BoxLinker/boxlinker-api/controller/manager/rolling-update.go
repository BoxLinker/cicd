package manager

import (
	"github.com/go-xorm/xorm"
	"k8s.io/client-go/kubernetes"
	"github.com/BoxLinker/boxlinker-api/pkg/amqp"
	"github.com/Sirupsen/logrus"
	"encoding/json"
	"github.com/BoxLinker/boxlinker-api/pkg/registry"
	"strings"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"fmt"
)

type RollingUpdateManager interface {
	Manager
}

type defaultRollingUpdateManager struct {
	DefaultManager
	config *DefaultRollingUpdateConfig
	engine *xorm.Engine
	clientSet *kubernetes.Clientset
	notifyMsg chan []byte
}

type DefaultRollingUpdateConfig struct {
	RegistryHost string
}

func NewDefaultRollingUpdateManager (config *DefaultRollingUpdateConfig, engine *xorm.Engine, clientSet *kubernetes.Clientset, amqpConsumerConfig *amqp.ConsumerConfig) (RollingUpdateManager, error) {
	notifyMsg := make(chan []byte)
	m := &defaultRollingUpdateManager{
		config: config,
		engine: engine,
		clientSet: clientSet,
		notifyMsg: notifyMsg,
	}
	amqpConsumerConfig.NotifyMsg = notifyMsg
	consumer, err := amqp.NewConsumer(amqpConsumerConfig)
	if err != nil {
		return nil, err
	}
	logrus.Debugf("start to run amqp consumer")
	logrus.Debugf("consumer config: (%+v)", amqpConsumerConfig)
	go consumer.Run()
	go m.RollingUpdate()
	return m, nil
}

func (m *defaultRollingUpdateManager) RollingUpdate() {
	var (
		msg []byte
		event *registry.Event
	)
	for {
		event = &registry.Event{}
		msg = <-m.notifyMsg
		if err := json.Unmarshal(msg, event); err == nil {
			go m.DoRollingUpdate(event)
		} else {
			logrus.Errorf("JSON parse err: (%v)", err.Error())
		}
	}
}

func (m *defaultRollingUpdateManager) DoRollingUpdate(event *registry.Event) {
	if event.Action != "push" {
		logrus.Debugf("rolling-update will not proceed event: (%s)", event.Action)
		return
	}
	parts := strings.Split(event.Target.Repository, "/")
	if len(parts) != 2 {
		logrus.Debugf("rolling-update will not proceed repository: (%s)", event.Target.Repository)
		return
	}
	namespace := parts[0]
	image := parts[1]
	tag := event.Target.Tag
	deployOperator := m.clientSet.AppsV1beta1().Deployments(namespace)
	deploys, err := deployOperator.List(metav1.ListOptions{})
	if err != nil {
		logrus.Errorf("get deployments err: (%s)", err.Error())
		return
	}
	imageName := fmt.Sprintf("%s/%s/%s", m.config.RegistryHost, namespace, image)
	imageURL := fmt.Sprintf("%s:%s", imageName, tag)
	for _, deploy := range deploys.Items {
		name := deploy.ObjectMeta.Name
		containers := deploy.Spec.Template.Spec.Containers
		if len(containers) != 1 {
			logrus.Warnf("rolling-update only proceed containers len equal 1")
			return
		}
		container := &containers[0]
		cImage := container.Image
		parts := strings.Split(cImage, ":")
		if len(parts) != 2 {
			logrus.Warnf("deployment [%s] got invalid image url: (%s)", name, cImage)
			return
		}

		if imageName != parts[0] {
			return
		}

		if imageURL == cImage {
			logrus.Warnf("the deployment to be updated get same image url with registry event. (%s)", imageURL)
		}

		container.Image = imageURL
		if _, err := deployOperator.Update(&deploy); err != nil {
			logrus.Errorf("rolling update deployment [%s] err: (%s)", name, err.Error())
			return
		}
		logrus.Debugf("rolling update success.")
	}
}