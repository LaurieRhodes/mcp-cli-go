# RAG (Retrieval-Augmented Generation)

RAG enables semantic search across vector databases connected via MCP. It combines the power of vector embeddings with intelligent retrieval to find the most relevant information from your data stores.

## Overview

**RAG** allows mcp-cli to:
- Search vector databases using semantic similarity
- Retrieve relevant documents based on meaning, not just keywords
- Support multiple search strategies and fusion methods
- Integrate seamlessly with workflows for AI-powered applications

## Quick Start

```bash
# Search for documents
mcp-cli rag search "authentication requirements"

# Use in a workflow
mcp-cli --workflow examples/rag/search.yaml --var query="your search here"
```

## Key Features

- **Semantic Search**: Find documents by meaning, not just exact matches
- **Multiple Strategies**: Search across different vector columns (default, context, technical)
- **Result Fusion**: Combine results using RRF, weighted, max, or avg methods
- **Query Expansion**: Automatically expand queries with synonyms and related terms
- **Workflow Integration**: Use RAG in multi-step workflows
- **OpenAI Embeddings**: Powered by text-embedding-3-small (1536 dimensions)

## Documentation

- [Configuration Guide](configuration.md) - Set up RAG servers and strategies
- [Usage Guide](usage.md) - Command-line usage and examples
- [Workflows Guide](workflows.md) - Using RAG in workflows
- [Troubleshooting](troubleshooting.md) - Common issues and solutions

## Architecture

```
mcp-cli
  ├── Embedding Service (OpenAI API)
  │   └── Generates query embeddings
  ├── RAG Service
  │   └── Orchestrates search and fusion
  └── MCP Server (pgvector)
      └── Vector database search
```

## Prerequisites

- MCP vector server (e.g., mcp-pgvector-go)
- OpenAI API key (set as `OPENAI_API_KEY` environment variable)
- Populated vector database with embeddings

## Next Steps

1. [Configure RAG](configuration.md) - Set up your RAG server
2. [Learn Usage](usage.md) - Master RAG commands
3. [Build Workflows](workflows.md) - Create RAG-powered applications
