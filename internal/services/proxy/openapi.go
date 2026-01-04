package proxy

import (
	"fmt"
)

// OpenAPISpec represents an OpenAPI 3.0 specification
type OpenAPISpec struct {
	OpenAPI    string                 `json:"openapi"`
	Info       OpenAPIInfo            `json:"info"`
	Servers    []OpenAPIServer        `json:"servers,omitempty"`
	Paths      map[string]interface{} `json:"paths"`
	Components OpenAPIComponents      `json:"components,omitempty"`
}

// OpenAPIInfo contains API metadata
type OpenAPIInfo struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Version     string `json:"version"`
}

// OpenAPIServer represents a server URL
type OpenAPIServer struct {
	URL         string `json:"url"`
	Description string `json:"description,omitempty"`
}

// OpenAPIComponents contains reusable components
type OpenAPIComponents struct {
	Schemas         map[string]interface{} `json:"schemas,omitempty"`
	SecuritySchemes map[string]interface{} `json:"securitySchemes,omitempty"`
}

// generateOpenAPISpec generates the OpenAPI specification for all tools
func (s *ProxyServer) generateOpenAPISpec() *OpenAPISpec {
	proxyConfig := s.config.ProxyConfig
	
	spec := &OpenAPISpec{
		OpenAPI: "3.0.0",
		Info: OpenAPIInfo{
			Title:       s.config.ServerInfo.Name,
			Description: s.config.ServerInfo.Description,
			Version:     s.config.ServerInfo.Version,
		},
		Paths: make(map[string]interface{}),
		Components: OpenAPIComponents{
			Schemas: make(map[string]interface{}),
			SecuritySchemes: map[string]interface{}{
				"ApiKeyAuth": map[string]interface{}{
					"type": "apiKey",
					"in":   "header",
					"name": "Authorization",
					"description": "API key authentication. Use 'Bearer <your-api-key>' or just '<your-api-key>'",
				},
			},
		},
	}
	
	// Add server URL if we can determine it
	port := proxyConfig.Port
	if port == 0 {
		port = 8080
	}
	spec.Servers = []OpenAPIServer{
		{
			URL:         fmt.Sprintf("http://localhost:%d%s", port, proxyConfig.BasePath),
			Description: "Local development server",
		},
	}
	
	// Add health endpoint
	spec.Paths["/health"] = map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "Health check",
			"description": "Check if the server is healthy",
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Server is healthy",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"type": "object",
								"properties": map[string]interface{}{
									"status":  map[string]interface{}{"type": "string"},
									"server":  map[string]interface{}{"type": "string"},
									"version": map[string]interface{}{"type": "string"},
									"tools":   map[string]interface{}{"type": "integer"},
								},
							},
						},
					},
				},
			},
		},
	}
	
	// Add tools list endpoint
	spec.Paths["/tools"] = map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "List tools",
			"description": "Get a list of all available tools",
			"security": []map[string][]string{
				{"ApiKeyAuth": {}},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "List of tools",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"type": "object",
								"properties": map[string]interface{}{
									"tools": map[string]interface{}{
										"type": "array",
										"items": map[string]interface{}{
											"type": "object",
											"properties": map[string]interface{}{
												"name":        map[string]interface{}{"type": "string"},
												"description": map[string]interface{}{"type": "string"},
												"template":    map[string]interface{}{"type": "string"},
												"endpoint":    map[string]interface{}{"type": "string"},
											},
										},
									},
									"count": map[string]interface{}{"type": "integer"},
								},
							},
						},
					},
				},
			},
		},
	}
	
	// Add each tool as a path
	for _, tool := range s.config.Tools {
		path := fmt.Sprintf("/%s", tool.Name)
		
		// Use the input schema directly - MCP servers provide complete JSON Schema objects
		// The inputSchema already contains type, properties, and required fields
		requestSchema := tool.InputSchema
		
		// Store schema in components for reuse
		schemaName := fmt.Sprintf("%sRequest", tool.Name)
		spec.Components.Schemas[schemaName] = requestSchema
		
		// Create response schema
		responseSchema := map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"success": map[string]interface{}{
					"type":        "boolean",
					"description": "Whether the operation was successful",
				},
				"result": map[string]interface{}{
					"type":        "string",
					"description": "The result from the tool execution",
				},
				"tool": map[string]interface{}{
					"type":        "string",
					"description": "The name of the tool that was executed",
				},
			},
		}
		
		responseSchemaName := fmt.Sprintf("%sResponse", tool.Name)
		spec.Components.Schemas[responseSchemaName] = responseSchema
		
		// Add path operation
		spec.Paths[path] = map[string]interface{}{
			"post": map[string]interface{}{
				"operationId": fmt.Sprintf("tool_%s_post", tool.Name),
				"summary":     tool.Name,
				"description": tool.Description,
				"tags":        []string{"Tools"},
				"security": []map[string][]string{
					{"ApiKeyAuth": {}},
				},
				"requestBody": map[string]interface{}{
					"required": true,
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"$ref": fmt.Sprintf("#/components/schemas/%s", schemaName),
							},
						},
					},
				},
				"responses": map[string]interface{}{
					"200": map[string]interface{}{
						"description": "Successful execution",
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"$ref": fmt.Sprintf("#/components/schemas/%s", responseSchemaName),
								},
							},
						},
					},
					"400": map[string]interface{}{
						"description": "Bad request - invalid input",
					},
					"401": map[string]interface{}{
						"description": "Unauthorized - invalid or missing API key",
					},
					"500": map[string]interface{}{
						"description": "Internal server error - execution failed",
					},
				},
			},
		}
	}
	
	return spec
}

// generateSwaggerUIHTML generates the HTML for Swagger UI
func (s *ProxyServer) generateSwaggerUIHTML() string {
	basePath := s.config.ProxyConfig.BasePath
	
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s - API Documentation</title>
    <link rel="stylesheet" type="text/css" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui.css">
    <style>
        body { margin: 0; padding: 0; }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
    <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui-standalone-preset.js"></script>
    <script>
        window.onload = function() {
            const ui = SwaggerUIBundle({
                url: "%s/openapi.json",
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIStandalonePreset
                ],
                plugins: [
                    SwaggerUIBundle.plugins.DownloadUrl
                ],
                layout: "StandaloneLayout"
            });
            window.ui = ui;
        };
    </script>
</body>
</html>`, s.config.ServerInfo.Name, basePath)
}
