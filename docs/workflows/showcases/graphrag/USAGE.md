# GraphRAG Usage Guide

Complete guide to using the GraphRAG system for querying the Australian ISM policy knowledge graph.

## Quick Start

```bash
# Navigate to the showcase directory
cd docs/workflows/showcases/graphrag

# Query the knowledge graph
./scripts/ask_graphrag_direct.sh "password requirements"

# Explore graph structure  
./scripts/explore_graphrag.sh stats

# Validate extraction quality
./scripts/validate_extraction.sh CHUNK-305
```

## Installation & Setup

### Prerequisites

- Python 3.8+
- jq (JSON processor)
- 50MB free disk space
- DeepSeek API key (only for building, not querying)

### Data Files Required

The system needs these data files (should be in `/tmp/mcp-outputs/rlm_poc/`):

```
/tmp/mcp-outputs/rlm_poc/
├── chunks.json                    # Original ISM chunks (Layer 1)
├── knowledge_graph.json           # Knowledge graph (Layer 3)
└── graph_chunks/                  # Entity files (Layer 2)
    ├── entities_CHUNK-001.json
    ├── entities_CHUNK-002.json
    └── ... (531 files total)
```

If you need to rebuild from scratch, see [Building the Graph](#building-the-graph).

## Querying the Knowledge Graph

### Method 1: Direct Query (Recommended)

Fast, no LLM needed, returns graph results with source text.

```bash
./scripts/ask_graphrag_direct.sh "your question"
```

**Examples:**

```bash
# Broad security topics
./scripts/ask_graphrag_direct.sh "cyber security"
# Returns: 586 entities, 1390 related, 487 chunks

# Specific requirements
./scripts/ask_graphrag_direct.sh "password length"
# Returns: 6 entities, 6 related, 6 chunks

# Technical queries
./scripts/ask_graphrag_direct.sh "email encryption"
# Returns: 93 entities, 375 related, 325 chunks

# ISM controls
./scripts/ask_graphrag_direct.sh "ISM-0417"
# Returns: Exact control with full context
```

**Output:**
```
✅ Query successful!

Summary:
  matching_entities: 6
  total_related: 6
  chunks_found: 6

Matching Entities:
  • [CONCEPT] passphrases
  • [CONCEPT] passwords
  • [CONCEPT] multi-factor authentication

Relevant Chunks (first 5):
  • CHUNK-305: 9 entities
  • CHUNK-298: 10 entities
  • CHUNK-309: 10 entities

Top 3 Chunks with Original Text:
[CHUNK-305]
A significant threat to the compromise of accounts is
credential cracking tools. When an adversary gains access...
```

### Method 2: Programmatic Access

Use Python script directly for automation:

```bash
cd /media/laurie/Data/Github/mcp-cli-go

./config/skills/python-context-builder/scripts/query_graphrag.py \
  "password length" \
  "/tmp/mcp-outputs/rlm_poc/knowledge_graph.json" \
  "/tmp/mcp-outputs/rlm_poc/graph_chunks" \
  "/tmp/mcp-outputs/rlm_poc/query_results.json" \
  2  # traversal depth
```

**Parameters:**
1. Query string
2. Path to knowledge graph
3. Path to entity files directory
4. Output file path
5. Traversal depth (1-3, default 2)

**Output:** JSON file with complete results

```json
{
  "query": "password length",
  "status": "success",
  "matching_entities": [...],
  "total_related_entities": 6,
  "relevant_chunks": 6,
  "chunks": [...],
  "subgraph": {...}
}
```

## Exploring the Graph

### Show Graph Statistics

```bash
./scripts/explore_graphrag.sh stats
```

**Output:**
```json
{
  "total_nodes": 2337,
  "total_edges": 2313,
  "entity_types": {
    "CONCEPT": 1115,
    "PROCESS": 390,
    "ACTOR": 173,
    "CONTROL": 119,
    ...
  },
  "relationship_types": {
    "REQUIRES": 276,
    "DEFINES": 271,
    "RELATED_TO": 166,
    ...
  }
}
```

### Find Most Connected Entities

```bash
./scripts/explore_graphrag.sh top 20
```

Shows the 20 most connected entities (graph hubs):

```
50  CONCEPT    multi-factor authentication
45  CONCEPT    encryption
42  ACTOR      organisation
38  CONCEPT    cyber security incidents
...
```

These are typically:
- Core security concepts
- Frequently referenced controls
- Key organizational entities

### List Entities by Type

```bash
./scripts/explore_graphrag.sh list CONCEPT
./scripts/explore_graphrag.sh list CONTROL
./scripts/explore_graphrag.sh list ACTOR
```

**Output:**
```
Entities of type CONCEPT (showing first 50):
multi-factor authentication    Authentication requiring multiple factors
encryption                     Data protection through cryptography
passphrases                   Random word sequences for authentication
...
```

## Validating Extraction Quality

### Validate Specific Chunk

```bash
./scripts/validate_extraction.sh CHUNK-305
```

**Shows:**
1. Original ISM policy text (word-for-word)
2. Extracted entities with types
3. Extracted relationships
4. Validation checklist
5. Extraction statistics

**Output:**
```
═══════════════════════════════════════════════════════════════════
📄 ORIGINAL ISM POLICY TEXT (Source of Truth):
═══════════════════════════════════════════════════════════════════

A significant threat to the compromise of accounts is credential
cracking tools. When an adversary gains access...

═══════════════════════════════════════════════════════════════════
🔍 EXTRACTED ENTITIES:
═══════════════════════════════════════════════════════════════════

[CONCEPT] credential cracking tools
  ➜ Tools used to crack credentials

[CONCEPT] multi-factor authentication
  ➜ Authentication requiring multiple factors
...

═══════════════════════════════════════════════════════════════════
✅ VALIDATION CHECKLIST:
═══════════════════════════════════════════════════════════════════

☑ Are all key concepts captured as entities?
☑ Are entity types appropriate?
☑ Do entity descriptions match the source text meaning?
☑ Are relationships between entities accurate?
...
```

### Validate Random Sample

```bash
# Validate 10 random chunks
for i in {1..10}; do
  CHUNK="CHUNK-$(printf "%03d" $((RANDOM % 531)))"
  ./scripts/validate_extraction.sh $CHUNK
  read -p "Press Enter to continue..."
done
```

## Advanced Usage

### Adjust Graph Traversal Depth

**1-hop (Direct connections only):**
```bash
# Modify query script or use Python directly
./config/skills/python-context-builder/scripts/query_graphrag.py \
  "password" \
  "/tmp/mcp-outputs/rlm_poc/knowledge_graph.json" \
  "/tmp/mcp-outputs/rlm_poc/graph_chunks" \
  "/tmp/output.json" \
  1  # 1-hop only
```

**2-hop (Default, Balanced):**
```bash
# Default in scripts
./scripts/ask_graphrag_direct.sh "password"
```

**3-hop (Extended network):**
```bash
./config/skills/python-context-builder/scripts/query_graphrag.py \
  "password" \
  "/tmp/mcp-outputs/rlm_poc/knowledge_graph.json" \
  "/tmp/mcp-outputs/rlm_poc/graph_chunks" \
  "/tmp/output.json" \
  3  # 3-hop extended
```

### Search Original Source Text

```bash
# Find all chunks mentioning "passphrases"
cat /tmp/mcp-outputs/rlm_poc/chunks.json | \
  jq '.[] | select(.content | contains("passphrases")) | .chunk_id'

# Get full text of specific chunk
cat /tmp/mcp-outputs/rlm_poc/chunks.json | \
  jq '.[] | select(.chunk_id == "CHUNK-305") | .content'
```

### Inspect Entity Files

```bash
# View entities extracted from CHUNK-305
cat /tmp/mcp-outputs/rlm_poc/graph_chunks/entities_CHUNK-305.json | jq '.'

# List all entity types in a chunk
cat /tmp/mcp-outputs/rlm_poc/graph_chunks/entities_CHUNK-305.json | \
  jq '.entities[].type' | sort -u
```

### Export Subgraph

```bash
# Query and extract subgraph for a topic
./config/skills/python-context-builder/scripts/query_graphrag.py \
  "multi-factor authentication" \
  "/tmp/mcp-outputs/rlm_poc/knowledge_graph.json" \
  "/tmp/mcp-outputs/rlm_poc/graph_chunks" \
  "/tmp/mfa_subgraph.json" \
  2

# View just the subgraph
cat /tmp/mfa_subgraph.json | jq '.subgraph'
```

## Building the Graph

If you need to rebuild the knowledge graph from scratch:

### Step 1: Prepare Source Document

```bash
# Place ISM PDF in source directory
cp /path/to/ism-document.pdf docs/workflows/showcases/graphrag/source/

# Chunk the document (separate process)
# Creates: chunks.json with 531 chunks
```

### Step 2: Extract Entities

```bash
cd /media/laurie/Data/Github/mcp-cli-go

# Run entity extraction workflow
./mcp-cli --workflow rlm_poc/workflows/production_graphrag_531 \
  < chunks.json

# Takes ~4 minutes with 20 workers
# Creates: 531 entity files in graph_chunks/
```

### Step 3: Build Knowledge Graph

```bash
# Run graph building workflow
./mcp-cli --workflow rlm_poc/workflows/rebuild_graph

# Takes ~30 seconds
# Creates: knowledge_graph.json
```

## Integration Examples

### Python Integration

```python
import json

# Load knowledge graph
with open('/tmp/mcp-outputs/rlm_poc/knowledge_graph.json') as f:
    graph = json.load(f)

# Find entities
def find_entities(graph, search_term):
    matches = []
    for node in graph['nodes']:
        if search_term.lower() in node['text'].lower():
            matches.append(node)
    return matches

# Find relationships
def find_relationships(graph, entity_id):
    relationships = []
    for edge in graph['edges']:
        if edge['from'] == entity_id or edge['to'] == entity_id:
            relationships.append(edge)
    return relationships

# Example usage
entities = find_entities(graph, "password")
for entity in entities[:5]:
    print(f"{entity['type']}: {entity['text']}")
```

### Bash Integration

```bash
#!/bin/bash
# Query GraphRAG and process results

QUERY="password requirements"

# Run query
./scripts/ask_graphrag_direct.sh "$QUERY" > /tmp/results.txt

# Extract chunk IDs
CHUNKS=$(cat /tmp/mcp-outputs/rlm_poc/query_results.json | \
  jq -r '.chunks[].chunk_id')

# Process each chunk
for CHUNK in $CHUNKS; do
  echo "Processing $CHUNK..."
  # Your processing logic here
done
```

### REST API Wrapper (Future)

```python
from flask import Flask, jsonify, request
import subprocess

app = Flask(__name__)

@app.route('/query', methods=['POST'])
def query_graphrag():
    question = request.json.get('question')
    
    # Run query script
    result = subprocess.run(
        ['./scripts/ask_graphrag_direct.sh', question],
        capture_output=True,
        text=True
    )
    
    # Parse JSON results
    with open('/tmp/mcp-outputs/rlm_poc/query_results.json') as f:
        data = json.load(f)
    
    return jsonify(data)

if __name__ == '__main__':
    app.run(port=5000)
```

## Troubleshooting

### No Results Found

**Problem:** Query returns 0 entities

**Solutions:**
1. Try broader search terms
2. Check spelling
3. List available entity types: `./scripts/explore_graphrag.sh stats`
4. Browse entities: `./scripts/explore_graphrag.sh list CONCEPT`

### Too Many Results

**Problem:** Query returns 500+ entities

**Solutions:**
1. Use more specific terms
2. Combine multiple concepts: "password AND length"
3. Reduce traversal depth to 1-hop
4. Filter by entity type

### Script Not Found

**Problem:** `./scripts/ask_graphrag_direct.sh: No such file or directory`

**Solutions:**
```bash
# Make sure you're in the right directory
cd docs/workflows/showcases/graphrag

# Make scripts executable
chmod +x scripts/*.sh

# Check files exist
ls -l scripts/
```

### Query Results File Missing

**Problem:** `/tmp/mcp-outputs/rlm_poc/query_results.json not found`

**Solutions:**
1. Run query script first
2. Check /tmp/mcp-outputs/rlm_poc/ directory exists
3. Verify permissions: `ls -la /tmp/mcp-outputs/`

## Tips & Best Practices

### Query Tips

1. **Start Broad, Then Narrow**
   - Begin with general terms
   - Drill down based on results
   - Use entity types to filter

2. **Use Multiple Queries**
   - Different phrasings reveal different connections
   - Combine results for comprehensive view
   - Example: "password" + "authentication" + "credentials"

3. **Follow the Graph**
   - Let relationships guide exploration
   - Check related entities
   - Traverse connections

4. **Validate Important Results**
   - Always check source text for critical decisions
   - Use validation script: `./scripts/validate_extraction.sh`
   - Trust but verify

### Performance Tips

1. **Cache Common Queries**
   ```bash
   # Save frequent query results
   ./scripts/ask_graphrag_direct.sh "password" > password_query_cache.txt
   ```

2. **Limit Result Sets**
   - Modify scripts to limit results
   - Reduce traversal depth for faster queries
   - Filter by entity type

3. **Use JSON Tools**
   ```bash
   # Filter results with jq
   cat /tmp/mcp-outputs/rlm_poc/query_results.json | \
     jq '.chunks | .[0:5]'  # First 5 chunks only
   ```

## Reference

### File Paths

```
# Data files
/tmp/mcp-outputs/rlm_poc/chunks.json
/tmp/mcp-outputs/rlm_poc/knowledge_graph.json
/tmp/mcp-outputs/rlm_poc/graph_chunks/entities_*.json

# Query results (temporary)
/tmp/mcp-outputs/rlm_poc/query_results.json

# Showcase directory
docs/workflows/showcases/graphrag/
```

### Script Locations

```
# Direct query
docs/workflows/showcases/graphrag/scripts/ask_graphrag_direct.sh

# Exploration
docs/workflows/showcases/graphrag/scripts/explore_graphrag.sh

# Validation
docs/workflows/showcases/graphrag/scripts/validate_extraction.sh

# Python scripts
docs/workflows/showcases/graphrag/skills/query_graphrag.py
docs/workflows/showcases/graphrag/skills/explore_graph.py
```

---

**Usage Guide Version:** 1.0
**Last Updated:** February 2026
**Status:** ✅ Ready to Use
