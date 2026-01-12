package stdio

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/logging"
	"github.com/LaurieRhodes/mcp-cli-go/internal/providers/mcp/messages"
)

const (
	// MaxBufferSize defines the maximum buffer size for reading from stdio
	// Set to 20MB to handle large security alert responses with multiple alerts
	MaxBufferSize = 20 * 1024 * 1024 // 20MB
)

// StdioClient handles communication with a server process via stdin/stdout
type StdioClient struct {
	params           StdioServerParameters
	cmd              *exec.Cmd
	stdin            io.WriteCloser
	stdout           io.ReadCloser
	stderr           io.ReadCloser
	readChan         chan *messages.JSONRPCMessage
	writeChan        chan *messages.JSONRPCMessage
	done             chan struct{}
	ctx              context.Context
	cancel           context.CancelFunc
	wg               sync.WaitGroup
	initialized      bool
	mu               sync.Mutex
	suppressConsole  bool           // Controls console output visibility
	stderrBuffer     *bytes.Buffer  // Captures server stderr for error analysis
	stderrMutex      sync.Mutex     // Protects stderr buffer access
	hasRealErrors    bool           // Indicates if server reported ACTUAL errors (not just info/debug logs)
	dispatcher       *ResponseDispatcher // Routes responses to waiting requests
}

// NewStdioClient creates a new stdio client with the given parameters
func NewStdioClient(params StdioServerParameters) *StdioClient {
	ctx, cancel := context.WithCancel(context.Background())
	
	// Determine console output suppression based on logging level
	// If logging is ERROR or above, we want clean user output
	suppressConsole := logging.GetDefaultLevel() >= logging.ERROR
	
	return &StdioClient{
		params:          params,
		readChan:        make(chan *messages.JSONRPCMessage, 10),
		writeChan:       make(chan *messages.JSONRPCMessage, 10),
		done:            make(chan struct{}),
		ctx:             ctx,
		cancel:          cancel,
		suppressConsole: suppressConsole,
		stderrBuffer:    &bytes.Buffer{},
		hasRealErrors:   false,
	}
}

// NewStdioClientWithOptions creates a new stdio client with custom options
func NewStdioClientWithOptions(params StdioServerParameters, suppressConsole bool) *StdioClient {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &StdioClient{
		params:          params,
		readChan:        make(chan *messages.JSONRPCMessage, 10),
		writeChan:       make(chan *messages.JSONRPCMessage, 10),
		done:            make(chan struct{}),
		ctx:             ctx,
		cancel:          cancel,
		suppressConsole: suppressConsole,
		stderrBuffer:    &bytes.Buffer{},
		hasRealErrors:   false,
	}
}

// NewStdioClientWithStderrOption creates a new stdio client with stderr suppression option
func NewStdioClientWithStderrOption(params StdioServerParameters, suppressStderr bool) *StdioClient {
	ctx, cancel := context.WithCancel(context.Background())
	
	// suppressStderr controls whether to show server debug output
	return &StdioClient{
		params:          params,
		readChan:        make(chan *messages.JSONRPCMessage, 10),
		writeChan:       make(chan *messages.JSONRPCMessage, 10),
		done:            make(chan struct{}),
		ctx:             ctx,
		cancel:          cancel,
		suppressConsole: suppressStderr, // Use suppressStderr to control console output
		stderrBuffer:    &bytes.Buffer{},
		hasRealErrors:   false,
	}
}

