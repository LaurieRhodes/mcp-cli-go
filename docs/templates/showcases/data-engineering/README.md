# Data Engineering & ML Templates

> **For:** Data Engineers, ML Engineers, AI Developers  
> **Purpose:** Automate ML data pipelines, RAG construction, vector operations, and data quality validation

---

## What This Showcase Contains

This section demonstrates how templates automate critical data engineering workflows for AI/ML applications. All examples solve real challenges: RAG pipeline construction, data quality validation, vector similarity search, feature engineering, and data drift detection.

### Available Use Cases

**ML & AI Workflows:**
1. **[RAG Pipeline Construction](use-cases/rag-pipeline.md)** - Document chunking → embeddings → vector database
2. **[ML Data Quality Validation](use-cases/data-quality-validation.md)** - Consensus validation before training
3. **[Vector Similarity Search](use-cases/vector-similarity.md)** - Semantic search, recommendations, duplicate detection
4. **[Feature Engineering Pipeline](use-cases/feature-engineering.md)** - Raw data → ML-ready features
5. **[Data Drift Detection](use-cases/data-drift.md)** - Monitor production vs training distribution
6. **[Schema Evolution Analysis](use-cases/schema-evolution.md)** - Detect breaking changes

---

## Why Templates Matter for Data Engineering

### 1. RAG Pipeline Automation

**The Challenge:** Building RAG (Retrieval-Augmented Generation) requires multiple steps: document processing, intelligent chunking, embedding generation, vector storage, and search quality validation.

**Template Solution:** End-to-end RAG pipeline automation

```yaml
# Automated RAG construction:
steps:
  # 1. Ingest and parse documents
  - name: parse_documents
    servers: [filesystem]
    prompt: "Extract text from: {{input_data.documents}}"
  
  # 2. Intelligent chunking
  - name: chunk_documents
    prompt: "Chunk semantically: {{parsed_docs}}"
    # Semantic chunks vs fixed size
  
  # 3. Generate embeddings (parallel for speed)
  - name: generate_embeddings
    for_each: "{{chunks}}"
    servers: [openai-embeddings]
    # Batch process for efficiency
  
  # 4. Store in vector DB
  - name: store_vectors
    servers: [pinecone]  # or Weaviate, ChromaDB, etc.
    prompt: "Store: {{embeddings}}"
  
  # 5. Validate search quality
  - name: test_search
    servers: [pinecone]
    prompt: "Test semantic search quality"
```

**Impact:**
- Manual setup: 4-8 hours per RAG pipeline
- Automated: 15 minutes
- **Consistent chunking strategy** (not ad-hoc)
- **Reproducible embeddings** (version controlled)

**Documentation:** [RAG Pipeline Construction](use-cases/rag-pipeline.md)

---

### 2. ML Data Quality Validation with Consensus

**The Challenge:** Bad training data = bad models. Single validation approach misses subtle issues. Data quality problems caught after expensive training runs.

**Template Solution:** Multi-provider consensus validation

```yaml
# Validate data quality with 3 AI models before training:
parallel:
  - provider: anthropic
    prompt: "Validate training data quality: {{dataset}}"
  
  - provider: openai
    prompt: "Validate training data quality: {{dataset}}"
  
  - provider: gemini
    prompt: "Validate training data quality: {{dataset}}"

# Cross-validate findings:
# - All 3 agree data is clean: Proceed with training
# - 2 of 3 find issues: Investigate before training
# - All find different issues: Major data problems, halt pipeline
```

**Real scenario:**
```
Training Dataset: 10,000 customer support tickets for sentiment analysis

Claude validation: Missing labels in 5% of data, detected label noise
GPT-4 validation: Missing labels confirmed, also found class imbalance (95% positive)
Gemini validation: Missing labels confirmed, found duplicate tickets (8%)

Consensus: HIGH confidence data issues
Decision: Fix labels, remove duplicates, balance classes BEFORE training
Result: Avoided training biased model, saved 4 GPU hours + retraining cost
```

**Cost consideration:**
- Validation cost: $0.50 for 10K samples
- Training cost wasted on bad data: $50+ (GPU hours)
- **ROI: 100× savings** by catching issues early

**Documentation:** [ML Data Quality Validation](use-cases/data-quality-validation.md)

---

### 3. Vector Similarity Search at Scale

**The Challenge:** Finding similar items in large datasets (product recommendations, duplicate detection, semantic search) requires efficient vector operations.

**Template Solution:** Automated similarity analysis pipeline

