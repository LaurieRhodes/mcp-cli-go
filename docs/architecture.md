# MCP CLI - Architecture Documentation

## Overview

MCP CLI is a sophisticated command-line interface for the Model Context Protocol (MCP), built with enterprise-grade Go architecture. The system enables seamless integration between Large Language Models and external tools through standardized server connections, supporting multiple operational modes and AI providers.

## Design Philosophy

### Core Principles

1. **Modular Architecture**: Clean separation of concerns with well-defined interfaces
2. **Interface-Based Design**: Provider abstraction enabling easy extensibility  
3. **Concurrent Processing**: Leveraging Go's goroutines for efficient parallel operations
4. **Robust Error Handling**: Comprehensive error recovery and logging throughout
5. **Configuration-Driven**: Flexible JSON-based configuration with environment variable support
6. **Automation-Friendly**: Designed for scripting, CI/CD, and multi-agent scenarios

### Architectural Patterns

- **Hexagonal Architecture**: Domain-centric design with external adapters
- **Factory Pattern**: Dynamic provider instantiation based on configuration
- **Strategy Pattern**: Interchangeable AI provider implementations
- **Repository Pattern**: Configuration and data access abstraction
- **Command Pattern**: CLI command structure with Cobra framework

## System Architecture

### High-Level Component Diagram

```
                           ┌─────────────────────────────────────┐
                           │            Command Layer            │
                           │        (Cobra CLI Framework)       │
                           └─────────────────┬───────────────────┘
                                             │
                           ┌─────────────────▼───────────────────┐
                           │           Service Layer             │
                           │      (Business Logic & Orchestration)
                           └─────────────────┬───────────────────┘
                                             │
                     ┌───────────────────────┼───────────────────────┐
                     │                       │                       │
         ┌───────────▼──────────┐ ┌─────────▼─────────┐ ┌───────────▼──────────┐
         │    Core Modules      │ │   Provider Layer  │ │   Infrastructure     │
         │  (Business Logic)    │ │   (AI & MCP)      │ │   (Config, Logging)  │
         └───────────┬──────────┘ └─────────┬─────────┘ └───────────┬──────────┘
                     │                       │                       │
         ┌───────────▼──────────┐ ┌─────────▼─────────┐ ┌───────────▼──────────┐
         │     Domain Layer     │ │ Presentation Layer│ │    External APIs     │
         │   (Core Types &      │ │  (Output Format)  │ │  (OpenAI, Anthropic, │
         │    Interfaces)       │ │                   │ │   Gemini, Ollama)    │
         └──────────────────────┘ └───────────────────┘ └──────────────────────┘
```

### Layer Responsibilities

| Layer | Responsibility | Key Components |
|-------|---------------|----------------|
| **Command** | CLI interface, flag parsing, user interaction | `cmd/` package with Cobra commands |
| **Service** | Business logic orchestration, workflow management | `internal/services/` |
| **Core** | Mode-specific implementations (chat, query, interactive) | `internal/core/` |
| **Provider** | AI provider clients, MCP protocol implementation | `internal/providers/` |
| **Infrastructure** | Configuration, logging, host management | `internal/infrastructure/` |
| **Domain** | Core types, interfaces, business rules | `internal/domain/` |
| **Presentation** | Output formatting, console interaction | `internal/presentation/` |

## Detailed Architecture

### 1. Command Layer (`cmd/`)

The command layer implements CLI commands using the Cobra framework, providing:

```go
// Command structure
├── root.go         // Global flags, configuration
├── chat.go         // Conversational mode
├── query.go        // Single-shot automation mode  
├── interactive.go  // Direct server interaction mode
└── production.go   // Production-specific commands
```

**Key Features:**
- Global flag management (`--provider`, `--model`, `--config`, `--verbose`)
- Command-specific flag handling
- Configuration loading and validation
- Provider initialization orchestration

### 2. Service Layer (`internal/services/`)

Orchestrates business logic and coordinates between components:

```go
// Service architecture
├── chat/           // Chat mode service
├── query/          // Query mode service  
└── interactive/    // Interactive mode service
```

**Responsibilities:**
- Mode-specific workflow orchestration
- Provider lifecycle management
- Configuration parsing and validation
- Error handling and recovery

