# Template Quick Reference

## Template Schema

### Minimal Template

```json
{
  "templates": {
    "template_name": {
      "name": "template_name",
      "description": "What this template does",
      "steps": [
        {
          "step": 1,
          "name": "Step Name",
          "system_prompt": "You are a helpful assistant...",
          "base_prompt": "Process this: {{stdin}}",
          "servers": ["server-name"],
          "output_variable": "result"
        }
      ]
    }
  }
}
```

## Property Reference

### Template Properties

| Property      | Required | Type   | Description            |
| ------------- | -------- | ------ | ---------------------- |
| `name`        | ✅        | string | Template identifier    |
| `description` | ✅        | string | What the template does |
| `steps`       | ✅        | array  | Workflow steps         |
| `variables`   | ❌        | object | Global variables       |

### Step Properties

| Property          | Required | Type    | Description                                            |
| ----------------- | -------- | ------- | ------------------------------------------------------ |
| `step`            | ✅        | integer | Sequential step number                                 |
| `name`            | ✅        | string  | Descriptive step name                                  |
| `system_prompt`   | ✅        | string  | AI role definition                                     |
| `base_prompt`     | ✅        | string  | Main prompt with variables                             |
| `servers`         | ❌        | array   | MCP servers to use.  Defaults to all if not specified. |
| `input_variables` | ❌        | array   | Data dependencies from previous steps or stdin         |
| `output_variable` | ❌        | string  | Variable to store output                               |
| `temperature`     | ❌        | float   | AI creativity (0.0-1.0)                                |

## Built-in Variables

| Variable          | Source         | Usage                                             |
| ----------------- | -------------- | ------------------------------------------------- |
| `{{stdin}}`       | Command input  | `"base_prompt": "Analyze: {{stdin}}"`             |
| `{{input_data}}`  | Command input  | Alias for `{{stdin}}`                             |
| `{{step_result}}` | Previous steps | `"base_prompt": "Continue from: {{step_result}}"` |

## Command Usage

```bash
# Execute template
mcp-cli --template "template_name"

# With input data
echo "data" | mcp-cli --template "template_name"
mcp-cli --template "template_name" --input-data "data"

# List templates
mcp-cli --list-templates

# Verbose output
mcp-cli --template "template_name" --verbose

# Override servers
mcp-cli --template "template_name" --server "specific-server"
```

## Examples

### Single Step

```json
{
  "templates": {
    "simple": {
      "name": "simple",
      "description": "Simple data analysis",
      "steps": [
        {
          "step": 1,
          "name": "Analyze",
          "system_prompt": "You are a data analyst.",
          "base_prompt": "Analyze: {{stdin}}",
          "servers": []
        }
      ]
    }
  }
}
```

### Multi-Step Chain

```json
{
  "templates": {
    "chain": {
      "name": "chain",
      "description": "Multi-step analysis",
      "steps": [
        {
          "step": 1,
          "name": "Parse Data",
          "system_prompt": "You are a data parser.",
          "base_prompt": "Parse: {{stdin}}",
          "servers": [],
          "output_variable": "parsed_data"
        },
        {
          "step": 2,
          "name": "Analyze Data", 
          "system_prompt": "You are an analyst.",
          "base_prompt": "Analyze: {{parsed_data}}",
          "servers": ["search"],
          "input_variables": ["parsed_data"]
        }
      ]
    }
  }
}
```

### Function App Integration

```json
{
  "templates": {
    "webhook": {
      "name": "webhook",
      "description": "Process webhook data",
      "steps": [
        {
          "step": 1,
          "name": "Process Webhook",
          "system_prompt": "You process webhook data from Function Apps.",
          "base_prompt": "Process this webhook: {{stdin}}\n\nExtract key fields and validate.",
          "servers": [],
          "output_variable": "processed_data"
        }
      ]
    }
  }
}
```

## Best Practices

### ✅ Do

- Use descriptive names for steps and variables
- Keep prompts clear and specific
- Specify only needed servers
- Use `{{stdin}}` for Function App integration
- Rely on provider defaults for timeout/retry

### ❌ Don't

- Override timeout/retry unless necessary
- Use complex `input_handling` configurations
- Create overly long single steps
- Use generic variable names like `result1`
- Specify servers you don't need

## Common Patterns

### Data Processing

```json
"base_prompt": "Process this data: {{stdin}}\n\nExtract key information and summarize."
```

### Chain Analysis

```json
"base_prompt": "Based on: {{previous_analysis}}\n\nPerform next level analysis."
```

### Web Research

```json
"base_prompt": "Research this topic: {{topic}}\n\nFind recent information and trends."
```

### Function App Response

```json
"base_prompt": "Process this Function App payload: {{stdin}}\n\nReturn structured response."
```
