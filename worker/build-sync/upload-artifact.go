package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/google/go-github/github"
	"github.com/marwan-at-work/baghdad/utils"
)

func uploadArtifact(ctx context.Context, repoOwner, repoName, name, tarGzPath string, releaseID, retries int) error {
	gh := utils.GetGithub(os.Getenv("ADMIN_TOKEN"))
	f, err := os.Open(tarGzPath)
	if err != nil {
		return err
	}

	_, _, err = gh.Repositories.UploadReleaseAsset(ctx, repoOwner, repoName, releaseID, &github.UploadOptions{
		Name: name,
	}, f)

	if err != nil && retries > 0 {
		fmt.Println("artifact upload errored out, retrying in one second. Err:", err)
		time.Sleep(time.Second)
		err = uploadArtifact(ctx, repoOwner, repoName, name, tarGzPath, releaseID, retries-1)
	}

	return err
}
