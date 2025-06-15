package logging

import "os"

var (
	// productionMode controls whether we're in production mode
	productionMode = false
)

// SetProductionMode enables or disables production mode
// In production mode, only WARN and above messages are logged
func SetProductionMode(enabled bool) {
	productionMode = enabled
	
	// Update the default level accordingly
	if enabled {
		SetDefaultLevel(WARN)
	}
}

// IsProduction returns whether production mode is enabled
func IsProduction() bool {
	return productionMode
}

// init initializes production mode based on environment
func init() {
	// Check environment variable for production mode
	if os.Getenv("PRODUCTION") == "1" {
		SetProductionMode(true)
	}
}
