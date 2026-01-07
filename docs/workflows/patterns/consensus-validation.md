# Consensus Validation Pattern

**Use multiple AI providers for critical decisions requiring high confidence.**

---

## Overview

The Consensus Validation pattern uses workflow v2.0's built-in `consensus` mode to get agreement from multiple AI providers on critical decisions.

**Verified feature:** This pattern uses the `consensus` field in StepV2, which executes the same prompt across multiple providers in parallel and evaluates agreement.

---

## When to Use

**Use this pattern when:**
- High-stakes decisions (medical, financial, legal)
- Fact-checking is critical
- Need confidence through multi-provider agreement
- Single-model bias is a concern

**Don't use when:**
- Simple, routine questions
- Creative tasks (multiple valid answers)
- Cost is primary concern (3x more expensive)
- Speed is critical (slower than single provider)

---

## Basic Structure

```yaml
$schema: "workflow/v2.0"
name: consensus_validation
version: 1.0.0
description: Validate critical information with consensus

execution:
  provider: anthropic
  model: claude-sonnet-4

steps:
  - name: validate
    consensus:
      prompt: "{{input}}"
      executions:
        - provider: anthropic
          model: claude-sonnet-4
        - provider: openai
          model: gpt-4o
        - provider: deepseek
          model: deepseek-chat
      require: unanimous
```

**How consensus works:**
1. Same prompt sent to all providers
2. Responses collected
3. Agreement evaluated
4. Result includes votes and confidence level

---

## Complete Examples

### Example 1: Fact Verification

```yaml
$schema: "workflow/v2.0"
name: fact_check
version: 1.0.0
description: Verify factual claims with consensus

execution:
  provider: anthropic
  model: claude-sonnet-4
  servers: [brave-search]

steps:
  # Extract claims to verify
  - name: extract_claims
    run: |
      Extract factual claims from this text:
      {{input}}
      
      List each claim that can be verified.

  # Verify each claim with consensus
  - name: verify_claim
    needs: [extract_claims]
    consensus:
      prompt: |
        Is this claim accurate? Answer YES or NO and explain why.
        
        Claim: {{extract_claims}}
        
        Search for evidence if needed.
      executions:
        - provider: anthropic
          model: claude-sonnet-4
          temperature: 0
        - provider: openai
          model: gpt-4o
          temperature: 0
        - provider: deepseek
          model: deepseek-chat
          temperature: 0
      require: unanimous
      timeout: 60s

  # Generate report
  - name: report
    needs: [verify_claim]
    run: |
      Create fact-check report:
      
      Claims: {{extract_claims}}
      Verification: {{verify_claim}}
      
      For each claim provide:
      - Claim text
      - Verification status
      - Confidence level
      - Evidence summary
```

**Usage:**
```bash
./mcp-cli --workflow fact_check \
  --input-data "The Earth is flat and NASA faked the moon landing."
```

---

### Example 2: Security Approval

```yaml
$schema: "workflow/v2.0"
name: security_approval
version: 1.0.0
description: Security review with consensus approval

execution:
  provider: anthropic
  model: claude-sonnet-4
  servers: [filesystem]

steps:
  # Analyze code for security issues
  - name: security_scan
    run: |
      Analyze this code for security vulnerabilities:
      {{input}}
      
      List any security issues found.

  # Consensus decision on safety
  - name: approve_deployment
    needs: [security_scan]
    consensus:
      prompt: |
        Based on this security analysis, approve deployment?
        
        Analysis: {{security_scan}}
        
        Answer:
        - APPROVE: No critical security issues
        - REJECT: Critical security issues found
        
        Explain your decision.
      executions:
        - provider: anthropic
          model: claude-sonnet-4
          temperature: 0
        - provider: openai
          model: gpt-4o
          temperature: 0
        - provider: deepseek
          model: deepseek-chat
          temperature: 0
      require: unanimous
      allow_partial: false

  # Generate approval report
  - name: approval_report
    needs: [approve_deployment]
    run: |
      Generate deployment approval report:
      
      Security Analysis: {{security_scan}}
      Approval Decision: {{approve_deployment}}
      
      Include:
      - Decision (approved/rejected)
      - Consensus level
      - All provider votes
      - Reasoning summary
```

---

### Example 3: Medical Information Validation

