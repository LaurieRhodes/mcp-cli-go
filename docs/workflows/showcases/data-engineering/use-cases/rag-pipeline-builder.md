# RAG Pipeline Builder

> **Workflow:** [rag_pipeline_builder.yaml](../workflows/rag_pipeline_builder.yaml)  
> **Pattern:** Systematic Pipeline Construction  
> **Best For:** Building RAG pipelines with validated chunking and cost estimation

---

## Problem Description

### The RAG Pipeline Challenge

**Manual RAG pipeline building:**

```
Day 1: Parse documents (2 hours trial and error)
Day 2: Chunk text (4 hours finding right size)
Day 3: Generate embeddings ($50 wasted on bad chunks)
Day 4: Realize chunking was wrong
Day 5: Start over ($50 more wasted)

Total: 3 days, $100 wasted, still not optimal
```

**Common problems:**
- **Bad chunking:** Breaks context, ruins retrieval
- **Unknown costs:** Surprise $200 embedding bill
- **No validation:** Deploy then discover issues
- **Trial and error:** Waste time and money

**Real incident:**
```
Team embedded entire 10GB documentation
→ Used tiny 200-token chunks
→ Context destroyed
→ RAG retrieval useless
→ Cost: $800 in embeddings
→ Had to re-do everything
→ Systematic pipeline would have caught this
```

---

## Workflow Solution

### What It Does

**Systematic RAG pipeline with validation:**

1. **Parse → Chunk → Plan → Validate → Report**
2. **Cost estimated BEFORE** spending on embeddings
3. **Chunking validated** for context preservation
4. **Step dependencies** ensure correct order

**Value:**
- Time: 8 hours → 5 minutes (99% savings)
- Cost: Know before spending
- Quality: Validated chunking strategy

### Key Features

```yaml
steps:
  - name: parse_documents
    # Extract and validate text
  
  - name: chunk_text
    needs: [parse_documents]
    # Smart chunking with overlap
  
  - name: plan_embeddings
    needs: [chunk_text]
    # Cost estimation BEFORE executing
  
  - name: validate_pipeline
    needs: [parse_documents, chunk_text, plan_embeddings]
    # Validates before you spend money
```

---

## Usage Example

**Input:** 50 markdown documentation files

```bash
./mcp-cli --workflow rag_pipeline_builder \
  --server filesystem \
  --input-data "$(cat docs/*.md)"
```

**Output:**

```markdown
# RAG Pipeline Construction Report

## Summary

**Documents:** 50 files
**Total Text:** 2.5M characters
**Chunks Generated:** 1,247
**Estimated Embedding Cost:** $0.18

---

## Document Parsing

✓ All 50 documents parsed successfully
✓ Text extracted: 2.5M characters
✓ No encoding issues
✓ Structure preserved (headers, lists)

---

## Chunking Strategy

**Configuration:**
- Target size: 500-1000 tokens
- Maximum: 1500 tokens
- Overlap: 100 tokens (preserves context)
- Method: Semantic (split at paragraphs)

**Results:**
- Chunks: 1,247
- Average size: 756 tokens
- Min: 124 tokens (edge case)
- Max: 1,498 tokens (within limit)
- Context preserved: ✓ Yes

**Quality Validation:**
✓ No orphaned sentences
✓ Semantic boundaries respected
✓ Overlap sufficient for context
✓ All chunks within size limits

---

## Embedding Plan

**Model:** text-embedding-3-small
**Dimensions:** 1,536
**Cost:** $0.02 per 1M tokens

**Calculation:**
- Total tokens: 942,372
- Batches: 13 (100 chunks per batch)
- Estimated cost: $0.18
- Time estimate: 45 seconds

**Batch Processing:**
```bash
# Batch 1 (chunks 0-99)
curl https://api.openai.com/v1/embeddings \
  -H "Authorization: Bearer $OPENAI_API_KEY" \
  -d '{"input": [...], "model": "text-embedding-3-small"}'

