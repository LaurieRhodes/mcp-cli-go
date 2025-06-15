package streaming

import (
	"io"
)

// Writer handles streaming output formatting
type Writer struct {
	output io.Writer
}

// NewWriter creates a new streaming writer
func NewWriter(output io.Writer) *Writer {
	return &Writer{
		output: output,
	}
}

// Write implements io.Writer interface
func (w *Writer) Write(p []byte) (n int, err error) {
	return w.output.Write(p)
}

// WriteString writes a string to the output
func (w *Writer) WriteString(s string) (n int, err error) {
	return w.output.Write([]byte(s))
}

// Flush ensures all data is written (if the underlying writer supports it)
func (w *Writer) Flush() error {
	if flusher, ok := w.output.(interface{ Flush() error }); ok {
		return flusher.Flush()
	}
	return nil
}
