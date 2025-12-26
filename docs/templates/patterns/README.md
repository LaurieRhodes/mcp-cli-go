# Template Patterns

Common design patterns for building robust, reusable templates.

---

## Available Patterns

### Core Patterns

These are complete, production-ready workflow patterns:

- **[Research Agent](research-agent.md)** - Autonomous research with iteration and follow-up
- **[Document Analysis](document-analysis.md)** - Systematic document processing pipelines
- **[Data Pipeline](data-pipeline.md)** - ETL (Extract-Transform-Load) workflows
- **[Multi-Provider Validation](validation.md)** - Get consensus from multiple AIs
- **[Conditional Routing](conditional-routing.md)** - Smart request routing to best handler

### Fundamental Building Blocks

These are simple patterns you can combine:

- **ETL (Extract-Transform-Load)** - Process data through stages
- **Map-Reduce** - Process many items, then combine results
- **Pipeline** - Sequential refinement through multiple steps
- **Fan-Out Fan-In** - Parallel processing with synthesis
- **Circuit Breaker** - Graceful fallback when things fail

---

## Pattern Selection Guide

**Choose the right pattern for your task:**

| Pattern | Best For | Complexity | Speed | Cost |
|---------|----------|------------|-------|------|
| **[Routing](conditional-routing.md)** | Classifying & directing | Low | ⚡⚡⚡ | $ |
| **[Document Analysis](document-analysis.md)** | Processing documents | Medium | ⚡⚡ | $$ |
| **[Data Pipeline](data-pipeline.md)** | ETL & transformation | Medium | ⚡⚡⚡ | $$ |
| **[Validation](validation.md)** | Critical decisions | Medium | ⚡ | $$$$ |
| **[Research Agent](research-agent.md)** | Deep exploration | High | ⚡ | $$$$ |
| **ETL** | Data transformation | Low | ⚡⚡⚡ | $ |
| **Map-Reduce** | Batch processing | Low | ⚡⚡⚡ | $$ |
| **Pipeline** | Sequential steps | Low | ⚡⚡ | $$ |
| **Fan-Out Fan-In** | Parallel analysis | Medium | ⚡⚡⚡ | $$$ |
| **Circuit Breaker** | Error handling | Medium | ⚡⚡ | $$ |

**Legend:**
- Speed: ⚡ = slow, ⚡⚡⚡ = fast
- Cost: $ = cheap, $$$$ = expensive

---

## When to Use Each Pattern

### Research Agent Pattern

**What it does:** Autonomously researches a topic through multiple iterations, following up on discoveries and synthesizing findings.

**Use when:**
- Need comprehensive topic coverage (not just surface-level answers)
- Topic requires multiple research iterations
- Following up on discoveries matters
- Want autonomous exploration (AI decides what to research next)

**Examples:**
- Market research: "Analyze the competitive landscape for AI coding tools"
- Technical investigation: "Research best practices for Kubernetes security"
- Literature review: "Survey recent papers on quantum computing applications"
- Competitive analysis: "How do our features compare to top 3 competitors?"

**Why choose this:**
- More thorough than single web search
- AI follows interesting leads automatically
- Builds comprehensive knowledge base

**Why not choose this:**
- Expensive (multiple AI calls + web searches)
- Slow (iterative process)
- Overkill for simple questions

→ [Read Pattern Guide](research-agent.md)

---

### Document Analysis Pattern

**What it does:** Systematically processes documents through extraction, classification, and analysis stages.

**Use when:**
- Processing long documents (> 5 pages)
- Need structured information extraction
- Multiple analysis perspectives required (security + quality + compliance)
- Document classification and routing needed

**Examples:**
- Legal document review: Extract parties, obligations, risks from contracts
- Technical documentation: Check completeness, clarity, accuracy
- Financial reports: Extract metrics, calculate ratios, identify trends
- Contract analysis: Find key terms, deadlines, and liabilities

