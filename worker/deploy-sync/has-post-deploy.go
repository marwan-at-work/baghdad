package main

import "github.com/marwan-at-work/baghdad"

func hasPostDeploy(b baghdad.Baghdad, env string) bool {
	hasPostDeployService := b.PostDeploy.SourceService != "" && b.PostDeploy.TargetService != ""
	hasEnv := false
	for _, e := range b.PostDeploy.Environments {
		if e == env {
			hasEnv = true
			break
		}
	}

	return hasPostDeployService && hasEnv
}
