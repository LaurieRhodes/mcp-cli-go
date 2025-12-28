# RAG Pipeline Construction

> **Template:** [rag_pipeline.yaml](../templates/rag_pipeline.yaml)  
> **Workflow:** Documents → Parse → Chunk → Embed → Store → Validate  
> **Best For:** Building production RAG (Retrieval-Augmented Generation) systems

---

## Problem Description

### The RAG Construction Challenge

**Building RAG requires multiple complex steps:**

1. **Document ingestion** - Handle PDFs, markdown, code, various formats
2. **Text extraction** - Clean, parse, preserve structure
3. **Intelligent chunking** - Preserve semantic meaning, avoid mid-sentence splits
4. **Embedding generation** - Choose model, batch process, manage costs
5. **Vector storage** - Select vector DB, configure indexing, manage metadata
6. **Search validation** - Test quality, tune parameters, measure relevance

**Manual RAG pipeline construction:**

```python
# Step 1: Parse documents (custom code for each format)
docs = []
for file in document_files:
    if file.endswith('.pdf'):
        docs.append(parse_pdf(file))  # Custom parser
    elif file.endswith('.md'):
        docs.append(parse_markdown(file))  # Different parser
    # ... handle each format differently

# Step 2: Chunk documents (ad-hoc strategy)
chunks = []
for doc in docs:
    # Fixed 500-token chunks? Semantic splits? Overlaps?
    chunks.extend(simple_chunk(doc, size=500))  # Not ideal

# Step 3: Generate embeddings (slow, one at a time)
embeddings = []
for chunk in chunks:  # This takes HOURS for 10K chunks
    emb = openai.Embedding.create(input=chunk, model="text-embedding-ada-002")
    embeddings.append(emb)
    time.sleep(0.1)  # Rate limiting

# Step 4: Store in vector DB (manual batching)
for i in range(0, len(embeddings), 100):
    batch = embeddings[i:i+100]
    pinecone.upsert(vectors=batch)  # Manual batch management

# Step 5: Test search (manual queries, subjective evaluation)
results = pinecone.query("test query", top_k=5)
# Is this good? How do we know? No systematic validation

# Result: 6-8 hours of work, hard to reproduce, inconsistent quality
```

**Problems:**

- **Time-consuming:** 4-8 hours per RAG pipeline
- **Inconsistent:** Different engineers use different chunking strategies
- **Error-prone:** Easy to miss edge cases (empty docs, very long docs)
- **Not reproducible:** Hard to version control Python scripts
- **Expensive:** Inefficient embedding generation (not batched)
- **Unvalidated:** Search quality untested until production

---

## Template Solution

### What It Does

This template implements **automated end-to-end RAG pipeline construction**:

1. **Parses documents** - Handles multiple formats automatically
2. **Chunks intelligently** - Semantic chunking that preserves context
3. **Generates embeddings** - Parallel batch processing for speed
4. **Stores in vector DB** - Optimized batching and metadata
5. **Validates search** - Systematic quality testing
6. **Reports metrics** - Cost, time, quality scores

### Template Structure

