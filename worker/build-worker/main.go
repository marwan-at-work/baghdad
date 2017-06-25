package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/marwan-at-work/baghdad"
	pb "github.com/marwan-at-work/baghdad/services"
	"github.com/marwan-at-work/baghdad/utils"
	"github.com/marwan-at-work/baghdad/worker"
	"github.com/streadway/amqp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	port = ":50051"
)

type server struct{}

type streamWriter struct {
	stream pb.ServiceBuilder_ServiceBuildServer
}

func (s *streamWriter) Write(p []byte) (int, error) {
	err := s.stream.Send(&pb.BuildResponse{Artifacts: p})
	return len(p), err
}

var w *worker.Worker

// AdminToken is the token the worker uses to perform github API operations
var AdminToken string

// WorkPath is the directory where the build worker clones, builds & tars projects.
const WorkPath = "/var/baghdad/workdir"

// ServiceBuild buidls a service to a docker image and pushes to repo-hosting service.
func (s *server) ServiceBuild(req *pb.BuildRequest, stream pb.ServiceBuilder_ServiceBuildServer) (err error) {
	bj, err := baghdad.DecodeBuildJob(req.GetMsg())
	if err != nil {
		err = fmt.Errorf("could not decode build job: %v", err)
		return
	}

	srvName := fmt.Sprintf("%v-%v", bj.Baghdad.Project, bj.Service.Name)
	logger := worker.NewLogger(srvName, bj.LogID, w)

	var closerFunc func() error
	var repoPath string
	var rc io.ReadCloser

	switch bj.Type {
	case baghdad.PushEvent:
		repoPath, closerFunc, rc, err = buildFromPush(stream.Context(), bj, logger)
	case baghdad.PullRequestEvent:
		err = buildFromPR(bj, logger)
	default:
		logger.Loglnf("err: build type unrecognized: %v", bj.Type)
	}

	if err != nil {
		fmt.Println("stopping the world", err)
		os.RemoveAll(repoPath)
	}

	if rc != nil {
		defer closerFunc()
		r := bufio.NewReader(rc)
		// stream back the artifacts
		r.WriteTo(&streamWriter{stream: stream})
	}

	return
}

func main() {
	godotenv.Load("/run/secrets/baghdad-vars")
	utils.ValidateEnvVars(getRequiredVars()...)
	AdminToken = os.Getenv("ADMIN_TOKEN")
	ensureBuildPath()
	var err error
	w, err = worker.NewWorker(utils.GetAMQPURL())
	utils.FailOnError(err, "could not dial amqp")
	defer w.Close()

	err = w.EnsureExchanges("logs", "remove-repo")
	utils.FailOnError(err, "could not declare exchanges")

	q, err := w.QueueDeclare(worker.QueueOpts{
		Name:       "", // create a unique name for this subscriber.
		Durable:    true,
		AutoDelete: false,
		Exclusive:  true,
		NoWait:     false,
		Args:       nil,
	})
	utils.FailOnError(err, "could not declare event queue")

	err = w.Ch.QueueBind(q.Name, "", "remove-repo", false, nil)
	utils.FailOnError(err, "could not bind to event queue")

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

	go listen(msgs, w)

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterServiceBuilderServer(s, &server{})
	reflection.Register(s)
	fmt.Println("listening for grpc build requests")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func listen(msgs <-chan amqp.Delivery, w *worker.Worker) {
	for d := range msgs {
		go consume(d, w)
	}
}

func consume(d amqp.Delivery, w *worker.Worker) {
	d.Ack(false)
	bj, err := baghdad.DecodeBuildJob(d.Body)
	if err != nil {
		fmt.Println(fmt.Sprintf("could not unmarshal deploy-sync message: %v", err))
		d.Ack(false)
		return
	}

	repoPath := getRepoPath(strconv.Itoa(bj.Type), bj.RepoName, bj.NextTag)
	err = os.RemoveAll(repoPath)
	if err != nil {
		fmt.Println("could not delete repo at ", repoPath, err)
	}
}
