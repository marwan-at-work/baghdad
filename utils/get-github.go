package utils

import (
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

// GetGithub given a token, it will return a client to perform Github API operations on.
func GetGithub(t string) *github.Client {
	var ts = oauth2.StaticTokenSource(&oauth2.Token{AccessToken: t})
	var tc = oauth2.NewClient(oauth2.NoContext, ts)
	return github.NewClient(tc)
}
