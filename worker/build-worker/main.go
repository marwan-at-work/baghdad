package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/marwan-at-work/baghdad"
	"github.com/marwan-at-work/baghdad/utils"
	"github.com/marwan-at-work/baghdad/worker"
	"github.com/streadway/amqp"
)

var (
	// AdminToken is the token the worker uses to perform github API operations
	AdminToken string
)

// WorkPath is the directory where the build worker clones, builds & tars projects.
const WorkPath = "/var/baghdad/workdir"

func consume(d amqp.Delivery, w *worker.Worker) {
	bj, err := baghdad.DecodeBuildJob(d.Body)
	if err != nil {
		w.Log(fmt.Sprintf("could not unmarshal build message: %v", err))
		d.Ack(false)
		return
	}

	var tag string
	switch bj.Type {
	case baghdad.PushEvent:
		tag, err = buildFromPush(bj, w)
	case baghdad.PullRequestEvent:
		// err = buildFromPR(bj, w)
	default:
		w.Log(fmt.Sprintf("%v err: build type unrecognized: %v", bj.RepoName, bj.Type))
	}

	if err != nil {
		w.Log(err.Error())
	}

	eventType := baghdad.BuildSuccessEvent
	if err != nil {
		eventType = baghdad.BuildFailureEvent
	}

	err = sendBuildEvent(bj, eventType, tag, w)
	if err != nil {
		w.Log(fmt.Sprintf("%v err: could not send build-event: %v", bj.RepoName, err))
	}

	releaseBuild("build--"+bj.RepoName, w)
	d.Ack(false)
}

func main() {
	godotenv.Load("/run/secrets/baghdad-vars")
	utils.ValidateEnvVars(getRequiredVars()...)
	AdminToken = os.Getenv("ADMIN_TOKEN")
	ensureBuildPath()
	w, err := worker.NewWorker(utils.GetAMQPURL())
	utils.FailOnError(err, "could not dial amqp")
	defer w.Conn.Close()
	defer w.Ch.Close()

	err = w.EnsureQueues("build", "logs")
	utils.FailOnError(err, "could not declare queue")
	err = w.EnsureExchange("event")
	utils.FailOnError(err, "could not declare exchange")

	msgs, err := w.Consume(worker.ConsumeOpts{
		Queue:     "build",
		Consumer:  "",
		AutoAck:   false,
		Exclusive: false,
		NoLocal:   false,
		NoWait:    false,
		Args:      nil,
	})
	utils.FailOnError(err, "could not consume")

	go listen(msgs, w)

	fmt.Println("[x] - Listening for messages on the build queue")
	<-make(chan bool) // wait forever
}

func listen(msgs <-chan amqp.Delivery, w *worker.Worker) {
	for d := range msgs {
		go consume(d, w)
	}
}
