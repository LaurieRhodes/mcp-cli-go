# Component Architecture

Deep dive into each architectural component, their responsibilities, internal structure, and interactions.

---

## Table of Contents

- [Command Layer](#command-layer)
- [Service Layer](#service-layer)
- [Core Layer](#core-layer)
- [Provider Layer](#provider-layer)
- [Infrastructure Layer](#infrastructure-layer)
- [Domain Layer](#domain-layer)
- [Presentation Layer](#presentation-layer)

---

## Command Layer

**Location:** `cmd/`

**Responsibility:** CLI interface, flag parsing, command routing

### Structure

```
cmd/
├── root.go              # Root command, global flags
├── chat.go              # Chat mode command
├── query.go             # Query mode command
├── interactive.go       # Interactive mode command
├── embeddings.go        # Embeddings command
└── serve.go             # Server mode command
```

### Root Command (root.go)

**Purpose:** Global command setup, shared flags, initialization

**Key Components:**
```go
var rootCmd = &cobra.Command{
    Use:   "mcp-cli",
    Short: "Model Context Protocol CLI",
    Long:  "...",
}

// Global flags
var (
    configFile    string
    provider      string
    model         string
    verbose       bool
    noisy         bool
    templateName  string
)

func init() {
    cobra.OnInitialize(initConfig)
    
    rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "config file path")
    rootCmd.PersistentFlags().StringVar(&provider, "provider", "", "AI provider")
    rootCmd.PersistentFlags().StringVar(&model, "model", "", "Model name")
    rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "Verbose output")
    // ...
}
```

**Responsibilities:**
- Initialize logging
- Load configuration
- Set up provider factory
- Register subcommands

---

### Chat Command (chat.go)

**Purpose:** Interactive conversation mode

**Flag Structure:**
```go
type ChatFlags struct {
    Provider      string
    Model         string
    Servers       []string
    SystemPrompt  string
    MaxTokens     int
    Temperature   float64
}
```

**Execution Flow:**
```go
func runChat(cmd *cobra.Command, args []string) error {
    // 1. Parse flags
    flags := parseChatFlags(cmd)
    
    // 2. Load configuration
    config, err := loadConfig(flags.ConfigFile)
    
    // 3. Create service
    chatService := services.NewChatService(config)
    
    // 4. Start chat session
    return chatService.Start(context.Background(), flags)
}
```

**Features:**
- Streaming response display
- Command handling (/help, /clear, etc.)
- Context management
- Tool execution
- History tracking

---

### Query Command (query.go)

**Purpose:** Single-shot query execution

**Flag Structure:**
```go
type QueryFlags struct {
    Provider     string
    Model        string
    Servers      []string
    JSON         bool
    OutputFile   string
    InputFile    string
    MaxTokens    int
}
```

**Execution Flow:**
```go
func runQuery(cmd *cobra.Command, args []string) error {
    // 1. Get query text (args, stdin, or file)
    query := getQueryText(args, flags)
    
    // 2. Create service
    queryService := services.NewQueryService(config)
    
    // 3. Execute query
    result, err := queryService.Execute(context.Background(), query, flags)
    
    // 4. Format and output
    return formatAndOutput(result, flags)
}
```

**Input Sources:**
1. Command arguments: `mcp-cli query "text"`
2. Stdin: `echo "text" | mcp-cli query`
3. File: `mcp-cli query --input-file file.txt`

**Output Formats:**
- Text (default)
- JSON (`--json`)
- File (`--output-file`)

---

### Interactive Command (interactive.go)

**Purpose:** Direct MCP tool interaction

**Command Loop:**
```go
func runInteractive(cmd *cobra.Command, args []string) error {
    service := services.NewInteractiveService(config)
    
    // Start REPL
    for {
        input := readUserInput()
        
        if isCommand(input) {
            handleCommand(input, service)
        } else {
            fmt.Println("Use /help for commands")
        }
        
        if shouldExit(input) {
            break
        }
    }
    
    return nil
}
```

**Commands:**
- `/help` - Show help
- `/tools` - List tools
- `/call <server> <tool> <args>` - Call tool
- `/exit` - Exit

---

## Service Layer

**Location:** `services/`

**Responsibility:** Business logic orchestration, workflow coordination

### Service Architecture

```
services/
├── chat/
│   ├── service.go          # Chat service implementation
│   ├── options.go          # Chat options/config
│   └── session.go          # Session management
│
├── query/
│   ├── service.go          # Query service implementation
│   └── options.go          # Query options
│
├── interactive/
│   ├── service.go          # Interactive service
│   └── commands.go         # Command handlers
│
└── template/
    ├── executor.go         # Template executor
    ├── parser.go           # Template parser
    └── validator.go        # Template validator
```

---

### Chat Service

**Interface:**
```go
type ChatService interface {
    Start(ctx context.Context, options ChatOptions) error
    SendMessage(ctx context.Context, message string) (*Response, error)
    GetHistory() []Message
    Clear() error
}
```

**Implementation:**
```go
type chatService struct {
    config      *config.Config
    provider    domain.LLMProvider
    mcpManager  *mcp.Manager
    chatCore    *core.ChatManager
    logger      logging.Logger
}

func (s *chatService) Start(ctx context.Context, opts ChatOptions) error {
    // 1. Initialize provider
    s.provider, err = s.createProvider(opts.Provider, opts.Model)
    
    // 2. Connect MCP servers
    s.mcpManager.ConnectServers(opts.Servers)
    
    // 3. Start chat core
    return s.chatCore.Start(ctx, s.provider, s.mcpManager)
}
```

**Key Responsibilities:**
- Provider initialization and selection
- MCP server connection management
- Chat context coordination
- Command routing
- Error handling and recovery

---

### Query Service

**Interface:**
```go
type QueryService interface {
    Execute(ctx context.Context, query string, options QueryOptions) (*QueryResult, error)
    ExecuteWithTemplate(ctx context.Context, template string, data map[string]interface{}) (*QueryResult, error)
}
```

**Implementation:**
```go
type queryService struct {
    config     *config.Config
    factory    *ai.ProviderFactory
    mcpManager *mcp.Manager
    logger     logging.Logger
}

func (s *queryService) Execute(ctx context.Context, query string, opts QueryOptions) (*QueryResult, error) {
    // 1. Select provider
    provider, err := s.selectProvider(opts)
    
    // 2. Prepare request
    request := s.buildRequest(query, opts)
    
    // 3. Execute with retry
    response, err := s.executeWithRetry(ctx, provider, request)
    
    // 4. Process response
    return s.processResponse(response, opts)
}
```

**Retry Logic:**
```go
func (s *queryService) executeWithRetry(ctx context.Context, provider domain.LLMProvider, request *domain.CompletionRequest) (*domain.CompletionResponse, error) {
    retryConfig := RetryConfig{
        MaxAttempts:  3,
        InitialDelay: 1 * time.Second,
        MaxDelay:     10 * time.Second,
        Multiplier:   2.0,
    }
    
    return retryWithBackoff(func() (*domain.CompletionResponse, error) {
        return provider.CreateCompletion(ctx, request)
    }, retryConfig)
}
```

---

### Template Executor

**Purpose:** Execute multi-step template workflows

**Structure:**
```go
type TemplateExecutor struct {
    template    *Template
    provider    domain.LLMProvider
    mcpManager  *mcp.Manager
    variables   map[string]interface{}
}

type Template struct {
    Name        string
    Description string
    Version     string
    Config      TemplateConfig
    Steps       []TemplateStep
}

type TemplateStep struct {
    Name       string
    Provider   string  // Optional override
    Model      string  // Optional override
    Prompt     string
    Output     string  // Variable name for output
    Condition  string  // Optional conditional
    Template   string  // Sub-template to call
    Servers    []string // MCP servers to use
}
```

**Execution Flow:**
```go
func (e *TemplateExecutor) Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
    e.variables = input
    
    for _, step := range e.template.Steps {
        // 1. Evaluate condition
        if !e.evaluateCondition(step.Condition) {
            continue
        }
        
        // 2. Substitute variables in prompt
        prompt := e.substituteVariables(step.Prompt)
        
        // 3. Select provider (step override or template default)
        provider := e.selectProvider(step.Provider)
        
        // 4. Execute step
        result, err := e.executeStep(ctx, provider, prompt, step)
        
        // 5. Store output
        if step.Output != "" {
            e.variables[step.Output] = result
        }
    }
    
    return e.variables, nil
}
```

**Variable Substitution:**
```go
func (e *TemplateExecutor) substituteVariables(template string) string {
    // Replace {{variable}} with actual values
    for key, value := range e.variables {
        placeholder := fmt.Sprintf("{{%s}}", key)
        template = strings.ReplaceAll(template, placeholder, fmt.Sprint(value))
    }
    return template
}
```

---

## Core Layer

**Location:** `core/`

**Responsibility:** Mode-specific business logic implementation

### Chat Core

**Location:** `core/chat/`

**Components:**
```
core/chat/
├── manager.go         # Chat workflow manager
├── context.go         # Conversation context
├── commands.go        # Command handlers
├── tools.go           # Tool execution
└── ui.go              # User interface
```

**Chat Manager:**
```go
type ChatManager struct {
    context     *ChatContext
    provider    domain.LLMProvider
    mcpManager  *mcp.Manager
    ui          *ChatUI
    
    streaming   bool
    maxTokens   int
    temperature float64
}

func (m *ChatManager) ProcessMessage(ctx context.Context, userMessage string) error {
    // 1. Add user message to context
    m.context.AddMessage(domain.Message{
        Role:    "user",
        Content: userMessage,
    })
    
    // 2. Build request with context
    request := m.buildRequest()
    
    // 3. Stream response
    response, err := m.streamResponse(ctx, request)
    
    // 4. Handle tool calls if present
    if response.HasToolCalls() {
        toolResults := m.executeTools(ctx, response.ToolCalls)
        return m.continueWithToolResults(ctx, toolResults)
    }
    
    // 5. Add assistant response to context
    m.context.AddMessage(domain.Message{
        Role:    "assistant",
        Content: response.Content,
    })
    
    return nil
}
```

**Context Management:**
```go
type ChatContext struct {
    mu           sync.RWMutex
    messages     []domain.Message
    systemPrompt string
    maxMessages  int
    tokenBudget  int
}

func (c *ChatContext) AddMessage(msg domain.Message) {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    c.messages = append(c.messages, msg)
    
    // Trim if exceeds limits
    if len(c.messages) > c.maxMessages {
        c.trimMessages()
    }
}

func (c *ChatContext) trimMessages() {
    // Keep system prompt and recent messages
    systemMsg := c.messages[0]
    recentMessages := c.messages[len(c.messages)-c.maxMessages+1:]
    c.messages = append([]domain.Message{systemMsg}, recentMessages...)
}
```

**Tool Execution:**
```go
func (m *ChatManager) executeTools(ctx context.Context, toolCalls []domain.ToolCall) []domain.ToolResult {
    results := make([]domain.ToolResult, len(toolCalls))
    
    // Execute tools in parallel
    var wg sync.WaitGroup
    for i, toolCall := range toolCalls {
        wg.Add(1)
        go func(idx int, call domain.ToolCall) {
            defer wg.Done()
            
            result, err := m.mcpManager.ExecuteTool(ctx, call.ServerName, call.ToolName, call.Arguments)
            if err != nil {
                results[idx] = domain.ToolResult{Error: err.Error()}
            } else {
                results[idx] = domain.ToolResult{Content: result}
            }
        }(i, toolCall)
    }
    
    wg.Wait()
    return results
}
```

---

### Query Handler

**Location:** `core/query/`

**Structure:**
```go
type QueryHandler struct {
    provider   domain.LLMProvider
    mcpManager *mcp.Manager
    formatter  *Formatter
}

func (h *QueryHandler) HandleQuery(ctx context.Context, query string, opts QueryOptions) (*QueryResult, error) {
    // 1. Build completion request
    request := &domain.CompletionRequest{
        Messages: []domain.Message{
            {Role: "user", Content: query},
        },
        MaxTokens:   opts.MaxTokens,
        Temperature: opts.Temperature,
    }
    
    // 2. Add tools if servers specified
    if len(opts.Servers) > 0 {
        request.Tools = h.mcpManager.GetAvailableTools(opts.Servers)
    }
    
    // 3. Execute request
    response, err := h.provider.CreateCompletion(ctx, request)
    if err != nil {
        return nil, err
    }
    
    // 4. Handle tool calls
    if response.HasToolCalls() {
        response, err = h.handleToolCalls(ctx, request, response)
    }
    
    // 5. Format result
    return h.formatter.FormatResult(response, opts.Format)
}
```

---

## Provider Layer

**Location:** `providers/`

**Responsibility:** External system integration (AI providers, MCP servers)

### AI Provider Structure

```
providers/ai/
├── factory.go           # Provider factory
├── service.go           # AI service coordinator
├── streaming/
│   ├── processor.go    # Stream processing
│   └── sse.go          # Server-Sent Events
└── clients/
    ├── openai.go       # OpenAI client
    ├── anthropic.go    # Anthropic client
    ├── gemini.go       # Gemini client
    └── ollama.go       # Ollama client
```

---

### Provider Factory

**Implementation:**
```go
type ProviderFactory struct {
    config *config.Config
    logger logging.Logger
}

func (f *ProviderFactory) CreateProvider(providerType ProviderType, modelOverride string) (domain.LLMProvider, error) {
    // 1. Get provider configuration
    providerConfig, interfaceType, err := f.config.GetProviderConfig(string(providerType))
    if err != nil {
        return nil, err
    }
    
    // 2. Select implementation based on interface type
    switch interfaceType {
    case OpenAICompatible:
        return f.createOpenAICompatible(providerConfig, modelOverride)
    case AnthropicNative:
        return f.createAnthropicNative(providerConfig, modelOverride)
    case GeminiNative:
        return f.createGeminiNative(providerConfig, modelOverride)
    case OllamaNative:
        return f.createOllamaNative(providerConfig, modelOverride)
    default:
        return nil, fmt.Errorf("unsupported interface type: %s", interfaceType)
    }
}
```

**Provider Mapping:**
```go
var providerInterfaceMap = map[ProviderType]InterfaceType{
    ProviderOpenAI:     OpenAICompatible,
    ProviderDeepSeek:   OpenAICompatible,
    ProviderOpenRouter: OpenAICompatible,
    ProviderAnthropic:  AnthropicNative,
    ProviderGemini:     GeminiNative,
    ProviderOllama:     OllamaNative,
}
```

---

### OpenAI Client

**Structure:**
```go
type OpenAIClient struct {
    config      *ProviderConfig
    httpClient  *http.Client
    rateLimiter *rate.Limiter
    
    apiKey      string
    baseURL     string
    model       string
}

func (c *OpenAIClient) CreateCompletion(ctx context.Context, req *domain.CompletionRequest) (*domain.CompletionResponse, error) {
    // 1. Rate limiting
    if err := c.rateLimiter.Wait(ctx); err != nil {
        return nil, err
    }
    
    // 2. Build API request
    apiRequest := c.buildAPIRequest(req)
    
    // 3. Make HTTP request
    httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/chat/completions", apiRequest)
    httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
    httpReq.Header.Set("Content-Type", "application/json")
    
    resp, err := c.httpClient.Do(httpReq)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    // 4. Parse response
    var apiResponse OpenAIResponse
    if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
        return nil, err
    }
    
    // 5. Convert to domain response
    return c.convertToDomainResponse(apiResponse), nil
}
```

**Streaming:**
```go
func (c *OpenAIClient) StreamCompletion(ctx context.Context, req *domain.CompletionRequest, writer io.Writer) (*domain.CompletionResponse, error) {
    // Set stream flag
    apiRequest := c.buildAPIRequest(req)
    apiRequest.Stream = true
    
    // Make request
    httpReq, _ := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/chat/completions", apiRequest)
    resp, _ := c.httpClient.Do(httpReq)
    defer resp.Body.Close()
    
    // Process SSE stream
    scanner := bufio.NewScanner(resp.Body)
    var fullContent string
    
    for scanner.Scan() {
        line := scanner.Text()
        
        if !strings.HasPrefix(line, "data: ") {
            continue
        }
        
        data := strings.TrimPrefix(line, "data: ")
        if data == "[DONE]" {
            break
        }
        
        var chunk OpenAIStreamChunk
        json.Unmarshal([]byte(data), &chunk)
        
        if len(chunk.Choices) > 0 {
            content := chunk.Choices[0].Delta.Content
            writer.Write([]byte(content))
            fullContent += content
        }
    }
    
    return &domain.CompletionResponse{Content: fullContent}, nil
}
```

---

### Anthropic Client

**Native Implementation:**
```go
type AnthropicClient struct {
    config     *ProviderConfig
    httpClient *http.Client
    
    apiKey  string
    baseURL string
    model   string
}

func (c *AnthropicClient) CreateCompletion(ctx context.Context, req *domain.CompletionRequest) (*domain.CompletionResponse, error) {
    // Anthropic uses different message format
    apiRequest := c.buildAnthropicRequest(req)
    
    httpReq, _ := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/v1/messages", apiRequest)
    httpReq.Header.Set("x-api-key", c.apiKey)
    httpReq.Header.Set("anthropic-version", "2023-06-01")
    httpReq.Header.Set("Content-Type", "application/json")
    
    resp, _ := c.httpClient.Do(httpReq)
    defer resp.Body.Close()
    
    var apiResponse AnthropicResponse
    json.NewDecoder(resp.Body).Decode(&apiResponse)
    
    return c.convertToDomainResponse(apiResponse), nil
}
```

**Tool Use Handling:**
```go
func (c *AnthropicClient) buildAnthropicRequest(req *domain.CompletionRequest) *AnthropicRequest {
    apiReq := &AnthropicRequest{
        Model:      c.model,
        MaxTokens:  req.MaxTokens,
        Messages:   c.convertMessages(req.Messages),
    }
    
    // Convert tools to Anthropic format
    if len(req.Tools) > 0 {
        apiReq.Tools = c.convertTools(req.Tools)
    }
    
    return apiReq
}

func (c *AnthropicClient) convertTools(tools []domain.Tool) []AnthropicTool {
    anthropicTools := make([]AnthropicTool, len(tools))
    for i, tool := range tools {
        anthropicTools[i] = AnthropicTool{
            Name:        tool.Name,
            Description: tool.Description,
            InputSchema: tool.Parameters,
        }
    }
    return anthropicTools
}
```

---

### MCP Provider

**Location:** `providers/mcp/`

**Structure:**
```
providers/mcp/
├── manager.go           # MCP server manager
├── messages/
│   ├── initialize/     # Initialization messages
│   ├── tools/          # Tool-related messages
│   └── json_rpc_message.go
└── transport/
    └── stdio/          # Standard I/O transport
```

**MCP Manager:**
```go
type Manager struct {
    servers map[string]*ServerConnection
    mu      sync.RWMutex
}

type ServerConnection struct {
    Name      string
    Process   *exec.Cmd
    Transport *stdio.Transport
    Tools     []domain.Tool
    State     ServerState
}

func (m *Manager) ConnectServer(ctx context.Context, serverConfig *ServerConfig) error {
    // 1. Start server process
    cmd := exec.CommandContext(ctx, serverConfig.Command, serverConfig.Args...)
    
    stdin, _ := cmd.StdinPipe()
    stdout, _ := cmd.StdoutPipe()
    stderr, _ := cmd.StderrPipe()
    
    cmd.Start()
    
    // 2. Create transport
    transport := stdio.NewTransport(stdin, stdout)
    
    // 3. Send initialize message
    initResponse, err := m.initialize(transport, serverConfig)
    
    // 4. List tools
    tools, err := m.listTools(transport)
    
    // 5. Store connection
    m.mu.Lock()
    m.servers[serverConfig.Name] = &ServerConnection{
        Name:      serverConfig.Name,
        Process:   cmd,
        Transport: transport,
        Tools:     tools,
        State:     StateConnected,
    }
    m.mu.Unlock()
    
    return nil
}
```

**Tool Execution:**
```go
func (m *Manager) ExecuteTool(ctx context.Context, serverName, toolName string, arguments map[string]interface{}) (interface{}, error) {
    m.mu.RLock()
    server := m.servers[serverName]
    m.mu.RUnlock()
    
    if server == nil {
        return nil, fmt.Errorf("server not found: %s", serverName)
    }
    
    // Build JSON-RPC request
    request := &messages.ToolCallRequest{
        JSONRPC: "2.0",
        ID:      generateID(),
        Method:  "tools/call",
        Params: messages.ToolCallParams{
            Name:      toolName,
            Arguments: arguments,
        },
    }
    
    // Send request
    response, err := server.Transport.SendRequest(ctx, request)
    if err != nil {
        return nil, err
    }
    
    return response.Result, nil
}
```

---

## Infrastructure Layer

**Location:** `infrastructure/`

**Components:**
```
infrastructure/
├── config/
│   ├── service.go      # Configuration service
│   ├── enhanced.go     # Modern config format
│   ├── legacy.go       # Legacy compatibility
│   └── validation.go   # Config validation
├── logging/
│   ├── logger.go       # Logger implementation
│   ├── levels.go       # Log levels
│   └── production.go   # Production config
└── host/
    ├── server_manager.go    # MCP server lifecycle
    ├── environment.go       # Environment handling
    └── ai_options.go        # AI provider options
```

### Configuration Service

```go
type ConfigurationService interface {
    LoadConfig(path string) (*ApplicationConfig, error)
    GetProviderConfig(provider string) (*ProviderConfig, InterfaceType, error)
    GetServerConfig(server string) (*ServerConfig, error)
    GetTemplateConfig(template string) (*TemplateConfig, error)
}

type configService struct {
    configPath string
    config     *ApplicationConfig
    cache      map[string]interface{}
    mu         sync.RWMutex
}
```

### Logger

```go
type Logger interface {
    Debug(msg string, fields ...Field)
    Info(msg string, fields ...Field)
    Warn(msg string, fields ...Field)
    Error(msg string, fields ...Field)
    Fatal(msg string, fields ...Field)
}

type Field struct {
    Key   string
    Value interface{}
}
```

---

## Domain Layer

**Location:** `domain/`

**Core Types:**
```go
// Provider interface
type LLMProvider interface {
    CreateCompletion(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error)
    StreamCompletion(ctx context.Context, req *CompletionRequest, writer io.Writer) (*CompletionResponse, error)
    GetProviderType() ProviderType
    GetInterfaceType() InterfaceType
    ValidateConfig() error
    Close() error
}

// Request/Response types
type CompletionRequest struct {
    Messages    []Message
    MaxTokens   int
    Temperature float64
    Tools       []Tool
    ToolChoice  string
}

type CompletionResponse struct {
    ID        string
    Content   string
    ToolCalls []ToolCall
    Usage     Usage
}

// Tool types
type Tool struct {
    Name        string
    Description string
    Parameters  map[string]interface{}
}

type ToolCall struct {
    ID         string
    ServerName string
    ToolName   string
    Arguments  map[string]interface{}
}
```

---

**Component architecture complete!** Continue to [Data Flow](data-flow.md) for request/response flows. 🔧
