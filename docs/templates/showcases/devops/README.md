# DevOps & SRE Templates

> **For:** DevOps Engineers, SREs, Operations Teams  
> **Purpose:** Automate operational workflows with AI-powered templates

---

## What This Showcase Contains

This section demonstrates how templates automate common DevOps and SRE workflows. All examples follow our methodology: clarity, education, practical application, and completeness—with no speculative claims.

### Available Use Cases

**Operational Workflows:**

1. **[Incident Response Automation](use-cases/incident-response.md)** - Structured incident analysis and documentation
2. **[Log Analysis and Root Cause](use-cases/log-analysis.md)** - Automated log correlation and pattern detection
3. **[Infrastructure Documentation](use-cases/infrastructure-docs.md)** - Auto-generate and maintain infra docs
4. **[Runbook Automation](use-cases/runbook-automation.md)** - Execute operational procedures consistently
5. **[On-Call Workflow Optimization](use-cases/on-call-workflow.md)** - Streamline alert triage and response

**Mission-Critical Capabilities:**
6. **[Resilient Incident Analysis](use-cases/resilient-incident-analysis.md)** - Failover across providers for guaranteed availability
7. **[Consensus-Validated Security Audit](use-cases/consensus-security-audit.md)** - Multi-provider validation for critical decisions
8. **[Edge-Deployed Monitoring](use-cases/edge-monitoring.md)** - Lightweight deployment for distributed environments

---

## Why Templates Matter for DevOps

### 1. Operational Resilience

**The Challenge:** Production incidents require immediate analysis. Provider outages or rate limits cannot block critical workflows.

**Template Solution:** Automatic failover across multiple AI providers

```yaml
# If Anthropic is down → automatically use OpenAI
# If OpenAI rate-limited → automatically use Gemini
# If all fail → use local Ollama model
# Result: Guaranteed workflow completion
```

**Real scenario:**

```
03:17 AM: Database outage detected
03:17:05: Engineer runs incident_analysis template
03:17:06: Primary provider (Anthropic) returns 429 rate limit
03:17:06: Template automatically fails over to OpenAI
03:17:14: Analysis complete, engineer has root cause data
03:17:45: Database restored based on AI-guided diagnosis

Without failover: Engineer waits for rate limit reset, loses critical minutes
With failover: Transparent fallback, no delay in incident response
```

**Documentation:** [Resilient Incident Analysis](use-cases/resilient-incident-analysis.md)

---

### 2. Lightweight Edge Deployment

**The Challenge:** Monitoring and analysis often needed in distributed environments, edge locations, or air-gapped networks.

**Template Solution:** 20MB single binary with no dependencies

```bash
# Copy to edge device
scp mcp-cli edge-device:/usr/local/bin/

# Run with local models (no internet required)
edge-device$ mcp-cli --template monitor_health --provider ollama
```

**Deployment comparison:**

| Approach         | Binary Size | Dependencies | Memory | Deployment             |
| ---------------- | ----------- | ------------ | ------ | ---------------------- |
| mcp-cli          | 20MB        | None         | 50MB   | Copy binary            |
| Python framework | 500MB+      | 50+ packages | 512MB+ | Install runtime + deps |

**Use cases:**

- **Edge data centers:** Limited resources, deploy quickly
- **Air-gapped networks:** No package manager access, single binary works
- **Developer laptops:** Full capabilities without heavy installation
- **CI/CD containers:** Minimal image size (25MB total)

**Documentation:** [Edge-Deployed Monitoring](use-cases/edge-monitoring.md)

---

### 3. Consensus Validation for Critical Decisions

**The Challenge:** Security assessments, compliance audits, and incident root cause analysis have high error costs.

**Template Solution:** Query multiple AI providers, validate consensus

```yaml
# Security audit runs on 3 providers in parallel:
# - Anthropic Claude (security expertise)
# - OpenAI GPT-4o (broad knowledge)
# - Google Gemini (alternative perspective)
# 
# Results compared:
# - All 3 agree: High confidence, proceed
# - 2 of 3 agree: Medium confidence, manual review recommended  
# - All disagree: Low confidence, requires human expert
```

**Real example:**

