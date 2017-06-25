package utils

import "os"

// GetAMQPURL performs an environment variable look up on $AMQP_URL - returns
// default url if empty
func GetAMQPURL() string {
	url := os.Getenv("AMQP_URL")
	if url == "" {
		url = "amqp://user:password@localhost:5672"
	}

	return url
}
