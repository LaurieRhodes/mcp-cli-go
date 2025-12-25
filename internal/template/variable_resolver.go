package template

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// VariableResolver handles variable storage and resolution for templates
type VariableResolver struct {
	variables map[string]interface{}
}

// NewVariableResolver creates a new variable resolver
func NewVariableResolver() *VariableResolver {
	return &VariableResolver{
		variables: make(map[string]interface{}),
	}
}

// SetVariable stores a variable with the given name and value
func (vr *VariableResolver) SetVariable(name string, value interface{}) {
	vr.variables[name] = value
}

// GetVariable retrieves a variable by name
func (vr *VariableResolver) GetVariable(name string) (interface{}, bool) {
	val, ok := vr.variables[name]
	return val, ok
}

// DeleteVariable removes a variable
func (vr *VariableResolver) DeleteVariable(name string) {
	delete(vr.variables, name)
}

// GetAllVariables returns all variables
func (vr *VariableResolver) GetAllVariables() map[string]interface{} {
	return vr.variables
}

// SetMultiple sets multiple variables at once
func (vr *VariableResolver) SetMultiple(vars map[string]interface{}) {
	for k, v := range vars {
		vr.variables[k] = v
	}
}

// ResolveString replaces {{variable}} patterns with actual values
func (vr *VariableResolver) ResolveString(template string) (string, error) {
	// Pattern matches {{variable}} or {{variable.field.subfield}}
	re := regexp.MustCompile(`\{\{([^}]+)\}\}`)

	result := template
	var lastErr error

	result = re.ReplaceAllStringFunc(result, func(match string) string {
		// Extract variable expression (remove {{ and }})
		expr := strings.TrimSpace(match[2 : len(match)-2])

		// Resolve the variable path
		value, err := vr.resolveExpression(expr)
		if err != nil {
			lastErr = err
			return match // Keep original if resolution fails
		}

		// Convert value to string
		return vr.valueToString(value)
	})

	return result, lastErr
}

// resolveExpression resolves a variable expression like "incident.severity" or "step1_output"
func (vr *VariableResolver) resolveExpression(expr string) (interface{}, error) {
	// Handle array index access: variable[0]
	if strings.Contains(expr, "[") {
		return vr.resolveArrayAccess(expr)
	}

	// Handle nested field access: variable.field.subfield
	if strings.Contains(expr, ".") {
		return vr.resolveNestedPath(expr)
	}

	// Handle filter operations: variable | length
	if strings.Contains(expr, "|") {
		return vr.resolveFilter(expr)
	}

	// Simple variable lookup
	value, ok := vr.variables[expr]
	if !ok {
		return nil, fmt.Errorf("variable not found: %s", expr)
	}

	return value, nil
}

// resolveNestedPath handles nested field access like "incident.severity"
func (vr *VariableResolver) resolveNestedPath(path string) (interface{}, error) {
	parts := strings.Split(path, ".")

	// Get root variable
	current, ok := vr.variables[parts[0]]
	if !ok {
		return nil, fmt.Errorf("variable not found: %s", parts[0])
	}

	// Navigate nested fields
	for i := 1; i < len(parts); i++ {
		current = vr.navigateToField(current, parts[i])
		if current == nil {
			return nil, fmt.Errorf("field not found: %s in path %s", parts[i], path)
		}
	}

	return current, nil
}

// navigateToField navigates to a field in a map or struct
func (vr *VariableResolver) navigateToField(current interface{}, field string) interface{} {
	// Handle map[string]interface{}
	if m, ok := current.(map[string]interface{}); ok {
		return m[field]
	}

	// Handle JSON unmarshaled data
	if m, ok := current.(map[interface{}]interface{}); ok {
		return m[field]
	}

	// If current is a string that looks like JSON, try parsing it
	if str, ok := current.(string); ok {
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(str), &data); err == nil {
			return data[field]
		}
	}

	return nil
}