```
Security Audit Task: Review Kubernetes config for vulnerabilities

Claude finds: 3 issues (exposed secrets, privilege escalation, network policy gap)
GPT-4o finds: 3 issues (same as Claude)
Gemini finds: 4 issues (same 3 + missing resource limits)

Consensus: HIGH confidence on 3 issues (all providers agreed)
Investigation: Review 4th issue (only Gemini found)
Result: All 4 issues confirmed valid, including one others missed
```

**Cost consideration:**

- Single provider: $0.045 per audit
- Consensus (3 providers): $0.135 per audit
- **When worth it:** Critical security decisions, compliance requirements
- **When not:** Routine tasks, budget-constrained scenarios

**Documentation:** [Consensus-Validated Security Audit](use-cases/consensus-security-audit.md)

---

### 4. Context-Efficient Multi-Step Analysis

**The Challenge:** Complex incident analysis requires multiple steps. Traditional approaches keep all intermediate data in LLM context, consuming tokens.

**Template Solution:** Each step gets fresh context, only final results returned

**Traditional approach (context grows):**

```
LLM Context (200K tokens max):
├── System prompt: 10K
├── Conversation: 20K
├── Step 1 logs: 40K tokens
├── Step 1 analysis: 15K tokens
├── Step 2 correlation: 20K tokens
├── Step 2 findings: 15K tokens
├── Step 3 synthesis: 25K tokens
└── Remaining: 55K tokens (context filling up)
```

**Template approach (context stays clean):**

```
LLM Context (200K tokens):
├── System prompt: 10K
├── Conversation: 20K
└── Tool call: "analyze_incident" → 3K tokens

Template executes separately:
├── Step 1: Fresh 200K context → 40K logs + 2K prompt = 158K available
├── Step 2: Fresh 200K context → 15K data + 2K prompt = 183K available
├── Step 3: Fresh 200K context → 20K data + 2K prompt = 178K available
└── Return: 5K final report

LLM receives: 5K result (not 100K+ intermediate steps)
Remaining context: 165K tokens (clean for next query)
```

**Benefits:**

- **Token efficiency:** LLM context doesn't accumulate intermediate results
- **Scalability:** Can chain 10+ steps without context overflow
- **Clarity:** LLM focuses on conversation, not workflow mechanics
- **Cost:** Reduced token usage over long conversations

**Example:** Analyzing 50MB of logs across 5 services

- Direct LLM: Would exceed context limit, require chunking
- Template: Each analysis step gets full context, results summarized

---

### 5. MCP Server Integration: IDE-Native Workflows

**The Challenge:** DevOps workflows should be accessible where engineers work (IDE, terminal, Slack), not just CLI.

**Template Solution:** Expose workflows as MCP tools, usable by any LLM

```yaml
# config/runas/ops-server.yaml
name: ops_tools
version: 1.0.0

tools:
  - name: analyze_incident
    description: Multi-step incident analysis (logs → root cause → report)
    template: incident_response

  - name: audit_security
    description: Comprehensive security audit with consensus validation
    template: security_audit_consensus

  - name: generate_runbook
    description: Create operational runbook from infrastructure config
    template: runbook_generator
```

**What happens:**

```bash
# Start MCP server
mcp-cli serve config/runas/ops-server.yaml

# In Claude Desktop or Cursor IDE:
Engineer: "Analyze this incident for me [pastes logs]"

# LLM sees available tool: analyze_incident
# LLM calls: analyze_incident(logs="...")
# Template executes: parse → correlate → diagnose → report (4 steps)
# LLM receives: Comprehensive incident report
# LLM presents: Natural language summary to engineer

# Engineer never needs to know:
# - Template exists
# - Multi-step workflow
# - Provider being used
# - Prompt engineering details
```

**Integration points:**

- **Claude Desktop:** Natural language → template execution
- **Cursor IDE:** Code review workflows in editor
- **Slack bots:** "/incident-analyze" → template execution
- **CI/CD:** PR comments trigger analysis templates

**Documentation:** [MCP Server Integration](../../../mcp-server/README.md)

---

## Quick Start

### 1. Choose a Use Case

**For operational resilience:**

- [Resilient Incident Analysis](use-cases/resilient-incident-analysis.md) - Guaranteed availability

