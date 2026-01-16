package workflow

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// BufferedLogger captures logs per step for ordered output
type BufferedLogger struct {
	mu      sync.RWMutex
	buffers map[string]*StepLogBuffer
	order   []string // Track execution order for flushing
	enabled bool
}

// StepLogBuffer holds logs for a single step
type StepLogBuffer struct {
	stepName  string
	logs      []LogEntry
	startTime time.Time
	endTime   time.Time
	mu        sync.Mutex
}

// LogEntry represents a single log message
type LogEntry struct {
	timestamp time.Time
	level     string // INFO, DEBUG, ERROR, OUTPUT
	message   string
}

// NewBufferedLogger creates a new buffered logger
func NewBufferedLogger() *BufferedLogger {
	return &BufferedLogger{
		buffers: make(map[string]*StepLogBuffer),
		order:   make([]string, 0),
		enabled: true,
	}
}

// StartStep begins buffering for a step
func (bl *BufferedLogger) StartStep(stepName string) {
	bl.mu.Lock()
	defer bl.mu.Unlock()

	buffer := &StepLogBuffer{
		stepName:  stepName,
		logs:      make([]LogEntry, 0),
		startTime: time.Now(),
	}
	
	bl.buffers[stepName] = buffer
	bl.order = append(bl.order, stepName)
}

// EndStep marks step completion
func (bl *BufferedLogger) EndStep(stepName string) {
	bl.mu.RLock()
	buffer, exists := bl.buffers[stepName]
	bl.mu.RUnlock()

	if exists {
		buffer.mu.Lock()
		buffer.endTime = time.Now()
		buffer.mu.Unlock()
	}
}

// Log adds a log entry to the step's buffer
func (bl *BufferedLogger) Log(stepName, level, format string, args ...interface{}) {
	if !bl.enabled {
		return
	}

	bl.mu.RLock()
	buffer, exists := bl.buffers[stepName]
	bl.mu.RUnlock()

	if !exists {
		// If buffer doesn't exist, create it
		bl.StartStep(stepName)
		bl.mu.RLock()
		buffer = bl.buffers[stepName]
		bl.mu.RUnlock()
	}

	message := fmt.Sprintf(format, args...)
	
	buffer.mu.Lock()
	buffer.logs = append(buffer.logs, LogEntry{
		timestamp: time.Now(),
		level:     level,
		message:   message,
	})
	buffer.mu.Unlock()
}

// FlushInOrder flushes all buffers in execution order
func (bl *BufferedLogger) FlushInOrder(logger *Logger) {
	bl.mu.RLock()
	order := make([]string, len(bl.order))
	copy(order, bl.order)
	bl.mu.RUnlock()

	for _, stepName := range order {
		bl.FlushStep(stepName, logger)
	}
}

// FlushStep flushes a single step's buffer
func (bl *BufferedLogger) FlushStep(stepName string, logger *Logger) {
	bl.mu.RLock()
	buffer, exists := bl.buffers[stepName]
	bl.mu.RUnlock()

	if !exists {
		return
	}

	buffer.mu.Lock()
	defer buffer.mu.Unlock()

	// Calculate duration
	var duration time.Duration
	if !buffer.endTime.IsZero() {
		duration = buffer.endTime.Sub(buffer.startTime)
	}

	// Print step header
	logger.Info("─────────────────────────────────────────────────────")
	if duration > 0 {
		logger.Info("Step: %s (duration: %v)", stepName, duration.Round(time.Millisecond))
	} else {
		logger.Info("Step: %s", stepName)
	}
	logger.Info("─────────────────────────────────────────────────────")

	// Print all logs
	for _, entry := range buffer.logs {
		switch entry.level {
		case "INFO":
			logger.Info(entry.message)
		case "DEBUG":
			logger.Debug(entry.message)
		case "ERROR":
			logger.Error(entry.message)
		case "OUTPUT":
			logger.Output(entry.message)
		default:
			logger.Info(entry.message)
		}
	}

	logger.Info("") // Blank line after step
}

// GetBuffer returns the buffer for a step (for testing)
func (bl *BufferedLogger) GetBuffer(stepName string) *StepLogBuffer {
	bl.mu.RLock()
	defer bl.mu.RUnlock()
	return bl.buffers[stepName]
}

// Clear clears all buffers
func (bl *BufferedLogger) Clear() {
	bl.mu.Lock()
	defer bl.mu.Unlock()
	
	bl.buffers = make(map[string]*StepLogBuffer)
	bl.order = make([]string, 0)
}

// Disable disables buffering (logs go through immediately)
func (bl *BufferedLogger) Disable() {
	bl.enabled = false
}

// Enable enables buffering
func (bl *BufferedLogger) Enable() {
	bl.enabled = true
}

// GetExecutionSummary returns a summary of all step executions
func (bl *BufferedLogger) GetExecutionSummary() string {
	bl.mu.RLock()
	defer bl.mu.RUnlock()

	var sb strings.Builder
	sb.WriteString("\n═══════════════════════════════════════════════════════\n")
	sb.WriteString("                 EXECUTION SUMMARY\n")
	sb.WriteString("═══════════════════════════════════════════════════════\n\n")

	totalDuration := time.Duration(0)
	
	for _, stepName := range bl.order {
		buffer := bl.buffers[stepName]
		if buffer == nil {
			continue
		}

		buffer.mu.Lock()
		duration := time.Duration(0)
		if !buffer.endTime.IsZero() {
			duration = buffer.endTime.Sub(buffer.startTime)
			totalDuration += duration
		}
		
		status := "✓"
		hasError := false
		for _, log := range buffer.logs {
			if log.level == "ERROR" {
				hasError = true
				break
			}
		}
		if hasError {
			status = "✗"
		}
		
		sb.WriteString(fmt.Sprintf("%s %-30s %8v\n", status, stepName, duration.Round(time.Millisecond)))
		buffer.mu.Unlock()
	}

	sb.WriteString("\n───────────────────────────────────────────────────────\n")
	sb.WriteString(fmt.Sprintf("Total Duration: %v\n", totalDuration.Round(time.Millisecond)))
	sb.WriteString("═══════════════════════════════════════════════════════\n")

	return sb.String()
}
