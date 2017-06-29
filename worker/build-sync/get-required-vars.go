package main

func getRequiredVars() []string {
	return []string{
		"BAGHDAD_BUILD_ENV",
		"BAGHDAD_BUILD_BRANCH",
		"BAGHDAD_DOMAIN_NAME",
		"BUILDER_ADDR",
		"ADMIN_TOKEN",
		"AMQP_URL",
	}
}
