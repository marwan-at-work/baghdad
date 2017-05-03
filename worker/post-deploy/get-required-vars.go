package main

func getRequiredVars() []string {
	return []string{
		"AMQP_URL",
		"ADMIN_TOKEN",
		"DOCKER_AUTH_USER",
		"DOCKER_AUTH_PASS",
	}
}
