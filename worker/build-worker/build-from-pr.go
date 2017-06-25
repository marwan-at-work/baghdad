package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"

	docker "github.com/docker/docker/client"
	"github.com/marwan-at-work/baghdad"
	"github.com/marwan-at-work/baghdad/worker"

	"github.com/docker/docker/api/types/filters"
)

func buildFromPR(b baghdad.BuildJob, logger *worker.Logger) (err error) {
	r := b.RepoName

	prNum := strconv.Itoa(b.PRNum)
	err = cloneRepo(b, logger)
	if err != nil {
		logger.Loglnf("err: could not clone repo: %v", err)
		return
	}

	c, err := docker.NewEnvClient()
	if err != nil {
		err = fmt.Errorf("could not get docker client: %v", err)
		return
	}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	c.ContainersPrune(ctx, filters.Args{})

	dockerOrg := os.Getenv("DOCKER_ORG")

	if len(b.Baghdad.Services) == 0 {
		return errors.New("project needs to have at least one service for deployment")
	}

	repoPath := getRepoPath(strconv.Itoa(b.Type), r, "")

	service := b.Service
	imgName := fmt.Sprintf("%v/%v-%v:pr-%v", dockerOrg, r, service.Name, prNum)
	err = buildImage(ctx, &buildImgOpts{
		c:              c,
		imgName:        imgName,
		repoPath:       repoPath,
		dockerfilePath: service.Dockerfile,
		stdout:         logger,
	})
	if err != nil {
		err = fmt.Errorf("could not build image for %v: %v", service.Name, err)
		return
	}

	return
}
