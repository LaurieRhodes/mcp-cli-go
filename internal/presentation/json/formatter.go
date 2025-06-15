package json

import (
	"encoding/json"
	"fmt"
)

// Formatter provides consistent JSON formatting across commands
type Formatter struct{}

// NewFormatter creates a new JSON formatter
func NewFormatter() *Formatter {
	return &Formatter{}
}

// Format formats any data structure as pretty-printed JSON
func (f *Formatter) Format(data interface{}) ([]byte, error) {
	return json.MarshalIndent(data, "", "  ")
}

// FormatString formats any data structure as a JSON string
func (f *Formatter) FormatString(data interface{}) (string, error) {
	bytes, err := f.Format(data)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// FormatCompact formats data as compact JSON without indentation
func (f *Formatter) FormatCompact(data interface{}) ([]byte, error) {
	return json.Marshal(data)
}

// FormatCompactString formats data as a compact JSON string
func (f *Formatter) FormatCompactString(data interface{}) (string, error) {
	bytes, err := f.FormatCompact(data)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// ValidateJSON checks if a string is valid JSON
func (f *Formatter) ValidateJSON(jsonStr string) error {
	var js interface{}
	return json.Unmarshal([]byte(jsonStr), &js)
}

// PrettyPrint formats and prints JSON data to stdout
func (f *Formatter) PrettyPrint(data interface{}) error {
	formatted, err := f.FormatString(data)
	if err != nil {
		return fmt.Errorf("failed to format JSON: %w", err)
	}
	fmt.Println(formatted)
	return nil
}
