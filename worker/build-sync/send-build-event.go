package main

import (
	"github.com/marwan-at-work/baghdad"
	"github.com/marwan-at-work/baghdad/worker"
)

func sendBuildEvent(bj baghdad.BuildJob, eventType int, tag string, w *worker.Worker) error {
	be := baghdad.BuildEvent{
		BuildJob:  bj,
		EventType: eventType,
		Tag:       tag,
		LogID:     bj.LogID,
	}

	p, err := baghdad.EncodeBuildEvent(be)
	if err != nil {
		return err
	}

	return w.Broadcast("event", "build", p)
}
