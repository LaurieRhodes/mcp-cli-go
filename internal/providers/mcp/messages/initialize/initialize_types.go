package initialize

// ClientInfo describes information about the client
type ClientInfo struct {
	// The name of the client
	Name string `json:"name"`

	// The version of the client
	Version string `json:"version"`
}

// ClientCapabilities defines what capabilities the client supports
type ClientCapabilities struct {
	// Whether the client supports configuration changes during runtime
	SupportsConfigurationChange bool `json:"supportsConfigurationChange,omitempty"`

	// Whether the client supports progress reporting
	SupportsProgressReporting bool `json:"supportsProgressReporting,omitempty"`

	// Whether the client supports cancellation
	SupportsCancellation bool `json:"supportsCancellation,omitempty"`
}

// InitializeParams represents the parameters for an initialize request
type InitializeParams struct {
	// The version of the Model Context Protocol that the client is using
	ProtocolVersion string `json:"protocolVersion"`

	// Information about the client
	ClientInfo ClientInfo `json:"clientInfo"`

	// Client capabilities
	Capabilities ClientCapabilities `json:"capabilities"`

	// Configuration settings for the server (optional)
	ServerConfig map[string]interface{} `json:"serverConfig,omitempty"`
}

// ServerInfo contains information about the server
type ServerInfo struct {
	// The name of the server
	Name string `json:"name"`

	// The version of the server
	Version string `json:"version"`

	// A description of the server
	Description string `json:"description,omitempty"`

	// The version of the Model Context Protocol implemented by the server
	ProtocolVersion string `json:"protocolVersion"`
}

// ServerCapabilities defines what capabilities the server supports
type ServerCapabilities struct {
	// Whether the server supports config changes during runtime
	SupportsConfigurationChange bool `json:"supportsConfigurationChange,omitempty"`

	// Whether the server supports progress reporting
	SupportsProgressReporting bool `json:"supportsProgressReporting,omitempty"`

	// Whether the server supports cancellation
	SupportsCancellation bool `json:"supportsCancellation,omitempty"`

	// Whether the server provides tools
	ProvidesTools bool `json:"providesTools,omitempty"`

	// Whether the server provides prompts
	ProvidesPrompts bool `json:"providesPrompts,omitempty"`

	// Whether the server provides resources
	ProvidesResources bool `json:"providesResources,omitempty"`
}

// InitializeResult represents the result of an initialize request
type InitializeResult struct {
	// Information about the server
	ServerInfo ServerInfo `json:"serverInfo"`

	// Server capabilities
	Capabilities ServerCapabilities `json:"capabilities"`
}
