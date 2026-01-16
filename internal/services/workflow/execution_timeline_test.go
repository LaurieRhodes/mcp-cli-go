package workflow

import (
	"testing"
	"time"
)

func TestExecutionTimeline_Basic(t *testing.T) {
	timeline := NewExecutionTimeline(3)
	timeline.Start()
	
	// Simulate step executions
	timeline.RecordStepStart("step1")
	time.Sleep(10 * time.Millisecond)
	timeline.RecordStepEnd("step1")
	
	timeline.RecordStepStart("step2")
	time.Sleep(15 * time.Millisecond)
	timeline.RecordStepEnd("step2")
	
	timeline.End()
	
	// Check executions
	executions := timeline.GetStepExecutions()
	if len(executions) != 2 {
		t.Errorf("Expected 2 executions, got %d", len(executions))
	}
	
	// Check duration
	totalDuration := timeline.GetTotalDuration()
	if totalDuration == 0 {
		t.Error("Total duration should not be zero")
	}
}

func TestExecutionTimeline_Parallelism(t *testing.T) {
	timeline := NewExecutionTimeline(3)
	timeline.Start()
	
	// Simulate parallel execution
	startTime := time.Now()
	timeline.RecordStepStart("step1")
	timeline.RecordStepStart("step2")
	timeline.RecordStepStart("step3")
	
	time.Sleep(20 * time.Millisecond)
	
	timeline.RecordStepEnd("step1")
	timeline.RecordStepEnd("step2")
	timeline.RecordStepEnd("step3")
	timeline.End()
	
	// Check parallelism level
	level := timeline.GetParallelismLevel()
	if level != 3 {
		t.Errorf("Expected parallelism level 3, got %d", level)
	}
	
	// Check that parallel duration is less than sequential would be
	totalDuration := timeline.GetTotalDuration()
	if totalDuration > 30*time.Millisecond {
		t.Errorf("Parallel duration too high: %v", totalDuration)
	}
	
	// Verify we didn't start recording before actual start
	if totalDuration > time.Since(startTime) + 5*time.Millisecond {
		t.Errorf("Timeline total duration exceeds actual time")
	}
}

func TestExecutionTimeline_Speedup(t *testing.T) {
	timeline := NewExecutionTimeline(3)
	timeline.Start()
	
	// Simulate 3 steps running in parallel (each 10ms)
	timeline.RecordStepStart("step1")
	timeline.RecordStepStart("step2")
	timeline.RecordStepStart("step3")
	
	time.Sleep(10 * time.Millisecond)
	
	timeline.RecordStepEnd("step1")
	timeline.RecordStepEnd("step2")
	timeline.RecordStepEnd("step3")
	timeline.End()
	
	// Sequential would be ~30ms, parallel is ~10ms
	sequential := timeline.GetSequentialEstimate()
	parallel := timeline.GetTotalDuration()
	speedup := timeline.GetSpeedup()
	
	// Speedup should be close to 3x
	if speedup < 2.0 {
		t.Errorf("Speedup too low: %.2fx (sequential: %v, parallel: %v)", 
			speedup, sequential, parallel)
	}
}

func TestExecutionTimeline_GanttChart(t *testing.T) {
	timeline := NewExecutionTimeline(2)
	timeline.Start()
	
	timeline.RecordStepStart("step1")
	time.Sleep(5 * time.Millisecond)
	timeline.RecordStepEnd("step1")
	
	timeline.RecordStepStart("step2")
	time.Sleep(5 * time.Millisecond)
	timeline.RecordStepEnd("step2")
	
	timeline.End()
	
	chart := timeline.GenerateGanttChart()
	
	// Check that chart contains expected elements
	if len(chart) == 0 {
		t.Error("Gantt chart should not be empty")
	}
	
	if !containsString(chart, "step1") {
		t.Error("Chart should contain step1")
	}
	
	if !containsString(chart, "step2") {
		t.Error("Chart should contain step2")
	}
	
	if !containsString(chart, "█") {
		t.Error("Chart should contain bar characters")
	}
}

func TestExecutionTimeline_ASCIITimeline(t *testing.T) {
	timeline := NewExecutionTimeline(2)
	timeline.Start()
	
	timeline.RecordStepStart("step1")
	time.Sleep(5 * time.Millisecond)
	timeline.RecordStepEnd("step1")
	
	timeline.End()
	
	asciiTimeline := timeline.GenerateASCIITimeline()
	
	if len(asciiTimeline) == 0 {
		t.Error("ASCII timeline should not be empty")
	}
	
	if !containsString(asciiTimeline, "step1") {
		t.Error("Timeline should contain step1")
	}
}

// Helper function
func containsString(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && 
		(s == substr || len(s) > len(substr) && 
		(s[:len(substr)] == substr || 
		s[len(s)-len(substr):] == substr ||
		findInString(s, substr)))
}

func findInString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
