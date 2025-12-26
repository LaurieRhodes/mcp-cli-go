# Multi-Provider Validation Pattern

Use multiple AI providers to validate, verify, and build consensus.

---

## Overview

The **Multi-Provider Validation Pattern** leverages multiple AI providers to validate critical information.

**What it does:**
- Gets answers from 3+ different AI models (Claude, GPT-4, DeepSeek, etc.)
- Compares their responses to find agreement and disagreement
- Builds consensus where models agree
- Flags areas where models disagree for further research

**Why use multiple AI providers:**
- **Catch errors:** If one AI hallucinates, others likely won't make same mistake
- **Build confidence:** Agreement = high confidence, disagreement = needs verification
- **Reduce bias:** Each AI has different training, perspectives balance out
- **Critical decisions:** Medical, financial, legal info needs highest accuracy

**Use when:**
- High-stakes decisions (affects money, health, legal standing)
- Fact-checking is critical (journalism, research, compliance)
- Need confidence in results (can't afford to be wrong)
- Reducing bias matters (want balanced perspective)

**Don't use when:**
- Simple questions with known answers ("What's 2+2?")
- Creative work (multiple valid approaches)
- Cost-sensitive (3-5x more expensive)
- Speed matters (3-5x slower)

**Cost reality:**
- Single provider: $0.03 per question
- Three providers: $0.09 per question
- **3x more expensive** but much higher confidence

---

## Pattern Structure

```
Input → Provider 1 (Claude) → Provider 2 (GPT-4) → Provider 3 (DeepSeek) → Compare → Consensus
```

**What happens:**
1. Same question asked to 3 different AIs
2. Each responds independently
3. Compare step analyzes where they agree/disagree
4. Consensus step provides validated answer with confidence level

### Basic Validation Pattern

**What it does:** Asks same question to 3 AIs, compares answers, provides consensus.

**Use when:** Need to verify important information.

```yaml
name: multi_provider_validation
description: Validate with multiple AI providers
version: 1.0.0

config:
  defaults:
    provider: anthropic
    model: claude-sonnet-4

steps:
  # Provider 1: Claude (strong at analysis)
  - name: claude_response
    provider: anthropic
    model: claude-sonnet-4
    prompt: |
      {{input_data.task}}
      
      Be accurate and cite sources where possible.
    output: claude_answer
  
  # Provider 2: GPT-4 (strong at reasoning)
  - name: gpt_response
    provider: openai
    model: gpt-4o
    prompt: |
      {{input_data.task}}
      
      Be accurate and cite sources where possible.
    output: gpt_answer
  
  # Provider 3: DeepSeek (strong at math/logic)
  - name: deepseek_response
    provider: deepseek
    prompt: |
      {{input_data.task}}
      
      Be accurate and cite sources where possible.
    output: deepseek_answer
  
  # Compare all three responses
  - name: compare_responses
    prompt: |
      Compare these three AI responses:
      
      **Claude says:**
      {{claude_answer}}
      
      **GPT-4 says:**
      {{gpt_answer}}
      
      **DeepSeek says:**
      {{deepseek_answer}}
      
      Analyze:
      1. Where do all three agree? (HIGH CONFIDENCE)
      2. Where do two agree but one differs? (MEDIUM CONFIDENCE)
      3. Where do all three disagree? (LOW CONFIDENCE - needs research)
      4. Which response seems most accurate and why?
      
      Return structured comparison.
    output: comparison
  
  # Build consensus answer
  - name: build_consensus
    prompt: |
      Based on this comparison:
      {{comparison}}
      
      Provide:
      1. **Consensus answer:** Where models agree
      2. **Confidence level:**
         - HIGH: All 3 agree
         - MEDIUM: 2 of 3 agree
         - LOW: All disagree or unclear
      3. **Areas of uncertainty:** Where they disagree
      4. **Recommendation:**
         - ACCEPT: High confidence, use answer
         - VERIFY: Medium confidence, double-check critical parts
         - RESEARCH MORE: Low confidence, needs investigation
      5. **Evidence:** Sources or reasoning provided
```

**Usage:**
```bash
# Example 1: Medical question (high stakes)
mcp-cli --template multi_provider_validation --input-data '{
  "task": "What are the evidence-based treatments for type 2 diabetes?"
}'

# Example 2: Financial calculation
mcp-cli --template multi_provider_validation --input-data '{
  "task": "If I invest $10,000 at 7% annual return for 30 years with monthly compounding, what will I have?"
}'

# Example 3: Legal question
mcp-cli --template multi_provider_validation --input-data '{
  "task": "Under California law, what is the statute of limitations for breach of contract?"
}'
```

**What happens:**
1. Claude, GPT-4, and DeepSeek all answer independently
2. Comparison analyzes their responses
3. Builds consensus with confidence level
4. Provides recommendation

**Example - When AIs Agree (High Confidence):**
```
Question: "What is the capital of France?"

Claude: "Paris"
GPT-4: "Paris"
DeepSeek: "Paris"

Comparison: All three agree
Confidence: HIGH
Recommendation: ACCEPT
Answer: "Paris is the capital of France"
```

**Example - When AIs Disagree (Low Confidence):**
```
Question: "Will cryptocurrency prices rise in 2025?"

Claude: "Impossible to predict with certainty. Multiple factors including regulation, adoption, and macro conditions will influence prices."

GPT-4: "Historical patterns suggest volatility will continue, but specific price predictions are unreliable."

DeepSeek: "Based on technical analysis, bullish indicators suggest potential upward trend, but this is speculative."

Comparison: All three provide different perspectives
Confidence: LOW
Recommendation: RESEARCH MORE
Issue: This is inherently unpredictable - no one can reliably forecast markets
```

**Cost:**
- 3 AI calls + 2 synthesis = 5 total calls
- ~$0.09-0.15 per validation
- **3-5x more expensive** than single provider

**When it's worth it:**
- Medical decisions: Absolutely worth 3x cost
- Financial advice: Yes, if substantial money involved
- Legal questions: Yes, mistakes are costly
- Casual questions: No, not worth the cost

---

## Pattern: Fact-Checking Validation

Validate factual claims with multiple sources.

```yaml
name: fact_check_validation
steps:
  # Step 1: Extract claims
  - name: extract_claims
    prompt: |
      Extract factual claims from:
      {{content}}
      
      Return as list of verifiable statements.
    output: claims
  
  # Step 2: Verify each claim with multiple providers
  - name: verify_claims
    for_each: "{{claims}}"
    item_name: claim
    parallel:
      - name: verify_claude
        provider: anthropic
        prompt: "Is this claim accurate? {{claim}}"
        output: claude_verification
      
      - name: verify_gpt
        provider: openai
        prompt: "Is this claim accurate? {{claim}}"
        output: gpt_verification
      
      - name: verify_search
        servers: [brave-search]
        prompt: "Search for evidence: {{claim}}"
        output: search_evidence
    max_concurrent: 3
    aggregate: merge
    output: verifications
  
  # Step 3: Assess consensus
  - name: assess_consensus
    prompt: |
      For each claim, assess verification results:
      {{verifications}}
      
      For each claim provide:
      - Claim text
      - Verification status (verified/disputed/uncertain)
      - Confidence level (high/medium/low)
      - Evidence summary
      - Recommendation (accept/reject/needs-research)
    output: fact_check_report
```

---

## Pattern: Code Review Validation

Validate code analysis with multiple providers.

```yaml
name: code_review_validation
steps:
  # Parallel code review by multiple providers
  - name: parallel_review
    parallel:
      # Claude for security focus
      - name: claude_security
        provider: anthropic
        model: claude-sonnet-4
        system_prompt: "You are a security expert"
        prompt: |
          Review code for security issues:
          {{code}}
          
          List vulnerabilities with severity.
        output: claude_security
      
      # GPT-4 for best practices
      - name: gpt_quality
        provider: openai
        model: gpt-4o
        system_prompt: "You are a code quality expert"
        prompt: |
          Review code for quality issues:
          {{code}}
          
          Focus on maintainability and best practices.
        output: gpt_quality
      
      # DeepSeek for performance
      - name: deepseek_performance
        provider: deepseek
        system_prompt: "You are a performance optimization expert"
        prompt: |
          Review code for performance issues:
          {{code}}
          
          Find bottlenecks and optimization opportunities.
        output: deepseek_performance
    max_concurrent: 3
    aggregate: merge
  
  # Cross-validate findings
  - name: cross_validate
    prompt: |
      Cross-validate code review findings:
      
      Security (Claude): {{parallel_review.claude_security}}
      Quality (GPT-4): {{parallel_review.gpt_quality}}
      Performance (DeepSeek): {{parallel_review.deepseek_performance}}
      
      For each finding:
      - Is it confirmed by multiple reviewers?
      - What's the severity consensus?
      - Are there conflicting opinions?
      
      Create consolidated report with confidence levels.
    output: validated_review
```

---

## Pattern: Decision Validation

Validate important decisions with multiple perspectives.

```yaml
name: decision_validation
steps:
  # Present decision to multiple providers
  - name: get_perspectives
    parallel:
      # Optimistic view
      - name: optimistic
        provider: anthropic
        system_prompt: "You are an optimistic advisor. Focus on opportunities."
        prompt: |
          Analyze this decision:
          {{decision}}
          
          What are the benefits and opportunities?
        output: optimistic_view
      
      # Pessimistic view
      - name: pessimistic
        provider: openai
        system_prompt: "You are a risk-aware advisor. Focus on risks."
        prompt: |
          Analyze this decision:
          {{decision}}
          
          What are the risks and downsides?
        output: pessimistic_view
      
      # Neutral view
      - name: neutral
        provider: ollama
        model: qwen2.5:32b
        system_prompt: "You are a balanced advisor."
        prompt: |
          Analyze this decision:
          {{decision}}
          
          Provide balanced analysis of pros and cons.
        output: neutral_view
    max_concurrent: 3
    aggregate: merge
  
  # Synthesize decision recommendation
  - name: synthesize_recommendation
    prompt: |
      Synthesize decision recommendation from these perspectives:
      
      Optimistic: {{get_perspectives.optimistic_view}}
      Pessimistic: {{get_perspectives.pessimistic_view}}
      Neutral: {{get_perspectives.neutral_view}}
      
      Provide:
      - Balanced assessment
      - Key risks to mitigate
      - Opportunities to maximize
      - Final recommendation with confidence level
      - Decision criteria checklist
```

---

## Pattern: Answer Verification

Verify answers to questions with multiple providers.

```yaml
name: answer_verification
steps:
  # Get answer from primary provider
  - name: primary_answer
    provider: anthropic
    model: claude-sonnet-4
    prompt: "{{question}}"
    output: answer
  
  # Verify with secondary provider
  - name: verify_answer
    provider: openai
    model: gpt-4o
    prompt: |
      Verify this answer:
      
      Question: {{question}}
      Answer: {{answer}}
      
      Is this answer:
      - Accurate? (yes/no and why)
      - Complete? (yes/no and what's missing)
      - Well-reasoned? (yes/no and issues)
    output: verification
  
  # Check facts with search
  - name: fact_check
    servers: [brave-search]
    prompt: |
      Verify facts in answer:
      {{answer}}
      
      Search for evidence supporting or contradicting claims.
    output: fact_check_results
  
  # Final validation
  - name: final_validation
    prompt: |
      Validate answer:
      
      Original: {{answer}}
      Verification: {{verification}}
      Fact Check: {{fact_check_results}}
      
      Provide:
      - Validated answer (corrected if needed)
      - Confidence level (high/medium/low)
      - Supporting evidence
      - Caveats or limitations
```

---

## Real-World Examples

### Example 1: Medical Information Validation

```yaml
name: medical_validation
description: Validate medical information with multiple sources
version: 1.0.0

steps:
  # Get medical information from multiple providers
  - name: get_medical_info
    parallel:
      - name: claude_medical
        provider: anthropic
        prompt: |
          Provide medical information about: {{condition}}
          
          Include:
          - Symptoms
          - Causes
          - Treatments
          - When to see a doctor
          
          Cite medical sources.
        output: claude_info
      
      - name: gpt_medical
        provider: openai
        prompt: |
          Provide medical information about: {{condition}}
          
          Include:
          - Symptoms
          - Causes
          - Treatments
          - When to see a doctor
          
          Cite medical sources.
        output: gpt_info
    max_concurrent: 2
    aggregate: merge
  
  # Search medical databases
  - name: search_medical_sources
    servers: [medical-search]
    prompt: "Search medical literature for: {{condition}}"
    output: medical_sources
  
  # Validate against medical sources
  - name: validate
    prompt: |
      Validate AI-provided information against medical sources:
      
      Claude: {{get_medical_info.claude_info}}
      GPT-4: {{get_medical_info.gpt_info}}
      Medical Sources: {{medical_sources}}
      
      For each piece of information:
      - Is it confirmed by medical sources?
      - Are there any inaccuracies?
      - What's the evidence quality?
      
      Provide validated medical information with confidence levels.
    output: validated_info
  
  # Generate disclaimer
  - name: add_disclaimer
    prompt: |
      Add appropriate medical disclaimer to:
      {{validated_info}}
      
      Emphasize:
      - Not medical advice
      - Consult healthcare professional
      - Emergency symptoms
```

### Example 2: Financial Analysis Validation

```yaml
name: financial_validation
steps:
  # Multiple providers analyze financials
  - name: financial_analysis
    parallel:
      - name: conservative_analysis
        provider: anthropic
        system_prompt: "Conservative financial analyst"
        prompt: "Analyze: {{financials}}"
        output: conservative
      
      - name: aggressive_analysis
        provider: openai
        system_prompt: "Growth-focused financial analyst"
        prompt: "Analyze: {{financials}}"
        output: aggressive
    aggregate: merge
  
  # Verify calculations
  - name: verify_calculations
    prompt: |
      Verify financial calculations:
      
      Conservative: {{financial_analysis.conservative}}
      Aggressive: {{financial_analysis.aggressive}}
      
      Check:
      - Math accuracy
      - Ratio calculations
      - Assumptions validity
    output: verified_calculations
  
  # Risk assessment
  - name: risk_assessment
    prompt: |
      Assess risk levels:
      
      Analyses: {{financial_analysis}}
      Verified: {{verified_calculations}}
      
      Provide balanced risk assessment with confidence level.
```

---

## Best Practices

### 1. Choose Diverse Providers

```yaml
# Good: Different providers, different strengths
parallel:
  - provider: anthropic     # Analysis
  - provider: openai        # Reasoning
  - provider: ollama        # Local, unbiased

# Bad: Same provider multiple times
parallel:
  - provider: openai
  - provider: openai  # No diversity
```

### 2. Use Appropriate System Prompts

```yaml
# Good: Different perspectives
- provider: anthropic
  system_prompt: "You are cautious and risk-aware"
- provider: openai
  system_prompt: "You are optimistic and opportunity-focused"

# Bad: Same instructions
- system_prompt: "Analyze this"
- system_prompt: "Analyze this"
```

### 3. Quantify Confidence

```yaml
# Good: Clear confidence levels
- prompt: |
    Rate confidence:
    - High: All providers agree, evidence strong
    - Medium: Providers mostly agree
    - Low: Providers disagree or evidence weak

# Bad: Vague confidence
- prompt: "Are you sure?"
```

### 4. Handle Disagreements

```yaml
# Good: Explicit disagreement handling
- name: handle_disagreement
  condition: "{{comparison}} contains 'DISAGREE'"
  prompt: |
    Providers disagree. 
    Research more: {{disagreement_points}}

# Bad: Ignore disagreements
- name: just_pick_one  # Risky!
```

---

## Cost Optimization

### Selective Validation

```yaml
# Validate only high-stakes items
- name: classify_stake
  prompt: "Rate importance: high/medium/low"

- name: full_validation
  condition: "{{stake}} == 'high'"
  parallel:
    - provider: anthropic
    - provider: openai
    - provider: deepseek

- name: single_provider
  condition: "{{stake}} == 'low'"
  provider: ollama  # Free local model
```

### Tiered Validation

```yaml
# Tier 1: Fast, cheap
- provider: ollama
  output: quick_answer

# Tier 2: Verify if uncertain
- condition: "{{quick_answer.confidence}} < 0.8"
  provider: anthropic

# Tier 3: Full validation if critical
- condition: "{{critical}} == true"
  parallel:
    - provider: anthropic
    - provider: openai
```

---

## Performance

### Parallel Execution

```yaml
# Fast: All providers in parallel
parallel:
  - provider: anthropic
  - provider: openai
  - provider: deepseek
max_concurrent: 3  # All at once
```

### Progressive Validation

```yaml
# Start with one, add more if needed
- provider: anthropic
  output: first

- condition: "{{first.confidence}} < 0.9"
  provider: openai
  output: second

- condition: "{{first}} != {{second}}"
  provider: deepseek  # Tie-breaker
```

---

## Complete Example

```yaml
name: comprehensive_validation
version: 1.0.0

steps:
  # 1. Parallel analysis
  - parallel:
      - name: claude
        provider: anthropic
        prompt: "{{task}}"
      - name: gpt
        provider: openai
        prompt: "{{task}}"
      - name: deepseek
        provider: deepseek
        prompt: "{{task}}"
    max_concurrent: 3
    aggregate: merge
    output: all_responses
  
  # 2. Compare
  - name: compare
    prompt: |
      Compare responses:
      {{all_responses}}
      
      Where do they agree/disagree?
    output: comparison
  
  # 3. Search for evidence
  - name: verify
    servers: [brave-search]
    prompt: "Find evidence for: {{all_responses}}"
    output: evidence
  
  # 4. Build consensus
  - name: consensus
    prompt: |
      Build consensus from:
      - Responses: {{all_responses}}
      - Comparison: {{comparison}}
      - Evidence: {{evidence}}
      
      Provide:
      - Validated answer
      - Confidence (high/medium/low)
      - Evidence summary
      - Caveats
```

---

## Quick Reference

```yaml
# Basic validation
provider1 → provider2 → compare → consensus

# Fact checking
extract_claims → verify_each → assess → report

# Decision validation
optimistic → pessimistic → neutral → synthesize

# Progressive
tier1 → if_uncertain → tier2 → if_critical → tier3
```

---

## Next Steps

- **[Conditional Routing](conditional-routing.md)** - Smart routing
- **[Research Pattern](research-agent.md)** - Deep research
- **[Examples](../examples/)** - Working examples

---

**Build confidence with validation!** ✓
