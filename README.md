# MCP CLI - Golang

[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Build Status](https://img.shields.io/badge/build-passing-brightgreen.svg)](https://github.com/LaurieRhodes/mcp-cli-go)

An enterprise-grade command-line interface for the [Model Context Protocol (MCP)](https://modelcontextprotocol.io/), written in Go.  This tool enables seamless interaction between Large Language Models (LLMs) and external tools/data sources through standardized server connections, with advanced **workflow automation** capabilities for complex multi-step AI tasks.

This utility enables LLM and MCP Tool integration into workflow processes using containers or Function Apps.

## 🌟 What Makes This Special

### 1. **Workflow Templates**

The standout feature - create multi-step AI workflows that chain reasoning, tool calls, and data processing with variable passing between steps.

### 2. **Enterprise Architecture**

Interface-based provider design, robust error handling, and professional logging make this production-ready.

### 3. **Platform Agnostic**

Perfect for Azure Functions, webhooks, containers and automation pipelines with clean JSON I/O.

### 4. **Universal Provider Support**

Works with all major MCP supporting AI providers through a unified interface with provider-specific optimizations.

## 🚀 Features

### Multi-Modal Operation

- **Chat Mode**: Interactive conversational interface with context-aware LLM interactions
- **Query Mode**: Single-shot queries perfect for scripting and automation  
- **Template Mode**: Multi-step workflow automation with variable chaining and server orchestration

### 🌟 Workflow Templates

Create powerful, reusable workflows that chain multiple AI requests across different providers and MCP servers:

- **Multi-Step Automation**: Chain AI reasoning, tool calls, and data processing
- **Variable System**: Pass outputs between steps with `{{variable}}` substitution
- **Server Orchestration**: Use different MCP servers for specific workflow steps
- **Function App Integration Capable**: Perfect for Azure Functions, webhooks, and serverless workflows
- **Enterprise Automation**: Security incident analysis, data processing, research workflows

### Universal AI Provider Support

- **OpenAI**
- **Anthropic**
- **Google Gemini**
- **Ollama**
- **OpenRouter**
- **DeepSeek**

### Enterprise-Grade Architecture

- **Interface-Based Design**: Modular provider architecture reducing code duplication
- **Robust Error Handling**: Comprehensive error recovery and retry logic (2 retries by default)
- **Auto-Configuration**: Automatically generates example config with guided setup
- **Professional Logging**: Structured logging with configurable levels

### Tool Integration

- **MCP Protocol**: Full compliance with Model Context Protocol specification
- **Multi-Server Support**: Connect to multiple MCP servers simultaneously
- **Tool Discovery**: Automatic tool enumeration and validation
- **Streaming Support**: Real-time response streaming for better user experience

## 📦 Installation

### Build from Source

```bash
git clone https://github.com/LaurieRhodes/mcp-cli-go.git
cd mcp-cli-go
go build -o mcp-cli.exe
```

### System Requirements

- Go 1.23+ (for building from source)

## ⚡ Quick Start

### 1. Auto-Configuration

```bash
# First run automatically creates example config
mcp-cli.exe --list-templates

# Outputs:
# 📋 Created example configuration file: server_config.json
# 🔧 Please edit the file to:
#    1. Replace '# your-api-key-here' with your actual API keys
#    2. For Windows users: Download exe servers from GitHub
#    3. Update server paths to point to your downloaded .exe files
#    4. Remove comment fields when ready
```

Note that the example Go ports of common MIT licensed MCP servers referenced in these examples are available at: [GitHub - LaurieRhodes/PUBLIC-Golang-MCP-Servers: Golang Ports of MCP Servers](https://github.com/LaurieRhodes/PUBLIC-Golang-MCP-Servers)

### 2. Basic Operations

```bash
# Simple query
mcp-cli.exe query "What files are in the current directory?"

# Interactive chat
mcp-cli.exe chat --provider openai --model gpt-4o

# List available workflow templates
mcp-cli.exe --list-templates
```

### 3. 🔥 Workflow Templates

```bash
# Execute a pre-built workflow template
mcp-cli.exe --template analyze_file

# Process data through a workflow
echo '{"incident_id": "INC-001"}' | mcp-cli.exe --template search_and_summarize

# File analysis workflow
mcp-cli.exe --template analyze_file --input-data "file_path=./README.md"

# Custom data analysis
cat data.json | mcp-cli.exe --template simple_analyze
```

## 🏗️ Workflow Templates - Complete Example

### Security Incident Analysis Workflow

```json
{
  "templates": {
    "incident_analysis": {
      "name": "incident_analysis",
      "description": "Comprehensive security incident analysis with threat intelligence",
      "steps": [
        {
          "step": 1,
          "name": "Get Incident Details",
          "system_prompt": "You are a security analyst. Retrieve incident data with precision.",
          "base_prompt": "Search for security incident: {{stdin}}",
          "servers": ["brave-search"],
          "output_variable": "incident_data"
        },
        {
          "step": 2,
          "name": "Threat Analysis", 
          "system_prompt": "You are a threat intelligence analyst.",
          "base_prompt": "Analyze this incident for threats: {{incident_data}}\n\nProvide risk assessment and IOCs.",
          "output_variable": "threat_analysis"
        },
        {
          "step": 3,
          "name": "Generate Report",
          "system_prompt": "You are a security manager. Create executive summaries.",
          "base_prompt": "Create incident report:\n\nData: {{incident_data}}\nAnalysis: {{threat_analysis}}\n\nProvide executive summary with recommendations.",
          "input_variables": ["incident_data", "threat_analysis"]
        }
      ]
    }
  }
}
```

## 🔧 Configuration

The tool automatically generates a comprehensive example configuration on first run:

### Auto-Generated Config Features

- **Multi-Provider Setup**: OpenAI, Anthropic, Gemini, Ollama, OpenRouter, DeepSeek
- **Interface-Based Architecture**: Clean separation between provider types
- **Example Templates**: Ready-to-use workflow templates for common tasks
- **Windows & Unix Support**: Works with both .exe servers and npx
- **Guided Setup**: Step-by-step instructions for API keys and server paths

### Example Provider Configuration

```json
{
  "ai": {
    "default_provider": "openrouter",
    "interfaces": {
      "openai_compatible": {
        "providers": {
          "openrouter": {
            "api_key": "# your-openrouter-api-key-here",
            "api_endpoint": "https://openrouter.ai/api/v1",
            "default_model": "qwen/qwen3-30b-a3b",
            "available_models": ["qwen/qwen3-30b-a3b"],
            "max_retries": 2
          }
        }
      },
      "ollama_native": {
        "providers": {
          "ollama": {
            "api_endpoint": "http://localhost:11434",
            "default_model": "ollama.com/ajindal/llama3.1-storm:8b",
            "available_models": ["ollama.com/ajindal/llama3.1-storm:8b", "qwen3:30b"],
            "max_retries": 2
          }
        }
      }
    }
  }
}
```

## 📚 Usage Examples

### Basic Operations

```bash
# Chat mode
mcp-cli chat --provider anthropic --model claude-3-5-sonnet-20241022

# Query with JSON output
mcp-cli query --json "List all .go files and their sizes"

# Query with specific servers
mcp-cli query --server filesystem,brave-search "Search for Go tutorials and save summary to analysis.md in this directory"
```

### 🔥 Workflow Templates

```bash
# Built-in templates
mcp-cli.exe --template analyze_file
mcp-cli.exe --template search_and_summarize  
mcp-cl.exe --template simple_analyze

# Process data through workflow
echo "Analyze this data for security threats" | mcp-cli.exe --template simple_analyze

# File analysis with variables
mcp-cli.exe --template analyze_file --input-data "file_path=./security_logs.txt"

# Automation pipeline
cat incident_data.json | mcp-cli.exe --template incident_analysis > security_report.txt
```

## 🏗️ Architecture

Clean, modular architecture designed for enterprise use:

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Command Layer │────│  Service Layer   │────│ Provider Layer  │
│   (Cobra CLI)   │    │  (Business Logic)│    │  (AI Clients)   │
└─────────────────┘    └──────────────────┘    └─────────────────┘
         │                        │                        │
         ▼                        ▼                        ▼
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│ Workflow Engine │    │ Infrastructure   │    │   MCP Protocol  │
│  (Templates)    │    │  (Config, Log)   │    │  (Server Comm)  │
└─────────────────┘    └──────────────────┘    └─────────────────┘
```

### Key Components

- **Domain Layer**: Core business types and interfaces
- **Workflow Engine**: Template processing and variable management
- **Service Layer**: Business logic and orchestration  
- **Provider Layer**: AI provider implementations with interface-based design
- **Infrastructure**: Auto-configuration, logging, and utilities
- **MCP Protocol**: Server communication and tool management

## 🛠️ Development

### Project Structure

```
mcp-cli-go/
├── cmd/                    # CLI commands (chat, query, template, interactive)
├── internal/
│   ├── core/              # Core business logic
│   ├── domain/            # Domain types, workflow definitions
│   ├── infrastructure/    # Config auto-generation, logging
│   ├── presentation/      # Output formatting
│   ├── providers/         # AI provider implementations  
│   └── services/          # Service layer (workflow engine)
├── docs/                  # Comprehensive documentation
├── examples/              # Example configurations and templates
└── main.go               # Application entry point
```

### Building

```bash
# Build for current platform
go build -o mcp-cli.exe
```

## 🎯 Use Cases

- **Security Operations**: Automated incident analysis and threat intelligence workflows
- **Data Processing**: Multi-step data analysis and enrichment pipelines
- **Research Automation**: Automated research with web search and document analysis
- **Function Apps**: Serverless AI processing for webhooks and APIs
- **DevOps Automation**: Automated analysis of logs, metrics, and system data
- **Content Generation**: Multi-step content creation with research and validation
- 

## 📖 Documentation

### Core Documentation

- [Getting Started Guide](docs/getting_started.md) - Quick Start Guide
- [Template System Guide](docs/templates.md) - Complete workflow template reference
- [Template Reference](docs/template_reference.md) - Schema and examples
- [Configuration Guide](docs/server_config.md) - Complete configuration reference
- [Architecture Overview](docs/architecture.md) - System design and components

## 🤝 Contributing

This project is shared as example code for your own development and alteration.  I'm not certain there would be a lot of interest or value in turning this into a maintained project.  If you think I'm wrong - contact me through details at https://laurierhodes.info

## 🙏 Acknowledgments

This project started in February 2025 as a Golang fork of Chris Hay's (https://github.com/chrishayuk/mcp-cli) as I needed a Go MCP server for use with Go and Function Apps as I experimented with MCP Server development.  That project has contined to grow and is well supported by a team of talented coders.  I'm grateful for the generous sharing of code under MIT License and encourage everyone to look at and support that project as it really is awesome!

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 

</div>
