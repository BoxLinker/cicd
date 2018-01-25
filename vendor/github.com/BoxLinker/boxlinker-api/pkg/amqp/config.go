package amqp

type ProducerConfig struct {
	URI string
	Exchange string
	ExchangeType string
	RoutingKey string
	Reliable bool
}

type ConsumerConfig struct {
	URI string
	Exchange string
	ExchangeType string
	QueueName string
	BindingKey string
	NotifyMsg chan []byte
}