```yaml
# Semantic similarity workflow:
steps:
  # 1. Generate embeddings for entire dataset
  - name: embed_dataset
    for_each: "{{dataset}}"
    servers: [openai-embeddings]
    output: embeddings
  
  # 2. Store in vector DB
  - name: store_vectors
    servers: [weaviate]
  
  # 3. Find similar items
  - name: find_similar
    servers: [weaviate]
    prompt: "Find top 10 similar items for each"
  
  # 4. Detect clusters
  - name: cluster_analysis
    prompt: "Identify semantic clusters: {{similarities}}"
  
  # 5. Find duplicates
  - name: detect_duplicates
    prompt: "Items with >0.95 similarity = duplicates"
```

**Use cases:**
- **E-commerce:** "Find similar products" recommendations
- **Customer support:** Find similar past tickets
- **Content deduplication:** Detect near-duplicate articles
- **Semantic search:** Natural language queries

**Performance:**
- 100K items processed in parallel: 20 minutes
- Similarity search: <100ms per query (vector DB)
- Duplicate detection: ~500 pairs found in 100K dataset

**Documentation:** [Vector Similarity Search](use-cases/vector-similarity.md)

---

### 4. Context-Efficient Large Dataset Processing

**The Challenge:** Analyzing large datasets (millions of rows) exceeds LLM context limits. Loading entire dataset = token overflow.

**Template Solution:** Chunk-based processing with context isolation

**Traditional approach (fails on large data):**
```
LLM Context (200K tokens):
├── Full dataset: 500K rows × 100 tokens = 50M tokens
└── ERROR: Context overflow
```

**Template approach (scalable):**
```
Process in chunks:

Chunk 1 (10K rows):
├── LLM analyzes: Fresh 200K context
└── Output: Summary stats for chunk 1

Chunk 2 (10K rows):
├── LLM analyzes: Fresh 200K context
└── Output: Summary stats for chunk 2

...

Chunk 50 (10K rows):
├── LLM analyzes: Fresh 200K context
└── Output: Summary stats for chunk 50

Final aggregation:
├── LLM receives: 50 chunk summaries (25K tokens)
└── Output: Overall insights from 500K rows
```

**Benefits:**
- Can process unlimited dataset size
- Each chunk gets full context window
- Parallel processing for speed
- Aggregated insights from all data

---

### 5. MCP Integration: Data Sources as AI-Accessible Tools

**The Challenge:** ML pipelines need data from Postgres, BigQuery, S3, vector DBs, feature stores—scattered across systems.

**Template Solution:** MCP servers expose data sources

```yaml
# Template can query multiple data systems:
steps:
  # Query training data
  - name: get_training_data
    servers: [postgres]
    prompt: "SELECT * FROM training_data WHERE created > '2024-01-01'"
  
  # Get feature store data
  - name: get_features
    servers: [feast]
    prompt: "Get features for entities: {{entity_ids}}"
  
  # Query vector DB for embeddings
  - name: get_embeddings
    servers: [pinecone]
    prompt: "Fetch embeddings for: {{item_ids}}"
  
  # Get validation metrics
  - name: get_metrics
    servers: [mlflow]
    prompt: "Get model metrics for run: {{run_id}}"
```

**What this enables:**
- Unified data access across systems
- No custom integration code per data source
- Templates become portable (same template, different data)
- Version control for data queries

---

### 6. Parallel Batch Processing for ML Workloads

**The Challenge:** ML workflows require batch operations (embedding 100K documents, validating 50K samples, processing 1M images).

**Template Solution:** Parallel execution with batch optimization

```yaml
# Generate embeddings for 100K documents in parallel:
steps:
  - name: batch_embed
    for_each: "{{documents}}"
    item_name: doc
    parallel:
      batch_size: 1000  # Process 1000 at a time
      max_concurrent: 10  # 10 parallel batches
    servers: [openai-embeddings]
    prompt: "Generate embedding for: {{doc}}"
    output: embeddings

# Result: 100K documents embedded in ~15 minutes
# vs. Sequential: Would take ~25 hours
```

**Performance gain:**
- Sequential: 100K × 1 second = 27 hours
- Parallel (10 concurrent): 100K / 10 × 1 second = 2.7 hours
- Batched + Parallel: 100 batches × 1 second = 100 seconds (1.7 minutes)
- **Speedup: 1000×**

---

## Quick Start

### 1. Choose Your Data Challenge

