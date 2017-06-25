package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/marwan-at-work/baghdad"
	pb "github.com/marwan-at-work/baghdad/services"
)

type buildResp struct {
	err      error
	s        baghdad.Service
	distPath string
	fileName string
}

func buildService(
	ctx context.Context,
	c pb.ServiceBuilderClient,
	bj baghdad.BuildJob,
	s baghdad.Service,
	respCh chan buildResp,
) {
	bj.Service = s
	bts, _ := baghdad.EncodeBuildJob(bj)

	br := buildResp{s: s}

	stream, err := c.ServiceBuild(ctx, &pb.BuildRequest{Msg: bts})
	if err != nil {
		err = fmt.Errorf("%v service could not be built: %v", s.Name, err)
		br.err = err
		respCh <- br
		return
	}

	if bj.Type == baghdad.PullRequestEvent {
		_, err = stream.Recv()
		if err != io.EOF {
			br.err = err
		}

		respCh <- br
		return
	}

	if s.HasArtifacts {
		folderPath := filepath.Join(ArtifactsPath, bj.RepoName, bj.NextTag)
		os.MkdirAll(folderPath, 0666)

		fileName := br.s.Name + ".tar"
		distPath := filepath.Join(folderPath, fileName)
		br.distPath = distPath
		br.fileName = fileName
		f, err := os.Create(distPath)
		if err != nil {
			br.err = err
			respCh <- br
			return
		}
		defer f.Close()

		for {
			r, err := stream.Recv()
			if err == io.EOF {
				break
			} else if err != nil {
				br.err = err
				break
			}

			_, err = f.Write(r.Artifacts)
			if err != nil {
				fmt.Println("got err write", err)
			}
		}
	} else {
		_, br.err = stream.Recv()
		if br.err == io.EOF {
			br.err = nil
		}
	}

	respCh <- br
}
