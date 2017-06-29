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
		TargetURL:   getTargetURL(repoOwner, logID),
	})

	return err
}

func getTargetURL(repoOwner, logID string) *string {
	isDev := os.Getenv("GO_ENV") == "development"
	var url string
	if isDev {
		url = fmt.Sprintf("http://localhost:3000/projects/%v/logs/%v", repoOwner, logID)
	} else {
		branch := os.Getenv("BAGHDAD_BUILD_BRANCH")
		env := os.Getenv("BAGHDAD_BUILD_ENV")
		domain := os.Getenv("BAGHDAD_DOMAIN_NAME")

		url = fmt.Sprintf(
			"http://%v-%v-baghdad-api.%v/projects/%v/logs/%v",
			branch,
			env,
			domain,
			repoOwner,
			logID,
		)
	}

	return &url
}
