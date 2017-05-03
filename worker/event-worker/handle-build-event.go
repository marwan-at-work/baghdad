package main

import (
	"fmt"

	"github.com/marwan-at-work/baghdad"
	"github.com/marwan-at-work/baghdad/worker"
	"github.com/streadway/amqp"
)

func handleBuildEvent(be baghdad.BuildEvent, w *worker.Worker) {
	if be.EventType == baghdad.BuildSuccessEvent {
		for env, config := range be.Baghdad.Environments {
			if config == "auto" {
				dj := baghdad.DeployJob{
					Baghdad:    be.Baghdad,
					BranchName: be.BranchName,
					Tag:        be.Tag,
					Env:        env,
					RepoName:   be.RepoName,
				}

				b, err := baghdad.EncodeDeployJob(dj)
				if err != nil {
					w.Log(fmt.Sprintf("could not marshal deploy job: %v", err))
					return
				}

				w.Publish(worker.PublishOpts{
					Exchange:  "",
					Key:       "deploy-sync",
					Mandatory: false,
					Immediate: false,
					Msg: amqp.Publishing{
						DeliveryMode: amqp.Persistent,
						Body:         b,
					},
				})
			}
		}
	}
}