```yaml
$schema: "workflow/v2.0"
name: medical_validation
version: 1.0.0
description: Validate medical information with consensus

execution:
  provider: anthropic
  model: claude-sonnet-4
  servers: [brave-search]
  temperature: 0

steps:
  # Get medical information
  - name: medical_info
    run: |
      Provide evidence-based medical information about: {{input}}
      
      Include:
      - Symptoms
      - Causes
      - Treatments
      - When to see a doctor
      
      Cite medical sources.

  # Validate with consensus
  - name: validate_info
    needs: [medical_info]
    consensus:
      prompt: |
        Is this medical information accurate and evidence-based?
        
        Information: {{medical_info}}
        
        Answer YES or NO and explain:
        - Is it medically accurate?
        - Are sources reliable?
        - Are there any errors or omissions?
      executions:
        - provider: anthropic
          model: claude-sonnet-4
        - provider: openai
          model: gpt-4o
        - provider: deepseek
          model: deepseek-chat
      require: unanimous
      timeout: 90s

  # Add medical disclaimer
  - name: final_output
    needs: [validate_info]
    run: |
      Present validated medical information:
      
      Information: {{medical_info}}
      Validation: {{validate_info}}
      
      Add appropriate disclaimers:
      - This is not medical advice
      - Consult healthcare professional
      - Emergency warning signs
```

---

## Consensus Requirements

The `require` field specifies agreement level needed:

```yaml
# All providers must agree
require: unanimous

# At least 2 of 3 must agree
require: 2/3

# More than half must agree  
require: majority
```

**Example with 2/3:**
```yaml
steps:
  - name: decision
    consensus:
      prompt: "Is this safe?"
      executions:
        - provider: anthropic
          model: claude-sonnet-4
        - provider: openai
          model: gpt-4o
        - provider: deepseek
          model: deepseek-chat
      require: 2/3  # Need 2 of 3 to agree
```

---

## Consensus Result Format

Consensus returns structured data:

```json
{
  "success": true,
  "result": "APPROVED",
  "agreement": 1.0,
  "votes": {
    "anthropic/claude-sonnet-4": "APPROVED",
    "openai/gpt-4o": "APPROVED",
    "deepseek/deepseek-chat": "APPROVED"
  },
  "confidence": "high"
}
```

**Confidence levels:**
- `high`: All agree (unanimous)
- `good`: 2/3 or majority
- `medium`: Meets requirement but close
- `low`: Doesn't meet requirement

---

## Property Overrides in Consensus

Override execution properties for each provider:

```yaml
steps:
  - name: validate
    consensus:
      prompt: "Verify this fact"
      executions:
        # Conservative model
        - provider: anthropic
          model: claude-sonnet-4
          temperature: 0
          max_tokens: 1000
        
        # Standard model
        - provider: openai
          model: gpt-4o
          temperature: 0
          max_tokens: 1000
        
        # Fast model with shorter timeout
        - provider: deepseek
          model: deepseek-chat
          temperature: 0
          max_tokens: 1000
          timeout: 30s
      require: unanimous
```

---

## Pattern Variations

### Variation 1: Tiered Validation

Start with single provider, escalate to consensus if uncertain:

```yaml
steps:
  # First, try single provider
  - name: initial_check
    run: "{{input}}"

  # If uncertain, use consensus
  - name: consensus_check
    needs: [initial_check]
    consensus:
      prompt: "Validate: {{initial_check}}"
      executions:
        - provider: anthropic
          model: claude-sonnet-4
        - provider: openai
          model: gpt-4o
      require: unanimous
```

### Variation 2: Progressive Consensus

Use 2/3 first, unanimous for final approval:

```yaml
steps:
  # Initial consensus (2/3)
  - name: initial_approval
    consensus:
      prompt: "Initial review: {{input}}"
      executions:
        - provider: anthropic
          model: claude-sonnet-4
        - provider: openai
          model: gpt-4o
        - provider: deepseek
          model: deepseek-chat
      require: 2/3

  # Final consensus (unanimous)
  - name: final_approval
    needs: [initial_approval]
    consensus:
      prompt: "Final approval: {{initial_approval}}"
      executions:
        - provider: anthropic
          model: claude-opus-4
        - provider: openai
          model: gpt-4o
      require: unanimous
```

### Variation 3: Specialized Consensus

Different models for different aspects:

```yaml
steps:
  # Security consensus
  - name: security_check
    consensus:
      prompt: "Security review: {{input}}"
      executions:
        - provider: anthropic
          model: claude-sonnet-4
        - provider: openai
          model: gpt-4o
      require: unanimous

  # Performance consensus (different providers)
  - name: performance_check
    needs: [security_check]
    consensus:
      prompt: "Performance review: {{input}}"
      executions:
        - provider: deepseek
          model: deepseek-chat
        - provider: anthropic
          model: claude-sonnet-4
      require: unanimous
```

---

## Best Practices

### 1. Use Deterministic Temperature

```yaml
# ✅ Good: Temperature 0 for consistency
consensus:
  executions:
    - provider: anthropic
      temperature: 0
    - provider: openai
      temperature: 0

# ❌ Bad: High temperature creates randomness
consensus:
  executions:
    - provider: anthropic
      temperature: 0.9  # Too random for consensus
```

### 2. Set Appropriate Timeouts

```yaml
# ✅ Good: Longer timeout for consensus
consensus:
  timeout: 90s  # Overall consensus timeout
  executions:
    - provider: anthropic
      timeout: 30s  # Per-provider timeout

# ❌ Bad: Too short
consensus:
  timeout: 10s  # May not complete
```

