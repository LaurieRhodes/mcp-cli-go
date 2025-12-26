# Document Analysis Pattern

Process and analyze documents systematically with AI.

---

## Overview

The **Document Analysis Pattern** structures document processing workflows for comprehensive analysis.

**What it does:**
- Extracts key information from documents (facts, entities, dates)
- Classifies document type and content
- Analyzes from multiple perspectives (legal, financial, technical)
- Generates structured summaries and reports

**Use when:**
- Processing long documents (> 5 pages, > 5,000 words)
- Need structured information extraction (not just summary)
- Multiple analysis perspectives required (security + compliance + quality)
- Processing batches of similar documents
- Document classification and routing needed

**Real-world examples:**
- Legal: Review contracts for obligations, deadlines, risks
- Financial: Extract metrics from earnings reports
- Technical: Analyze documentation completeness and accuracy
- Compliance: Check documents against regulatory requirements

**Why use this pattern:**
- Systematic: Same analysis every time
- Comprehensive: Multiple perspectives in one pass
- Structured: Get data in consistent format
- Scalable: Process many documents with same workflow

---

## Pattern Structure

```
Document → Extract Info → Classify → Analyze → Generate Report
```

**What happens:**
1. **Extract:** Pull out key information (who, what, when, where)
2. **Classify:** Determine document type and category
3. **Analyze:** Deep analysis based on classification
4. **Report:** Structure findings in useful format

### Basic Document Analysis

**What it does:** Complete document processing - extract, classify, analyze, report.

**Use when:** Processing any document that needs structured analysis.

```yaml
name: document_analysis
description: Comprehensive document processing
version: 1.0.0

config:
  defaults:
    provider: anthropic
    model: claude-sonnet-4  # Best for long documents

steps:
  # Step 1: Extract key information
  - name: extract_key_info
    prompt: |
      Extract key information from this document:
      
      {{input_data.document}}
      
      Extract:
      - Document type (contract, report, email, article, technical, other)
      - Main topics (list of key themes)
      - Key entities:
        * People (names and roles)
        * Organizations (companies mentioned)
        * Dates (important dates and deadlines)
        * Locations (places mentioned)
      - Important facts (critical information)
      - Action items (things that need to be done)
      
      Return as structured JSON.
    output: extracted_info
  
  # Step 2: Classify document
  - name: classify_content
    prompt: |
      Classify this document:
      
      Extracted info: {{extracted_info}}
      
      Determine:
      - Primary category: legal, technical, financial, marketing, HR, other
      - Sub-categories (specific type within category)
      - Topics covered (detailed list)
      - Audience: technical, general, executive
      - Urgency: low, medium, high, critical
      - Complexity: simple, moderate, complex
      
      Return as JSON.
    output: classification
  
  # Step 3: Analyze content based on classification
  - name: analyze_content
    prompt: |
      Analyze document based on its classification:
      
      Classification: {{classification}}
      Content: {{extracted_info}}
      
      Provide:
      - Summary (3-5 sentences of main points)
      - Key insights (what's significant or surprising)
      - Important findings (critical information)
      - Recommendations (suggested actions)
      - Risks or concerns (potential issues)
      
      Return as structured JSON.
    output: analysis
  
  # Step 4: Generate structured report
  - name: create_report
    prompt: |
      Create comprehensive analysis report:
      
      # Document Analysis Report
      
      ## Document Information
      - Type: {{classification.primary_category}}
      - Urgency: {{classification.urgency}}
      - Complexity: {{classification.complexity}}
      
      ## Key Information
      {{extracted_info}}
      
      ## Analysis
      {{analysis}}
      
      ## Summary
      [3-5 sentence executive summary]
      
      ## Recommendations
      [Numbered list of suggested actions]
      
      ## Next Steps
      [What should happen next]
      
      Format as markdown.
    output: final_report
```

