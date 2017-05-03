package main

import (
	"fmt"

	"github.com/marwan-at-work/baghdad"
	"github.com/marwan-at-work/baghdad/utils"
	"github.com/marwan-at-work/baghdad/worker"
	"github.com/streadway/amqp"
)

func consume(d amqp.Delivery, w *worker.Worker) {
	switch d.RoutingKey {
	case "build":
		be, err := baghdad.DecodeBuildEvent(d.Body)
		if err != nil {
			w.Log(fmt.Sprintf("could not unmarshal body: %v", err))
			d.Ack(false)
			return
		}

		handleBuildEvent(be, w)
	default:
		w.Log(fmt.Sprintf("unrecognized event: %v", d.RoutingKey))
	}

	d.Ack(false)
}

func main() {
	w, err := worker.NewWorker(utils.GetAMQPURL())
	utils.FailOnError(err, "could not dial amqp")
	defer w.Conn.Close()
	defer w.Ch.Close()

	err = w.EnsureQueues("logs", "deploy")
	utils.FailOnError(err, "could not declare queues")
	err = w.EnsureExchange("event")
	utils.FailOnError(err, "could not declare exchange")

	q, err := w.QueueDeclare(worker.QueueOpts{
		Name:       "", // create a unique name for this subscriber.
		Durable:    true,
		AutoDelete: false,
		Exclusive:  true,
		NoWait:     false,
		Args:       nil,
	})
	utils.FailOnError(err, "could not declare event queue")

	err = w.Ch.QueueBind(q.Name, "", "event", false, nil)
	utils.FailOnError(err, "could not bind to event queue")

	msgs, err := w.Consume(worker.ConsumeOpts{
		Queue:     q.Name,
		Consumer:  "",
		AutoAck:   false,
		Exclusive: false,
		NoLocal:   false,
		NoWait:    false,
		Args:      nil,
	})
	utils.FailOnError(err, "could not consume")

	go listen(msgs, w)

	fmt.Println("[x] - Listening for messages on the event exchange")
	<-make(chan bool) // wait forever
}

func listen(msgs <-chan amqp.Delivery, w *worker.Worker) {
	for d := range msgs {
		go consume(d, w)
	}
}
