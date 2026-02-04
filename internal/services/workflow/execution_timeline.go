package workflow

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// ExecutionTimeline tracks step execution timing for visualization
type ExecutionTimeline struct {
	mu         sync.RWMutex
	events     []TimelineEvent
	startTime  time.Time
	endTime    time.Time
	maxWorkers int
}

// TimelineEvent represents a step start or end
type TimelineEvent struct {
	stepName  string
	eventType string // "start" or "end"
	timestamp time.Time
}

// StepExecution represents a complete step execution
type StepExecution struct {
	stepName  string
	startTime time.Time
	endTime   time.Time
	duration  time.Duration
}

// NewExecutionTimeline creates a new timeline tracker
func NewExecutionTimeline(maxWorkers int) *ExecutionTimeline {
	return &ExecutionTimeline{
		events:     make([]TimelineEvent, 0),
		maxWorkers: maxWorkers,
	}
}

// Start marks the beginning of workflow execution
func (et *ExecutionTimeline) Start() {
	et.mu.Lock()
	defer et.mu.Unlock()
	et.startTime = time.Now()
}

// End marks the end of workflow execution
func (et *ExecutionTimeline) End() {
	et.mu.Lock()
	defer et.mu.Unlock()
	et.endTime = time.Now()
}

// RecordStepStart records when a step starts
func (et *ExecutionTimeline) RecordStepStart(stepName string) {
	et.mu.Lock()
	defer et.mu.Unlock()

	et.events = append(et.events, TimelineEvent{
		stepName:  stepName,
		eventType: "start",
		timestamp: time.Now(),
	})
}

// RecordStepEnd records when a step ends
func (et *ExecutionTimeline) RecordStepEnd(stepName string) {
	et.mu.Lock()
	defer et.mu.Unlock()

	et.events = append(et.events, TimelineEvent{
		stepName:  stepName,
		eventType: "end",
		timestamp: time.Now(),
	})
}

// GetStepExecutions returns all complete step executions
func (et *ExecutionTimeline) GetStepExecutions() []StepExecution {
	et.mu.RLock()
	defer et.mu.RUnlock()

	// Build map of start times
	starts := make(map[string]time.Time)
	ends := make(map[string]time.Time)

	for _, event := range et.events {
		if event.eventType == "start" {
			starts[event.stepName] = event.timestamp
		} else {
			ends[event.stepName] = event.timestamp
		}
	}

	// Build executions
	var executions []StepExecution
	for stepName, startTime := range starts {
		if endTime, exists := ends[stepName]; exists {
			executions = append(executions, StepExecution{
				stepName:  stepName,
				startTime: startTime,
				endTime:   endTime,
				duration:  endTime.Sub(startTime),
			})
		}
	}

	// Sort by start time
	sort.Slice(executions, func(i, j int) bool {
		return executions[i].startTime.Before(executions[j].startTime)
	})

	return executions
}

// GetTotalDuration returns the total workflow duration
func (et *ExecutionTimeline) GetTotalDuration() time.Duration {
	et.mu.RLock()
	defer et.mu.RUnlock()

	if et.endTime.IsZero() || et.startTime.IsZero() {
		return 0
	}

	return et.endTime.Sub(et.startTime)
}

// GetParallelismLevel returns the maximum number of concurrent steps
func (et *ExecutionTimeline) GetParallelismLevel() int {
	et.mu.RLock()
	defer et.mu.RUnlock()

	if len(et.events) == 0 {
		return 0
	}

	// Sort events by timestamp
	events := make([]TimelineEvent, len(et.events))
	copy(events, et.events)
	sort.Slice(events, func(i, j int) bool {
		return events[i].timestamp.Before(events[j].timestamp)
	})

	maxConcurrent := 0
	currentConcurrent := 0

	for _, event := range events {
		if event.eventType == "start" {
			currentConcurrent++
			if currentConcurrent > maxConcurrent {
				maxConcurrent = currentConcurrent
			}
		} else {
			currentConcurrent--
		}
	}

	return maxConcurrent
}

