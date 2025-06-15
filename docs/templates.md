# MCP-CLI Templates Documentation

## Overview

MCP-CLI Templates provide a powerful workflow system that allows you to chain multiple AI requests across different providers and MCP servers. Templates enable you to create reusable automation workflows for complex tasks like security incident analysis, data processing, and multi-step AI reasoning.

## Table of Contents

- [Quick Start](#quick-start)
- [Template Schema Reference](#template-schema-reference)
- [Variable System](#variable-system)
- [Server Configuration](#server-configuration)
- [Examples](#examples)
- [Function App Integration](#function-app-integration)
- [Best Practices](#best-practices)

## Quick Start

### Basic Template Execution

```bash
# Execute a template
mcp-cli --template "incident_review"

# Execute with input data
echo '{"incident_id": "INC-001"}' | mcp-cli --template "process_data"

# List available templates
mcp-cli --list-templates
```

### Simple Template Example

```json
{
  "templates": {
    "simple_analysis": {
      "name": "simple_analysis",
      "description": "Analyze input data and provide insights",
      "steps": [
        {
          "step": 1,
          "name": "Data Analysis",
          "system_prompt": "You are a data analyst. Provide clear insights.",
          "base_prompt": "Analyze this data: {{stdin}}\n\nProvide key findings and recommendations.",
          "servers": [],
          "output_variable": "analysis_result"
        }
      ]
    }
  }
}
```

## Template Schema Reference

### Root Template Object

| Property      | Type   | Required | Description                                          |
| ------------- | ------ | -------- | ---------------------------------------------------- |
| `name`        | string | ✅        | Unique identifier for the template                   |
| `description` | string | ✅        | Human-readable description of what the template does |
| `steps`       | array  | ✅        | Array of workflow steps to execute sequentially      |
| `variables`   | object | ❌        | Global template variables (key-value pairs)          |

### Step Object

| Property          | Type    | Required | Description                                                                       |
| ----------------- | ------- | -------- | --------------------------------------------------------------------------------- |
| `step`            | integer | ✅        | Step number (must be sequential: 1, 2, 3...)                                      |
| `name`            | string  | ✅        | Descriptive name for the step                                                     |
| `system_prompt`   | string  | ✅        | Instructions that define the AI's role and behavior                               |
| `base_prompt`     | string  | ✅        | The main prompt with variable substitution support                                |
| `servers`         | array   | ❌        | List of MCP server names to use for this step.  Defaults to all if not specified. |
| `input_variables` | array   | ❌        | Variables this step depends on from previous steps                                |
| `output_variable` | string  | ❌        | Variable name to store this step's output                                         |
| `temperature`     | float   | ❌        | AI creativity level (0.0-1.0, default: provider setting)                          |

## Variable System

Templates support a powerful variable substitution system using double curly braces: `{{variable_name}}`

### Built-in Variables

| Variable         | Source        | Description                                      |
| ---------------- | ------------- | ------------------------------------------------ |
| `{{stdin}}`      | Command input | Data piped to the command or from `--input-data` |
| `{{input_data}}` | Command input | Alias for `{{stdin}}`                            |

### Step Variables

Variables are created by steps and can be referenced in subsequent steps:

```json
{
  "step": 1,
  "output_variable": "incident_details"
},
{
  "step": 2,
  "base_prompt": "Based on: {{incident_details}}\n\nPerform threat analysis",
  "input_variables": ["incident_details"]
}
```

### Template Variables

Global variables defined at the template level:

```json
{
  "templates": {
    "my_template": {
      "variables": {
        "company_name": "Acme Corp",
        "default_priority": "medium"
      },
      "steps": [
        {
          "base_prompt": "Analyze this incident for {{company_name}} with {{default_priority}} priority"
        }
      ]
    }
  }
}
```

### Variable Precedence

1. **Step outputs** (highest priority)
2. **Command input** (`{{stdin}}`, `{{input_data}}`)
3. **Template variables** (lowest priority)

## Server Configuration

### Step-Level Server Specification

Each step can specify which MCP servers to use:

```json
{
  "step": 1,
  "servers": ["GraphSecurityIncidents"],
  "base_prompt": "Get security incident details"
},
{
  "step": 2, 
  "servers": ["brave-search"],
  "base_prompt": "Research threat intelligence for: {{incident_details}}"
}
```

### Server Selection Logic

1. **Command-line override**: `--server` parameter overrides template settings
2. **Template specification**: Use servers listed in step's `servers` array
3. **No servers specified**: Use all available configured servers

### Empty Servers Array

```json
{
  "servers": []  // Uses no MCP servers, AI reasoning only
}
```

## Examples

### Security Incident Analysis

```json
{
  "templates": {
    "incident_review": {
      "name": "incident_review",
      "description": "Comprehensive security incident analysis",
      "steps": [
        {
          "step": 1,
          "name": "Get Incident Details",
          "system_prompt": "You are a security analyst. Retrieve and analyze security incidents with precision.",
          "base_prompt": "Get the full details of the latest security incident",
          "servers": ["GraphSecurityIncidents"],
          "output_variable": "incident_data"
        },
        {
          "step": 2,
          "name": "Threat Intelligence Research",
          "system_prompt": "You are a threat intelligence analyst. Research potential threats and attack patterns.",
          "base_prompt": "Based on this incident: {{incident_data}}\n\nResearch similar threats, IOCs, and attack patterns using web search.",
          "servers": ["brave-search"],
          "input_variables": ["incident_data"],
          "output_variable": "threat_intel",
          "temperature": 0.1
        },
        {
          "step": 3,
          "name": "Risk Assessment",
          "system_prompt": "You are a security operations manager. Provide executive-level risk assessments.",
          "base_prompt": "Create a comprehensive risk assessment based on:\n\nIncident: {{incident_data}}\n\nThreat Intelligence: {{threat_intel}}\n\nProvide risk level, business impact, and recommended actions.",
          "servers": [],
          "input_variables": ["incident_data", "threat_intel"],
          "output_variable": "risk_assessment"
        }
      ]
    }
  }
}
```

### Function App Data Processing

```json
{
  "templates": {
    "process_webhook_data": {
      "name": "process_webhook_data", 
      "description": "Process JSON data from Azure Function App webhooks",
      "steps": [
        {
          "step": 1,
          "name": "Parse and Validate",
          "system_prompt": "You are a data processing assistant. Parse and validate incoming webhook data.",
          "base_prompt": "Process this webhook data: {{stdin}}\n\nValidate the structure, extract key fields, and flag any anomalies.",
          "servers": [],
          "output_variable": "parsed_data"
        },
        {
          "step": 2,
          "name": "Enrich with External Data",
          "system_prompt": "You are a data enrichment specialist. Enhance data with additional context.",
          "base_prompt": "Enrich this data: {{parsed_data}}\n\nSearch for additional context, related information, or background data.",
          "servers": ["brave-search"],
          "input_variables": ["parsed_data"],
          "output_variable": "enriched_data"
        }
      ]
    }
  }
}
```

### Simple Query Processing

```json
{
  "templates": {
    "simple_query": {
      "name": "simple_query",
      "description": "Process a simple question with context",
      "steps": [
        {
          "step": 1,
          "name": "Answer Question",
          "system_prompt": "You are a helpful assistant. Answer questions clearly and concisely.",
          "base_prompt": "{{stdin}}",
          "servers": []
        }
      ]
    }
  }
}
```

## Function App Integration

Templates are perfect for Azure Function Apps and serverless workflows:

### Function App Example

```javascript
// Azure Function App
const { exec } = require('child_process');

module.exports = async function (context, req) {
    const inputData = JSON.stringify(req.body);

    return new Promise((resolve, reject) => {
        const mcpProcess = exec('mcp-cli --template "process_webhook_data"', 
            { cwd: '/path/to/mcp-cli' });

        // Send data via stdin
        mcpProcess.stdin.write(inputData);
        mcpProcess.stdin.end();

        let output = '';
        mcpProcess.stdout.on('data', (data) => {
            output += data.toString();
        });

        mcpProcess.on('close', (code) => {
            if (code === 0) {
                context.res = {
                    status: 200,
                    body: { result: output.trim() }
                };
            } else {
                context.res = {
                    status: 500,
                    body: { error: 'Processing failed' }
                };
            }
            resolve();
        });
    });
};
```

### Command Line Usage

```bash
# Direct execution with JSON
echo '{"user_id": 123, "action": "login"}' | mcp-cli --template "process_webhook_data"

# From file
cat webhook_payload.json | mcp-cli --template "process_webhook_data"

# With input data parameter
mcp-cli --template "simple_query" --input-data "What is the weather like today?"
```

## Best Practices

### Template Design

1. **Keep Steps Focused**: Each step should have a single, clear purpose
2. **Use Descriptive Names**: Step names should explain what they do
3. **Logical Server Assignment**: Match servers to step requirements
4. **Clear Variable Names**: Use descriptive variable names like `incident_details`, not `step1_result`

### Variable Management

1. **Document Dependencies**: Use `input_variables` to declare dependencies
2. **Consistent Naming**: Use snake_case for variable names
3. **Meaningful Names**: `threat_assessment` is better than `step2_output`

### Server Selection

1. **Minimize Server Usage**: Only specify servers you actually need
2. **Step-Specific Servers**: Different steps can use different servers
3. **Empty Arrays**: Use `"servers": []` for pure AI reasoning steps

### Prompt Engineering

1. **Clear System Prompts**: Define the AI's role explicitly
2. **Context in Base Prompts**: Provide clear context and instructions
3. **Variable Integration**: Use variables naturally in prompts

### Configuration Management

1. **Use Provider Defaults**: Don't override timeout/retry unless necessary
2. **Template Variables**: Use global variables for common values
3. **Simple Configuration**: Avoid over-engineering with complex settings

## Troubleshooting

### Common Issues

1. **Empty Output**: Check that `output_variable` is set and referenced correctly
2. **Variable Not Found**: Ensure variable is created before being referenced
3. **Server Not Found**: Verify server names match your configuration
4. **Template Not Found**: Check template name spelling and configuration file path

### Debug Mode

Use verbose logging to debug template execution:

```bash
mcp-cli --template "my_template" --verbose
```

### Validation

Templates are validated on execution. Common validation errors:

- Missing required fields (`name`, `step`, `base_prompt`)
- Invalid step sequence (must be 1, 2, 3...)
- Undefined variable references
- Non-existent server names

# 