// resolveArrayAccess handles array access like "items[0]" or "results[2].field"
func (vr *VariableResolver) resolveArrayAccess(expr string) (interface{}, error) {
	// Find the array index
	idxStart := strings.Index(expr, "[")
	idxEnd := strings.Index(expr, "]")

	if idxStart == -1 || idxEnd == -1 || idxEnd < idxStart {
		return nil, fmt.Errorf("invalid array access syntax: %s", expr)
	}

	// Extract variable name and index
	varName := expr[:idxStart]
	indexStr := expr[idxStart+1 : idxEnd]
	remaining := ""
	if idxEnd+1 < len(expr) {
		remaining = expr[idxEnd+1:]
		// Remove leading dot if present
		if strings.HasPrefix(remaining, ".") {
			remaining = remaining[1:]
		}
	}

	// Get the array variable
	value, ok := vr.variables[varName]
	if !ok {
		return nil, fmt.Errorf("variable not found: %s", varName)
	}

	// Convert to array
	arr, ok := value.([]interface{})
	if !ok {
		return nil, fmt.Errorf("variable %s is not an array", varName)
	}

	// Parse index
	idx, err := strconv.Atoi(strings.TrimSpace(indexStr))
	if err != nil {
		return nil, fmt.Errorf("invalid array index: %s", indexStr)
	}

	// Check bounds
	if idx < 0 || idx >= len(arr) {
		return nil, fmt.Errorf("array index out of bounds: %d (len=%d)", idx, len(arr))
	}

	// Get array element
	element := arr[idx]

	// If there's remaining path, navigate to it
	if remaining != "" {
		if strings.Contains(remaining, ".") {
			parts := strings.Split(remaining, ".")
			for _, part := range parts {
				element = vr.navigateToField(element, part)
				if element == nil {
					return nil, fmt.Errorf("field not found: %s", part)
				}
			}
		} else {
			element = vr.navigateToField(element, remaining)
			if element == nil {
				return nil, fmt.Errorf("field not found: %s", remaining)
			}
		}
	}

	return element, nil
}

// resolveFilter handles filter operations like "items | length" or "items | filter(...)"
func (vr *VariableResolver) resolveFilter(expr string) (interface{}, error) {
	parts := strings.SplitN(expr, "|", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid filter syntax: %s", expr)
	}

	// Resolve the variable part
	varExpr := strings.TrimSpace(parts[0])
	filterExpr := strings.TrimSpace(parts[1])

	value, err := vr.resolveExpression(varExpr)
	if err != nil {
		return nil, err
	}

	// Apply filter
	return vr.applyFilter(value, filterExpr)
}

// applyFilter applies a filter function to a value
func (vr *VariableResolver) applyFilter(value interface{}, filter string) (interface{}, error) {
	switch filter {
	case "length":
		return vr.filterLength(value)
	
	default:
		// Check for filter with arguments: filter(arg1, arg2)
		if strings.Contains(filter, "(") && strings.HasSuffix(filter, ")") {
			filterName := filter[:strings.Index(filter, "(")]
			argsStr := filter[strings.Index(filter, "(")+1 : len(filter)-1]
			args := vr.parseFilterArgs(argsStr)
			
			switch filterName {
			case "filter":
				return vr.filterArray(value, args)
			default:
				return nil, fmt.Errorf("unknown filter: %s", filterName)
			}
		}
		
		return nil, fmt.Errorf("unknown filter: %s", filter)
	}
}

// filterLength returns the length of an array or string
func (vr *VariableResolver) filterLength(value interface{}) (interface{}, error) {
	switch v := value.(type) {
	case []interface{}:
		return len(v), nil
	case string:
		return len(v), nil
	case map[string]interface{}:
		return len(v), nil
	default:
		return nil, fmt.Errorf("cannot get length of type %T", value)
	}
}

// filterArray filters an array based on a condition
func (vr *VariableResolver) filterArray(value interface{}, args []string) (interface{}, error) {
	arr, ok := value.([]interface{})
	if !ok {
		return nil, fmt.Errorf("filter can only be applied to arrays")
	}

	if len(args) == 0 {
		return arr, nil
	}

	_ = args[0] // condition - unused in Phase 1
	
	// Simple filter implementation - can be extended
	filtered := make([]interface{}, 0)
	for _, item := range arr {
		// For now, just return the array as-is
		// Full implementation would evaluate condition against each item
		filtered = append(filtered, item)
	}

	return filtered, nil
}

