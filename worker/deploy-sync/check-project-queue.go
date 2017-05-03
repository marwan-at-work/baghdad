package main

import (
	"fmt"
	"os"

	"github.com/go-redis/redis"
)

// checkProjectQueue checks the database if a queue has ever been created.
// if not, this is the first time a project-env is being deployed and it should
// not wait.
func checkProjectQueue(projectQName string) (shouldWait bool, err error) {
	shouldWait = true

	rc := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_URL"),
		Password: "",
		DB:       0,
	})

	val, err := rc.Get(projectQName).Result()
	if err == redis.Nil || (err == nil && val == "idle") {
		shouldWait = false
		err = rc.Set(projectQName, "processing", 0).Err()
		if err != nil {
			err = fmt.Errorf("could not set %v in redis: %v", projectQName, err)
			return
		}
	} else if err != nil {
		err = fmt.Errorf("could not get %v from redis: %v", projectQName, err)
		return
	}

	return
}
