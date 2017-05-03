package main

import (
	"context"

	"github.com/google/go-github/github"
	"github.com/marwan-at-work/baghdad/utils"
)

func updateGithubStatus(repoOwner, repoName, sha, statusState string) error {
	gh := utils.GetGithub(AdminToken)
	ctx := context.Background()

	statusCtx := "Baghdad"
	statusDescription := "Baghdad build system."
	_, _, err := gh.Repositories.CreateStatus(ctx, repoOwner, repoName, sha, &github.RepoStatus{
		State:       &statusState,
		Context:     &statusCtx,
		Description: &statusDescription,
	})

	return err
}