```yaml
name: rag_pipeline
description: End-to-end RAG pipeline construction with validation
version: 1.0.0

config:
  defaults:
    provider: anthropic
    model: claude-3-5-sonnet
    temperature: 0.3

steps:
  # Step 1: Parse documents
  - name: parse_documents
    prompt: |
      Parse and extract text from these documents:
      {{input_data.documents}}

      For each document:
      - Extract all text content
      - Preserve structure (headings, paragraphs, lists)
      - Clean formatting artifacts
      - Maintain document metadata (title, source, date)

      Return structured document data.
    output: parsed_docs

  # Step 2: Intelligent semantic chunking
  - name: chunk_documents
    prompt: |
      Chunk these documents semantically:

      {{parsed_docs}}

      Chunking strategy:
      - Target size: {{input_data.chunk_size}} tokens (default: 512)
      - MUST NOT split mid-sentence
      - MUST NOT split mid-paragraph if possible
      - Preserve semantic coherence
      - Include overlap: {{input_data.overlap}} tokens (default: 50)
      - Maintain source document reference

      For each chunk, include:
      - chunk_id (unique identifier)
      - text (the actual content)
      - source_document (which doc it came from)
      - chunk_index (position in document)
      - token_count (actual size)
      - metadata (title, section, etc.)

      Return array of chunks with metadata.
    output: chunks

  # Step 3: Generate embeddings in parallel batches
  - name: generate_embeddings
    for_each: "{{chunks}}"
    item_name: chunk
    servers: [openai-embeddings]  # Or cohere, local models, etc.
    prompt: |
      Generate embedding for:
      {{chunk.text}}

      Model: {{input_data.embedding_model}}
      Dimensions: {{input_data.embedding_dimensions}}
    parallel:
      batch_size: 100  # Process 100 chunks per batch
      max_concurrent: 5  # 5 batches in parallel
    output: embeddings

  # Step 4: Prepare vectors for storage
  - name: prepare_vectors
    prompt: |
      Prepare vectors for {{input_data.vector_db}} storage:

      Chunks: {{chunks}}
      Embeddings: {{embeddings}}

      For each vector:
      - id: {{chunk.chunk_id}}
      - values: {{embedding.vector}}
      - metadata:
          text: {{chunk.text}}
          source: {{chunk.source_document}}
          chunk_index: {{chunk.chunk_index}}
          token_count: {{chunk.token_count}}
          {{chunk.metadata}}

      Format for {{input_data.vector_db}} bulk upsert.
    output: vectors

  # Step 5: Store in vector database
  - name: store_vectors
    servers: [{{input_data.vector_db}}]  # Pinecone, Weaviate, ChromaDB, etc.
    prompt: |
      Upsert vectors to index:

      Index: {{input_data.index_name}}
      Namespace: {{input_data.namespace}}
      Vectors: {{vectors}}

      Batch size: 100 (optimal for most vector DBs)
    output: storage_result

  # Step 6: Test search quality
  - name: validate_search
    servers: [{{input_data.vector_db}}]
    prompt: |
      Test semantic search quality with these queries:

      {{input_data.test_queries}}

      For each query:
      1. Perform vector search (top 5 results)
      2. Check if results are semantically relevant
      3. Score relevance (1-5 scale)
      4. Document any irrelevant results

      Return search quality report.
    output: search_quality

  # Step 7: Generate pipeline report
  - name: generate_report
    prompt: |
      # RAG Pipeline Construction Report

      **Index:** {{input_data.index_name}}
      **Date:** {{execution.timestamp}}
      **Template:** {{template.name}} v{{template.version}}

      ---

      ## Pipeline Summary

      **Documents Processed:** {{parsed_docs.count}}
      **Chunks Created:** {{chunks.count}}
      **Embeddings Generated:** {{embeddings.count}}
      **Vectors Stored:** {{storage_result.count}}

      **Execution Time:** {{execution.duration}}
      **Total Cost:** {{execution.cost}}

      ---

      ## Document Processing

      {{parsed_docs.summary}}

      **Document Types:**
      {{parsed_docs.types}}

      **Total Tokens:** {{parsed_docs.total_tokens}}

      ---

      ## Chunking Strategy

      **Strategy:** Semantic chunking with overlap
      **Target Chunk Size:** {{input_data.chunk_size}} tokens
      **Overlap:** {{input_data.overlap}} tokens
      **Actual Chunk Sizes:** 
      - Min: {{chunks.min_size}}
      - Max: {{chunks.max_size}}
      - Average: {{chunks.avg_size}}

      **Chunks per Document:**
      {{chunks.distribution}}

      ---

      ## Embeddings

      **Model:** {{input_data.embedding_model}}
      **Dimensions:** {{input_data.embedding_dimensions}}
      **Batch Processing:** {{embeddings.batches}} batches
      **Parallel Workers:** {{embeddings.concurrent}}

      **Cost Breakdown:**
      - Embedding generation: {{embeddings.cost}}
      - Vector storage: {{storage_result.cost}}
      - Total: {{execution.total_cost}}

      ---

      ## Vector Database

      **Platform:** {{input_data.vector_db}}
      **Index:** {{input_data.index_name}}
      **Namespace:** {{input_data.namespace}}

      **Storage:**
      - Vectors stored: {{storage_result.count}}
      - Index size: {{storage_result.index_size}}
      - Storage cost: {{storage_result.monthly_cost}}/month

      ---

      ## Search Quality Validation

      {{search_quality.summary}}

      **Test Queries:** {{search_quality.queries_tested}}
      **Average Relevance Score:** {{search_quality.avg_score}}/5

      **Quality Breakdown:**
      - Excellent (5/5): {{search_quality.excellent_count}}
      - Good (4/5): {{search_quality.good_count}}
      - Acceptable (3/5): {{search_quality.acceptable_count}}
      - Poor (<3/5): {{search_quality.poor_count}}

      {% if search_quality.poor_count > 0 %}
      **⚠️ Warning:** Some queries returned poor results. Consider:
      - Adjusting chunk size
      - Different embedding model
      - Tuning search parameters
      {% endif %}

      ---

      ## Recommendations

      **Chunk Size:** {% if chunks.avg_size > 600 %}Consider smaller chunks for better granularity{% elif chunks.avg_size < 300 %}Consider larger chunks to preserve context{% else %}Current size is optimal{% endif %}

      **Search Quality:** {% if search_quality.avg_score >= 4.0 %}✓ Excellent - ready for production{% elif search_quality.avg_score >= 3.0 %}⚠️ Acceptable - consider tuning{% else %}❌ Poor - needs optimization{% endif %}

      **Cost Optimization:**
      - Current embedding cost: {{embeddings.cost}}
      - Monthly storage: {{storage_result.monthly_cost}}
      {% if embeddings.cost > 1.0 %}
      - Consider: Local embedding models for cost reduction
      {% endif %}

      ---

      ## Next Steps

      1. {% if search_quality.avg_score >= 4.0 %}Deploy to production{% else %}Tune chunking/embedding parameters{% endif %}
      2. Monitor search quality with real user queries
      3. Update index as new documents added
      4. Track cost and performance metrics

      **RAG Pipeline Status:** {% if search_quality.avg_score >= 4.0 %}✓ READY{% elif search_quality.avg_score >= 3.0 %}⚠️ NEEDS TUNING{% else %}❌ REQUIRES OPTIMIZATION{% endif %}
```

