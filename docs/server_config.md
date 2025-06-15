# Server Configuration Documentation

This document explains the complete structure and options available in the MCP CLI server configuration file (`server_config.json`), including the new **Template System** for workflow automation.

## Overview

The `server_config.json` file is the central configuration file for MCP-CLI, defining:

- **Templates** - Multi-step AI workflows for automation
- **Servers** - MCP server configurations
- **AI Providers** - LLM provider settings with timeout/retry
- **System Integration** - Function App and serverless workflows

## File Structure

The configuration file has the following top-level sections:

```json
{
  "templates": {
    // Workflow template definitions (NEW)
  },
  "servers": {
    // MCP server configurations
  },
  "ai": {
    // AI provider configurations with enhanced settings
  }
}
```

## Templates Configuration

The `templates` section defines multi-step AI workflows that can chain requests across different providers and MCP servers:

```json
"templates": {
  "incident_review": {
    "name": "incident_review",
    "description": "Comprehensive security incident analysis workflow",
    "steps": [
      {
        "step": 1,
        "name": "Get Incident Details",
        "system_prompt": "You are a security analyst. Use available tools to retrieve and analyze security incidents with precision.",
        "base_prompt": "Get the full details of the latest security incident",
        "servers": ["GraphSecurityIncidents"],
        "output_variable": "incident_data"
      },
      {
        "step": 2,
        "name": "Threat Intelligence Research",
        "system_prompt": "You are a threat intelligence analyst. Research potential threats and attack patterns.",
        "base_prompt": "Based on this incident: {{incident_data}}\n\nResearch similar threats, IOCs, and attack patterns.",
        "servers": ["brave-search"],
        "input_variables": ["incident_data"],
        "output_variable": "threat_intel",
        "temperature": 0.1
      }
    ],
    "variables": {
      "company_name": "YourCompany",
      "priority_level": "high"
    }
  },
  "process_function_app_data": {
    "name": "process_function_app_data",
    "description": "Process JSON data from Azure Function Apps",
    "steps": [
      {
        "step": 1,
        "name": "Parse and Validate",
        "system_prompt": "You are a data processing assistant. Parse and validate incoming data.",
        "base_prompt": "Process this Function App data: {{stdin}}\n\nValidate structure and extract key fields.",
        "servers": [],
        "output_variable": "parsed_data"
      }
    ]
  }
}
```

### Template Configuration Options

| Option        | Type   | Required | Description                        |
| ------------- | ------ | -------- | ---------------------------------- |
| `name`        | String | ✅        | Unique identifier for the template |
| `description` | String | ✅        | Human-readable description         |
| `steps`       | Array  | ✅        | Array of workflow steps            |
| `variables`   | Object | ❌        | Optional global template variables |

### Step Configuration Options

| Option            | Type    | Required | Description                                |
| ----------------- | ------- | -------- | ------------------------------------------ |
| `step`            | Integer | ✅        | Sequential step number (1, 2, 3...)        |
| `name`            | String  | ✅        | Descriptive step name                      |
| `system_prompt`   | String  | ✅        | AI role and behavior instructions          |
| `base_prompt`     | String  | ✅        | Main prompt with variable substitution     |
| `servers`         | Array   | ❌        | Tarfetted MCP servers to use for this step |
| `output_variable` | String  | ❌        | Variable name to store step output as      |
| `temperature`     | Float   | ❌        | AI creativity level (0.0-1.0)              |

## # Temperature Settings Guide for MCP CLI Go Workflows

## 🌡️ Temperature Overview

Temperature controls the randomness/creativity of AI responses:

- **0.0** - Completely deterministic (same input = same output)
- **0.1-0.3** - Low randomness, high consistency
- **0.4-0.7** - Balanced creativity and consistency
- **0.8-1.0** - High creativity, more variation
- **1.0+** - Very creative, potentially chaotic

## 📊 Recommended Settings by Use Case

### **🔒 Security & Analysis (Temperature: 0.1-0.2)**

### **📝 Technical Documentation (Temperature: 0.2-0.3)**

### **🔍 Research & Summarization (Temperature: 0.2-0.4)**

### **💡 Creative Content (Temperature: 0.5-0.8)**

### **🧠 Brainstorming (Temperature: 0.6-1.0)**

 

Note that Temperature is **hardcoded to 0.7** in other modes (query, chat, interactive)!



### Template Variables

Templates support powerful variable substitution using `{{variable_name}}`:

