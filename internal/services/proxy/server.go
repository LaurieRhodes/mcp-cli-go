package proxy

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/config"
	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/runas"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/host"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/logging"
	"github.com/LaurieRhodes/mcp-cli-go/internal/providers/mcp/messages/initialize"
	"github.com/LaurieRhodes/mcp-cli-go/internal/providers/mcp/messages/tools"
	"github.com/LaurieRhodes/mcp-cli-go/internal/providers/mcp/transport/stdio"
	"github.com/LaurieRhodes/mcp-cli-go/internal/services/skills"
	"github.com/LaurieRhodes/mcp-cli-go/internal/services/workflow"
)

// ProxyServer is an HTTP server that exposes MCP tools as REST endpoints
type ProxyServer struct {
	config          *runas.RunAsConfig
	appConfig       *config.ApplicationConfig
	workflowService *workflow.Orchestrator
	skillsService   *skills.Service
	mcpServers      []*host.ServerConnection
	httpServer      *http.Server
	toolHandlers    map[string]*ToolHandler
	openAPISpec     *OpenAPISpec
}

// NewServer creates a new HTTP proxy server
// This matches the expected signature from serve.go
func NewServer(runasConfig *runas.RunAsConfig, appConfig *config.ApplicationConfig, workflowSvc *workflow.Orchestrator, skillsSvc *skills.Service) *ProxyServer {
	return &ProxyServer{
		config:          runasConfig,
		appConfig:       appConfig,
		workflowService: workflowSvc,
		skillsService:   skillsSvc,
		mcpServers:      []*host.ServerConnection{},
		toolHandlers:    make(map[string]*ToolHandler),
	}
}

// NewProxyServer creates a new HTTP proxy server (alternative constructor)
func NewProxyServer(runasConfig *runas.RunAsConfig, appConfig *config.ApplicationConfig) *ProxyServer {
	return NewServer(runasConfig, appConfig, nil, nil)
}

// SetMCPServers sets the MCP server connections
func (s *ProxyServer) SetMCPServers(servers []*host.ServerConnection) {
	s.mcpServers = servers
}

// Start starts the HTTP server
func (s *ProxyServer) Start() error {
	// Validate proxy config
	if s.config.ProxyConfig == nil {
		return fmt.Errorf("proxy_config is required for proxy server types")
	}
	
	proxyConfig := s.config.ProxyConfig
	
	// Set defaults
	port := proxyConfig.Port
	if port == 0 {
		port = 8080
	}
	
	host := proxyConfig.Host
	if host == "" {
		host = "0.0.0.0"
	}
	
	// Validate API key
	if proxyConfig.APIKey == "" {
		return fmt.Errorf("api_key is required in proxy_config for security")
	}
	
	// If config_source is specified, connect to source MCP server and discover tools
	if s.config.ConfigSource != "" {
		logging.Info("Discovering tools from source MCP server: %s", s.config.ConfigSource)
		if err := s.discoverToolsFromSource(); err != nil {
			return fmt.Errorf("failed to discover tools from source: %w", err)
		}
	}
	
	// Initialize tool handlers
	if err := s.initializeToolHandlers(); err != nil {
		return fmt.Errorf("failed to initialize tool handlers: %w", err)
	}
	
	// Generate OpenAPI spec
	s.openAPISpec = s.generateOpenAPISpec()
	
	// Create HTTP mux
	mux := http.NewServeMux()
	
	// Register routes
	s.registerRoutes(mux)
	
	// Apply middleware
	handler := s.applyMiddleware(mux)
	
	// Create HTTP server
	addr := fmt.Sprintf("%s:%d", host, port)
	s.httpServer = &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	
	// Start server in goroutine
	errChan := make(chan error, 1)
	go func() {
		logging.Info("Starting HTTP proxy server on %s", addr)
		logging.Info("OpenAPI docs available at: http://%s/docs", addr)
		logging.Info("Server: %s v%s", s.config.ServerInfo.Name, s.config.ServerInfo.Version)
		logging.Info("Exposed tools: %d", len(s.config.Tools))
		
		if proxyConfig.TLS != nil && proxyConfig.TLS.CertFile != "" && proxyConfig.TLS.KeyFile != "" {
			logging.Info("TLS enabled")
			errChan <- s.httpServer.ListenAndServeTLS(proxyConfig.TLS.CertFile, proxyConfig.TLS.KeyFile)
		} else {
			errChan <- s.httpServer.ListenAndServe()
		}
	}()
	
	// Wait for interrupt signal or server error
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	
	select {
	case err := <-errChan:
		if err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("server error: %w", err)
		}
	case sig := <-sigChan:
		logging.Info("Received signal %v, shutting down...", sig)
		return s.Shutdown()
	}
	
	return nil
}