---

## Usage Examples

### Example 1: Build RAG for Documentation

**Scenario:** Create semantic search for product documentation (100 markdown files)

**Input:**

```json
{
  "documents": "path/to/docs/*.md",
  "chunk_size": 512,
  "overlap": 50,
  "embedding_model": "text-embedding-ada-002",
  "embedding_dimensions": 1536,
  "vector_db": "pinecone",
  "index_name": "product-docs",
  "namespace": "v1",
  "test_queries": [
    "How do I authenticate API requests?",
    "What are the rate limits?",
    "How to deploy to production?",
    "Troubleshooting connection errors"
  ]
}
```

**Execution:**

```bash
mcp-cli --template rag_pipeline --input-data @config.json --verbose
```

**What Happens:**

```
[14:30:00] Starting rag_pipeline
[14:30:00] Step: parse_documents
[14:30:05] ✓ Parsed 100 markdown files
  - Total text: 250,000 tokens
  - Average doc size: 2,500 tokens

[14:30:05] Step: chunk_documents
[14:30:12] ✓ Created 485 chunks
  - Average size: 515 tokens
  - Size range: 412-597 tokens
  - All chunks preserve sentence boundaries

[14:30:12] Step: generate_embeddings (parallel batching)
[14:30:12] → Batch 1 (100 chunks) processing...
[14:30:12] → Batch 2 (100 chunks) processing...
[14:30:12] → Batch 3 (100 chunks) processing...
[14:30:12] → Batch 4 (100 chunks) processing...
[14:30:12] → Batch 5 (85 chunks) processing...
[14:30:18] ✓ Generated 485 embeddings in 6 seconds
  - Model: text-embedding-ada-002
  - Dimensions: 1536
  - Cost: $0.049 (250K tokens)

[14:30:18] Step: prepare_vectors
[14:30:20] ✓ Prepared 485 vectors with metadata

[14:30:20] Step: store_vectors (Pinecone)
[14:30:25] ✓ Upserted 485 vectors to pinecone
  - Index: product-docs
  - Namespace: v1
  - Storage: ~750KB

[14:30:25] Step: validate_search
[14:30:30] ✓ Search quality validated
  - Query 1: "How do I authenticate API requests?"
    → Top result: "Authentication - API Keys" (Relevance: 5/5)
  - Query 2: "What are the rate limits?"
    → Top result: "Rate Limiting - Quotas and Throttling" (Relevance: 5/5)
  - Query 3: "How to deploy to production?"
    → Top result: "Production Deployment Guide" (Relevance: 5/5)
  - Query 4: "Troubleshooting connection errors"
    → Top result: "Common Connection Issues" (Relevance: 4/5)

  Average relevance: 4.75/5 ✓ Excellent

[14:30:30] Step: generate_report
[14:30:32] ✓ Pipeline report generated

[14:30:32] ✓ Template completed (32 seconds total)
```

**Output Report:**

```markdown
# RAG Pipeline Construction Report

**Index:** product-docs
**Status:** ✓ READY FOR PRODUCTION

## Summary

- Documents: 100
- Chunks: 485
- Embeddings: 485
- Time: 32 seconds
- Cost: $0.051

## Search Quality: ✓ EXCELLENT (4.75/5)

All test queries returned highly relevant results. Pipeline ready for production deployment.

## Cost Breakdown

- Embedding generation: $0.049
- Vector storage: $0.002/month (Pinecone)
- Total setup: $0.051
```

**Time saved:**

- Manual setup: 6 hours
- Automated: 32 seconds
- **Savings: 99.9%**

---

### Example 2: RAG for Code Search

**Scenario:** Build semantic code search for large codebase (10K Python files)

**Input:**

```json
{
  "documents": "src/**/*.py",
  "chunk_size": 1024,
  "overlap": 100,
  "embedding_model": "text-embedding-ada-002",
  "vector_db": "weaviate",
  "index_name": "codebase",
  "test_queries": [
    "authentication middleware implementation",
    "database connection pooling",
    "async task queue handlers"
  ]
}
```

