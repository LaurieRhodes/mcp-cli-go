# Command Layer Architecture

This directory contains the command-line interface layer for the MCP CLI application. It follows a clean architecture pattern with clear separation of concerns.

## Directory Structure

```
cmd/
├── README.md          # This documentation
├── root.go            # Root command and global configuration
├── chat.go            # Unified chat command implementation
├── query.go           # Query command for scripting/automation
├── interactive.go     # Interactive mode command
├── handlers/          # Command handlers (business logic)
│   ├── chat_handler.go
│   ├── query_handler.go
│   └── interactive_handler.go
└── formatters/        # Output formatting utilities
    ├── json_formatter.go
    ├── interactive_formatter.go
    └── result_formatter.go
```

## Architecture Principles

### 1. Thin Command Layer

Commands in this directory are **thin adapters** that:

- Parse command-line arguments and flags
- Validate input parameters
- Delegate to appropriate handlers in the `handlers/` subdirectory
- Format and output results using `formatters/`

### 2. Separation of Concerns

- **Commands**: CLI argument parsing and user interaction
- **Handlers**: Business logic and orchestration
- **Formatters**: Output formatting and display logic
- **Services**: Core business logic (in `internal/services/`)

### 3. Dependency Injection

Commands receive their dependencies through constructor injection rather than creating them directly. This improves:

- Testability
- Modularity
- Maintainability

### 4. Error Handling

All commands use consistent error handling patterns:

- Domain-specific error types from `internal/errors/`
- Proper error propagation with context
- User-friendly error messages

## Command Descriptions

### chat.go

**Purpose**: Interactive conversational interface with LLM providers
**Use Cases**: 

- Development and debugging
- Exploratory data analysis
- Interactive assistance

**Architecture**: Uses the unified chat handler that supports all provider types through the factory pattern

### query.go

**Purpose**: Single-shot query execution for automation
**Use Cases**:

- CI/CD pipelines
- Scripting and automation
- Multi-agent workflows
- Batch processing

**Architecture**: Optimized for programmatic use with structured output options

### interactive.go

**Purpose**: Server management and tool exploration
**Use Cases**:

- Server administration
- Tool discovery and testing
- MCP protocol debugging

**Architecture**: Direct server interaction without LLM intermediation

## Usage Patterns

### For Development

```bash
# Interactive exploration
mcp-cli interactive --server filesystem

# Chat-based development
mcp-cli chat --provider anthropic --model claude-3-sonnet
```

### For Automation

```bash
# Scripted queries
mcp-cli query --json --provider openai "Analyze the logs in /var/log"

# Pipeline integration
mcp-cli query --error-code-only --output results.json "Generate report"
```

### For System Administration

```bash
# Server diagnostics
mcp-cli interactive --server filesystem,brave-search

# Tool management
mcp-cli interactive /tools-all
```

## Development Guidelines

### Adding New Commands

1. **Create the command file** in `cmd/`
2. **Implement the handler** in `cmd/handlers/`
3. **Add formatters** in `cmd/formatters/` if needed
4. **Register the command** in `root.go`
5. **Update this README**

### Command Structure Template

```go
// new_command.go
package cmd

import (
    "github.com/LaurieRhodes/mcp-cli-go/cmd/handlers"
    "github.com/spf13/cobra"
)

var NewCmd = &cobra.Command{
    Use:   "new",
    Short: "Brief description",
    Long:  "Detailed description",
    RunE: func(cmd *cobra.Command, args []string) error {
        // Parse flags and arguments
        config := parseNewCommandConfig(cmd, args)

        // Create handler with dependencies
        handler := handlers.NewCommandHandler(
            // inject dependencies
        )

        // Execute business logic
        result, err := handler.Execute(config)
        if err != nil {
            return err
        }

        // Format and output result
        return formatOutput(result, outputFormat)
    },
}
```

### Handler Structure Template

```go
// handlers/new_handler.go
package handlers

type NewCommandHandler struct {
    // dependencies
}

type NewCommandConfig struct {
    // configuration
}

func NewNewCommandHandler(deps...) *NewCommandHandler {
    return &NewCommandHandler{
        // initialize dependencies
    }
}

func (h *NewCommandHandler) Execute(config *NewCommandConfig) (*Result, error) {
    // implement business logic
    // use services from internal/services/
    // return structured result
}
```

# 
