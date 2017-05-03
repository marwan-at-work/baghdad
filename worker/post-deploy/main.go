package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	docker "github.com/docker/docker/client"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/term"
	"github.com/joho/godotenv"
	"github.com/marwan-at-work/baghdad"
	"github.com/marwan-at-work/baghdad/utils"
	"github.com/marwan-at-work/baghdad/worker"
	"github.com/streadway/amqp"
)

func consume(d amqp.Delivery, w *worker.Worker) {
	pdj, err := baghdad.DecodePostDeployJob(d.Body)
	if err != nil {
		w.Log(fmt.Sprintf("could not unmarshal deploy-sync message: %v", err))
		d.Ack(false)
		return
	}

	projectQ := pdj.Baghdad.Project + "--" + pdj.Env

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Minute*10)
	defer cancel()
	err = runPostDeployJob(ctx, w, pdj)

	if err != nil {
		w.Log(fmt.Sprintf("error: %v", err))
	}

	utils.ReleaseDeploy(projectQ, w)
	d.Ack(false)
}

func runPostDeployJob(
	ctx context.Context,
	w *worker.Worker,
	pdj baghdad.PostDeployJob,
) (err error) {
	dc, err := docker.NewClient(os.Getenv("DOCKER_REMOTE_API_URL"), "1.26", nil, nil)
	if err != nil {
		return
	}

	authConfig := types.AuthConfig{
		Username: os.Getenv("DOCKER_AUTH_USER"),
		Password: os.Getenv("DOCKER_AUTH_PASS"),
	}

	_, err = dc.RegistryLogin(ctx, authConfig)
	if err != nil {
		return
	}

	img := fmt.Sprintf("%v/%v:%v", os.Getenv("DOCKER_ORG"), pdj.Baghdad.PostDeploy.SourceService, pdj.Tag)
	registryAuth, err := utils.EncodeAuthToBase64(authConfig)
	if err != nil {
		err = fmt.Errorf("could not encode auth: %v", err)
		return
	}

	imgOut, err := dc.ImagePull(ctx, img, types.ImagePullOptions{
		RegistryAuth: registryAuth,
	})
	if err != nil {
		err = fmt.Errorf("could not pull %v image: %v", img, err)
		return
	}
	defer imgOut.Close()
	fd, isTerm := term.GetFdInfo(w)
	err = jsonmessage.DisplayJSONMessagesStream(imgOut, w, fd, isTerm, nil)
	if err != nil {
		fmt.Println("could not display docker build output, process running though.")
	}

	tenMinutes := 10 * 60
	ct, err := dc.ContainerCreate(
		ctx,
		&container.Config{
			Image:       img,
			StopTimeout: &tenMinutes,
			Env: []string{
				fmt.Sprintf("SITE_URL=%v", pdj.SiteURL),
				fmt.Sprintf("PROJECT_NAME=%v", pdj.ProjectName),
				fmt.Sprintf("DEPLOYED_TAG=%v", pdj.Tag),
				fmt.Sprintf("DEPLOYED_ENV=%v", pdj.Env),
				fmt.Sprintf("BRANCH_NAME=%v", pdj.BranchName),
			},
		},
		&container.HostConfig{
			AutoRemove: false,
		},
		nil,
		"",
	)
	if err != nil {
		err = fmt.Errorf("could not create container %v: %v", img, err)
		return
	}

	if err = dc.ContainerStart(ctx, ct.ID, types.ContainerStartOptions{}); err != nil {
		return
	}

	out, err := dc.ContainerLogs(ctx, ct.ID, types.ContainerLogsOptions{
		ShowStdout: true,
		Follow:     true,
		ShowStderr: true,
	})
	if err != nil {
		fmt.Println("could not get container logs, but job is still running")
	}
	defer out.Close()

	io.Copy(w, out)
	if _, err = dc.ContainerWait(ctx, ct.ID); err != nil {
		err = fmt.Errorf("container run failed %v: %v", img, err)
		return
	}

	dc.ContainersPrune(ctx, filters.Args{})

	return
}

func listen(msgs <-chan amqp.Delivery, w *worker.Worker) {
	for d := range msgs {
		go consume(d, w)
	}
}

func main() {
	godotenv.Load("/run/secrets/baghdad-vars")
	utils.ValidateEnvVars(getRequiredVars()...)
	ensureDownloadPath()
	w, err := worker.NewWorker(utils.GetAMQPURL())
	utils.FailOnError(err, "could not dial amqp")
	defer w.Conn.Close()
	defer w.Ch.Close()

	err = w.EnsureQueues("post-deploy", "logs")

	utils.FailOnError(err, "could not declare queues")

	msgs, err := w.Consume(worker.ConsumeOpts{
		Queue:     "post-deploy",
		Consumer:  "",
		AutoAck:   false,
		Exclusive: false,
		NoLocal:   false,
		NoWait:    false,
		Args:      nil,
	})
	utils.FailOnError(err, "could not consume")

	go listen(msgs, w)

	fmt.Println("[x] - Listening for messages on the post-deploy queue")
	<-make(chan bool) // wait forever
}

func ensureDownloadPath() {
	os.MkdirAll("/projects", 0666)
}
