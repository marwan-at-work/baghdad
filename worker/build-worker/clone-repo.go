package main

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"sync"

	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"

	"github.com/marwan-at-work/baghdad"
	"github.com/marwan-at-work/baghdad/worker"
)

var repoMapper = map[string]chan cloneData{}
var repoMutex sync.Mutex

type cloneData struct {
	path   string
	resp   chan error
	bj     baghdad.BuildJob
	logger *worker.Logger
}

func cloner(ch chan cloneData) {
	for cd := range ch {
		bj := cd.bj

		var repo *git.Repository
		var err error
		if _, err = os.Stat(cd.path); os.IsNotExist(err) {
			gitURL, _ := url.Parse(bj.GitURL)
			gitURL.User = url.UserPassword(AdminToken, "x-oauth-basic")
			repo, err = git.PlainClone(cd.path, false, &git.CloneOptions{
				URL:               gitURL.String(),
				Progress:          cd.logger,
				RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
				ReferenceName:     plumbing.ReferenceName(fmt.Sprintf("refs/heads/%v", bj.BranchName)),
			})

			if err != nil {
				err = fmt.Errorf("could not clone repo for %v: %v", bj.RepoName, err)
				cd.resp <- err
				continue
			}
			var wt *git.Worktree
			wt, err = repo.Worktree()
			if err != nil {
				err = fmt.Errorf("could not get worktree for %v: %v", bj.RepoName, err)
				cd.resp <- err
				continue
			}

			err = wt.Checkout(plumbing.NewHash(bj.SHA))
			if err != nil {
				err = fmt.Errorf("could not checkout %v: %v", bj.RepoName, err)
			}
		}

		cd.resp <- err
	}
}

func cloneRepo(bj baghdad.BuildJob, logger *worker.Logger) (err error) {
	path := getRepoPath(strconv.Itoa(bj.Type), bj.RepoName, bj.NextTag)

	repoMutex.Lock()
	ch, ok := repoMapper[path]
	if !ok {
		ch = make(chan cloneData)
		repoMapper[path] = ch
		go cloner(ch)
	}
	repoMutex.Unlock()

	rCh := make(chan error)
	cd := cloneData{
		path:   path,
		resp:   rCh,
		bj:     bj,
		logger: logger,
	}

	ch <- cd
	err = <-rCh

	return
}
