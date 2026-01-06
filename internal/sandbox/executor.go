package sandbox

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/logging"
)

// Executor executes scripts in sandboxed environments
type Executor interface {
	// IsAvailable checks if the executor can run
	IsAvailable() bool

	// ExecutePython runs a Python script in a sandbox
	ExecutePython(ctx context.Context, skillDir, scriptPath string, args []string) (string, error)

	// ExecuteBash runs a Bash script in a sandbox
	ExecuteBash(ctx context.Context, skillDir, scriptPath string, args []string) (string, error)
	
	// ExecutePythonCode runs Python code with dual mount support
	// workspaceDir: read-write workspace for files and code execution
	// skillLibsDir: read-only skill directory for importing helper libraries
	ExecutePythonCode(ctx context.Context, workspaceDir, skillLibsDir, scriptPath string, args []string) (string, error)

	// GetInfo returns executor information
	GetInfo() string
}

// ExecutorConfig holds common configuration
type ExecutorConfig struct {
	PythonImage  string
	Timeout      time.Duration
	MemoryLimit  string
	CPULimit     string
	OutputsDir   string      // Persistent directory for skill outputs
	ImageMapping interface{} // Holds *skills.SkillImageMapping to avoid circular dependency
}

// DefaultConfig returns default executor configuration
func DefaultConfig() ExecutorConfig {
	return ExecutorConfig{
		PythonImage: "python:3.11-alpine",
		Timeout:     30 * time.Second,
		MemoryLimit: "256m",
		CPULimit:    "0.5",
		OutputsDir:   "/media/laurie/Data/outputs",
	}
}

// GetImageForSkill returns the appropriate image for a skill based on its directory path
// If a mapping is configured, uses that; otherwise falls back to default PythonImage
func (c *ExecutorConfig) GetImageForSkill(skillLibsDir string) string {
	// Extract skill name from path (e.g., /path/to/skills/docx -> "docx")
	skillName := filepath.Base(skillLibsDir)
	
	// If no mapping, use default
	if c.ImageMapping == nil {
		return c.PythonImage
	}
	
	// Type assert the mapping (avoid circular dependency by using interface{})
	type imageMapper interface {
		GetImageForSkill(string) string
	}
	
	if mapper, ok := c.ImageMapping.(imageMapper); ok {
		image := mapper.GetImageForSkill(skillName)
		// Log the image being used for debugging
		logging.Debug("Skill '%s' -> Image '%s' (from mapping)", skillName, image)
		return image
	}
	
	logging.Debug("Skill '%s' -> Image '%s' (default, no mapping)", skillName, c.PythonImage)
	return c.PythonImage
}

// DetectExecutor determines which executor to use
func DetectExecutor(config ExecutorConfig) (Executor, error) {
	// Check if we're running in a container
	if isRunningInContainer() {
		// Try DooD executor first (for containerized deployments)
		if exec, err := NewDooDockerExecutor(config); err == nil && exec.IsAvailable() {
			return exec, nil
		}
	}

	// Fall back to native executor (for native deployments)
	if exec, err := NewNativeExecutor(config); err == nil && exec.IsAvailable() {
		return exec, nil
	}

	return nil, fmt.Errorf("no Docker executor available")
}

// isRunningInContainer checks if we're inside a Docker container
func isRunningInContainer() bool {
	// Check for /.dockerenv file (most reliable indicator)
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}

	// Check cgroup for docker/containerd
	data, err := os.ReadFile("/proc/1/cgroup")
	if err == nil {
		content := string(data)
		if strings.Contains(content, "docker") || 
		   strings.Contains(content, "containerd") ||
		   strings.Contains(content, "kubepods") {
			return true
		}
	}

	return false
}
