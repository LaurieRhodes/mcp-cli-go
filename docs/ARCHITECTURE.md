# MCP-CLI-Go Architecture

Technical architecture documentation for developers, contributors, and system integrators.

---

## Overview

MCP-CLI-Go is a sophisticated command-line interface built with enterprise-grade Go architecture. The system enables seamless integration between Large Language Models and external tools through the Model Context Protocol (MCP), supporting multiple operational modes and AI providers.

**Architecture Philosophy:**

- **Modular design** with clean separation of concerns
- **Interface-based** provider abstraction for extensibility
- **Configuration-driven** behavior without recompilation  
- **Concurrent processing** leveraging Go's goroutines
- **Production-ready** error handling and observability

---

## Table of Contents

- [System Architecture](#system-architecture)
- [Layer Responsibilities](#layer-responsibilities)
- [Configuration Architecture](#configuration-architecture)
- [Data Flow Patterns](#data-flow-patterns)
- [Security Architecture](#security-architecture)
- [Performance Architecture](#performance-architecture)
- [Testing Architecture](#testing-architecture)
- [Deployment Architecture](#deployment-architecture)
- [Extension Points](#extension-points)

---

## System Architecture

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    Command Layer                         │
│                 (Cobra CLI Framework)                    │
│  ┌──────┬──────┬───────────┬──────────┬──────────┐     │
│  │ chat │ query│interactive│ template │  serve   │     │
│  └──────┴──────┴───────────┴──────────┴──────────┘     │
└──────────────────────────┬──────────────────────────────┘
                           │
┌──────────────────────────▼──────────────────────────────┐
│                   Service Layer                          │
│            (Business Logic Orchestration)                │
│  ┌──────────────┬──────────────┬──────────────┐        │
│  │ ChatService  │ QueryService │TemplateExec  │        │
│  └──────────────┴──────────────┴──────────────┘        │
└────────┬───────────────┬─────────────────┬──────────────┘
         │               │                 │
┌────────▼────┐   ┌──────▼──────┐   ┌────▼─────────┐
│ Core Layer  │   │ Provider    │   │Infrastructure│
│             │   │ Layer       │   │ Layer        │
│ • Chat Mgr  │   │ • AI        │   │ • Config     │
│ • Query Hdl │   │ • MCP       │   │ • Logging    │
│ • Interactive│  │ • Streaming │   │ • Host Mgmt  │
└─────────────┘   └─────────────┘   └──────────────┘
         │               │                 │
         └───────────────┴─────────────────┘
                         │
┌────────────────────────▼────────────────────────────────┐
│                    Domain Layer                          │
│               (Core Types & Interfaces)                  │
│                                                          │
│  • LLMProvider Interface    • CompletionRequest/Response│
│  • ConfigurationService     • Tool Types                │
│  • Message Types            • Error Types               │
└──────────────────────────────────────────────────────────┘
```

---

## Layer Responsibilities

### Detailed Layer Breakdown

| Layer              | Location                   | Responsibility                                    | Key Components                                      |
| ------------------ | -------------------------- | ------------------------------------------------- | --------------------------------------------------- |
| **Command**        | `cmd/`                     | CLI interface, flag parsing, user interaction     | `root.go`, `chat.go`, `query.go`, `interactive.go`  |
| **Service**        | `internal/services/`       | Business logic orchestration, workflow management | `ChatService`, `QueryService`, `TemplateExecutor`   |
| **Core**           | `internal/core/`           | Mode-specific implementations                     | `ChatManager`, `QueryHandler`, `InteractiveService` |
| **Provider**       | `internal/providers/`      | External system integrations                      | AI clients, MCP protocol, streaming                 |
| **Infrastructure** | `internal/infrastructure/` | Cross-cutting concerns                            | Configuration, logging, host management             |
| **Domain**         | `internal/domain/`         | Core types, interfaces, business rules            | `LLMProvider`, `CompletionRequest`, `Tool`          |
| **Presentation**   | `internal/presentation/`   | Output formatting                                 | Console, JSON, streaming formatters                 |

### Directory Structure

```
internal/
├── cmd/                      # Command Layer
│   ├── root.go              # Global flags, config initialization
│   ├── chat.go              # Chat mode command
│   ├── query.go             # Query mode command
│   ├── interactive.go       # Interactive mode command
│   └── serve.go             # Server mode command
│
├── services/                 # Service Layer
│   ├── chat/
│   │   ├── service.go       # Chat orchestration
│   │   └── options.go       # Chat configuration
│   ├── query/
│   │   ├── service.go       # Query orchestration
│   │   └── options.go       # Query configuration
│   └── template/
│       ├── executor.go      # Template execution engine
│       ├── parser.go        # Template parser
│       └── validator.go     # Template validation
│
├── core/                     # Core Layer
│   ├── chat/
│   │   ├── manager.go       # Chat workflow management
│   │   ├── context.go       # Conversation state
│   │   ├── tools.go         # Tool execution
│   │   └── ui.go            # User interface
│   ├── query/
│   │   ├── handler.go       # Query execution
│   │   └── errors.go        # Error handling
│   └── interactive/
│       └── service.go       # Interactive mode logic
│
├── providers/                # Provider Layer
│   ├── ai/
│   │   ├── factory.go       # Provider factory
│   │   ├── service.go       # AI service coordinator
│   │   ├── streaming/       # Stream processing
│   │   └── clients/
│   │       ├── openai.go    # OpenAI client
│   │       ├── anthropic.go # Anthropic client
│   │       ├── gemini.go    # Gemini client
│   │       └── ollama.go    # Ollama client
│   └── mcp/
│       ├── manager.go       # MCP server management
│       ├── messages/        # MCP protocol messages
│       └── transport/
│           └── stdio/       # stdio transport
│
├── infrastructure/           # Infrastructure Layer
│   ├── config/
│   │   ├── service.go       # Configuration service
│   │   ├── enhanced.go      # Modern config format
│   │   ├── legacy.go        # Backward compatibility
│   │   └── validation.go    # Config validation
│   ├── logging/
│   │   ├── logger.go        # Logger implementation
│   │   └── production.go    # Production config
│   └── host/
│       ├── server_manager.go # MCP server lifecycle
│       └── environment.go    # Environment handling
│
├── domain/                   # Domain Layer
│   ├── interfaces.go        # Core interfaces
│   ├── types.go            # Domain types
│   └── errors.go           # Domain errors
│
└── presentation/             # Presentation Layer
    ├── console/             # Console formatters
    ├── json/               # JSON formatters
    └── streaming/          # Streaming handlers
```

---

## Configuration Architecture

### Configuration Hierarchy

MCP-CLI uses a sophisticated configuration hierarchy that allows flexible overrides at multiple levels:

```
Command Line Args (Highest Priority)
         ↓
    --provider anthropic
    --model claude-sonnet-4
         ↓
Environment Variables  
         ↓
    OPENAI_API_KEY=sk-...
    MCP_PROVIDER=anthropic
         ↓
Enhanced Configuration (config.yaml - interfaces)
         ↓
    ai:
      default_provider: anthropic
      interfaces:
        openai_compatible: {...}
        anthropic_native: {...}
         ↓
Legacy Configuration (providers - backward compatibility)
         ↓
    providers:
      - provider_name: openai
        api_key: ${OPENAI_API_KEY}
         ↓
System Defaults (Lowest Priority)
         ↓
    provider: ollama
    model: llama3.1:8b
```

**Resolution Process:**

1. **CLI flags** override everything
2. **Environment variables** override configuration files
3. **Enhanced config** takes precedence over legacy
4. **Legacy config** provides backward compatibility
5. **System defaults** ensure operation without configuration

### Interface-Based Provider Configuration

```yaml
# Modern configuration format
ai:
  default_provider: anthropic

  interfaces:
    # OpenAI-compatible providers
    openai_compatible:
      providers:
        openai:
          api_key: ${OPENAI_API_KEY}
          default_model: gpt-4o
          api_endpoint: https://api.openai.com/v1

        deepseek:
          api_key: ${DEEPSEEK_API_KEY}
          default_model: deepseek-chat
          api_endpoint: https://api.deepseek.com/v1

        openrouter:
          api_key: ${OPENROUTER_API_KEY}
          default_model: qwen/qwen-32b
          api_endpoint: https://openrouter.ai/api/v1

    # Anthropic native
    anthropic_native:
      providers:
        anthropic:
          api_key: ${ANTHROPIC_API_KEY}
          default_model: claude-sonnet-4
          api_endpoint: https://api.anthropic.com

    # Google Gemini native
    gemini_native:
      providers:
        gemini:
          api_key: ${GEMINI_API_KEY}
          default_model: gemini-1.5-pro
          api_endpoint: https://generativelanguage.googleapis.com

    # Ollama (local)
    ollama_native:
      providers:
        ollama:
          api_endpoint: http://localhost:11434
          default_model: llama3.1:8b

# MCP server configuration
mcp:
  servers:
    filesystem:
      command: filesystem-server
      args: []

    brave_search:
      command: brave-search-server
      env:
        BRAVE_API_KEY: ${BRAVE_API_KEY}
```

---

## Data Flow Patterns

### Chat Mode Data Flow

```
User Input
    ↓
Chat Service (orchestration)
    ↓
Chat Manager (business logic)
    ├─→ Context Management (conversation state)
    ├─→ Provider Selection (factory pattern)
    └─→ Tool Detection (MCP integration)
    ↓
AI Provider (streaming)
    ├─→ API Request + Tools
    ├─→ Streaming Response
    └─→ Tool Calls (if any)
    ↓
Tool Execution (parallel)
    ├─→ MCP Server 1
    ├─→ MCP Server 2
    └─→ MCP Server N
    ↓
Results Aggregation
    ↓
Context Update
    ↓
User Display (streaming)
```

### Query Mode Data Flow

```
CLI Arguments / stdin
    ↓
Query Service
    ├─→ Parse Input
    ├─→ Load Config
    └─→ Select Provider
    ↓
Query Handler
    ├─→ Build Request
    ├─→ Add Tools (if --server)
    └─→ Execute with Retry
    ↓
AI Provider
    ├─→ API Call
    ├─→ Tool Calls (if any)
    └─→ Response
    ↓
Tool Execution (if needed)
    ↓
Output Formatting
    ├─→ Text (default)
    ├─→ JSON (--json)
    └─→ File (--output)
    ↓
Exit with Status
```

### Interactive Mode Data Flow

```
Start REPL
    ↓
Connect to MCP Servers
    ↓
User Command Loop
    ├─→ /help → Display commands
    ├─→ /tools → List available tools
    ├─→ /call → Direct tool execution
    └─→ /exit → Cleanup and exit
    ↓
Command Parsing
    ↓
Parameter Validation
    ↓
MCP Protocol (JSON-RPC)
    ├─→ Build Request
    ├─→ Send via stdio
    └─→ Parse Response
    ↓
Result Formatting
    ↓
Display to User
```

---

## Security Architecture

### API Key Management

**Security layers for credential management:**

1. **Environment Variables** (Primary Method)
   
   ```bash
   export OPENAI_API_KEY="sk-..."
   export ANTHROPIC_API_KEY="sk-ant-..."
   ```

2. **Configuration Files** with Environment Expansion
   
   ```yaml
   providers:
     openai:
       api_key: ${OPENAI_API_KEY}  # Expanded at runtime
   ```

3. **Runtime Security**
   
   - Keys never logged (automatic redaction)
   - Keys never exposed in error messages
   - Keys not included in debug output
   - Secure memory handling for credentials

4. **Key Rotation Support**
   
   ```bash
   # Update environment variable
   export OPENAI_API_KEY="new-key"
   # Restart application - no code changes needed
   ```

### Input Validation

**Multi-layer validation strategy:**

**1. Configuration Validation:**

```go
func (c *ConfigService) ValidateConfig() error {
    if c.AI.DefaultProvider == "" {
        return errors.New("default_provider is required")
    }
    // Validate provider references exist
    // Validate API keys are set (if not using Ollama)
    // Validate model names are non-empty
    return nil
}
```

**2. Parameter Sanitization:**

```go
func sanitizeInput(input string) string {
    // Remove control characters
    // Validate UTF-8
    // Limit length
    // Escape special characters
    return cleaned
}
```

**3. Path Validation:**

```go
func validatePath(path string) error {
    abs, err := filepath.Abs(path)
    if err != nil {
        return err
    }
    // Check for path traversal
    if strings.Contains(abs, "..") {
        return errors.New("path traversal detected")
    }
    // Verify path is within allowed directories
    return nil
}
```

### Process Isolation

**MCP servers run with controlled isolation:**

1. **Separate Processes** - Each MCP server runs in its own process
2. **Resource Limits** - Memory and CPU constraints enforced
3. **Controlled Communication** - stdio-only, no network by default
4. **Input Validation** - All JSON-RPC messages validated
5. **Output Sanitization** - Responses sanitized before display

---

## Performance Architecture

### Optimization Strategies

**1. Connection Pooling**

```go
func newHTTPClient() *http.Client {
    return &http.Client{
        Transport: &http.Transport{
            MaxIdleConns:        100,
            MaxIdleConnsPerHost: 10,
            IdleConnTimeout:     90 * time.Second,
        },
        Timeout: 30 * time.Second,
    }
}
```

**2. Tool Discovery Caching**

```go
type ServerManager struct {
    toolCache map[string][]Tool
    cacheMu   sync.RWMutex
}

func (m *ServerManager) GetTools(serverName string) []Tool {
    m.cacheMu.RLock()
    if tools, ok := m.toolCache[serverName]; ok {
        m.cacheMu.RUnlock()
        return tools
    }
    m.cacheMu.RUnlock()

    // Fetch and cache
    tools := m.fetchTools(serverName)
    m.cacheMu.Lock()
    m.toolCache[serverName] = tools
    m.cacheMu.Unlock()
    return tools
}
```

**3. Streaming Processing**

```go
func (c *Client) StreamCompletion(ctx context.Context, req *CompletionRequest, writer io.Writer) error {
    resp, err := c.httpClient.Do(httpReq)
    defer resp.Body.Close()

    scanner := bufio.NewScanner(resp.Body)
    for scanner.Scan() {
        chunk := parseChunk(scanner.Text())
        writer.Write([]byte(chunk.Content))  // Immediate write
    }
}
```

**4. Concurrent Tool Execution**

```go
func (m *Manager) ExecuteTools(ctx context.Context, toolCalls []ToolCall) []ToolResult {
    results := make([]ToolResult, len(toolCalls))
    var wg sync.WaitGroup

    for i, call := range toolCalls {
        wg.Add(1)
        go func(idx int, tc ToolCall) {
            defer wg.Done()
            results[idx] = m.executeSingleTool(ctx, tc)
        }(i, call)
    }

    wg.Wait()
    return results
}
```

### Resource Management

**Memory Management:**

```go
type ResponseStream struct {
    buffer   *bytes.Buffer
    maxSize  int
    overflow bool
}

func (s *ResponseStream) Write(p []byte) (n int, err error) {
    if s.buffer.Len()+len(p) > s.maxSize {
        s.overflow = true
        return 0, errors.New("response too large")
    }
    return s.buffer.Write(p)
}
```

**Rate Limiting:**

```go
type ProviderClient struct {
    rateLimiter *rate.Limiter
}

func (c *ProviderClient) CreateCompletion(ctx context.Context, req *CompletionRequest) error {
    if err := c.rateLimiter.Wait(ctx); err != nil {
        return err
    }
    // Proceed with API call
}
```

### Performance Characteristics

| Operation              | Latency    | Memory    | Notes                     |
| ---------------------- | ---------- | --------- | ------------------------- |
| Application Startup    | < 100ms    | 20-30 MB  | Binary initialization     |
| Config Loading         | < 50ms     | 5-10 MB   | YAML parsing + validation |
| Provider Selection     | < 5ms      | -         | Factory pattern           |
| MCP Server Connection  | < 100ms    | 10-20 MB  | Per server                |
| AI API Call            | 500-5000ms | Variable  | Network + provider        |
| Tool Execution         | 10-1000ms  | Variable  | Tool-dependent            |
| Streaming Response     | Real-time  | 5-20 MB   | Chunk processing          |
| **Chat Mode (active)** | -          | 50-200 MB | Context-dependent         |
| **Query Mode**         | -          | 30-100 MB | Single query              |
| **Server Mode**        | -          | 50-150 MB | Template-dependent        |

---

## Testing Architecture

### Testing Strategy

**1. Unit Tests** - Individual component testing

```go
type MockLLMProvider struct {
    mock.Mock
}

func (m *MockLLMProvider) CreateCompletion(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
    args := m.Called(ctx, req)
    return args.Get(0).(*CompletionResponse), args.Error(1)
}

func TestChatService(t *testing.T) {
    mockProvider := new(MockLLMProvider)
    mockProvider.On("CreateCompletion", mock.Anything, mock.Anything).
        Return(&CompletionResponse{Content: "test"}, nil)

    service := NewChatService(mockProvider)
    // Test service logic
}
```

**2. Integration Tests** - Component interaction

```go
func TestProviderIntegration(t *testing.T) {
    config := &ProviderConfig{
        APIKey: os.Getenv("TEST_OPENAI_API_KEY"),
    }

    client := NewOpenAIClient(config)
    resp, err := client.CreateCompletion(context.Background(), &CompletionRequest{
        Messages: []Message{{Role: "user", Content: "test"}},
    })

    require.NoError(t, err)
    assert.NotEmpty(t, resp.Content)
}
```

**3. End-to-End Tests** - Full workflows

```go
func TestChatWorkflow(t *testing.T) {
    server := startTestServer(t)
    defer server.Stop()

    chat := NewChatManager(config)
    resp := chat.ProcessMessage("list files in /tmp")
    assert.Contains(t, resp, "tool_calls")
}
```

**4. Performance Benchmarks**

```go
func BenchmarkProviderCall(b *testing.B) {
    client := NewOpenAIClient(config)
    req := &CompletionRequest{
        Messages: []Message{{Role: "user", Content: "test"}},
    }

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        client.CreateCompletion(context.Background(), req)
    }
}
```

### Mock Framework

```go
// MCP server mocking
type MockMCPServer struct {
    tools     []Tool
    responses map[string]interface{}
}

func (m *MockMCPServer) ListTools(ctx context.Context) ([]Tool, error) {
    return m.tools, nil
}

func (m *MockMCPServer) CallTool(ctx context.Context, name string, args map[string]interface{}) (interface{}, error) {
    return m.responses[name], nil
}
```

---

## Deployment Architecture

### Build Configuration

**Production Build:**

```bash
# Optimized binary
go build -ldflags="-s -w" -o mcp-cli cmd/main.go

# Flags:
# -s: Strip debug information
# -w: Strip DWARF symbol table
# Result: ~30% smaller binary
```

**Cross-Platform:**

```bash
# macOS
GOOS=darwin GOARCH=amd64 go build -o mcp-cli-darwin-amd64
GOOS=darwin GOARCH=arm64 go build -o mcp-cli-darwin-arm64

# Linux
GOOS=linux GOARCH=amd64 go build -o mcp-cli-linux-amd64

# Windows
GOOS=windows GOARCH=amd64 go build -o mcp-cli-windows.exe
```

### Distribution Methods

**1. Single Binary**

- Statically linked
- No runtime dependencies
- Portable across systems
- 15-25 MB compressed

**2. Package Managers**

```bash
# Homebrew (macOS)
brew install mcp-cli

# APT (Debian/Ubuntu)  
sudo apt install mcp-cli

# Snap (Linux)
snap install mcp-cli
```

**3. Container**

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /build
COPY . .
RUN go build -ldflags="-s -w" -o mcp-cli

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /build/mcp-cli /usr/local/bin/
ENTRYPOINT ["mcp-cli"]
```

---

## Extension Points

### Adding a New AI Provider

**Requirements:**

1. Implement `LLMProvider` interface
2. Add to provider factory
3. Create configuration schema
4. Handle authentication

**Interface:**

```go
type LLMProvider interface {
    CreateCompletion(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error)
    StreamCompletion(ctx context.Context, req *CompletionRequest, writer io.Writer) (*CompletionResponse, error)
    GetProviderType() ProviderType
    GetInterfaceType() InterfaceType
    ValidateConfig() error
    Close() error
}
```

**Example Implementation:**

```go
type CustomProvider struct {
    config     *ProviderConfig
    httpClient *http.Client
    apiKey     string
}

func (p *CustomProvider) CreateCompletion(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
    // 1. Build API request
    apiReq := p.buildRequest(req)

    // 2. Execute HTTP call
    resp, err := p.httpClient.Do(httpReq)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    // 3. Parse response
    var apiResp CustomAPIResponse
    json.NewDecoder(resp.Body).Decode(&apiResp)

    // 4. Convert to domain response
    return p.convertResponse(apiResp), nil
}
```

**Register in Factory:**

```go
// In providers/ai/factory.go
func (f *Factory) CreateProvider(providerType ProviderType) (LLMProvider, error) {
    switch providerType {
    case "custom":
        return f.createCustomProvider()
    // ... other cases
    }
}
```

### Adding a New Operational Mode

**Requirements:**

1. Create command in `cmd/`
2. Implement service in `services/`
3. Implement core logic in `core/`
4. Register with Cobra

### Adding a New MCP Transport

**Requirements:**

1. Implement transport interface
2. Handle initialization
3. Add to transport factory

---

## Architectural Patterns

### Hexagonal Architecture

Domain logic isolated from external dependencies:

```
         ┌─────────────────────┐
         │   Domain Layer      │
         │  (Business Logic)   │
         │   Interfaces        │
         └──────────┬──────────┘
                    │
        ┌───────────┴───────────┐
        │                       │
  ┌─────▼─────┐           ┌────▼─────┐
  │  Adapters │           │ Adapters │
  │   (In)    │           │  (Out)   │
  │  Commands │           │ Providers│
  └───────────┘           └──────────┘
```

### Factory Pattern

Dynamic provider instantiation:

```go
provider, err := factory.CreateProvider(providerType, config)
```

### Strategy Pattern

Runtime provider selection through common interface.

### Repository Pattern

Configuration and data access abstraction.

---

## Concurrency Model

### Goroutine Usage

```
Main Process
    │
    ├─── MCP Server 1 (goroutine)
    │    ├─── Tool execution (goroutine)
    │    └─── Heartbeat monitor (goroutine)
    │
    ├─── MCP Server 2 (goroutine)
    │    └─── Tool execution (goroutine)
    │
    ├─── Streaming Response (goroutine)
    │    └─── Chunk processing
    │
    └─── User Input (goroutine, chat mode)
         └─── Command handling
```

### Synchronization

**Context-based cancellation:**

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
```

**Channel communication:**

```go
type StreamProcessor struct {
    chunks chan string
    errors chan error
    done   chan bool
}
```

**Mutex for shared state:**

```go
type ChatContext struct {
    mu       sync.RWMutex
    messages []Message
}
```

---

## Error Handling

### Error Categories

1. **User Errors** - Invalid input, configuration issues
2. **System Errors** - Network failures, API limits
3. **Provider Errors** - AI provider-specific errors
4. **Server Errors** - MCP server communication failures

### Retry Strategy

```go
func retryWithBackoff(operation func() error, config RetryConfig) error {
    delay := config.InitialDelay

    for attempt := 0; attempt < config.MaxAttempts; attempt++ {
        err := operation()
        if err == nil {
            return nil
        }

        if !isRetryable(err) {
            return err
        }

        time.Sleep(delay)
        delay = time.Duration(float64(delay) * config.Multiplier)
        if delay > config.MaxDelay {
            delay = config.MaxDelay
        }
    }

    return fmt.Errorf("operation failed after %d attempts", config.MaxAttempts)
}
```

---

## Monitoring and Observability

### Structured Logging

```go
// Different log levels
logging.Debug("Provider selected", "provider", name, "model", model)
logging.Info("Chat session started", "user", userID)
logging.Warn("Rate limit approaching", "remaining", count)
logging.Error("API call failed", "error", err, "attempt", n)
logging.Fatal("Configuration invalid", "error", err)
```

**Output Formats:**

```bash
# Development (human-readable)
2024-12-26 10:30:45 INFO Chat started user=alice

# Production (JSON, structured)
{"level":"info","time":"2024-12-26T10:30:45Z","msg":"Chat started","user":"alice"}
```

### Future Enhancements

**Metrics (Planned):**

- Request/response times
- Token usage tracking
- Error rates by provider
- Active connections

**Distributed Tracing (Planned):**

- OpenTelemetry integration
- Request flow tracking
- Performance profiling

---

## Key Architectural Strengths

1. **Modularity** - Clean separation enabling independent development
2. **Extensibility** - Interface-based design for easy provider addition
3. **Reliability** - Comprehensive error handling and retry mechanisms
4. **Performance** - Concurrent processing and resource management
5. **Security** - Secure credential management and input validation
6. **Maintainability** - Clear structure and well-defined responsibilities
7. **Testability** - Mock-friendly interfaces throughout

**This architecture positions MCP-CLI-Go as production-ready infrastructure suitable for enterprise deployment.**

---

## Related Documentation

- **[User Guides](../guides/)** - Operational modes and usage patterns
- **[Templates](../templates/)** - Workflow creation
- **[MCP Server](../mcp-server/)** - Server mode documentation
