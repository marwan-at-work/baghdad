package utils

import (
	"os"

	"github.com/go-redis/redis"
	"github.com/marwan-at-work/baghdad/worker"
)

// ReleaseDeploy sends an empty message to the projectQ, and makes redis project idle.
func ReleaseDeploy(projectQ string, w *worker.Worker, logger *worker.Logger) {
	err := w.Publish(projectQ, nil)
	if err != nil {
		logger.Loglnf("%v: could not be released from q: %v", projectQ, err)
	}

	rc := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_URL"),
		Password: "",
		DB:       0,
	})

	err = rc.Set(projectQ, "idle", 0).Err()
	if err != nil {
		logger.Loglnf("%v: could not release redis val: %v", projectQ, err)
	}
}
