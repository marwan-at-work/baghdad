package utils

import (
	"fmt"

	"github.com/marwan-at-work/baghdad"
	yaml "gopkg.in/yaml.v2"
)

// TagStackServices creates docker-compose file that is valid for `docker stack deploy` within
// the baghdad ecosystem.
// this can be used to bootstrap the Baghdad microservices.
func TagStackServices(
	stackFile []byte,
	b baghdad.Baghdad,
	tag,
	branch,
	env,
	domain string,
) ([]byte, error) {
	c := ComposeFile{}
	err := yaml.Unmarshal(stackFile, &c)
	if err != nil {
		return []byte{}, err
	}

	for _, bs := range b.Services {
		s, ok := c.Services[bs.Name]
		if !ok {
			continue
		}
		if !bs.IsExternal {
			s.Image = s.Image + ":" + tag
		}

		if envs, ok := s.Environment.([]interface{}); ok {
			envs = append(envs, fmt.Sprintf("BAGHDAD_BUILD_VERSION=%v", tag))
			s.Environment = envs
		} else if envDic, ok := s.Environment.(map[string]interface{}); ok {
			envDic["BAGHDAD_BUILD_VERSION"] = tag
			s.Environment = envDic
		}

		labels := []string{"traefik.docker.network=traefik-net"} // all services may not need to be on traefik-net
		if bs.IsExposed {
			subDomain := fmt.Sprintf("%v-%v-%v-%v", branch, env, b.Project, bs.Name)
			labels = append(
				labels,
				fmt.Sprintf("traefik.frontend.rule=Host:%v.%v", subDomain, domain),
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
