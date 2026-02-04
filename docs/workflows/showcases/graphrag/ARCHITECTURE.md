# GraphRAG Architecture

## System Overview

This GraphRAG implementation uses a three-layer architecture that maintains complete traceability from query results back to original source text.

```
┌─────────────────────────────────────────────────────────────────┐
│ Layer 1: ORIGINAL SOURCE TEXT (100% Preserved)                  │
│ File: chunks.json                                                │
│                                                                  │
│ • All 531 chunks of ISM policy                                  │
│ • Complete original wording (UNCHANGED)                          │
│ • Full metadata (chunk_id, section, page)                       │
│ • Word counts and structure                                     │
│                                                                  │
│ ✅ GROUND TRUTH - Never modified, always available              │
└─────────────────────────────────────────────────────────────────┘
                             ↓
                    Entity Extraction
                             ↓
┌─────────────────────────────────────────────────────────────────┐
│ Layer 2: ENTITY EXTRACTION (Linked to Source)                   │
│ Files: entities_CHUNK-*.json (531 files)                        │
│                                                                  │
│ • chunk_id (LINKS BACK TO LAYER 1) ✅                           │
│ • Extracted entities (id, type, description)                    │
│ • Relationships (from, to, type)                                │
│ • One file per chunk                                            │
│                                                                  │
│ ✅ TRACEABLE - Every entity links to source chunk               │
└─────────────────────────────────────────────────────────────────┘
                             ↓
                    Graph Aggregation
                             ↓
┌─────────────────────────────────────────────────────────────────┐
│ Layer 3: KNOWLEDGE GRAPH (Global View)                          │
│ File: knowledge_graph.json                                      │
│                                                                  │
│ • All 2,337 unique entities                                     │
│ • All 2,313 relationships                                       │
│ • Entity types and descriptions                                 │
│ • No duplicate entities                                         │
│ • Optimized for graph traversal                                 │
│                                                                  │
│ ✅ QUERYABLE - Fast multi-hop graph queries                     │
└─────────────────────────────────────────────────────────────────┘
```

## Data Flow

### Phase 1: Entity Extraction

```
ISM Document (PDF)
↓
Chunked into 531 pieces
↓
Parallel Processing (20 workers)
↓
For each chunk:
  ├─ Send to DeepSeek R1
  ├─ Extract entities (id, type, text)
  ├─ Extract relationships (from, to, type)
  └─ Save: entities_CHUNK-XXX.json
↓
Result: 531 entity files
```

**Key Components:**
- **Workflow:** `extract_graph_chunk_production_deepseek.yaml`
- **Script:** `extract_graph_entities.py`
- **LLM:** DeepSeek R1 with "terse" prompts
- **Workers:** 20 parallel (configurable)
- **Time:** ~30s per chunk, ~4 minutes total

### Phase 2: Graph Building

```
Load all 531 entity files
↓
Aggregate entities:
  ├─ Deduplicate by entity ID
  ├─ Merge descriptions
  ├─ Track source chunks
  └─ Build nodes list (2,337 unique)
↓
Aggregate relationships:
  ├─ Deduplicate by (from, type, to)
  ├─ Track occurrences
  └─ Build edges list (2,313 unique)
↓
Save: knowledge_graph.json
```

**Key Components:**
- **Workflow:** `rebuild_graph.yaml`
- **Script:** `build_graph.py` (or inline Python)
- **Time:** ~30 seconds
- **Output:** Single JSON file with complete graph

### Phase 3: Query & Retrieval

```
User Query: "password length"
↓
Search graph for matching entities
  ├─ Text search on entity IDs and descriptions
  ├─ Returns: 6 matching entities
  └─ Entity IDs: passphrases, passwords, etc.
↓
Multi-hop traversal (2 hops by default)
  ├─ Find all connected entities
  └─ Returns: 6 total related entities
↓
Find source chunks
  ├─ Read entity files for chunk_ids
  ├─ Collect unique chunks
  └─ Returns: 6 relevant chunks
↓
Retrieve original text
  ├─ Load chunks.json
  ├─ Find chunks by chunk_id
  └─ Returns: Full ISM policy text
```

**Key Components:**
- **Script:** `query_graphrag.py`
- **Wrapper:** `ask_graphrag_direct.sh`
- **Time:** <1 second
- **Output:** JSON with entities, chunks, source text

## Entity Types

The system automatically identifies and categorizes entities:

**Top Entity Types (100+ total):**
- **CONCEPT** (1,115): Policy concepts, definitions
- **PROCESS** (390): Procedures, workflows
- **ACTOR** (173): Organizational roles, personnel
- **CONTROL** (119): ISM security controls
- **DOCUMENT** (51): Referenced documents
- **SYSTEM** (34): Technical systems
- And 94+ more types...

## Relationship Types

Relationships capture dependencies and connections:

