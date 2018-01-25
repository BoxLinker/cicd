package amqp

import (
	"github.com/streadway/amqp"
	log "github.com/Sirupsen/logrus"
	"fmt"
	"time"
	"errors"
)

type Consumer struct {
	config *ConsumerConfig

	conn *amqp.Connection
	channel *amqp.Channel
	Tag string
	done chan error

	NotifyMsg chan []byte
}

func NewConsumer(config *ConsumerConfig) (*Consumer, error) {
	return &Consumer{
		config: config,
	}, nil
}

func (client *Consumer) Run() {
	for {
		err := client.connect()
		log.Errorf("amqp connect err: (%s)\nstart to reconnect.", err.Error())
		time.Sleep(time.Second*3)
	}
}
func (client *Consumer) connect() error {
	errCh := make(chan error)
	config := client.config
	var err error
	log.Debugf("AMQP dial: %s", config.URI)
	client.conn, err = amqp.Dial(config.URI)

	if err != nil {
		return err
	}

	go func(){
		err := <-client.conn.NotifyClose(make(chan *amqp.Error))
		log.Debugf("AMQP close: %s", err.Error())
		errCh <- errors.New(err.Error())
	}()

	log.Debugf("Got Connection, getting channel")
	client.channel, err = client.conn.Channel()

	if err != nil {
		return err
	}

	log.Debugf("Got Channel, declaring Exchange: (%q)", config.Exchange)
	if err = client.channel.ExchangeDeclare(
		config.Exchange,
		config.ExchangeType,
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return err
	}

	log.Debugf("Declared Exchange, declaring queue: %s", config.QueueName)
	queue, err := client.channel.QueueDeclare(
		config.QueueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	log.Debugf("declared Queue (%q %d messages, %d consumers), binding to Exchange (key %q)",
		queue.Name, queue.Messages, queue.Consumers, config.BindingKey)

	if err = client.channel.QueueBind(
		queue.Name,
		config.BindingKey,
		config.Exchange,
		false,
		nil,
	); err != nil {
		return err
	}

	log.Debugf("Queue bound to Exchange, starting Consume (consumer tag %q)", client.Tag)
	deliveries, err := client.channel.Consume(
		queue.Name,
		client.Tag,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}
	client.NotifyMsg = client.config.NotifyMsg
	go handle(deliveries, client.NotifyMsg, client.done)

	return <-errCh
}

func (client *Consumer) Shutdown() error {
	// will close() the deliveries channel
	if err := client.channel.Cancel(client.Tag, true); err != nil {
		return fmt.Errorf("Consumer cancel failed: %s", err)
	}

	if err := client.conn.Close(); err != nil {
		return fmt.Errorf("AMQP connection close error: %s", err)
	}

	defer log.Printf("AMQP shutdown OK")

	// wait for handle() to exit
	return <-client.done
}

func handle(deliveries <-chan amqp.Delivery, notifyMsg chan []byte, done chan error){
	for d := range deliveries {
		log.Printf(
			"got %dB delivery: [%v] %q",
			len(d.Body),
			d.DeliveryTag,
			d.Body,
		)
		notifyMsg <- d.Body
		d.Ack(false)
	}
	log.Debugf("handle: deliveries channel closed")
	done <- nil
}