**Building RAG application?**
- [RAG Pipeline Construction](use-cases/rag-pipeline.md) - Automated chunking → embeddings → vector DB

**Training ML model?**
- [ML Data Quality Validation](use-cases/data-quality-validation.md) - Validate data before expensive training

**Need semantic search?**
- [Vector Similarity Search](use-cases/vector-similarity.md) - Find similar items, detect duplicates

**Feature engineering?**
- [Feature Engineering Pipeline](use-cases/feature-engineering.md) - Raw data → ML features

**Model monitoring?**
- [Data Drift Detection](use-cases/data-drift.md) - Detect when retraining needed

**Schema changes?**
- [Schema Evolution Analysis](use-cases/schema-evolution.md) - Validate schema migrations

### 2. Set Up MCP Integrations

Data engineering templates integrate with data infrastructure:

```yaml
# Vector database integration
servers:
  pinecone:
    command: "pinecone-mcp-server"
    env:
      API_KEY: "${PINECONE_API_KEY}"
      ENVIRONMENT: "${PINECONE_ENV}"
  
  weaviate:
    command: "weaviate-mcp-server"
    env:
      URL: "${WEAVIATE_URL}"
      API_KEY: "${WEAVIATE_API_KEY}"

# Data warehouse integration
servers:
  bigquery:
    command: "bigquery-mcp-server"
    env:
      PROJECT_ID: "${GCP_PROJECT_ID}"
      CREDENTIALS: "${GCP_CREDENTIALS}"
  
  postgres:
    command: "postgres-mcp-server"
    env:
      CONNECTION_STRING: "${POSTGRES_URL}"

# ML platform integration
servers:
  mlflow:
    command: "mlflow-mcp-server"
    env:
      TRACKING_URI: "${MLFLOW_URI}"
```

### 3. Run Template Against Your Data

```bash
# Build RAG pipeline
mcp-cli --template rag_pipeline --input-data "{
  \"documents_path\": \"./docs/\",
  \"vector_db\": \"pinecone\",
  \"index_name\": \"my-rag-index\"
}"

# Validate training data
mcp-cli --template data_quality_validation --input-data "{
  \"dataset_path\": \"./training_data.csv\",
  \"target_column\": \"sentiment\"
}"

# Find similar items
mcp-cli --template vector_similarity --input-data "{
  \"dataset\": \"products.json\",
  \"similarity_threshold\": 0.85
}"
```

---

## Integration Patterns

### Pattern 1: RAG Pipeline → Production

**Complete RAG workflow:**

```yaml
name: production_rag_pipeline

steps:
  # 1. Ingest documents
  - name: ingest
    servers: [s3]
    prompt: "List documents in: {{bucket}}/docs/"
    output: doc_list
  
  # 2. Parse and extract text
  - name: parse
    for_each: "{{doc_list}}"
    servers: [document-parser]
    output: parsed_docs
  
  # 3. Intelligent chunking
  - name: chunk
    prompt: |
      Chunk these documents semantically:
      {{parsed_docs}}
      
      Strategy: Preserve context, ~500 tokens per chunk
    output: chunks
  
  # 4. Generate embeddings
  - name: embed
    for_each: "{{chunks}}"
    parallel:
      batch_size: 100
      max_concurrent: 5
    servers: [openai-embeddings]
    output: embeddings
  
  # 5. Store in vector DB
  - name: store
    servers: [pinecone]
    prompt: |
      Upsert vectors:
      Index: {{input_data.index_name}}
      Namespace: {{input_data.namespace}}
      Vectors: {{embeddings}}
  
  # 6. Test search quality
  - name: validate
    servers: [pinecone]
    prompt: |
      Test queries:
      - "How do I configure authentication?"
      - "What are the API rate limits?"
      - "Deployment best practices"
      
      Expected: Relevant chunks returned
    output: search_quality
  
  # 7. Generate pipeline report
  - name: report
    prompt: |
      RAG Pipeline Report:
      - Documents processed: {{doc_list.count}}
      - Chunks created: {{chunks.count}}
      - Embeddings generated: {{embeddings.count}}
      - Search quality: {{search_quality.score}}
      - Cost: {{execution.cost}}
      - Time: {{execution.duration}}
```

**Result:** Production RAG index ready to serve queries

---

### Pattern 2: ML Training Pipeline with Validation

**Data quality gate before training:**

