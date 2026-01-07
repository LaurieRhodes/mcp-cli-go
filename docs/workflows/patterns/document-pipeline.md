# Document Pipeline Pattern

**Sequential document processing through extraction, transformation, and analysis stages.**

---

## Overview

The Document Pipeline pattern processes documents through multiple stages: extract → transform → analyze → report.

**Key features:**
- Sequential processing with `needs` dependencies
- Property inheritance for consistent configuration
- Template composition for reusable stages

---

## When to Use

**Use when:**
- Processing documents systematically
- Need structured extraction and analysis
- Multiple processing stages required
- Reproducible document workflows needed

**Examples:**
- Contract analysis (extract parties, terms, risks)
- Technical documentation review (check completeness, accuracy)
- Financial report processing (extract metrics, calculate ratios)
- Legal document review (find key clauses, obligations)

---

## Basic Structure

```yaml
$schema: "workflow/v2.0"
name: document_pipeline
version: 1.0.0

execution:
  provider: anthropic
  model: claude-sonnet-4
  servers: [filesystem]

steps:
  # Stage 1: Extract
  - name: extract
    run: "Extract key information from: {{input}}"

  # Stage 2: Transform
  - name: transform
    needs: [extract]
    run: "Structure the extracted data: {{extract}}"

  # Stage 3: Analyze
  - name: analyze
    needs: [transform]
    run: "Analyze the structured data: {{transform}}"

  # Stage 4: Report
  - name: report
    needs: [analyze]
    run: "Generate report from: {{analyze}}"
```

---

## Complete Examples

### Example 1: Contract Analysis

```yaml
$schema: "workflow/v2.0"
name: contract_analyzer
version: 1.0.0
description: Analyze legal contracts systematically

execution:
  provider: anthropic
  model: claude-sonnet-4
  servers: [filesystem]
  temperature: 0.3

steps:
  # Extract basic information
  - name: extract_parties
    run: |
      Extract from this contract:
      {{input}}
      
      Identify:
      - All parties involved
      - Their roles
      - Contact information

  # Extract terms
  - name: extract_terms
    needs: [extract_parties]
    run: |
      Extract contract terms from:
      {{input}}
      
      Find:
      - Obligations of each party
      - Payment terms
      - Deliverables
      - Timelines

  # Extract risks
  - name: extract_risks
    needs: [extract_parties]
    run: |
      Extract risk-related clauses from:
      {{input}}
      
      Identify:
      - Liability limitations
      - Indemnification clauses
      - Termination conditions
      - Dispute resolution

  # Analyze obligations
  - name: analyze_obligations
    needs: [extract_terms]
    run: |
      Analyze obligations:
      {{extract_terms}}
      
      For each obligation:
      - Is it clearly defined?
      - Is it measurable?
      - Is it realistic?
      - Are there dependencies?

  # Risk assessment
  - name: assess_risks
    needs: [extract_risks]
    run: |
      Assess risks:
      {{extract_risks}}
      
      For each risk:
      - Severity (high/medium/low)
      - Likelihood
      - Mitigation measures
      - Acceptability

  # Generate report
  - name: final_report
    needs: [analyze_obligations, assess_risks]
    run: |
      Generate contract analysis report:
      
      Parties: {{extract_parties}}
      Terms: {{extract_terms}}
      Risks: {{extract_risks}}
      Obligation Analysis: {{analyze_obligations}}
      Risk Assessment: {{assess_risks}}
      
      Provide:
      - Executive summary
      - Key obligations list
      - Risk summary
      - Recommendations
```

**Usage:**
```bash
./mcp-cli --workflow contract_analyzer \
  --server filesystem \
  --input-data "$(cat contract.pdf)"
```

---

### Example 2: Technical Documentation Review

