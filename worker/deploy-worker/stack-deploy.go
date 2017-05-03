package main

import (
	"io"
	"os"
	"os/exec"
)

func stackDeploy(stackName, stackfile, tag string, opts stackDeployOpts) error {
	args := []string{}
	if remoteURL := os.Getenv("DOCKER_REMOTE_API_URL"); remoteURL != "" {
		args = append(args, "-H", remoteURL)
	}
	args = append(args, "stack", "deploy", "--compose-file", stackfile, stackName)

	cmd := exec.Command("docker", args...)
	cmd.Stdout = opts.stdout
	cmd.Stderr = opts.stdout
	return cmd.Run()
}

type stackDeployOpts struct {
	stdout io.Writer
}
