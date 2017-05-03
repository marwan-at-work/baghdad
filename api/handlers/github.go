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

	fmt.Println("sending push job", pe.Repo.GetName())
	go sendBuildJob(b, bj)
	w.WriteHeader(http.StatusOK)
}

func handlePull(w http.ResponseWriter, r *http.Request, pre *github.PullRequestEvent, b bus.Publisher) {
	if !pullValid(pre) {
		w.WriteHeader(http.StatusOK)
		return
	}

	bj := baghdad.BuildJob{
		GitURL:    pre.PullRequest.Head.Repo.GetCloneURL(),
		RepoName:  pre.PullRequest.Head.Repo.GetName(),
		RepoOwner: pre.PullRequest.Head.Repo.Owner.GetName(),
		SHA:       pre.PullRequest.Head.GetSHA(),
		Type:      baghdad.PullRequestEvent,
		PRNum:     pre.PullRequest.GetNumber(),
	}

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

	fmt.Println("sending pr job", pre.Repo.GetName())
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

	b.Publish("build-sync", p)
}

func getBranchFromRef(s string) string {
	arr := strings.Split(s, "/")
	return arr[len(arr)-1]
}
