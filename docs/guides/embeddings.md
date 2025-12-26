# Embeddings Guide

Turn text into vectors (numbers) that AI can use for search, recommendations, and more.

**What are embeddings?** Numbers that represent the meaning of text.

**Think of it like:** GPS coordinates for text meaning.

- "cat" → [0.8, 0.2, 0.1, ...] (animal coordinates)
- "kitten" → [0.79, 0.21, 0.11, ...] (nearby coordinates!)
- "democracy" → [-0.3, 0.9, -0.4, ...] (far away coordinates)

**Why use embeddings?**

- ✅ **Semantic search** - Find similar meaning, not just matching words
- ✅ **Smart recommendations** - "More like this" features
- ✅ **Detect duplicates** - Find similar/duplicate content
- ✅ **RAG systems** - Retrieval Augmented Generation (give AI relevant context)
- ✅ **Clustering** - Group similar documents

**Real example:**

- Keyword search: "python" finds "python" only
- Semantic search: "python" finds "Python", "programming language", "coding", "snake" 🐍

**Cost:** ~$0.0001 per 1,000 words (very cheap!)

**Use when:**

- Building search (better than keyword search)
- Recommendations system
- Detecting similar content
- RAG (giving AI relevant context)

**Don't use when:**

- Simple keyword matching works fine
- Exact text matching needed (use grep, database queries)
- Cost-sensitive and don't need semantic understanding

---

## Table of Contents

