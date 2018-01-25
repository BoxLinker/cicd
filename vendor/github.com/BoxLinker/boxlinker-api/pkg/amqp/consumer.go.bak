package amqp

import (
	"github.com/streadway/amqp"
	log "github.com/Sirupsen/logrus"
	"fmt"
)

type Consumer struct {
	URI string
	Exchange string
	ExchangeType string
	QueueName string
	BindingKey string

	conn *amqp.Connection
	channel *amqp.Channel
	Tag string
	done chan error

	NotifyMsg chan []byte

}

func (client *Consumer) Run() (error) {

	//client := &Consumer{
	//	uri: 			config.URI,
	//	exchange: 	 	config.Exchange,
	//	exchangeType: 	config.ExchangeType,
	//	queueName: 		config.QueueName,
	//	bindingKey: 	config.BindingKey,
	//	tag: 			config.ConsumerTag,
	//	NotifyMsg: 		config.NotifyMsg,
	//}

	var err error
	log.Debugf("AMQP dial: %s", client.URI)
	client.conn, err = amqp.Dial(client.URI)

	if err != nil {
		return err
	}

	go func(){
		log.Debugf("AMQP close: %s", <-client.conn.NotifyClose(make(chan *amqp.Error)))
	}()

	log.Debugf("Got Connection, getting channel")
	client.channel, err = client.conn.Channel()

	if err != nil {
		return err
	}

	log.Debugf("Got Channel, declaring Exchange: (%q)", client.Exchange)
	if err = client.channel.ExchangeDeclare(
		client.Exchange,
		client.ExchangeType,
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return err
	}

	log.Debugf("Declared Exchange, declaring queue: %s", client.QueueName)
	queue, err := client.channel.QueueDeclare(
		client.QueueName,
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
		queue.Name, queue.Messages, queue.Consumers, client.BindingKey)

	if err = client.channel.QueueBind(
		queue.Name,
		client.BindingKey,
		client.Exchange,
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

	go handle(deliveries, client.NotifyMsg)

	return nil
}

func (c *Consumer) Shutdown() error {
	// will close() the deliveries channel
	if err := c.channel.Cancel(c.Tag, true); err != nil {
		return fmt.Errorf("Consumer cancel failed: %s", err)
	}

	if err := c.conn.Close(); err != nil {
		return fmt.Errorf("AMQP connection close error: %s", err)
	}

	defer log.Printf("AMQP shutdown OK")

	// wait for handle() to exit
	return <-c.done
}


func handle(deliveries <-chan amqp.Delivery, notifyMsg chan []byte){
	for d := range deliveries {
		//log.Printf(
		//	"got %dB delivery: [%v] %q",
		//	len(d.Body),
		//	d.DeliveryTag,
		//	d.Body,
		//)
		notifyMsg <- d.Body
		d.Ack(false)
	}
	log.Debugf("handle: deliveries channel closed")
	//done <- nil
}