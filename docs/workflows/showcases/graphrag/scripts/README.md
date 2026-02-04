# GraphRAG Helper Scripts

Bash wrapper scripts for easy GraphRAG usage.

## Scripts

### ask_graphrag_direct.sh
Query the knowledge graph and get results with source text.

**Usage:**
```bash
./ask_graphrag_direct.sh "your question"
```

**What it does:**
1. Calls query_graphrag.py
2. Displays summary statistics
3. Shows matching entities
4. Shows relevant chunks
5. Displays top 3 chunks with original ISM text

**Example:**
```bash
./ask_graphrag_direct.sh "password length"

# Returns:
# - 6 matching entities
# - 6 related entities
# - 6 relevant chunks
# - Top 3 chunks with full text
```

### explore_graphrag.sh
Explore the knowledge graph structure.

**Usage:**
```bash
# Show graph statistics
./explore_graphrag.sh stats

# Find most connected entities
./explore_graphrag.sh top 20

# List entities by type
./explore_graphrag.sh list CONCEPT
./explore_graphrag.sh list CONTROL
./explore_graphrag.sh list ACTOR
```

**Output:**
- `stats`: Total nodes, edges, entity types, relationship types
- `top N`: N most connected entities (graph hubs)
- `list TYPE`: All entities of a specific type

### validate_extraction.sh
Validate entity extraction against original source text.

**Usage:**
```bash
./validate_extraction.sh CHUNK-305
```

**Shows:**
1. Original ISM policy text (word-for-word)
2. Extracted entities with types and descriptions
3. Extracted relationships
4. Validation checklist
5. Extraction statistics (entities/word, etc.)

**Use for:**
- Quality assurance
- Debugging extraction
- Understanding entity types
- Verifying accuracy

## Installation

```bash
# Make scripts executable
chmod +x *.sh

# Verify they work
./ask_graphrag_direct.sh "test"
```

## Configuration

Scripts assume data is in:
```
/tmp/mcp-outputs/rlm_poc/
├── chunks.json
├── knowledge_graph.json
└── graph_chunks/
    └── entities_*.json
```

To change paths, edit variables at top of each script:
```bash
GRAPH_FILE="/tmp/mcp-outputs/rlm_poc/knowledge_graph.json"
CHUNKS_FILE="/tmp/mcp-outputs/rlm_poc/chunks.json"
ENTITIES_DIR="/tmp/mcp-outputs/rlm_poc/graph_chunks"
```

## See Also

- [../skills/README.md](../skills/README.md) - Python scripts called by these wrappers
- [../USAGE.md](../USAGE.md) - Complete usage guide
- [../examples/README.md](../examples/README.md) - Example queries