**Why choose this:**
- Handles long documents efficiently
- Structured, reproducible analysis
- Can process multiple documents in batch

**Why not choose this:**
- Too complex for simple summaries
- Use simple prompt for short documents

→ [Read Pattern Guide](document-analysis.md)

---

### Data Pipeline Pattern

**What it does:** Moves data through Extract → Transform → Load stages with validation and quality checks.

**Use when:**
- ETL (Extract-Transform-Load) workflows needed
- Data quality is critical (healthcare, finance)
- Multi-stage transformation required
- Reproducible, auditable pipelines needed

**Examples:**
- Log processing: Extract errors, classify severity, generate alerts
- Sales data ETL: Pull from CRM, clean, enrich, load to warehouse
- API integration: Fetch from multiple APIs, normalize, deduplicate, store
- Data migration: Extract from old system, transform format, load to new

**Why choose this:**
- Handles data quality validation
- Reproducible and auditable
- Good for scheduled/automated runs

**Why not choose this:**
- Too structured for exploratory analysis
- Overkill for one-time transformations

→ [Read Pattern Guide](data-pipeline.md)

---

### Multi-Provider Validation Pattern

**What it does:** Gets opinions from multiple AI models (Claude, GPT-4, DeepSeek) and builds consensus.

**Use when:**
- High-stakes decisions (medical, financial, legal)
- Fact-checking is critical
- Need confidence through consensus
- Reducing single-model bias matters

**Examples:**
- Medical information: Verify treatment recommendations from multiple AIs
- Financial analysis: Get investment advice from multiple models
- Code review: Critical security review for production code
- Decision validation: Verify important business decisions

**Why choose this:**
- Highest confidence through consensus
- Catches single-model errors
- Reduces bias

**Why not choose this:**
- Very expensive (3-5x normal cost)
- Slow (sequential AI calls)
- Overkill for routine tasks

**Cost example:** Single query costs $0.03, validation costs $0.09-0.15

→ [Read Pattern Guide](validation.md)

---

### Conditional Routing Pattern

**What it does:** Classifies requests and routes them to the best handler (AI model, workflow, or tool).

**Use when:**
- Different inputs need different processing (code vs documents vs data)
- Provider selection matters (GPT-4 for code, Claude for docs)
- Cost optimization important (cheap model for simple, expensive for complex)
- Processing complexity varies greatly

**Examples:**
- Support ticket routing: Bug → engineering, billing → finance team
- Code analysis: Python → Python expert prompt, Go → Go expert prompt
- Content classification: Urgent → fast model, detailed → best model
- Intent-based handling: Question → answer, task → execute, feedback → log

**Why choose this:**
- Saves money (cheap models for simple tasks)
- Better quality (right tool for each job)
- Faster (skip unnecessary processing)

**Why not choose this:**
- Adds classification overhead
- Not needed if all requests are similar

**Cost savings example:** Route 80% to free local model, 20% to paid API → save 80%

→ [Read Pattern Guide](conditional-routing.md)

---

## Pattern Combinations

Patterns work powerfully together. Here are proven combinations:

### Research + Validation

**Use case:** Deep research with fact-checking

```yaml
name: validated_research
steps:
  # Step 1: Autonomous research (expensive but thorough)
  - name: research_topic
    template: research_agent
    template_input: "{{input_data.question}}"
    output: findings
  
  # Step 2: Validate key claims (multiple AIs check facts)
  - name: validate_findings
    template: multi_provider_validation
    template_input: "{{findings}}"
    output: validated_findings
  
  # Step 3: Generate final report with confidence levels
  - name: create_report
    prompt: |
      Create research report with confidence indicators:
      {{validated_findings}}
      
      Mark claims as:
      - HIGH CONFIDENCE: All AIs agree
      - MEDIUM CONFIDENCE: Most AIs agree
      - LOW CONFIDENCE: AIs disagree
```

**Usage:**
```bash
mcp-cli --template validated_research --input-data '{
  "question": "What are the proven benefits of intermittent fasting?"
}'
```

