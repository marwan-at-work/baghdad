package utils

import (
	"fmt"
	"os"
)

// ValidateEnvVars panics if any given environment variable is an empty string.
func ValidateEnvVars(vars ...string) {
	for _, v := range vars {
		if os.Getenv(v) == "" {
			panic(fmt.Errorf("env var cannot be empty: %v", v))
		}
	}
}
