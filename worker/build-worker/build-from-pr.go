package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	docker "github.com/docker/docker/client"
	"github.com/marwan-at-work/baghdad"
	"github.com/marwan-at-work/baghdad/worker"

	"github.com/docker/docker/api/types/filters"
)

func buildFromPR(b baghdad.BuildJob, w *worker.Worker) (err error) {
	r := b.RepoName
	sha := b.SHA
	o := b.RepoOwner

	err = updateGithubStatus(o, r, sha, "pending")
	if err != nil {
		w.Log(fmt.Sprintf("%v err: could not update status: %v", r, err))
	}

	prNum := strconv.Itoa(b.PRNum)
	_, err = cloneRepo(b)
	if err != nil {
		w.Log(fmt.Sprintf("%v err: could not clone repo: %v", r, err))
		return
	}

	c, err := docker.NewEnvClient()
	if err != nil {
		err = fmt.Errorf("could not get docker client: %v", err)
		updateGithubStatus(o, r, sha, "error")
		return
	}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	c.ContainersPrune(ctx, filters.Args{})

	dockerOrg := os.Getenv("DOCKER_ORG")

	if len(b.Baghdad.Services) == 0 {
		updateGithubStatus(o, r, sha, "error")
		return errors.New("project needs to have at least one service for deployment")
	}

	repoPath := getRepoPath(r, strconv.Itoa(b.Type))
	errChan := make(chan error)

	internalServices := 0
	for _, service := range b.Baghdad.Services {
		if service.IsExternal {
			continue
		}

		internalServices++
		go func(service baghdad.Service) {
			imgName := fmt.Sprintf("%v/%v-%v", dockerOrg, r, service.Name)
			er := buildImage(ctx, &buildImgOpts{
				c:              c,
				imgName:        imgName,
				repoPath:       repoPath,
				dockerfilePath: service.Dockerfile,
				stdout:         w,
			})
			if er != nil {
				err = fmt.Errorf("could not build image for %v: %v", service.Name, er)
				errChan <- er
				return
			}

			errChan <- nil
		}(service)
	}

	for i := 0; i < internalServices; i++ {
		select {
		case err = <-errChan:
			if err != nil {
				break
			}
		case <-time.After(time.Minute * 5):
			err = errors.New("building services timed out")
		}
	}

	if err != nil {
		cancel()
		updateGithubStatus(o, r, sha, "error")
		w.Log(err.Error())
		return err
	}

	rmErr := os.RemoveAll(repoPath)
	if rmErr != nil {
		fmt.Println("could not remove repo:", repoPath, rmErr)
	}

	updateErr := updateGithubStatus(o, r, sha, "success")
	if updateErr != nil {
		w.Log(fmt.Sprintf("%v: built was successful but could not update github status %v", r, prNum))
	}

	w.Log(fmt.Sprintf("%v: successfully built %v", r, prNum))

	return
}