**Why combine these:**
- Research finds comprehensive information
- Validation ensures accuracy
- Perfect for high-stakes research

**Cost:** High (research + 3x validation)
**Time:** Slow (sequential processing)
**Quality:** Highest confidence

---

### Routing + Pipeline

**Use case:** Route requests to appropriate processing pipeline

```yaml
name: smart_routing_pipeline
steps:
  # Step 1: Classify the request type
  - name: classify
    template: conditional_routing
    template_input: "{{input_data}}"
    output: category
  
  # Step 2: Route data requests to ETL pipeline
  - name: process_data
    condition: "{{category}} == 'data'"
    template: data_pipeline
    template_input: "{{input_data}}"
    output: result
  
  # Step 3: Route document requests to document pipeline
  - name: process_document
    condition: "{{category}} == 'document'"
    template: document_analysis
    template_input: "{{input_data}}"
    output: result
```

**Why combine these:**
- One entry point handles everything
- Each type gets appropriate processing
- Easy to add new categories

**Cost:** Moderate (only runs one pipeline)
**Time:** Fast (skips unnecessary work)

---

### Document Analysis + Validation

**Use case:** Analyze documents with fact-checking

```yaml
name: validated_analysis
steps:
  # Step 1: Analyze document
  - name: analyze
    template: document_analysis
    template_input: "{{input_data.document}}"
    output: analysis
  
  # Step 2: Validate key findings
  - name: validate
    template: multi_provider_validation
    template_input: "{{analysis}}"
    output: validated_analysis
  
  # Step 3: Generate report with confidence scores
  - name: report
    prompt: |
      Create analysis report:
      {{validated_analysis}}
      
      Include confidence scores for each finding.
```

**Why combine these:**
- Thorough document analysis
- Validated findings
- Good for legal/financial docs

**Cost:** High
**Quality:** Very high confidence

---

## Fundamental Patterns

These are simple building blocks you can combine to create custom workflows.

### ETL (Extract-Transform-Load)

**What it is:** Get data from somewhere, clean/modify it, put it somewhere else.

**Structure:**
```yaml
steps:
  # Extract: Get the data
  - name: extract
    prompt: "Extract data from: {{input_data.source}}"
    output: raw_data
  
  # Transform: Clean and modify
  - name: transform
    prompt: "Clean and format: {{raw_data}}"
    output: clean_data
  
  # Load: Save results
  - name: load
    prompt: "Save to: {{input_data.destination}}"
```

**Use when:** Moving and cleaning data between systems

**Example:** Pull customer data from CSV, fix phone formats, save to database

→ See [Data Pipeline Pattern](data-pipeline.md) for complete version

---

### Map-Reduce

**What it is:** Process each item in a list, then combine all the results.

**Structure:**
```yaml
steps:
  # Map: Process each item
  - name: process_items
    for_each: "{{input_data.items}}"
    item_name: item
    prompt: "Process: {{item}}"
    output: results
  
  # Reduce: Combine results
  - name: combine
    prompt: "Combine all results: {{results}}"
```

**Use when:** Batch processing multiple items

**Example:** Analyze 50 customer reviews, then summarize common themes

**Performance:** Can parallelize the "map" phase for speed

---

### Pipeline

**What it is:** Pass output through multiple refinement stages.

**Structure:**
```yaml
steps:
  # Stage 1: Initial draft
  - name: draft
    prompt: "Create draft: {{input_data}}"
    output: draft
  
  # Stage 2: Improve it
  - name: improve
    prompt: "Improve: {{draft}}"
    output: improved
  
  # Stage 3: Final polish
  - name: polish
    prompt: "Polish: {{improved}}"
    output: final
```

**Use when:** Quality improves with multiple passes

**Example:** Draft email → improve tone → fix grammar → final version

**Cost:** Higher (multiple AI calls), but quality improves each stage

---

### Fan-Out Fan-In

**What it is:** Run multiple analyses in parallel, then combine results.

