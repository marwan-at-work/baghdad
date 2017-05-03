package utils

import (
	"fmt"
	"log"
)

// FailOnError helper function to panic on error.
func FailOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}
