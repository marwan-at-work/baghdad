package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/marwan-at-work/baghdad"
	"github.com/marwan-at-work/baghdad/utils"
	"github.com/marwan-at-work/baghdad/worker"
	"github.com/streadway/amqp"
)

func consume(d amqp.Delivery, w *worker.Worker) {
	dj, err := baghdad.DecodeDeployJob(d.Body)
	if err != nil {
		w.Log(fmt.Sprintf("could not unmarshal deploy message: %v", err))
		d.Ack(false)
		return
	}

	projectQ := dj.Baghdad.Project + "--" + dj.Env

	f, err := getStackCompose(utils.GetGithub(os.Getenv("ADMIN_TOKEN")), stackComposeGetOpts{
		Ctx:      context.Background(),
		Owner:    dj.RepoOwner,
		RepoName: dj.RepoName,
		Tag:      dj.Tag,
	})
	if err != nil {
		w.Log(fmt.Sprintf("could not get stack-compose.yml: %v", err))
		utils.ReleaseDeploy(projectQ, w)
		d.Ack(false)
		return
	}

	out, err := tagStackServices([]byte(f), dj)
	if err != nil {
		w.Log(fmt.Sprintf("could not marshal yaml: %v", err))
		utils.ReleaseDeploy(projectQ, w)
		d.Ack(false)
		return
	}

	// name space environment because event-worker could parallalize multi-env deployments.
	folderPath := filepath.Join("/compose-files", dj.RepoName, dj.Env, dj.Tag)
	filePath := filepath.Join(folderPath, "stack-compose.yml")
	err = os.MkdirAll(folderPath, 777)
	if err != nil {
		w.Log(fmt.Sprintf("could not make folder path: %v - %v", folderPath, err))
		utils.ReleaseDeploy(projectQ, w)
		d.Ack(false)
		return
	}
	err = ioutil.WriteFile(filePath, out, 777)
	if err != nil {
		w.Log(fmt.Sprintf("could not write stack-compose.yml to disk: %v", err))
		utils.ReleaseDeploy(projectQ, w)
		d.Ack(false)
		return
	}

	w.Log("deploying stack")
	err = stackDeploy(dj.RepoName, filePath, dj.Tag, stackDeployOpts{stdout: w})
	if err != nil {
		w.Log(fmt.Sprintf("could not deploy stack: %v", err))
		utils.ReleaseDeploy(projectQ, w)
		d.Ack(false)
		return
	}

	w.Log("checking for post deploy")
	if hasPostDeploy(dj.Baghdad, dj.Env) {
		siteURL := getSiteURL(dj, dj.Baghdad.PostDeploy.TargetService)
		pdj := baghdad.PostDeployJob{
			Baghdad:     dj.Baghdad,
			ProjectName: dj.RepoName,
			Tag:         dj.Tag,
			Env:         dj.Env,
			BranchName:  dj.BranchName,
			SiteURL:     siteURL,
		}
		b, _ := baghdad.EncodePostDeployJob(pdj)

		err = w.Publish(worker.PublishOpts{
			Exchange:  "",
			Key:       "post-deploy",
			Mandatory: false,
			Immediate: false,
			Msg: amqp.Publishing{
				DeliveryMode: amqp.Persistent,
				Body:         b,
			},
		})

		if err != nil {
			w.Log(fmt.Sprintf("%v could not send post-deploy, ignoring: %v", dj.RepoName, err))
			utils.ReleaseDeploy(projectQ, w)
		}

		d.Ack(false)
		return
	}

	w.Log("done")
	utils.ReleaseDeploy(projectQ, w)
	d.Ack(false)
}

func main() {
	godotenv.Load("/run/secrets/baghdad-vars")
	utils.ValidateEnvVars(getRequiredVars()...)
	w, err := worker.NewWorker(utils.GetAMQPURL())
	utils.FailOnError(err, "could not dial amqp")
	defer w.Conn.Close()
	defer w.Ch.Close()

	err = w.EnsureQueues("deploy", "post-deploy", "logs")
	utils.FailOnError(err, "could not declare queue")

	msgs, err := w.Consume(worker.ConsumeOpts{
		Queue:     "deploy",
		Consumer:  "",
		AutoAck:   false,
		Exclusive: false,
		NoLocal:   false,
		NoWait:    false,
		Args:      nil,
	})
	utils.FailOnError(err, "could not consume")

	go listen(msgs, w)

	fmt.Println("[x] - Listening for messages on the deploy queue")
	<-make(chan bool) // wait forever
}

func listen(msgs <-chan amqp.Delivery, w *worker.Worker) {
	for d := range msgs {
		go consume(d, w)
	}
}