```json
{
  "base_prompt": "Analyze incident: {{stdin}} for {{company_name}} with {{incident_data}} context"
}
```

**Built-in Variables:**

- `{{stdin}}` - Data from command input or Function Apps
- `{{input_data}}` - Alias for stdin
- `{{step_variable}}` - Output from previous steps

## Servers Configuration

The `servers` section defines MCP servers that templates and queries can use:

```json
"servers": {
  "GraphSecurityIncidents": {
    "command": "graph-security-incidents.exe",
    "system_prompt": "You are a Security Operations Center (SOC) assistant working with Microsoft Graph Security API.",
    "settings": {
      "max_tool_follow_up": 3,
      "request_timeout_seconds": 300,
      "command_timeout_seconds": 900
    }
  },
  "brave-search": {
    "command": "D:\\Github\\PUBLIC-MCPServer-Brave-Search-Golang\\brave-search-mcp.exe"
  },
  "filesystem": {
    "command": "D:\\Github\\PUBLIC-MCPServer-FileSystem-Golang\\filesystem-mcp.exe"
  }
}
```

### Server Configuration Options

| Option          | Type   | Description                                      |
| --------------- | ------ | ------------------------------------------------ |
| `command`       | String | **Required**. Executable path for the MCP server |
| `args`          | Array  | Command-line arguments (optional)                |
| `env`           | Object | Environment variables (optional)                 |
| `system_prompt` | String | Custom system prompt for this server             |
| `settings`      | Object | Server-specific settings                         |

### Server Settings

| Setting                   | Type    | Description                                  |
| ------------------------- | ------- | -------------------------------------------- |
| `max_tool_follow_up`      | Integer | Maximum tool follow-up requests (default: 2) |
| `request_timeout_seconds` | Integer | Request timeout in seconds                   |
| `command_timeout_seconds` | Integer | Command execution timeout                    |

## AI Configuration

The `ai` section configures AI providers with enhanced settings for reliability:

```json
"ai": {
  "default_provider": "openrouter",
  "default_system_prompt": "You are a helpful assistant that answers questions concisely and accurately. You have access to tools and should use them when necessary to answer the question.",
  "interfaces": {
    "openai_compatible": {
      "providers": {
        "openai": {
          "api_key": "sk-proj-your-openai-key",
          "api_endpoint": "https://api.openai.com/v1",
          "default_model": "gpt-4o",
          "available_models": ["gpt-4o", "gpt-4o-mini", "gpt-3.5-turbo"],
          "timeout_seconds": 300,
          "max_retries": 5
        },
        "openrouter": {
          "api_key": "sk-or-v1-your-openrouter-key",
          "api_endpoint": "https://openrouter.ai/api/v1",
          "default_model": "qwen/qwen3-30b-a3b",
          "available_models": ["qwen/qwen3-30b-a3b", "qwen/qwen3-14b"],
          "timeout_seconds": 300,
          "max_retries": 5
        },
        "deepseek": {
          "api_key": "sk-your-deepseek-key",
          "api_endpoint": "https://api.deepseek.com/v1",
          "default_model": "deepseek-chat",
          "available_models": ["deepseek-chat"],
          "timeout_seconds": 300,
          "max_retries": 5
        }
      }
    },
    "anthropic_native": {
      "providers": {
        "anthropic": {
          "api_key": "sk-ant-your-anthropic-key",
          "api_endpoint": "https://api.anthropic.com",
          "default_model": "claude-3-sonnet-20240229",
          "available_models": [
            "claude-3-opus-20240229", 
            "claude-3-sonnet-20240229"
            "claude-3-5-sonnet-20240620",
            "claude-3-7-sonnet-20250219"
          ],
          "timeout_seconds": 300,
          "max_retries": 5
        }
      }
    },
    "gemini_native": {
      "providers": {
        "gemini": {
          "api_key": "AIzaSyA-your-gemini-key",
          "default_model": "gemini-1.5-pro",
          "available_models": ["gemini-1.5-pro", "gemini-1.5-flash"],
          "timeout_seconds": 300,
          "max_retries": 5
        }
      }
    },
    "ollama_native": {
      "providers": {
        "ollama": {
          "api_endpoint": "http://localhost:11434",
          "default_model": "qwen3:30b",
          "available_models": ["qwen3:14b", "llama3.1:8b", "llama3.1:70b"],
          "timeout_seconds": 300,
          "max_retries": 5
        }
      }
    }
  }
}
```

### AI Configuration Options