**Top Relationship Types (400+ total):**
- **REQUIRES** (276): Dependencies
- **DEFINES** (271): Definitions
- **RELATED_TO** (166): Associations
- **USES** (59): Usage patterns
- **REFERENCES** (55): Citations
- And 395+ more types...

## Graph Traversal

### Single-Hop Query

```
Query: "passwords"
↓
Finds: passwords (CONCEPT)
↓
1-hop: Entities directly connected
  ├─ passphrases (via ALTERNATIVE_TO)
  ├─ multi-factor authentication (via SUPERSEDED_BY)
  └─ credential storage (via REQUIRES)
```

### Multi-Hop Query (Default: 2 hops)

```
Query: "passwords"
↓
Finds: passwords (CONCEPT)
↓
1-hop: Direct connections
  ├─ passphrases
  ├─ multi-factor authentication
  └─ credential storage
↓
2-hop: Extended network
  ├─ From passphrases:
  │   ├─ random words
  │   └─ minimum length
  ├─ From multi-factor authentication:
  │   ├─ security tokens
  │   └─ biometrics
  └─ From credential storage:
      ├─ password managers
      └─ secure vaults
```

**Configurable Depth:**
- 1-hop: Very focused (direct connections only)
- 2-hop: Balanced (default, comprehensive)
- 3-hop: Broad (extended network)

## Scalability

### Current Performance

- **Chunks:** 531
- **Entities:** 2,337
- **Relationships:** 2,313
- **Query Time:** <1 second
- **Memory Usage:** ~50MB (graph in memory)

### Scaling Considerations

**For 10x scale (5,310 chunks):**
- Entity extraction: ~40 minutes (20 workers)
- Graph building: ~5 minutes
- Entities: ~23,000 (estimated)
- Relationships: ~23,000 (estimated)
- Query time: Still <2 seconds
- Memory usage: ~500MB

**For 100x scale (53,100 chunks):**
- Entity extraction: ~7 hours (20 workers)
- Graph building: ~30 minutes
- Entities: ~230,000 (estimated)
- Relationships: ~230,000 (estimated)
- Query time: 2-5 seconds
- Consider: Graph database (Neo4j) instead of JSON

## Technology Choices

### Why DeepSeek R1?

**Pros:**
- ✅ Excellent entity extraction quality
- ✅ Fast inference (~30s per chunk)
- ✅ Cost effective ($0.015 per chunk)
- ✅ Good at following structured instructions

**Cons:**
- ⚠️ Can be verbose (solved with "terse" prompts)
- ⚠️ Occasional list misalignment (solved with forgiving scripts)

### Why JSON instead of Graph Database?

**For current scale (531 chunks):**
- ✅ Simple to inspect and validate
- ✅ No database overhead
- ✅ Easy to version control
- ✅ Fast enough (<1s queries)
- ✅ Portable and self-contained

**When to switch to graph DB:**
- ❌ >10,000 chunks
- ❌ Need complex graph algorithms
- ❌ Need concurrent writes
- ❌ Need fine-grained access control

### Why Three Layers?

**Separation of Concerns:**
1. **Layer 1 (Source):** Ground truth, never modified
2. **Layer 2 (Entities):** Extracted knowledge, linked to source
3. **Layer 3 (Graph):** Aggregated view, optimized for queries

**Benefits:**
- ✅ Can validate extractions against source
- ✅ Can re-build graph without re-extracting
- ✅ Can improve extraction without losing source
- ✅ Complete audit trail

## Error Handling

### Extraction Phase

**Issue:** LLM returns invalid JSON
**Solution:** Try-catch with retry logic (up to 3 attempts)

**Issue:** LLM doesn't call tools (verbose responses)
**Solution:** "Terse" prompts with explicit instructions

**Issue:** List index misalignment
**Solution:** Forgiving scripts that handle partial results

### Query Phase

**Issue:** No matching entities
**Solution:** Return empty results with suggestions

**Issue:** Too many results
**Solution:** Limit to top N, sorted by relevance

## Security & Privacy

**ISM Policy Content:**
- Original document is public (Australian Government)
- No sensitive extraction data
- All processing done locally

**GraphRAG System:**
- All data stored locally
- No external API calls after building
- Source text never sent to external services
- Entity extraction controlled by local LLM calls

## Future Enhancements

**Potential Improvements:**
1. **Better Entity Typing:** Use ontology for consistent types
2. **Relationship Confidence:** Add confidence scores to edges
3. **Temporal Tracking:** Track which ISM version each entity came from
4. **Semantic Embeddings:** Add vector embeddings for similarity search
5. **Graph Database:** For >10,000 chunks, migrate to Neo4j
6. **Real-time Updates:** Watch for ISM updates and incrementally update graph
7. **UI/Visualization:** Web interface for graph exploration

---

**Architecture Version:** 1.0
**Last Updated:** February 2026
