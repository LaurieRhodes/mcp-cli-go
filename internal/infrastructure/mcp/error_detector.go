package mcp

import (
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/logging"
	"github.com/LaurieRhodes/mcp-cli-go/internal/providers/mcp/messages/tools"
)

// ErrorDetector provides enhanced error detection for MCP tool results
// Following MCP spec and Gemini CLI's approach, we check for errors in multiple formats:
// 1. Top-level isError field (MCP spec compliant)
// 2. Nested error.isError field (legacy/backward compatibility)
type ErrorDetector struct{}

// NewErrorDetector creates a new error detector
func NewErrorDetector() *ErrorDetector {
	return &ErrorDetector{}
}

// IsMCPError checks if a tool result indicates an error
// This checks both the MCP-compliant format and legacy formats
func (d *ErrorDetector) IsMCPError(result *tools.ToolsCallResult) bool {
	if result == nil {
		return false
	}

	// Check 1: Top-level IsError field (MCP spec compliant)
	if result.IsError {
		logging.Debug("Detected error via top-level IsError field")
		return true
	}

	// Check 2: Legacy nested error object (backward compatibility)
	// Some MCP servers return: { "error": { "isError": true, "message": "..." } }
	if result.Content != nil {
		if contentMap, ok := result.Content.(map[string]interface{}); ok {
			if errorObj, ok := contentMap["error"].(map[string]interface{}); ok {
				// Check for nested isError boolean
				if isError, ok := errorObj["isError"].(bool); ok && isError {
					logging.Debug("Detected error via nested error.isError field")
					return true
				}
				// Check for nested isError string "true" (some implementations)
				if isError, ok := errorObj["isError"].(string); ok && isError == "true" {
					logging.Debug("Detected error via nested error.isError string field")
					return true
				}
			}
		}
	}

	return false
}

// IsMCPErrorFromMap checks if a raw result map indicates an error
// This is useful when working with raw map[string]interface{} results
func (d *ErrorDetector) IsMCPErrorFromMap(result map[string]interface{}) bool {
	if result == nil {
		return false
	}

	// Check 1: Top-level isError field
	if isError, ok := result["isError"].(bool); ok && isError {
		logging.Debug("Detected error via top-level isError field (map)")
		return true
	}
	if isError, ok := result["isError"].(string); ok && isError == "true" {
		logging.Debug("Detected error via top-level isError string field (map)")
		return true
	}

	// Check 2: Nested error.isError field
	if errorObj, ok := result["error"].(map[string]interface{}); ok {
		if isError, ok := errorObj["isError"].(bool); ok && isError {
			logging.Debug("Detected error via nested error.isError field (map)")
			return true
		}
		if isError, ok := errorObj["isError"].(string); ok && isError == "true" {
			logging.Debug("Detected error via nested error.isError string field (map)")
			return true
		}
	}

	// Check 3: Legacy content.error.isError (some servers nest deeper)
	if content, ok := result["content"].(map[string]interface{}); ok {
		if errorObj, ok := content["error"].(map[string]interface{}); ok {
			if isError, ok := errorObj["isError"].(bool); ok && isError {
				logging.Debug("Detected error via nested content.error.isError field (map)")
				return true
			}
		}
	}

	return false
}

// GetErrorMessage extracts the error message from a tool result
// Returns the error message and whether an error was found
func (d *ErrorDetector) GetErrorMessage(result *tools.ToolsCallResult) (string, bool) {
	if result == nil {
		return "", false
	}

	// Check top-level error message
	if result.IsError && result.Error != "" {
		return result.Error, true
	}

	// Check nested error message
	if result.Content != nil {
		if contentMap, ok := result.Content.(map[string]interface{}); ok {
			if errorObj, ok := contentMap["error"].(map[string]interface{}); ok {
				// Try to get message field
				if msg, ok := errorObj["message"].(string); ok && msg != "" {
					return msg, true
				}
				// Try to get error field
				if msg, ok := errorObj["error"].(string); ok && msg != "" {
					return msg, true
				}
				// Try to convert entire error object to string
				if isError, _ := errorObj["isError"].(bool); isError {
					return "MCP tool error (see nested error object)", true
				}
			}
		}
	}

	return "", false
}

// GetErrorMessageFromMap extracts error message from raw map
func (d *ErrorDetector) GetErrorMessageFromMap(result map[string]interface{}) (string, bool) {
	if result == nil {
		return "", false
	}

	// Check top-level error field
	if errorStr, ok := result["error"].(string); ok && errorStr != "" {
		return errorStr, true
	}

	// Check nested error.message
	if errorObj, ok := result["error"].(map[string]interface{}); ok {
		if msg, ok := errorObj["message"].(string); ok && msg != "" {
			return msg, true
		}
		if msg, ok := errorObj["error"].(string); ok && msg != "" {
			return msg, true
		}
	}

	// Check content.error.message
	if content, ok := result["content"].(map[string]interface{}); ok {
		if errorObj, ok := content["error"].(map[string]interface{}); ok {
			if msg, ok := errorObj["message"].(string); ok && msg != "" {
				return msg, true
			}
		}
	}

	return "", false
}


// ExtractTextFromContent extracts text content from various MCP response formats
// Handles both array format [{"text": "...", "type": "text"}] and string format
func (d *ErrorDetector) ExtractTextFromContent(content interface{}) string {
	if content == nil {
		return ""
	}
	
	// Handle string content
	if str, ok := content.(string); ok {
		return str
	}
	
	// Handle array of content items (MCP standard format)
	if contentArray, ok := content.([]interface{}); ok {
		for _, item := range contentArray {
			if itemMap, ok := item.(map[string]interface{}); ok {
				// Extract text field if it exists
				if text, ok := itemMap["text"].(string); ok && text != "" {
					return text
				}
			}
		}
	}
	
	// Handle map format
	if contentMap, ok := content.(map[string]interface{}); ok {
		if text, ok := contentMap["text"].(string); ok && text != "" {
			return text
		}
	}
	
	return ""
}

// LogErrorDetails logs detailed error information for debugging
func (d *ErrorDetector) LogErrorDetails(result *tools.ToolsCallResult) {
	if result == nil {
		return
	}

	logging.Debug("=== MCP Tool Result Error Analysis ===")
	logging.Debug("Top-level IsError: %v", result.IsError)
	logging.Debug("Top-level Error message: %s", result.Error)
	
	if result.Content != nil {
		logging.Debug("Content type: %T", result.Content)
		if contentMap, ok := result.Content.(map[string]interface{}); ok {
			logging.Debug("Content is map with %d keys", len(contentMap))
			if errorObj, ok := contentMap["error"]; ok {
				logging.Debug("Content contains 'error' key: %+v", errorObj)
			}
		}
	}
	
	isError := d.IsMCPError(result)
	errorMsg, hasMsg := d.GetErrorMessage(result)
	
	logging.Debug("Error detected: %v", isError)
	if hasMsg {
		logging.Debug("Error message: %s", errorMsg)
	}
	logging.Debug("=== End Error Analysis ===")
}