### 3. Core Layer (`internal/core/`)

Contains mode-specific business logic implementations:

```go
// Core business logic
├── chat/
│   ├── manager.go    // Chat workflow management
│   ├── context.go    // Conversation state management
│   └── ui.go         // User interface handling
├── query/
│   ├── handler.go    // Query execution logic
│   ├── errors.go     // Error handling
│   └── queryresult.go // Result structures
└── interactive/
    └── service.go    // Interactive mode logic
```

**Key Features:**
- **Chat Manager**: Handles conversation flow, context management, tool execution
- **Query Handler**: Processes single-shot queries with tool integration
- **Interactive Service**: Manages slash commands and direct server communication

### 4. Provider Layer (`internal/providers/`)

Implements external system integrations:

#### AI Providers (`internal/providers/ai/`)

```go
// AI provider architecture
├── factory.go        // Provider factory with interface mapping
├── service.go        // Centralized AI service management
├── streaming/        // Streaming response processors
└── clients/
    ├── openai.go     // OpenAI-compatible providers
    ├── anthropic.go  // Anthropic native implementation
    ├── gemini.go     // Google Gemini native implementation
    └── ollama.go     // Ollama native implementation
```

**Interface-Based Design:**
```go
// Provider types and interfaces
type ProviderType string
const (
    ProviderOpenAI     ProviderType = "openai"
    ProviderAnthropic  ProviderType = "anthropic" 
    ProviderGemini     ProviderType = "gemini"
    ProviderOllama     ProviderType = "ollama"
    ProviderDeepSeek   ProviderType = "deepseek"
    ProviderOpenRouter ProviderType = "openrouter"
)

type InterfaceType string  
const (
    OpenAICompatible InterfaceType = "openai_compatible"
    AnthropicNative  InterfaceType = "anthropic_native"
    GeminiNative     InterfaceType = "gemini_native"
    OllamaNative     InterfaceType = "ollama_native"
)
```

**Key Features:**
- **Factory Pattern**: Dynamic provider instantiation based on configuration
- **Interface Abstraction**: Common `LLMProvider` interface for all providers
- **Native Implementations**: Provider-specific optimizations (streaming, tool calling)
- **Retry Logic**: Exponential backoff with provider-specific error handling

#### MCP Providers (`internal/providers/mcp/`)

```go
// MCP protocol implementation
├── messages/
│   ├── initialize/   // Server initialization messages
│   ├── tools/        // Tool discovery and execution
│   └── json_rpc_message.go // JSON-RPC protocol handling
└── transport/
    └── stdio/        // Standard I/O transport
```

### 5. Infrastructure Layer (`internal/infrastructure/`)

Provides cross-cutting concerns and utilities:

```go
// Infrastructure components  
├── config/
│   ├── service.go      // Configuration service implementation
│   ├── enhanced.go     // Modern interface-based configuration
│   └── legacy.go       // Backward compatibility
├── host/
│   ├── server_manager.go // MCP server lifecycle management
│   ├── environment.go    // Environment variable handling
│   └── ai_options.go     // AI provider option resolution
└── logging/
    ├── logger.go       // Structured logging implementation
    └── production.go   // Production logging configuration
```

**Key Features:**
- **Configuration Hierarchy**: Enhanced > Legacy > Environment Variables > Defaults
- **Server Management**: Process lifecycle, health monitoring, automatic restart
- **Structured Logging**: Configurable levels, colored output, production modes

### 6. Domain Layer (`internal/domain/`)

Defines core business types and interfaces:

```go
// Core domain types
├── interfaces.go     // Primary interfaces and types
└── errors.go        // Domain-specific error types
```

**Key Types:**
```go
// Core domain interfaces
type LLMProvider interface {
    CreateCompletion(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error)
    StreamCompletion(ctx context.Context, req *CompletionRequest, writer io.Writer) (*CompletionResponse, error)
    GetProviderType() ProviderType
    GetInterfaceType() InterfaceType
    ValidateConfig() error
    Close() error
}

type ConfigurationService interface {
    LoadConfig(filePath string) (*ApplicationConfig, error)
    GetProviderConfig(providerName string) (*ProviderConfig, InterfaceType, error)
    // ... other configuration methods
}
```

