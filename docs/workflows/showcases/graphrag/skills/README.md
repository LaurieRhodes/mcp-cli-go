# GraphRAG Skills (Python Scripts)

Python scripts for entity extraction, graph building, and querying.

## Scripts

### extract_graph_entities.py
Extracts entities and relationships from a document chunk.

**Input:**
- Chunk content (text)
- Chunk ID

**Output:**
```json
{
  "chunk_id": "CHUNK-305",
  "entities": [
    {
      "id": "passphrases",
      "type": "CONCEPT",
      "text": "Random word sequences for authentication"
    }
  ],
  "relationships": [
    {
      "from": "passphrases",
      "type": "ALTERNATIVE_TO",
      "to": "passwords"
    }
  ]
}
```

**Called by:** extract_graph_chunk workflows

### query_graphrag.py
Query the knowledge graph for matching entities and retrieve source chunks.

**Usage:**
```bash
./query_graphrag.py \
  "password length" \
  "/tmp/mcp-outputs/rlm_poc/knowledge_graph.json" \
  "/tmp/mcp-outputs/rlm_poc/graph_chunks" \
  "/tmp/mcp-outputs/rlm_poc/query_results.json" \
  2  # traversal depth
```

**Process:**
1. Search graph for matching entities
2. Multi-hop traversal (configurable depth)
3. Find chunks containing entities
4. Return results with source text

**Output:**
```json
{
  "query": "password length",
  "status": "success",
  "matching_entities": 6,
  "total_related_entities": 6,
  "relevant_chunks": 6,
  "chunks": [...],
  "subgraph": {...}
}
```

### explore_graph.py
Explore knowledge graph structure and statistics.

**Commands:**
```bash
# Show statistics
./explore_graph.py stats graph.json output.json

# Find most connected entities
./explore_graph.py top graph.json output.json 20

# List entities by type
./explore_graph.py list graph.json output.json CONCEPT
```

**Use cases:**
- Understanding graph structure
- Finding key entities
- Discovering entity types
- Quality assessment

## Helper Functions

**Common patterns across scripts:**

```python
# Load knowledge graph
def load_graph(graph_file):
    with open(graph_file, 'r') as f:
        return json.load(f)

# Search entities
def search_entities(graph, query_terms):
    matches = []
    for node in graph['nodes']:
        if query in node['text'].lower():
            matches.append(node)
    return matches

# Find related entities (multi-hop)
def get_related_entities(graph, entity_ids, depth=2):
    related = set(entity_ids)
    for _ in range(depth):
        new_entities = set()
        for edge in graph['edges']:
            if edge['from'] in related:
                new_entities.add(edge['to'])
            if edge['to'] in related:
                new_entities.add(edge['from'])
        related.update(new_entities)
    return list(related)
```

## Dependencies

```python
# Standard library only
import json
import sys
import os
from pathlib import Path
from collections import Counter
```

**No external dependencies required!**

## See Also

- [../workflows/README.md](../workflows/README.md) - Workflow definitions
- [../ARCHITECTURE.md](../ARCHITECTURE.md) - System architecture
- [../examples/README.md](../examples/README.md) - Example usage
