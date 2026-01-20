package sandbox

import (
	"time"
)

// ExecutorConfigV2 extends ExecutorConfig with support for V2 features
type ExecutorConfigV2 struct {
	*ExecutorConfig
	// V2-specific fields can go here if needed
}

// GetResourceLimitsForSkill returns resource limits for a skill
// Returns: memory (bytes), cpu (cores), timeout (duration)
func (c *ExecutorConfig) GetResourceLimitsForSkill(skillLibsDir string) (int64, float64, time.Duration) {
	// Default values
	defaultMemory := int64(256 * 1024 * 1024) // 256MB
	defaultCPU := 0.5
	defaultTimeout := 60 * time.Second
	
	// TODO: Check if mapping supports V2 and extract resource limits
	// For now, return defaults (will implement full resolution in next phase)
	
	// V1 config or error - use defaults
	return defaultMemory, defaultCPU, defaultTimeout
}

func extractSkillName(skillLibsDir string) string {
	// Extract skill name from path
	// e.g., /path/to/skills/docx -> "docx"
	parts := []rune(skillLibsDir)
	for i := len(parts) - 1; i >= 0; i-- {
		if parts[i] == '/' || parts[i] == '\\' {
			return string(parts[i+1:])
		}
	}
	return skillLibsDir
}
