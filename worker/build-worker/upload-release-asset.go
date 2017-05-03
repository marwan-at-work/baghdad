package main

import (
	"context"
	"os"

	"github.com/google/go-github/github"
	"github.com/marwan-at-work/baghdad/utils"
)

func uploadReleaseAsset(ctx context.Context, repoOwner, repoName, name, tarGzPath string, releaseID int) error {
	gh := utils.GetGithub(AdminToken)
	f, err := os.Open(tarGzPath)
	if err != nil {
		return err
	}

	_, _, err = gh.Repositories.UploadReleaseAsset(ctx, repoOwner, repoName, releaseID, &github.UploadOptions{
		Name: name,
	}, f)

	return err
}
