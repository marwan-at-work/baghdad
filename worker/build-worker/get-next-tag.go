package main

import (
	"fmt"
	"regexp"
	"strconv"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"

	"github.com/marwan-at-work/baghdad"
)

func getNextTag(b baghdad.BuildJob, repo *git.Repository) (string, error) {
	tags, err := repo.Tags()
	if err != nil {
		return "", err
	}

	sprint := b.Baghdad.Branches[b.BranchName].Version
	expr := fmt.Sprintf(`%v-%v-build\.(\d+)`, b.BranchName, sprint)
	r, err := regexp.Compile(expr)
	if err != nil {
		return "", err
	}
	latestBuildNumber := 0

	tags.ForEach(func(ref *plumbing.Reference) error {
		tag := ref.Name().String()
		validTag := r.MatchString(tag)
		if !validTag {
			return nil
		}

		subs := r.FindStringSubmatch(tag)
		buildNumber, _ := strconv.Atoi(subs[len(subs)-1])
		if buildNumber > latestBuildNumber {
			latestBuildNumber = buildNumber
		}

		return nil
	})

	nextTag := fmt.Sprintf("%v-%v-build.%v", b.BranchName, sprint, latestBuildNumber+1)
	return nextTag, nil
}
