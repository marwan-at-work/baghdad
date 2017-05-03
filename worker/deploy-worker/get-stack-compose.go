package main

import (
	"context"

	"github.com/google/go-github/github"
)

func getStackCompose(c *github.Client, opts stackComposeGetOpts) (s string, err error) {
	f, _, _, err := c.Repositories.GetContents(
		opts.Ctx,
		opts.Owner,
		opts.RepoName,
		"stack-compose.yml",
		&github.RepositoryContentGetOptions{
			Ref: opts.Tag,
		},
	)
	if err != nil {
		return
	}

	s, err = f.GetContent()
	return
}

type stackComposeGetOpts struct {
	Ctx      context.Context
	Owner    string
	RepoName string
	Tag      string
}
