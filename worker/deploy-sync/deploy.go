package main

import (
	"github.com/marwan-at-work/baghdad/worker"
	"github.com/streadway/amqp"
)

func deploy(projectQName string, bj []byte, w *worker.Worker) error {
	return w.Publish(worker.PublishOpts{
		Exchange:  "",
		Key:       "deploy",
		Mandatory: false,
		Immediate: false,
		Msg: amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			Body:         bj,
		},
	})
}