**Usage:**
```bash
# Example 1: Analyze contract
mcp-cli --template document_analysis --input-data '{
  "document": "CONTRACT AGREEMENT\n\nThis agreement dated December 1, 2024..."
}'

# Example 2: Analyze from file
mcp-cli --template document_analysis --input-data "{
  \"document\": \"$(cat contract.txt)\"
}"

# Example 3: Analyze technical documentation
mcp-cli --template document_analysis --input-data '{
  "document": "API DOCUMENTATION\n\nEndpoint: POST /api/v1/users..."
}'
```

**What happens:**
1. Extract: Identifies contract type, parties (Acme Corp, Widget Inc), deadline (Jan 15)
2. Classify: Legal document, sub-type: service agreement, high urgency, moderate complexity
3. Analyze: Summarizes terms, identifies key obligation (30-day notice), notes risk (auto-renewal)
4. Report: Generates markdown report with recommendations

**Example output:**
```markdown
# Document Analysis Report

## Document Information
- Type: Legal - Service Agreement
- Urgency: High (deadline in 20 days)
- Complexity: Moderate

## Key Parties
- Provider: Acme Corp
- Client: Widget Inc

## Critical Dates
- Start: Dec 1, 2024
- Renewal: Jan 15, 2025
- Notice: 30 days before renewal

## Key Terms
- Service: Cloud hosting
- Payment: $5,000/month
- Term: 12 months with auto-renewal

## Analysis
This is a standard service agreement with auto-renewal clause. Key risk: 
Auto-renewal requires 30-day notice to cancel. Deadline approaching.

## Recommendations
1. Review if we want to renew (decide by Dec 16)
2. If canceling, send notice immediately
3. Archive this contract in contracts database
```

**Performance:**
- Short docs (< 5 pages): ~30 seconds
- Long docs (5-20 pages): ~2 minutes
- Very long (20+ pages): Use chunking pattern

**Cost:**
- ~$0.03-0.10 per document depending on length

---

## Pattern: Multi-Document Analysis

Analyze and compare multiple documents.

```yaml
name: multi_document_analysis
steps:
  # Process each document
  - name: analyze_documents
    for_each: "{{documents}}"
    item_name: doc
    prompt: |
      Analyze document {{index}}:
      
      {{doc}}
      
      Extract:
      - Main points
      - Key data
      - Conclusions
    output: document_analyses
  
  # Compare documents
  - name: compare_documents
    prompt: |
      Compare these document analyses:
      
      {{document_analyses}}
      
      Identify:
      - Common themes
      - Contradictions
      - Unique insights per document
      - Overall pattern
    output: comparison
  
  # Synthesize findings
  - name: synthesize
    prompt: |
      Create comprehensive synthesis:
      
      {{comparison}}
      
      Provide:
      - Unified narrative
      - Key takeaways
      - Recommendations based on all documents
```

---

## Pattern: Question-Answering from Documents

Extract specific information to answer questions.

```yaml
name: document_qa
description: Answer questions from document content
version: 1.0.0

steps:
  # Step 1: Index document content
  - name: index_document
    prompt: |
      Break this document into logical sections:
      
      {{document}}
      
      Return sections with:
      - Section title
      - Content
      - Key topics
    output: indexed_sections
  
  # Step 2: Find relevant sections
  - name: find_relevant_sections
    prompt: |
      Given this question: {{question}}
      
      Which sections are most relevant?
      {{indexed_sections}}
      
      Return top 3 relevant sections.
    output: relevant_sections
  
  # Step 3: Answer question
  - name: answer_question
    prompt: |
      Answer this question using only the provided sections:
      
      Question: {{question}}
      
      Relevant sections:
      {{relevant_sections}}
      
      Provide:
      - Direct answer
      - Supporting evidence from document
      - Section references
    output: answer
  
  # Step 4: Verify answer
  - name: verify_answer
    prompt: |
      Verify this answer is supported by document:
      
      Question: {{question}}
      Answer: {{answer}}
      Document: {{document}}
      
      Is answer:
      - Accurate? (yes/no)
      - Complete? (yes/no)
      - Well-supported? (yes/no)
      
      If issues found, provide corrections.
    output: verified_answer
```

---

## Pattern: Document Chunking and Embedding

**What it does:** Breaks large documents into manageable pieces, analyzes each, then synthesizes.

