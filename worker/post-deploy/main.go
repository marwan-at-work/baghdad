package main

import (
	"fmt"
	"log"
	"net"

	"github.com/joho/godotenv"
	"github.com/marwan-at-work/baghdad"
	pb "github.com/marwan-at-work/baghdad/services"
	"github.com/marwan-at-work/baghdad/utils"
	"github.com/marwan-at-work/baghdad/worker"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	port = ":50051"
)

type server struct {
	w *worker.Worker
}

func (s *server) Run(ctx context.Context, req *pb.PostDeployRequest) (pdr *pb.PostDeployResponse, err error) {
	pdr = &pb.PostDeployResponse{}
	pdj, err := baghdad.DecodePostDeployJob(req.GetMsg())
	if err != nil {
		err = fmt.Errorf("could not decode build job: %v", err)
		return
	}

	logger := worker.NewLogger(pdj.Baghdad.Project, "", s.w)

	err = run(ctx, pdj, logger)
	if err != nil {
		logger.Loglnf("error: %v", err)
	}
	return
}

func main() {
	utils.FailOnError(
		godotenv.Load("/run/secrets/baghdad-vars"),
		"could not get secrets file",
	)

	utils.ValidateEnvVars(getRequiredVars()...)
	w, err := worker.NewWorker(utils.GetAMQPURL())
	utils.FailOnError(err, "could not start worker")

	utils.FailOnError(w.EnsureExchanges("logs"), "could not create logs exchange")

	lis, err := net.Listen("tcp", port)
	utils.FailOnError(err, "failed to listen")

	s := grpc.NewServer()
	pb.RegisterPostDeployServer(s, &server{w: w})
	reflection.Register(s)
	fmt.Println("listening for grpc post-deploy requests")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
