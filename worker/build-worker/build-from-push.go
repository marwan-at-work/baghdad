package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	docker "github.com/docker/docker/client"
	"github.com/marwan-at-work/baghdad"
	"github.com/marwan-at-work/baghdad/worker"
)

func buildFromPush(ctx context.Context, b baghdad.BuildJob, logger *worker.Logger) (repoPath string, closeFunc func() error, rc io.ReadCloser, err error) {
	r := b.RepoName
	err = cloneRepo(b, logger)
	if err != nil {
		logger.Loglnf("err: could not clone repo: %v", err)
		return
	}

	nextTag := b.NextTag

	c, err := docker.NewEnvClient()
	if err != nil {
		err = fmt.Errorf("could not get docker client: %v", err)
		return
	}

	dockerOrg := os.Getenv("DOCKER_ORG")

	repoPath = getRepoPath(strconv.Itoa(b.Type), r, nextTag)

	service := b.Service
	imgName := fmt.Sprintf("%v/%v-%v:%v", dockerOrg, r, service.Name, nextTag)
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

	err = pushImage(ctx, &pushImageOpts{
		c:       c,
		imgName: imgName,
		stdout:  logger,
	})
	if err != nil {
		err = fmt.Errorf("could not push img to repo: %v", err)
		return
	}

	if service.HasArtifacts {
		ctr, err := c.ContainerCreate(ctx, &container.Config{
			Image: imgName,
		}, &container.HostConfig{
			AutoRemove: false,
		}, nil, "")
		if err != nil {
			err = fmt.Errorf("could not create container: %v", err)
			return "", nil, nil, err
		}

		rc, _, err = c.CopyFromContainer(ctx, ctr.ID, service.ArtifactsPath)
		if err != nil {
			err = fmt.Errorf("could not copy from container: %v", err)
			return "", nil, nil, err
		}

		closeFunc = func() error {
			c.ContainerRemove(context.Background(), ctr.ID, types.ContainerRemoveOptions{
				RemoveVolumes: true,
				RemoveLinks:   true,
				Force:         true,
			})

			rc.Close()

			return nil
		}
	}

	// clean up previous builds of this service. But keep the current one, for future builds to be cached.
	imgPrefix := fmt.Sprintf("%v/%v-%v:", dockerOrg, r, service.Name)
	imgs, _ := c.ImageList(ctx, types.ImageListOptions{})
	for _, img := range imgs {
		for _, t := range img.RepoTags {
			if (strings.HasPrefix(t, imgPrefix) && t != imgName) || t == "<none>:<none>" {
				removeChan := make(chan struct{}, 1)
				imageRemovalContext, cancelImgRemove := context.WithCancel(ctx)
				go func() {
					c.ImageRemove(imageRemovalContext, img.ID, types.ImageRemoveOptions{
						Force: true,
					})

					removeChan <- struct{}{}
				}()

				select {
				case <-removeChan:
					cancelImgRemove()
				case <-time.After(time.Second * 3):
					go func() {
						time.Sleep(time.Second * 7)
						cancelImgRemove()
					}()
				}
			}
		}
	}

	return
}
