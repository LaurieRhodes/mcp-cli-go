package skills

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/LaurieRhodes/mcp-cli-go/internal/domain"
	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/config"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/host"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/logging"
	"github.com/LaurieRhodes/mcp-cli-go/internal/providers/mcp/messages/tools"
)

// HostServerManager adapts host.ServerConnection to domain.MCPServerManager interface
type HostServerManager struct {
	connections []*host.ServerConnection
}

// NewHostServerManager creates a new host server manager
func NewHostServerManager(connections []*host.ServerConnection) *HostServerManager {
	return &HostServerManager{connections: connections}
}

func (hsm *HostServerManager) StartServer(ctx context.Context, serverName string, cfg *config.ServerConfig) (domain.MCPServer, error) {
	for _, conn := range hsm.connections {
		if conn.Name == serverName {
			return &HostServerAdapter{connection: conn}, nil
		}
	}
	return nil, fmt.Errorf("server '%s' not found in host connections", serverName)
}

func (hsm *HostServerManager) StopServer(serverName string) error {
	return nil
}

func (hsm *HostServerManager) GetServer(serverName string) (domain.MCPServer, bool) {
	for _, conn := range hsm.connections {
		if conn.Name == serverName {
			return &HostServerAdapter{connection: conn}, true
		}
	}
	return nil, false
}

func (hsm *HostServerManager) ListServers() map[string]domain.MCPServer {
	servers := make(map[string]domain.MCPServer)
	for _, conn := range hsm.connections {
		servers[conn.Name] = &HostServerAdapter{connection: conn}
	}
	return servers
}

func (hsm *HostServerManager) GetAvailableTools() ([]domain.Tool, error) {
	var toolsList []domain.Tool
	
	for _, conn := range hsm.connections {
		adapter := &HostServerAdapter{connection: conn}
		serverTools, err := adapter.GetTools()
		if err != nil {
			logging.Warn("Failed to get tools from server %s: %v", conn.Name, err)
			continue
		}
		toolsList = append(toolsList, serverTools...)
	}
	
	return toolsList, nil
}

func (hsm *HostServerManager) ExecuteTool(ctx context.Context, toolName string, arguments map[string]interface{}) (string, error) {
	for _, conn := range hsm.connections {
		adapter := &HostServerAdapter{connection: conn}
		toolsList, err := adapter.GetTools()
		if err != nil {
			continue
		}
		
		// Check both prefixed and unprefixed tool names
		serverPrefix := conn.Name + "_"
		serverPrefixUnderscore := strings.ReplaceAll(conn.Name, "-", "_") + "_"
		
		for _, tool := range toolsList {
			// Extract original tool name (strip server prefix if present)
			originalName := tool.Function.Name
			if strings.HasPrefix(originalName, serverPrefix) {
				originalName = strings.TrimPrefix(originalName, serverPrefix)
			} else if strings.HasPrefix(originalName, serverPrefixUnderscore) {
				originalName = strings.TrimPrefix(originalName, serverPrefixUnderscore)
			}
			
			// Match against both original name and prefixed name
			if tool.Function.Name == toolName || originalName == toolName {
				return adapter.ExecuteTool(ctx, toolName, arguments)
			}
		}
	}
	
	return "", fmt.Errorf("tool '%s' not found on any server", toolName)
}

func (hsm *HostServerManager) StopAll() error {
	return nil
}

// HostServerAdapter adapts host.ServerConnection to domain.MCPServer interface
type HostServerAdapter struct {
	connection  *host.ServerConnection
	toolsCache  []domain.Tool
	toolsCached bool
}

func (hsa *HostServerAdapter) Start(ctx context.Context) error {
	return nil
}

func (hsa *HostServerAdapter) Stop() error {
	return nil
}

func (hsa *HostServerAdapter) IsRunning() bool {
	return hsa.connection.Client != nil
}

func formatToolNameForOpenAI(serverName, toolName string) string {
	serverName = strings.ReplaceAll(serverName, ".", "_")
	serverName = strings.ReplaceAll(serverName, " ", "_")
	serverName = strings.ReplaceAll(serverName, "-", "_")
	
	toolName = strings.ReplaceAll(toolName, ".", "_")
	toolName = strings.ReplaceAll(toolName, " ", "_")
	
	return fmt.Sprintf("%s_%s", serverName, toolName)
}

