package amqp

import (
	log "github.com/Sirupsen/logrus"
	"github.com/streadway/amqp"
	"fmt"
)

type Producer struct {
	uri string
	exchange string
	exchangeType string
	routingKey string
	reliable bool
}

type ProducerOptions struct {
	URI string
	Exchange string
	ExchangeType string
	RoutingKey string
	Reliable bool
}

func NewProducer(config ProducerOptions) *Producer {

	return &Producer{
		uri: config.URI,
		exchange: config.Exchange,
		exchangeType: config.ExchangeType,
		routingKey: config.RoutingKey,
		reliable: config.Reliable,
	}

}

func (p *Producer) Conn() (*amqp.Connection, error) {
	log.Debugf("dialing %q", p.uri)
	return amqp.Dial(p.uri)
}


func (p *Producer) PublishOnce(body string) error {
	conn, err := p.Conn()
	if err != nil {
		return fmt.Errorf("Dial: %s", err)
	}

	defer conn.Close()

	log.Debugf("got connection, getting channel")
	channel, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("Channel: %s", err)
	}

	log.Debugf("got channel, declaring %q Exchange (%q)", p.exchangeType, p.exchange)
	if err := channel.ExchangeDeclare(
		p.exchange,
		p.exchangeType,
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("Exchange declare %s", err)
	}

	log.Debugf("declared exchange, publish %dB body (%q)", len(body), body)
	if err := channel.Publish(
		p.exchange,   // publish to an exchange
		p.routingKey, // routing to 0 or more queues
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			Headers:         amqp.Table{},
			ContentType:     "text/plain",
			ContentEncoding: "",
			Body:            []byte(body),
			DeliveryMode:    amqp.Transient, // 1=non-persistent, 2=persistent
			Priority:        0,              // 0-9
			// a bunch of application/implementation-specific fields
		},
	); err != nil {
		return fmt.Errorf("Exchange publish: %s", err)
	}
	log.Debugf("Exchange publish ok!")
	return nil
}