**Use when:**
- Documents are very long (> 20 pages, > 20,000 words)
- Document exceeds AI context window (> 100,000 tokens)
- Want to process document in parallel for speed

**Why chunk documents:**
- AI models have context limits (can't process 200-page PDF at once)
- Smaller chunks = faster processing
- Can process chunks in parallel
- Better accuracy (AI focuses on smaller section)

```yaml
name: large_document_analysis
description: Handle documents too large for single pass
version: 1.0.0

steps:
  # Step 1: Split document intelligently
  - name: chunk_document
    prompt: |
      Split this document into semantic chunks:
      
      {{input_data.document}}
      
      Rules:
      - Keep related content together (don't split mid-topic)
      - Each chunk should be coherent standalone
      - Target ~1000-1500 words per chunk
      - Include context headers (chapter name, section title)
      - Number each chunk for reference
      
      Return array of chunks with metadata.
    output: chunks
  
  # Step 2: Analyze each chunk
  - name: analyze_chunks
    for_each: "{{chunks}}"
    item_name: chunk
    prompt: |
      Analyze chunk {{index}} of {{total}}:
      
      Context: {{chunk.header}}
      Content: {{chunk.text}}
      
      Extract:
      - Main points (key takeaways from this section)
      - Key entities (people, organizations, dates mentioned)
      - Important data (numbers, metrics, statistics)
      - Action items (tasks mentioned in this section)
      
      Return as JSON.
    output: chunk_analyses
  
  # Step 3: Synthesize all chunk analyses
  - name: synthesize_analyses
    prompt: |
      Synthesize analyses from all document chunks:
      
      {{chunk_analyses}}
      
      Create complete document analysis:
      - Overall document summary (what is this document about)
      - Key findings across all chunks
      - Important patterns or themes
      - Complete entity list (all people, orgs mentioned)
      - Complete action item list
      - Timeline (all important dates in order)
      
      Return as structured JSON.
    output: document_summary
```

**Usage:**
```bash
# Large PDF document (100 pages)
mcp-cli --template large_document_analysis --input-data "{
  \"document\": \"$(cat large-report.txt)\"
}"

# Annual report
mcp-cli --template large_document_analysis --input-data '{
  "document": "ANNUAL REPORT 2024\n\n[50,000 words of content]..."
}'
```

**What happens:**
1. Chunk: Splits 100-page document → 67 chunks (~1,500 words each)
2. Analyze: Processes each chunk → 67 chunk analyses (can run in parallel!)
3. Synthesize: Combines all analyses → complete document summary

**Performance comparison:**
- **Single pass (impossible):** Document too large for context window
- **Sequential chunking:** 67 chunks × 10 seconds = 11 minutes
- **Parallel chunking (20 concurrent):** ~35 seconds

**Cost:**
- 67 chunks + 1 synthesis = 68 AI calls
- ~$2.00 per 100-page document

**Chunking strategy example:**

Input: 100-page technical document
```
Document → Split by chapters →
  Chapter 1 (5 pages) → 3 chunks
  Chapter 2 (15 pages) → 10 chunks
  Chapter 3 (8 pages) → 5 chunks
  ...
Total: 67 chunks
```

Each chunk:
```json
{
  "chunk_id": 5,
  "header": "Chapter 2: System Architecture",
  "text": "The system consists of...",
  "word_count": 1,450
}
```

**Why this works:**
- Keeps related content together (doesn't split mid-paragraph)
- Each chunk has context (chapter header)
- Synthesis combines insights from all chunks
- No information lost

---

## Pattern: Parallel Document Processing

Process different aspects concurrently.

```yaml
name: parallel_document_processing
steps:
  - name: parallel_analysis
    parallel:
      # Content analysis
      - name: content_analysis
        prompt: |
          Analyze content quality:
          {{document}}
          
          Check:
          - Clarity
          - Completeness
          - Accuracy
        output: content_quality
      
      # Structure analysis
      - name: structure_analysis
        prompt: |
          Analyze document structure:
          {{document}}
          
          Check:
          - Organization
          - Flow
          - Section coherence
        output: structure_quality
      
      # Technical analysis
      - name: technical_analysis
        prompt: |
          Analyze technical aspects:
          {{document}}
          
          Check:
          - Terminology correctness
          - Technical accuracy
          - Appropriate level
        output: technical_quality
    
    max_concurrent: 3
    aggregate: merge
  
  # Combine analyses
  - name: comprehensive_review
    prompt: |
      Create comprehensive document review:
      
      {{parallel_analysis}}
      
      Provide:
      - Overall assessment
      - Strengths
      - Weaknesses
      - Recommendations
```

---

## Real-World Examples

### Example 1: Legal Document Analysis

```yaml
name: legal_document_analysis
steps:
  # Extract legal elements
  - name: extract_legal_elements
    prompt: |
      Extract from legal document:
      {{document}}
      
      - Parties involved
      - Key terms and conditions
      - Obligations
      - Deadlines
      - Remedies
      - Jurisdiction
    output: legal_elements
  
  # Identify risks
  - name: identify_risks
    prompt: |
      Identify legal risks in:
      {{legal_elements}}
      
      Rate each risk:
      - High: Significant exposure
      - Medium: Moderate concern
      - Low: Minor issue
    output: risk_analysis
  
  # Generate summary
  - name: legal_summary
    prompt: |
      Create legal summary:
      
      Elements: {{legal_elements}}
      Risks: {{risk_analysis}}
      
      Format for legal team review.
```

### Example 2: Technical Documentation Review

```yaml
name: tech_doc_review
steps:
  # Completeness check
  - name: check_completeness
    prompt: |
      Check technical documentation completeness:
      {{document}}
      
      Should include:
      - Overview
      - Prerequisites
      - Step-by-step instructions
      - Examples
      - Troubleshooting
      - FAQ
      
      List missing sections.
    output: completeness
  
  # Clarity check
  - name: check_clarity
    prompt: |
      Check documentation clarity:
      {{document}}
      
      Find:
      - Unclear explanations
      - Missing definitions
      - Confusing steps
      - Ambiguous instructions
    output: clarity_issues
  
  # Accuracy check
  - name: check_accuracy
    servers: [code-analyzer]
    prompt: |
      Verify code examples in:
      {{document}}
      
      Check for:
      - Syntax errors
      - Deprecated APIs
      - Security issues
    output: accuracy_check
  
  # Generate review report
  - name: review_report
    prompt: |
      Create documentation review:
      
      Completeness: {{completeness}}
      Clarity: {{clarity_issues}}
      Accuracy: {{accuracy_check}}
      
      Priority fixes and recommendations.
```

### Example 3: Financial Report Analysis

```yaml
name: financial_report_analysis
steps:
  # Extract financial data
  - name: extract_financials
    prompt: |
      Extract from financial report:
      {{document}}
      
      - Revenue figures
      - Expenses
      - Profit/loss
      - Cash flow
      - Key ratios
      - YoY comparisons
    output: financial_data
  
  # Calculate metrics
  - name: calculate_metrics
    prompt: |
      Calculate financial metrics:
      {{financial_data}}
      
      - Profit margin
      - Growth rate
      - Liquidity ratios
      - Efficiency ratios
    output: metrics
  
  # Trend analysis
  - name: analyze_trends
    prompt: |
      Analyze financial trends:
      
      Data: {{financial_data}}
      Metrics: {{metrics}}
      
      Identify:
      - Positive trends
      - Concerning trends
      - Anomalies
    output: trends
  
  # Executive summary
  - name: executive_summary
    prompt: |
      Create executive summary:
      
      Financials: {{financial_data}}
      Metrics: {{metrics}}
      Trends: {{trends}}
      
      Focus on key insights for decision-makers.
```

---

## Best Practices

### 1. Structure Extraction

```yaml
# Good: Structured extraction
- prompt: |
    Extract as JSON:
    {
      "type": "...",
      "entities": [...],
      "key_points": [...]
    }

# Bad: Unstructured
- prompt: "Tell me about the document"
```

### 2. Chunking for Large Documents

```yaml
# Good: Semantic chunking
- prompt: "Split by logical sections"
- for_each: "{{chunks}}"

# Bad: Hard character limits
# (Might split mid-sentence)
```

### 3. Multi-Pass Analysis

```yaml
# Good: Multiple perspectives
- name: extract
- name: classify
- name: analyze
- name: verify

# Bad: Single pass
- name: do_everything
```

### 4. Preserve Context

```yaml
# Good: Context preservation
- prompt: |
    Chunk {{index}} of {{total}}
    
    Previous context: {{previous_summary}}
    Current chunk: {{chunk}}
    
    Analyze with context.

# Bad: Isolated chunks
- prompt: "Analyze: {{chunk}}"
```

---

## Integration with MCP Servers

### Filesystem for Document Loading

```yaml
- name: load_document
  servers: [filesystem]
  prompt: "Read file: {{filepath}}"
  output: document
```

### OCR for Scanned Documents

```yaml
- name: extract_text
  servers: [ocr-service]
  prompt: "Extract text from image: {{image_path}}"
  output: extracted_text
```

### Database for Metadata

```yaml
- name: store_analysis
  servers: [database]
  prompt: |
    Store analysis:
    Document ID: {{doc_id}}
    Analysis: {{analysis}}
```

---

## Performance Optimization

### Parallel Processing

```yaml
# Fast: Parallel analysis
parallel:
  - name: extract
  - name: classify
  - name: validate
max_concurrent: 3
```

### Chunking Strategy

```yaml
# For 100-page documents
- prompt: "Split into ~10 page chunks"
- for_each: "{{chunks}}"
  parallel: true  # Process chunks in parallel
  max_concurrent: 5
```

### Incremental Processing

```yaml
# Process only what's needed
- name: quick_scan
  prompt: "High-level overview: {{document}}"

- name: detailed_analysis
  condition: "{{user_needs_details}} == true"
  prompt: "Deep analysis: {{document}}"
```

---

## Error Handling

```yaml
steps:
  - name: load_document
    servers: [filesystem]
    prompt: "Load: {{filepath}}"
    error_handling:
      on_failure: stop
      default_output: "LOAD_FAILED"
  
  - name: analyze
    condition: "{{load_document}} not contains 'FAILED'"
    prompt: "Analyze: {{load_document}}"
```

---

## Complete Example

```yaml
name: comprehensive_document_analysis
version: 1.0.0

config:
  defaults:
    provider: anthropic
    model: claude-sonnet-4

steps:
  # 1. Load and validate
  - name: load
    servers: [filesystem]
    prompt: "Load document: {{filepath}}"
    output: document
    error_handling:
      on_failure: stop
  
  # 2. Parallel analysis
  - name: analyze_all
    parallel:
      - name: extract
        prompt: "Extract key info: {{document}}"
        output: extraction
      
      - name: classify
        prompt: "Classify: {{document}}"
        output: classification
      
      - name: entities
        prompt: "Extract entities: {{document}}"
        output: entities
    max_concurrent: 3
    aggregate: merge
  
  # 3. Focused analysis
  - name: deep_analysis
    prompt: |
      Deep analysis:
      {{analyze_all}}
      
      Provide insights based on classification.
    output: insights
  
  # 4. Generate output
  - name: report
    prompt: |
      Create report:
      
      Classification: {{analyze_all.classification}}
      Entities: {{analyze_all.entities}}
      Extraction: {{analyze_all.extraction}}
      Insights: {{insights}}
      
      Format as markdown.
```

---

## Quick Reference

```yaml
# Basic analysis
extract → classify → analyze → report

# Multi-document
for_each documents → compare → synthesize

# Q&A
index → search → answer → verify

# Large documents
chunk → analyze_chunks → synthesize

# Parallel
parallel {extract, classify, validate} → combine
```

---

## Next Steps

- **[Data Pipeline Pattern](data-pipeline.md)** - ETL workflows
- **[Research Pattern](research-agent.md)** - Deep research
- **[Examples](../examples/)** - Working templates

---

**Process documents efficiently!** 📄