### 7. Presentation Layer (`internal/presentation/`)

Handles output formatting and user interface:

```go
// Presentation components
├── console/          // Console-specific formatters
├── json/            // JSON output formatting  
└── streaming/       // Streaming response handling
```

## Configuration Architecture

### Hierarchical Configuration System

```
Command Line Args (Highest Priority)
         ↓
Environment Variables  
         ↓
Enhanced Configuration (interfaces)
         ↓
Legacy Configuration (providers)
         ↓
Default Values (Lowest Priority)
```

### Interface-Based Provider Configuration

```json
{
  "ai": {
    "default_provider": "anthropic",
    "interfaces": {
      "openai_compatible": {
        "providers": {
          "openai": {"api_key": "...", "default_model": "gpt-4o"},
          "deepseek": {"api_key": "...", "default_model": "deepseek-chat"},
          "openrouter": {"api_key": "...", "default_model": "qwen/qwen-32b"}
        }
      },
      "anthropic_native": {
        "providers": {
          "anthropic": {"api_key": "...", "default_model": "claude-3-5-sonnet-20240620"}
        }
      },
      "gemini_native": {
        "providers": {
          "gemini": {"api_key": "...", "default_model": "gemini-1.5-pro"}
        }
      },
      "ollama_native": {
        "providers": {
          "ollama": {"api_endpoint": "http://localhost:11434", "default_model": "llama3.1:8b"}
        }
      }
    }
  }
}
```

## Data Flow Architecture

### Chat Mode Data Flow

```
User Input → Chat Service → Chat Manager → AI Provider
    ↓              ↓              ↓            ↓
Context       Configuration   Context      API Request
Management    Resolution      Update       + Tools
    ↓              ↓              ↓            ↓
Tool          Server         Streaming     Response
Execution     Management     Handler       Processing
    ↓              ↓              ↓            ↓
MCP Server → Tool Results → Context → User Display
```

### Query Mode Data Flow

```
CLI Args → Query Service → Query Handler → AI Provider
    ↓           ↓              ↓             ↓
Parameter   Configuration   Request       API Call
Parsing     Loading         Building      + Tools
    ↓           ↓              ↓             ↓
Context     Server         Tool          Response
Loading     Selection      Execution     Processing
    ↓           ↓              ↓             ↓
Output File ← Formatter ← Results ← Tool Results
```

### Interactive Mode Data Flow

```
User Command → Interactive Service → MCP Protocol
     ↓              ↓                     ↓
Command         Server              Tool/Resource
Parsing         Selection           Operations
     ↓              ↓                     ↓
Parameter       Direct              Response
Validation      Communication       Formatting
     ↓              ↓                     ↓
Result Display ← Formatter ← Server Response
```

## Concurrency Model

### Goroutine Usage Patterns

1. **Server Connections**: Each MCP server runs in a separate goroutine
2. **Tool Execution**: Tool calls are executed concurrently when possible
3. **Streaming Responses**: Response processing happens asynchronously
4. **User Input**: Input handling is separated from processing

### Synchronization Mechanisms

```go
// Context-based cancellation
ctx, cancel := context.WithTimeout(context.Background(), timeout)
defer cancel()

// Channel-based communication
type StreamProcessor struct {
    chunks  chan string
    errors  chan error
    done    chan bool
}

// Mutex for shared state
type ChatContext struct {
    mu       sync.RWMutex
    messages []domain.Message
    // ...
}
```

## Error Handling Strategy

### Error Categories

1. **User Errors**: Invalid input, configuration issues
2. **System Errors**: Network failures, API limits
3. **Provider Errors**: AI provider-specific errors
4. **Server Errors**: MCP server communication failures

### Error Handling Patterns

```go
// Structured error types
type ConfigError struct {
    Type    string
    Field   string
    Message string
}

// Retry with exponential backoff
func (c *Client) retryWithBackoff(operation func() error) error {
    for attempt := 0; attempt < maxRetries; attempt++ {
        if err := operation(); err != nil {
            if !isRetryable(err) {
                return err
            }
            time.Sleep(calculateBackoff(attempt))
            continue
        }
        return nil
    }
    return fmt.Errorf("operation failed after %d attempts", maxRetries)
}
```

