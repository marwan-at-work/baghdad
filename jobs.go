package baghdad

// DeployJob use to send a deploy job to the deploy worker
type DeployJob struct {
	Baghdad    Baghdad
	BranchName string
	Env        string
	RepoName   string
	Tag        string
	RepoOwner  string
	LogID      string
}

// BuildJob use to send a build job to the build worker
type BuildJob struct {
	Baghdad    Baghdad
	BranchName string
	GitURL     string
	PRNum      int
	RepoName   string
	RepoOwner  string
	SHA        string
	Type       int // use PushEvent/PullEvent enums
	Service    Service
	NextTag    string
	ErrString  string
	LogID      string
}

// SecretsJob use to save a secret to swarm.
type SecretsJob struct {
	ProjectName string
	SecretName  string
	SecretBody  []byte
}

// PostDeployJob use to send a post deploy job to rabbitmq
type PostDeployJob struct {
	Baghdad     Baghdad
	ProjectName string
	Tag         string
	Env         string
	BranchName  string
	SiteURL     string
}