```yaml
name: ml_training_pipeline

steps:
  # 1. Load training data
  - name: load_data
    servers: [postgres]
    prompt: "SELECT * FROM training_data"
    output: raw_data
  
  # 2. Consensus data validation
  - name: validate_quality
    template: data_quality_validation
    template_input: "{{raw_data}}"
    output: validation_result
  
  # 3. HALT if data quality issues
  - name: check_quality
    condition: "{{validation_result.quality}} == 'PASS'"
    prompt: "Proceed to training"
  
  # 4. Feature engineering
  - name: engineer_features
    template: feature_engineering
    template_input: "{{raw_data}}"
    output: features
  
  # 5. Train/test split
  - name: split_data
    prompt: "80/20 train/test split on: {{features}}"
    output: splits
  
  # 6. Log to MLflow
  - name: log_experiment
    servers: [mlflow]
    prompt: |
      Create experiment run:
      - Dataset: {{raw_data.id}}
      - Features: {{features.count}}
      - Validation: {{validation_result.quality}}
```

**Benefit:** Catches data issues BEFORE wasting GPU hours on bad training run

---

### Pattern 3: Vector Similarity for Recommendations

**Product recommendations at scale:**

```yaml
name: product_recommendations

steps:
  # 1. Get product catalog
  - name: get_products
    servers: [postgres]
    prompt: "SELECT id, name, description FROM products"
    output: products
  
  # 2. Generate product embeddings
  - name: embed_products
    for_each: "{{products}}"
    parallel:
      batch_size: 1000
      max_concurrent: 10
    servers: [openai-embeddings]
    prompt: "Embed: {{item.name}} {{item.description}}"
    output: product_embeddings
  
  # 3. Store in vector DB
  - name: store_embeddings
    servers: [weaviate]
    prompt: "Store product vectors: {{product_embeddings}}"
  
  # 4. For each product, find similar
  - name: find_similar
    for_each: "{{products}}"
    servers: [weaviate]
    prompt: "Find 10 most similar products to: {{item.id}}"
    output: similarity_matrix
  
  # 5. Generate recommendation rules
  - name: create_rules
    prompt: |
      Create recommendation logic:
      
      For product {{item.id}}:
      Recommend: {{similarity_matrix[item.id].top_10}}
      
      Store in Redis for fast lookup
    servers: [redis]
```

**Performance:**
- 100K products × 10 recommendations = 1M similarities
- Generated in ~30 minutes
- Lookup: <10ms via Redis

---

## Best Practices

### RAG Pipeline Design

**✅ Do:**
- Use semantic chunking (preserve context) over fixed-size chunks
- Test multiple chunk sizes (256, 512, 1024 tokens)
- Validate search quality with test queries
- Version control chunking strategy
- Monitor embedding costs (track token usage)
- Use batch processing for embeddings

**❌ Don't:**
- Split chunks mid-sentence
- Use same chunk size for all content types
- Skip search quality validation
- Forget to handle edge cases (empty docs, very long docs)
- Generate embeddings one-by-one (slow and expensive)

### Data Quality Validation

**✅ Do:**
- Use consensus validation for critical datasets
- Define quality metrics upfront (missing %, outliers, etc.)
- Validate BEFORE expensive training
- Track validation results over time
- Document data quality requirements

**❌ Don't:**
- Skip validation to save time (costs more later)
- Use single validation approach for critical data
- Ignore minority findings from consensus
- Train on data that failed validation
- Forget to version datasets

### Vector Operations

**✅ Do:**
- Batch embeddings generation (100-1000 at a time)
- Use appropriate embedding models for use case
- Set similarity thresholds based on use case
- Monitor vector DB performance
- Implement caching for frequent queries

**❌ Don't:**
- Generate embeddings individually (slow)
- Use oversized embeddings (1536d when 384d sufficient)
- Ignore embedding costs (can be significant at scale)
- Skip vector DB indexing (slow searches)
- Forget to clean up old embeddings

---

## Measuring Success

### RAG Pipeline Metrics

**Before templates:**
- Manual pipeline setup: 4-8 hours
- Chunking strategy: Ad-hoc, inconsistent
- Search quality: Untested until production
- Reproducibility: Low (manual steps)

**After templates:**
- Automated setup: 15 minutes
- Chunking strategy: Consistent, version controlled
- Search quality: Validated before deployment
- Reproducibility: High (YAML-defined)
- **Time saved: 4-8 hours per RAG pipeline**

### Data Quality Impact

**Before consensus validation:**
- Bad data caught: After training (wasted GPU hours)
- Validation approach: Single method
- Issues missed: Label noise, class imbalance, duplicates
- Training failures: 15-20% of runs

