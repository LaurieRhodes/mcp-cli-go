package sandbox

import (
	"testing"
)

func TestDetectExecutor(t *testing.T) {
	config := DefaultConfig()
	executor, err := DetectExecutor(config)
	
	if err != nil {
		t.Skip("Docker not available:", err)
		return
	}
	
	if !executor.IsAvailable() {
		t.Fatal("Executor reports unavailable but should be available")
	}
	
	info := executor.GetInfo()
	t.Logf("✅ Detected executor: %s", info)
}

func TestIsRunningInContainer(t *testing.T) {
	inContainer := isRunningInContainer()
	t.Logf("Running in container: %v", inContainer)
}
