package sandbox

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/logging"
)

// NativeExecutor uses Docker/Podman CLI from host (for native deployments)
type NativeExecutor struct {
	config     ExecutorConfig
	command    string // "docker" or "podman"
}

// NewNativeExecutor creates a new native Docker/Podman executor
func NewNativeExecutor(config ExecutorConfig) (*NativeExecutor, error) {
	executor := &NativeExecutor{config: config}
	
	// Try docker first, then podman
	if cmd := exec.Command("docker", "version"); cmd.Run() == nil {
		executor.command = "docker"
	} else if cmd := exec.Command("podman", "version"); cmd.Run() == nil {
		executor.command = "podman"
	} else {
		return nil, fmt.Errorf("neither docker nor podman found")
	}
	
	return executor, nil
}

// IsAvailable checks if Docker/Podman CLI is available
func (n *NativeExecutor) IsAvailable() bool {
	return n.command != ""
}

// ExecutePython runs a Python script using Docker/Podman CLI
func (n *NativeExecutor) ExecutePython(ctx context.Context, skillDir, scriptPath string, args []string) (string, error) {
	// Build docker/podman run command with security constraints
	cmdArgs := []string{
		"run",
		"--rm",                                    // Remove container after execution
		"--read-only",                             // Read-only root filesystem
		"--network=" + n.config.NetworkMode,       // Network mode from config
		"--memory=" + n.config.MemoryLimit,       // Memory limit
		"--cpus=" + n.config.CPULimit,            // CPU limit
		"--pids-limit=100",                        // Process limit
		"--security-opt=no-new-privileges",        // No privilege escalation
		"--cap-drop=ALL",                          // Drop all capabilities
		"-v", fmt.Sprintf("%s:/skill:ro", skillDir), // Mount skill dir read-only
		"-v", fmt.Sprintf("%s:/outputs:rw", n.config.OutputsDir),  // Persistent outputs directory
		"-w", "/skill",                            // Working directory
		n.config.PythonImage,                      // Python image
		"python", scriptPath,                      // Command
	}
	cmdArgs = append(cmdArgs, args...)

	cmd := exec.CommandContext(ctx, n.command, cmdArgs...)
	output, err := cmd.CombinedOutput()

	// Check for timeout
	if ctx.Err() == context.DeadlineExceeded {
		return "", fmt.Errorf("execution timeout after %v", n.config.Timeout)
	}

	if err != nil {
		return string(output), fmt.Errorf("script execution failed: %w\nOutput: %s", err, output)
	}

	return string(output), nil
}

// ExecuteBash runs a Bash script using Docker/Podman CLI
func (n *NativeExecutor) ExecuteBash(ctx context.Context, skillDir, scriptPath string, args []string) (string, error) {
	cmdArgs := []string{
		"run",
		"--rm",
		"--read-only",
		"--network=" + n.config.NetworkMode,
		"--memory=" + n.config.MemoryLimit,
		"--cpus=" + n.config.CPULimit,
		"--pids-limit=100",
		"--security-opt=no-new-privileges",
		"--cap-drop=ALL",
		"-v", fmt.Sprintf("%s:/skill:ro", skillDir),
		"-w", "/skill",
		"alpine:latest", // Lightweight image for bash
		"sh", scriptPath,
	}
	cmdArgs = append(cmdArgs, args...)

	cmd := exec.CommandContext(ctx, n.command, cmdArgs...)
	output, err := cmd.CombinedOutput()

	if ctx.Err() == context.DeadlineExceeded {
		return "", fmt.Errorf("execution timeout after %v", n.config.Timeout)
	}

	if err != nil {
		return string(output), fmt.Errorf("script execution failed: %w\nOutput: %s", err, output)
	}

	return string(output), nil
}

// GetInfo returns information about the Docker/Podman installation
func (n *NativeExecutor) GetInfo() string {
	cmd := exec.Command(n.command, "version", "--format", "{{.Server.Version}}")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Sprintf("%s (version unknown)", strings.Title(n.command))
	}
	version := strings.TrimSpace(string(output))
	return fmt.Sprintf("%s %s (native)", strings.Title(n.command), version)
}

