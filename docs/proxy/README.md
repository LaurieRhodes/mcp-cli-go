# HTTP Proxy Server Documentation

Documentation for the HTTP Proxy feature that converts MCP servers into REST APIs.

## 📚 Documentation

### **[Server Guide](proxy-server-guide.md)** - Complete Documentation

The comprehensive guide covering everything about the HTTP Proxy:

**Getting Started:**
- What is the HTTP Proxy?
- How it works (architecture & request flow)
- Quick start (2 minutes to first API call)

**Configuration:**
- All configuration options explained
- Environment variables and secrets
- TLS/HTTPS setup
- CORS configuration

**Features:**
- Auto-discovery from MCP servers
- OpenAPI 3.0 specification generation
- Swagger UI integration
- API key authentication
- Production deployment

**Integration:**
- OpenWebUI setup and fixes
- nginx reverse proxy
- systemd service
- Docker deployment

**Troubleshooting:**
- Common issues and solutions
- Debugging commands
- Production monitoring

---

### **[Quick Reference](proxy-quick-reference.md)** - Fast Lookup

Quick reference for common tasks:

- 2-minute setup
- Configuration examples
- Testing commands
- Common patterns
- Debugging tips

---

## ⚡ Quick Start

### 1. Create Proxy Config

```yaml
# config/proxy/my-server.yaml
runas_type: proxy
version: "1.0"
config_source: config/servers/bash.yaml

proxy_config:
  port: 4000
  api_key: "${MY_API_KEY}"
  cors_origins: ["*"]
```

### 2. Set API Key

```bash
export MY_API_KEY="$(openssl rand -hex 32)"
```

### 3. Start Server

```bash
./mcp-cli serve config/proxy/my-server.yaml
```

### 4. Test

```bash
# View docs
open http://localhost:4000/docs

# Call endpoint
curl -X POST http://localhost:4000/bash \
  -H "Authorization: Bearer $MY_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command": "echo Hello", "description": "Test"}'
```

---

## 🎯 What Gets Created

When you start a proxy server, it automatically:

1. **Connects to MCP Server** - Uses `config_source` to find and start the MCP server
2. **Discovers Tools** - Sends `tools/list` request to get available tools
3. **Generates REST Endpoints** - Creates `POST /{tool_name}` for each tool
4. **Creates OpenAPI Spec** - Auto-generates documentation at `/openapi.json`
5. **Enables Swagger UI** - Interactive testing at `/docs`

**Example:**

```
bash MCP Server → Proxy discovers 1 tool → Creates:
  • POST /bash - Execute bash commands
  • GET /health - Health check
  • GET /tools - List tools
  • GET /openapi.json - OpenAPI spec
  • GET /docs - Swagger UI
```

---

## 🔍 Key Features

### ✅ Auto-Discovery

No manual configuration needed - proxy automatically:
- Reads MCP server config from `config_source`
- Connects via stdio
- Discovers all available tools
- Generates REST endpoints with proper schemas

### ✅ OpenAPI 3.0

Auto-generated specification includes:
- Complete endpoint documentation
- Request/response schemas from MCP tool definitions
- Security schemes (API key auth)
- Example values and descriptions

### ✅ Security

Production-ready security:
- Required API key authentication
- CORS support with configurable origins
- TLS/HTTPS support
- Health endpoint (no auth required)

### ✅ OpenWebUI Compatible

Tested and working with OpenWebUI:
- Fixed schema structure issues
- Proper endpoint paths
- Correct operationId fields
- Ready for tool aggregation

---

## 📊 Use Cases

### 1. Web Application Integration

```
React App → REST API → Proxy → bash MCP Server
                             → filesystem MCP Server
                             → brave-search MCP Server
```

### 2. OpenWebUI Tool Aggregation

```
OpenWebUI → http://localhost:4000 (bash proxy)
          → http://localhost:4001 (filesystem proxy)
          → http://localhost:4002 (brave-search proxy)
```

### 3. Legacy System Integration

```
Old System (REST only) → Proxy → Modern MCP Server
```

---

## 🔧 Configuration Examples

### Minimal (Development)

```yaml
runas_type: proxy
config_source: config/servers/bash.yaml
proxy_config:
  api_key: "dev-key-12345"
```

### Production (with TLS)

```yaml
runas_type: proxy
config_source: config/servers/bash.yaml

proxy_config:
  port: 443
  host: "127.0.0.1"  # localhost only, nginx handles external
  api_key: "${PROXY_API_KEY}"
  cors_origins:
    - "https://app.example.com"
  tls:
    cert_file: /path/to/cert.crt
    key_file: /path/to/key.key
```

### Multiple Proxies (Different Ports)

```yaml
# config/proxy/bash.yaml
proxy_config:
  port: 4000

# config/proxy/filesystem.yaml
proxy_config:
  port: 4001

# config/proxy/brave-search.yaml
proxy_config:
  port: 4002
```

---

## 🐛 Troubleshooting

### Server Won't Start

```
Error: api_key is required
```

**Fix:** Set API key in config or environment variable.

### Tools Not Discovered

**Check logs:**
```bash
./mcp-cli serve --verbose config/proxy/my-server.yaml 2>&1 | grep discovered
```

**Expected:**
```
Discovered 1 tools from bash
```

### OpenWebUI Shows 0 Tools

**Verify spec:**
```bash
curl http://localhost:4000/openapi.json | jq '.paths | keys'
```

### CORS Errors

Add your app's origin to `cors_origins`.

---

## 📈 Production Status

**✅ Fully Implemented and Tested:**

- Auto-discovery from MCP servers
- OpenAPI 3.0 spec generation
- Swagger UI documentation
- API key authentication
- CORS support
- TLS/HTTPS
- Health checks
- OpenWebUI integration

**Verified with:**
- OpenWebUI (fixed Issue #6427)
- cURL and other HTTP clients
- Swagger UI
- Web browsers

---

## 🚀 Next Steps

1. **Read the [Server Guide](proxy-server-guide.md)** for complete documentation
2. **Use the [Quick Reference](proxy-quick-reference.md)** for fast lookups
3. **Check example configs** in `/config/proxy/`
4. **Deploy to production** with the deployment guides

---

**Last Updated:** January 4, 2026  
**Status:** Production-ready ✅
