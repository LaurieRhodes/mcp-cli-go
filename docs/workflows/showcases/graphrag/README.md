# GraphRAG Showcase: Australian ISM Policy Knowledge Graph

A complete, production-ready GraphRAG implementation for the Australian Government Information Security Manual (ISM) policy document.

## 🎯 Overview

This showcase demonstrates a full GraphRAG (Graph-based Retrieval Augmented Generation) pipeline that:

1. **Extracts entities and relationships** from policy documents using LLMs
2. **Builds a knowledge graph** with 2,337 nodes and 2,313 edges
3. **Enables semantic queries** with multi-hop graph traversal
4. **Maintains full traceability** to original source text
5. **Provides validation tools** for quality assurance

## 📊 Results Summary

**Data Processed:**
- 531 ISM policy chunks
- 2,337 unique entities extracted
- 2,313 relationships mapped
- 100+ entity types identified
- 400+ relationship types found

**Performance:**
- Entity extraction: ~30s per chunk (DeepSeek R1)
- Graph building: ~30s (all 531 chunks)
- Query response time: <1 second
- Total cost: ~$8-10

**Quality:**
- 100% source text preservation
- Full entity traceability
- No hallucinations detected
- Complete audit trail

## 🚀 Quick Start

### Query the Knowledge Graph

```bash
# Direct query (fast, no LLM)
./scripts/ask_graphrag_direct.sh "password length"

# Results: 6 matching entities, 6 related entities, 6 relevant chunks
```

### Explore the Graph

```bash
# Show statistics
./scripts/explore_graphrag.sh stats

# Find most connected entities
./scripts/explore_graphrag.sh top 20

# List entities by type
./scripts/explore_graphrag.sh list CONCEPT
```

### Validate Extraction

```bash
# Validate entity extraction against source text
./scripts/validate_extraction.sh CHUNK-305
```

## 📁 Directory Structure

```
graphrag/
├── workflows/         # GraphRAG workflows (YAML)
├── skills/            # Python extraction/query scripts
├── examples/          # Example queries and outputs
├── validation/        # Validation examples
├── scripts/           # Helper scripts
├── ARCHITECTURE.md    # System architecture details
├── RESULTS.md         # Detailed results and metrics
└── USAGE.md           # Complete usage guide
```

## 🏗️ Architecture

### Three-Layer Design

**Layer 1: Original Source (100% Preserved)**
- File: `chunks.json`
- All 531 chunks with exact ISM wording
- Complete metadata and structure
- Ground truth never modified

**Layer 2: Entity Extraction (Linked to Source)**
- Files: `entities_CHUNK-*.json` (531 files)
- Each entity has `chunk_id` linking to Layer 1
- Extracted entities with types and descriptions
- Relationship mappings

**Layer 3: Knowledge Graph (Queryable)**
- File: `knowledge_graph.json`
- 2,337 unique entities aggregated
- 2,313 relationships mapped
- Optimized for fast graph traversal
- Can always trace back to source

### Traceability Flow

```
Query → Entities → Chunk IDs → Original Text

User asks: "password length"
↓
Graph finds: passphrases, passwords, authentication (6 entities)
↓
Returns: CHUNK-305, CHUNK-298, CHUNK-309 (6 chunks)
↓
Shows: Full ISM policy text (original wording)
```

## 💡 Key Features

### 1. Semantic Search
- Finds concepts even when query terms don't match exactly
- Multi-hop graph traversal discovers related entities
- Example: "password" finds passphrases, credentials, authentication

### 2. Complete Traceability
- Every entity links to source chunk
- Every chunk contains original ISM text
- Full validation possible
- Audit-ready

### 3. High Quality
- DeepSeek R1 for entity extraction (excellent results)
- "Terse" prompts to prevent LLM verbosity
- Forgiving scripts handle edge cases
- Sample validation shows 100% accuracy

### 4. Production Ready
- Fast queries (<1 second)
- No ongoing costs after building
- Simple JSON data format
- Easy to integrate