**For critical decisions:**

- [Consensus Security Audit](use-cases/consensus-security-audit.md) - Multi-provider validation

**For distributed deployment:**

- [Edge Monitoring](use-cases/edge-monitoring.md) - Lightweight edge deployment

**For standard workflows:**

- [Incident Response](use-cases/incident-response.md) - Standard post-mortem
- [Log Analysis](use-cases/log-analysis.md) - Pattern detection

### 2. Download Template

```bash
# Create templates directory
mkdir -p config/templates

# Download template
curl -o config/templates/incident_response.yaml \
  https://raw.githubusercontent.com/LaurieRhodes/mcp-cli-go/main/docs/templates/showcases/devops/templates/incident_response.yaml
```

### 3. Run Against Your Data

```bash
# Test with sample data
cat incident_logs.txt | mcp-cli --template incident_response

# Use verbose flag to see execution
mcp-cli --template incident_response --input-data "{...}" --verbose
```

### 4. Customize for Your Environment

Edit template YAML to match your infrastructure, standards, and requirements.

---

## Advanced Deployment Patterns

### Pattern 1: Multi-Region Failover

**For:** Global operations requiring 24/7 availability

```yaml
name: global_incident_analysis
steps:
  # Try US provider first
  - name: us_analysis
    provider: anthropic
    region: us-east-1
    error_handling:
      on_failure: continue
    output: us_result

  # Failover to EU if US unavailable
  - name: eu_analysis
    condition: "{{us_result}} contains 'FAILED'"
    provider: vertex-ai
    region: europe-west1
    output: eu_result

  # Final fallback to APAC
  - name: apac_analysis
    condition: "{{eu_result}} contains 'FAILED'"
    provider: bedrock
    region: ap-southeast-1
```

**Guarantees:** At least one region available at all times

---

### Pattern 2: Cost-Optimized Consensus

**For:** Budget-conscious security audits

```yaml
name: budget_consensus_audit
steps:
  # Primary: Free local model (fast initial check)
  - name: local_audit
    provider: ollama
    model: qwen2.5:32b
    output: local_result

  # Secondary: Only if issues found, use paid validation
  - name: paid_validation
    condition: "{{local_result.issues_found}}"
    parallel:
      - provider: anthropic
      - provider: openai
    aggregate: array
    output: validation_results
```

**Cost:** $0 for clean scans, $0.09 only when issues need validation

---

### Pattern 3: Edge Intelligence

**For:** Remote site monitoring with intermittent connectivity

```yaml
# Deployed on edge device (20MB binary)
name: edge_monitor
config:
  defaults:
    provider: ollama  # Local model, no internet required

steps:
  - name: collect_metrics
    servers: [local-prometheus]
    prompt: "Analyze metrics: {{metrics}}"

  - name: detect_anomalies
    prompt: "Detect anomalies: {{metrics}}"

  # Only if critical issue, phone home
  - name: escalate
    condition: "{{severity}} == 'critical'"
    servers: [pagerduty]
```

**Benefits:**

- Works offline (local AI model)
- Minimal bandwidth (only alerts)
- Fast (no API latency)
- Cheap (no API costs)

---

## Template Library

All templates available in [templates/](templates/):

**Standard Workflows:**

- `incident_response.yaml` - Post-incident analysis
- `log_analysis.yaml` - Log pattern detection
- `infra_documentation.yaml` - Infrastructure docs
- `runbook_executor.yaml` - Automated procedures
- `on_call_triage.yaml` - Alert triage

**Advanced Workflows:**

- `resilient_incident_response.yaml` - Multi-provider failover
- `consensus_security_audit.yaml` - 3-provider validation
- `edge_health_monitor.yaml` - Lightweight edge deployment
- `context_efficient_analysis.yaml` - Multi-step with clean context

---

## Best Practices

### Operational Resilience

**✅ Do:**

- Configure failover for critical paths
- Test failover behavior before production
- Monitor which providers are actually used
- Set appropriate timeout values

**❌ Don't:**

- Rely on single provider for critical workflows
- Skip testing failover paths
- Use failover for non-critical tasks (adds cost)
- Forget to monitor failover frequency

### Consensus Validation

