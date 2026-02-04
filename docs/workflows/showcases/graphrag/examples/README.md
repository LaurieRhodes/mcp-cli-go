# GraphRAG Examples

Example queries and outputs demonstrating the GraphRAG system.

## Query Examples

### queries/cyber_security.json
Broad query finding cyber security related concepts.

**Results:**
- 586 matching entities
- 1,390 related entities (via graph traversal)
- 487 relevant chunks (92% of document)

**Demonstrates:**
- Comprehensive coverage for broad topics
- Graph expansion discovering related concepts
- High recall

### queries/password_length.json
Specific query about password requirements.

**Results:**
- 6 matching entities
- 6 related entities
- 6 relevant chunks

**Demonstrates:**
- Precision for specific queries
- Focused results when appropriate
- Direct entity matches

### queries/email_encryption.json
Technical query about email security.

**Results:**
- 93 matching entities
- 375 related entities
- 325 relevant chunks

**Demonstrates:**
- Technical concept handling
- Balanced breadth and depth
- Multi-hop relationship discovery

## Output Examples

### outputs/sample_entities_CHUNK-305.json
Complete entity extraction for CHUNK-305 (password/passphrase requirements).

**Contains:**
- 9 extracted entities
- 2 relationships
- Entity types: CONCEPT, ACTOR, PROCESS
- Linked to source: chunk_id "CHUNK-305"

### outputs/knowledge_graph_sample.json
Sample of knowledge graph structure (first 50 nodes and edges).

**Shows:**
- Node structure (id, type, text, source_chunks)
- Edge structure (from, to, type)
- Entity types distribution
- Relationship patterns

### outputs/query_results_password_length.json
Complete query results for "password length" query.

**Contains:**
- Matching entities (6)
- Related entities via graph (6)
- Relevant chunks (6)
- Subgraph of connected entities
- Statistics and metadata

## Running Examples

### Test Query Examples

```bash
cd ../

# Run each example query
./scripts/ask_graphrag_direct.sh "cyber security"
./scripts/ask_graphrag_direct.sh "password length"
./scripts/ask_graphrag_direct.sh "email encryption"

# Compare results to expected outputs in queries/*.json
```

### Inspect Example Outputs

```bash
# View sample entity extraction
cat outputs/sample_entities_CHUNK-305.json | jq '.'

# View knowledge graph sample
cat outputs/knowledge_graph_sample.json | jq '.'

# View query results
cat outputs/query_results_password_length.json | jq '.'
```

## Creating Your Own Examples

### Add New Query Example

```bash
# Create query file
cat > queries/my_query.json << 'EOF'
{
  "query": "my search terms",
  "description": "What I'm looking for",
  "expected_results": {
    "matching_entities": "approximate count",
    "related_entities": "approximate count",
    "relevant_chunks": "approximate count"
  }
}
EOF

# Test the query
../scripts/ask_graphrag_direct.sh "my search terms"
```

### Capture Output Example

```bash
# Run query and save results
../scripts/ask_graphrag_direct.sh "my search terms" > my_query_output.txt

# Save JSON results
cp /tmp/mcp-outputs/rlm_poc/query_results.json \
   outputs/query_results_my_query.json
```

## See Also

- [../USAGE.md](../USAGE.md) - How to run queries
- [../RESULTS.md](../RESULTS.md) - Complete results and metrics
- [../validation/](../validation/) - Validation examples
