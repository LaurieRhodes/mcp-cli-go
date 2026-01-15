# RAG Usage Guide

## Basic Commands

### Search

```bash
# Simple search
mcp-cli rag search "authentication requirements"

# Search with more results
mcp-cli rag search "access control policies" --top-k 10

# Use specific server
mcp-cli rag search "encryption" --server pgvector
```

### View Configuration

```bash
# Show RAG configuration
mcp-cli rag config

# Show detailed configuration
mcp-cli rag config --verbose
```

## Command-Line Options

### --top-k

Number of results to return:

```bash
# Get 3 results
mcp-cli rag search "password policies" --top-k 3

# Get 20 results
mcp-cli rag search "security controls" --top-k 20
```

**Default**: 5 results

### --server

Specify which RAG server to use:

```bash
# Use specific server
mcp-cli rag search "query" --server pgvector

# Use alternative server
mcp-cli rag search "query" --server vectordb2
```

**Default**: From `default_server` in settings.yaml

### --strategies

Specify which search strategies to use:

```bash
# Single strategy
mcp-cli rag search "authentication" --strategies default

# Multiple strategies (comma-separated)
mcp-cli rag search "authentication" --strategies default,context

# All available strategies
mcp-cli rag search "authentication" --strategies default,context,technical
```

**Default**: `[default]`

### --fusion

Combine results from multiple strategies:

```bash
# Reciprocal Rank Fusion (recommended)
mcp-cli rag search "query" --strategies default,context --fusion rrf

# Weighted fusion
mcp-cli rag search "query" --strategies default,context --fusion weighted

# Maximum score
mcp-cli rag search "query" --strategies default,context --fusion max

# Average score
mcp-cli rag search "query" --strategies default,context --fusion avg
```

**Default**: From `default_fusion` in settings.yaml (usually `rrf`)

### --expand

Enable query expansion:

```bash
# Expand "MFA" to include "multi-factor authentication", "two-factor", etc.
mcp-cli rag search "MFA" --expand

# Expand acronyms and synonyms
mcp-cli rag search "auth" --expand
```

**Default**: Disabled

## Examples by Use Case

### Finding Security Controls

```bash
# Find authentication requirements
mcp-cli rag search "multi-factor authentication requirements"

# Find access control policies
mcp-cli rag search "role-based access control implementation"

# Find encryption standards
mcp-cli rag search "data encryption at rest and in transit"
```

### Compliance Research

```bash
# Search multiple strategies with fusion
mcp-cli rag search "GDPR compliance requirements" \
  --strategies default,context \
  --fusion rrf \
  --top-k 10

# Expand query for comprehensive results
mcp-cli rag search "privacy controls" --expand --top-k 15
```

### Technical Documentation

```bash
# Search with specific server
mcp-cli rag search "API authentication methods" \
  --server technical_docs \
  --top-k 5

# High-precision search
mcp-cli rag search "OAuth 2.0 implementation" \
  --strategies technical \
  --top-k 3
```

### Exploratory Search

```bash
# Broad search with many results
mcp-cli rag search "security best practices" \
  --top-k 20 \
  --expand

# Multi-strategy comprehensive search
mcp-cli rag search "incident response procedures" \
  --strategies default,context,technical \
  --fusion rrf \
  --top-k 15
```

## Understanding Results

### Output Format

```json
{
  "query": "authentication requirements",
  "results": [
    {
      "id": "555",
      "text": {
        "identifier": "ISM-1680",
        "description": "Multi-factor authentication...",
        "guideline": "Guidelines for system hardening",
        "section": "Authentication hardening",
        "topic": "Multi-factor authentication"
      },
      "metadata": {
        "id": 555,
        "index": 3,
        "revision": 1,
        "source": "description_vector"
      },
      "combined_score": 0.015625,
      "component_scores": {
        "description_vector": 0.501537
      },
      "source": "description_vector"
    }
  ],
  "fusion": "rrf",
  "total_results": 5,
  "execution_time_ms": 1464
}
```

### Key Fields

- **query**: Your search query
- **results**: Array of matched documents
  - **id**: Unique document identifier
  - **text**: Document content (all text columns)
  - **metadata**: Additional information
  - **combined_score**: Final relevance score
  - **component_scores**: Scores per strategy
  - **source**: Which vector column matched
- **total_results**: Number of results returned
- **execution_time_ms**: Search duration

### Interpreting Scores

**Similarity Scores** (component_scores):
- **0.9 - 1.0**: Nearly identical
- **0.7 - 0.9**: Very similar
- **0.5 - 0.7**: Related/relevant (typical good matches)
- **0.3 - 0.5**: Somewhat related
- **< 0.3**: Weak similarity

**Combined Scores** (with fusion):
- Normalized across strategies
- Higher = more relevant
- Depends on fusion method

## Advanced Usage

### Piping Results

```bash
# Extract identifiers only
mcp-cli rag search "authentication" | jq -r '.results[].text.identifier'

# Get top result description
mcp-cli rag search "encryption" --top-k 1 | jq -r '.results[0].text.description'

# Save results to file
mcp-cli rag search "security controls" --top-k 20 > results.json
```

### Batch Queries

```bash
# Search multiple queries
for query in "authentication" "authorization" "encryption"; do
  echo "=== $query ==="
  mcp-cli rag search "$query" --top-k 3
  echo ""
done

# Save all results
for query in "auth" "access" "crypto"; do
  mcp-cli rag search "$query" > "results_${query}.json"
done
```

### Comparison Search

```bash
# Compare results from different strategies
mcp-cli rag search "authentication" --strategies default > default.json
mcp-cli rag search "authentication" --strategies context > context.json

# Compare fusion methods
mcp-cli rag search "auth" --fusion rrf > rrf.json
mcp-cli rag search "auth" --fusion weighted > weighted.json
```

## Performance Tips

### Optimize Result Count

```bash
# Quick lookup - fewer results
mcp-cli rag search "specific control ISM-1234" --top-k 1

# Comprehensive research - more results
mcp-cli rag search "broad security topic" --top-k 20
```

### Strategy Selection

```bash
# Fast - single strategy
mcp-cli rag search "query" --strategies default

# Thorough - multiple strategies
mcp-cli rag search "query" --strategies default,context,technical
```

### Query Specificity

```bash
# Specific queries = faster, more precise
mcp-cli rag search "ISM-1680 multi-factor authentication"

# Broad queries = slower, more results
mcp-cli rag search "security"
```

## Debug Mode

```bash
# Show detailed logging
mcp-cli rag search "authentication" --log-level debug

# Show trace logging (very verbose)
mcp-cli rag search "authentication" --log-level verbose
```

## Next Steps

- [Build Workflows](workflows.md) - Use RAG in multi-step workflows
- [Configuration](configuration.md) - Customize RAG behavior
- [Troubleshooting](troubleshooting.md) - Fix common issues