// isRealError determines if a stderr line represents an actual error vs normal logging
func isRealError(line string) bool {
	lowerLine := strings.ToLower(line)
	
	// Ignore normal logging patterns and configuration messages
	if strings.HasPrefix(lowerLine, "debug:") ||
	   strings.HasPrefix(lowerLine, "info:") ||
	   strings.Contains(lowerLine, "loading configuration") ||
	   strings.Contains(lowerLine, "successfully obtained token") ||
	   strings.Contains(lowerLine, "api call succeeded") ||
	   strings.Contains(lowerLine, "configuration loaded") ||
	   strings.Contains(lowerLine, "server started") ||
	   strings.Contains(lowerLine, "starting on stdin/stdout") ||
	   strings.Contains(lowerLine, "registered") ||
	   strings.Contains(lowerLine, "processing") ||
	   strings.Contains(lowerLine, "sending:") ||
	   strings.Contains(lowerLine, "received:") ||
	   strings.Contains(lowerLine, "response:") ||
	   strings.Contains(lowerLine, "sending response:") ||
	   strings.Contains(lowerLine, "received message:") ||
	   strings.Contains(lowerLine, "parsed message") ||
	   strings.Contains(lowerLine, "command timeout:") || // Configuration message, not error
	   strings.Contains(lowerLine, "timeout:") || // Generic timeout config messages
	   strings.Contains(lowerLine, "looking for config") ||
	   strings.Contains(lowerLine, "reading config") ||
	   strings.Contains(lowerLine, "executable directory") {
		return false
	}
	
	// Ignore agentic loop intermediate errors (auto-corrected by follow-up iterations)
	// These are expected during multi-turn tool execution and shouldn't alarm users
	if strings.Contains(lowerLine, "code execution failed") ||
	   strings.Contains(lowerLine, "invalid file paths detected") ||
	   strings.Contains(lowerLine, "nameerror:") ||
	   strings.Contains(lowerLine, "syntaxerror:") ||
	   strings.Contains(lowerLine, "typeerror:") ||
	   strings.Contains(lowerLine, "valueerror:") ||
	   strings.Contains(lowerLine, "attributeerror:") ||
	   strings.Contains(lowerLine, "importerror:") ||
	   strings.Contains(lowerLine, "keyerror:") ||
	   strings.Contains(lowerLine, "indexerror:") {
		return false
	}
	
	// These indicate actual problems (but not timeout configs)
	return strings.Contains(lowerLine, "error:") ||
	       strings.Contains(lowerLine, "failed:") ||
	       strings.Contains(lowerLine, "panic:") ||
	       strings.Contains(lowerLine, "fatal:") ||
	       strings.Contains(lowerLine, "authentication failed") ||
	       strings.Contains(lowerLine, "connection refused") ||
	       strings.Contains(lowerLine, "timed out") || // Actual timeout event, not config
	       strings.Contains(lowerLine, "timeout exceeded") || // Actual timeout event
	       strings.Contains(lowerLine, "permission denied")
}

// Start initiates the connection to the server
func (c *StdioClient) Start() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.initialized {
		return fmt.Errorf("client already started")
	}

	logging.Debug("Starting stdio client with command: %s", c.params.Command)
	
	// Create the command
	c.cmd = exec.CommandContext(c.ctx, c.params.Command, c.params.Args...)

	// Set environment variables
	if c.params.Env != nil {
		env := os.Environ()
		for k, v := range c.params.Env {
			env = append(env, fmt.Sprintf("%s=%s", k, v))
			logging.Debug("Setting environment variable: %s=%s", k, v)
		}
		c.cmd.Env = env
	}

	// Get stdin/stdout pipes
	var err error
	c.stdin, err = c.cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdin pipe: %w", err)
	}

	c.stdout, err = c.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	// REFINED FIX: Always capture stderr for error analysis
	// but be much more intelligent about what we consider "errors"
	c.stderr, err = c.cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to get stderr pipe: %w", err)
	}

	// Start the command
	logging.Debug("Starting server process")
	if err := c.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %w", err)
	}
	logging.Info("Server process started with PID %d", c.cmd.Process.Pid)

	// Start the reader, writer, and stderr monitor goroutines
	c.wg.Add(3)
	go c.readLoop()
	go c.writeLoop()
	go c.stderrLoop()

	// Initialize and start the response dispatcher
	c.dispatcher = NewResponseDispatcher(c)
	c.dispatcher.Start()

	c.initialized = true
	return nil
}

// readLoop reads JSON-RPC messages from the server's stdout
func (c *StdioClient) readLoop() {
	defer c.wg.Done()
	defer close(c.readChan)

	logging.Debug("Starting stdout reader loop with %d MB buffer size", MaxBufferSize/(1024*1024))
	scanner := bufio.NewScanner(c.stdout)
	
	// Create a custom buffer with increased size to handle large security alert responses
	buf := make([]byte, MaxBufferSize)
	scanner.Buffer(buf, MaxBufferSize)
	
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 {
			continue
		}

		logging.Debug("Received line of length: %d bytes", len(line))

		// Check if line is valid JSON-RPC message
		var msg messages.JSONRPCMessage
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			// If not a valid JSON-RPC message, only log at debug level
			// This prevents non-JSON server output from cluttering console
			logging.Debug("Received non-JSON line: %s", line)
			continue
		}

		// Valid JSON-RPC message
		logging.Debug("Received data: %s", line)
		logging.Debug("Parsed message ID: %s, Method: %s", msg.ID, msg.Method)
		select {
		case c.readChan <- &msg:
			logging.Debug("Message sent to read channel successfully")
		case <-c.ctx.Done():
			logging.Debug("Context done, exiting read loop")
			return
		}
	}

	if err := scanner.Err(); err != nil {
		logging.Error("Error reading from stdout: %v", err)
	}
	logging.Debug("Exiting stdout reader loop")
}

