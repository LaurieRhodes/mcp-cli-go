# HTTP Proxy Server Guide

Complete guide to exposing MCP servers as REST APIs using the HTTP proxy server.

## Table of Contents

- [Overview](#overview)
- [What is the HTTP Proxy?](#what-is-the-http-proxy)
- [How It Works](#how-it-works)
- [Quick Start](#quick-start)
- [Configuration Reference](#configuration-reference)
- [OpenAPI Integration](#openapi-integration)
- [Authentication](#authentication)
- [Troubleshooting](#troubleshooting)
- [Production Deployment](#production-deployment)

---

## Overview

The **HTTP Proxy Server** converts MCP (Model Context Protocol) servers into standard REST APIs with:

✅ **Auto-discovery** - Tools from MCP servers → REST endpoints  
✅ **OpenAPI spec** - Auto-generated documentation  
✅ **API Key auth** - Secure access control  
✅ **CORS support** - Browser-friendly  
✅ **TLS/HTTPS** - Production-ready security  

### Production Status

**✅ Fully Functional:**
- MCP server connection and tool discovery
- REST API endpoint generation
- OpenAPI 3.0 specification
- API key authentication
- CORS middleware
- Health checks
- Swagger UI documentation
- TLS/HTTPS support

**Verified with:** OpenWebUI, cURL, web browsers

---

## What is the HTTP Proxy?

### The Problem

**MCP servers use stdio/SSE** - great for desktop apps, not for web:
```
Desktop App ←→ stdin/stdout ←→ MCP Server ✅
Web App     ←→ stdin/stdout ←→ MCP Server ❌
```

**HTTP Proxy bridges the gap:**
```
Web App ←→ HTTP/REST ←→ Proxy ←→ stdin/stdout ←→ MCP Server ✅
```

### Use Cases

**1. Web Integration**
```
React/Vue App → REST API → Proxy → MCP Server (filesystem, brave-search, etc.)
```

**2. Tool Aggregation (like OpenWebUI)**
```
OpenWebUI → REST API → Proxy → bash MCP Server
          → REST API → Proxy → filesystem MCP Server  
          → REST API → Proxy → brave-search MCP Server
```

**3. Legacy System Integration**
```
Old System (REST only) → Proxy → Modern MCP Server
```

---

## How It Works

### Architecture

```
┌──────────────────────────────────────────────────┐
│ HTTP Proxy Server                                │
│                                                  │
│  ┌────────────┐     ┌─────────────────────────┐ │
│  │ REST API   │────▶│ Auto-Discovery          │ │
│  │ Endpoints  │     │                         │ │
│  │            │     │ 1. Read config_source   │ │
│  │ POST /bash │     │ 2. Connect to MCP server│ │
│  │ POST /read │     │ 3. List tools           │ │
│  │ GET  /docs │     │ 4. Generate endpoints   │ │
│  └────────────┘     └─────────────────────────┘ │
│                                                  │
│  ┌────────────┐     ┌─────────────────────────┐ │
│  │ Middleware │────▶│ - API Key Auth          │ │
│  │            │     │ - CORS Headers          │ │
│  │            │     │ - Request Logging       │ │
│  └────────────┘     └─────────────────────────┘ │
│                                                  │
│  ┌────────────┐     ┌─────────────────────────┐ │
│  │ OpenAPI    │────▶│ Auto-generated from:    │ │
│  │ Generator  │     │ - MCP tool schemas      │ │
│  │            │     │ - Tool descriptions     │ │
│  └────────────┘     └─────────────────────────┘ │
│         │                                        │
└─────────┼──────────────────────────────────────┘
          │
          ▼
┌──────────────────────────────────────────────────┐
│ MCP Server (stdio)                               │
│                                                  │
│  Examples: bash, filesystem, brave-search        │
└──────────────────────────────────────────────────┘
```

### Request Flow

**1. Client sends REST request:**
```http
POST /bash HTTP/1.1
Host: localhost:4000
Authorization: Bearer my-api-key
Content-Type: application/json

{
  "command": "ls -la",
  "description": "List files"
}
```

**2. Proxy validates & authenticates:**
```
✓ API key valid
✓ Content-Type application/json
✓ Request body valid
```

**3. Proxy converts to MCP protocol:**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "bash",
    "arguments": {
      "command": "ls -la",
      "description": "List files"
    }
  }
}
```

**4. MCP server executes & responds:**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "content": [{
      "type": "text",
      "text": "total 48\ndrwxr-xr-x  ..."
    }]
  }
}
```

**5. Proxy converts to REST response:**
```json
{
  "success": true,
  "tool": "bash",
  "result": "total 48\ndrwxr-xr-x  ..."
}
```

---

## Quick Start

### 1. Choose an MCP Server

Any MCP server in `config/servers/`:
```bash
ls config/servers/
# bash.yaml
# filesystem.yaml
# brave-search.yaml
```

### 2. Create Proxy Config

Create `config/proxy/my-server.yaml`:

```yaml
runas_type: proxy
version: "1.0"

# Point to the MCP server config
config_source: config/servers/bash.yaml

proxy_config:
  port: 4000
  api_key: "${MY_PROXY_API_KEY}"  # Or hardcoded: "my-secret-key"
  cors_origins: ["*"]
```

### 3. Set API Key

```bash
# Option 1: Environment variable
export MY_PROXY_API_KEY="your-secure-key-here"

# Option 2: .env file
echo "MY_PROXY_API_KEY=your-secure-key-here" >> .env

# Option 3: Hardcode in YAML (not recommended for production)
api_key: "your-secure-key-here"
```

### 4. Start the Proxy

```bash
./mcp-cli serve config/proxy/my-server.yaml
```

**Output:**
```
[INFO] Starting HTTP proxy server on 0.0.0.0:4000
[INFO] OpenAPI docs available at: http://0.0.0.0:4000/docs
[INFO] Server: bash-proxy v1.0.0
[INFO] Discovered 1 tools from bash
[INFO] Exposed tools: 1
```

### 5. Test It

**View OpenAPI docs:**
```bash
open http://localhost:4000/docs
```

**Call a tool:**
```bash
curl -X POST http://localhost:4000/bash \
  -H "Authorization: Bearer your-secure-key-here" \
  -H "Content-Type: application/json" \
  -d '{
    "command": "echo Hello World",
    "description": "Test command"
  }'
```

**Response:**
```json
{
  "success": true,
  "tool": "bash",
  "result": "Hello World"
}
```

---

## Configuration Reference

### Minimal Configuration

```yaml
runas_type: proxy
version: "1.0"
config_source: config/servers/bash.yaml

proxy_config:
  api_key: "your-key"
```

That's it! Defaults apply for everything else.

### Full Configuration

```yaml
runas_type: proxy
version: "1.0"

# MCP server to proxy
config_source: config/servers/bash.yaml

# Optional: Server metadata (shown in OpenAPI)
server_info:
  name: bash-proxy
  version: 1.0.0
  description: "REST API for bash MCP server"

proxy_config:
  # Port to listen on
  # Default: 8080
  port: 4000
  
  # Host to bind to
  # Default: "0.0.0.0" (all interfaces)
  # Production: Consider "127.0.0.1" (localhost only) with reverse proxy
  host: "0.0.0.0"
  
  # API key for authentication (REQUIRED)
  # Use environment variable: "${VAR_NAME}"
  # Or hardcode (not recommended): "actual-key"
  api_key: "${BASH_PROXY_API_KEY}"
  
  # CORS allowed origins
  # Default: ["*"] (allow all)
  # Production: ["https://yourdomain.com"]
  cors_origins:
    - "*"
    - "https://app.example.com"
    - "https://admin.example.com"
  
  # Enable Swagger UI at /docs
  # Default: true
  enable_docs: true
  
  # Base path for all endpoints
  # Default: "" (root)
  # Example: "/api/v1" → endpoints at /api/v1/bash, /api/v1/docs
  base_path: ""
  
  # Optional: TLS/HTTPS configuration
  tls:
    cert_file: /path/to/server.crt
    key_file: /path/to/server.key
```

### Configuration Field Details

#### `runas_type`
**Required.** Must be `proxy`.

#### `config_source`
**Required.** Path to MCP server config file.

**Examples:**
```yaml
config_source: config/servers/bash.yaml
config_source: config/servers/filesystem.yaml
config_source: config/servers/brave-search.yaml
```

**What it does:**
1. Reads the server config
2. Starts the MCP server (stdio connection)
3. Sends `tools/list` request
4. Auto-generates REST endpoints from tool schemas

#### `proxy_config.port`
**Optional.** Default: `8080`

Port for HTTP server.

**Examples:**
```yaml
port: 4000  # http://localhost:4000
port: 8080  # http://localhost:8080 (default)
port: 3000  # http://localhost:3000
```

#### `proxy_config.host`
**Optional.** Default: `"0.0.0.0"` (all interfaces)

**Options:**
- `"0.0.0.0"` - Listen on all network interfaces
- `"127.0.0.1"` - Localhost only (recommended with reverse proxy)

**Example:**
```yaml
# Development: accept from anywhere
host: "0.0.0.0"

# Production: localhost only + nginx reverse proxy
host: "127.0.0.1"
```

#### `proxy_config.api_key`
**Required.** API key for authentication.

**Best practices:**
```yaml
# ✅ GOOD: Environment variable
api_key: "${MY_API_KEY}"

# ⚠️ OK: For development only
api_key: "dev-key-12345"

# ❌ BAD: Hardcoded production key
api_key: "prod-key-abc123"  # Don't commit this!
```

**Usage:**
```bash
# Set environment variable
export MY_API_KEY="$(openssl rand -hex 32)"

# Or in .env file
echo "MY_API_KEY=$(openssl rand -hex 32)" >> .env
```

**Client usage:**
```http
Authorization: Bearer your-api-key
# or just:
Authorization: your-api-key
```

#### `proxy_config.cors_origins`
**Optional.** Default: `["*"]`

List of allowed origins for CORS.

**Examples:**
```yaml
# Development: allow all
cors_origins: ["*"]

# Production: specific domains only
cors_origins:
  - "https://app.example.com"
  - "https://admin.example.com"

# Localhost for testing
cors_origins:
  - "http://localhost:3000"
  - "http://localhost:5173"
```

#### `proxy_config.enable_docs`
**Optional.** Default: `true`

Enable Swagger UI documentation at `/docs`.

```yaml
enable_docs: true   # ✅ /docs available
enable_docs: false  # ❌ /docs returns 404
```

#### `proxy_config.base_path`
**Optional.** Default: `""` (root)

Prefix for all endpoints.

**Examples:**
```yaml
base_path: ""        # → /bash, /docs, /health
base_path: "/api"    # → /api/bash, /api/docs, /api/health
base_path: "/api/v1" # → /api/v1/bash, /api/v1/docs, /api/v1/health
```

#### `proxy_config.tls`
**Optional.** Enable HTTPS.

```yaml
tls:
  cert_file: /path/to/server.crt
  key_file: /path/to/server.key
```

**Generate self-signed cert for development:**
```bash
openssl req -x509 -newkey rsa:4096 -nodes \
  -keyout server.key -out server.crt \
  -days 365 -subj "/CN=localhost"
```

---

## OpenAPI Integration

### Auto-Generated Specification

The proxy automatically generates OpenAPI 3.0 spec from MCP tools.

**Access spec:**
```bash
curl http://localhost:4000/openapi.json
```

**Example output:**
```json
{
  "openapi": "3.0.0",
  "info": {
    "title": "bash-proxy",
    "version": "1.0.0"
  },
  "paths": {
    "/bash": {
      "post": {
        "operationId": "tool_bash_post",
        "summary": "bash",
        "description": "Execute bash commands...",
        "requestBody": {
          "content": {
            "application/json": {
              "schema": {
                "$ref": "#/components/schemas/bashRequest"
              }
            }
          }
        }
      }
    }
  },
  "components": {
    "schemas": {
      "bashRequest": {
        "type": "object",
        "properties": {
          "command": {
            "type": "string",
            "description": "The bash command to execute"
          }
        },
        "required": ["command"]
      }
    }
  }
}
```

### Swagger UI Documentation

**Access interactive docs:**
```
http://localhost:4000/docs
```

**Features:**
- ✅ Try out API calls directly
- ✅ See request/response examples
- ✅ View schemas and parameters
- ✅ Test authentication

### Integration with OpenWebUI

**OpenWebUI Issue #6427 Fix:**

During development, we discovered OpenWebUI requires:
1. ✅ Flat schema structure (properties at root level)
2. ✅ Endpoint paths at `/{tool_name}` (not `/tools/{tool_name}`)
3. ✅ `operationId` field in each operation

**All fixed!** The proxy now works perfectly with OpenWebUI.

**OpenWebUI Configuration:**
```json
{
  "name": "MCP Tools",
  "url": "http://localhost:4000/openapi.json",
  "api_key": "your-api-key"
}
```

---

## Authentication

### API Key Authentication

**Every request must include:**
```http
Authorization: Bearer your-api-key
```

**Or simply:**
```http
Authorization: your-api-key
```

Both formats work.

### Setting the API Key

**1. Environment Variable (Recommended):**
```bash
export MY_API_KEY="$(openssl rand -hex 32)"
./mcp-cli serve config/proxy/my-server.yaml
```

**2. .env File:**
```bash
echo "MY_API_KEY=$(openssl rand -hex 32)" >> .env
./mcp-cli serve config/proxy/my-server.yaml
```

**3. Hardcoded (Development Only):**
```yaml
proxy_config:
  api_key: "dev-key-12345"
```

### Generating Secure Keys

```bash
# 256-bit random key (recommended)
openssl rand -hex 32

# UUID
uuidgen

# Base64 encoded random bytes
openssl rand -base64 32
```

### Testing Authentication

**Valid key:**
```bash
curl -H "Authorization: Bearer correct-key" \
  http://localhost:4000/health
# → 200 OK
```

**Invalid key:**
```bash
curl -H "Authorization: Bearer wrong-key" \
  http://localhost:4000/health
# → 401 Unauthorized
```

**Missing key:**
```bash
curl http://localhost:4000/health
# → 401 Unauthorized
```

---

## Troubleshooting

### Issue: Server won't start

**Error:**
```
Failed to start HTTP proxy server: api_key is required
```

**Fix:**
```yaml
proxy_config:
  api_key: "${MY_API_KEY}"  # Set this environment variable
```

### Issue: Tools not discovered

**Error:**
```
Failed to discover tools from source: server not found
```

**Diagnosis:**
```bash
# Check if server config exists
ls config/servers/bash.yaml

# Check config_source path in proxy config
grep config_source config/proxy/my-server.yaml
```

**Fix:**
```yaml
# Ensure path is correct
config_source: config/servers/bash.yaml  # ✅ Correct
config_source: servers/bash.yaml         # ❌ Wrong
```

### Issue: OpenWebUI shows 0 tools

**Symptom:** Proxy connects but no tools visible.

**Diagnosis:**
```bash
# Check OpenAPI spec
curl http://localhost:4000/openapi.json | jq '.paths | keys'
```

**Expected:**
```json
[
  "/bash",
  "/health",
  "/tools"
]
```

**If tools missing:**
```bash
# Check verbose logs
./mcp-cli serve --verbose config/proxy/my-server.yaml 2>&1 | grep -i "discovered"
```

**Common causes:**
1. MCP server failed to start
2. Tool discovery returned empty
3. Schema generation failed

**Fix:** Check MCP server config and logs.

### Issue: CORS errors in browser

**Error:**
```
Access to fetch at 'http://localhost:4000/bash' has been blocked by CORS policy
```

**Fix:**
```yaml
proxy_config:
  cors_origins:
    - "http://localhost:3000"  # Add your app's origin
    - "*"                       # Or allow all (development only)
```

### Issue: TLS certificate errors

**Error:**
```
x509: certificate signed by unknown authority
```

**For development (self-signed cert):**
```bash
# cURL: Skip verification
curl -k https://localhost:4000/health

# Browser: Accept certificate warning
```

**For production:** Use proper CA-signed certificate.

---

## Production Deployment

### Checklist

Before deploying to production:

- [ ] Generate strong API key (32+ bytes random)
- [ ] Use environment variables for secrets
- [ ] Configure specific CORS origins (not `["*"]`)
- [ ] Enable TLS/HTTPS
- [ ] Use reverse proxy (nginx/caddy)
- [ ] Set up monitoring
- [ ] Configure rate limiting
- [ ] Test all endpoints
- [ ] Document API for users

### Reverse Proxy Setup

**nginx configuration:**
```nginx
upstream mcp_proxy {
    server 127.0.0.1:4000;
}

server {
    listen 443 ssl http2;
    server_name api.example.com;

    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;

    location / {
        proxy_pass http://mcp_proxy;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # WebSocket support (if needed)
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}
```

**mcp-cli config:**
```yaml
proxy_config:
  host: "127.0.0.1"  # Localhost only - nginx handles external traffic
  port: 4000
  api_key: "${PROXY_API_KEY}"
  cors_origins:
    - "https://app.example.com"
```

### Systemd Service

**`/etc/systemd/system/mcp-proxy.service`:**
```ini
[Unit]
Description=MCP Proxy Server
After=network.target

[Service]
Type=simple
User=mcp
WorkingDirectory=/opt/mcp-cli
Environment="PROXY_API_KEY=your-key-here"
ExecStart=/opt/mcp-cli/mcp-cli serve /opt/mcp-cli/config/proxy/production.yaml
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

**Commands:**
```bash
sudo systemctl enable mcp-proxy
sudo systemctl start mcp-proxy
sudo systemctl status mcp-proxy
```

### Docker Deployment

**Dockerfile:**
```dockerfile
FROM golang:1.21 AS builder
WORKDIR /app
COPY . .
RUN go build -o mcp-cli

FROM debian:bookworm-slim
COPY --from=builder /app/mcp-cli /usr/local/bin/
COPY config /app/config
WORKDIR /app

ENV PROXY_API_KEY=""
EXPOSE 4000

CMD ["mcp-cli", "serve", "config/proxy/production.yaml"]
```

**docker-compose.yml:**
```yaml
version: '3.8'
services:
  mcp-proxy:
    build: .
    ports:
      - "4000:4000"
    environment:
      - PROXY_API_KEY=${PROXY_API_KEY}
    restart: unless-stopped
```

### Monitoring

**Health check endpoint:**
```bash
curl http://localhost:4000/health
```

**Expected response:**
```json
{
  "status": "healthy",
  "server": "bash-proxy",
  "version": "1.0.0",
  "tools": 1
}
```

**Monitoring script:**
```bash
#!/bin/bash
# health-check.sh

HEALTH_URL="http://localhost:4000/health"
API_KEY="your-key"

response=$(curl -s -H "Authorization: $API_KEY" "$HEALTH_URL")
status=$(echo "$response" | jq -r '.status')

if [ "$status" == "healthy" ]; then
    echo "OK: Server healthy"
    exit 0
else
    echo "CRITICAL: Server unhealthy"
    exit 2
fi
```

---

## Summary

**The HTTP Proxy provides:**

✅ **Auto-discovery** - Tools from MCP → REST endpoints  
✅ **OpenAPI 3.0** - Auto-generated documentation  
✅ **Swagger UI** - Interactive testing  
✅ **API Key auth** - Secure access control  
✅ **CORS support** - Browser-friendly  
✅ **Production-ready** - TLS, reverse proxy support  
✅ **OpenWebUI compatible** - Fixed schema/path issues  

**Recommended setup:**
```yaml
runas_type: proxy
config_source: config/servers/your-server.yaml
proxy_config:
  port: 4000
  api_key: "${MY_API_KEY}"
  cors_origins: ["*"]
```

**Next steps:**
- Choose an MCP server
- Create proxy config
- Start and test
- Deploy to production

---

**Last Updated:** January 4, 2026  
**Tested With:** OpenWebUI, cURL, Swagger UI  
**Issues Fixed:** OpenWebUI schema, endpoint paths, operationId
