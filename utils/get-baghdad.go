package utils

import (
	"context"

	"github.com/BurntSushi/toml"
	"github.com/google/go-github/github"
	"github.com/marwan-at-work/baghdad"
)

// GetBaghdad takes a github client, a sha, and returns the corresponding baghdad.toml file.
func GetBaghdad(c *github.Client, opts GetBaghdadOpts) (v baghdad.Baghdad, err error) {
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
		return
	}

	str, err := f.GetContent()
	if err != nil {
		return
	}

	b := baghdad.Baghdad{}

	_, err = toml.Decode(str, &b)
	return
}

// GetBaghdadOpts verbose options for get baghdad
type GetBaghdadOpts struct {
	SHA      string
	Owner    string
	RepoName string
	Ctx      context.Context
}