**Structure:**
```yaml
steps:
  # Fan-Out: Parallel processing
  - name: parallel_analysis
    parallel:
      - name: task1
        prompt: "Analyze for X: {{input_data}}"
      - name: task2
        prompt: "Analyze for Y: {{input_data}}"
      - name: task3
        prompt: "Analyze for Z: {{input_data}}"
    max_concurrent: 3
    aggregate: merge
    output: all_results
  
  # Fan-In: Synthesize
  - name: synthesize
    prompt: "Combine insights: {{all_results}}"
```

**Use when:** Need independent perspectives combined

**Example:** Code review from 3 angles (security, performance, style) → combined report

**Performance:** 3x faster than sequential (but same API cost)

---

### Circuit Breaker

**What it is:** Try main approach, fall back to alternative if it fails.

**Structure:**
```yaml
steps:
  # Try primary approach
  - name: primary
    prompt: "{{input_data}}"
    error_handling:
      on_failure: continue
      default_output: "FAILED"
    output: result
  
  # Fallback if primary fails
  - name: fallback
    condition: "{{result}} contains 'FAILED'"
    prompt: "Use fallback approach: {{input_data}}"
    output: result
  
  # Final fallback
  - name: default
    condition: "{{result}} contains 'FAILED'"
    prompt: "Use simplest approach: {{input_data}}"
```

**Use when:** Need reliability and graceful degradation

**Example:** Try expensive API → if fails, use free API → if fails, use cached data

**Reliability:** Ensures something always works, even if not ideal

---

## Real-World Examples

### Example 1: Comprehensive Code Review

```yaml
# Combines: Routing + Fan-Out + Validation
name: comprehensive_review

steps:
  # Route by language
  - template: conditional_routing
    output: language
  
  # Parallel analysis
  - parallel:
      - template: security_analysis
      - template: quality_analysis
      - template: performance_analysis
    aggregate: merge
  
  # Validate findings
  - template: multi_provider_validation
    template_input: "{{parallel_results}}"
```

### Example 2: Research with Fact-Checking

```yaml
# Combines: Research + Validation
name: validated_research

steps:
  # Research topic
  - template: research_agent
    output: findings
  
  # Validate key claims
  - template: multi_provider_validation
    template_input: "{{findings}}"
    output: validated
  
  # Generate report
  - prompt: |
      Create report with confidence levels:
      {{validated}}
```

### Example 3: Smart Document Processing

```yaml
# Combines: Routing + Document Analysis + Pipeline
name: smart_doc_processing

steps:
  # Classify document
  - template: conditional_routing
    output: doc_type
  
  # Route to appropriate analyzer
  - name: analyze_legal
    condition: "{{doc_type}} == 'legal'"
    template: legal_doc_analysis
  
  - name: analyze_technical
    condition: "{{doc_type}} == 'technical'"
    template: technical_doc_analysis
  
  # Process results through pipeline
  - template: data_pipeline
    template_input: "{{analysis_result}}"
```

---

## Pattern Best Practices

### 1. Start Simple

```yaml
# Good: Start with basic pattern
- name: extract
- name: transform
- name: load

# Then add complexity
- name: extract
  error_handling: ...
- name: validate
- name: transform
```

### 2. Compose Patterns

```yaml
# Good: Combine patterns
- template: routing
- template: validation
- template: pipeline

# Bad: One giant template
- name: do_everything
  prompt: "..." # Too complex
```

### 3. Use Appropriate Pattern

```yaml
# Good: Right pattern for task
- template: research_agent  # For research
- template: data_pipeline   # For ETL

# Bad: Wrong pattern
- template: data_pipeline   # For research (inefficient)
```

### 4. Error Handling

```yaml
# Good: Every pattern has error handling
- error_handling:
    on_failure: retry
    max_retries: 3

# Bad: No error handling
- prompt: "..."  # Will fail unexpectedly
```

---

## Performance Considerations

