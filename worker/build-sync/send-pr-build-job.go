package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"google.golang.org/grpc"

	"github.com/marwan-at-work/baghdad"
	pb "github.com/marwan-at-work/baghdad/services"
	"github.com/marwan-at-work/baghdad/worker"
)

func sendPRBuildJob(ctx context.Context, bj baghdad.BuildJob, w *worker.Worker, ch buildCh, logger *worker.Logger) (err error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	numServices := 0
	addr := os.Getenv("BUILDER_ADDR")
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		err = fmt.Errorf("could not connect to service builder service: %v", err)
		return
	}
	defer conn.Close()

	c := pb.NewServiceBuilderClient(conn)
	respCh := make(chan buildResp)
	for _, s := range bj.Baghdad.Services {
		if s.IsExternal {
			continue
		}
		numServices++

		go buildService(ctx, c, bj, s, respCh)
	}

	errs := []error{}
	for i := 0; i < numServices; i++ {
		if br := <-respCh; br.err != nil {
			cancel()
			errs = append(errs, br.err)
		}
	}

	ss := []string{}
	for _, e := range errs {
		if err != nil {
			ss = append(ss, e.Error())
		}
	}

	errMsg := strings.Join(ss, "\n")
	if len(errs) > 0 {
		return errors.New(errMsg)
	}

	logger.Loglnf("successfully built %v", bj.PRNum)
	return
}
