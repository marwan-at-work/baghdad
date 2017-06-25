package bus

import (
	"github.com/streadway/amqp"
)

// Broker wraps a rabbitmq functionality. It implements the publisher and broadcaster interfaces.
type Broker struct {
	AmqpURL string
}

// NewBroker takes a rabbitmq host url and returns a broker struct
func NewBroker(AmqpURL string) *Broker {
	return &Broker{AmqpURL: AmqpURL}
}

// Publish sends a message to a rabbitmq server with default settings
func (b *Broker) Publish(qName string, msg []byte) (err error) {
	conn, err := amqp.Dial(b.AmqpURL)
	if err != nil {
		return
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(qName, true, false, false, false, nil)
	if err != nil {
		return
	}

	err = ch.Publish("", q.Name, false, false, amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		Body:         msg,
	})

	return
}

// Broadcast sends a message to a rabbitmq fanout exchange.
func (b *Broker) Broadcast(exName, routingKey string, msg []byte) (err error) {
	conn, err := amqp.Dial(b.AmqpURL)
	if err != nil {
		return
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return
	}
	defer ch.Close()

	err = ch.ExchangeDeclare(
		exName,
		"fanout",
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		return
	}

	err = ch.Publish(exName, routingKey, false, false, amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		Body:         msg,
	})

	return
}