- [Quick Start](#quick-start)
- [What are Embeddings?](#what-are-embeddings-explained)
- [Basic Usage](#basic-usage)
- [Chunking Strategies](#chunking-strategies)
- [Output Formats](#output-formats)
- [Use Cases](#use-cases)
- [Best Practices](#best-practices)

---

## Quick Start

**Generate embeddings for a single sentence:**

```bash
echo "The cat sat on the mat" | mcp-cli embeddings
```

**Expected output (truncated):**

```json
{
  "model": "text-embedding-3-small",
  "embeddings": [
    {
      "vector": [0.023, -0.451, 0.789, ..., 0.234],
      "dimensions": 1536
    }
  ]
}
```

**What just happened:**

1. Text sent to OpenAI embedding model
2. Model converted text → 1,536 numbers (vector)
3. Vector represents the semantic meaning
4. JSON output with the vector

**Cost:** ~$0.00001 (1/100,000th of a penny)

**Time:** ~0.5 seconds

---

## What are Embeddings? (Explained)

### The Simple Explanation

**Embeddings = Numbers that represent meaning**

**Example:**

```
"dog" → [0.8, 0.3, 0.1, 0.9, -0.2, ...]
"puppy" → [0.79, 0.31, 0.09, 0.88, -0.19, ...]  # Similar numbers!
"car" → [-0.2, 0.7, 0.9, -0.3, 0.5, ...]       # Different numbers!
```

**Why similar words have similar numbers:**

- "dog" and "puppy" mean related things
- Their vectors are close in "meaning space"
- "car" is unrelated, so its vector is far away

---

### How It Works

**Input text:**

```
"The cat sat on the mat"
```

**↓ Embedding Model ↓**

**Output vector (1,536 numbers):**

```
[0.023, -0.451, 0.789, 0.234, -0.123, 0.567, ...]
```

**Now you can:**

- **Compare similarity:** Are two texts about similar things?
- **Search semantically:** Find texts with similar meaning
- **Cluster documents:** Group similar content together
- **Detect duplicates:** Find nearly identical content

---

### Semantic Similarity Example

**Three sentences:**

1. "The cat sat on the mat"
2. "A feline rested on the rug"  
3. "The weather is sunny today"

**Their embeddings:**

```
Sentence 1: [0.8, 0.3, 0.1, ...]
Sentence 2: [0.79, 0.31, 0.09, ...] ← Very close to #1!
Sentence 3: [-0.2, 0.7, 0.9, ...]  ← Far from #1 and #2
```

**Similarity scores (0-1, higher = more similar):**

```
Sentence 1 vs 2: 0.94 (94% similar - both about cats/rugs)
Sentence 1 vs 3: 0.12 (12% similar - different topics)
```

**This is how semantic search works!**

---

### Why Not Just Use Keywords?

**Keyword search:**

```
Query: "python programming"
Finds: Documents containing "python" AND "programming"
Misses: "coding in Python", "Python development", "py script"
```

**Semantic search (with embeddings):**

```
Query: "python programming" → vector: [0.7, 0.3, ...]
Finds: All documents with similar vectors:
  - "python programming" ✓
  - "coding in Python" ✓ (similar meaning)
  - "Python development" ✓ (similar meaning)
  - "py script" ✓ (similar meaning)
  - "Java programming" ✓ (related but different language)
```

**Benefit:** Finds relevant content even with different words!

---

## Basic Usage

### From Text Directly

```bash
# Single line
mcp-cli embeddings "The quick brown fox"

# Multiple lines (use echo)
echo "Line 1
Line 2
Line 3" | mcp-cli embeddings
```

**Output:** JSON with vector(s)

---

### From File

```bash
# Small file (< 8,000 words)
mcp-cli embeddings --input-file document.txt

# Large file (auto-chunks)
mcp-cli embeddings --input-file large-doc.txt \
  --output-file embeddings.json
```

**What happens with large files:**

1. File read (e.g., 10,000 words)
2. Split into chunks (e.g., 10 chunks of ~1,000 words each)
3. Each chunk embedded separately
4. Output contains 10 vectors (one per chunk)

**Why chunk?** Embedding models have limits (8,000 words max)

**Cost:** 10,000 words = ~$0.001 (1/10th of a penny)

---

### Save to File

```bash
# Save as JSON
mcp-cli embeddings --input-file doc.txt \
  --output-file doc-vectors.json

# Save as CSV
mcp-cli embeddings --input-file doc.txt \
  --output-file doc-vectors.csv \
  --output-format csv
```

**File sizes:**

- 1,000 words → ~25KB JSON file
- 10,000 words (10 chunks) → ~250KB JSON file

---

### Use Specific Model

```bash
# Small model (cheaper, faster, 1536 dimensions)
mcp-cli embeddings --model text-embedding-3-small "text"

# Large model (more accurate, 3072 dimensions)
mcp-cli embeddings --model text-embedding-3-large "text"

# Legacy model
mcp-cli embeddings --model text-embedding-ada-002 "text"
```

**Cost comparison per 1M words:**

- text-embedding-3-small: $0.02 
- text-embedding-3-large: $0.13
- text-embedding-ada-002: $0.10

**Accuracy:**

- text-embedding-3-large: Best (100%)
- text-embedding-3-small: Great (95%)
- text-embedding-ada-002: Good (90%)

**Recommendation:** Use text-embedding-3-small (best value)

---

## Chunking Strategies

For large texts, MCP-CLI automatically chunks content.

### Why Chunk?

**Problem:** Embedding models have token limits (e.g., 8192 tokens)

**Solution:** Split large text into chunks, embed each chunk

### Available Strategies

#### 1. Sentence (Default)

Splits at sentence boundaries to preserve meaning.

```bash
mcp-cli embeddings --chunk-strategy sentence \
  --max-chunk-size 512 \
  --input-file document.txt
```

**Best for:**

- Documents with clear sentence structure
- Natural language text
- Preserving semantic coherence

**Example:**

```
Input:
"First sentence. Second sentence. Third sentence."

Chunks:
1. "First sentence. Second sentence."
2. "Third sentence."
```

#### 2. Paragraph

Splits at paragraph boundaries to preserve structure.

```bash
mcp-cli embeddings --chunk-strategy paragraph \
  --max-chunk-size 1024 \
  --input-file document.txt
```

**Best for:**

- Well-structured documents
- Preserving document hierarchy
- Longer context per chunk

**Example:**

```
Input:
"Paragraph 1.\n\nParagraph 2.\n\nParagraph 3."

Chunks:
1. "Paragraph 1."
2. "Paragraph 2."
3. "Paragraph 3."
```

#### 3. Fixed

Splits into fixed-size chunks with optional overlap.

```bash
mcp-cli embeddings --chunk-strategy fixed \
  --max-chunk-size 512 \
  --overlap 50 \
  --input-file document.txt
```

**Best for:**

- Code or structured data
- Consistent chunk sizes
- When semantic boundaries don't matter

**Example with overlap:**

```
Input (600 tokens):
"Token1 Token2 Token3 ... Token600"

Chunks (512 tokens, 50 overlap):
1. Tokens 1-512
2. Tokens 463-600 (overlaps last 50 from chunk 1)
```

### Show Available Strategies

```bash
mcp-cli embeddings --show-strategies

Available Chunking Strategies:
=============================
  sentence       - Splits text at sentence boundaries while preserving semantic meaning
  paragraph      - Splits text at paragraph boundaries to preserve document structure
  fixed          - Splits text into fixed-size chunks with configurable overlap
```

---

## Output Formats

### JSON (Default)

Full metadata and vectors.

```bash
mcp-cli embeddings --output-format json "Text here"
```

**Output:**

```json
{
  "id": "emb_123",
  "model": "text-embedding-3-small",
  "provider": "openai",
  "chunks": [
    {
      "index": 0,
      "text": "Text here",
      "token_count": 2,
      "start_pos": 0,
      "end_pos": 9
    }
  ],
  "embeddings": [
    {
      "chunk": { ... },
      "vector": [0.023, -0.451, 0.789, ..., 0.234],
      "model": "text-embedding-3-small",
      "dimensions": 1536
    }
  ],
  "metadata": {
    "cli_version": "1.0.0",
    "source": "argument",
    "created_at": "2024-12-26T10:30:00Z"
  }
}
```

### CSV

Tabular format with vectors.

```bash
mcp-cli embeddings --output-format csv --output-file embeddings.csv \
  --input-file document.txt
```

**Output:**

```csv
chunk_index,text,vector_json,start_pos,end_pos,token_count
0,"First sentence.","[0.023, -0.451, ...]",0,15,3
1,"Second sentence.","[0.045, -0.332, ...]",16,32,3
```

### Compact

Minimal JSON with just vectors.

```bash
mcp-cli embeddings --output-format compact "Text here"
```

**Output:**

```json
{
  "model": "text-embedding-3-small",
  "vectors": [
    [0.023, -0.451, 0.789, ..., 0.234]
  ]
}
```

**Use when:**

- You only need vectors
- Minimizing file size
- Integrating with vector databases

### Without Metadata

```bash
mcp-cli embeddings --include-metadata=false \
  --output-format json "Text here"
```

**Output:**

```json
{
  "model": "text-embedding-3-small",
  "vectors": [
    [0.023, -0.451, ...]
  ]
}
```

---

## Advanced Features

### Custom Dimensions

Some models support custom dimensions (smaller = faster, cheaper):

```bash
# Full dimensions (1536 for text-embedding-3-small)
mcp-cli embeddings --model text-embedding-3-small "Text"

# Reduced dimensions (faster, cheaper)
mcp-cli embeddings --model text-embedding-3-small \
  --dimensions 512 "Text"
```

**Trade-off:**

- **More dimensions** → Better accuracy, slower, more storage
- **Fewer dimensions** → Faster, cheaper, slightly less accurate

### Encoding Format

```bash
# Float (default)
mcp-cli embeddings --encoding-format float "Text"

# Base64 (smaller file size)
mcp-cli embeddings --encoding-format base64 "Text"
```

### Chunk Overlap

Preserve context across chunk boundaries:

```bash
mcp-cli embeddings --chunk-strategy fixed \
  --max-chunk-size 512 \
  --overlap 50 \
  --input-file long-document.txt
```

**Why overlap?**

- Prevents losing context at boundaries
- Improves semantic search accuracy
- Typical: 10-20% of chunk size

### Show Available Models

```bash
mcp-cli embeddings --show-models

Available Embedding Models:
==========================
  text-embedding-3-small       (max tokens: 8191)
  text-embedding-3-large       (max tokens: 8191)
  text-embedding-ada-002       (max tokens: 8191)
```

---

## Use Cases

### Use Case 1: Semantic Search

**Goal:** Find documents similar to a query

```bash
# 1. Generate embeddings for all documents
for doc in documents/*.txt; do
    mcp-cli embeddings --input-file "$doc" \
      --output-file "embeddings/$(basename $doc).json"
done

# 2. Embed user query
echo "How to deploy kubernetes?" | \
    mcp-cli embeddings --output-file query-embedding.json

# 3. Compare vectors (use your vector DB or similarity function)
python compare_embeddings.py query-embedding.json embeddings/*.json
```

### Use Case 2: Document Clustering

**Goal:** Group similar documents together

```bash
# Generate embeddings for corpus
mcp-cli embeddings --input-file corpus.txt \
  --chunk-strategy paragraph \
  --output-format compact \
  --output-file corpus-embeddings.json

# Cluster with your ML tool
python cluster.py corpus-embeddings.json
```

### Use Case 3: RAG (Retrieval-Augmented Generation)

**Goal:** Find relevant context for AI queries

```bash
#!/bin/bash
# rag-pipeline.sh

# 1. Chunk and embed knowledge base
mcp-cli embeddings --input-file knowledge-base.txt \
  --chunk-strategy sentence \
  --max-chunk-size 512 \
  --output-file kb-embeddings.json

# 2. Embed user question
USER_QUERY="What is the refund policy?"
echo "$USER_QUERY" | mcp-cli embeddings \
  --output-file query-embedding.json

# 3. Find similar chunks (using vector similarity)
RELEVANT_CONTEXT=$(python find_similar.py \
  query-embedding.json kb-embeddings.json)

# 4. Query with context
mcp-cli query "Based on this context: $RELEVANT_CONTEXT
Answer: $USER_QUERY"
```

### Use Case 4: Content Recommendations

**Goal:** Recommend similar articles

```bash
# Embed all articles
for article in articles/*.md; do
    mcp-cli embeddings --input-file "$article" \
      --output-file "vectors/$(basename $article).json"
done

# For given article, find similar ones
python recommend.py vectors/current-article.json vectors/*.json
```

### Use Case 5: Duplicate Detection

**Goal:** Find duplicate or near-duplicate content

```bash
# Generate embeddings
mcp-cli embeddings --input-file submissions.txt \
  --chunk-strategy paragraph \
  --output-file submissions-embeddings.json

# Find duplicates (cosine similarity > 0.95)
python detect_duplicates.py submissions-embeddings.json 0.95
```

---

## Best Practices

### 1. Choose Right Chunk Size

```bash
# General content (512 tokens)
--max-chunk-size 512

# Code or technical (256 tokens for precision)
--max-chunk-size 256

# Long documents (1024 tokens for context)
--max-chunk-size 1024
```

**Rule of thumb:** 

- 512 tokens ≈ 380 words ≈ 2-3 paragraphs

### 2. Use Appropriate Strategy

```bash
# Natural language → sentence
--chunk-strategy sentence

# Well-structured docs → paragraph
--chunk-strategy paragraph

# Code/data → fixed
--chunk-strategy fixed
```

### 3. Add Overlap for Better Search

```bash
# Good for search
--overlap 50

# Good for classification
--overlap 0
```

### 4. Choose Model Based on Need

```bash
# Best quality (more expensive)
--model text-embedding-3-large

# Good balance (recommended)
--model text-embedding-3-small

# Legacy (cheaper)
--model text-embedding-ada-002
```

### 5. Store Metadata

```bash
# Include metadata for debugging
--include-metadata=true

# Skip metadata for production (smaller files)
--include-metadata=false
```

### 6. Batch Processing

```bash
#!/bin/bash
# Process many files efficiently

for file in data/*.txt; do
    output="embeddings/$(basename $file .txt).json"

    if [ ! -f "$output" ]; then
        mcp-cli embeddings --input-file "$file" \
          --output-file "$output" \
          --output-format compact

        echo "Processed: $file"
        sleep 1  # Rate limiting
    fi
done
```

---

## Integration Examples

### Python Integration

```python
import json
import numpy as np

# Load embeddings
with open('embeddings.json') as f:
    data = json.load(f)

vectors = [emb['vector'] for emb in data['embeddings']]
vectors = np.array(vectors)

# Compute similarity
from sklearn.metrics.pairwise import cosine_similarity

similarity = cosine_similarity(vectors)
print(f"Similarity matrix shape: {similarity.shape}")
```

### Vector Database (Pinecone)

```python
import pinecone
import json

# Load embeddings
with open('embeddings.json') as f:
    data = json.load(f)

# Initialize Pinecone
pinecone.init(api_key="...")
index = pinecone.Index("my-index")

# Upsert vectors
vectors = []
for i, emb in enumerate(data['embeddings']):
    vectors.append({
        'id': f'chunk-{i}',
        'values': emb['vector'],
        'metadata': {
            'text': emb['chunk']['text']
        }
    })

index.upsert(vectors)
```

### Vector Database (ChromaDB)

```python
import chromadb
import json

# Load embeddings
with open('embeddings.json') as f:
    data = json.load(f)

# Create client
client = chromadb.Client()
collection = client.create_collection("my-docs")

# Add vectors
texts = [emb['chunk']['text'] for emb in data['embeddings']]
embeddings = [emb['vector'] for emb in data['embeddings']]
ids = [f'id-{i}' for i in range(len(texts))]

collection.add(
    embeddings=embeddings,
    documents=texts,
    ids=ids
)

# Query
results = collection.query(
    query_embeddings=[query_vector],
    n_results=5
)
```

---

## Troubleshooting

### "No input provided"

```bash
# Problem: No input text
mcp-cli embeddings

# Solution: Provide input
echo "Text" | mcp-cli embeddings
# Or
mcp-cli embeddings "Text"
# Or
mcp-cli embeddings --input-file file.txt
```

### "Failed to generate embeddings"

```bash
# Check API key
echo $OPENAI_API_KEY

# Check provider configuration
cat config/providers/openai.yaml

# Test with verbose
mcp-cli --verbose embeddings "Test"
```

### "Input text is empty"

```bash
# Check file exists and has content
cat input-file.txt

# Verify file is not binary
file input-file.txt
```

### Token Limit Exceeded

```bash
# Reduce chunk size
--max-chunk-size 256

# Or use different model
--model text-embedding-3-large  # 8191 tokens
```

---

## Quick Reference

```bash
# Basic
echo "text" | mcp-cli embeddings

# From file
mcp-cli embeddings --input-file doc.txt

# Custom model
mcp-cli embeddings --model text-embedding-3-large "text"

# Chunking
mcp-cli embeddings --chunk-strategy sentence \
  --max-chunk-size 512 --overlap 50 \
  --input-file doc.txt

# Output formats
mcp-cli embeddings --output-format csv --output-file out.csv
mcp-cli embeddings --output-format compact  # Minimal JSON
mcp-cli embeddings --include-metadata=false # No metadata

# Info
mcp-cli embeddings --show-models
mcp-cli embeddings --show-strategies
```

---

## Next Steps

- **[Automation Guide](automation.md)** - Batch processing embeddings
- **[Query Mode](query-mode.md)** - Use embeddings in RAG
- **Vector Databases** - Store and search embeddings

---

## Resources

- **OpenAI Embeddings Guide**: https://platform.openai.com/docs/guides/embeddings
- **Vector similarity**: Cosine similarity, dot product
- **Vector databases**: Pinecone, ChromaDB, Weaviate, Qdrant

---

**Ready to create embeddings?** Start with a simple document! 🔢
