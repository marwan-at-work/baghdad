package main

import (
	"fmt"
	"net/url"
	"os"
	"strconv"

	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"

	"github.com/marwan-at-work/baghdad"
)

func cloneRepo(b baghdad.BuildJob) (repo *git.Repository, err error) {
	path := getRepoPath(b.RepoName, strconv.Itoa(b.Type))
	err = os.RemoveAll(path)
	if err != nil {
		return
	}

	gitURL, _ := url.Parse(b.GitURL)
	gitURL.User = url.UserPassword(AdminToken, "x-oauth-basic")
	repo, err = git.PlainClone(path, false, &git.CloneOptions{
		URL:               gitURL.String(),
		Progress:          os.Stdout, // should be worker
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
	})

	if err != nil {
		err = fmt.Errorf("could not clone repo for %v: %v", b.RepoName, err)
		return
	}

	wt, err := repo.Worktree()
	if err != nil {
		err = fmt.Errorf("could not get worktree for %v: %v", b.RepoName, err)
		return
	}

	err = wt.Checkout(plumbing.NewHash(b.SHA))
	if err != nil {
		err = fmt.Errorf("could not checkout %v: %v", b.RepoName, err)
	}

	return
}
