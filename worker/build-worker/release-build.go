package main

import (
	"fmt"

	"github.com/marwan-at-work/baghdad/worker"
	"github.com/streadway/amqp"
)

func releaseBuild(q string, w *worker.Worker) {
	err := w.Publish(worker.PublishOpts{
		Exchange: "",
		Key:      q,
		Msg: amqp.Publishing{
			DeliveryMode: amqp.Persistent,
		},
	})
	if err != nil {
		w.Log(fmt.Sprintf("%v: could not be released from q: %v", q, err))
	}
}
