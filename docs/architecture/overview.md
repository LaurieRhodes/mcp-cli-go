# System Overview

High-level architecture of MCP-CLI-Go, design philosophy, and system-wide patterns.

---

## Table of Contents

- [System Architecture](#system-architecture)
- [Design Philosophy](#design-philosophy)
- [Layered Architecture](#layered-architecture)
- [Operational Modes](#operational-modes)
- [Provider System](#provider-system)
- [Configuration Architecture](#configuration-architecture)
- [Concurrency Model](#concurrency-model)

---

## System Architecture

### High-Level System Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                         User Layer                               │
│  (CLI Commands, Terminal Input, Configuration Files)            │
└───────────────────────────────┬─────────────────────────────────┘
                                │
┌───────────────────────────────▼─────────────────────────────────┐
│                      Command Layer                               │
│              (Cobra CLI Framework)                               │
│  ┌──────────┬──────────┬──────────┬──────────┬──────────┐     │
│  │  chat    │  query   │interactive│ template │  serve   │     │
│  └──────────┴──────────┴──────────┴──────────┴──────────┘     │
└───────────────────────────────┬─────────────────────────────────┘
                                │
┌───────────────────────────────▼─────────────────────────────────┐
│                      Service Layer                               │
│           (Business Logic Orchestration)                         │
│  ┌──────────────┬──────────────┬──────────────────┐            │
│  │ ChatService  │ QueryService │ InteractiveService│            │
│  └──────────────┴──────────────┴──────────────────┘            │
└─────────┬────────────────────┬────────────────┬─────────────────┘
          │                    │                │
┌─────────▼────────┐  ┌───────▼────────┐  ┌───▼──────────┐
│   Core Layer     │  │ Provider Layer  │  │Infrastructure│
│                  │  │                 │  │   Layer      │
│ ┌──────────────┐ │  │ ┌────────────┐ │  │┌────────────┐│
│ │ Chat Manager │ │  │ │ AI Factory │ │  ││Config Svc  ││
│ │              │ │  │ │            │ │  ││            ││
│ │ Query Handler│ │  │ │  Clients:  │ │  ││Logger      ││
│ │              │ │  │ │ • OpenAI   │ │  ││            ││
│ │ Interactive  │ │  │ │ • Anthropic│ │  ││Host Mgr    ││
│ │   Service    │ │  │ │ • Gemini   │ │  │└────────────┘│
│ │              │ │  │ │ • Ollama   │ │  │              │
│ └──────────────┘ │  │ └────────────┘ │  │              │
│                  │  │                 │  │              │
│                  │  │ ┌────────────┐ │  │              │
│                  │  │ │ MCP Proto  │ │  │              │
│                  │  │ │ • Messages │ │  │              │
│                  │  │ │ • Transport│ │  │              │
│                  │  │ └────────────┘ │  │              │
└──────────────────┘  └─────────────────┘  └──────────────┘
          │                    │                    │
          └────────────────────┼────────────────────┘
                               │
┌──────────────────────────────▼──────────────────────────────────┐
│                       Domain Layer                               │
│              (Core Types & Interfaces)                           │
│                                                                  │
│  • LLMProvider Interface      • ConfigurationService Interface  │
│  • CompletionRequest/Response • Tool Types                      │
│  • Message Types              • Error Types                     │
└──────────────────────────────────────────────────────────────────┘
          │                    │                    │
          ▼                    ▼                    ▼
┌─────────────────┐  ┌──────────────────┐  ┌─────────────────┐
│  External APIs  │  │   MCP Servers    │  │  File System    │
│                 │  │                  │  │                 │
│ • OpenAI API    │  │ • filesystem     │  │ • Config Files  │
│ • Anthropic API │  │ • brave-search   │  │ • Templates     │
│ • Gemini API    │  │ • database       │  │ • .env          │
│ • Ollama API    │  │ • custom servers │  │                 │
└─────────────────┘  └──────────────────┘  └─────────────────┘
```

---

## Design Philosophy

### Core Principles

#### 1. Separation of Concerns
Each layer has a distinct responsibility:
- **Command Layer:** User interaction and flag parsing
- **Service Layer:** Workflow orchestration
- **Core Layer:** Mode-specific business logic
- **Provider Layer:** External system integration
- **Infrastructure Layer:** Cross-cutting concerns
- **Domain Layer:** Core types and contracts

**Benefit:** Changes in one layer don't cascade to others. Easy to test and maintain.

#### 2. Interface-Based Design
Components communicate through interfaces, not concrete implementations.

```go
// Domain interface
type LLMProvider interface {
    CreateCompletion(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error)
}

// Multiple implementations
type OpenAIClient struct { ... }
type AnthropicClient struct { ... }
type OllamaClient struct { ... }

// All satisfy the same interface
var provider LLMProvider = &OpenAIClient{...}
```

**Benefit:** Easy to swap implementations, mock for testing, and extend.

#### 3. Configuration Over Code
Behavior controlled through configuration:

```yaml
# Change provider without code changes
provider: anthropic  # or openai, gemini, ollama

# Change model without code changes
model: claude-sonnet-4  # or gpt-4o, gemini-pro

# Change templates without code changes
template: code_review  # or security_scan, data_analysis
```

**Benefit:** Users customize behavior. No recompilation needed.

#### 4. Fail-Safe Defaults
System works out of the box with sensible defaults:

```go
// Default to local Ollama if no API keys
if apiKey == "" {
    provider = "ollama"
}

// Default to reasonable token limits
if maxTokens == 0 {
    maxTokens = 4096
}
```

**Benefit:** New users can start immediately.

#### 5. Explicit Over Implicit
Operations are explicit and visible:

```go
// Explicit provider selection
mcp-cli query --provider anthropic "question"

// Explicit template execution  
mcp-cli --template code_review

// Explicit server connection
mcp-cli chat --server filesystem
```

**Benefit:** Users understand what's happening. No surprises.

---

## Layered Architecture

### Layer Responsibilities

```
┌────────────────────────────────────────────────┐
│           Command Layer (cmd/)                  │
│  Responsibilities:                              │
│  • Parse CLI flags and arguments               │
│  • Validate user input                         │
│  • Invoke appropriate service                  │
│  • Format and display output                   │
│  Dependencies: Service Layer, Domain Layer     │
└────────────────┬───────────────────────────────┘
                 │
┌────────────────▼───────────────────────────────┐
│          Service Layer (services/)              │
│  Responsibilities:                              │
│  • Orchestrate business workflows              │
│  • Coordinate between providers                │
│  • Manage transaction boundaries               │
│  • Handle cross-cutting concerns               │
│  Dependencies: Core, Provider, Infrastructure  │
└────────────────┬───────────────────────────────┘
                 │
     ┌───────────┼───────────┐
     │           │           │
┌────▼─────┐ ┌──▼────────┐ ┌▼──────────────┐
│   Core   │ │ Provider  │ │Infrastructure │
│  Layer   │ │  Layer    │ │    Layer      │
│          │ │           │ │               │
│ Business │ │ External  │ │ Cross-cutting │
│  Logic   │ │ Integrations│ Concerns      │
└──────────┘ └───────────┘ └───────────────┘
     │           │           │
     └───────────┼───────────┘
                 │
┌────────────────▼───────────────────────────────┐
│          Domain Layer (domain/)                 │
│  Responsibilities:                              │
│  • Define core types and models                │
│  • Define interfaces (contracts)               │
│  • Define business rules                       │
│  • Define domain errors                        │
│  Dependencies: None (pure domain logic)        │
└────────────────────────────────────────────────┘
```

### Dependency Rules

**Strict Dependency Direction:**
- Command → Service → Core/Provider/Infrastructure → Domain
- Domain depends on nothing
- Upper layers depend on lower layers
- Lower layers NEVER depend on upper layers

**Dependency Inversion:**
```go
// Service layer depends on interface (Domain)
type ChatService struct {
    provider domain.LLMProvider  // Interface
}

// Provider layer implements interface
type OpenAIClient struct { ... }
func (c *OpenAIClient) CreateCompletion(...) { ... }

// Service is decoupled from concrete provider
```

---

## Operational Modes

### Mode Architecture

Each operational mode has three components:

```
┌─────────────────────────────────────────────┐
│              Mode Structure                  │
│                                              │
│  ┌──────────┐    ┌──────────┐    ┌────────┐│
│  │ Command  │───▶│ Service  │───▶│  Core  ││
│  │  (CLI)   │    │(Workflow)│    │(Logic) ││
│  └──────────┘    └──────────┘    └────────┘│
│       │                │              │     │
│       └────────────────┴──────────────┘     │
│                   │                         │
│             Uses Domain Types               │
│             Uses Providers                  │
└─────────────────────────────────────────────┘
```

### Chat Mode

**Purpose:** Interactive conversation with AI and tools

**Components:**
```
cmd/chat.go
    ↓
services/chat/service.go
    ↓
core/chat/manager.go
    ├─→ AI Provider (streaming)
    ├─→ MCP Servers (tool execution)
    └─→ Context Management
```

**Key Characteristics:**
- Maintains conversation context
- Automatic tool execution
- Streaming responses
- Interactive commands (/help, /clear, etc.)

### Query Mode

**Purpose:** Single-shot queries for automation

**Components:**
```
cmd/query.go
    ↓
services/query/service.go
    ↓
core/query/handler.go
    ├─→ AI Provider (completion)
    ├─→ MCP Servers (if needed)
    └─→ Output Formatting
```

**Key Characteristics:**
- Stateless (no conversation history)
- Single request-response
- Scriptable
- Multiple output formats (text, JSON)

### Interactive Mode

**Purpose:** Direct MCP server tool testing

**Components:**
```
cmd/interactive.go
    ↓
services/interactive/service.go
    ↓
core/interactive/service.go
    ├─→ MCP Protocol (direct)
    └─→ Tool Inspection
```

**Key Characteristics:**
- No AI involvement
- Manual tool calling
- Direct MCP communication
- Tool schema inspection

### Workflow Mode

**Purpose:** Multi-step AI workflows

**Components:**
```
cmd/root.go (--template flag)
    ↓
services/template/executor.go
    ↓
core/template/engine.go
    ├─→ Template Parser
    ├─→ Variable Substitution
    ├─→ Step Execution
    │   ├─→ AI Provider
    │   └─→ MCP Servers
    └─→ Result Aggregation
```

**Key Characteristics:**
- Multi-step execution
- Variable interpolation
- Conditional logic
- Template composition

### Server Mode

**Purpose:** Expose templates as MCP server

**Components:**
```
cmd/serve.go
    ↓
services/server/service.go
    ↓
core/server/handler.go
    ├─→ JSON-RPC Server
    ├─→ Tool Registration
    ├─→ Template Mapping
    └─→ Parameter Translation
```

**Key Characteristics:**
- JSON-RPC protocol
- Tool discovery
- Parameter mapping
- Template execution

---

## Provider System

### Provider Architecture

```
┌─────────────────────────────────────────────────┐
│           Provider Factory Pattern               │
│                                                  │
│  Request: (providerType, config)                │
│      │                                           │
│      ▼                                           │
│  ┌──────────────────────────────────────┐       │
│  │         Provider Factory             │       │
│  │  • Maps type to implementation       │       │
│  │  • Creates provider instance          │       │
│  │  • Validates configuration            │       │
│  └──────────────────────────────────────┘       │
│      │                                           │
│      ├──────┬──────┬──────┬──────┬──────┐       │
│      ▼      ▼      ▼      ▼      ▼      ▼       │
│  ┌──────┐┌─────┐┌──────┐┌──────┐┌──────┐┌────┐ │
│  │OpenAI││Anthr││Gemini││Ollama││DeepS.││OR  │ │
│  │Client││opicC││Client││Client││Client││Clnt│ │
│  └──────┘└─────┘└──────┘└──────┘└──────┘└────┘ │
│                                                  │
│  All implement: domain.LLMProvider interface    │
└─────────────────────────────────────────────────┘
```

### Interface Types

Providers are categorized by interface compatibility:

```go
type InterfaceType string

const (
    OpenAICompatible InterfaceType = "openai_compatible"
    AnthropicNative  InterfaceType = "anthropic_native"
    GeminiNative     InterfaceType = "gemini_native"
    OllamaNative     InterfaceType = "ollama_native"
)
```

**OpenAI-Compatible Providers:**
- OpenAI (official)
- DeepSeek
- OpenRouter
- Any provider with OpenAI-compatible API

**Native Providers:**
- Anthropic (streaming, tool use)
- Gemini (native protocol)
- Ollama (local inference)

### Provider Selection Flow

```
User Request
    │
    ▼
┌─────────────────────────┐
│ Determine Provider      │
│ 1. CLI flag (--provider)│
│ 2. Config default       │
│ 3. Template override    │
│ 4. System default       │
└────────┬────────────────┘
         │
         ▼
┌─────────────────────────┐
│ Validate Configuration  │
│ • API key present?      │
│ • Model available?      │
│ • Network accessible?   │
└────────┬────────────────┘
         │
         ▼
┌─────────────────────────┐
│ Create Provider Instance│
│ via Factory             │
└────────┬────────────────┘
         │
         ▼
┌─────────────────────────┐
│ Execute Request         │
│ • Retry on failure      │
│ • Stream if supported   │
│ • Handle tools          │
└─────────────────────────┘
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

### Configuration Structure

```yaml
# config.yaml - Modern structure
version: 2.0

ai:
  default_provider: anthropic
  
  interfaces:
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
    
    anthropic_native:
      providers:
        anthropic:
          api_key: ${ANTHROPIC_API_KEY}
          default_model: claude-sonnet-4
          api_endpoint: https://api.anthropic.com

mcp:
  servers:
    filesystem:
      command: filesystem-server
      args: []
    
    brave_search:
      command: brave-search-server
      env:
        BRAVE_API_KEY: ${BRAVE_API_KEY}

templates:
  paths:
    - ./config/templates
    - ./templates
  
  defaults:
    provider: anthropic
    model: claude-sonnet-4
```

### Configuration Loading Process

```
Application Start
    │
    ▼
┌─────────────────────────┐
│ Load config.yaml        │
│ • Parse YAML            │
│ • Expand env vars       │
│ • Validate schema       │
└────────┬────────────────┘
         │
         ▼
┌─────────────────────────┐
│ Load Provider Configs   │
│ • Load from config/     │
│ • Merge with main       │
│ • Validate credentials  │
└────────┬────────────────┘
         │
         ▼
┌─────────────────────────┐
│ Load MCP Server Configs │
│ • Load from config/     │
│ • Validate binaries     │
│ • Check permissions     │
└────────┬────────────────┘
         │
         ▼
┌─────────────────────────┐
│ Load Templates          │
│ • Scan template paths   │
│ • Parse YAML            │
│ • Build template index  │
└────────┬────────────────┘
         │
         ▼
┌─────────────────────────┐
│ Configuration Ready     │
│ • Provider factory init │
│ • Server manager init   │
│ • Template engine init  │
└─────────────────────────┘
```

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
    ├─── Streaming Response Processor (goroutine)
    │    ├─── Chunk reader
    │    ├─── Chunk writer
    │    └─── Buffer management
    │
    └─── User Input Handler (goroutine, chat mode)
         └─── Command processing
```

### Synchronization Patterns

**Context for Cancellation:**
```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

response, err := provider.CreateCompletion(ctx, request)
```

**Channels for Communication:**
```go
type StreamProcessor struct {
    chunks  chan string        // Data flow
    errors  chan error         // Error reporting
    done    chan bool          // Completion signal
}
```

**Mutex for Shared State:**
```go
type ChatContext struct {
    mu       sync.RWMutex
    messages []domain.Message
    
    func (c *ChatContext) AddMessage(msg domain.Message) {
        c.mu.Lock()
        defer c.mu.Unlock()
        c.messages = append(c.messages, msg)
    }
}
```

### MCP Server Process Management

```
Server Manager
    │
    ├─── Start Server Process
    │    ├─── exec.Command(serverPath)
    │    ├─── Set stdio pipes
    │    ├─── Start process
    │    └─── Store process handle
    │
    ├─── Monitor Health (goroutine)
    │    ├─── Periodic ping
    │    ├─── Check process alive
    │    └─── Restart if crashed
    │
    ├─── Handle Tool Calls
    │    ├─── Send JSON-RPC request
    │    ├─── Wait for response
    │    └─── Parse result
    │
    └─── Shutdown
         ├─── Send shutdown message
         ├─── Wait for graceful exit
         └─── Force kill if timeout
```

---

## Error Handling Architecture

### Error Flow

```
Error Occurs
    │
    ▼
┌─────────────────────────┐
│ Classify Error          │
│ • User error            │
│ • System error          │
│ • Provider error        │
│ • Network error         │
└────────┬────────────────┘
         │
         ▼
┌─────────────────────────┐
│ Determine Handling      │
│ • Retryable?            │
│ • Recoverable?          │
│ • Fatal?                │
└────────┬────────────────┘
         │
    ┌────┼────┐
    │    │    │
    ▼    ▼    ▼
┌────┐┌────┐┌────┐
│Retry││Log││Exit│
└────┘└────┘└────┘
```

### Retry Logic

```go
type RetryConfig struct {
    MaxAttempts  int
    InitialDelay time.Duration
    MaxDelay     time.Duration
    Multiplier   float64
}

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

## Performance Characteristics

### Request Latency

```
Component                 Latency
─────────────────────────────────────
CLI Parsing               < 1ms
Config Loading            < 10ms
Provider Selection        < 1ms
Template Parsing          < 5ms
MCP Server Connection     < 50ms
AI API Call               500-5000ms  (variable)
Tool Execution            10-1000ms   (tool-dependent)
Response Formatting       < 10ms
Total (simple query)      500-6000ms
Total (with tools)        1000-10000ms
```

### Memory Usage

```
Component                 Memory
─────────────────────────────────────
Base Application          20-30 MB
Configuration             1-5 MB
Chat Context              5-50 MB    (grows with history)
Template Cache            1-10 MB
MCP Server Process        10-50 MB   (per server)
Provider Client           5-20 MB
Stream Buffer             1-5 MB
Total (typical)           50-200 MB
```

---

---

## State Management Strategy

### Stateful vs. Stateless Operations

**Stateless Operations (Query, Template):**
- No conversation history
- Each request independent
- No shared state between requests
- Memory released after completion

**Stateful Operations (Chat, Server):**
- Conversation history maintained
- Context carries across requests
- Shared state requires synchronization
- Memory grows with usage

### Chat State Management

**State Components:**
```go
type ChatContext struct {
    mu           sync.RWMutex
    messages     []Message          // Conversation history
    systemPrompt string            // System instructions
    metadata     map[string]interface{}  // Session metadata
    
    // Resource management
    maxMessages  int               // Trim threshold
    tokenBudget  int               // Token limit
}
```

**State Lifecycle:**

```
Session Start
    │
    ▼
Initialize Context
    ├─→ Load system prompt
    ├─→ Set resource limits
    └─→ Initialize metadata
    │
    ▼
Message Loop
    ├─→ Add user message → Lock
    ├─→ Generate response → Lock
    ├─→ Add assistant message → Lock
    ├─→ Check limits → Trim if needed
    │
    ▼
Session End
    └─→ Release resources
```

**Concurrency Control:**
- Read-write mutex for message list
- Single-writer guarantee (only chat manager modifies)
- Multiple readers allowed (context inspection)
- Lock granularity: per-operation, not per-message

**Memory Management:**
```go
func (c *ChatContext) trimMessages() {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    if len(c.messages) <= c.maxMessages {
        return
    }
    
    // Keep system prompt + recent messages
    systemMsg := c.messages[0]
    recentMessages := c.messages[len(c.messages)-c.maxMessages+1:]
    
    c.messages = append([]Message{systemMsg}, recentMessages...)
    
    runtime.GC()  // Hint to GC to reclaim old messages
}
```

### MCP Server State

**Server Connection State:**
```go
type ServerState int

const (
    StateUninitialized ServerState = iota
    StateConnecting
    StateInitialized
    StateReady
    StateError
    StateStopped
)

type ServerConnection struct {
    mu      sync.RWMutex
    state   ServerState
    process *exec.Cmd
    toolCache map[string][]Tool  // Cached tool list
}
```

**State Transitions:**
```
Uninitialized
    │
    ├─→ Start() → Connecting
    │                │
    │                ├─→ Success → Initialized
    │                │                │
    │                │                ├─→ ListTools() → Ready
    │                │                │
    │                │                └─→ Error → Error
    │                │
    │                └─→ Failure → Error
    │
    └─→ Any State → Stop() → Stopped
```

**Race Condition Prevention:**
```go
func (s *ServerConnection) GetTools() ([]Tool, error) {
    s.mu.RLock()
    if s.state != StateReady {
        s.mu.RUnlock()
        return nil, errors.New("server not ready")
    }
    
    // Check cache
    if tools, ok := s.toolCache["tools"]; ok {
        s.mu.RUnlock()
        return tools, nil
    }
    s.mu.RUnlock()
    
    // Need to fetch - upgrade to write lock
    s.mu.Lock()
    defer s.mu.Unlock()
    
    // Double-check after acquiring write lock
    if tools, ok := s.toolCache["tools"]; ok {
        return tools, nil
    }
    
    // Fetch and cache
    tools, err := s.fetchTools()
    if err == nil {
        s.toolCache["tools"] = tools
    }
    return tools, err
}
```

---

## Distributed System Concerns

### Process Coordination

**Challenge:** MCP servers run as separate processes. Must coordinate:
- Server lifecycle (start, stop, restart)
- Request/response correlation
- Concurrent tool calls
- Failure detection and recovery

**Solution: Process Manager Pattern**

```go
type ProcessManager struct {
    mu        sync.RWMutex
    processes map[string]*ManagedProcess
}

type ManagedProcess struct {
    cmd       *exec.Cmd
    stdin     io.WriteCloser
    stdout    io.ReadCloser
    stderr    io.ReadCloser
    
    // Coordination
    requests  map[int]chan Response  // Request ID → response channel
    requestMu sync.Mutex
    nextID    int
    
    // Health monitoring
    healthy   bool
    lastPing  time.Time
}
```

**Request Correlation:**
```go
func (p *ManagedProcess) SendRequest(ctx context.Context, method string, params interface{}) (Response, error) {
    // Allocate request ID
    p.requestMu.Lock()
    id := p.nextID
    p.nextID++
    responseChan := make(chan Response, 1)
    p.requests[id] = responseChan
    p.requestMu.Unlock()
    
    // Send request
    request := JSONRPCRequest{ID: id, Method: method, Params: params}
    json.NewEncoder(p.stdin).Encode(request)
    
    // Wait for response with timeout
    select {
    case response := <-responseChan:
        return response, nil
    case <-ctx.Done():
        return Response{}, ctx.Err()
    case <-time.After(30 * time.Second):
        return Response{}, errors.New("request timeout")
    }
}

// Response handler goroutine
func (p *ManagedProcess) handleResponses() {
    scanner := bufio.NewScanner(p.stdout)
    for scanner.Scan() {
        var response JSONRPCResponse
        json.Unmarshal(scanner.Bytes(), &response)
        
        p.requestMu.Lock()
        if ch, ok := p.requests[response.ID]; ok {
            ch <- response.Result
            delete(p.requests, response.ID)
        }
        p.requestMu.Unlock()
    }
}
```

### Partial Failure Handling

**Scenario:** User requests analysis that requires 3 tools. Tool 1 succeeds, Tool 2 fails, Tool 3 succeeds.

**Strategy:** Collect all results, return partial success with errors.

```go
type ToolResult struct {
    ToolName string
    Success  bool
    Result   interface{}
    Error    error
}

func (m *Manager) ExecuteTools(ctx context.Context, toolCalls []ToolCall) []ToolResult {
    results := make([]ToolResult, len(toolCalls))
    var wg sync.WaitGroup
    
    for i, call := range toolCalls {
        wg.Add(1)
        go func(idx int, tc ToolCall) {
            defer wg.Done()
            
            result, err := m.executeSingleTool(ctx, tc)
            results[idx] = ToolResult{
                ToolName: tc.ToolName,
                Success:  err == nil,
                Result:   result,
                Error:    err,
            }
        }(i, call)
    }
    
    wg.Wait()
    return results  // Returns all results, even if some failed
}
```

**AI Provider receives:**
```json
{
  "tool_results": [
    {"tool": "analyze_code", "success": true, "result": "..."},
    {"tool": "check_security", "success": false, "error": "server timeout"},
    {"tool": "suggest_improvements", "success": true, "result": "..."}
  ]
}
```

**AI can decide:** Continue with partial results or retry failed tools.

### Timeout and Deadline Propagation

**Problem:** Nested operations need coordinated timeouts.

```
User Request (30s timeout)
    │
    ├─→ Chat Service (25s remaining)
    │       │
    │       ├─→ AI Provider Call (20s remaining)
    │       │       │
    │       │       └─→ Network Request (15s remaining)
    │       │
    │       └─→ Tool Execution (10s remaining)
    │               │
    │               └─→ MCP Server Call (5s remaining)
```

**Solution: Context-based deadline propagation**

```go
func ProcessUserRequest(userTimeout time.Duration) error {
    // Create parent context with deadline
    ctx, cancel := context.WithTimeout(context.Background(), userTimeout)
    defer cancel()
    
    // Pass context down - automatically inherits deadline
    return chatService.HandleMessage(ctx, message)
}

func (s *ChatService) HandleMessage(ctx context.Context, msg string) error {
    // Check if we still have time
    if deadline, ok := ctx.Deadline(); ok {
        remaining := time.Until(deadline)
        if remaining < 5*time.Second {
            return errors.New("insufficient time remaining")
        }
    }
    
    // Call provider - automatically respects parent deadline
    return s.provider.CreateCompletion(ctx, request)
}
```

---

## Bounded Contexts and Domain Boundaries

### Context Map

```
┌─────────────────────────────────────────────────────┐
│              Command Context (CLI)                   │
│                                                      │
│  Responsibility: User interaction, routing           │
│  Language: Commands, flags, arguments                │
│  Dependencies: None (entry point)                    │
└──────────────────┬──────────────────────────────────┘
                   │ Commands
                   ▼
┌─────────────────────────────────────────────────────┐
│           Execution Context (Business Logic)         │
│                                                      │
│  Responsibility: Workflow orchestration              │
│  Language: Services, handlers, managers              │
│  Dependencies: Domain, Provider, Infrastructure      │
└──────────────────┬──────────────────────────────────┘
                   │ Requests
                   ▼
┌─────────────────────────────────────────────────────┐
│           Provider Context (Integration)             │
│                                                      │
│  Responsibility: External system communication       │
│  Language: Clients, adapters, protocols              │
│  Dependencies: Domain (interfaces only)              │
└─────────────────────────────────────────────────────┘
```

### Anti-Corruption Layers

**Problem:** External APIs have different models than our domain.

**Solution:** Adapter pattern with translation layer.

```go
// Domain model (our language)
type CompletionRequest struct {
    Messages    []Message
    MaxTokens   int
    Temperature float64
}

// OpenAI API model (their language)
type OpenAIRequest struct {
    Model       string                 `json:"model"`
    Messages    []OpenAIMessage        `json:"messages"`
    MaxTokens   int                    `json:"max_tokens"`
    Temperature float64                `json:"temperature"`
}

// Anti-corruption layer
func (c *OpenAIClient) CreateCompletion(req *CompletionRequest) (*CompletionResponse, error) {
    // Translate our model to their model
    apiReq := &OpenAIRequest{
        Model:       c.model,
        Messages:    convertMessages(req.Messages),  // Translation
        MaxTokens:   req.MaxTokens,
        Temperature: req.Temperature,
    }
    
    // Call their API
    apiResp, err := c.callOpenAI(apiReq)
    
    // Translate their model back to our model
    return convertResponse(apiResp), err  // Translation
}
```

**Benefit:** Domain model stays clean. API changes isolated to adapter.

---

## Scalability Analysis

### Vertical Scaling (Single Instance)

**Current Limits:**

| Resource | Limit | Reason |
|----------|-------|--------|
| **Chat Context** | 1000 messages | Memory growth (500MB) |
| **Concurrent MCP Servers** | 20 servers | Memory overhead (1GB) |
| **Template Nesting** | 10 levels | Stack depth |
| **Concurrent Tool Calls** | 1000 goroutines | Goroutine overhead |

**Bottlenecks:**
1. **Memory** - Linear growth with chat history and MCP servers
2. **Provider API Rate Limits** - External constraint
3. **Sequential Template Steps** - Can't parallelize without refactoring

**Optimization Strategies:**
- Context compression (summarize old messages)
- Lazy MCP server loading
- Template step parallelization (future)
- Response caching

### Horizontal Scaling (Multiple Instances)

**Current State:** Not supported (no shared state mechanism)

**Challenges:**
- Chat sessions are stateful (can't distribute)
- MCP servers are process-local (can't share)
- No coordination mechanism

**Future Architecture for Scale:**

```
┌─────────────┐   ┌─────────────┐   ┌─────────────┐
│ MCP-CLI     │   │ MCP-CLI     │   │ MCP-CLI     │
│ Instance 1  │   │ Instance 2  │   │ Instance 3  │
└──────┬──────┘   └──────┬──────┘   └──────┬──────┘
       │                 │                 │
       └─────────────────┼─────────────────┘
                         │
                    ┌────▼─────┐
                    │  Redis   │  Shared state
                    │ (Cache)  │
                    └──────────┘
```

**Required Changes:**
- Externalize chat context (Redis/PostgreSQL)
- Centralized MCP server pool
- Session affinity or sticky routing
- Distributed locking

---

## Performance Optimization Patterns

### 1. Connection Pooling

**Problem:** Creating new HTTP connection for each API call is slow.

**Solution:** Reuse connections via `http.Client` connection pool.

```go
func newHTTPClient() *http.Client {
    return &http.Client{
        Transport: &http.Transport{
            // Connection pooling
            MaxIdleConns:        100,
            MaxIdleConnsPerHost: 10,
            IdleConnTimeout:     90 * time.Second,
            
            // Connection reuse
            DisableKeepAlives: false,
            
            // TLS optimization
            TLSHandshakeTimeout: 10 * time.Second,
        },
        Timeout: 30 * time.Second,
    }
}
```

**Benefit:** 50-100ms latency reduction per API call (eliminates TCP handshake + TLS handshake).

### 2. Tool Discovery Caching

**Problem:** Listing tools from MCP server requires JSON-RPC call.

**Solution:** Cache tool list after first retrieval.

```go
func (m *ServerManager) GetTools(serverName string) ([]Tool, error) {
    m.cacheMu.RLock()
    if tools, ok := m.toolCache[serverName]; ok {
        m.cacheMu.RUnlock()
        return tools, nil  // Return cached
    }
    m.cacheMu.RUnlock()
    
    // Fetch from server
    tools, err := m.fetchToolsFromServer(serverName)
    if err != nil {
        return nil, err
    }
    
    // Cache for future calls
    m.cacheMu.Lock()
    m.toolCache[serverName] = tools
    m.cacheMu.Unlock()
    
    return tools, nil
}
```

**Benefit:** 10-50ms latency reduction for subsequent tool calls.

### 3. Streaming Response Processing

**Problem:** Waiting for complete response before displaying is slow UX.

**Solution:** Stream chunks as they arrive.

```go
func (c *Client) StreamCompletion(ctx context.Context, req *CompletionRequest, writer io.Writer) error {
    // Make request
    resp, err := c.httpClient.Do(httpReq)
    defer resp.Body.Close()
    
    // Process Server-Sent Events
    scanner := bufio.NewScanner(resp.Body)
    for scanner.Scan() {
        line := scanner.Text()
        
        if strings.HasPrefix(line, "data: ") {
            chunk := parseChunk(line)
            writer.Write([]byte(chunk.Content))  // Immediate display
        }
    }
}
```

**Benefit:** First token appears ~500ms faster. Better perceived performance.

### 4. Goroutine Pooling (Not Implemented)

**Problem:** Creating 1000 goroutines for 1000 tool calls has overhead.

**Proposed Solution:** Worker pool pattern.

```go
type WorkerPool struct {
    workers   int
    taskQueue chan func()
}

func NewWorkerPool(workers int) *WorkerPool {
    pool := &WorkerPool{
        workers:   workers,
        taskQueue: make(chan func(), workers*2),
    }
    
    for i := 0; i < workers; i++ {
        go pool.worker()
    }
    
    return pool
}

func (p *WorkerPool) worker() {
    for task := range p.taskQueue {
        task()  // Execute task
    }
}

func (p *WorkerPool) Submit(task func()) {
    p.taskQueue <- task
}
```

**Benefit:** Reduced goroutine creation overhead for large batches.

---

## Next Steps

- **[Components](components.md)** - Detailed component architecture
- **[Data Flow](data-flow.md)** - Request/response flows
- **[API & Domain](api-domain.md)** - Interfaces and types

---

**Understanding the architecture?** Continue to [Components](components.md) for detailed component design. 🏗️
