package main

import (
	"fmt"
	"sync"

	"github.com/marwan-at-work/baghdad"
	"github.com/marwan-at-work/baghdad/utils"
	"github.com/marwan-at-work/baghdad/worker"

	"github.com/streadway/amqp"
)

type rawBuildJob []byte

var syncher = map[string]chan rawBuildJob{}
var m sync.Mutex

func buildWorker(ch chan rawBuildJob, w *worker.Worker, done <-chan amqp.Delivery) {
	for bj := range ch {
		w.Publish(worker.PublishOpts{
			Exchange: "",
			Key:      "build",
			Msg: amqp.Publishing{
				DeliveryMode: amqp.Persistent,
				Body:         bj,
			},
		})

		<-done
	}
}

func consume(d amqp.Delivery, w *worker.Worker) {
	bj, err := baghdad.DecodeBuildJob(d.Body)
	if err != nil {
		w.Log(fmt.Sprintf("could not unmarshal deploy-sync message: %v", err))
		d.Ack(false)
		return
	}

	q := "build--" + bj.RepoName
	m.Lock()
	ch, ok := syncher[q]
	if !ok {
		w.QueueDeclare(worker.QueueOpts{
			Name:       q,
			Durable:    true,
			AutoDelete: false,
			Exclusive:  false,
			NoWait:     false,
			Args:       nil,
		})

		done, _ := w.Consume(worker.ConsumeOpts{
			Queue:     q,
			Exclusive: true,
			Args:      nil,
		})

		ch = make(chan rawBuildJob)
		go buildWorker(ch, w, done)
		syncher[q] = ch
	}
	m.Unlock()
	d.Ack(false)
	ch <- d.Body
}

func listen(msgs <-chan amqp.Delivery, w *worker.Worker) {
	for d := range msgs {
		go consume(d, w)
	}
}

func main() {
	utils.ValidateEnvVars(getRequiredVars()...)
	w, err := worker.NewWorker(utils.GetAMQPURL())
	utils.FailOnError(err, "could not connect to rabbitmq")
	defer w.Conn.Close()
	defer w.Ch.Close()

	err = w.EnsureQueues("build-sync", "build", "logs")
	utils.FailOnError(err, "could not declare queue")

	msgs, err := w.Consume(worker.ConsumeOpts{
		Queue:     "build-sync",
		Consumer:  "",
		AutoAck:   false,
		Exclusive: false,
		NoLocal:   false,
		NoWait:    false,
		Args:      nil,
	})
	utils.FailOnError(err, "could not consume deploy-sync")

	go listen(msgs, w)

	fmt.Println("[x] - Listening for messages on the deploy-sync queue")
	<-make(chan bool) // wait forever
}
