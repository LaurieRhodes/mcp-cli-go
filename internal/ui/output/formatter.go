package output

import (
	"encoding/json"
	"fmt"
	"io"
)

// Format represents output format type
type Format string

const (
	FormatJSON   Format = "json"
	FormatText   Format = "text"
	FormatPretty Format = "pretty"
)

// Formatter handles output formatting
type Formatter struct {
	format Format
	writer io.Writer
	indent string
}

// NewFormatter creates a new formatter
func NewFormatter(format Format, writer io.Writer) *Formatter {
	return &Formatter{
		format: format,
		writer: writer,
		indent: "  ",
	}
}

// Format formats and writes output
func (f *Formatter) Format(data interface{}) error {
	switch f.format {
	case FormatJSON:
		return f.formatJSON(data)
	case FormatText:
		return f.formatText(data)
	case FormatPretty:
		return f.formatPretty(data)
	default:
		return f.formatText(data)
	}
}

// formatJSON formats as JSON
func (f *Formatter) formatJSON(data interface{}) error {
	encoder := json.NewEncoder(f.writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// formatText formats as plain text
func (f *Formatter) formatText(data interface{}) error {
	_, err := fmt.Fprintln(f.writer, data)
	return err
}

// formatPretty formats with pretty printing
func (f *Formatter) formatPretty(data interface{}) error {
	// Try to pretty print as JSON first
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		// Fall back to text format
		return f.formatText(data)
	}

	_, err = f.writer.Write(jsonData)
	if err != nil {
		return err
	}

	_, err = f.writer.Write([]byte("\n"))
	return err
}

// SetIndent sets the indentation string
func (f *Formatter) SetIndent(indent string) {
	f.indent = indent
}

// Write writes raw data
func (f *Formatter) Write(data []byte) (int, error) {
	return f.writer.Write(data)
}

// WriteString writes a string
func (f *Formatter) WriteString(s string) (int, error) {
	return f.writer.Write([]byte(s))
}

// WriteLine writes a line with newline
func (f *Formatter) WriteLine(s string) error {
	_, err := f.writer.Write([]byte(s + "\n"))
	return err
}

// FormatResponse is a helper for formatting common response types
type FormatResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// NewSuccessResponse creates a success response
func NewSuccessResponse(message string, data interface{}) *FormatResponse {
	return &FormatResponse{
		Success: true,
		Message: message,
		Data:    data,
	}
}

// NewErrorResponse creates an error response
func NewErrorResponse(err error) *FormatResponse {
	return &FormatResponse{
		Success: false,
		Error:   err.Error(),
	}
}