**After consensus validation:**
- Bad data caught: Before training
- Validation approach: 3-provider consensus
- Issues caught: 95% of problems
- Training failures: <5% of runs
- **GPU hours saved: 40-60 hours/month**

### Vector Similarity Performance

**Processing 100K products:**
- Embedding generation: 20 minutes (parallel)
- Vector storage: 2 minutes
- Similarity computation: 5 minutes
- **Total: 27 minutes for 100K items**

**Search performance:**
- Query latency: <100ms
- Recommendations: <50ms
- Duplicate detection: 500 pairs found from 100K items

---

## Cost Analysis

### RAG Pipeline Costs

**Per 1,000 documents:**
- Document parsing: Minimal (local)
- Chunking (AI): $0.50
- Embedding generation (OpenAI): $0.13 (1M tokens)
- Vector DB storage (Pinecone): $0.096/month
- **Total setup: $0.63 + $0.096/month**

**ROI:**
- Manual setup time: 6 hours @ $100/hr = $600
- Automated: $0.63
- **Savings: 99.9%** (plus ongoing consistency)

### Data Quality Validation Costs

**Per 10K training samples:**
- Single validation: $0.15
- Consensus validation (3 providers): $0.45
- **Cost to catch issues early: $0.45**

**vs. Training on bad data:**
- GPU hours wasted: 4 hours @ $3/hr = $12
- Retraining cost: $12
- **ROI: 27× savings** by validating first

---

## Example: Complete ML Pipeline

```yaml
name: complete_ml_pipeline

steps:
  # 1. Load raw data
  - name: load_raw_data
    servers: [bigquery]
    prompt: "SELECT * FROM ml_training.raw_features"
    output: raw_data
  
  # 2. Consensus data quality validation
  - name: validate_data
    template: data_quality_validation
    template_input: "{{raw_data}}"
    output: validation
  
  # 3. Feature engineering
  - name: engineer_features
    condition: "{{validation.quality}} == 'PASS'"
    template: feature_engineering
    template_input: "{{raw_data}}"
    output: features
  
  # 4. Detect data drift
  - name: check_drift
    template: data_drift_detection
    template_input: |
      Training: {{historical_data}}
      Current: {{features}}
    output: drift_analysis
  
  # 5. Log to MLflow
  - name: log_run
    servers: [mlflow]
    prompt: |
      Log experiment:
      - Validation: {{validation}}
      - Features: {{features.count}}
      - Drift: {{drift_analysis.score}}
  
  # 6. Trigger training (if quality OK + no drift)
  - name: trigger_training
    condition: "{{validation.quality}} == 'PASS' AND {{drift_analysis.drift}} == 'LOW'"
    servers: [airflow]
    prompt: "Trigger training DAG with features: {{features}}"
```

**This pipeline demonstrates:**
- Data quality gates (consensus validation)
- Feature engineering automation
- Drift detection
- MLflow integration
- Conditional training trigger

---

## Template Library

All templates available in [templates/](templates/):

**RAG & Vector Operations:**
- `rag_pipeline.yaml` - End-to-end RAG construction
- `vector_similarity.yaml` - Similarity search and clustering
- `semantic_deduplication.yaml` - Find near-duplicates

**ML Data Quality:**
- `data_quality_validation.yaml` - Consensus validation
- `feature_engineering.yaml` - Raw data → ML features
- `data_drift_detection.yaml` - Monitor distribution shifts

**Schema & Compliance:**
- `schema_evolution.yaml` - Detect breaking changes
- `pii_detection.yaml` - Find sensitive data

---

## Next Steps

1. **Review use cases** - Read detailed documentation for each workflow
2. **Set up MCP servers** - Configure vector DB, data warehouse integrations
3. **Test with sample data** - Run templates on small datasets first
4. **Measure baseline** - Track current pipeline execution times
5. **Deploy incrementally** - Start with one workflow, expand gradually

---

## Additional Resources

- **[MCP Server Integration](../../../mcp-server/README.md)** - Expose templates as tools
- **[Why Templates Matter](../../WHY_TEMPLATES_MATTER.md)** - Context management for large datasets
- **[DevOps Showcase](../devops/)** - Failover patterns applicable to data pipelines
- **[Template Authoring Guide](../../authoring-guide.md)** - Create custom data engineering templates

---

**Data engineering with AI: Automate RAG pipelines, validate data quality, scale vector operations.**

Templates transform data engineering from manual scripts to reproducible, version-controlled ML workflows.
