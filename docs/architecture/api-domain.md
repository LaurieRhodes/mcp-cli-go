# API & Domain Model

Core types, interfaces, and API contracts that define the MCP-CLI domain model and extension points.

---

## Table of Contents

- [Domain Interfaces](#domain-interfaces)
- [Core Types](#core-types)
- [Provider API](#provider-api)
- [Configuration Schema](#configuration-schema)
- [MCP Protocol Types](#mcp-protocol-types)
- [Template Schema](#template-schema)
- [Extension Points](#extension-points)

---

## Domain Interfaces

### LLMProvider Interface

**Purpose:** Abstraction for AI provider implementations

**Location:** `internal/domain/interfaces.go`

```go
type LLMProvider interface {
    // CreateCompletion executes a completion request
    CreateCompletion(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error)
    
    // StreamCompletion executes a streaming completion request
    StreamCompletion(ctx context.Context, req *CompletionRequest, writer io.Writer) (*CompletionResponse, error)
    
    // GetProviderType returns the provider identifier
    GetProviderType() ProviderType
    
    // GetInterfaceType returns the API interface type
    GetInterfaceType() InterfaceType
    
    // ValidateConfig validates provider configuration
    ValidateConfig() error
    
    // Close cleans up provider resources
    Close() error
}
```

**Implementation Requirements:**

1. **Thread-Safe:** Must be safe for concurrent use
2. **Context-Aware:** Must respect context cancellation
3. **Error Handling:** Return structured errors
4. **Resource Management:** Properly clean up in Close()

**Example Implementation Skeleton:**

```go
type MyProvider struct {
    config     *ProviderConfig
    httpClient *http.Client
    apiKey     string
}

func (p *MyProvider) CreateCompletion(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
    // 1. Validate request
    if err := p.validateRequest(req); err != nil {
        return nil, fmt.Errorf("invalid request: %w", err)
    }
    
    // 2. Build API request
    apiReq := p.buildAPIRequest(req)
    
    // 3. Execute HTTP request
    httpReq, _ := http.NewRequestWithContext(ctx, "POST", p.endpoint, apiReq)
    resp, err := p.httpClient.Do(httpReq)
    if err != nil {
        return nil, fmt.Errorf("API call failed: %w", err)
    }
    defer resp.Body.Close()
    
    // 4. Parse response
    var apiResp MyAPIResponse
    if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
        return nil, fmt.Errorf("parse failed: %w", err)
    }
    
    // 5. Convert to domain response
    return p.convertResponse(apiResp), nil
}
```

---

### ConfigurationService Interface

**Purpose:** Configuration loading and management

**Location:** `internal/domain/interfaces.go`

```go
type ConfigurationService interface {
    // LoadConfig loads configuration from file
    LoadConfig(path string) (*ApplicationConfig, error)
    
    // GetProviderConfig retrieves provider configuration
    GetProviderConfig(provider string) (*ProviderConfig, InterfaceType, error)
    
    // GetServerConfig retrieves MCP server configuration
    GetServerConfig(server string) (*ServerConfig, error)
    
    // GetTemplateConfig retrieves template configuration
    GetTemplateConfig(template string) (*TemplateConfig, error)
    
    // ValidateConfig validates entire configuration
    ValidateConfig() error
    
    // ReloadConfig reloads configuration from disk
    ReloadConfig() error
}
```

---

### WorkflowExecutor Interface

**Purpose:** Template execution engine

```go
type TemplateExecutor interface {
    // Execute runs template with given input
    Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error)
    
    // ValidateTemplate checks template structure
    ValidateTemplate(template *Template) error
    
    // GetTemplate loads template by name
    GetTemplate(name string) (*Template, error)
    
    // ListTemplates returns all available templates
    ListTemplates() ([]TemplateInfo, error)
}
```

---

## Core Types

### CompletionRequest

**Purpose:** Standardized completion request across all providers

```go
type CompletionRequest struct {
    // Messages in the conversation
    Messages []Message `json:"messages"`
    
    // Maximum tokens to generate
    MaxTokens int `json:"max_tokens,omitempty"`
    
    // Temperature (0.0-2.0, typically 0.0-1.0)
    Temperature float64 `json:"temperature,omitempty"`
    
    // Tools available for use
    Tools []Tool `json:"tools,omitempty"`
    
    // How to use tools: "none", "auto", "required"
    ToolChoice string `json:"tool_choice,omitempty"`
    
    // Stop sequences
    Stop []string `json:"stop,omitempty"`
    
    // Presence penalty (-2.0 to 2.0)
    PresencePenalty float64 `json:"presence_penalty,omitempty"`
    
    // Frequency penalty (-2.0 to 2.0)
    FrequencyPenalty float64 `json:"frequency_penalty,omitempty"`
    
    // Top P (0.0-1.0)
    TopP float64 `json:"top_p,omitempty"`
    
    // Model identifier (optional override)
    Model string `json:"model,omitempty"`
}
```

---

### CompletionResponse

**Purpose:** Standardized completion response

```go
type CompletionResponse struct {
    // Unique response identifier
    ID string `json:"id"`
    
    // Response content
    Content string `json:"content"`
    
    // Tool calls requested by model
    ToolCalls []ToolCall `json:"tool_calls,omitempty"`
    
    // Token usage information
    Usage Usage `json:"usage"`
    
    // Model that generated response
    Model string `json:"model"`
    
    // Finish reason: "stop", "length", "tool_calls"
    FinishReason string `json:"finish_reason"`
    
    // Provider-specific metadata
    Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// HasToolCalls returns true if response contains tool calls
func (r *CompletionResponse) HasToolCalls() bool {
    return len(r.ToolCalls) > 0
}
```

---

### Message

**Purpose:** Conversation message

```go
type Message struct {
    // Role: "system", "user", "assistant", "tool"
    Role string `json:"role"`
    
    // Message content
    Content string `json:"content"`
    
    // Name (optional, for function calls)
    Name string `json:"name,omitempty"`
    
    // Tool calls (for assistant messages)
    ToolCalls []ToolCall `json:"tool_calls,omitempty"`
    
    // Tool call ID (for tool role messages)
    ToolCallID string `json:"tool_call_id,omitempty"`
}
```

**Role Types:**

| Role | Purpose | Who Sends |
|------|---------|-----------|
| `system` | System instructions | User/Config |
| `user` | User messages | User |
| `assistant` | AI responses | AI Provider |
| `tool` | Tool execution results | MCP Server |

---

### Tool

**Purpose:** Tool definition for AI

```go
type Tool struct {
    // Tool name (must be unique)
    Name string `json:"name"`
    
    // Human-readable description
    Description string `json:"description"`
    
    // JSON Schema for parameters
    Parameters map[string]interface{} `json:"parameters"`
    
    // Server providing this tool (internal)
    ServerName string `json:"-"`
}
```

**Parameter Schema Example:**

```go
tool := Tool{
    Name: "read_file",
    Description: "Read contents of a file",
    Parameters: map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "path": map[string]interface{}{
                "type": "string",
                "description": "Path to file",
            },
        },
        "required": []string{"path"},
    },
}
```

---

### ToolCall

**Purpose:** AI request to execute a tool

```go
type ToolCall struct {
    // Unique identifier for this call
    ID string `json:"id"`
    
    // Tool name to call
    ToolName string `json:"tool_name"`
    
    // Server providing the tool (resolved internally)
    ServerName string `json:"server_name"`
    
    // Arguments as JSON object
    Arguments map[string]interface{} `json:"arguments"`
}
```

---

### ToolResult

**Purpose:** Result from tool execution

```go
type ToolResult struct {
    // Tool call ID this result corresponds to
    ToolCallID string `json:"tool_call_id"`
    
    // Result content
    Content interface{} `json:"content"`
    
    // Error message if execution failed
    Error string `json:"error,omitempty"`
    
    // Whether execution was successful
    Success bool `json:"success"`
}
```

---

### Usage

**Purpose:** Token usage tracking

```go
type Usage struct {
    // Tokens in the prompt
    PromptTokens int `json:"prompt_tokens"`
    
    // Tokens in the completion
    CompletionTokens int `json:"completion_tokens"`
    
    // Total tokens used
    TotalTokens int `json:"total_tokens"`
}
```

---

## Provider API

### Provider Types

```go
type ProviderType string

const (
    ProviderOpenAI     ProviderType = "openai"
    ProviderAnthropic  ProviderType = "anthropic"
    ProviderGemini     ProviderType = "gemini"
    ProviderOllama     ProviderType = "ollama"
    ProviderDeepSeek   ProviderType = "deepseek"
    ProviderOpenRouter ProviderType = "openrouter"
)
```

### Interface Types

```go
type InterfaceType string

const (
    // OpenAI-compatible API (chat/completions endpoint)
    OpenAICompatible InterfaceType = "openai_compatible"
    
    // Anthropic native API (messages endpoint)
    AnthropicNative InterfaceType = "anthropic_native"
    
    // Google Gemini native API
    GeminiNative InterfaceType = "gemini_native"
    
    // Ollama native API (local)
    OllamaNative InterfaceType = "ollama_native"
)
```

### ProviderConfig

```go
type ProviderConfig struct {
    // Provider identifier
    ProviderName string `yaml:"provider_name" json:"provider_name"`
    
    // API key (supports env var substitution)
    APIKey string `yaml:"api_key" json:"api_key"`
    
    // API endpoint URL
    APIEndpoint string `yaml:"api_endpoint" json:"api_endpoint"`
    
    // Default model
    DefaultModel string `yaml:"default_model" json:"default_model"`
    
    // Available models
    Models []ModelConfig `yaml:"models" json:"models"`
    
    // Rate limits
    RateLimit *RateLimitConfig `yaml:"rate_limit,omitempty" json:"rate_limit,omitempty"`
    
    // Timeout in seconds
    Timeout int `yaml:"timeout,omitempty" json:"timeout,omitempty"`
    
    // Retry configuration
    Retry *RetryConfig `yaml:"retry,omitempty" json:"retry,omitempty"`
}
```

### ModelConfig

```go
type ModelConfig struct {
    // Model identifier
    Name string `yaml:"name" json:"name"`
    
    // Maximum context tokens
    MaxTokens int `yaml:"max_tokens" json:"max_tokens"`
    
    // Whether model supports tools
    SupportsTools bool `yaml:"supports_tools" json:"supports_tools"`
    
    // Whether model supports streaming
    SupportsStreaming bool `yaml:"supports_streaming" json:"supports_streaming"`
    
    // Cost per 1K input tokens (USD)
    InputCost float64 `yaml:"input_cost,omitempty" json:"input_cost,omitempty"`
    
    // Cost per 1K output tokens (USD)
    OutputCost float64 `yaml:"output_cost,omitempty" json:"output_cost,omitempty"`
}
```

---

## Configuration Schema

### ApplicationConfig

**Location:** `internal/infrastructure/config/types.go`

```go
type ApplicationConfig struct {
    // Configuration version
    Version string `yaml:"version" json:"version"`
    
    // AI configuration (enhanced format)
    AI *AIConfig `yaml:"ai,omitempty" json:"ai,omitempty"`
    
    // MCP configuration
    MCP *MCPConfig `yaml:"mcp,omitempty" json:"mcp,omitempty"`
    
    // Template configuration
    Templates *TemplatesConfig `yaml:"templates,omitempty" json:"templates,omitempty"`
    
    // Logging configuration
    Logging *LoggingConfig `yaml:"logging,omitempty" json:"logging,omitempty"`
    
    // Legacy providers (backward compatibility)
    Providers []ProviderConfig `yaml:"providers,omitempty" json:"providers,omitempty"`
}
```

### AIConfig

```go
type AIConfig struct {
    // Default provider to use
    DefaultProvider string `yaml:"default_provider" json:"default_provider"`
    
    // Interface-based provider organization
    Interfaces map[string]*InterfaceConfig `yaml:"interfaces" json:"interfaces"`
}

type InterfaceConfig struct {
    // Providers implementing this interface
    Providers map[string]*ProviderConfig `yaml:"providers" json:"providers"`
}
```

**Example:**

```yaml
ai:
  default_provider: anthropic
  
  interfaces:
    openai_compatible:
      providers:
        openai:
          api_key: ${OPENAI_API_KEY}
          default_model: gpt-4o
        
        deepseek:
          api_key: ${DEEPSEEK_API_KEY}
          default_model: deepseek-chat
    
    anthropic_native:
      providers:
        anthropic:
          api_key: ${ANTHROPIC_API_KEY}
          default_model: claude-sonnet-4
```

---

### MCPConfig

```go
type MCPConfig struct {
    // MCP server configurations
    Servers map[string]*ServerConfig `yaml:"servers" json:"servers"`
    
    // Default timeout for server operations
    DefaultTimeout int `yaml:"default_timeout,omitempty" json:"default_timeout,omitempty"`
}

type ServerConfig struct {
    // Server name
    Name string `yaml:"name" json:"name"`
    
    // Command to execute
    Command string `yaml:"command" json:"command"`
    
    // Command arguments
    Args []string `yaml:"args,omitempty" json:"args,omitempty"`
    
    // Environment variables
    Env map[string]string `yaml:"env,omitempty" json:"env,omitempty"`
    
    // Working directory
    WorkingDir string `yaml:"working_dir,omitempty" json:"working_dir,omitempty"`
    
    // Auto-restart on failure
    AutoRestart bool `yaml:"auto_restart,omitempty" json:"auto_restart,omitempty"`
}
```

---

### WorkflowsConfig

```go
type TemplatesConfig struct {
    // Paths to search for templates
    Paths []string `yaml:"paths" json:"paths"`
    
    // Default template settings
    Defaults *TemplateDefaults `yaml:"defaults,omitempty" json:"defaults,omitempty"`
}

type TemplateDefaults struct {
    // Default provider
    Provider string `yaml:"provider,omitempty" json:"provider,omitempty"`
    
    // Default model
    Model string `yaml:"model,omitempty" json:"model,omitempty"`
    
    // Default temperature
    Temperature float64 `yaml:"temperature,omitempty" json:"temperature,omitempty"`
    
    // Default max tokens
    MaxTokens int `yaml:"max_tokens,omitempty" json:"max_tokens,omitempty"`
}
```

---

## MCP Protocol Types

### JSON-RPC Message

```go
type JSONRPCMessage struct {
    // Protocol version (always "2.0")
    JSONRPC string `json:"jsonrpc"`
    
    // Request ID (for request-response matching)
    ID interface{} `json:"id,omitempty"`
    
    // Method name (for requests)
    Method string `json:"method,omitempty"`
    
    // Parameters (for requests)
    Params interface{} `json:"params,omitempty"`
    
    // Result (for responses)
    Result interface{} `json:"result,omitempty"`
    
    // Error (for error responses)
    Error *JSONRPCError `json:"error,omitempty"`
}

type JSONRPCError struct {
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}
```

### Initialize Request/Response

```go
type InitializeRequest struct {
    ProtocolVersion string       `json:"protocolVersion"`
    Capabilities    Capabilities `json:"capabilities"`
    ClientInfo      ClientInfo   `json:"clientInfo"`
}

type InitializeResponse struct {
    ProtocolVersion string          `json:"protocolVersion"`
    Capabilities    ServerCapabilities `json:"capabilities"`
    ServerInfo      ServerInfo      `json:"serverInfo"`
}

type Capabilities struct {
    // Client capabilities
    Experimental map[string]interface{} `json:"experimental,omitempty"`
}

type ServerCapabilities struct {
    // Server supports tools
    Tools *ToolCapabilities `json:"tools,omitempty"`
    
    // Server supports resources
    Resources *ResourceCapabilities `json:"resources,omitempty"`
}

type ServerInfo struct {
    Name    string `json:"name"`
    Version string `json:"version"`
}
```

### Tool List Request/Response

```go
type ListToolsRequest struct {
    // Optional cursor for pagination
    Cursor string `json:"cursor,omitempty"`
}

type ListToolsResponse struct {
    Tools      []ToolDefinition `json:"tools"`
    NextCursor string           `json:"nextCursor,omitempty"`
}

type ToolDefinition struct {
    Name        string                 `json:"name"`
    Description string                 `json:"description"`
    InputSchema map[string]interface{} `json:"inputSchema"`
}
```

### Tool Call Request/Response

```go
type CallToolRequest struct {
    Name      string                 `json:"name"`
    Arguments map[string]interface{} `json:"arguments"`
}

type CallToolResponse struct {
    Content []ContentBlock `json:"content"`
    IsError bool           `json:"isError,omitempty"`
}

type ContentBlock struct {
    Type string `json:"type"` // "text", "image", "resource"
    
    // For text content
    Text string `json:"text,omitempty"`
    
    // For image content
    Data     string `json:"data,omitempty"`
    MimeType string `json:"mimeType,omitempty"`
}
```

---

## Workflow Schema

### Workflow Structure

```go
type Template struct {
    // Template name (must be unique)
    Name string `yaml:"name" json:"name"`
    
    // Human-readable description
    Description string `yaml:"description" json:"description"`
    
    // Semantic version
    Version string `yaml:"version" json:"version"`
    
    // Configuration
    Config *TemplateConfig `yaml:"config,omitempty" json:"config,omitempty"`
    
    // Execution steps
    Steps []TemplateStep `yaml:"steps" json:"steps"`
}
```

### WorkflowStep

```go
type TemplateStep struct {
    // Step name (must be unique within template)
    Name string `yaml:"name" json:"name"`
    
    // Prompt template (supports {{variable}} substitution)
    Prompt string `yaml:"prompt,omitempty" json:"prompt,omitempty"`
    
    // Provider override
    Provider string `yaml:"provider,omitempty" json:"provider,omitempty"`
    
    // Model override
    Model string `yaml:"model,omitempty" json:"model,omitempty"`
    
    // Output variable name
    Output string `yaml:"output,omitempty" json:"output,omitempty"`
    
    // Condition (skip if false)
    Condition string `yaml:"condition,omitempty" json:"condition,omitempty"`
    
    // Sub-template to call
    Template string `yaml:"template,omitempty" json:"template,omitempty"`
    
    // Input data for sub-template
    TemplateInput string `yaml:"template_input,omitempty" json:"template_input,omitempty"`
    
    // MCP servers to use
    Servers []string `yaml:"servers,omitempty" json:"servers,omitempty"`
    
    // Loop over array
    ForEach string `yaml:"for_each,omitempty" json:"for_each,omitempty"`
}
```

**YAML Example:**

```yaml
name: code_review
description: Multi-step code review
version: 1.0.0

config:
  defaults:
    provider: openai
    model: gpt-4o

steps:
  - name: analyze
    prompt: |
      Analyze this code:
      {{input_data.code}}
    output: analysis
  
  - name: suggest
    condition: "{{analysis}} != ''"
    prompt: |
      Based on this analysis:
      {{analysis}}
      
      Suggest improvements.
    output: suggestions
```

---

## Extension Points

### Adding a New Provider

**Required Steps:**

1. **Implement LLMProvider Interface:**
```go
type MyProvider struct {
    config *ProviderConfig
}

func (p *MyProvider) CreateCompletion(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
    // Implementation
}
// ... other interface methods
```

2. **Register in Factory:**
```go
// In providers/ai/factory.go
func (f *ProviderFactory) createProvider(providerType ProviderType) (LLMProvider, error) {
    switch providerType {
    case "myprovider":
        return f.createMyProvider()
    // ... other cases
    }
}
```

3. **Add Configuration Schema:**
```yaml
# config/providers/myprovider.yaml
provider_name: myprovider
api_key: ${MYPROVIDER_API_KEY}
api_endpoint: https://api.myprovider.com/v1
default_model: my-model-1
```

4. **Add Provider Type Constant:**
```go
const (
    ProviderMyProvider ProviderType = "myprovider"
)
```

---

### Adding a New MCP Transport

**Required Steps:**

1. **Implement Transport Interface:**
```go
type Transport interface {
    Send(ctx context.Context, message *JSONRPCMessage) error
    Receive(ctx context.Context) (*JSONRPCMessage, error)
    Close() error
}
```

2. **Create Transport Implementation:**
```go
// providers/mcp/transport/mytransport/transport.go
type MyTransport struct {
    // Transport-specific fields
}

func (t *MyTransport) Send(ctx context.Context, msg *JSONRPCMessage) error {
    // Implementation
}
```

3. **Register in Transport Factory:**
```go
func NewTransport(transportType string, config *TransportConfig) (Transport, error) {
    switch transportType {
    case "mytransport":
        return NewMyTransport(config)
    // ... other cases
    }
}
```

---

### Adding a New Template Function

**Built-in Functions:**

Templates support these built-in functions in variable substitution:

```
{{variable}}              - Simple substitution
{{variable.field}}        - Nested field access
{{variable[0]}}           - Array index access
{% if condition %}...{% endif %}  - Conditional
{% for item in array %}...{% endfor %}  - Loop
```

**Adding Custom Functions:**

Extend the template engine in `services/template/engine.go`:

```go
func (e *Engine) evaluateFunction(funcName string, args []interface{}) (interface{}, error) {
    switch funcName {
    case "upper":
        return strings.ToUpper(args[0].(string)), nil
    case "lower":
        return strings.ToLower(args[0].(string)), nil
    case "myfunction":
        return e.myCustomFunction(args)
    }
}
```

---

**API documentation complete!** All interfaces, types, and contracts are now documented. 🎯