// GenerateASCIITimeline creates an ASCII visualization of the timeline
func (et *ExecutionTimeline) GenerateASCIITimeline() string {
	executions := et.GetStepExecutions()
	if len(executions) == 0 {
		return "No executions recorded"
	}

	var sb strings.Builder
	sb.WriteString("\n")
	sb.WriteString("═══════════════════════════════════════════════════════\n")
	sb.WriteString("                 EXECUTION TIMELINE\n")
	sb.WriteString("═══════════════════════════════════════════════════════\n\n")

	// Find global start and end
	globalStart := et.startTime
	globalEnd := et.endTime
	totalDuration := globalEnd.Sub(globalStart)

	if totalDuration == 0 {
		return "Timeline duration is zero"
	}

	// Calculate time buckets (1 second resolution)
	timelineWidth := 60 // characters
	bucketDuration := totalDuration / time.Duration(timelineWidth)
	if bucketDuration < time.Millisecond {
		bucketDuration = time.Millisecond
	}

	// Group steps by time buckets
	type bucket struct {
		startTime time.Time
		steps     []string
	}

	buckets := make([]bucket, timelineWidth)
	for i := 0; i < timelineWidth; i++ {
		buckets[i] = bucket{
			startTime: globalStart.Add(time.Duration(i) * bucketDuration),
			steps:     make([]string, 0),
		}
	}

	// Place executions in buckets
	for _, exec := range executions {
		relativeStart := exec.startTime.Sub(globalStart)
		bucketIdx := int(relativeStart / bucketDuration)
		if bucketIdx >= 0 && bucketIdx < timelineWidth {
			buckets[bucketIdx].steps = append(buckets[bucketIdx].steps, exec.stepName)
		}
	}

	// Print timeline
	sb.WriteString("Time     Steps\n")
	sb.WriteString("───────  ──────────────────────────────────────────────\n")

	for i, b := range buckets {
		elapsed := time.Duration(i) * bucketDuration
		timeStr := fmt.Sprintf("T+%-5s", elapsed.Round(100*time.Millisecond).String())

		if len(b.steps) > 0 {
			stepsStr := strings.Join(b.steps, ", ")
			if len(stepsStr) > 50 {
				stepsStr = stepsStr[:47] + "..."
			}
			sb.WriteString(fmt.Sprintf("%s  %s\n", timeStr, stepsStr))
		}
	}

	sb.WriteString("\n")
	sb.WriteString("───────────────────────────────────────────────────────\n")
	sb.WriteString(fmt.Sprintf("Total Duration: %v\n", totalDuration.Round(time.Millisecond)))
	sb.WriteString(fmt.Sprintf("Max Parallelism: %d steps (limit: %d)\n",
		et.GetParallelismLevel(), et.maxWorkers))
	sb.WriteString("═══════════════════════════════════════════════════════\n")

	return sb.String()
}

// GenerateGanttChart creates a Gantt-style chart of execution
func (et *ExecutionTimeline) GenerateGanttChart() string {
	executions := et.GetStepExecutions()
	if len(executions) == 0 {
		return "No executions recorded"
	}

	var sb strings.Builder
	sb.WriteString("\n")
	sb.WriteString("═══════════════════════════════════════════════════════\n")
	sb.WriteString("                   GANTT CHART\n")
	sb.WriteString("═══════════════════════════════════════════════════════\n\n")

	globalStart := et.startTime
	totalDuration := et.GetTotalDuration()

	if totalDuration == 0 {
		return "Timeline duration is zero"
	}

	chartWidth := 50

	// Find longest step name for padding
	maxNameLen := 0
	for _, exec := range executions {
		if len(exec.stepName) > maxNameLen {
			maxNameLen = len(exec.stepName)
		}
	}

	// Print each execution as a bar
	for _, exec := range executions {
		relativeStart := exec.startTime.Sub(globalStart)
		startPos := int((float64(relativeStart) / float64(totalDuration)) * float64(chartWidth))
		barLen := int((float64(exec.duration) / float64(totalDuration)) * float64(chartWidth))

		if barLen < 1 {
			barLen = 1
		}
		if startPos+barLen > chartWidth {
			barLen = chartWidth - startPos
		}

		// Build the bar
		line := strings.Repeat(" ", chartWidth)
		lineBytes := []rune(line)
		for i := 0; i < chartWidth; i++ {
			if i >= startPos && i < startPos+barLen {
				lineBytes[i] = '█'
			}
		}

		sb.WriteString(fmt.Sprintf("%-*s |%s| %v\n",
			maxNameLen, exec.stepName, string(lineBytes), exec.duration.Round(time.Millisecond)))
	}

	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("Total: %v\n", totalDuration.Round(time.Millisecond)))
	sb.WriteString("═══════════════════════════════════════════════════════\n")

	return sb.String()
}

// GetSequentialEstimate estimates sequential execution time
func (et *ExecutionTimeline) GetSequentialEstimate() time.Duration {
	executions := et.GetStepExecutions()
	total := time.Duration(0)

	for _, exec := range executions {
		total += exec.duration
	}

	return total
}

// GetSpeedup calculates the speedup ratio
func (et *ExecutionTimeline) GetSpeedup() float64 {
	parallel := et.GetTotalDuration()
	sequential := et.GetSequentialEstimate()

	if parallel == 0 {
		return 0
	}

	return float64(sequential) / float64(parallel)
}