// parseFilterArgs parses filter arguments from string like "arg1, arg2"
func (vr *VariableResolver) parseFilterArgs(argsStr string) []string {
	if argsStr == "" {
		return []string{}
	}

	parts := strings.Split(argsStr, ",")
	args := make([]string, len(parts))
	for i, part := range parts {
		args[i] = strings.TrimSpace(part)
	}

	return args
}

// valueToString converts a value to its string representation
func (vr *VariableResolver) valueToString(value interface{}) string {
	if value == nil {
		return ""
	}

	switch v := value.(type) {
	case string:
		return v
	case int, int32, int64, uint, uint32, uint64:
		return fmt.Sprintf("%d", v)
	case float32, float64:
		return fmt.Sprintf("%v", v)
	case bool:
		return fmt.Sprintf("%t", v)
	case []interface{}, map[string]interface{}:
		// For complex types, marshal to JSON
		if jsonBytes, err := json.Marshal(v); err == nil {
			return string(jsonBytes)
		}
		return fmt.Sprintf("%v", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// SetStepOutput stores output from a step using the step name
func (vr *VariableResolver) SetStepOutput(stepName string, output interface{}) {
	vr.SetVariable(stepName, output)
}

// SetOutputs stores multiple outputs from a step
func (vr *VariableResolver) SetOutputs(outputs map[string]interface{}) {
	for name, value := range outputs {
		vr.SetVariable(name, value)
	}
}

// ResolveExpression resolves a variable expression without {{}} wrapper
// This is useful for programmatic variable resolution
func (vr *VariableResolver) ResolveExpression(expr string) (interface{}, error) {
	return vr.resolveExpression(expr)
}

// EvaluateCondition evaluates a condition expression
func (vr *VariableResolver) EvaluateCondition(condition string) (bool, error) {
	// First, resolve any {{variable}} references in the condition
	resolved, err := vr.ResolveString(condition)
	if err != nil {
		return false, err
	}

	// Simple condition evaluation
	// This is a basic implementation - can be enhanced with a proper expression parser
	
	// Handle simple equality
	if strings.Contains(resolved, "==") {
		parts := strings.SplitN(resolved, "==", 2)
		left := strings.TrimSpace(parts[0])
		right := strings.TrimSpace(parts[1])
		
		// Try to resolve left side as a variable if it's not already resolved
		if val, ok := vr.variables[left]; ok {
			left = vr.valueToString(val)
		}
		
		// Remove quotes if present
		left = strings.Trim(left, "'\"")
		right = strings.Trim(right, "'\"")
		return left == right, nil
	}

	// Handle inequality
	if strings.Contains(resolved, "!=") {
		parts := strings.SplitN(resolved, "!=", 2)
		left := strings.TrimSpace(parts[0])
		right := strings.TrimSpace(parts[1])
		
		// Try to resolve left side as a variable if it's not already resolved
		if val, ok := vr.variables[left]; ok {
			left = vr.valueToString(val)
		}
		
		left = strings.Trim(left, "'\"")
		right = strings.Trim(right, "'\"")
		return left != right, nil
	}

	// Handle 'or' operator
	if strings.Contains(resolved, " or ") {
		parts := strings.Split(resolved, " or ")
		for _, part := range parts {
			result, err := vr.EvaluateCondition(strings.TrimSpace(part))
			if err != nil {
				continue
			}
			if result {
				return true, nil
			}
		}
		return false, nil
	}

	// Handle 'and' operator
	if strings.Contains(resolved, " and ") {
		parts := strings.Split(resolved, " and ")
		for _, part := range parts {
			result, err := vr.EvaluateCondition(strings.TrimSpace(part))
			if err != nil {
				return false, err
			}
			if !result {
				return false, nil
			}
		}
		return true, nil
	}

	// Handle boolean values
	resolved = strings.ToLower(strings.TrimSpace(resolved))
	if resolved == "true" {
		return true, nil
	}
	if resolved == "false" {
		return false, nil
	}

	// Handle non-empty string as true
	if resolved != "" && resolved != "0" {
		return true, nil
	}

	return false, nil
}
