package worker

import (
	"errors"
	"fmt"

	"github.com/streadway/amqp"
)

// Worker covers rabbitmq's consume essentials: connecting, logging, and acknowledging.
type Worker struct {
	Ch   *amqp.Channel
	Conn *amqp.Connection
}

// NewWorker takes in a queue name and returns a Worker struct
func NewWorker(amqpURL string) (w *Worker, err error) {
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

// Write implements the io.Writer interface, this way the execute commands
// can directly link their Stdout & Stderr to the worker logger.
func (w *Worker) Write(p []byte) (n int, err error) {
	n = len(p)
	w.Log(string(p))

	return
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

// EnsureExchange calles ExchangeDeclare on passed string. Currently used for
// the event worker. So no need to make a flexible api.
func (w *Worker) EnsureExchange(ex string) error {
	return w.Ch.ExchangeDeclare(
		ex,
		"fanout",
		true,
		false,
		false,
		false,
		nil,
	)
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

// Publish sends a rabbitmq message on the worker's channel.
func (w *Worker) Publish(opts PublishOpts) error {
	return w.Ch.Publish(
		opts.Exchange,
		opts.Key,
		opts.Mandatory,
		opts.Immediate,
		opts.Msg,
	)
}

// Log sends a message to the "logs" channel & stdout.
func (w *Worker) Log(msg string) {
	fmt.Println(msg)
	err := w.Publish(PublishOpts{
		Exchange:  "",
		Key:       "logs",
		Mandatory: false,
		Immediate: false,
		Msg: amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			Body:         []byte(msg),
		},
	})

	if err != nil {
		fmt.Println("could not publish log:", err)
	}
}

// ExPub helper function to publish to an exchange, with default options.
func (w *Worker) ExPub(ex, routingKey string, msg []byte) error {
	return w.Publish(PublishOpts{
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
