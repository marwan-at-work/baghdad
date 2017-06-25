package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/marwan-at-work/baghdad"
	pb "github.com/marwan-at-work/baghdad/services"
	"github.com/marwan-at-work/baghdad/utils"
	"github.com/marwan-at-work/baghdad/worker"
	"google.golang.org/grpc"
)

// ArtifactsPath where to save artifacts to upload them to github.
var ArtifactsPath = "/var/baghdad/artifacts"

func sendPushBuildJob(ctx context.Context, bj baghdad.BuildJob, w *worker.Worker, ch buildCh, logger *worker.Logger) (nextTag string, err error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	o := bj.RepoOwner
	r := bj.RepoName

	tags, err := getAllTags(ctx, o, r)
	if err != nil {
		err = fmt.Errorf("could not get all tags: %v", err)
		return
	}

	sprint := bj.Baghdad.Branches[bj.BranchName].Version
	nextTag, err = getNextTag(bj.BranchName, sprint, tags)
	if err != nil {
		err = fmt.Errorf("could not get next tag: %v", err)
		return
	}
	bj.NextTag = nextTag
	// move this to the events. Baghdad should over all send events along every step, and have the event worker
	// do most of the heavy lifting, potentially even the logger part. This way you don't have to define a logger anywhere,
	// and you can just have an event interface.
	slackURL := bj.Baghdad.SlackURL
	if slackURL != "" {
		utils.SendSlackMessage(slackURL, fmt.Sprintf("starting %v build job: %v", r, nextTag))
	}

	if len(bj.Baghdad.Services) == 0 {
		err = errors.New("project needs to have at least one service for deployment")
		return
	}

	if err != nil {
		err = fmt.Errorf("could not declare cancel queue: %v", err)
		return
	}

	numServices := 0
	respCh := make(chan buildResp)
	for _, s := range bj.Baghdad.Services {
		if s.IsExternal {
			continue
		}
		numServices++

		addr := os.Getenv("BUILDER_ADDR")
		var conn *grpc.ClientConn
		conn, err = grpc.Dial(addr, grpc.WithInsecure())
		if err != nil {
			err = fmt.Errorf("could not connect to service builder service: %v", err)
			cancel()
			return nextTag, err
		}
		defer conn.Close()

		c := pb.NewServiceBuilderClient(conn)

		go buildService(ctx, c, bj, s, respCh)
	}

	errs := []error{}
	brs := []buildResp{}
	for i := 0; i < numServices; i++ {
		if br := <-respCh; br.err != nil {
			cancel()
			errs = append(errs, br.err)
		} else if br.s.HasArtifacts {
			brs = append(brs, br)
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
		return nextTag, errors.New(errMsg)
	}

	releaseID, err := createRelease(bj, nextTag, 1)
	if err != nil {
		return
	}

	for _, br := range brs {
		err = uploadArtifact(ctx, o, r, br.fileName, br.distPath, releaseID, 5)
		if err != nil {
			err = fmt.Errorf("could not upload asset for %v: %v", br.s.Name, err)
			return
		}

		os.RemoveAll(br.distPath)
	}

	logger.Loglnf("successfully finished build job %v", nextTag)
	return
}
