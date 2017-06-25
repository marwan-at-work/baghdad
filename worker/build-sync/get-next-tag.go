package main

import (
	"fmt"
	"regexp"
	"strconv"
)

func getNextTag(branchName, sprint string, tags []string) (string, error) {
	expr := fmt.Sprintf(`%v-%v-build\.(\d+)`, branchName, sprint)
	r, err := regexp.Compile(expr)
	if err != nil {
		return "", err
	}
	latestBuildNumber := 0

	for _, tag := range tags {
		validTag := r.MatchString(tag)
		if !validTag {
			continue
		}

		subs := r.FindStringSubmatch(tag)
		buildNumber, _ := strconv.Atoi(subs[len(subs)-1])
		if buildNumber > latestBuildNumber {
			latestBuildNumber = buildNumber
		}
	}

	nextTag := fmt.Sprintf("%v-%v-build.%v", branchName, sprint, latestBuildNumber+1)
	return nextTag, nil
}