**What Happens:**

```
Documents: 10,000 Python files
Total tokens: 5M tokens
Chunks created: 8,500
Embedding cost: $0.625 (5M tokens × $0.13/1M)
Processing time: 8 minutes (parallel batching)
Search quality: 4.3/5 (Good)

Result: Semantic code search operational
```

**Use cases enabled:**

- "Find examples of OAuth implementation"
- "Locate error handling patterns"
- "Search for API endpoint definitions"

---

## When to Use

### ✅ Appropriate Use Cases

**Building RAG Applications:**

- Documentation Q&A systems
- Knowledge base search
- Code search
- Customer support chatbots with context
- Research paper analysis

**Multiple Document Formats:**

- PDFs, markdown, code, HTML
- Need consistent processing across formats
- Want reproducible chunking strategy

**Production RAG Systems:**

- Need validated search quality
- Require cost tracking
- Want version-controlled pipeline
- Need to update/rebuild indexes

**Team Standardization:**

- Multiple engineers building RAG
- Want consistent chunking approach
- Need reproducible results

### ❌ Inappropriate Use Cases

**One-Time Experiments:**

- Quick prototype, manual is faster
- Exploring different approaches
- Not worth template setup

**Very Small Datasets:**

- <10 documents
- Manual processing feasible
- Overhead not justified

**Highly Custom Requirements:**

- Need specialized chunking logic
- Custom embedding models not supported
- Very specific metadata requirements

---

## Trade-offs

### Advantages

**Time Savings:**

- Manual: 6-8 hours setup
- Automated: 15-30 minutes
- **Savings: 95%+** for subsequent builds

**Consistency:**

- Same chunking strategy every time
- No ad-hoc decisions
- Version-controlled approach
- Team alignment

**Quality Assurance:**

- Systematic search validation
- Quality metrics tracked
- Issues caught before production

**Cost Optimization:**

- Parallel embedding generation (faster)
- Optimal batching (cheaper)
- Cost tracking built-in

### Limitations

**Setup Required:**

- MCP servers for vector DB integration
- Template configuration learning curve
- Initial time investment

**Less Flexible:**

- Fixed workflow steps
- Customization requires template editing
- Not as flexible as custom code

**MCP Dependencies:**

- Requires MCP server for vector DB
- API credentials needed
- Integration setup time

---

## Customization

### Adjust Chunking Strategy

```yaml
# In chunk_documents step:
Chunking strategy:
  - Target size: 256 tokens  # Smaller for fine-grained search
  - Overlap: 25 tokens

OR

  - Target size: 1024 tokens  # Larger to preserve context
  - Overlap: 100 tokens
```

### Use Different Embedding Models

```yaml
# OpenAI (default)
servers: [openai-embeddings]
model: text-embedding-ada-002
dimensions: 1536

# Cohere (multilingual)
servers: [cohere-embeddings]
model: embed-multilingual-v3.0
dimensions: 1024

# Local model (free, private)
servers: [ollama]
model: nomic-embed-text
dimensions: 768
```

### Different Vector Databases

```yaml
# Pinecone (managed)
servers: [pinecone]

# Weaviate (self-hosted)
servers: [weaviate]

# ChromaDB (local)
servers: [chromadb]

# Qdrant (hybrid cloud)
servers: [qdrant]
```

---

## Best Practices

**Before Building RAG:**

**✅ Do:**

- Define test queries upfront (validates quality)
- Start with small dataset (test pipeline)
- Experiment with chunk sizes (256, 512, 1024)
- Check embedding costs (can be significant)
- Plan for index updates (how to add new docs)

**❌ Don't:**

- Skip search quality validation
- Use same chunk size for all content types
- Forget to track costs
- Build without test queries
- Ignore edge cases (empty docs, huge docs)

**After Building:**

**✅ Do:**

- Monitor search quality with real queries
- Track user feedback on results
- Update index as docs change
- Optimize based on usage patterns
- Version control pipeline configuration

**❌ Don't:**

- Deploy without validation
- Ignore poor search results
- Let index get stale
- Skip cost monitoring
- Forget to document pipeline

---

## Related Resources

- **[Template File](../templates/rag_pipeline.yaml)** - Download complete template
- **[Vector Similarity Search](vector-similarity.md)** - Related vector operations
- **[Data Quality Validation](data-quality-validation.md)** - Validate document quality
- **[Why Templates Matter](../../../WHY_TEMPLATES_MATTER.md)** - Context management explained

---

**RAG pipeline automation: From 6 hours manual work to 15 minutes reproducible pipeline.**

Remember: Test search quality is critical. A RAG system that returns irrelevant results is worse than no RAG at all.