// stderrLoop monitors server stderr for ACTUAL errors (not normal logging)
func (c *StdioClient) stderrLoop() {
	defer c.wg.Done()

	logging.Debug("Starting stderr monitor loop")
	scanner := bufio.NewScanner(c.stderr)
	
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 {
			continue
		}

		// Always capture stderr for potential debugging needs
		c.stderrMutex.Lock()
		c.stderrBuffer.WriteString(line + "\n")
		
		// REFINED: Only flag as error if it's actually an error, not normal logging
		if isRealError(line) {
			c.hasRealErrors = true
		}
		c.stderrMutex.Unlock()

		// Log the stderr output at debug level for troubleshooting
		logging.Debug("Server stderr: %s", line)

		// REFINED DISPLAY LOGIC: Only display to user if:
		// 1. Not suppressing console output (verbose/noisy mode), OR
		// 2. This line contains an ACTUAL error (not just info/debug logging)
		if !c.suppressConsole {
			// In verbose/noisy mode, show all server output
			fmt.Fprintf(os.Stderr, "%s\n", line)
		} else if isRealError(line) {
			// In clean mode, only show actual errors
			fmt.Fprintf(os.Stderr, "Server Error: %s\n", line)
		}
		// If suppressConsole=true AND this is just normal logging, don't show it
	}

	if err := scanner.Err(); err != nil {
		logging.Error("Error reading from stderr: %v", err)
	}
	logging.Debug("Exiting stderr monitor loop")
}

// writeLoop sends JSON-RPC messages to the server's stdin
func (c *StdioClient) writeLoop() {
	defer c.wg.Done()

	logging.Debug("Starting stdin writer loop")
	for {
		select {
		case msg, ok := <-c.writeChan:
			if !ok {
				logging.Debug("Write channel closed, exiting write loop")
				return
			}

			data, err := json.Marshal(msg)
			if err != nil {
				logging.Error("Error marshaling JSON-RPC message: %v", err)
				continue
			}

			// Add newline to delimit messages
			data = append(data, '\n')

			logging.Debug("Sending data: %s", string(data))
			if _, err := c.stdin.Write(data); err != nil {
				logging.Error("Error writing to stdin: %v", err)
				c.Stop()
				return
			}
			logging.Debug("Data sent successfully")

		case <-c.ctx.Done():
			logging.Debug("Context done, exiting write loop")
			return
		}
	}
}

// Read returns a channel of incoming JSON-RPC messages
func (c *StdioClient) Read() <-chan *messages.JSONRPCMessage {
	return c.readChan
}

// Write sends a JSON-RPC message to the server
func (c *StdioClient) Write(msg *messages.JSONRPCMessage) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.initialized {
		return fmt.Errorf("client not started")
	}

	logging.Debug("Queuing message to write: ID=%s, Method=%s", msg.ID, msg.Method)
	select {
	case c.writeChan <- msg:
		logging.Debug("Message queued successfully")
		return nil
	case <-c.ctx.Done():
		return fmt.Errorf("client stopped")
	}
}

