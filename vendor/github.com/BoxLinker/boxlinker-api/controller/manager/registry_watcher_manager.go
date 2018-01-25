package manager

import (
	"github.com/BoxLinker/boxlinker-api/pkg/amqp"
	"github.com/BoxLinker/boxlinker-api/pkg/registry"
	"encoding/json"
)

type RegistryWatcherManager interface {
	Manager
	Publish(event *registry.Event) error
}

type defaultRegistryWatcherManager struct {
	DefaultManager
	amqpProducer *amqp.Producer
}

type DefaultRegistryWatcherManagerOptions struct {
	AmqpProducer *amqp.Producer
}

func NewDefaultRegistryWatcherManagerOptions(options DefaultRegistryWatcherManagerOptions) RegistryWatcherManager {
	return &defaultRegistryWatcherManager{
		amqpProducer: options.AmqpProducer,
	}
}

func (m *defaultRegistryWatcherManager) Publish(event *registry.Event) error {
	b, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return m.amqpProducer.PublishOnce(string(b))
}