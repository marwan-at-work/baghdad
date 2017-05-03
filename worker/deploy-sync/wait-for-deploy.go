package main

import (
	"fmt"
	"time"

	"github.com/marwan-at-work/baghdad/worker"
)

// waitForDeploy declares a queue and waits on it. It also times out after 30 minutes.
func waitForDeploy(projectQName string, w *worker.Worker) (err error) {
	msgs, err := w.Consume(worker.ConsumeOpts{
		Queue:     projectQName,
		Consumer:  "",
		AutoAck:   false,
		Exclusive: false,
		NoLocal:   false,
		NoWait:    false,
		Args:      nil,
	})

	if err != nil {
		err = fmt.Errorf("could not consume%v q: %v", projectQName, err)
		return
	}

	select {
	case d := <-msgs:
		err = d.Ack(false)
		if err != nil {
			return
		}
		err = w.Ch.Cancel(d.ConsumerTag, false)
	case <-time.After(time.Minute * 30):
		w.Log(fmt.Sprintf("%v - waiting for previous deploy to finish timed out", projectQName))
	}

	return
}
