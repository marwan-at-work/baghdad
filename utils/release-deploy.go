package utils

// import (
// 	"fmt"
// 	"os"
//
// 	"github.com/go-redis/redis"
// 	"github.com/streadway/amqp"
// 	"github.com/marwan-at-work/baghdad/worker"
// )
//
// // ReleaseDeploy sends an empty message to the projectQ, and makes redis project idle.
// func ReleaseDeploy(projectQ string, w *worker.Worker) {
// 	err := w.Publish(worker.PublishOpts{
// 		Exchange:  "",
// 		Key:       projectQ,
// 		Mandatory: false,
// 		Immediate: false,
// 		Msg: amqp.Publishing{
// 			DeliveryMode: amqp.Persistent,
// 		},
// 	})
// 	if err != nil {
// 		w.Log(fmt.Sprintf("%v: could not be released from q: %v", projectQ, err))
// 	}
//
// 	rc := redis.NewClient(&redis.Options{
// 		Addr:     os.Getenv("REDIS_URL"),
// 		Password: "",
// 		DB:       0,
// 	})
//
// 	err = rc.Set(projectQ, "idle", 0).Err()
// 	if err != nil {
// 		w.Log(fmt.Sprintf("%v: could not release redis val: %v", projectQ, err))
// 	}
// }
