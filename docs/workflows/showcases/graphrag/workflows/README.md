# GraphRAG Workflows

YAML workflow files for the GraphRAG system.

## Workflows

### main_production_graphrag.yaml
Complete production workflow that orchestrates all three phases:
1. Entity extraction from all 531 chunks (parallel)
2. Knowledge graph building (aggregation)
3. Optional: Query processing

**Usage:**
```bash
./mcp-cli --workflow rlm_poc/workflows/production_graphrag_531 < chunks.json
```

### extract_graph_chunk_production_deepseek.yaml
Single chunk entity extraction using DeepSeek R1.

**Features:**
- Extracts entities (id, type, description)
- Identifies relationships (from, to, type)
- Saves to JSON file per chunk

**Input:** Single chunk content
**Output:** entities_CHUNK-XXX.json

### extract_graph_chunk_terse.yaml
Ultra-directive version preventing LLM verbosity.

**Use when:**
- LLM explaining instead of calling tools
- Hitting max iterations
- Need guaranteed tool calling

### query_graph_chunk_terse.yaml
Terse query workflow for Phase 3 (context synthesis).

**Purpose:**
- Query knowledge graph for chunk's entities
- Extract context from chunk content  
- Save enriched context

**Note:** Currently not used (direct scripts preferred)

### rebuild_graph.yaml
Aggregates all entity files into single knowledge graph.

**Process:**
1. Load all 531 entity files
2. Deduplicate entities
3. Merge relationships
4. Save knowledge_graph.json

**Usage:**
```bash
./mcp-cli --workflow rlm_poc/workflows/rebuild_graph
```

### ask_graphrag.yaml
Query workflow with LLM synthesis (experimental).

**Note:** Variable substitution issues led to direct script approach.
See `../scripts/ask_graphrag_direct.sh` for working version.

### query_graphrag_simple.yaml
Simple graph query without LLM synthesis.

**Alternative:** Use `query_graphrag.py` script directly.

## Workflow Development Notes

**Lessons Learned:**
- ✅ Parallel processing very effective (20 workers)
- ✅ "Terse" prompts essential for DeepSeek
- ✅ Direct scripts more reliable than complex workflows
- ⚠️ Variable substitution in workflows can be tricky
- ⚠️ LLMs may write code instead of calling tools

**Best Practices:**
- Use explicit, directive prompts
- Verify server readiness before tool calls
- Handle exceptions with forgiving logic
- Prefer direct scripts for operational tools
- Keep workflows simple and focused

## See Also

- [../skills/README.md](../skills/README.md) - Python scripts
- [../ARCHITECTURE.md](../ARCHITECTURE.md) - System architecture
- [../USAGE.md](../USAGE.md) - Usage guide