| Option                  | Type   | Description                                      |
| ----------------------- | ------ | ------------------------------------------------ |
| `default_provider`      | String | Default AI provider for templates and queries    |
| `default_system_prompt` | String | Default system prompt when not specified         |
| `interfaces`            | Object | Interface configurations for different API types |

### Provider Configuration Options

| Option             | Type    | Required | Description                             |
| ------------------ | ------- | -------- | --------------------------------------- |
| `api_key`          | String  | ✅*       | API key (*not needed for Ollama)        |
| `api_endpoint`     | String  | ✅        | Base URL for the provider's API         |
| `default_model`    | String  | ✅        | Default model for this provider         |
| `available_models` | Array   | ✅        | List of available models                |
| `timeout_seconds`  | Integer | ✅        | Request timeout (recommended: 300)      |
| `max_retries`      | Integer | ✅        | Maximum retry attempts (recommended: 5) |

### Interface Types

MCP-CLI supports four interface types:

1. **`openai_compatible`** - OpenAI-compatible APIs (OpenAI, OpenRouter, DeepSeek)
2. **`anthropic_native`** - Anthropic Claude native API
3. **`gemini_native`** - Google Gemini native API  
4. **`ollama_native`** - Ollama local API

## Template Usage

### Command Line Execution

```bash
# Execute a template
mcp-cli.exe --template "incident_review"

# With Function App data via stdin
echo '{"incident_id": "INC-001", "severity": "high"}' | mcp-cli --template "process_function_app_data"

# With input data parameter
mcp-cli.exe --template "incident_review" --input-data "custom data"

# List available templates
mcp-cli.exe --list-templates

# Verbose output for debugging
mcp-cli.exe --template "incident_review" --verbose

# Override server selection
mcp-cli.exe --template "incident_review" --server "GraphSecurityIncidents"
```

### Configuration Examples

### Security-Focused Template Configuration

```json
{
  "templates": {
    "security_incident_analysis": {
      "name": "security_incident_analysis",
      "description": "Complete security incident analysis workflow",
      "steps": [
        {
          "step": 1,
          "name": "Retrieve Incident",
          "system_prompt": "You are a SOC analyst. Retrieve security incidents with precision.",
          "base_prompt": "Get details for the latest high-severity security incident",
          "servers": ["GraphSecurityIncidents"],
          "output_variable": "incident_details",
          "temperature": 0.1
        },
        {
          "step": 2,
          "name": "Threat Research",
          "system_prompt": "You are a threat intelligence analyst. Research threats thoroughly.",
          "base_prompt": "Research threats related to: {{incident_details}}\n\nFocus on IOCs, TTPs, and similar campaigns.",
          "servers": ["brave-search"],
          "output_variable": "threat_intelligence",
          "temperature": 0.2
        },
        {
          "step": 3,
          "name": "Risk Assessment",
          "system_prompt": "You are a CISO. Provide executive-level risk assessment.",
          "base_prompt": "Based on:\nIncident: {{incident_details}}\nThreats: {{threat_intelligence}}\n\nProvide risk level, business impact, and recommendations.",
          "servers": [],
          "temperature": 0.3
        }
      ]
    }
  },
  "servers": {
    "GraphSecurityIncidents": {
      "command": "graph-security-incidents.exe",
      "system_prompt": "You are a Security Operations Center (SOC) assistant with Microsoft Graph Security API access.",
      "settings": {
        "max_tool_follow_up": 3,
        "request_timeout_seconds": 300
      }
    },
    "brave-search": {
      "command": "brave-search-mcp.exe"
    }
  },
  "ai": {
    "default_provider": "openrouter",
    "interfaces": {
      "openai_compatible": {
        "providers": {
          "openrouter": {
            "api_key": "sk-or-v1-your-key",
            "api_endpoint": "https://openrouter.ai/api/v1",
            "default_model": "qwen/qwen3-30b-a3b",
            "timeout_seconds": 300,
            "max_retries": 5
          }
        }
      }
    }
  }
}
```

### Function App Data Processing Configuration

