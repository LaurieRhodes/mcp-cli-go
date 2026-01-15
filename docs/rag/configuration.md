# RAG Configuration Guide

## Configuration Files

RAG is configured through two main files:

1. **`config/settings.yaml`** - Global RAG defaults
2. **`config/rag/*.yaml`** - Server-specific configurations

## Global Settings (settings.yaml)

```yaml
rag:
  default_server: pgvector      # Default server for RAG queries
  default_fusion: rrf            # Default fusion method (rrf, weighted, max, avg)
  default_top_k: 5              # Default number of results
  query_expansion:
    enabled: true               # Enable query expansion
    max_expansions: 5          # Maximum query variations
    synonyms_file: ""          # Optional synonyms file
    acronyms_file: ""          # Optional acronyms file
```

## Server Configuration (config/rag/pgvector.yaml)

```yaml
server_name: pgvector
config:
  # MCP server details
  mcp_server: pgvector          # MCP server name (from config/servers/)
  search_tool: search_vectors   # Tool name for vector search
  table: ism_controls_vectors   # Target table name
  
  # Embedding configuration
  query_embedding:
    type: service               # Use mcp-cli's embedding service
    provider: openai           # Embedding provider
    model: text-embedding-3-small  # Model name (1536 dimensions)
  
  # Search strategies
  strategies:
    - name: default
      vector_column: description_vector
      weight: 1.0
      threshold: 0.7
      
    - name: context
      vector_column: combined_context_vector
      weight: 1.0
      threshold: 0.7
  
  # Table schema
  text_columns:
    - identifier
    - description
    - guideline
    - section
    - topic
  
  metadata_columns:
    - revision
    - updated
```

## Embedding Configuration

### Type: service (Recommended)

Uses mcp-cli's internal embedding service with your OpenAI API key:

```yaml
query_embedding:
  type: service
  provider: openai
  model: text-embedding-3-small
```

**Requirements:**
- Set `OPENAI_API_KEY` environment variable
- Configure `config/embeddings/openai.yaml`

### Type: mcp_tool (Alternative)

Calls the MCP server's embedding tool:

```yaml
query_embedding:
  type: mcp_tool
  tool_name: text_to_embeddings
  default_params:
    provider: openai
    model: text-embedding-3-small
    normalize: false
```

**Requirements:**
- MCP server must support embedding generation
- API key configured in MCP server's config

## Search Strategies

Strategies define which vector columns to search:

```yaml
strategies:
  - name: default              # Strategy name
    vector_column: description_vector  # Column to search
    weight: 1.0               # Relative importance (for fusion)
    threshold: 0.7            # Minimum similarity score
```

### Common Strategies

- **default**: Primary semantic search (description/content)
- **context**: Contextual information (metadata, tags)
- **technical**: Technical terms and identifiers
- **semantic**: General semantic understanding

## Fusion Methods

Combine results from multiple strategies:

### RRF (Reciprocal Rank Fusion)
```yaml
default_fusion: rrf
```
Balances diverse ranking signals. Good default choice.

### Weighted
```yaml
default_fusion: weighted
```
Uses strategy weights. Better when some columns are more important.

### Max
```yaml
default_fusion: max
```
Takes the best score for each result.

### Average
```yaml
default_fusion: avg
```
Averages scores across strategies.

## Complete Example

**config/settings.yaml:**
```yaml
rag:
  default_server: pgvector
  default_fusion: rrf
  default_top_k: 5
  query_expansion:
    enabled: true
    max_expansions: 5
```

**config/rag/pgvector.yaml:**
```yaml
server_name: pgvector
config:
  mcp_server: pgvector
  search_tool: search_vectors
  table: documents_vectors
  
  query_embedding:
    type: service
    provider: openai
    model: text-embedding-3-small
  
  strategies:
    - name: default
      vector_column: content_vector
      weight: 1.0
      threshold: 0.7
    
    - name: metadata
      vector_column: metadata_vector
      weight: 0.5
      threshold: 0.6
  
  text_columns:
    - title
    - content
    - summary
  
  metadata_columns:
    - created
    - author
    - category
```

## Verification

Check your configuration:

```bash
# View configuration
mcp-cli rag config

# View with details
mcp-cli rag config --verbose

# Test search
mcp-cli rag search "test query"
```

## Environment Variables

Required:
```bash
export OPENAI_API_KEY="sk-your-key-here"
```

Optional:
```bash
export RAG_DEBUG=true          # Enable debug logging
export RAG_CACHE_TTL=300       # Cache results (seconds)
```

## Next Steps

- [Learn Usage](usage.md) - Command-line examples
- [Build Workflows](workflows.md) - RAG in workflows
- [Troubleshooting](troubleshooting.md) - Fix common issues
