package baghdad

const (
	// PushEvent enum
	PushEvent = 1 + iota
	// PullRequestEvent enum
	PullRequestEvent
	// BuildSuccessEvent enum
	BuildSuccessEvent
	// BuildFailureEvent enum
	BuildFailureEvent
)
