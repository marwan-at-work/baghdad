package main

import (
	"fmt"
	"os"

	"github.com/marwan-at-work/baghdad"
)

// getSiteURL returns the public URL for a service that you can interact with.
// If the service is not exposed, the url will be unreachable.
func getSiteURL(dj baghdad.DeployJob, serviceName string) string {
	return getSub(dj, serviceName) + "." + getDomain()
}

func getSub(dj baghdad.DeployJob, serviceName string) string {
	return fmt.Sprintf(
		"http://%v-%v-%v-%v",
		dj.BranchName,
		dj.Env,
		dj.Baghdad.Project,
		serviceName,
	)
}

func getDomain() string {
	return os.Getenv("BAGHDAD_DOMAIN_NAME")
}
