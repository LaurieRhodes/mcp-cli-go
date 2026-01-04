package skills

// ExecutionMode determines how scripts are executed
type ExecutionMode string

const (
	// ExecutionModePassive loads skills as documentation only (no script execution)
	ExecutionModePassive ExecutionMode = "passive"
	
	// ExecutionModeActive executes scripts (requires Docker/Podman)
	ExecutionModeActive ExecutionMode = "active"
	
	// ExecutionModeAuto uses active if Docker/Podman available, otherwise passive
	ExecutionModeAuto ExecutionMode = "auto"
)

// ScriptExecution represents a script execution request
type ScriptExecution struct {
	ScriptPath string   // Relative path to script within skill directory
	Args       []string // Arguments to pass to script
	Timeout    int      // Timeout in seconds (0 = use default)
}

// ExecutionResult represents the result of script execution
type ExecutionResult struct {
	Output   string // Combined stdout/stderr
	ExitCode int    // Exit code (0 = success)
	Error    error  // Error if execution failed
	Duration int64  // Execution time in milliseconds
}

// CodeExecutionRequest represents a request to execute arbitrary code with skill context
type CodeExecutionRequest struct {
	SkillName string            // Which skill's libraries to use
	Language  string            // "python" or "node"
	Code      string            // Code to execute
	Files     map[string][]byte // Optional files to make available in workspace
	Timeout   int               // Timeout in seconds (0 = use default)
}
