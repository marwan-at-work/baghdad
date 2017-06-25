package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/joho/godotenv"
	"github.com/marwan-at-work/baghdad"
	pb "github.com/marwan-at-work/baghdad/services"
	"github.com/marwan-at-work/baghdad/utils"
	"github.com/marwan-at-work/baghdad/worker"
	"github.com/streadway/amqp"
)

type deployCh chan baghdad.DeployJob

var pc = map[string]deployCh{}
var m sync.Mutex

func deploy(w *worker.Worker, ch deployCh) {
	for dj := range ch {
		logger := worker.NewLogger(dj.Baghdad.Project, dj.LogID, w)
		logger.Log("starting deploy job")

		f, err := getStackCompose(utils.GetGithub(os.Getenv("ADMIN_TOKEN")), stackComposeGetOpts{
			Ctx:      context.Background(),
			Owner:    dj.RepoOwner,
			RepoName: dj.RepoName,
			Tag:      dj.Tag,
		})
		if err != nil {
			logger.Loglnf("could not get stack-compose.yml: %v", err)
			continue
		}

		domain := os.Getenv("BAGHDAD_DOMAIN_NAME")
		out, err := utils.TagStackServices([]byte(f), dj.Baghdad, dj.Tag, dj.BranchName, dj.Env, domain)
		if err != nil {
			logger.Loglnf("could not marshal yaml: %v", err)
			continue
		}

		// name space environment because event-worker could parallalize multi-env deployments.
		folderPath := filepath.Join("/compose-files", dj.RepoName, dj.Env, dj.Tag)
		filePath := filepath.Join(folderPath, "stack-compose.yml")
		err = os.MkdirAll(folderPath, 777)
		if err != nil {
			logger.Loglnf("could not make folder path: %v - %v", folderPath, err)
			continue
		}
		err = ioutil.WriteFile(filePath, out, 777)
		if err != nil {
			logger.Loglnf("could not write stack-compose.yml to disk: %v", err)
			continue
		}

		stackName := fmt.Sprintf("%v_%v", dj.RepoName, dj.Env)
		logger.Log("deploying stack:", stackName)
		err = stackDeploy(stackName, filePath, dj.Tag, stackDeployOpts{stdout: logger})
		if err != nil {
			logger.Loglnf("could not deploy stack: %v", err)
			logger.Write(out)
			continue
		}

		if hasPostDeploy(dj.Baghdad, dj.Env) {
			logger.Log("starting post deploy job")
			siteURL := getSiteURL(dj, dj.Baghdad.PostDeploy.TargetService)
			pdj := baghdad.PostDeployJob{
				Baghdad:     dj.Baghdad,
				ProjectName: dj.Baghdad.Project,
				Tag:         dj.Tag,
				Env:         dj.Env,
				BranchName:  dj.BranchName,
				SiteURL:     siteURL,
			}
			b, _ := baghdad.EncodePostDeployJob(pdj)

			addr := os.Getenv("POST_DEPLOY_ADDR")
			var conn *grpc.ClientConn
			conn, err = grpc.Dial(addr, grpc.WithInsecure())
			if err != nil {
				logger.Logf("could not connect to service builder service: %v", err)
				continue
			}

			c := pb.NewPostDeployClient(conn)
			ctx, cancel := context.WithTimeout(context.Background(), time.Minute*10)
			_, err = c.Run(ctx, &pb.PostDeployRequest{Msg: b})

			if err != nil {
				logger.Logf("could not run post deploy: %v", err)
			}

			conn.Close()
			cancel()
		}
	}
}

func consume(d amqp.Delivery, w *worker.Worker) {
	d.Ack(true)
	dj, err := baghdad.DecodeDeployJob(d.Body)
	if err != nil {
		fmt.Printf("could not unmarshal deploy-sync message: %v\n", err)
		return
	}

	m.Lock()
	ch, ok := pc[dj.Baghdad.Project]
	if !ok {
		ch = make(deployCh)
		pc[dj.Baghdad.Project] = ch
		go deploy(w, ch)
	}
	m.Unlock()

	ch <- dj
}

func main() {
	utils.FailOnError(
		godotenv.Load("/run/secrets/baghdad-vars"),
		"could not get secrets file",
	)

	utils.ValidateEnvVars(getRequiredVars()...)

	w, err := worker.NewWorker(utils.GetAMQPURL())
	utils.FailOnError(err, "could not connect to rabbitmq")
	defer w.Close()

	err = w.EnsureQueues("deploy-sync")
	utils.FailOnError(err, "could not declare queues")

	err = w.EnsureExchanges("logs")
	utils.FailOnError(err, "could not declare exchange")

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

func listen(msgs <-chan amqp.Delivery, w *worker.Worker) {
	for d := range msgs {
		go consume(d, w)
	}
}