// ExecutePythonCode runs Python code with dual mount support
// workspaceDir: read-write workspace for files and code execution
// skillLibsDir: read-only skill directory for importing helper libraries
func (n *NativeExecutor) ExecutePythonCode(ctx context.Context, workspaceDir, skillLibsDir, scriptPath string, args []string) (string, error) {
	// Get the appropriate image and network mode for this skill
	image := n.config.GetImageForSkill(skillLibsDir)
	networkMode := n.config.GetNetworkModeForSkill(skillLibsDir)
	logging.Info("🐳 Executing skill from '%s' with image '%s' (network: %s)", skillLibsDir, image, networkMode)
	
	// Build docker/podman run command with dual mounts
	cmdArgs := []string{
		"run",
		"--rm",                                      // Remove container after execution
		"--read-only",                               // Read-only root filesystem
		"--network=" + networkMode,                  // Network mode for this skill
		"--memory=" + n.config.MemoryLimit,         // Memory limit
		"--cpus=" + n.config.CPULimit,              // CPU limit
		"--pids-limit=100",                          // Process limit
		"--security-opt=no-new-privileges",          // No privilege escalation
		"--cap-drop=ALL",                            // Drop all capabilities
		"-v", fmt.Sprintf("%s:/workspace:rw", workspaceDir),     // Read-write workspace
		"-v", fmt.Sprintf("%s:/skill:ro", skillLibsDir),         // Read-only skill libs
		"-v", fmt.Sprintf("%s:/outputs:rw", n.config.OutputsDir),  // Persistent outputs directory
		"-w", "/workspace",                          // Working directory
		"-e", "PYTHONPATH=/skill",                   // Can import from /skill
		"--tmpfs", "/tmp:rw,exec,size=100m",        // Writable /tmp for Python
		image,                                       // Use skill-specific image
		"python", scriptPath,                        // Command (relative to /workspace)
	}
	cmdArgs = append(cmdArgs, args...)

	cmd := exec.CommandContext(ctx, n.command, cmdArgs...)
	output, err := cmd.CombinedOutput()

	// Check for timeout
	if ctx.Err() == context.DeadlineExceeded {
		return "", fmt.Errorf("execution timeout after %v", n.config.Timeout)
	}

	if err != nil {
		return string(output), fmt.Errorf("code execution failed: %w\nOutput: %s", err, output)
	}

	return string(output), nil
}

// ExecuteBashCode runs Bash code with dual mount support
// workspaceDir: read-write workspace for files and code execution
// skillLibsDir: read-only skill directory (for future bash libraries)
func (n *NativeExecutor) ExecuteBashCode(ctx context.Context, workspaceDir, skillLibsDir, scriptPath string, args []string) (string, error) {
	// Get the appropriate image and network mode for this skill
	image := n.config.GetImageForSkill(skillLibsDir)
	networkMode := n.config.GetNetworkModeForSkill(skillLibsDir)
	logging.Info("🐳 Executing bash skill from '%s' with image '%s' (network: %s)", skillLibsDir, image, networkMode)
	
	// Build docker/podman run command with dual mounts
	cmdArgs := []string{
		"run",
		"--rm",                                      // Remove container after execution
		"--read-only",                               // Read-only root filesystem
		"--network=" + networkMode,                  // Network mode for this skill
		"--memory=" + n.config.MemoryLimit,         // Memory limit
		"--cpus=" + n.config.CPULimit,              // CPU limit
		"--pids-limit=100",                          // Process limit
		"--security-opt=no-new-privileges",          // No privilege escalation
		"--cap-drop=ALL",                            // Drop all capabilities
		"-v", fmt.Sprintf("%s:/workspace:rw", workspaceDir),     // Read-write workspace
		"-v", fmt.Sprintf("%s:/skill:ro", skillLibsDir),         // Read-only skill libs
		"-v", fmt.Sprintf("%s:/outputs:rw", n.config.OutputsDir),  // Persistent outputs directory
		"-w", "/workspace",                          // Working directory
		"--tmpfs", "/tmp:rw,exec,size=100m",        // Writable /tmp
		image,                                       // Use skill-specific image
		"bash", scriptPath,                          // Command (relative to /workspace)
	}
	cmdArgs = append(cmdArgs, args...)

	cmd := exec.CommandContext(ctx, n.command, cmdArgs...)
	output, err := cmd.CombinedOutput()

	// Check for timeout
	if ctx.Err() == context.DeadlineExceeded {
		return "", fmt.Errorf("execution timeout after %v", n.config.Timeout)
	}

	if err != nil {
		return string(output), fmt.Errorf("code execution failed: %w\nOutput: %s", err, output)
	}

	return string(output), nil
}
