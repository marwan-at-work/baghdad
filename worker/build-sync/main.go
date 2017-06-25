package main

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"github.com/marwan-at-work/baghdad"
	"github.com/marwan-at-work/baghdad/utils"
	"github.com/marwan-at-work/baghdad/worker"

	"github.com/streadway/amqp"
)

type buildCh chan baghdad.BuildJob

var pc = map[string]buildCh{}
var m sync.Mutex

func build(w *worker.Worker, ch buildCh) {
	for bj := range ch {
		logger := worker.NewLogger(bj.Baghdad.Project, bj.LogID, w)
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute*30)

		err := updateGithubStatus(bj.RepoOwner, bj.RepoName, bj.SHA, "pending", bj.LogID)
		if err != nil {
			logger.Loglnf("err: could not update status: %v", err)
		}

		var nextTag string
		if bj.Type == baghdad.PushEvent {
			nextTag, err = sendPushBuildJob(ctx, bj, w, ch, logger)
			eventType := baghdad.BuildSuccessEvent
			if err != nil {
				eventType = baghdad.BuildFailureEvent
				logger.Log("sending build failure event")
			} else {
				logger.Log("sending build success event")
			}

			if eventErr := sendBuildEvent(bj, eventType, nextTag, w); eventErr != nil {
				logger.Loglnf("err: could not send build-event: %v", err)
			}
		} else if bj.Type == baghdad.PullRequestEvent {
			err = sendPRBuildJob(ctx, bj, w, ch, logger)
		} else {
			err = errors.New("unrecognized build event")
		}
		if err != nil {
			logger.Log(err)
			updateGithubStatus(bj.RepoOwner, bj.RepoName, bj.SHA, "error", bj.LogID)
		} else {
			updateGithubStatus(bj.RepoOwner, bj.RepoName, bj.SHA, "success", bj.LogID)
		}

		bj.NextTag = nextTag
		rawBj, _ := baghdad.EncodeBuildJob(bj)
		w.Broadcast("remove-repo", "", rawBj)
		cancel()
	}
}

func consume(d amqp.Delivery, w *worker.Worker) {
	bj, err := baghdad.DecodeBuildJob(d.Body)
	if err != nil {
		fmt.Println(fmt.Sprintf("could not unmarshal deploy-sync message: %v", err))
		d.Ack(false)
		return
	}

	m.Lock()
	ch, ok := pc[bj.Baghdad.Project]
	if !ok {
		ch = make(chan baghdad.BuildJob)
		pc[bj.Baghdad.Project] = ch
		go build(w, ch)
	}
	m.Unlock()

	ch <- bj
	d.Ack(false)
}

func listen(msgs <-chan amqp.Delivery, w *worker.Worker) {
	for d := range msgs {
		go consume(d, w)
	}
}

func main() {
	godotenv.Load("/run/secrets/baghdad-vars")
	utils.ValidateEnvVars(getRequiredVars()...)
	w, err := worker.NewWorker(utils.GetAMQPURL())
	utils.FailOnError(err, "could not connect to rabbitmq")
	defer w.Close()
	ensureBuildPath()

	err = w.EnsureQueues("build-sync", "build")
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