```yaml
$schema: "workflow/v2.0"
name: doc_reviewer
version: 1.0.0
description: Review technical documentation quality

execution:
  provider: anthropic
  model: claude-sonnet-4
  temperature: 0.5

steps:
  # Check structure
  - name: check_structure
    run: |
      Check documentation structure:
      {{input}}
      
      Verify:
      - Table of contents present
      - Logical section organization
      - Proper heading hierarchy
      - Navigation elements

  # Check completeness
  - name: check_completeness
    needs: [check_structure]
    run: |
      Check documentation completeness:
      {{input}}
      
      Verify:
      - All features documented
      - Examples provided
      - Prerequisites listed
      - Troubleshooting section

  # Check accuracy
  - name: check_accuracy
    needs: [check_completeness]
    run: |
      Check technical accuracy:
      {{input}}
      
      Verify:
      - Code examples work
      - Commands are correct
      - API references accurate
      - Links are valid

  # Check clarity
  - name: check_clarity
    needs: [check_accuracy]
    run: |
      Check documentation clarity:
      {{input}}
      
      Assess:
      - Language is clear
      - Jargon explained
      - Audience appropriate
      - Steps are numbered

  # Generate review report
  - name: review_report
    needs: [check_structure, check_completeness, check_accuracy, check_clarity]
    run: |
      Generate documentation review report:
      
      Structure: {{check_structure}}
      Completeness: {{check_completeness}}
      Accuracy: {{check_accuracy}}
      Clarity: {{check_clarity}}
      
      Provide:
      - Overall quality score (1-10)
      - Strengths
      - Issues found
      - Recommendations for improvement
```

---

### Example 3: Financial Report Processing

```yaml
$schema: "workflow/v2.0"
name: financial_processor
version: 1.0.0
description: Process financial reports

execution:
  provider: deepseek
  model: deepseek-chat
  servers: [filesystem]

steps:
  # Extract financials
  - name: extract_financials
    run: |
      Extract financial data from:
      {{input}}
      
      Extract:
      - Revenue figures
      - Expense breakdown
      - Net income
      - Cash flow
      - Balance sheet items

  # Calculate ratios
  - name: calculate_ratios
    needs: [extract_financials]
    run: |
      Calculate financial ratios:
      {{extract_financials}}
      
      Calculate:
      - Profit margin
      - Current ratio
      - Debt-to-equity ratio
      - Return on assets
      - Operating cash flow ratio

  # Trend analysis
  - name: trend_analysis
    needs: [extract_financials]
    run: |
      Analyze financial trends:
      {{extract_financials}}
      
      Compare to previous periods:
      - Revenue growth
      - Expense trends
      - Margin changes
      - Cash position changes

  # Generate insights
  - name: generate_insights
    needs: [calculate_ratios, trend_analysis]
    run: |
      Generate financial insights:
      
      Ratios: {{calculate_ratios}}
      Trends: {{trend_analysis}}
      
      Provide:
      - Financial health assessment
      - Key strengths
      - Areas of concern
      - Recommendations
```

---

## Pattern with Template Composition

Break complex pipelines into reusable templates:

**Main pipeline:**
```yaml
$schema: "workflow/v2.0"
name: document_analysis
version: 1.0.0

execution:
  provider: anthropic
  model: claude-sonnet-4

steps:
  # Use extraction template
  - name: extract
    template:
      name: document_extractor
      with:
        document: "{{input}}"

  # Use analysis template
  - name: analyze
    needs: [extract]
    template:
      name: document_analyzer
      with:
        data: "{{extract}}"

  # Use reporting template
  - name: report
    needs: [analyze]
    template:
      name: report_generator
      with:
        analysis: "{{analyze}}"
```

**Extraction template (document_extractor.yaml):**
```yaml
$schema: "workflow/v2.0"
name: document_extractor
version: 1.0.0

execution:
  provider: anthropic
  model: claude-sonnet-4

steps:
  - name: extract_text
    run: "Extract text from: {{document}}"

  - name: extract_metadata
    run: "Extract metadata from: {{document}}"

  - name: structure_data
    needs: [extract_text, extract_metadata]
    run: |
      Structure extracted data:
      Text: {{extract_text}}
      Metadata: {{extract_metadata}}
```

---

