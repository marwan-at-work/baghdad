package main

import "path/filepath"

func getRepoPath(buildType, repoName, tag string) string {
	if tag == "" {
		return filepath.Join(WorkPath, buildType, repoName)
	}

	return filepath.Join(WorkPath, buildType, repoName, tag)
}
