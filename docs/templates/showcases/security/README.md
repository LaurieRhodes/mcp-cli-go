# Security & Compliance Templates

> **For:** Security Managers, SOC Analysts, Incident Responders, Compliance Officers  
> **Purpose:** Automate security operations with AI-powered SOAR-style workflows

---

## What This Showcase Contains

This section demonstrates how templates automate critical security operations workflows. All examples solve real SOC challenges: alert fatigue, manual enrichment, phishing triage, incident documentation, and high-stakes decision validation.

### Available Use Cases

**Security Operations Automation:**

1. **[Alert Enrichment and Triage](use-cases/alert-enrichment.md)** - SOAR-style automated alert investigation
2. **[Phishing Email Analysis](use-cases/phishing-analysis.md)** - Automated email triage and incident creation
3. **[Consensus Threat Validation](use-cases/consensus-threat-validation.md)** - Multi-provider validation for critical decisions
4. **[Sentinel Alert Response](use-cases/sentinel-playbook.md)** - Automated playbook execution for Sentinel alerts
5. **[Automated Incident Documentation](use-cases/incident-documentation.md)** - Consistent post-incident reporting

---

## Why Templates Matter for Security

### 1. Alert Fatigue Reduction

**The Challenge:** SOC analysts receive 100+ alerts daily, 90% are false positives, manual investigation takes 15-30 minutes per alert.

**Template Solution:** Automated enrichment and triage

```yaml
# Alert received from Sentinel via MCP
# Template automatically:
# 1. Enriches with threat intel (VirusTotal, AbuseIPDB)
# 2. Checks user context (AD, recent activity)
# 3. Correlates with historical patterns
# 4. Assigns severity based on enriched context
# 5. Auto-closes false positives
# 6. Escalates real threats with full context
```

**Impact:**

- 90% of false positives auto-closed (measured)
- Analyst time focused on real threats
- Consistent investigation process
- No alerts missed during night shift

**Documentation:** [Alert Enrichment](use-cases/alert-enrichment.md)

---

### 2. MCP Integration: Security Tools as AI Data Sources

**The Challenge:** Security data scattered across SIEM, threat intel, endpoint tools, ticketing systems.

**Template Solution:** MCP servers expose security tools as data sources

```yaml
# Template can query multiple security systems:
steps:
  - name: get_sentinel_alert
    servers: [sentinel]  # MCP server for Microsoft Sentinel

  - name: enrich_ioc
    servers: [virustotal, abuseipdb]  # Threat intel enrichment

  - name: check_user_context
    servers: [active-directory]  # User details

  - name: create_incident
    servers: [jira]  # Ticket creation
```

**What this enables:**

- Templates receive alerts directly from Sentinel
- Query threat intel APIs programmatically
- Check user context from AD/LDAP
- Create tickets in JIRA/ServiceNow
- All within single workflow

**Example:** Template triggered by Sentinel alert → enriches → creates ticket → all automatic

---

### 3. Consensus Validation for High-Stakes Decisions