## Security Architecture

### API Key Management

1. **Environment Variables**: Primary method for API key storage
2. **Configuration Files**: Encrypted or secured configuration storage
3. **Runtime Security**: Keys never logged or exposed in output

### Input Validation

1. **Configuration Validation**: Schema validation for all configuration
2. **Parameter Sanitization**: Input sanitization for all user parameters
3. **Path Validation**: Secure file path handling

### Process Isolation

1. **Server Sandboxing**: MCP servers run in separate processes
2. **Resource Limits**: Memory and CPU limits for server processes
3. **Network Isolation**: Controlled network access for servers

## Performance Architecture

### Optimization Strategies

1. **Connection Pooling**: Reuse HTTP connections for API calls
2. **Caching**: Tool discovery results and configuration caching
3. **Streaming**: Real-time response processing for better UX
4. **Concurrent Execution**: Parallel tool execution when possible

### Resource Management

```go
// Memory management for large responses
type ResponseStream struct {
    buffer   []byte
    maxSize  int
    overflow bool
}

// Connection management
type ProviderClient struct {
    httpClient *http.Client
    rateLimit  *rate.Limiter
    timeout    time.Duration
}
```

## Testing Architecture

### Testing Strategy

1. **Unit Tests**: Individual component testing with mocks
2. **Integration Tests**: Component interaction testing
3. **End-to-End Tests**: Full workflow testing
4. **Performance Tests**: Load and stress testing

### Mock Framework

```go
// Provider mocking
type MockLLMProvider struct {
    responses []domain.CompletionResponse
    errors    []error
    callCount int
}

// Server mocking  
type MockMCPServer struct {
    tools     []domain.Tool
    responses map[string]interface{}
}
```

## Deployment Architecture

### Build Configuration

```go
// Build tags for different environments
// +build production
package main

// Conditional compilation for features
// +build !testing
func initProduction() {
    // Production-specific initialization
}
```

### Distribution

1. **Single Binary**: Statically linked binary with no dependencies
2. **Cross-Platform**: Windows, macOS, Linux support
3. **Containerization**: Docker support for cloud deployments
4. **Package Managers**: Distribution through package managers

## Extensibility Architecture

### Plugin Architecture (Future)

```go
// Plugin interface for future extensions
type Plugin interface {
    Name() string
    Version() string
    Initialize(config map[string]interface{}) error
    Execute(context PluginContext) (interface{}, error)
}
```

### Provider Extension

```go
// Adding new AI providers
type CustomProvider struct {
    config *domain.ProviderConfig
    client *http.Client
}

func (p *CustomProvider) CreateCompletion(ctx context.Context, req *domain.CompletionRequest) (*domain.CompletionResponse, error) {
    // Custom implementation
}
```

## Monitoring and Observability

### Logging Architecture

```go
// Structured logging with levels
logging.Info("Operation completed successfully", 
    "provider", providerName,
    "model", modelName, 
    "duration", elapsed)

logging.Error("Operation failed",
    "error", err,
    "context", operationContext)
```

### Metrics Collection (Future)

```go
// Metrics for monitoring
type Metrics struct {
    RequestCount    prometheus.Counter
    RequestDuration prometheus.Histogram
    ErrorRate       prometheus.Gauge
}
```

## Conclusion

The MCP CLI architecture provides a robust, scalable foundation for Model Context Protocol interactions. The modular design, interface-based provider system, and comprehensive error handling make it suitable for both individual use and enterprise deployment. The architecture supports multiple operational modes while maintaining consistency and reliability across all use cases.

Key architectural strengths:

1. **Modularity**: Clean separation enabling independent development and testing
2. **Extensibility**: Interface-based design allowing easy addition of new providers
3. **Reliability**: Comprehensive error handling and retry mechanisms
4. **Performance**: Concurrent processing and efficient resource management
5. **Security**: Secure credential management and input validation
6. **Maintainability**: Clear structure and well-defined responsibilities

This architecture positions MCP CLI as a professional-grade tool suitable for automation, integration, and enterprise deployment scenarios.
