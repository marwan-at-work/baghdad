package main

import (
	"context"
	"fmt"
	"os"

	"github.com/google/go-github/github"
	"github.com/marwan-at-work/baghdad/utils"
)

func updateGithubStatus(repoOwner, repoName, sha, statusState, logID string) error {
	gh := utils.GetGithub(os.Getenv("ADMIN_TOKEN"))
	ctx := context.Background()

	statusCtx := "Baghdad"
	statusDescription := "Baghdad build system."
	_, _, err := gh.Repositories.CreateStatus(ctx, repoOwner, repoName, sha, &github.RepoStatus{
		State:       &statusState,
		Context:     &statusCtx,
		Description: &statusDescription,
		TargetURL:   getTargetURL(logID),
	})

	return err
}

func getTargetURL(logID string) *string {
	url := fmt.Sprintf("http://master-cd-baghdad-api.marwan.io/status/%v", logID)

	return &url
}
