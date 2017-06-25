package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"golang.org/x/net/context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
	docker "github.com/docker/docker/client"
	"github.com/joho/godotenv"
	"github.com/marwan-at-work/baghdad"
	pb "github.com/marwan-at-work/baghdad/services"
	"github.com/marwan-at-work/baghdad/utils"
	"github.com/marwan-at-work/baghdad/worker"
	"github.com/streadway/amqp"
)

const (
	port = ":50051"
)

type server struct{}

// Get implmenets the Secret Get gRPC
func (s *server) Get(ctx context.Context, req *pb.SecretRequest) (res *pb.SecretResponse, err error) {
	bts, err := ioutil.ReadFile(fmt.Sprintf("/run/secrets/%v", req.GetName()))
	if err != nil {
		return nil, fmt.Errorf("%v secret does not exist", req.GetName())
	}

	res = &pb.SecretResponse{
		Exists: true,
		Data:   bts,
	}

	return
}

func consume(d amqp.Delivery, w *worker.Worker) {
	sj, err := baghdad.DecodeSecretsJob(d.Body)
	if err != nil {
		fmt.Printf("could not marshal secrets job: %v\n", err)
		d.Ack(false)
		return
	}

	logger := worker.NewLogger(sj.ProjectName, "", w)
	logger.Log("adding secret")

	dc, err := docker.NewClient(os.Getenv("DOCKER_REMOTE_API_URL"), "1.29", nil, nil)
	if err != nil {
		logger.Loglnf("could not get docker client: %v", err)
		d.Ack(false)
		return
	}

	ctx := context.Background()
	scrt, err := dc.SecretCreate(ctx, swarm.SecretSpec{
		Annotations: swarm.Annotations{Name: sj.SecretName},
		Data:        sj.SecretBody,
	})

	if err != nil {
		logger.Loglnf("could not create docker secret: %v", err)
		d.Ack(false)
		return
	}

	// get the deployed environment then construct this.
	serviceName := "baghdad_cd_secrets-worker"
	service, _, err := dc.ServiceInspectWithRaw(ctx, serviceName, types.ServiceInspectOptions{})
	if err != nil {
		logger.Loglnf("could not inspect service %v", serviceName)
		d.Ack(false)
		return
	}

	service.Spec.TaskTemplate.ContainerSpec.Secrets = append(
		service.Spec.TaskTemplate.ContainerSpec.Secrets,
		&swarm.SecretReference{
			SecretID:   scrt.ID,
			SecretName: sj.SecretName,
			File: &swarm.SecretReferenceFileTarget{
				GID:  "0",
				Mode: 0444,
				UID:  "0",
				Name: sj.SecretName,
			},
		},
	)

	go func() {
		logger.Log("updating secrets worker to store your secret")
		_, err := dc.ServiceUpdate(ctx, service.ID, service.Version, service.Spec, types.ServiceUpdateOptions{})
		if err != nil {
			logger.Loglnf("could not update service: %v", err)
		}

		logger.Log("secret worker updated")
	}()

	logger.Log("secret created")
	d.Ack(false)
}

func listen(msgs <-chan amqp.Delivery, w *worker.Worker) {
	for d := range msgs {
		go consume(d, w)
	}
}

func main() {
	godotenv.Load("/run/secrets/baghdad-vars")
	utils.ValidateEnvVars(getRequiredVars()...)

	w, err := worker.NewWorker(utils.GetAMQPURL())
	utils.FailOnError(err, "could not dial amqp")
	defer w.Close()

	err = w.EnsureQueues("secrets")
	utils.FailOnError(err, "could not declare queues")
	err = w.EnsureExchanges("logs")
	utils.FailOnError(err, "could not declare exchanges")

	msgs, err := w.Consume(worker.ConsumeOpts{
		Queue:     "secrets",
		Consumer:  "",
		AutoAck:   false,
		Exclusive: false,
		NoLocal:   false,
		NoWait:    false,
		Args:      nil,
	})
	utils.FailOnError(err, "could not consume")

	go listen(msgs, w)

	fmt.Println("[x] - Listening for messages on the secrets queue")

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterSecretServer(s, &server{})
	reflection.Register(s)
	fmt.Println("listening for grpc build requests")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
