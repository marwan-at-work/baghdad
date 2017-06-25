package main

import (
	"context"
	"fmt"
	"io"

	"github.com/docker/docker/api/types"
	docker "github.com/docker/docker/client"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/term"
	"github.com/marwan-at-work/baghdad/utils"
)

type buildImgOpts struct {
	c              *docker.Client
	imgName        string
	repoPath       string
	dockerfilePath string
	stdout         io.Writer
}

func buildImage(ctx context.Context, opts *buildImgOpts) (err error) {
	dockerClient := opts.c
	body, err := utils.CreateTar(opts.repoPath, opts.dockerfilePath)
	if err != nil {
		return
	}

	resp, err := dockerClient.ImageBuild(ctx, body, types.ImageBuildOptions{
		Dockerfile: opts.dockerfilePath,
		Tags:       []string{opts.imgName},
	})

	if err != nil {
		err = fmt.Errorf("could not build image: %v", err)
		return
	}
	defer resp.Body.Close()

	fd, isTerm := term.GetFdInfo(opts.stdout)
	err = jsonmessage.DisplayJSONMessagesStream(resp.Body, opts.stdout, fd, isTerm, nil)

	return
}