// Shutdown gracefully shuts down the server
func (s *ProxyServer) Shutdown() error {
	if s.httpServer == nil {
		return nil
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	logging.Info("Shutting down HTTP proxy server...")
	return s.httpServer.Shutdown(ctx)
}

// registerRoutes registers all HTTP routes
func (s *ProxyServer) registerRoutes(mux *http.ServeMux) {
	basePath := s.config.ProxyConfig.BasePath
	
	// Health check endpoint
	mux.HandleFunc(basePath+"/health", s.handleHealth)
	
	// OpenAPI spec endpoint
	mux.HandleFunc(basePath+"/openapi.json", s.handleOpenAPISpec)
	
	// OpenAPI docs endpoint (if enabled)
	if s.config.ProxyConfig.EnableDocs {
		mux.HandleFunc(basePath+"/docs", s.handleDocs)
	}
	
	// Tool endpoints - one per tool (at root level to match mcpo behavior)
	for toolName, handler := range s.toolHandlers {
		path := fmt.Sprintf("%s/%s", basePath, toolName)
		mux.HandleFunc(path, s.createToolEndpoint(handler))
		logging.Debug("Registered tool endpoint: POST %s", path)
	}
	
	// List tools endpoint
	mux.HandleFunc(basePath+"/tools", s.handleListTools)
}

// applyMiddleware applies middleware to the handler chain
func (s *ProxyServer) applyMiddleware(handler http.Handler) http.Handler {
	// Apply middleware in reverse order (last applied is executed first)
	
	// CORS middleware
	handler = corsMiddleware(s.config.ProxyConfig.CORSOrigins)(handler)
	
	// API key authentication middleware
	handler = apiKeyMiddleware(s.config.ProxyConfig.APIKey)(handler)
	
	// Logging middleware
	handler = loggingMiddleware(handler)
	
	return handler
}

// handleHealth handles health check requests
func (s *ProxyServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "healthy",
		"server":  s.config.ServerInfo.Name,
		"version": s.config.ServerInfo.Version,
		"tools":   len(s.config.Tools),
	})
}

// handleListTools handles tool listing requests
func (s *ProxyServer) handleListTools(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	tools := make([]map[string]interface{}, 0, len(s.config.Tools))
	for _, tool := range s.config.Tools {
		tools = append(tools, map[string]interface{}{
			"name":        tool.Name,
			"description": tool.Description,
			"template":    tool.Template,
			"endpoint":    fmt.Sprintf("/%s", tool.Name),
		})
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"tools": tools,
		"count": len(tools),
	})
}

// handleOpenAPISpec serves the OpenAPI specification
func (s *ProxyServer) handleOpenAPISpec(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.openAPISpec)
}

// handleDocs serves the OpenAPI documentation UI
func (s *ProxyServer) handleDocs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// Serve Swagger UI HTML
	html := s.generateSwaggerUIHTML()
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

