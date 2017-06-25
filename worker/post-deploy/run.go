package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"google.golang.org/grpc"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	docker "github.com/docker/docker/client"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/term"
	"github.com/marwan-at-work/baghdad"
	pb "github.com/marwan-at-work/baghdad/services"
	"github.com/marwan-at-work/baghdad/utils"
	"github.com/marwan-at-work/baghdad/worker"
)

func run(
	ctx context.Context,
	pdj baghdad.PostDeployJob,
	logger *worker.Logger,
) (err error) {
	logger.Log("running post deploy job")
	dc, err := docker.NewEnvClient()
	if err != nil {
		return
	}

	authConfig := types.AuthConfig{
		Username: os.Getenv("DOCKER_AUTH_USER"),
		Password: os.Getenv("DOCKER_AUTH_PASS"),
	}

	_, err = dc.RegistryLogin(ctx, authConfig)
	if err != nil {
		return
	}

	img := fmt.Sprintf(
		"%v/%v-%v:%v",
		os.Getenv("DOCKER_ORG"),
		pdj.Baghdad.Project,
		pdj.Baghdad.PostDeploy.SourceService,
		pdj.Tag,
	)

	registryAuth, err := utils.EncodeAuthToBase64(authConfig)
	if err != nil {
		err = fmt.Errorf("could not encode auth: %v", err)
		return
	}

	imgOut, err := dc.ImagePull(ctx, img, types.ImagePullOptions{
		RegistryAuth: registryAuth,
	})
	if err != nil {
		err = fmt.Errorf("could not pull %v image: %v", img, err)
		return
	}
	defer imgOut.Close()
	fd, isTerm := term.GetFdInfo(logger)
	err = jsonmessage.DisplayJSONMessagesStream(imgOut, logger, fd, isTerm, nil)
	if err != nil {
		err = fmt.Errorf("could not pull %v image: %v", img, err)
		return
	}

	sec := pdj.Baghdad.PostDeploy.Secrets
	var data []byte
	if sec != "" {
		addr := os.Getenv("SECRET_ADDR")
		var conn *grpc.ClientConn
		conn, err = grpc.Dial(addr, grpc.WithInsecure())
		if err != nil {
			err = fmt.Errorf("could not connect to service builder service: %v", err)
			return
		}
		defer conn.Close()

		c := pb.NewSecretClient(conn)

		var resp *pb.SecretResponse
		resp, err = c.Get(ctx, &pb.SecretRequest{Name: sec})
		if err != nil {
			err = fmt.Errorf("could not get secret from secret worker: %v", err)
			return err
		}

		data = resp.GetData()
	}

	env := []string{
		fmt.Sprintf("SITE_URL=%v", pdj.SiteURL),
		fmt.Sprintf("PROJECT_NAME=%v", pdj.ProjectName),
		fmt.Sprintf("DEPLOYED_TAG=%v", pdj.Tag),
		fmt.Sprintf("DEPLOYED_ENV=%v", pdj.Env),
		fmt.Sprintf("BRANCH_NAME=%v", pdj.BranchName),
	}

	if sec != "" {
		env = append(env, fmt.Sprintf("%v=%v", sec, string(data)))
	}

	tenMinutes := 10 * 60
	ct, err := dc.ContainerCreate(
		ctx,
		&container.Config{
			Image:       img,
			StopTimeout: &tenMinutes,
			Env:         env,
		},
		&container.HostConfig{
			AutoRemove: false,
		},
		nil,
		"",
	)
	if err != nil {
		err = fmt.Errorf("could not create container %v: %v", img, err)
		return
	}

	if err = dc.ContainerStart(ctx, ct.ID, types.ContainerStartOptions{}); err != nil {
		return
	}

	out, err := dc.ContainerLogs(ctx, ct.ID, types.ContainerLogsOptions{
		ShowStdout: true,
		Follow:     true,
		ShowStderr: true,
	})
	if err != nil {
		fmt.Println("could not get container logs, but job is still running")
	}
	defer out.Close()

	io.Copy(logger, out)
	if _, err = dc.ContainerWait(ctx, ct.ID); err != nil {
		err = fmt.Errorf("container run failed %v: %v", img, err)
		return
	}

	dc.ContainersPrune(ctx, filters.Args{})

	return
}