### Pattern Performance

| Pattern | Speed | Cost | Accuracy |
|---------|-------|------|----------|
| **Routing** | ⚡⚡⚡ | $ | ⭐⭐⭐ |
| **Pipeline** | ⚡⚡ | $$ | ⭐⭐⭐⭐ |
| **Map-Reduce** | ⚡⚡⚡ | $$$ | ⭐⭐⭐ |
| **Fan-Out** | ⚡⚡⚡ | $$$ | ⭐⭐⭐⭐ |
| **Research** | ⚡ | $$$$ | ⭐⭐⭐⭐⭐ |
| **Validation** | ⚡ | $$$$ | ⭐⭐⭐⭐⭐ |

### Optimization Tips

**1. Parallel where possible:**
```yaml
# Fast: Parallel
parallel:
  - task1
  - task2
  - task3

# Slow: Sequential
- task1
- task2
- task3
```

**2. Cache expensive operations:**
```yaml
# Cache reference data
- name: load_reference
  output: cache

# Reuse cache
- prompt: "Use {{cache}}"
- prompt: "Use {{cache}}"
```

**3. Route to appropriate models:**
```yaml
# Cheap for simple
- condition: "simple"
  provider: ollama

# Expensive for complex
- condition: "complex"
  provider: anthropic
```

---

## Pattern Catalog

### By Use Case

**Research & Analysis:**
- [Research Agent](research-agent.md)
- [Multi-Provider Validation](validation.md)

**Document Processing:**
- [Document Analysis](document-analysis.md)
- Pipeline Pattern

**Data Processing:**
- [Data Pipeline](data-pipeline.md)
- ETL Pattern
- Map-Reduce Pattern

**Optimization:**
- [Conditional Routing](conditional-routing.md)
- Circuit Breaker Pattern

---

## Quick Reference

**Pattern decision tree:**

```
Need to answer a question?
├─ Simple factual → Single prompt (no pattern needed)
├─ Requires research → Research Agent Pattern
└─ Need high confidence → Research + Validation

Need to process data?
├─ Single transformation → ETL Pattern
├─ Multiple items → Map-Reduce Pattern
└─ Complex pipeline → Data Pipeline Pattern

Need to process documents?
├─ Short document → Single prompt
├─ Long document → Document Analysis Pattern
└─ Multiple documents → Document Analysis + loops

Need to make decision?
├─ Routine → Single prompt
└─ High-stakes → Multi-Provider Validation

Need to handle variety?
└─ Different types of input → Conditional Routing Pattern
```

**Pattern workflows:**

```yaml
# Research workflow
decompose_question → search_each_part → follow_up → synthesize

# Document workflow
extract_info → classify_content → analyze_sections → create_report

# Data Pipeline workflow
extract_from_source → validate_data → transform_data → load_to_destination

# Validation workflow
ask_claude → ask_gpt4 → ask_deepseek → compare → find_consensus

# Routing workflow
classify_input → route_to_handler_A | handler_B | handler_C

# ETL workflow
extract → transform → load

# Map-Reduce workflow
for_each_item → process → aggregate_results

# Pipeline workflow
draft → improve → polish → finalize

# Fan-Out Fan-In workflow
parallel_process → wait_for_all → combine_results

# Circuit Breaker workflow
try_primary → if_fail_try_secondary → if_fail_use_default
```

---

## Next Steps

**New to patterns?**
1. Start with [Conditional Routing](conditional-routing.md)
2. Try [Data Pipeline](data-pipeline.md)
3. Explore [Document Analysis](document-analysis.md)

**Building complex workflows?**
1. Study [Research Agent](research-agent.md)
2. Learn [Multi-Provider Validation](validation.md)
3. Combine multiple patterns

**Need examples?**
1. See [Template Examples](../examples/)
2. Read [Authoring Guide](../authoring-guide.md)
3. Check [Automation Guide](../../guides/automation.md)

---

**Choose the right pattern, build better templates!** 🎯
