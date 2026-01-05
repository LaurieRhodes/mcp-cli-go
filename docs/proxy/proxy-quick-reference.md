# HTTP Proxy Quick Reference

Fast lookup reference for the HTTP proxy server.

## ⚡ Quick Setup (2 Minutes)

### 1. Create Proxy Config

```yaml
# config/proxy/my-server.yaml
runas_type: proxy
version: "1.0"
config_source: config/servers/bash.yaml  # MCP server to proxy

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
  -d '{"command": "echo Hello"}'
```

---

## 📁 Files & Paths

```
config/
├── servers/
│   ├── bash.yaml           ← MCP server configs
│   ├── filesystem.yaml
│   └── brave-search.yaml
└── proxy/
    ├── bash.yaml           ← Proxy configs (point to servers/)
    ├── filesystem.yaml
    └── my-server.yaml
```

---

## 🔧 Configuration Options

### Minimal
```yaml
runas_type: proxy
version: "1.0"
config_source: config/servers/bash.yaml
proxy_config:
  api_key: "your-key"
```

### Common Options
```yaml
proxy_config:
  port: 4000                    # Default: 8080
  host: "0.0.0.0"              # Default: "0.0.0.0"
  api_key: "${MY_API_KEY}"     # Required!
  cors_origins: ["*"]          # Default: ["*"]
  enable_docs: true            # Default: true
  base_path: ""                # Default: "" (root)
```

### With TLS
```yaml
proxy_config:
  port: 443
  api_key: "${API_KEY}"
  tls:
    cert_file: /path/to/cert.crt
    key_file: /path/to/key.key
```

---

## 🌐 Endpoints

### Auto-Generated (from MCP tools)

```
POST /{tool_name}    # Execute tool
```

**Example:**
```
POST /bash           # bash tool
POST /read_file      # filesystem tool
POST /search         # brave-search tool
```

### Standard Endpoints

```
GET  /health         # Health check (no auth required)
GET  /tools          # List all tools (requires auth)
GET  /openapi.json   # OpenAPI specification (no auth)
GET  /docs           # Swagger UI (no auth, if enabled)
```

---

## 🔐 Authentication

### Format

```http
Authorization: Bearer your-api-key
```

or

```http
Authorization: your-api-key
```

### Generate Key

```bash
# Secure random key
openssl rand -hex 32

# UUID
uuidgen
```

### Set Key

```bash
# Environment variable
export MY_API_KEY="your-key"

# .env file
echo "MY_API_KEY=your-key" >> .env
```

---

## 📝 Example Requests

### Health Check (No Auth)

```bash
curl http://localhost:4000/health
```

**Response:**
```json
{
  "status": "healthy",
  "server": "bash-proxy",
  "version": "1.0.0",
  "tools": 1
}
```

### List Tools

```bash
curl -H "Authorization: Bearer $MY_API_KEY" \
  http://localhost:4000/tools
```

### Call Tool

```bash
curl -X POST http://localhost:4000/bash \
  -H "Authorization: Bearer $MY_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "command": "ls -la",
    "description": "List files"
  }'
```

**Response:**
```json
{
  "success": true,
  "tool": "bash",
  "result": "total 48\ndrwxr-xr-x  ..."
}
```

---

## 🐛 Troubleshooting

### Server won't start

```
Error: api_key is required
```

**Fix:** Set API key in config or environment.

### Tools not discovered

**Check logs:**
```bash
./mcp-cli serve --verbose config/proxy/my-server.yaml 2>&1 | grep discovered
```

**Expected:**
```
Discovered 1 tools from bash
```

### OpenWebUI shows 0 tools

**Verify OpenAPI spec:**
```bash
curl http://localhost:4000/openapi.json | jq '.paths | keys'
```

**Should show:**
```json
["/bash", "/health", "/tools"]
```

### CORS errors

```yaml
proxy_config:
  cors_origins:
    - "http://localhost:3000"  # Your app's origin
```

### Authentication fails

**Test:**
```bash
# Good key
curl -H "Authorization: Bearer correct-key" http://localhost:4000/health
# → 200 OK

# Bad key
curl -H "Authorization: Bearer wrong-key" http://localhost:4000/health
# → 401 Unauthorized
```

---

## 🚀 Testing Commands

### View OpenAPI Spec

```bash
curl http://localhost:4000/openapi.json | jq .
```

### List Paths

```bash
curl http://localhost:4000/openapi.json | jq '.paths | keys'
```

### Check Tool Schema

```bash
curl http://localhost:4000/openapi.json | \
  jq '.components.schemas.bashRequest'
```

### Test with Python

```python
import requests

url = "http://localhost:4000/bash"
headers = {
    "Authorization": "Bearer your-api-key",
    "Content-Type": "application/json"
}
data = {
    "command": "echo 'Hello from Python'",
    "description": "Test"
}

response = requests.post(url, headers=headers, json=data)
print(response.json())
```

---

## 📊 Common Patterns

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

**Start all:**
```bash
./mcp-cli serve config/proxy/bash.yaml &
./mcp-cli serve config/proxy/filesystem.yaml &
./mcp-cli serve config/proxy/brave-search.yaml &
```

### OpenWebUI Integration

**OpenWebUI Settings → Functions → Add Function:**

```
Name: MCP Bash
URL: http://localhost:4000/openapi.json
API Key: your-api-key
```

### Reverse Proxy (nginx)

```nginx
location /api/ {
    proxy_pass http://127.0.0.1:4000/;
    proxy_set_header Host $host;
    proxy_set_header Authorization $http_authorization;
}
```

**mcp-cli config:**
```yaml
proxy_config:
  host: "127.0.0.1"  # localhost only
  base_path: "/api"
```

---

## ✅ Production Checklist

- [ ] Strong API key (32+ bytes)
- [ ] Environment variable for API key
- [ ] Specific CORS origins (not `["*"]`)
- [ ] TLS/HTTPS enabled
- [ ] Reverse proxy configured
- [ ] Health monitoring setup
- [ ] Firewall rules configured
- [ ] Logs monitored
- [ ] Backup strategy
- [ ] Documentation for users

---

## 🔍 Debugging

### Verbose Logs

```bash
./mcp-cli serve --verbose config/proxy/my-server.yaml
```

### Check MCP Server Connection

```bash
# Should see these in logs:
# [INFO] Connected to bash v1.0.0
# [INFO] Discovered 1 tools from bash
```

### Validate Config

```bash
# Test config loads
./mcp-cli serve config/proxy/my-server.yaml --help
```

### Monitor Requests

```bash
# Watch access logs
tail -f /var/log/mcp-proxy/access.log
```

---

## 📚 Full Documentation

For complete details, see:
- **[HTTP Proxy Server Guide](proxy-server-guide.md)** - Complete documentation
- **[MCP Server Configs](../config/servers/)** - Available MCP servers

---

**Last Updated:** January 4, 2026  
**Tested With:** OpenWebUI, cURL, Swagger UI
