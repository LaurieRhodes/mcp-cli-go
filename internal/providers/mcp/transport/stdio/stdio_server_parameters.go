package stdio

// StdioServerParameters contains the parameters needed to start and connect to a server process.
type StdioServerParameters struct {
	// Command is the executable path to run
	Command string `json:"command"`
	
	// Args are the command-line arguments to pass to the command
	Args []string `json:"args"`
	
	// Env is the environment variables to set for the process
	// If nil, the current process's environment will be used
	Env map[string]string `json:"env,omitempty"`
}