```json
{
  "templates": {
    "webhook_processor": {
      "name": "webhook_processor",
      "description": "Process webhook data from Function Apps",
      "steps": [
        {
          "step": 1,
          "name": "Validate Webhook",
          "system_prompt": "You are a data validation specialist. Ensure data integrity.",
          "base_prompt": "Validate this webhook data: {{stdin}}\n\nCheck required fields, data types, and business rules.",
          "servers": [],
          "output_variable": "validated_data",
          "temperature": 0.1
        },
        {
          "step": 2,
          "name": "Enrich Data",
          "system_prompt": "You are a data enrichment specialist. Add valuable context.",
          "base_prompt": "Enrich this data: {{validated_data}}\n\nAdd relevant context and related information.",
          "servers": ["brave-search"],
          "output_variable": "enriched_data",
          "temperature": 0.2
        }
      ]
    },
    "simple_processor": {
      "name": "simple_processor",
      "description": "Simple single-step data processing",
      "steps": [
        {
          "step": 1,
          "name": "Process Data",
          "system_prompt": "You are a helpful data processor.",
          "base_prompt": "Process this data: {{stdin}}\n\nExtract key information and provide summary.",
          "servers": [],
          "temperature": 0.1
        }
      ]
    }
  },
  "servers": {
    "brave-search": {
      "command": "D:\\Github\\PUBLIC-MCPServer-Brave-Search-Golang\\brave-search-mcp.exe"
    }
  },
  "ai": {
    "default_provider": "openai",
    "interfaces": {
      "openai_compatible": {
        "providers": {
          "openai": {
            "api_key": "sk-proj-your-openai-key",
            "api_endpoint": "https://api.openai.com/v1",
            "default_model": "gpt-4o-mini",
            "timeout_seconds": 300,
            "max_retries": 5
          }
        }
      }
    }
  }
}
```

### Multi-Provider Configuration

```json
{
  "ai": {
    "default_provider": "anthropic",
    "interfaces": {
      "openai_compatible": {
        "providers": {
          "openai": {
            "api_key": "sk-proj-your-openai-key",
            "api_endpoint": "https://api.openai.com/v1",
            "default_model": "gpt-4o",
            "timeout_seconds": 300,
            "max_retries": 5
          },
          "openrouter": {
            "api_key": "sk-or-v1-your-openrouter-key",
            "api_endpoint": "https://openrouter.ai/api/v1",
            "default_model": "qwen/qwen3-30b-a3b",
            "timeout_seconds": 300,
            "max_retries": 5
          }
        }
      },
      "anthropic_native": {
        "providers": {
          "anthropic": {
            "api_key": "sk-ant-your-anthropic-key",
            "api_endpoint": "https://api.anthropic.com",
            "default_model": "claude-3-sonnet-20240229",
            "timeout_seconds": 300,
            "max_retries": 5
          }
        }
      },
      "ollama_native": {
        "providers": {
          "ollama": {
            "api_endpoint": "http://localhost:11434",
            "default_model": "qwen3:30b",
            "timeout_seconds": 300,
            "max_retries": 3
          }
        }
      }
    }
  }
}
```

## Best Practices

### Template Design

1. **Keep steps focused** - Each step should have a single purpose
2. **Use descriptive names** - Clear step and variable names
3. **Specify required servers** - Only connect to needed servers
4. **Use appropriate temperature** - 0.1 for data reliability, higher for creativity

### Provider Configuration

1. **Set reasonable timeouts** - 300 seconds is recommended
2. **Configure retries** - 5 retries for reliability
3. **Choose appropriate models** - Balance cost and capability
4. **Test provider connectivity** - Verify API keys and endpoints

### Function App Integration

1. **Use {{stdin}} variable** - Perfect for JSON payloads
2. **Validate input data** - Always validate Function App data
3. **Handle errors gracefully** - Proper error responses
4. **Keep templates simple** - Avoid overly complex workflows

## Environment Variables

Override API keys using environment variables:

- `OPENAI_API_KEY` - OpenAI API key
- `ANTHROPIC_API_KEY` - Anthropic API key  
- `OPENROUTER_API_KEY` - OpenRouter API key
- `DEEPSEEK_API_KEY` - DeepSeek API key
- `GEMINI_API_KEY` - Google Gemini API key

## Troubleshooting

### Template Issues

```bash
# Debug template execution
mcp-cli.exe --template "template_name" --verbose

# List available templates
mcp-cli.exe --list-templates --verbose

# Test with sample data
echo '{"test": "data"}' | mcp-cli.exe --template "template_name" --verbose
```

## File Location

Default configuration file: `server_config.json` in current directory

Override with `--config` flag:

```bash
mcp-cli.exe --template "name" --config path/to/custom_config.json
```

The new template system transforms MCP-CLI into a powerful workflow automation tool perfect for Function Apps, security operations, and data processing workflows! 🚀