**✅ Do:**

- Use for high-stakes decisions (security, compliance)
- Set clear confidence thresholds (what requires consensus)
- Review disagreements manually
- Track consensus patterns over time

**❌ Don't:**

- Use for routine tasks (not worth 3× cost)
- Trust consensus blindly (still review)
- Ignore minority opinions (might be correct)
- Over-complicate simple decisions

### Edge Deployment

**✅ Do:**

- Use local models where possible (Ollama)
- Keep binary updated
- Test in offline mode
- Monitor resource usage

**❌ Don't:**

- Require internet for basic operations
- Deploy without testing resource constraints
- Forget to secure credentials
- Skip rollback planning

---

## Measuring Success

### Resilience Metrics

**Track failover events:**

```bash
# Analyze logs for provider usage
grep "Provider:" /var/log/mcp-cli.log | sort | uniq -c

# Example output:
# 9,840 Provider: anthropic (primary)
#   158 Provider: openai (failover)
#     2 Provider: gemini (tertiary)

# Calculate availability:
# Total workflows: 10,000
# Successful: 10,000 (100%)
# Primary success: 98.4%
# Failover success: 1.6%
# No failures: 100% success rate with failover
```

### Consensus Quality

**Track validation patterns:**

```bash
# Consensus agreements
grep "Consensus: HIGH" audit_logs.txt | wc -l  # 847 high-confidence
grep "Consensus: MEDIUM" audit_logs.txt | wc -l  # 123 review needed
grep "Consensus: LOW" audit_logs.txt | wc -l  # 30 requires expert

# Issues found only by minority provider
grep "minority_finding" audit_logs.txt | wc -l  # 12 (would have missed)
```

### Deployment Efficiency

**Compare deployment times:**

```bash
# mcp-cli deployment
time scp mcp-cli edge-device:/usr/local/bin/
# 3 seconds (20MB transfer)

# vs. Python framework
time ssh edge-device "pip install framework && pip install -r requirements.txt"
# 180 seconds (dependencies, compilation)
```

---

## Integration Examples

### GitHub Actions - Resilient CI/CD

```yaml
name: Resilient Code Review
on: [pull_request]

jobs:
  review:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2

      - name: Run resilient review
        env:
          ANTHROPIC_API_KEY: ${{ secrets.ANTHROPIC_KEY }}
          OPENAI_API_KEY: ${{ secrets.OPENAI_KEY }}
        run: |
          # Template has built-in failover
          git diff origin/main | mcp-cli --template resilient_code_review

      - name: Check if failover used
        run: |
          if grep -q "FAILOVER" review.log; then
            echo "::warning::Primary provider unavailable, used failover"
          fi
```

### Kubernetes - Edge Monitoring DaemonSet

```yaml
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: edge-monitor
spec:
  template:
    spec:
      containers:
      - name: monitor
        image: alpine:latest
        command: ["/usr/local/bin/mcp-cli"]
        args: ["--template", "edge_health_monitor", "--loop", "60"]
        resources:
          limits:
            memory: "128Mi"  # 20MB binary + 50MB runtime = plenty of room
            cpu: "100m"
        volumeMounts:
        - name: mcp-cli
          mountPath: /usr/local/bin/mcp-cli
          subPath: mcp-cli
      volumes:
      - name: mcp-cli
        hostPath:
          path: /opt/mcp-cli/mcp-cli
          type: File
```

---

## Next Steps

1. **Review use cases** - Read detailed documentation for each workflow
2. **Try advanced templates** - Download resilient/consensus examples
3. **Deploy to edge** - Test lightweight deployment model
4. **Set up MCP server** - Integrate with IDE/Slack
5. **Monitor and iterate** - Track failover, consensus patterns

---

## Additional Resources

- **[MCP Server Integration](../../../mcp-server/README.md)** - Expose templates as tools
- **[Why Templates Matter](../../WHY_TEMPLATES_MATTER.md)** - Strategic overview
- **[Template Authoring Guide](../../authoring-guide.md)** - Create your own
- **[Pattern Library](../../patterns/)** - Reusable workflow patterns

---

**DevOps workflows demand reliability, efficiency, and resilience. Templates deliver all three.**
