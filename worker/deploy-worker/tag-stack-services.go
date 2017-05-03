package main

import (
	"fmt"
	"os"

	"github.com/marwan-at-work/baghdad"
	yaml "gopkg.in/yaml.v2"
)

func tagStackServices(stackFile []byte, dj baghdad.DeployJob) ([]byte, error) {
	c := ComposeFile{}
	err := yaml.Unmarshal(stackFile, &c)
	if err != nil {
		return []byte{}, err
	}

	for _, bs := range dj.Baghdad.Services {
		s := c.Services[bs.Name]
		if !bs.IsExternal {
			s.Image = s.Image + ":" + dj.Tag
		}

		labels := []string{"traefik.docker.network=traefik-net"} // all services may not need to be on traefik-net
		if bs.IsExposed {
			labels = append(
				labels,
				"traefik.frontend.rule=Host:"+getSiteURL(dj, bs.Name),
				"traefik.port="+bs.Port,
			)
		}

		s.Deploy.Labels = append(s.Deploy.Labels, labels...)

		if ns, ok := s.Networks.(map[interface{}]interface{}); ok {
			ns["traefik-net"] = map[interface{}]interface{}{}
			s.Networks = ns
		} else if a, ok := s.Networks.([]interface{}); ok {
			a = append(a, "traefik-net")
			s.Networks = a
		} else {
			s.Networks = []string{"traefik-net"}
		}

		c.Services[bs.Name] = s
	}

	if c.Networks == nil {
		c.Networks = map[string]ComposeNetwork{
			"traefik-net": ComposeNetwork{External: true},
		}
	} else {
		c.Networks["traefik-net"] = ComposeNetwork{External: true}
	}

	return yaml.Marshal(c)
}

func getSiteURL(dj baghdad.DeployJob, serviceName string) string {
	return getSub(dj, serviceName) + "." + getDomain()
}

func getSub(dj baghdad.DeployJob, serviceName string) string {
	return fmt.Sprintf(
		"%v-%v-%v-%v",
		dj.BranchName,
		dj.Env,
		dj.Baghdad.Project,
		serviceName,
	)
}

func getDomain() string {
	return os.Getenv("BAGHDAD_DOMAIN_NAME")
}