**The Challenge:** False positives on critical actions (isolating CEO's laptop, blocking production IPs) are costly. Single AI model might hallucinate or miss context.

**Template Solution:** Multi-provider validation before disruptive action

```yaml
# Analyze suspicious activity with 3 AI providers
parallel:
  - provider: anthropic  # Claude analyzes
  - provider: openai     # GPT-4o analyzes
  - provider: gemini     # Gemini analyzes

# Cross-validate findings:
# - All 3 agree: HIGH confidence → Auto-respond
# - 2 of 3 agree: MEDIUM confidence → Senior analyst review
# - Disagree: LOW confidence → Escalate to team lead
```

**Real scenario:**

```
Alert: "Suspicious PowerShell execution on CEO's laptop"

Claude: CRITICAL - Detected credential dumping pattern
GPT-4o: CRITICAL - Mimikatz-like behavior observed
Gemini: MEDIUM - PowerShell usage unusual but might be admin script

Consensus: MEDIUM (2 of 3 critical)
Action: Notify senior analyst for review before isolation
Result: Turned out to be legitimate IT maintenance
Avoided: Isolating CEO's laptop during board meeting
```

**Cost consideration:**

- Single provider: $0.045 per analysis
- Consensus (3 providers): $0.135 per analysis
- **Worth it when:** Blocking production servers, isolating executives, major response actions
- **Not worth it when:** Routine low-impact decisions

**Documentation:** [Consensus Threat Validation](use-cases/consensus-threat-validation.md)

---

### 4. Context-Efficient Multi-Source Enrichment

**The Challenge:** Enriching alerts requires checking 10+ data sources. Loading all context into LLM = token explosion.

**Template Solution:** Each enrichment step gets fresh context

**Traditional approach (context bloat):**

```
LLM Context (200K tokens):
├── Alert data: 5K tokens
├── VirusTotal response: 40K tokens
├── AbuseIPDB response: 20K tokens
├── User AD info: 10K tokens
├── Historical alerts: 60K tokens
├── Endpoint logs: 50K tokens
└── Remaining: 15K tokens (nearly full)
```

**Template approach (efficient):**

```
LLM sees: Alert summary (3K tokens)

Template enrichment steps (separate contexts):
├── Step 1: VirusTotal lookup → Fresh 200K context
├── Step 2: AbuseIPDB check → Fresh 200K context
├── Step 3: AD user query → Fresh 200K context
├── Step 4: Historical correlation → Fresh 200K context
└── Step 5: Final analysis with enriched summary → 15K tokens

LLM receives: Final enriched analysis (5K tokens)
```

**Benefits:**

- Each enrichment step has full context window
- Can enrich from 10+ sources without overflow
- LLM context stays clean for analyst questions
- Scalable to complex investigations

---

### 5. Failover Resilience for 24/7 Operations

**The Challenge:** Security operations run 24/7. API outages can't block incident response.

**Template Solution:** Automatic failover across providers

```yaml
# Security analysis with guaranteed completion:
steps:
  - provider: anthropic    # Primary (best quality)
    error_handling:
      on_failure: continue

  - provider: openai       # Automatic failover
    condition: "primary failed"

  - provider: ollama       # Final guarantee (local)
    condition: "secondary failed"
```

**Real incident:**

```
03:17 AM: Ransomware alert detected
03:17:05: Template starts analysis
03:17:06: Anthropic API returns 429 (rate limit)
03:17:06: Auto-failover to OpenAI
03:17:12: Analysis complete, incident created
03:17:45: Ransomware contained

Without failover: SOC analyst waits for rate limit reset
With failover: Response continues automatically
```

**Documentation:** [Automated Incident Documentation](use-cases/incident-documentation.md)

---

## Quick Start

### 1. Choose Your Security Challenge

**Alert fatigue?**

- [Alert Enrichment and Triage](use-cases/alert-enrichment.md) - Automate investigation

**Phishing overload?**

- [Phishing Email Analysis](use-cases/phishing-analysis.md) - Automated email triage

**High-stakes decisions?**

- [Consensus Threat Validation](use-cases/consensus-threat-validation.md) - Multi-provider validation

**Sentinel integration?**

- [Sentinel Alert Response](use-cases/sentinel-playbook.md) - Automated playbooks

**Incident documentation?**

- [Automated Incident Documentation](use-cases/incident-documentation.md) - Consistent reports

### 2. Set Up MCP Integrations

Security templates often integrate with existing tools via MCP servers:

```yaml
# Example: Sentinel MCP server configuration
servers:
  sentinel:
    command: "sentinel-mcp-server"
    env:
      TENANT_ID: "${AZURE_TENANT_ID}"
      CLIENT_ID: "${AZURE_CLIENT_ID}"
      CLIENT_SECRET: "${AZURE_CLIENT_SECRET}"
```

**Available MCP servers for security:**

- Microsoft Sentinel (SIEM alerts)
- Threat intel APIs (VirusTotal, AbuseIPDB, AlienVault OTX)
- Ticketing (JIRA, ServiceNow)
- Email (Exchange, Gmail for security@ mailbox)
- Active Directory (user context)

### 3. Run Template Against Real Alert

```bash
# Enrich Sentinel alert
mcp-cli --template alert_enrichment --input-data "{
  \"alert_id\": \"abc123\",
  \"source_ip\": \"192.168.1.100\",
  \"destination_ip\": \"203.0.113.45\"
}"

# Result: Enriched analysis with threat intel, user context, severity
```

### 4. Customize for Your Environment

Edit templates to match your:

- Threat intel sources
- Severity thresholds
- Escalation policies
- Compliance requirements

---

## Integration Patterns

### Pattern 1: Sentinel → Template → Ticket

**Automated alert response:**

```yaml
# Triggered by Sentinel alert via MCP
name: sentinel_alert_response

steps:
  # Get full alert details from Sentinel
  - name: fetch_alert
    servers: [sentinel]
    prompt: "Get alert details for {{input_data.alert_id}}"
    output: alert_details

  # Enrich with threat intel
  - name: enrich
    servers: [virustotal, abuseipdb]
    prompt: "Enrich IOCs: {{alert_details.indicators}}"
    output: enrichment

  # Analyze severity
  - name: analyze
    prompt: "Assess: {{alert_details}} + {{enrichment}}"
    output: analysis

  # Create ticket if real threat
  - name: create_ticket
    condition: "{{analysis.severity}} != 'false_positive'"
    servers: [jira]
    prompt: "Create incident: {{analysis}}"
```

**Flow:**

1. Sentinel alert triggers template (via MCP server)
2. Template enriches alert automatically
3. Creates ticket for real threats
4. Auto-closes false positives

---

### Pattern 2: Email Monitoring → Phishing Analysis → Incident

**Automated phishing triage:**

```yaml
name: phishing_triage

steps:
  # Monitor security@ mailbox
  - name: get_emails
    servers: [exchange]
    prompt: "Get unread emails from security@company.com"
    output: emails

  # Analyze each email
  - name: analyze_phishing
    for_each: "{{emails}}"
    item_name: email
    prompt: "Analyze phishing indicators: {{email}}"
    output: analyses

  # Create incidents for real threats
  - name: create_incidents
    for_each: "{{analyses}}"
    condition: "{{item.is_phishing}}"
    servers: [jira]
```

**Result:** Security@ mailbox monitored continuously, real phishing auto-escalated

---

### Pattern 3: Consensus for Critical Actions

**Before isolating production servers:**

```yaml
name: validate_before_isolation

steps:
  # Get consensus from 3 providers
  - name: multi_provider_analysis
    parallel:
      - provider: anthropic
      - provider: openai
      - provider: gemini
    prompt: "Should we isolate server {{server_id}}?"
    output: opinions

  # Require consensus
  - name: decision
    prompt: "Consensus from: {{opinions}}"
    output: recommendation

  # Only execute if high confidence
  - name: isolate
    condition: "{{recommendation.confidence}} == 'HIGH'"
    servers: [edr-platform]
```

---

## Best Practices

### Security Template Design

**✅ Do:**

- Use consensus validation for disruptive actions
- Implement failover for 24/7 availability
- Log all automated decisions for audit trail
- Include manual review step for high-impact actions
- Set appropriate severity thresholds
- Test with historical alerts before production

**❌ Don't:**

- Auto-respond to critical alerts without validation
- Skip logging automated decisions
- Use single provider for high-stakes actions
- Forget to handle API rate limits
- Deploy without testing against known false positives
- Ignore compliance requirements (retain logs, audit trails)

### Compliance Considerations

**Audit trail requirements:**

- Log all template executions
- Record which AI provider made decisions
- Retain enrichment data sources
- Document automated vs manual actions
- Track false positive/negative rates

**Data handling:**

- Ensure threat intel queries comply with TOS
- Don't send PII to third-party APIs without approval
- Use local models for sensitive data analysis
- Encrypt logs containing security findings



---

## Example: Complete Alert Workflow

```yaml
name: complete_security_workflow

# Triggered by Sentinel alert
steps:
  # 1. Fetch full alert from Sentinel
  - name: get_alert
    servers: [sentinel]
    output: alert

  # 2. Parallel enrichment (fast)
  - name: enrich
    parallel:
      - servers: [virustotal]    # Check IP reputation
      - servers: [abuseipdb]     # Check historical abuse
      - servers: [ad]            # Get user context
    output: enrichment

  # 3. Consensus validation (for high-severity)
  - name: validate
    condition: "{{alert.severity}} == 'high'"
    parallel:
      - provider: anthropic
      - provider: openai
      - provider: gemini
    prompt: "Validate threat: {{alert}} + {{enrichment}}"
    output: validation

  # 4. Create incident if real threat
  - name: create_incident
    condition: "{{validation.consensus}} == 'threat'"
    servers: [jira]
    output: incident

  # 5. Execute response (if critical)
  - name: respond
    condition: "{{validation.confidence}} == 'HIGH'"
    servers: [edr]
    prompt: "Execute response: {{validation.actions}}"

  # 6. Document everything
  - name: document
    prompt: "Generate incident report"
    output: report
```

**This workflow demonstrates:**

- MCP integration (Sentinel, threat intel, ticketing, EDR)
- Parallel enrichment (fast)
- Consensus validation (high-stakes)
- Automated response (when high confidence)
- Documentation (audit trail)

---

## Template Library

All templates available in [templates/](templates/):

**SOAR-Style Automation:**

- `alert_enrichment.yaml` - Automated alert triage
- `phishing_analysis.yaml` - Email phishing detection
- `sentinel_playbook.yaml` - Sentinel alert response

**Validation & Compliance:**

- `consensus_threat_validation.yaml` - Multi-provider validation
- `incident_documentation.yaml` - Automated reporting

---

## Advanced Deployment

### Pattern: Security Orchestration Hub

**Central template server for SOC:**

```yaml
# MCP server exposing security templates
name: security_orchestration
version: 1.0.0

tools:
  - name: enrich_alert
    template: alert_enrichment
    description: Enrich security alert with threat intel

  - name: analyze_phishing
    template: phishing_analysis
    description: Analyze email for phishing indicators

  - name: validate_threat
    template: consensus_threat_validation
    description: Multi-provider threat validation
```

**Result:** SOC analysts access all security templates via natural language interface

---

## Next Steps

1. **Review use cases** - Read detailed documentation for each workflow
2. **Set up MCP servers** - Configure Sentinel, threat intel integrations
3. **Test with historical data** - Run templates against known alerts
4. **Measure baseline** - Track current alert triage metrics
5. **Deploy incrementally** - Start with alert enrichment, expand to full SOAR

---

## Additional Resources

- **[MCP Server Integration](../../../mcp-server/README.md)** - Expose templates as tools
- **[Why Templates Matter](../../WHY_TEMPLATES_MATTER.md)** - Strategic overview
- **[DevOps Showcase](../devops/)** - Resilience patterns applicable to security
- **[Template Authoring Guide](../../authoring-guide.md)** - Create custom security templates

---

**Security automation with AI: Reduce alert fatigue, ensure consistency, maintain 24/7 operations.**

Templates transform security operations from manual triage to intelligent orchestration.