// discoverToolsFromSource connects to source MCP server and discovers available tools
func (s *ProxyServer) discoverToolsFromSource() error {
	// Extract server name from config_source path
	// e.g., "config/servers/filesystem.yaml" → "filesystem"
	serverName := strings.TrimSuffix(filepath.Base(s.config.ConfigSource), filepath.Ext(s.config.ConfigSource))
	
	logging.Info("Discovering tools from MCP server: %s", serverName)
	
	// Get server config from appConfig
	serverConfig, exists := s.appConfig.Servers[serverName]
	if !exists {
		return fmt.Errorf("server %s not found in application config", serverName)
	}
	
	// Create infrastructure server config
	infraServerConfig := config.ServerConfig{
		Command: serverConfig.Command,
		Args:    serverConfig.Args,
		Env:     serverConfig.Env,
	}
	
	// Create and start stdio client
	params := stdio.StdioServerParameters{
		Command: infraServerConfig.Command,
		Args:    infraServerConfig.Args,
		Env:     infraServerConfig.Env,
	}
	client := stdio.NewStdioClient(params)
	
	if err := client.Start(); err != nil {
		return fmt.Errorf("failed to start MCP server %s: %w", serverName, err)
	}
	
	// Send initialize request
	initResult, err := initialize.SendInitialize(client)
	if err != nil {
		client.Stop()
		return fmt.Errorf("failed to initialize MCP server %s: %w", serverName, err)
	}
	
	logging.Info("Connected to %s v%s", initResult.ServerInfo.Name, initResult.ServerInfo.Version)
	
	// Create server connection
	conn := &host.ServerConnection{
		Name:         serverName,
		Client:       client,
		ServerInfo:   initResult.ServerInfo,
		Capabilities: initResult.Capabilities,
	}
	
	// Store connection
	s.mcpServers = append(s.mcpServers, conn)
	
	// Auto-populate server_info if not set in proxy config
	if s.config.ServerInfo.Name == "" {
		s.config.ServerInfo.Name = fmt.Sprintf("%s-proxy", initResult.ServerInfo.Name)
		s.config.ServerInfo.Version = initResult.ServerInfo.Version
		s.config.ServerInfo.Description = fmt.Sprintf("HTTP proxy for %s", initResult.ServerInfo.Name)
		logging.Info("Auto-populated server_info from source: %s v%s", 
			s.config.ServerInfo.Name, s.config.ServerInfo.Version)
	}
	
	// Get tools from server using tools/list
	toolsResult, err := tools.SendToolsList(client, nil)
	if err != nil {
		return fmt.Errorf("failed to list tools from %s: %w", serverName, err)
	}
	
	logging.Info("Discovered %d tools from %s", len(toolsResult.Tools), serverName)
	
	// Convert MCP tools to RunAs tools
	s.config.Tools = make([]runas.ToolExposure, len(toolsResult.Tools))
	for i, tool := range toolsResult.Tools {
		s.config.Tools[i] = runas.ToolExposure{
			Name:        tool.Name,
			Description: tool.Description,
			InputSchema: tool.InputSchema,
			// Mark as MCP server tool
			MCPServer: serverName,
			MCPTool:   tool.Name,
			Template:  "", // No template for MCP tools
		}
		
		logging.Debug("Discovered tool: %s - %s", tool.Name, tool.Description)
	}
	
	return nil
}

// initializeToolHandlers creates handlers for all tools
func (s *ProxyServer) initializeToolHandlers() error {
	// Create schema generator with MCP server connections
	schemaGen := NewSchemaGenerator(s.appConfig)
	
	for i := range s.config.Tools {
		tool := &s.config.Tools[i]
		
		// Auto-generate schema and description if not provided
		if tool.InputSchema == nil || tool.Description == "" {
			schema, description, err := schemaGen.GenerateForTool(tool)
			if err != nil {
				logging.Warn("Failed to auto-generate schema for tool %s: %v, using generic schema", tool.Name, err)
				// Use generic schema as fallback
				schema = map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"input_data": map[string]interface{}{
							"type":        "string",
							"description": "Input data for the tool",
						},
					},
					"required": []interface{}{"input_data"},
				}
				description = fmt.Sprintf("Execute %s", tool.Name)
			}
			
			// Update tool with generated values
			if tool.InputSchema == nil {
				tool.InputSchema = schema
				logging.Debug("Auto-generated schema for tool %s", tool.Name)
			}
			if tool.Description == "" {
				tool.Description = description
				logging.Debug("Auto-generated description for tool %s: %s", tool.Name, description)
			}
		}
		
		// Create handler for this tool
		handler := NewToolHandler(tool, s); _ = handler
		
		s.toolHandlers[tool.Name] = handler
		logging.Debug("Initialized handler for tool: %s", tool.Name)
	}
	
	return nil
}

// createToolEndpoint creates an HTTP handler for a specific tool
func (s *ProxyServer) createToolEndpoint(handler *ToolHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		
		// Handle the tool execution
		handler.Handle(w, r)
	}
}
