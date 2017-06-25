package main

import (
	"fmt"

	"github.com/marwan-at-work/baghdad"
	"github.com/marwan-at-work/baghdad/utils"
	"github.com/marwan-at-work/baghdad/worker"
)

func handleBuildEvent(be baghdad.BuildEvent, w *worker.Worker, logger *worker.Logger) {
	switch be.EventType {
	case baghdad.BuildSuccessEvent:
		if be.Baghdad.SlackURL != "" {
			utils.SendSlackMessage(be.Baghdad.SlackURL, fmt.Sprintf("%v built %v successfully", be.RepoName, be.Tag))
		}
		for env, config := range be.Baghdad.Environments {
			if config == "auto" {
				dj := baghdad.DeployJob{
					Baghdad:    be.Baghdad,
					BranchName: be.BranchName,
					Tag:        be.Tag,
					Env:        env,
					RepoName:   be.RepoName,
					RepoOwner:  be.RepoOwner,
					LogID:      be.LogID,
				}

				b, err := baghdad.EncodeDeployJob(dj)
				if err != nil {
					logger.Loglnf("could not marshal deploy job: %v", err)
					return
				}

				logger.Log("sending deploy-sync job")

				w.Publish("deploy-sync", b)
			}
		}
	case baghdad.BuildFailureEvent:
		if be.Baghdad.SlackURL != "" {
			utils.SendSlackMessage(be.Baghdad.SlackURL, fmt.Sprintf("build job for %v/%v failed", be.RepoName, be.Tag))
		}
	}
	if be.EventType == baghdad.BuildSuccessEvent {
		for env, config := range be.Baghdad.Environments {
			if config == "auto" {
				dj := baghdad.DeployJob{
					Baghdad:    be.Baghdad,
					BranchName: be.BranchName,
					Tag:        be.Tag,
					Env:        env,
					RepoName:   be.RepoName,
					RepoOwner:  be.RepoOwner,
				}

				b, err := baghdad.EncodeDeployJob(dj)
				if err != nil {
					logger.Loglnf("could not marshal deploy job: %v", err)
					return
				}

				logger.Log("sending deploy-sync job")

				w.Publish("deploy-sync", b)
			}
		}
	}
}
