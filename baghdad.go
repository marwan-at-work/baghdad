package baghdad

// Baghdad top level config
type Baghdad struct {
	Title        string            `toml:"title"`
	Services     []Service         `toml:"services"`
	Environments map[string]string `toml:"environments"`
	Branches     map[string]Branch `toml:"branches"`
	PostDeploy   PostDeploy        `toml:"post-deploy"`
}

// Service config
type Service struct {
	Name       string `toml:"name"`
	Dockerfile string `toml:"dockerfile"`
	IsExposed  bool   `toml:"isExposed"`
	Port       string `toml:"port"`
	Image      string `toml:"image"`
	IsExternal bool   `toml:"isExternal"`
}

// Branch config
type Branch struct {
	Version string `toml:"version"`
}

// PostDeploy config
type PostDeploy struct {
	SourceService string   `toml:"source-service"`
	TargetService string   `toml:"target-service"`
	Environments  []string `toml:"environments"`
}
