package host

import (
	"os"
	"strings"
)

// GetDefaultEnvironment returns the default environment variables.
// This is used when spawning server processes.
func GetDefaultEnvironment() map[string]string {
	env := make(map[string]string)
	
	// Populate from current process environment
	for _, item := range os.Environ() {
		parts := strings.SplitN(item, "=", 2)
		if len(parts) == 2 {
			key := parts[0]
			value := parts[1]
			env[key] = value
		}
	}
	
	return env
}

// MergeEnvironment merges the provided environment variables with the default ones.
// Values in the provided environment take precedence over the defaults.
func MergeEnvironment(env map[string]string) map[string]string {
	merged := GetDefaultEnvironment()
	
	// Add or overwrite values from the provided environment
	for key, value := range env {
		merged[key] = value
	}
	
	return merged
}

// EnvToSlice converts an environment map to a slice of "KEY=VALUE" strings.
// This is useful for passing to exec.Command.
func EnvToSlice(env map[string]string) []string {
	result := make([]string, 0, len(env))
	
	for key, value := range env {
		result = append(result, key+"="+value)
	}
	
	return result
}
