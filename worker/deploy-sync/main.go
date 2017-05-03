package main

import (
	"fmt"

	"github.com/marwan-at-work/baghdad"
	"github.com/marwan-at-work/baghdad/utils"
	"github.com/marwan-at-work/baghdad/worker"

	"github.com/streadway/amqp"
)

func consume(d amqp.Delivery, w *worker.Worker) {
	bj, err := baghdad.DecodeDeployJob(d.Body)
	if err != nil {
		w.Log(fmt.Sprintf("could not unmarshal deploy-sync message: %v", err))
		d.Ack(false)
		return
	}

	projectQName := bj.Baghdad.Project + "--" + bj.Env
	_, err = w.QueueDeclare(worker.QueueOpts{
		Name:       projectQName,
		Durable:    true,
		AutoDelete: false,
		Exclusive:  false,
		NoWait:     false,
		Args: amqp.Table{
			"x-max-length": int32(1),
		},
	})

	if err != nil {
		w.Log(fmt.Sprintf("could not declare %v queue: %v", projectQName, err))
		d.Ack(false)
		return
	}

	shouldWait, err := checkProjectQueue(projectQName)
	if err != nil {
		w.Log(fmt.Sprintf("could not get %v Q from redis db: %v", projectQName, err))
		d.Ack(false)
		return
	}

	if shouldWait {
		w.Log(fmt.Sprintf("waiting for the next deploy to finish, time out in 30 minutes: %v", projectQName))
		err = waitForDeploy(projectQName, w)
	}

	if err != nil {
		w.Log(fmt.Sprintf("error while waiting for previous deploy q: %v", err))
		d.Ack(false)
		return
	}

	err = deploy(projectQName, d.Body, w)

	if err != nil {
		w.Log(fmt.Sprintf("could not send deploy job for %v: %v", projectQName, err))
	}

	d.Ack(false)
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

	w.EnsureQueues("deploy-sync", "deploy", "logs")
	utils.FailOnError(err, "could not declare queue")

	msgs, err := w.Consume(worker.ConsumeOpts{
		Queue:     "deploy-sync",
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
