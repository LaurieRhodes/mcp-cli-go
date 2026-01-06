# Proxy Directory Structure

## Final Configuration

All proxy configs now follow the **same consistent pattern**:

```yaml
runas_type: proxy
version: "1.0"
config_source: config/servers/NAME.yaml      # For MCP servers
# OR
config_source: config/templates/NAME.yaml    # For workflow templates

proxy_config:
  port: XXXX
  api_key: "${NAME_PROXY_API_KEY}"
  cors_origins: ["*"]
```

## Files

### MCP Servers (Ports 8000-8999)

**filesystem.yaml**
- Config: `config/servers/filesystem.yaml`
- Port: 8080
- Key: `FILESYSTEM_PROXY_API_KEY`
- Tools: read_file, write_file, list_directory, etc.

**brave-search.yaml**
- Config: `config/servers/search.yaml`
- Port: 8081
- Key: `BRAVE_PROXY_API_KEY`
- Tools: brave_web_search, brave_local_search

### Workflow Templates (Ports 9000-9999)

**simple-analysis.yaml**
- Config: `config/templates/simple_analysis.yaml`
- Port: 9000
- Key: `SIMPLE_ANALYSIS_PROXY_API_KEY`
- Tool: simple_analysis (basic AI analysis)

**parallel-analysis.yaml**
- Config: `config/templates/parallel_analysis.yaml`
- Port: 9001
- Key: `PARALLEL_ANALYSIS_PROXY_API_KEY`
- Tool: parallel_analysis (parallel processing workflow)

**sentiment-analysis.yaml**
- Config: `config/templates/sentiment_analysis.yaml`
- Port: 9002
- Key: `SENTIMENT_ANALYSIS_PROXY_API_KEY`
- Tool: sentiment_analysis (sentiment analysis workflow)

**document-intelligence.yaml**
- Config: `config/templates/document_intelligence.yaml`
- Port: 9003
- Key: `DOCUMENT_INTELLIGENCE_PROXY_API_KEY`
- Tool: document_intelligence (document processing workflow)

### Skills (Auto-discovery)

**skills.yaml**
- Type: `proxy-skills` (auto-discovery)
- Port: 9010
- Key: `SKILLS_PROXY_API_KEY`
- Tools: Auto-discovered from `config/skills/`

## Pattern Benefits

✅ **Consistent** - Same pattern for servers and templates  
✅ **Explicit** - config_source clearly shows what's proxied  
✅ **Isolated** - One resource per proxy, one port per service  
✅ **Traceable** - Easy to follow: proxy → source → resource  
✅ **Scalable** - Simple to add new services  

## Quick Start

```bash
# Set environment variables
export FILESYSTEM_PROXY_API_KEY="key1"
export BRAVE_PROXY_API_KEY="key2"
export SIMPLE_ANALYSIS_PROXY_API_KEY="key3"
export SENTIMENT_ANALYSIS_PROXY_API_KEY="key4"
export SKILLS_PROXY_API_KEY="key5"

# Start services
mcp-cli serve config/proxy/filesystem.yaml &          # 8080
mcp-cli serve config/proxy/brave-search.yaml &        # 8081
mcp-cli serve config/proxy/simple-analysis.yaml &     # 9000
mcp-cli serve config/proxy/sentiment-analysis.yaml &  # 9002
mcp-cli serve config/proxy/skills.yaml &              # 9010
```

## Open WebUI Integration

Add each service individually:

1. **Filesystem Tools** - http://localhost:8080
2. **Search Tools** - http://localhost:8081
3. **Simple Analysis** - http://localhost:9000
4. **Sentiment Analysis** - http://localhost:9002
5. **Skills** - http://localhost:9010

Each with its own API key!
