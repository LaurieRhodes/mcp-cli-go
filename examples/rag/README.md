# RAG Workflow Examples

This directory contains example workflows demonstrating RAG (Retrieval-Augmented Generation) capabilities in mcp-cli.

## Prerequisites

1. **MCP Vector Server**: A connected MCP server with vector search capabilities (e.g., mcp-pgvector-go)
2. **RAG Configuration**: `config/rag.yaml` configured with your server and strategies
3. **Embeddings**: Data already embedded and stored in your vector database

## Configuration

Copy and customize the example RAG configuration:

```bash
cp config/rag.yaml.example config/rag.yaml
# Edit config/rag.yaml with your server details
```

## Examples

### 1. Simple Search (`simple-search.yaml`)

Basic RAG search using a single embedding strategy.

**Run:**
```bash
mcp-cli workflow run examples/rag/simple-search.yaml \
  --var user_query="What are the MFA requirements?"
```

**What it does:**
- Retrieves top 5 results using default strategy
- Generates answer based on retrieved documents

### 2. Multi-Strategy Search (`multi-strategy-search.yaml`)

Search across multiple embedding strategies with RRF fusion.

**Run:**
```bash
mcp-cli workflow run examples/rag/multi-strategy-search.yaml \
  --var user_query="Explain multi-factor authentication controls"
```

**What it does:**
- Searches using default, technical, and compliance strategies simultaneously
- Fuses results using Reciprocal Rank Fusion (RRF)
- Analyzes results from multiple perspectives

### 3. Advanced Search (`advanced-search.yaml`)

Advanced RAG with query expansion and iterative refinement.

**Run:**
```bash
mcp-cli workflow run examples/rag/advanced-search.yaml \
  --var user_query="What encryption standards apply to data at rest?"
```

**What it does:**
- Expands query with synonyms and acronyms
- Evaluates initial results
- Performs refined search if needed
- Synthesizes comprehensive answer

## RAG Configuration Reference

### Basic Structure

```yaml
rag:
  default_server: pgvector
  default_fusion: rrf
  default_top_k: 5
  
  servers:
    pgvector:
      mcp_server: pgvector  # References MCP server from main config
      strategies:
        - name: default
          vector_column: embedding_default
          weight: 1.0
          threshold: 0.7
```

### Workflow RAG Step

```yaml
steps:
  - name: retrieve
    rag:
      query: "{{user_query}}"      # Query text
      server: pgvector               # RAG server name
      strategies: [default, technical]  # Strategies to use
      fusion: rrf                    # rrf, weighted, max, avg
      top_k: 5                       # Results to return
      expand_query: true             # Enable query expansion
```

## Fusion Methods

### Reciprocal Rank Fusion (RRF)
- **Best for**: Combining results from diverse strategies
- **How it works**: Position-based scoring (higher rank = higher score)
- **Use when**: Strategies have different ranking characteristics

### Weighted
- **Best for**: Trusting certain strategies more than others
- **How it works**: Multiplies similarity scores by strategy weights
- **Use when**: You know some strategies are more reliable

### Max
- **Best for**: Finding the single best match
- **How it works**: Takes highest score across all strategies
- **Use when**: You want the absolute best result

### Average
- **Best for**: Balanced results
- **How it works**: Averages scores across strategies
- **Use when**: All strategies equally reliable

## Query Expansion

Enable query expansion to automatically expand queries with:
- **Synonyms**: "authentication" → "auth", "verification"
- **Acronyms**: "MFA" → "Multi-Factor Authentication"
- **Domain terms**: Technical vocabulary mapping

Configure in `config/rag.yaml`:
```yaml
query_expansion:
  enabled: true
  synonyms_file: config/rag/synonyms.yaml
  acronyms_file: config/rag/acronyms.yaml
  max_expansions: 5
```

## Multi-Server Fusion (Future)

Coming soon: Search multiple vector databases simultaneously:

```yaml
steps:
  - name: hybrid_search
    rag:
      query: "{{user_query}}"
      servers:
        - pgvector:
            strategies: [default, technical]
            weight: 0.6
        - pinecone:
            strategies: [default]
            weight: 0.4
      fusion: rrf
```

## Tips

1. **Start Simple**: Use `simple-search.yaml` first to verify your setup
2. **Tune Thresholds**: Adjust similarity thresholds in config for your data
3. **Experiment with Fusion**: Try different fusion methods for your use case
4. **Monitor Results**: Use `--verbose` to see strategy-specific scores
5. **Iterate**: Use advanced workflows for complex queries

## Troubleshooting

### No Results
- Check similarity thresholds (lower threshold = more results)
- Verify embeddings exist for your data
- Try query expansion

### Poor Quality Results
- Adjust strategy weights
- Try different fusion methods
- Refine your embedding strategy configuration

### Server Not Found
```bash
# Check RAG configuration
mcp-cli rag config

# Verify MCP server connection
mcp-cli servers list
```

## See Also

- [RAG Configuration Reference](../../docs/rag/configuration.md)
- [Query Expansion Guide](../../docs/rag/query-expansion.md)
- [Multi-Strategy Best Practices](../../docs/rag/multi-strategy.md)
