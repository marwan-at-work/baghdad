package handlers

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/google/go-github/github"
	"github.com/marwan-at-work/baghdad"
	"github.com/marwan-at-work/baghdad/bus"
	"github.com/marwan-at-work/baghdad/utils"
	"github.com/marwan-at-work/baghdad/worker"
	"github.com/satori/go.uuid"
)

// GithubHook handler for all incoming github webhooks
func GithubHook(b bus.Publisher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		payload, err := github.ValidatePayload(r, []byte("baghdad"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, err.Error())
			return
		}

		event, err := github.ParseWebHook(github.WebHookType(r), payload)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, err.Error())
			return
		}

		switch event.(type) {
		case *github.PushEvent:
			handlePush(w, r, event.(*github.PushEvent), b)
			return
		case *github.PullRequestEvent:
			handlePull(w, r, event.(*github.PullRequestEvent), b)
			return
		default:
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}
}

func handlePush(w http.ResponseWriter, r *http.Request, pe *github.PushEvent, b bus.Publisher) {
	if !pushValid(pe) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "invalid push")
		return
	}

	bj := baghdad.BuildJob{
		GitURL:     pe.Repo.GetCloneURL(),
		RepoName:   pe.Repo.GetName(),
		RepoOwner:  pe.Repo.Owner.GetName(),
		SHA:        pe.HeadCommit.GetID(),
		Type:       baghdad.PushEvent,
		BranchName: getBranchFromRef(pe.GetRef()),
	}

	bj.LogID = fmt.Sprintf("%v-%v", bj.RepoName, uuid.NewV4().String())

	bgd, err := utils.GetBaghdad(utils.GetGithub(os.Getenv("ADMIN_TOKEN")), utils.GetBaghdadOpts{
		SHA:      bj.SHA,
		Owner:    bj.RepoOwner,
		RepoName: bj.RepoName,
		Ctx:      context.Background(),
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "could not get baghdad.toml: %v", err)
		return
	}

	if _, ok := bgd.Branches[bj.BranchName]; !ok {
		fmt.Fprintf(w, "untracked branch: %v", bj.BranchName)
		return
	}
	bj.Baghdad = bgd

	go sendBuildJob(b, bj)
	w.WriteHeader(http.StatusOK)
}

func handlePull(w http.ResponseWriter, r *http.Request, pre *github.PullRequestEvent, b bus.Publisher) {
	if !pullValid(pre) {
		w.WriteHeader(http.StatusOK)
		return
	}

	owner := pre.PullRequest.Head.Repo.Owner.GetName()
	if owner == "" {
		owner = pre.PullRequest.Head.Repo.Owner.GetLogin()
	}

	bj := baghdad.BuildJob{
		BranchName: pre.PullRequest.Head.GetRef(),
		GitURL:     pre.PullRequest.Head.Repo.GetCloneURL(),
		RepoName:   pre.PullRequest.Head.Repo.GetName(),
		RepoOwner:  owner,
		SHA:        pre.PullRequest.Head.GetSHA(),
		Type:       baghdad.PullRequestEvent,
		PRNum:      pre.PullRequest.GetNumber(),
	}

	bj.LogID = fmt.Sprintf("%v-%v", bj.RepoName, uuid.NewV4().String())

	bgd, err := utils.GetBaghdad(utils.GetGithub(os.Getenv("ADMIN_TOKEN")), utils.GetBaghdadOpts{
		SHA:      bj.SHA,
		Owner:    bj.RepoOwner,
		RepoName: bj.RepoName,
		Ctx:      context.Background(),
	})
	if err != nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	bj.Baghdad = bgd

	go sendBuildJob(b, bj)
	w.WriteHeader(http.StatusOK)
}

func pushValid(pe *github.PushEvent) bool {
	return !pe.GetDeleted() && pe.HeadCommit != nil // add the infinite loop check.
}

func pullValid(pre *github.PullRequestEvent) bool {
	a := pre.GetAction()
	return a == "opened" || a == "reopened" || a == "synchronize"
}

func sendBuildJob(b bus.Publisher, bj baghdad.BuildJob) {
	p, _ := baghdad.EncodeBuildJob(bj)

	var jobType string
	if bj.Type == baghdad.PullRequestEvent {
		jobType = "pull request"
	} else {
		jobType = "push"
	}

	logger := worker.NewLogger(bj.RepoName, bj.LogID, b.(bus.BroadcastPublisher))
	logger.Loglnf("sending build job - %v: %v", jobType, bj.SHA)
	b.Publish("build-sync", p)
}

func getBranchFromRef(s string) string {
	arr := strings.Split(s, "/")
	return arr[len(arr)-1]
}
