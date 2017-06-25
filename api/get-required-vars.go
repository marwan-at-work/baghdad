package main

func getRequiredVars() []string {
	return []string{
		"DOCKER_REMOTE_API_URL",
		"AMQP_URL",
		"ADMIN_TOKEN",
	}
}
