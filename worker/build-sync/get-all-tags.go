package main

import (
	"context"
	"os"

	"github.com/google/go-github/github"
	"github.com/marwan-at-work/baghdad/utils"
)

func getAllTags(ctx context.Context, owner, repo string) (tags []string, err error) {
	c := utils.GetGithub(os.Getenv("ADMIN_TOKEN"))
	allTags := []*github.RepositoryTag{}
	page := 1
	for {
		paginatedTags, _, err := c.Repositories.ListTags(context.Background(), owner, repo, &github.ListOptions{
			PerPage: 100,
			Page:    page,
		})

		if err != nil {
			return tags, err
		}

		if len(paginatedTags) == 0 {
			break
		}

		allTags = append(allTags, paginatedTags...)
		page++
	}

	for _, t := range allTags {
		tags = append(tags, t.GetName())
	}

	return
}
