package worker

import (
	"errors"

	"github.com/streadway/amqp"
)

// Worker covers rabbitmq's consume essentials: connecting, logging, and acknowledging.
type Worker struct {
	Ch   *amqp.Channel
	Conn *amqp.Connection
}

// NewWorker takes in a queue name and returns a Worker struct
func NewWorker(amqpURL string) (w *Worker, err error) {
	setLogLevel()
	if amqpURL == "" {
		err = errors.New("AMQP_URL cannot be empty")
		return
	}

	w = &Worker{}
	w.Conn, err = amqp.Dial(amqpURL)
	if err != nil {
		return
	}

	w.Ch, err = w.Conn.Channel()
	if err != nil {
		return
	}

	return
}

// Close closes the rabbitmq channel and connection.
func (w *Worker) Close() {
	w.Ch.Close()
	w.Conn.Close()
}

// EnsureQueues calles QueueDeclare on every string passed. It uses baghdad default
// settings. For different settings per queue, use the QueueDeclare directly.
func (w *Worker) EnsureQueues(qs ...string) (err error) {
	for _, q := range qs {
		_, err = w.QueueDeclare(QueueOpts{
			Name:       q,
			Durable:    true,
			AutoDelete: false,
			Exclusive:  false,
			NoWait:     false,
			Args:       nil,
		})
		if err != nil {
			return
		}
	}

	return
}

// EnsureExchanges calles ExchangeDeclare on passed string. Currently used for
// the event worker. So no need to make a flexible api.
func (w *Worker) EnsureExchanges(exs ...string) (err error) {
	for _, ex := range exs {
		err = w.Ch.ExchangeDeclare(
			ex,
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
	}

	return
}

// QueueDeclare calls QueueDeclare on the registered channel. Provides better
// readability for the callers Queue options
func (w *Worker) QueueDeclare(opts QueueOpts) (amqp.Queue, error) {
	return w.Ch.QueueDeclare(
		opts.Name,
		opts.Durable,
		opts.AutoDelete,
		opts.Exclusive,
		opts.NoWait,
		opts.Args,
	)
}

// Consume calls amqp.Channel.Consume with Consume options.
func (w *Worker) Consume(opts ConsumeOpts) (<-chan amqp.Delivery, error) {
	return w.Ch.Consume(
		opts.Queue,
		opts.Consumer,
		opts.AutoAck,
		opts.Exclusive,
		opts.NoLocal,
		opts.NoWait,
		opts.Args,
	)
}

// Publish satisfies the bus.Publisher interface. Because rabbitmq's publish comes
// with a lot of configuration. The app almost always sends msgs with the configuration
// inside this method. Therefore, abstraction this makes a reasonable interface.
func (w *Worker) Publish(qName string, msg []byte) error {
	return w.RawPublish(PublishOpts{
		Exchange:  "",
		Key:       qName,
		Mandatory: false,
		Immediate: false,
		Msg: amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			Body:         msg,
		},
	})
}

// RawPublish sends a rabbitmq message on the worker's channel.
func (w *Worker) RawPublish(opts PublishOpts) error {
	return w.Ch.Publish(
		opts.Exchange,
		opts.Key,
		opts.Mandatory,
		opts.Immediate,
		opts.Msg,
	)
}

// Broadcast helper function to publish to an exchange, with default options.
func (w *Worker) Broadcast(ex, routingKey string, msg []byte) error {
	return w.RawPublish(PublishOpts{
		Exchange:  ex,
		Key:       routingKey,
		Mandatory: false,
		Immediate: false,
		Msg: amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			Body:         msg,
		},
	})
}

// PublishOpts explicit options for amqp.Ch.Publish
type PublishOpts struct {
	Exchange  string
	Key       string
	Mandatory bool
	Immediate bool
	Msg       amqp.Publishing
}

// ConsumeOpts explicit options for amqp.Ch.Consume
type ConsumeOpts struct {
	Queue     string
	Consumer  string
	AutoAck   bool
	Exclusive bool
	NoLocal   bool
	NoWait    bool
	Args      amqp.Table
}

// QueueOpts explicit options for amqp.Ch.QueueDeclare
type QueueOpts struct {
	Name       string
	Durable    bool
	AutoDelete bool
	Exclusive  bool
	NoWait     bool
	Args       amqp.Table
}
