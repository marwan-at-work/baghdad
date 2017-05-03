package main

import "path/filepath"

func getRepoPath(repoName string, buildType string) string {
	return filepath.Join(WorkPath, buildType, repoName)
}
