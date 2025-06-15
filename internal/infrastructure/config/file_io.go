package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// ReadFile reads the content of a file
func ReadFile(path string) ([]byte, error) {
	// Ensure the path is absolute
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}
	
	// Read the file
	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	
	return data, nil
}

// WriteFile writes data to a file
func WriteFile(path string, data []byte) error {
	// Ensure the path is absolute
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}
	
	// Write the file
	if err := os.WriteFile(absPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	
	return nil
}

// FileExists checks if a file exists
func FileExists(path string) bool {
	// Ensure the path is absolute
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	
	// Check if the file exists
	_, err = os.Stat(absPath)
	return err == nil
}

// CreateConfigDirectory creates the directory for a config file if it doesn't exist
func CreateConfigDirectory(configPath string) error {
	// Get the directory part of the path
	dir := filepath.Dir(configPath)
	
	// Create the directory if it doesn't exist
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}
	
	return nil
}
