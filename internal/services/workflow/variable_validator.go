package workflow

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/config"
)

// Variable interpolation pattern
var variablePattern = regexp.MustCompile(`\{\{([^}]+)\}\}`)

// validateVariableSyntax checks for invalid variable interpolations in a field
// This catches common mistakes like {{input.text}} in contexts where only {{input}} is supported
func (v *WorkflowValidator) validateVariableSyntax(step *config.StepV2, field string, value string) {
	if value == "" {
		return
	}

	matches := variablePattern.FindAllStringSubmatch(value, -1)
	
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		
		varExpr := strings.TrimSpace(match[1])
		
		// Check for nested field access (contains dots)
		if !strings.Contains(varExpr, ".") {
			continue // Base variables are always OK
		}
		
		parts := strings.Split(varExpr, ".")
		base := parts[0]
		rest := strings.Join(parts[1:], ".")
		
		// Determine if nested access is allowed based on field and base variable
		switch base {
		case "input":
			// input.field is NEVER supported - always use {{input}} and parse in run
			v.addError(step.Name, field,
				fmt.Sprintf("{{input.%s}} not supported", rest),
				"Use {{input}} to receive the entire object, then parse it in your 'run' prompt. Example:\n"+
					"  run: |\n"+
					"    Input data: {{input}}\n"+
					"    Extract the text field and process it...")
		
		case "step":
			// step.name is OK, but step.name.field is not (except in 'run')
			if strings.Contains(rest, ".") && field != "run" {
				v.addError(step.Name, field,
					fmt.Sprintf("{{%s}} uses deep nesting which is not supported in %s fields", varExpr, field),
					fmt.Sprintf("Use {{step.%s}} to get the step's output, then parse it in a 'run' field", parts[1]))
			}
		
		case "env", "loop":
			// env.VAR and loop.iteration are OK in any context
			continue
		
		default:
			// Unknown base with dot notation
			if field != "run" {
				v.addWarning(step.Name, field,
					fmt.Sprintf("{{%s}} uses dot notation which may not be supported in %s fields", varExpr, field),
					"Consider using just the base variable and parsing in a 'run' field if this doesn't work")
			}
		}
	}
}

// addWarning adds a validation warning (non-fatal)
func (v *WorkflowValidator) addWarning(step, field, message, hint string) {
	// For now, treat as errors. Could later separate warnings from errors
	v.addError(step, field, message, hint)
}

// validateRagVariables specifically validates RAG query variables
func (v *WorkflowValidator) validateRagVariables(step *config.StepV2) {
	if step.Rag == nil || step.Rag.Query == "" {
		return
	}
	
	// Check for common mistakes in RAG queries
	query := step.Rag.Query
	
	// Check for {{input.field}} pattern
	if strings.Contains(query, "{{input.") {
		v.addError(step.Name, "rag.query",
			"RAG queries cannot use {{input.field}} syntax",
			"RAG query fields don't support variable interpolation with nested access.\n"+
				"Solution: Use a previous step to prepare the query string:\n"+
				"  - name: prepare_query\n"+
				"    run: |\n"+
				"      Input: {{input}}\n"+
				"      Extract text: {{input.text}}\n"+
				"      Output just the text value\n"+
				"  - name: search\n"+
				"    rag:\n"+
				"      query: \"{{prepare_query}}\"\n"+
				"\nOr embed RAG in a 'run' field where variables work:\n"+
				"  - name: search\n"+
				"    run: |\n"+
				"      Text to search: {{input.text}}\n"+
				"      Use RAG to search for relevant documents")
	}
}

// validateLoopVariables specifically validates loop configuration variables
func (v *WorkflowValidator) validateLoopVariables(step *config.StepV2) {
	if step.Loop == nil {
		return
	}
	
	// Check loop items field
	if step.Loop.Items != "" && strings.Contains(step.Loop.Items, "{{input.") {
		v.addError(step.Name, "loop.items",
			"Loop items cannot use {{input.field}} syntax",
			"In iterate mode, each item is passed to the child workflow as {{input}}.\n"+
				"The items source should be:\n"+
				"  • file:///path/to/items.json\n"+
				"  • {{env.items_var}}\n"+
				"  • {{step.previous_step}}\n"+
				"  • {{statements}}\n"+
				"\nThe child workflow receives each item as {{input}}, not {{input.field}}")
	}
	
	// Check loop mode and provide helpful messages
	if step.Loop.Mode == "iterate" {
		// Document expected behavior
		// (This is just informational, not an error)
	}
}