func (hsa *HostServerAdapter) GetTools() ([]domain.Tool, error) {
	if hsa.toolsCached {
		return hsa.toolsCache, nil
	}

	// Type assert to stdio client
	stdioClient := hsa.connection.GetStdioClient()
	if stdioClient == nil {
		return nil, fmt.Errorf("server %s does not support stdio protocol", hsa.connection.Name)
	}
	
	result, err := tools.SendToolsList(stdioClient, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get tools from MCP server %s: %w", hsa.connection.Name, err)
	}

	var domainTools []domain.Tool
	for _, tool := range result.Tools {
		formattedName := formatToolNameForOpenAI(hsa.connection.Name, tool.Name)
		
		domainTool := domain.Tool{
			Type: "function",
			Function: domain.ToolFunction{
				Name:        formattedName,
				Description: fmt.Sprintf("[%s] %s", hsa.connection.Name, tool.Description),
				Parameters:  tool.InputSchema,
			},
		}
		domainTools = append(domainTools, domainTool)
	}

	hsa.toolsCache = domainTools
	hsa.toolsCached = true

	logging.Debug("Successfully got %d tools from server %s", len(domainTools), hsa.connection.Name)
	return domainTools, nil
}

func (hsa *HostServerAdapter) ExecuteTool(ctx context.Context, toolName string, arguments map[string]interface{}) (string, error) {
	actualToolName := toolName
	serverPrefix := hsa.connection.Name + "_"
	serverPrefixUnderscore := strings.ReplaceAll(hsa.connection.Name, "-", "_") + "_"
	
	if strings.HasPrefix(toolName, serverPrefix) {
		actualToolName = strings.TrimPrefix(toolName, serverPrefix)
	} else if strings.HasPrefix(toolName, serverPrefixUnderscore) {
		actualToolName = strings.TrimPrefix(toolName, serverPrefixUnderscore)
	}

	logging.Debug("Executing tool %s (actual: %s) on server %s", toolName, actualToolName, hsa.connection.Name)

	// Type assert to stdio client
	stdioClient := hsa.connection.GetStdioClient()
	if stdioClient == nil {
		return "", fmt.Errorf("server %s does not support stdio protocol", hsa.connection.Name)
	}
	
	result, err := tools.SendToolsCall(stdioClient, stdioClient.GetDispatcher(), actualToolName, arguments)
	if err != nil {
		return "", fmt.Errorf("MCP tool execution failed for %s: %w", actualToolName, err)
	}

	if result.IsError {
		return "", fmt.Errorf("tool execution failed: %s", result.Error)
	}

	// Extract text from content blocks
	var resultStr string
	switch content := result.Content.(type) {
	case string:
		// Direct string response
		resultStr = content
	case []interface{}:
		// Content blocks array (standard MCP format)
		// Extract text from the first text-type content block
		for _, item := range content {
			if block, ok := item.(map[string]interface{}); ok {
				if blockType, hasType := block["type"].(string); hasType && blockType == "text" {
					if text, hasText := block["text"].(string); hasText {
						resultStr = text
						break
					}
				}
			}
		}
		if resultStr == "" {
			// No text content found, marshal the whole thing as fallback
			resultBytes, err := json.Marshal(content)
			if err != nil {
				return "", fmt.Errorf("failed to marshal tool result: %w", err)
			}
			resultStr = string(resultBytes)
		}
	default:
		// Unknown format, marshal it
		resultBytes, err := json.Marshal(content)
		if err != nil {
			return "", fmt.Errorf("failed to marshal tool result: %w", err)
		}
		resultStr = string(resultBytes)
	}

	logging.Debug("Tool %s executed successfully on server %s", actualToolName, hsa.connection.Name)
	return resultStr, nil
}

func (hsa *HostServerAdapter) GetServerName() string {
	return hsa.connection.Name
}

func (hsa *HostServerAdapter) GetConfig() *config.ServerConfig {
	return &config.ServerConfig{
		Command: "mock",
		Args:    []string{},
	}
}