# Batch 2 (chunks 100-199)
...
```

---

## Validation Results

**Status:** ✅ APPROVED TO PROCEED

**Validation Checks:**
✓ All documents parsed successfully
✓ Chunking preserves context
✓ Token counts within limits
✓ Cost within budget ($0.18 < $1.00)
✓ No duplicate chunks detected

**Decision:** PROCEED with embedding generation

**Confidence:** HIGH (all validations passed)

---

## Next Steps

1. **Execute Embeddings:**
   ```bash
   python generate_embeddings.py --plan embedding_plan.json
   ```

2. **Store in Vector DB:**
   ```python
   import pinecone
   pinecone.upsert(vectors=embeddings, metadata=chunks)
   ```

3. **Test Retrieval:**
   ```python
   query = "How do I configure authentication?"
   results = pinecone.query(vector=embed(query), top_k=5)
   ```

4. **Build RAG Application:**
   - Query → Embed → Search → Retrieve → Generate

---

## Cost Breakdown

**Embedding Generation:**
- OpenAI API: $0.18
- One-time cost

**Vector Database:**
- Pinecone: ~$0.10/month (1,247 vectors)
- Recurring cost

**Total First Month:** $0.28
**Ongoing:** $0.10/month

**Compare to:**
- Manual pipeline: 8 hours × $150/hour = $1,200
- Automated: 5 minutes × $0.05 = $0.05
- **Savings: $1,199.95 (99.996%)**

---

## Quality Assurance

**Prevented Issues:**
✓ No bad chunking ($200 re-embedding saved)
✓ No surprise costs (knew $0.18 upfront)
✓ No context loss (validated semantic chunking)
✓ No wasted time (systematic approach)

**vs Manual Approach:**
- Manual: 3 days trial-and-error
- Automated: 5 minutes systematic
- **Time savings: 99%+**
```

---

## When to Use

### ✅ Appropriate Use Cases

**RAG Applications:**
- Documentation search
- Knowledge base Q&A
- Code search
- Semantic document retrieval

**Cost-Sensitive:**
- Need to know embedding costs upfront
- Budget-constrained projects
- Validate before spending
- Prevent wasted API calls

**Quality-Critical:**
- Context preservation important
- Retrieval quality matters
- Systematic validation needed
- Production deployments

### ❌ Not Needed For

**Tiny Datasets:**
- < 10 documents
- Trivial embedding cost
- Manual chunking fine
- Overkill

**Experimental:**
- Just trying things
- Cost doesn't matter
- Learning/exploration
- Quick prototypes

---

## Trade-offs

### Advantages

**Cost Known Upfront:**
- $0.18 estimated before spending
- No surprise bills
- Budget validation
- **Prevents $50-200 waste**

**Validated Chunking:**
- Context preserved
- Semantic boundaries
- Overlap configured
- **No retrieval quality issues**

**99% Time Savings:**
- Manual: 8 hours
- Automated: 5 minutes
- Systematic approach
- **Focus on application, not pipeline**

### Limitations

**Requires Text Files:**
- Works with: Markdown, text, HTML, code
- Doesn't work: Binary files, images
- PDF requires preprocessing

**Not Magic:**
- Still need good documents
- Can't fix bad source material
- Chunking strategy matters
- **Validates but doesn't create**

---

## Best Practices

**Before Building:**
- Review source documents
- Set budget limits
- Define chunk size requirements
- Choose embedding model

**After Validation:**
- Review cost estimate
- Check chunking samples
- Verify context preservation
- Execute if approved

**After Deployment:**
- Test retrieval quality
- Monitor embedding costs
- Track query performance
- Iterate on chunk size if needed

---

## Related Resources

- **[Workflow File](../workflows/rag_pipeline_builder.yaml)**
- **[ML Data Quality Validator](ml-data-quality-validator.md)**
- **[Data Transformation Pipeline](data-transformation-pipeline.md)**

---

**Systematic RAG pipelines: Know costs, validate quality, deploy confidently.**

Remember: Bad chunking wastes money and ruins retrieval. Validate before you embed.
