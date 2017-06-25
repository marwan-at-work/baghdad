package main

import (
	"context"
	"fmt"
	"os"

	"github.com/google/go-github/github"
	"github.com/marwan-at-work/baghdad/utils"
	yaml "gopkg.in/yaml.v2"
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
		return buildGenericStackCompose(opts)
	}

	s, err = f.GetContent()
	return
}

// if the repo does not have a stack-compose.yml, then we assume it had a Dockerfile and we're building
// a generic one for the built "main" service. In the future, we should check that Dockerfile does exist
// which narrows the chances that the docker was built on Baghdad.
func buildGenericStackCompose(opts stackComposeGetOpts) (string, error) {
	cf := utils.ComposeFile{
		Version: "3.2",
		Services: map[string]utils.ComposeService{"main": utils.ComposeService{
			Image: fmt.Sprintf("%v/%v-%v", os.Getenv("DOCKER_ORG"), opts.RepoName, "main"),
		}},
	}

	out, err := yaml.Marshal(cf)
	if err != nil {
		return "", err
	}

	return string(out), nil
}

type stackComposeGetOpts struct {
	Ctx      context.Context
	Owner    string
	RepoName string
	Tag      string
}
