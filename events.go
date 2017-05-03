package baghdad

// BuildEvent is used at the end of a build to let subscribes know if a build failed/succeeded.
type BuildEvent struct {
	BuildJob
	EventType int
	Tag       string
}
