package utils

// ComposeFile top level compose file.
type ComposeFile struct {
	Version  string                    `yaml:"version,omitempty"`
	Services map[string]ComposeService `yaml:"services,omitempty"`
	Networks map[string]ComposeNetwork `yaml:"networks,omitempty"`
	Volumes  interface{}               `yaml:"volumes,omitempty"`
	Secrets  interface{}               `yaml:"secrets,omitempty"`
}

// ComposeService top level compose service.
type ComposeService struct {
	CapAdd          interface{}          `yaml:"cap_add,omitempty"`
	CapDrop         interface{}          `yaml:"cap_drop,omitempty"`
	CgroupParent    interface{}          `yaml:"cgroup_parent,omitempty"`
	Command         interface{}          `yaml:"command,omitempty"`
	ContainerName   interface{}          `yaml:"container_name,omitempty"`
	DependsOn       interface{}          `yaml:"depends_on,omitempty"`
	Deploy          ComposeServiceDeploy `yaml:"deploy,omitempty"`
	Devices         interface{}          `yaml:"devices,omitempty"`
	DNS             interface{}          `yaml:"dns,omitempty"`
	DNSSearch       interface{}          `yaml:"dns_search,omitempty"`
	DomainName      interface{}          `yaml:"domainname,omitempty"`
	Entrypoint      interface{}          `yaml:"entrypoint,omitempty"`
	Environment     interface{}          `yaml:"environment,omitempty"`
	Expose          interface{}          `yaml:"expose,omitempty"`
	ExternalLinks   interface{}          `yaml:"external_links,omitempty"`
	ExtraHosts      interface{}          `yaml:"extra_hosts,omitempty"`
	Hostname        interface{}          `yaml:"hostname,omitempty"`
	HealthCheck     interface{}          `yaml:"healthcheck,omitempty"`
	Image           string               `yaml:"image,omitempty"`
	Ipc             interface{}          `yaml:"ipc,omitempty"`
	Labels          interface{}          `yaml:"labels,omitempty"`
	Links           interface{}          `yaml:"links,omitempty"`
	Logging         interface{}          `yaml:"logging,omitempty"`
	MacAddress      interface{}          `yaml:"mac_address,omitempty"`
	NetworkMode     interface{}          `yaml:"network_mode,omitempty"`
	Networks        interface{}          `yaml:"networks,omitempty"`
	Pid             interface{}          `yaml:"pid,omitempty"`
	Ports           interface{}          `yaml:"ports,omitempty"`
	Privileged      interface{}          `yaml:"privileged,omitempty"`
	ReadOnly        interface{}          `yaml:"read_only,omitempty"`
	Restart         interface{}          `yaml:"Restart,omitempty"`
	Secrets         interface{}          `yaml:"secrets,omitempty"`
	SecurityOpt     interface{}          `yaml:"security_opt,omitempty"`
	StdinOpen       interface{}          `yaml:"stdin_open,omitempty"`
	StopGracePeriod interface{}          `yaml:"stop_grace_period,omitempty"`
	StopSignal      interface{}          `yaml:"stop_signal,omitempty"`
	Tmpfs           interface{}          `yaml:"tmpfs,omitempty"`
	Tty             interface{}          `yaml:"tty,omitempty"`
	Ulimits         interface{}          `yaml:"ulimits,omitempty"`
	User            interface{}          `yaml:"user,omitempty"`
	Volumes         interface{}          `yaml:"volumes,omitempty"`
	WorkingDir      interface{}          `yaml:"working_dir,omitempty"`
}

// ComposeNetwork top level compose network
type ComposeNetwork struct {
	Driver     interface{} `yaml:"driver,omitempty"`
	DriverOpts interface{} `yaml:"driver_opts,omitempty"`
	Ipam       interface{} `yaml:"ipam,omitempty"`
	External   interface{} `yaml:"external,omitempty"`
	Labels     interface{} `yaml:"labels,omitempty"`
}

// ComposeServiceDeploy Deploy config for a compose service
type ComposeServiceDeploy struct {
	Mode          interface{} `yaml:"mode,omitempty"`
	Replicas      interface{} `yaml:"replicas,omitempty"`
	Labels        []string    `yaml:"labels,omitempty"`
	UpdateConfig  interface{} `yaml:"update_config,omitempty"`
	Resources     interface{} `yaml:"resources,omitempty"`
	RestartPolicy interface{} `yaml:"restart_policy,omitempty"`
	Placement     interface{} `yaml:"placement,omitempty"`
}
