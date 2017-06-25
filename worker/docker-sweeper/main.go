package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	docker "github.com/docker/docker/client"
	"github.com/marwan-at-work/baghdad/utils"
	"github.com/marwan-at-work/baghdad/worker"
	"github.com/streadway/amqp"
)

func consume(d amqp.Delivery, w *worker.Worker) {
	d.Ack(false)
	c, err := docker.NewEnvClient()
	if err != nil {
		fmt.Println("could not get docker client")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*10)
	_, err = c.ContainersPrune(ctx, filters.Args{})
	if err != nil {
		fmt.Println("could not prune containers")
		cancel()
	}

	_, err = c.VolumesPrune(ctx, filters.Args{})
	if err != nil {
		fmt.Println("could not prune volumes", err)
	}

	_, err = c.ImagesPrune(ctx, filters.Args{})
	if err != nil {
		fmt.Println("could not prune images", err)
	}

	images, err := c.ImageList(ctx, types.ImageListOptions{})
	if err != nil {
		fmt.Println("could not get images", err)
		cancel()
	}

	for _, image := range images {
		_, err = c.ImageRemove(
			ctx,
			image.ID,
			types.ImageRemoveOptions{Force: true, PruneChildren: true},
		)

		if err != nil {
			continue
		}
		fmt.Println("removed: ", formatImageID(image.ID))
	}

	cancel()
}

func main() {
	w, err := worker.NewWorker(utils.GetAMQPURL())
	utils.FailOnError(err, "could not dial amqp")
	defer w.Close()

	err = w.EnsureExchanges("sweep")
	utils.FailOnError(err, "could not declare exchanges")

	q, err := w.QueueDeclare(worker.QueueOpts{
		Name:       "", // create a unique name for this subscriber.
		Durable:    true,
		AutoDelete: false,
		Exclusive:  true,
		NoWait:     false,
		Args:       nil,
	})
	utils.FailOnError(err, "could not declare sweeper bound queue")

	err = w.Ch.QueueBind(q.Name, "", "sweep", false, nil)
	utils.FailOnError(err, "could not bind to sweep exchange")

	msgs, err := w.Consume(worker.ConsumeOpts{
		Queue:     q.Name,
		Consumer:  "",
		AutoAck:   false,
		Exclusive: false,
		NoLocal:   false,
		NoWait:    false,
		Args:      nil,
	})
	utils.FailOnError(err, "could not consume")

	fmt.Println("[x] - Listening for messages on the sweeper exchange")
	listen(msgs, w)
}

func listen(msgs <-chan amqp.Delivery, w *worker.Worker) {
	for d := range msgs {
		consume(d, w)
	}
}

func formatImageID(s string) string {
	if strings.HasPrefix(s, "sha256:") {
		id := strings.Split(s, ":")[1]
		if len(id) > 12 {
			return id[:12]
		}

		return id
	}

	return s
}
