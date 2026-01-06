# HTTP Proxy Configurations

HTTP proxy configurations for exposing MCP servers and workflow templates as REST APIs.

## Consistent Pattern

**Every proxy config follows the same template:**

```yaml
runas_type: proxy
version: "1.0"

# Source of what you're proxying
config_source: config/servers/NAME.yaml      # For MCP server
# OR
config_source: config/templates/NAME.yaml    # For workflow template

proxy_config:
  port: XXXX
  api_key: "${NAME_PROXY_API_KEY}"
  cors_origins: ["*"]
  
  # Optional: Enable HTTPS
  # tls:
  #   cert_file: /path/to/cert.crt
  #   key_file: /path/to/key.key
```

**One proxy config = One resource (MCP server OR workflow template)**

## Key Features

### 1. Explicit Config Source
- Points directly to the config file
- Works for MCP servers AND workflow templates
- Same pattern for both types

### 2. TLS/HTTPS Support
- Optional certificate configuration
- Production-ready security

### 3. Clear API Key Separation
- Service API keys: In `config/servers/` (authenticate to external services)
- Proxy API keys: In `config/proxy/` (authenticate HTTP clients to your proxy)

## Available Proxies

### MCP Servers

**filesystem.yaml** - Filesystem operations
```yaml
config_source: config/servers/filesystem.yaml
port: 8080
```

**brave-search.yaml** - Web search
```yaml
config_source: config/servers/search.yaml
port: 8081
```

### Workflow Templates

**simple-analysis.yaml** - Simple AI analysis
```yaml
config_source: config/templates/simple_analysis.yaml
port: 9000
```

**parallel-analysis.yaml** - Parallel processing workflow
```yaml
config_source: config/templates/parallel_analysis.yaml
port: 9001
```

**sentiment-analysis.yaml** - Sentiment analysis workflow
```yaml
config_source: config/templates/sentiment_analysis.yaml
port: 9002
```

**document-intelligence.yaml** - Document processing workflow
```yaml
config_source: config/templates/document_intelligence.yaml
port: 9003
```

### Skills

**skills.yaml** - Auto-discovered Anthropic skills
```yaml
runas_type: proxy-skills  # Different type - auto-discovery
port: 9010
```

## Usage Examples

### MCP Server Proxy

```bash
# Set proxy auth key
export FILESYSTEM_PROXY_API_KEY="secure-key-1"

# Start proxy
mcp-cli serve config/proxy/filesystem.yaml

# Access
curl -H "Authorization: ${FILESYSTEM_PROXY_API_KEY}" \
  http://localhost:8080/tools
```

### Workflow Template Proxy

```bash
# Set proxy auth key
export SIMPLE_ANALYSIS_PROXY_API_KEY="secure-key-2"

# Start proxy
mcp-cli serve config/proxy/simple-analysis.yaml

# Access
curl -H "Authorization: ${SIMPLE_ANALYSIS_PROXY_API_KEY}" \
  -X POST http://localhost:9000/tools/simple_analysis \
  -H "Content-Type: application/json" \
  -d '{"input_data": "Analyze this text"}'
```

### Multiple Proxies (Different Services)

```bash
# Start all services on different ports
mcp-cli serve config/proxy/filesystem.yaml &        # Port 8080
mcp-cli serve config/proxy/brave-search.yaml &      # Port 8081
mcp-cli serve config/proxy/simple-analysis.yaml &   # Port 9000
mcp-cli serve config/proxy/sentiment-analysis.yaml & # Port 9002
```

## Directory Structure

```
config/
├── servers/                    # MCP server definitions
│   ├── filesystem.yaml         # Filesystem MCP server config
│   └── search.yaml             # Search MCP server config
│
├── templates/                  # Workflow template definitions
│   ├── simple_analysis.yaml    # Simple analysis workflow
│   ├── sentiment_analysis.yaml # Sentiment workflow
│   └── ...
│
└── proxy/                      # HTTP proxy configs (this directory)
    ├── filesystem.yaml         # config_source → servers/filesystem.yaml
    ├── brave-search.yaml       # config_source → servers/search.yaml
    ├── simple-analysis.yaml    # config_source → templates/simple_analysis.yaml
    ├── sentiment-analysis.yaml # config_source → templates/sentiment_analysis.yaml
    └── skills.yaml             # Auto-discovery type
```

## Pattern Consistency

**MCP Server Proxy:**
```
config/proxy/brave-search.yaml
  ↓ config_source
config/servers/search.yaml
  ↓ server_name: search
config.yaml (servers.search)
  ↓ connection details
Brave Search MCP Server
```

**Workflow Template Proxy:**
```
config/proxy/simple-analysis.yaml
  ↓ config_source
config/templates/simple_analysis.yaml
  ↓ template definition
Simple Analysis Workflow
```

**Same pattern, different resources!**

## API Key Convention

### Two Different Keys