## Property Inheritance

Use execution context for consistent configuration:

```yaml
$schema: "workflow/v2.0"
name: consistent_pipeline
version: 1.0.0

execution:
  # Default for all steps
  provider: anthropic
  model: claude-sonnet-4
  servers: [filesystem]
  temperature: 0.3
  logging: verbose

steps:
  - name: extract
    run: "..."
    # Inherits: anthropic/claude-sonnet-4, temp=0.3

  - name: analyze
    needs: [extract]
    run: "..."
    # Inherits: anthropic/claude-sonnet-4, temp=0.3

  - name: report
    needs: [analyze]
    run: "..."
    temperature: 0.7  # Override for creative report
    # Uses: anthropic/claude-sonnet-4, temp=0.7
```

---

## Best Practices

### 1. Use Clear Stage Names

```yaml
# ✅ Good: Descriptive names
steps:
  - name: extract_parties
  - name: extract_terms
  - name: analyze_risks

# ❌ Bad: Generic names
steps:
  - name: step1
  - name: step2
  - name: step3
```

### 2. Declare Dependencies

```yaml
# ✅ Good: Explicit dependencies
- name: analyze
  needs: [extract, transform]

# ❌ Bad: Implicit ordering
- name: analyze  # Unclear what it needs
```

### 3. Use Appropriate Temperature

```yaml
# ✅ Good: Different temps for different stages
- name: extract
  temperature: 0.1  # Deterministic extraction

- name: analyze
  temperature: 0.5  # Balanced analysis

- name: report
  temperature: 0.7  # Creative reporting
```

### 4. Provide Context

```yaml
# ✅ Good: Clear instructions
- name: extract
  run: |
    Extract these specific items:
    - Item 1
    - Item 2
    
    From document: {{input}}

# ❌ Bad: Vague instructions
- name: extract
  run: "Extract stuff from {{input}}"
```

---

## Error Handling

Add error handling to critical steps:

```yaml
steps:
  - name: extract
    run: "Extract data from: {{input}}"
    on_error:
      retry: 3
      backoff: exponential
      fallback: extract_manual

  - name: extract_manual
    run: |
      Manual extraction fallback:
      {{input}}
      
      Use simpler extraction method.
```

---

## Performance

**Sequential execution:**
- Step 1: 2-4s
- Step 2: 2-4s
- Step 3: 2-4s
- Total: 6-12s for 3 steps

**Optimization:**
- Steps with same `needs` could run in parallel (not yet implemented)
- Use faster models where appropriate
- Cache extracted data for reuse

---

## Common Patterns

### ETL Pattern

```yaml
steps:
  # Extract
  - name: extract
    run: "Extract from: {{source}}"

  # Transform
  - name: transform
    needs: [extract]
    run: "Transform: {{extract}}"

  # Load
  - name: load
    needs: [transform]
    run: "Load to: {{destination}}"
```

### Multi-Stage Analysis

```yaml
steps:
  # Stage 1: Syntax
  - name: syntax_check
    run: "Check syntax: {{input}}"

  # Stage 2: Semantics
  - name: semantic_check
    needs: [syntax_check]
    run: "Check semantics: {{input}}"

  # Stage 3: Quality
  - name: quality_check
    needs: [semantic_check]
    run: "Check quality: {{input}}"
```

### Progressive Refinement

```yaml
steps:
  # Draft
  - name: draft
    run: "Create draft: {{input}}"

  # Improve
  - name: improve
    needs: [draft]
    run: "Improve: {{draft}}"

  # Polish
  - name: polish
    needs: [improve]
    run: "Polish: {{improve}}"
```

---

## Related Patterns

- **[Iterative Refinement](iterative-refinement.md)** - Add loops for quality improvement
- **[Consensus Validation](consensus-validation.md)** - Validate critical findings

---

## See Also

- [Schema Reference](../SCHEMA.md) - Step dependencies
- [Authoring Guide](../AUTHORING_GUIDE.md) - Writing workflows

---

**Systematic document processing!** 📄
