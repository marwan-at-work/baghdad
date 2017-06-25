package utils

import (
	"context"
	"errors"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/google/go-github/github"
	"github.com/marwan-at-work/baghdad"
	"github.com/marwan-at-work/dfparser"
)

// GetBaghdadOpts verbose options for get baghdad
type GetBaghdadOpts struct {
	SHA      string
	Owner    string
	RepoName string
	Ctx      context.Context
}

// GetBaghdad takes a github client, a sha, and returns the corresponding baghdad.toml file.
func GetBaghdad(c *github.Client, opts GetBaghdadOpts) (b baghdad.Baghdad, err error) {
	f, _, _, err := c.Repositories.GetContents(
		opts.Ctx,
		opts.Owner,
		opts.RepoName,
		"baghdad.toml",
		&github.RepositoryContentGetOptions{
			Ref: opts.SHA,
		},
	)
	if err != nil {
		return getBaghdadFromDockerfile(c, opts)
	}

	str, err := f.GetContent()
	if err != nil {
		return
	}

	b = baghdad.Baghdad{}
	_, err = toml.Decode(str, &b)

	return
}

func getBaghdadFromDockerfile(c *github.Client, opts GetBaghdadOpts) (b baghdad.Baghdad, err error) {
	f, _, _, err := c.Repositories.GetContents(
		opts.Ctx,
		opts.Owner,
		opts.RepoName,
		"Dockerfile",
		&github.RepositoryContentGetOptions{
			Ref: opts.SHA,
		},
	)
	if err != nil {
		return
	}

	str, err := f.GetContent()
	if err != nil {
		return
	}

	r := strings.NewReader(str)
	dockerfile, err := dfparser.Parse(r)
	if err != nil {
		return
	}

	if len(dockerfile.Expose) == 0 {
		return b, errors.New("Dockerfile must have an Expose command")
	}

	b = baghdad.Baghdad{
		Project:      opts.RepoName,
		Environments: map[string]string{"cd": "manual"},
		Branches:     map[string]baghdad.Branch{"master": baghdad.Branch{Version: "0.0.0"}},
		Services: []baghdad.Service{{
			Name:       "main",
			Dockerfile: "Dockerfile",
			IsExposed:  true,
			Port:       dockerfile.Expose[0],
		}},
	}

	return
}
