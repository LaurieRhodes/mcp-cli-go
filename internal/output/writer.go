package output

import (
	"fmt"
	"io"
	"os"
	"runtime"
)

// Writer provides platform-aware output functionality
type Writer struct {
	stdout io.Writer
	stderr io.Writer
}

// NewWriter creates a platform-aware output writer
func NewWriter() *Writer {
	return &Writer{
		stdout: getStdoutWriter(),
		stderr: os.Stderr,
	}
}

// getStdoutWriter returns the appropriate stdout writer for the platform
func getStdoutWriter() io.Writer {
	// On Linux/Unix, use stderr as a fallback for clean output
	// This avoids any issues with stdout being captured by subprocess management
	// while still being visible to the user
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		// Use stderr - it's always visible and doesn't get captured
		return os.Stderr
	}
	
	// On Windows, use standard stdout
	return os.Stdout
}

// Println writes a line to stdout with platform-appropriate handling
func (w *Writer) Println(msg string) error {
	_, err := fmt.Fprintln(w.stdout, msg)
	
	// Force flush if writer supports it
	if syncer, ok := w.stdout.(interface{ Sync() error }); ok {
		syncer.Sync()
	}
	
	return err
}

// Print writes to stdout without newline
func (w *Writer) Print(msg string) error {
	_, err := fmt.Fprint(w.stdout, msg)
	
	// Force flush
	if syncer, ok := w.stdout.(interface{ Sync() error }); ok {
		syncer.Sync()
	}
	
	return err
}

// Printf writes formatted output to stdout
func (w *Writer) Printf(format string, args ...interface{}) error {
	_, err := fmt.Fprintf(w.stdout, format, args...)
	
	// Force flush
	if syncer, ok := w.stdout.(interface{ Sync() error }); ok {
		syncer.Sync()
	}
	
	return err
}

// Errorln writes a line to stderr
func (w *Writer) Errorln(msg string) error {
	_, err := fmt.Fprintln(w.stderr, msg)
	return err
}

// Errorf writes formatted output to stderr
func (w *Writer) Errorf(format string, args ...interface{}) error {
	_, err := fmt.Fprintf(w.stderr, format, args...)
	return err
}

// Close closes any open resources
func (w *Writer) Close() error {
	if closer, ok := w.stdout.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}
