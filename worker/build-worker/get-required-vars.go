package main

func getRequiredVars() []string {
	return []string{
		"ADMIN_TOKEN",
		"AMQP_URL",
		"DOCKER_AUTH_USER",
		"DOCKER_AUTH_PASS",
		"DOCKER_ORG",
	}
}
