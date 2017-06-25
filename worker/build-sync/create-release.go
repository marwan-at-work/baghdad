package main

import (
	"context"
	"fmt"
	"os"

	"github.com/google/go-github/github"
	"github.com/marwan-at-work/baghdad"
	"github.com/marwan-at-work/baghdad/utils"
)

func createRelease(b baghdad.BuildJob, nextTag string, retries int) (releaseID int, err error) {
	ctx := context.Background()
	repoName := b.RepoName
	repoOwner := b.RepoOwner
	sha := b.SHA
	gh := utils.GetGithub(os.Getenv("ADMIN_TOKEN"))
	tagType := "commit"
	msg := "automated-baghdad-tags"
	_, _, err = gh.Git.CreateTag(ctx, repoOwner, repoName, &github.Tag{
		Tag: &nextTag,
		SHA: &sha,
		Object: &github.GitObject{
			SHA:  &sha,
			Type: &tagType,
		},
		Message: &msg,
	})

	if err != nil {
		err = fmt.Errorf("%v err: could not create tag: %v", repoName, err)
		if retries > 0 {
			return createRelease(b, nextTag, retries-1)
		}

		return
	}

	ref := "tags/" + nextTag
	_, _, err = gh.Git.CreateRef(ctx, repoOwner, repoName, &github.Reference{
		Ref: &ref,
		Object: &github.GitObject{
			SHA:  &sha,
			Type: &tagType,
		},
	})
	if err != nil {
		err = fmt.Errorf("%v err: could not create ref: %v", repoName, err)
		if retries > 0 {
			return createRelease(b, nextTag, retries-1)
		}
		return
	}

	fp := false
	pre := true
	r, _, err := gh.Repositories.CreateRelease(ctx, repoOwner, repoName, &github.RepositoryRelease{
		TagName:         &nextTag,
		TargetCommitish: &sha,
		Name:            &nextTag,
		Body:            &ref,
		Draft:           &fp,
		Prerelease:      &pre,
	})
	if err != nil {
		err = fmt.Errorf("%v err: could not create prerelease: %v", repoName, err)
		if retries > 0 {
			return createRelease(b, nextTag, retries-1)
		}
		return
	}

	releaseID = r.GetID()
	return
}
