package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/docker/docker/api/types"
	docker "github.com/docker/docker/client"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/term"
	"github.com/marwan-at-work/baghdad/utils"
)

type pushImageOpts struct {
	c       *docker.Client
	imgName string
	stdout  io.Writer
}

func pushImage(ctx context.Context, opts *pushImageOpts) (err error) {
	authConfig := types.AuthConfig{
		Username: os.Getenv("DOCKER_AUTH_USER"),
		Password: os.Getenv("DOCKER_AUTH_PASS"),
	}

	c := opts.c

	_, err = c.RegistryLogin(ctx, authConfig)
	if err != nil {
		err = fmt.Errorf("could not login to docker registry: %v", err)
		return
	}

	registryAuth, err := utils.EncodeAuthToBase64(authConfig)
	if err != nil {
		err = fmt.Errorf("could not encode auth: %v", err)
		return err
	}

	resp, err := opts.c.ImagePush(ctx, opts.imgName, types.ImagePushOptions{
		RegistryAuth: registryAuth,
	})
	defer resp.Close()
	if err != nil {
		return
	}

	fd, isTerm := term.GetFdInfo(opts.stdout)
	err = jsonmessage.DisplayJSONMessagesStream(resp, opts.stdout, fd, isTerm, nil)
	if err != nil {
		fmt.Println("could not display output")
		err = nil
	}
	return
}