**Service API Key** (in `config/servers/`)
- Authenticates to external service
- Example: `BRAVE_API_KEY` for Brave Search service
- Set in server config environment

**Proxy API Key** (in `config/proxy/`)
- Authenticates HTTP clients to YOUR proxy
- Example: `BRAVE_PROXY_API_KEY` for brave-search proxy
- Set in proxy config

### Naming Convention

```bash
# For MCP servers: {SERVER}_PROXY_API_KEY
export FILESYSTEM_PROXY_API_KEY="key1"
export BRAVE_PROXY_API_KEY="key2"

# For templates: {TEMPLATE}_PROXY_API_KEY
export SIMPLE_ANALYSIS_PROXY_API_KEY="key3"
export SENTIMENT_ANALYSIS_PROXY_API_KEY="key4"

# For skills: SKILLS_PROXY_API_KEY
export SKILLS_PROXY_API_KEY="key5"
```

## How config_source Works

### For MCP Servers

1. Proxy config specifies: `config_source: config/servers/search.yaml`
2. Read that file to extract: `server_name: search`
3. Find server in main config.yaml: `servers.search`
4. Connect to MCP server and auto-discover tools
5. Expose tools via HTTP

### For Workflow Templates

1. Proxy config specifies: `config_source: config/templates/simple_analysis.yaml`
2. Extract template name from filename: `simple_analysis`
3. Load template definition from config
4. Expose as HTTP endpoint
5. Execute workflow when called

## For Open WebUI Integration

```bash
# Set all proxy keys
export FILESYSTEM_PROXY_API_KEY="key1"
export BRAVE_PROXY_API_KEY="key2"
export SIMPLE_ANALYSIS_PROXY_API_KEY="key3"
export SENTIMENT_ANALYSIS_PROXY_API_KEY="key4"

# Start proxies (each on its own port)
mcp-cli serve config/proxy/filesystem.yaml &
mcp-cli serve config/proxy/brave-search.yaml &
mcp-cli serve config/proxy/simple-analysis.yaml &
mcp-cli serve config/proxy/sentiment-analysis.yaml &

# Add each to Open WebUI:
# Admin → Settings → Tools → Add Tool Server
# - Filesystem: http://localhost:8080
# - Search: http://localhost:8081
# - Simple Analysis: http://localhost:9000
# - Sentiment Analysis: http://localhost:9002
```

## HTTPS Example

```yaml
# config/proxy/simple-analysis.yaml
runas_type: proxy
version: "1.0"

config_source: config/templates/simple_analysis.yaml

proxy_config:
  port: 9443  # HTTPS port
  api_key: "${SIMPLE_ANALYSIS_PROXY_API_KEY}"
  
  tls:
    cert_file: /etc/ssl/certs/analysis.crt
    key_file: /etc/ssl/private/analysis.key
```

```bash
# Access via HTTPS
curl -H "Authorization: ${SIMPLE_ANALYSIS_PROXY_API_KEY}" \
  https://localhost:9443/tools
```

## Creating New Proxies

### For a New MCP Server

1. Create server config in `config/servers/myserver.yaml`
2. Add to main `config.yaml` servers section
3. Create proxy config:

```yaml
# config/proxy/myserver.yaml
runas_type: proxy
version: "1.0"

config_source: config/servers/myserver.yaml

proxy_config:
  port: 8090
  api_key: "${MYSERVER_PROXY_API_KEY}"
  cors_origins: ["*"]
```

### For a New Workflow Template

1. Create template in `config/templates/myworkflow.yaml`
2. Create proxy config:

```yaml
# config/proxy/myworkflow.yaml
runas_type: proxy
version: "1.0"

config_source: config/templates/myworkflow.yaml

proxy_config:
  port: 9010
  api_key: "${MYWORKFLOW_PROXY_API_KEY}"
  cors_origins: ["*"]
```

## Benefits

✅ **Consistent Pattern** - Same structure for servers and templates  
✅ **One Resource Per Proxy** - Clear isolation  
✅ **Explicit Source** - Easy to trace what's being proxied  
✅ **Individual Ports** - Each service isolated  
✅ **Individual Auth** - Separate API keys per service  
✅ **HTTPS Ready** - TLS support built-in  
✅ **Scalable** - Add new services easily  

## Troubleshooting

**Error: "config_source points to template 'X' but it's not defined"**
- Check template exists in `config/templates/`
- Verify filename matches template name

**Port already in use:**
- Each proxy needs unique port
- Check for running processes: `lsof -i :PORT`

**Certificate errors with HTTPS:**
- Verify cert and key files exist
- Check permissions on certificate files
- Test with: `openssl s_client -connect localhost:PORT`

**API key not working:**
- Check environment variable is set: `echo $NAME_PROXY_API_KEY`
- Verify using correct key (proxy key, not service key)
- Include in Authorization header: `Authorization: your-key`