## 📈 Example Queries

### Broad Query: "cyber security"
```json
{
  "matching_entities": 586,
  "related_entities": 1390,
  "relevant_chunks": 487,
  "response_time": "<1 second"
}
```

### Specific Query: "password length"
```json
{
  "matching_entities": 6,
  "related_entities": 6,
  "relevant_chunks": 6,
  "response_time": "<1 second"
}
```

### Technical Query: "email encryption"
```json
{
  "matching_entities": 93,
  "related_entities": 375,
  "relevant_chunks": 325,
  "response_time": "<1 second"
}
```

## 🔍 Validation

All entity extractions can be validated against source text:

```bash
./scripts/validate_extraction.sh CHUNK-305
```

Shows:
- ✅ Original ISM policy text (exact wording)
- ✅ Extracted entities with types
- ✅ Extracted relationships
- ✅ Validation checklist
- ✅ Extraction statistics

See [validation/CHUNK-305_validation.md](validation/CHUNK-305_validation.md) for a complete example.

## 🛠️ Technology Stack

**LLM:** DeepSeek R1 (via DeepSeek Direct API)
- Excellent entity extraction quality
- Fast inference (~30s per chunk)
- 60% cost savings vs OpenRouter

**Workflow Engine:** MCP-CLI-Go
- Parallel processing (20 workers)
- Skill-based architecture
- Python script integration

**Data Format:** JSON
- Simple, human-readable
- Easy to inspect and validate
- Standard tooling (jq, Python)

## 📚 Documentation

- **[ARCHITECTURE.md](ARCHITECTURE.md)** - Detailed system architecture
- **[RESULTS.md](RESULTS.md)** - Complete results and metrics  
- **[USAGE.md](USAGE.md)** - Comprehensive usage guide
- **[validation/](validation/)** - Validation examples

## 🎓 Lessons Learned

**What Worked:**
- ✅ DeepSeek R1 excellent for entity extraction
- ✅ Parallel processing with 20 workers (fast!)
- ✅ "Terse" prompts prevent LLM verbosity
- ✅ Forgiving scripts handle edge cases
- ✅ Direct Python scripts better than complex workflows
- ✅ Three-layer architecture enables validation

**Challenges Overcome:**
- ❌ OpenRouter rate limits → Switched to DeepSeek Direct API
- ❌ Race conditions → Fixed with server verification
- ❌ DeepSeek verbosity → Created ultra-directive prompts
- ❌ List misalignment → Built forgiving exception handlers
- ❌ Workflow complexity → Simplified to direct scripts

**Key Insights:**
- Multi-hop graph traversal finds related concepts brilliantly
- Entity types provide useful structure
- Traceability critical for enterprise use
- Validation tools build trust
- Simple beats complex for operational tools

## 🚀 Use Cases

**Policy Research & Compliance:**
- Find all requirements for a specific control
- Understand relationships between policies
- Identify compliance gaps

**Security Assessments:**
- Discover all security controls for a domain
- Map control dependencies
- Validate implementations

**Knowledge Navigation:**
- Explore policy structure
- Find related concepts
- Trace requirement sources

**Q&A Systems:**
- Answer questions with source citations
- Provide context from graph
- Enable drilling down into details

## 📝 Citation

If you use this showcase, please reference:

```
GraphRAG Implementation for Australian ISM Policy
Built with: MCP-CLI-Go + DeepSeek R1
Data: Australian Government Information Security Manual
2024-2026
```

## 🤝 Contributing

This is a showcase/reference implementation. Feel free to:
- Use as a template for your own GraphRAG projects
- Adapt workflows for different document types
- Improve entity extraction prompts
- Add new query capabilities

## 📄 License

Workflows and scripts: MIT License (see repository root)
ISM Policy Content: © Australian Government (see source)

---

**Status:** ✅ Production Ready
**Last Updated:** February 2026
**Contact:** See repository contributors
