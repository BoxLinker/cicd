package pubsub

import (
	"errors"
	"context"
)

var ErrNotFound = errors.New("topic not found")

type Message struct {
	ID string `json:"id,omitempty"`
	Data []byte `json:"data"`
	Labels map[string]string `json:"labels,omitempty"`
}

type Receiver func(Message)

type Publisher interface {
	Create(c context.Context, topic string) error

	// publish the message
	Publish(c context.Context, topic string, message Message) error

	// subscribes to the topic.
	// Receiver func is a callback func that receives published messages.
	Subscribe(c context.Context, topic string, receiver Receiver) error

	// removes the named topic
	Remove(c context.Context, topic string) error
}

