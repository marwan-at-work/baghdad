package main

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"time"

	docker "github.com/docker/docker/client"
	"github.com/marwan-at-work/baghdad"
	"github.com/marwan-at-work/baghdad/worker"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
)

func buildFromPush(b baghdad.BuildJob, w *worker.Worker) (tag string, err error) {
	r := b.RepoName
	sha := b.SHA
	o := b.RepoOwner

	err = updateGithubStatus(o, r, sha, "pending")
	if err != nil {
		w.Log(fmt.Sprintf("%v err: could not update status: %v", r, err))
	}

	repo, err := cloneRepo(b)
	if err != nil {
		w.Log(fmt.Sprintf("%v err: could not clone repo: %v", r, err))
		return
	}

	nextTag, err := getNextTag(b, repo)
	if err != nil {
		w.Log(fmt.Sprintf("%v err: could not get next versioning tag: %v", r, err))
		updateGithubStatus(o, r, sha, "error")
		return
	}

	releaseID, err := createRelease(b, nextTag, 1)
	if err != nil {
		w.Log(fmt.Sprintf("%v err: could not create release: %v", r, err))
		updateGithubStatus(o, r, sha, "error")
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
		return nextTag, errors.New("project needs to have at least one service for deployment")
	}

	repoPath := getRepoPath(r, strconv.Itoa(b.Type))
	errChan := make(chan error)

	for _, service := range b.Baghdad.Services {
		if service.IsExternal {
			continue
		}

		go func(service baghdad.Service) {
			imgName := fmt.Sprintf("%v/%v-%v:%v", dockerOrg, r, service.Name, nextTag)
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

			er = pushImage(ctx, &pushImageOpts{
				c:       c,
				imgName: imgName,
				stdout:  w,
			})
			if er != nil {
				er = fmt.Errorf("could not push img to repo: %v", err)
				errChan <- er
				return
			}

			ctr, er := c.ContainerCreate(ctx, &container.Config{
				Image: imgName,
			}, &container.HostConfig{
				AutoRemove: false,
			}, nil, "")
			if er != nil {
				er = fmt.Errorf("could not create container: %v", err)
				errChan <- er
				return
			}

			srcPath, er := getDockerfileWorkdir(filepath.Join(repoPath, service.Dockerfile))
			if er != nil {
				er = fmt.Errorf("could not get dockerfile workdir: %v", err)
				errChan <- er
				return
			}

			rc, _, er := c.CopyFromContainer(ctx, ctr.ID, srcPath)
			if er != nil {
				er = fmt.Errorf("could not copy from container: %v", err)
				errChan <- er
				return
			}

			bts, er := ioutil.ReadAll(rc)
			if er != nil {
				er = fmt.Errorf("could not read bytes from container: %v", err)
				errChan <- er
				return
			}

			folderPath := filepath.Join(WorkPath, "artifacts", r)
			os.MkdirAll(folderPath, 0666)

			fileName := service.Name + ".tar"
			distPath := filepath.Join(folderPath, fileName)
			er = ioutil.WriteFile(distPath, bts, 0666)
			if er != nil {
				er = fmt.Errorf("could not write tar: %v", err)
				errChan <- er
				return
			}

			er = uploadReleaseAsset(ctx, o, r, fileName, distPath, releaseID)
			if er != nil {
				er = fmt.Errorf("could not upload asset for %v: %v", service.Name, err)
				errChan <- er
				return
			}

			er = os.RemoveAll(distPath)
			if er != nil {
				fmt.Println("could not remove", distPath)
			}

			errChan <- nil
		}(service)
	}

	select {
	case err = <-errChan:
	case <-time.After(time.Minute * 5):
		err = errors.New("building services timed out")
	}

	if err != nil {
		cancel()
		updateGithubStatus(o, r, sha, "error")
		return nextTag, err
	}

	rmErr := os.RemoveAll(repoPath)
	if rmErr != nil {
		fmt.Println("could not remove repo:", repoPath, rmErr)
	}

	updateErr := updateGithubStatus(o, r, sha, "success")
	if updateErr != nil {
		w.Log(fmt.Sprintf("%v: built was successful but could not update github status %v", r, nextTag))
	}

	w.Log(fmt.Sprintf("%v: successfully built %v", r, nextTag))

	return
}
