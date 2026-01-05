package mcp

import (
	"encoding/json"
	"fmt"

	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/logging"
)

// LenientSchemaValidator provides tolerant validation for MCP tool schemas
// Some MCP servers return complex schemas with $defs, $ref, and nested structures
// that can cause strict validation to fail. This validator falls back to accepting
// schemas that fail validation, logging a warning but allowing tool discovery to continue.
//
// This matches the approach used by Gemini CLI to handle real-world MCP servers
// that may not strictly follow JSON Schema specifications.
type LenientSchemaValidator struct {
	// Could add strict validator here if needed in the future
}

// NewLenientSchemaValidator creates a new lenient schema validator
func NewLenientSchemaValidator() *LenientSchemaValidator {
	return &LenientSchemaValidator{}
}

// ValidateSchema validates a JSON schema with lenient fallback behavior
//
// Returns:
//   - nil if schema is valid or if validation fails but we accept it anyway
//   - error only for catastrophic failures (not schema validation issues)
func (v *LenientSchemaValidator) ValidateSchema(schema map[string]interface{}) error {
	if schema == nil {
		logging.Warn("MCP tool schema is nil, accepting anyway for lenient discovery")
		return nil
	}

	// Try to validate basic structure
	if err := v.validateBasicStructure(schema); err != nil {
		// Log the validation failure but don't reject the tool
		logging.Warn("MCP tool schema validation failed, accepting anyway: %v", err)
		logging.Debug("Problematic schema: %+v", schema)
		return nil // Accept despite validation failure
	}

	return nil
}

// validateBasicStructure performs basic structural validation
// This is a lightweight check that catches only egregious schema issues
func (v *LenientSchemaValidator) validateBasicStructure(schema map[string]interface{}) error {
	// Check if it can be marshaled to JSON (basic sanity check)
	if _, err := json.Marshal(schema); err != nil {
		return fmt.Errorf("schema cannot be marshaled to JSON: %w", err)
	}

	// Check for basic JSON Schema type field (if present)
	if typeField, ok := schema["type"]; ok {
		// Type field should be a string or array
		switch typeField.(type) {
		case string:
			// Valid
		case []interface{}:
			// Valid (array of types)
		default:
			return fmt.Errorf("schema type field has invalid type: %T", typeField)
		}
	}

	// If properties exist, they should be an object
	if properties, ok := schema["properties"]; ok {
		if _, ok := properties.(map[string]interface{}); !ok {
			return fmt.Errorf("schema properties field is not an object")
		}
	}

	// Schema looks basically valid
	return nil
}

// ValidateSchemaStrict performs strict validation (for future use)
// Currently unused but kept for potential future enhancement
func (v *LenientSchemaValidator) ValidateSchemaStrict(schema map[string]interface{}) error {
	// TODO: Implement strict validation using a JSON Schema validator library
	// For now, just do basic validation
	return v.validateBasicStructure(schema)
}

// ShouldAcceptSchema determines if a schema should be accepted
// This always returns true in lenient mode
func (v *LenientSchemaValidator) ShouldAcceptSchema(schema map[string]interface{}) bool {
	// In lenient mode, we accept all schemas that can be marshaled to JSON
	if schema == nil {
		return false // Reject nil schemas
	}

	// Try to marshal - if it works, accept it
	if _, err := json.Marshal(schema); err != nil {
		logging.Error("Schema cannot be marshaled to JSON, rejecting: %v", err)
		return false
	}

	return true
}

// LogSchemaForDebugging logs schema information for debugging purposes
func (v *LenientSchemaValidator) LogSchemaForDebugging(toolName string, schema map[string]interface{}) {
	logging.Debug("Tool schema for %s:", toolName)
	
	if schemaJSON, err := json.MarshalIndent(schema, "  ", "  "); err == nil {
		logging.Debug("  %s", string(schemaJSON))
	} else {
		logging.Debug("  (failed to marshal schema: %v)", err)
	}

	// Log specific potentially problematic fields
	if defs, ok := schema["$defs"]; ok {
		logging.Debug("  Schema contains $defs: %+v", defs)
	}
	if ref, ok := schema["$ref"]; ok {
		logging.Debug("  Schema contains $ref: %s", ref)
	}
	if definitions, ok := schema["definitions"]; ok {
		logging.Debug("  Schema contains definitions: %+v", definitions)
	}
	if oneOf, ok := schema["oneOf"]; ok {
		logging.Debug("  Schema contains oneOf with %d options", len(oneOf.([]interface{})))
	}
	if anyOf, ok := schema["anyOf"]; ok {
		logging.Debug("  Schema contains anyOf with %d options", len(anyOf.([]interface{})))
	}
}