### 3. Choose Right Requirement

```yaml
# Critical decisions: unanimous
consensus:
  prompt: "Approve nuclear reactor startup?"
  require: unanimous

# Important decisions: 2/3
consensus:
  prompt: "Approve code deployment?"
  require: 2/3

# Routine decisions: majority
consensus:
  prompt: "Accept pull request?"
  require: majority
```

### 4. Provide Context

```yaml
# ✅ Good: Clear context
consensus:
  prompt: |
    Review this code for security vulnerabilities.
    
    Code: {{code}}
    
    Answer APPROVED or REJECTED and explain.

# ❌ Bad: Vague prompt
consensus:
  prompt: "Is this okay?"
```

---

## Cost Considerations

**Consensus multiplies costs:**
- 3 providers = 3x cost
- Plus comparison overhead

**Example costs:**
- Single provider: ~$0.01
- Consensus (3): ~$0.03
- With comparison: ~$0.04

**When it's worth it:**
- ✅ Security approvals
- ✅ Medical validation
- ✅ Financial decisions
- ✅ Legal reviews
- ❌ Casual questions
- ❌ Creative writing

---

## Performance

**Execution:**
- Providers run in parallel
- Waits for all to complete
- Evaluates agreement

**Typical timing:**
- Single provider: 2-4s
- Consensus (3): 4-8s (parallel execution)
- With comparison: 6-10s

**Optimization:**
```yaml
# Faster: Use faster models
executions:
  - provider: deepseek  # Fast
  - provider: openai    # Medium
  - provider: anthropic # Quality

# Slower: All premium models
executions:
  - provider: anthropic
    model: claude-opus-4
  - provider: openai
    model: gpt-4o
```

---

## Troubleshooting

### Consensus Never Reached

**Problem:** Providers never agree

**Solutions:**
1. Use more specific prompts
2. Lower requirement (unanimous → 2/3)
3. Use temperature 0 for determinism
4. Check if question has objective answer

### Partial Failures

**Problem:** One provider times out or errors

**Solution:** Use `allow_partial`
```yaml
consensus:
  allow_partial: true  # Continue with 2/3 if one fails
  require: 2/3
```

### Costs Too High

**Problem:** Consensus too expensive

**Solutions:**
1. Use consensus only for critical decisions
2. Use cheaper models (deepseek)
3. Use 2/3 instead of unanimous
4. Use tiered validation

---

## Complete Example: Production Deployment

```yaml
$schema: "workflow/v2.0"
name: deployment_approval
version: 1.0.0
description: Multi-stage consensus for deployment

execution:
  provider: anthropic
  model: claude-sonnet-4
  servers: [filesystem]
  timeout: 120s

steps:
  # Stage 1: Security review
  - name: security_review
    run: |
      Review security of deployment:
      {{input}}
      
      Check:
      - Vulnerabilities
      - Secrets management
      - Access controls

  # Stage 2: Security consensus
  - name: security_consensus
    needs: [security_review]
    consensus:
      prompt: |
        Security Review: {{security_review}}
        
        Approve from security perspective?
        Answer APPROVED or REJECTED.
      executions:
        - provider: anthropic
          model: claude-sonnet-4
          temperature: 0
        - provider: openai
          model: gpt-4o
          temperature: 0
      require: unanimous
      timeout: 60s

  # Stage 3: Performance review
  - name: performance_review
    needs: [security_consensus]
    run: |
      Review performance impact:
      {{input}}
      
      Check:
      - Resource usage
      - Scalability
      - Bottlenecks

  # Stage 4: Final approval consensus
  - name: final_approval
    needs: [performance_review]
    consensus:
      prompt: |
        Overall approval for deployment?
        
        Security: {{security_consensus}}
        Performance: {{performance_review}}
        
        Answer APPROVED or REJECTED with reasoning.
      executions:
        - provider: anthropic
          model: claude-sonnet-4
          temperature: 0
        - provider: openai
          model: gpt-4o
          temperature: 0
        - provider: deepseek
          model: deepseek-chat
          temperature: 0
      require: unanimous
      timeout: 60s

  # Generate approval document
  - name: approval_document
    needs: [final_approval]
    run: |
      Generate deployment approval document:
      
      Security Review: {{security_review}}
      Security Consensus: {{security_consensus}}
      Performance Review: {{performance_review}}
      Final Approval: {{final_approval}}
      
      Include:
      - Approval status
      - All consensus votes
      - Risk summary
      - Approval timestamp
```

---

## Related Patterns

- **[Iterative Refinement](iterative-refinement.md)** - Improve with loops, validate with consensus
- **[Document Pipeline](document-pipeline.md)** - Process documents, validate with consensus

---

## See Also

- [Schema Reference](../SCHEMA.md) - Consensus mode documentation
- [Examples](../examples/) - More consensus examples

---

**Build confidence through consensus!** ✓
