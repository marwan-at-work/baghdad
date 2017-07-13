package worker

import (
	"fmt"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/go-redis/redis"
	"github.com/marwan-at-work/baghdad/bus"
)

func init() {
	logrus.SetFormatter(&LogFormatter{})
	level := os.Getenv("LOG_LEVEL")

	switch level {
	case "DEBUG":
		logrus.SetLevel(logrus.DebugLevel)
	default:
		logrus.SetLevel(logrus.WarnLevel)
	}
}

// LogFormatter overrides the default text formatter for logrus.
// this is so the output is not too cluttered. The only addition to the log
// that is needed is the project name.
type LogFormatter struct{}

// Format formats an incomming logrus log.
func (w *LogFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	msg := entry.Message
	project, ok := entry.Data["project"]
	if !ok {
		return nil, fmt.Errorf("log did not have a project field: %v", msg)
	}

	return []byte(fmt.Sprintf("%v: %v", project, msg)), nil
}

// Logger knows how to send logs to a project-specific channel
type Logger struct {
	Project     string
	Broadcaster bus.Broadcaster
	logCtx      *logrus.Entry
	ID          string
	redis       *redis.Client
}

// NewLogger Logger constructor
func NewLogger(project, id string, broadcaster bus.Broadcaster) *Logger {
	return &Logger{
		Project:     project,
		ID:          id,
		Broadcaster: broadcaster,
		logCtx:      logrus.WithField("project", project),
		redis: redis.NewClient(&redis.Options{
			Addr:     os.Getenv("REDIS_URL"),
			Password: "",
			DB:       0,
		}),
	}
}

// Write implements the io.Writer interface, this way the execute commands
// can directly link their Stdout & Stderr to project logs.
func (l *Logger) Write(p []byte) (n int, err error) {
	n = len(p)

	l.logCtx.Debug(string(p))

	err = l.Broadcaster.Broadcast("logs", l.Project, p)
	if err != nil {
		l.logCtx.Warnf("could not broadcast logs: %v\n", err)
		err = nil
	}

	if l.ID != "" {
		err = l.redis.LPush(l.ID, string(p)).Err()
		if err != nil {
			l.logCtx.Warnf("could not persist %v log to redis: %v\n", l.ID, err)
			err = nil
		}
	}

	return
}

// Logf like fmt.Printf but instead of printing to stdout, it calls the Logger Write.
func (l *Logger) Logf(format string, a ...interface{}) (n int, err error) {
	s := fmt.Sprintf(format, a...)

	return l.Write([]byte(s))
}

// Loglnf like Logf with \n at the end.
func (l *Logger) Loglnf(format string, a ...interface{}) (n int, err error) {
	s := fmt.Sprintln(fmt.Sprintf(format, a...))

	return l.Write([]byte(s))
}

// Log like fmt.Print but calls l.Write
func (l *Logger) Log(a ...interface{}) (n int, err error) {
	s := fmt.Sprintln(a...)

	return l.Write([]byte(s))
}
