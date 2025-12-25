package output

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/models"
	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/ports"
	"github.com/LaurieRhodes/mcp-cli-go/internal/ui/console"
)

// Manager implements ports.OutputManager
type Manager struct {
	config *models.OutputConfig
	writer ports.OutputWriter
	mu     sync.RWMutex
}

// NewManager creates a new output manager
func NewManager(config *models.OutputConfig) *Manager {
	if config == nil {
		config = models.NewDefaultOutputConfig()
	}
	
	return &Manager{
		config: config,
		writer: NewWriter(config, os.Stdout),
	}
}

// GetWriter returns the output writer
func (m *Manager) GetWriter() ports.OutputWriter {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.writer
}

// SetVerbosity sets the output verbosity level
func (m *Manager) SetVerbosity(level models.OutputLevel) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.config.Level = level
	// Update writer config
	if w, ok := m.writer.(*Writer); ok {
		w.SetConfig(m.config)
	}
}

// GetVerbosity returns the current verbosity level
func (m *Manager) GetVerbosity() models.OutputLevel {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config.Level
}

// ShouldShowConnectionMessages returns true if connection messages should be shown
func (m *Manager) ShouldShowConnectionMessages() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config.ShouldShowConnectionMessages()
}

// ShouldShowStartupInfo returns true if startup info should be shown
func (m *Manager) ShouldShowStartupInfo() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config.ShouldShowStartupInfo()
}

// ShouldSuppressServerStderr returns true if server stderr should be suppressed
func (m *Manager) ShouldSuppressServerStderr() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config.SuppressServerStderr
}

// Writer implements ports.OutputWriter
type Writer struct {
	config *models.OutputConfig
	output io.Writer
	mu     sync.RWMutex
}

// NewWriter creates a new output writer
func NewWriter(config *models.OutputConfig, output io.Writer) *Writer {
	if config == nil {
		config = models.NewDefaultOutputConfig()
	}
	if output == nil {
		output = os.Stdout
	}
	
	// Configure console colors
	console.SetColorsEnabled(config.ShowColors)
	
	return &Writer{
		config: config,
		output: output,
	}
}

// WriteInfo writes an informational message
func (w *Writer) WriteInfo(format string, args ...interface{}) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	
	if w.config.ShouldShow(models.OutputNormal) {
		msg := fmt.Sprintf(format, args...)
		fmt.Fprintln(w.output, console.Info(msg))
	}
}

// WriteSuccess writes a success message
func (w *Writer) WriteSuccess(format string, args ...interface{}) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	
	if w.config.ShouldShow(models.OutputQuiet) {
		msg := fmt.Sprintf(format, args...)
		fmt.Fprintln(w.output, console.Success(msg))
	}
}

// WriteError writes an error message
func (w *Writer) WriteError(format string, args ...interface{}) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	
	// Always show errors
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintln(w.output, console.Error(msg))
}

// WriteWarning writes a warning message
func (w *Writer) WriteWarning(format string, args ...interface{}) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	
	if w.config.ShouldShow(models.OutputNormal) {
		msg := fmt.Sprintf(format, args...)
		fmt.Fprintln(w.output, console.Warning(msg))
	}
}

// WriteDebug writes a debug message
func (w *Writer) WriteDebug(format string, args ...interface{}) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	
	if w.config.ShouldShow(models.OutputVerbose) {
		msg := fmt.Sprintf(format, args...)
		if w.config.ShowTimestamps {
			msg = fmt.Sprintf("[%s] %s", time.Now().Format("15:04:05"), msg)
		}
		fmt.Fprintln(w.output, console.Dim(msg))
	}
}

// WriteLine writes a plain line
func (w *Writer) WriteLine(format string, args ...interface{}) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintln(w.output, msg)
}

// GetConfig returns the current output configuration
func (w *Writer) GetConfig() *models.OutputConfig {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.config.Clone()
}

// SetConfig updates the output configuration
func (w *Writer) SetConfig(config *models.OutputConfig) {
	w.mu.Lock()
	defer w.mu.Unlock()
	
	w.config = config
	console.SetColorsEnabled(config.ShowColors)
}

// Global output manager instance
var globalManager *Manager
var globalOnce sync.Once

// GetGlobalManager returns the global output manager
func GetGlobalManager() *Manager {
	globalOnce.Do(func() {
		globalManager = NewManager(models.NewDefaultOutputConfig())
	})
	return globalManager
}

// SetGlobalManager sets the global output manager
func SetGlobalManager(manager *Manager) {
	globalManager = manager
}