// Stop terminates the connection to the server
func (c *StdioClient) Stop() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.initialized {
		logging.Debug("Client not initialized, nothing to stop")
		return
	}

	logging.Debug("Stopping stdio client")

	// Cancel the context to signal goroutines to stop
	c.cancel()
	logging.Debug("Context cancelled")

	// Close the write channel to stop the write loop
	close(c.writeChan)
	logging.Debug("Write channel closed")

	// Wait for goroutines to finish
	logging.Debug("Waiting for goroutines to finish")
	c.wg.Wait()
	logging.Debug("Goroutines finished")

	// Close the stdin pipe to ensure the process receives EOF
	if c.stdin != nil {
		logging.Debug("Closing stdin pipe")
		c.stdin.Close()
	}

	// Close other pipes
	if c.stdout != nil {
		c.stdout.Close()
	}
	if c.stderr != nil {
		c.stderr.Close()
	}

	// Wait for the process to exit
	if c.cmd.Process != nil {
		logging.Debug("Attempting to terminate process with PID %d", c.cmd.Process.Pid)
		// First try to terminate gracefully
		if err := c.cmd.Process.Signal(os.Interrupt); err != nil {
			logging.Warn("Failed to send interrupt signal: %v", err)
			logging.Debug("Forcefully killing process")
			_ = c.cmd.Process.Kill()
		}

		// Wait for the process to exit
		logging.Debug("Waiting for process to exit")
		err := c.cmd.Wait()
		if err != nil {
			logging.Debug("Process exited with error: %v", err)
			// Only show error diagnostics if there were ACTUAL errors
			c.showErrorDiagnostics()
		} else {
			logging.Debug("Process exited successfully")
		}
	}

	logging.Info("Stdio client stopped")
	c.initialized = false
}

// showErrorDiagnostics displays captured stderr only if there were REAL errors
func (c *StdioClient) showErrorDiagnostics() {
	c.stderrMutex.Lock()
	defer c.stderrMutex.Unlock()
	
	// REFINED: Only show diagnostics if there were actual errors, not normal logging
	if c.hasRealErrors && c.stderrBuffer.Len() > 0 {
		logging.Error("Server reported actual errors. Stderr output:")
		fmt.Fprintf(os.Stderr, "\n--- Server Error Output ---\n")
		
		// Filter the buffer to only show actual error lines
		lines := strings.Split(c.stderrBuffer.String(), "\n")
		hasErrorsToShow := false
		for _, line := range lines {
			if isRealError(line) {
				fmt.Fprintf(os.Stderr, "%s\n", line)
				hasErrorsToShow = true
			}
		}
		
		if hasErrorsToShow {
			fmt.Fprintf(os.Stderr, "--- End Server Error Output ---\n\n")
		} else {
			fmt.Fprintf(os.Stderr, "(No actual errors found - process may have exited normally)\n")
			fmt.Fprintf(os.Stderr, "--- End Server Error Output ---\n\n")
		}
	}
}

// IsRunning returns whether the client is running
func (c *StdioClient) IsRunning() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.initialized
}

// SetSuppressConsole sets whether to suppress console output
func (c *StdioClient) SetSuppressConsole(suppress bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.suppressConsole = suppress
}

// GetSuppressConsole returns whether console output is being suppressed
func (c *StdioClient) GetSuppressConsole() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.suppressConsole
}

// HasRealErrors returns whether the server reported any ACTUAL errors (not just logs)
func (c *StdioClient) HasRealErrors() bool {
	c.stderrMutex.Lock()
	defer c.stderrMutex.Unlock()
	return c.hasRealErrors
}

// GetStderrContent returns the captured stderr content
func (c *StdioClient) GetStderrContent() string {
	c.stderrMutex.Lock()
	defer c.stderrMutex.Unlock()
	return c.stderrBuffer.String()
}

// Legacy methods for backward compatibility - deprecated
// SetQuiet is deprecated, use SetSuppressConsole instead
func (c *StdioClient) SetQuiet(quiet bool) {
	logging.Warn("SetQuiet is deprecated, use SetSuppressConsole instead")
	c.SetSuppressConsole(quiet)
}

// HasErrors is deprecated, use HasRealErrors instead
func (c *StdioClient) HasErrors() bool {
	logging.Warn("HasErrors is deprecated, use HasRealErrors instead")
	return c.HasRealErrors()
}

// SetSuppressStderr is deprecated - stderr is now always captured but intelligently managed
func (c *StdioClient) SetSuppressStderr(suppress bool) {
	logging.Warn("SetSuppressStderr is deprecated - stderr is now always captured for error analysis but display is controlled by suppressConsole")
	// No-op - we always capture stderr now but manage display intelligently
}

// GetSuppressStderr is deprecated - always returns false since we always capture stderr
func (c *StdioClient) GetSuppressStderr() bool {
	logging.Warn("GetSuppressStderr is deprecated - stderr is now always captured")
	return false // We always capture stderr now
}

// GetDispatcher returns the response dispatcher (for concurrent request handling)
func (c *StdioClient) GetDispatcher() *ResponseDispatcher {
	return c.dispatcher
}

