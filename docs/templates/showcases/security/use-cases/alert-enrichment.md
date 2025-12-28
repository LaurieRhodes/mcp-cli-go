# Alert Enrichment and Triage

> **Template:** [alert_enrichment.yaml](../templates/alert_enrichment.yaml)  
> **Workflow:** Alert → Enrich → Analyze → Triage → Create Incident  
> **Best For:** SOAR-style automated alert investigation and triage

---

## Problem Description

### The Alert Fatigue Challenge

**SOC analysts drown in alerts:**
- 100-200 alerts per day per analyst
- 90% are false positives
- 15-30 minutes manual investigation per alert
- Critical alerts buried in noise
- Analyst burnout and turnover

**Manual enrichment process:**
```
1. Analyst receives alert from SIEM
2. Copy suspicious IP address
3. Open VirusTotal, paste IP, wait
4. Open AbuseIPDB, paste IP, wait
5. Check internal AD for user context
6. Search historical alerts for patterns
7. Check endpoint logs
8. Correlate across 10+ data sources
9. Make triage decision
10. Document findings
11. Create ticket if real threat

Time: 15-30 minutes per alert
Error-prone: Easy to miss correlation
Inconsistent: Varies by analyst skill
```

**Consequences:**
- Real threats missed in alert storm
- Analysts waste time on false positives
- Inconsistent investigation quality
- No coverage during night shift
- Alert backlog grows

---

## Template Solution

### What It Does

This template implements **SOAR-style automated alert enrichment and triage**:

1. **Receives alert** from Sentinel (or other SIEM) via MCP
2. **Extracts IOCs** (IPs, domains, hashes, URLs)
3. **Enriches in parallel** (threat intel, user context, historical patterns)
4. **Analyzes** enriched context for threat indicators
5. **Assigns severity** based on comprehensive analysis
6. **Auto-closes** obvious false positives
7. **Creates incident** for real threats with full context
8. **Escalates** critical threats immediately

### Template Structure

```yaml
name: alert_enrichment
description: SOAR-style automated alert enrichment and intelligent triage
version: 1.0.0

config:
  defaults:
    provider: anthropic
    model: claude-3-5-sonnet
    temperature: 0.2  # Lower for security analysis

steps:
  # Step 1: Extract IOCs from alert
  - name: extract_iocs
    prompt: |
      Extract indicators of compromise from this security alert:
      
      {{input_data.alert}}
      
      Extract and structure:
      - IP addresses (source, destination, both internal and external)
      - Domain names
      - URLs
      - File hashes (MD5, SHA1, SHA256)
      - Email addresses
      - User accounts involved
      - Process names
      - Registry keys modified
      
      Return as structured JSON.
    output: iocs

  # Step 2: Parallel threat intel enrichment (fast)
  - name: threat_intel_enrichment
    parallel:
      # VirusTotal reputation check
      - name: virustotal_check
        servers: [virustotal]
        prompt: |
          Check VirusTotal reputation for:
          IPs: {{iocs.ip_addresses}}
          Domains: {{iocs.domains}}
          Hashes: {{iocs.file_hashes}}
        output: vt_results
      
      # AbuseIPDB historical abuse check
      - name: abuseipdb_check
        servers: [abuseipdb]
        prompt: |
          Check AbuseIPDB for historical abuse reports:
          IPs: {{iocs.ip_addresses}}
        output: abuseipdb_results
      
      # User context from Active Directory
      - name: user_context
        servers: [active-directory]
        prompt: |
          Get user details for:
          Users: {{iocs.user_accounts}}
          
          Include:
          - Department
          - Job title
          - Manager
          - Account status
          - Recent password changes
          - Group memberships
        output: user_details
    
    max_concurrent: 3
    aggregate: merge
    output: enrichment_results

  # Step 3: Check historical patterns
  - name: historical_correlation
    servers: [sentinel]  # Query Sentinel for historical data
    prompt: |
      Search last 30 days for similar alerts:
      
      Same source IP: {{iocs.source_ip}}
      Same user: {{iocs.user_accounts}}
      Same process: {{iocs.process_names}}
      
      Return:
      - Similar alert count
      - Previous severity levels
      - Previous dispositions
      - Any patterns identified
    output: historical_data

  # Step 4: Analyze threat level
  - name: analyze_threat
    prompt: |
      Comprehensive threat analysis:
      
      **Original Alert:**
      {{input_data.alert}}
      
      **Extracted IOCs:**
      {{iocs}}
      
      **Threat Intelligence:**
      {{enrichment_results}}
      
      **Historical Context:**
      {{historical_data}}
      
      Analyze for:
      
      **Threat Indicators:**
      - Known malicious IPs/domains (from VirusTotal, AbuseIPDB)
      - Suspicious user behavior patterns
      - Anomalous process execution
      - Lateral movement indicators
      - Data exfiltration signs
      - Privilege escalation attempts
      
      **Benign Indicators:**
      - Legitimate business process
      - Known safe IPs/domains
      - Expected user behavior
      - Authorized software
      - Normal administrative activity
      
      **Severity Assessment:**
      Assign severity (CRITICAL, HIGH, MEDIUM, LOW, FALSE_POSITIVE):
      
      - CRITICAL: Active threat, immediate response required
      - HIGH: Likely threat, investigation needed within 1 hour
      - MEDIUM: Suspicious, investigate within 4 hours
      - LOW: Minor anomaly, review when time permits
      - FALSE_POSITIVE: Benign activity, auto-close
      
      **Confidence Level:**
      How confident in this assessment? (HIGH, MEDIUM, LOW)
      
      **Recommended Actions:**
      What should SOC analyst do?
      
      **Key Evidence:**
      What specific findings support this assessment?
    output: threat_analysis

  # Step 5: Create incident for real threats
  - name: create_incident
    condition: "{{threat_analysis.severity}} != 'FALSE_POSITIVE'"
    servers: [jira]  # Or ServiceNow, etc.
    prompt: |
      Create security incident:
      
      **Title:** {{input_data.alert.title}} - {{threat_analysis.severity}}
      
      **Severity:** {{threat_analysis.severity}}
      
      **Description:**
      Original alert from Sentinel:
      {{input_data.alert}}
      
      **Automated Analysis:**
      {{threat_analysis.summary}}
      
      **Key Evidence:**
      {{threat_analysis.evidence}}
      
      **Indicators of Compromise:**
      {{iocs}}
      
      **Threat Intel Findings:**
      {{enrichment_results.summary}}
      
      **Historical Context:**
      {{historical_data.summary}}
      
      **Recommended Actions:**
      {{threat_analysis.recommended_actions}}
      
      **Confidence:** {{threat_analysis.confidence}}
      
      **Priority:** {{threat_analysis.priority}}
      
      **Assigned To:** {{escalation_assignment}}
    output: incident

  # Step 6: Escalate critical threats immediately
  - name: escalate_critical
    condition: "{{threat_analysis.severity}} == 'CRITICAL'"
    servers: [pagerduty]  # Or Slack, Teams, etc.
    prompt: |
      CRITICAL SECURITY ALERT
      
      Alert: {{input_data.alert.title}}
      Severity: CRITICAL
      Confidence: {{threat_analysis.confidence}}
      
      Threat Analysis:
      {{threat_analysis.summary}}
      
      Immediate Actions Required:
      {{threat_analysis.recommended_actions}}
      
      Incident Created: {{incident.ticket_id}}
      
      Full Details: {{incident.url}}
      
      @oncall-security immediate response required
    output: escalation

  # Step 7: Close false positives
  - name: close_false_positive
    condition: "{{threat_analysis.severity}} == 'FALSE_POSITIVE'"
    servers: [sentinel]
    prompt: |
      Close alert {{input_data.alert.id}} as false positive.
      
      Reason: {{threat_analysis.explanation}}
      
      Supporting Evidence: {{threat_analysis.evidence}}
      
      Add comment for audit trail.
    output: closure

  # Step 8: Generate summary report
  - name: summary_report
    prompt: |
      Generate triage summary:
      
      # Alert Triage Report
      
      **Alert ID:** {{input_data.alert.id}}
      **Timestamp:** {{execution.timestamp}}
      **Processing Time:** {{execution.duration}}
      
      ## Disposition
      
      **Severity:** {{threat_analysis.severity}}
      **Confidence:** {{threat_analysis.confidence}}
      
      {% if incident %}
      **Incident Created:** {{incident.ticket_id}}
      **Status:** Escalated for investigation
      {% endif %}
      
      {% if closure %}
      **Status:** Closed as false positive
      **Reason:** {{threat_analysis.explanation}}
      {% endif %}
      
      ## Analysis Summary
      
      {{threat_analysis.summary}}
      
      ## Key Findings
      
      **Threat Intelligence:**
      {{enrichment_results.key_findings}}
      
      **Historical Patterns:**
      {{historical_data.patterns}}
      
      **IOCs Identified:**
      {{iocs.summary}}
      
      ## Recommended Actions
      
      {{threat_analysis.recommended_actions}}
      
      ---
      
      **Automated Analysis:** Yes
      **Human Review Required:** {% if threat_analysis.confidence == 'LOW' %}Yes{% else %}No{% endif %}
```

---

## Usage Examples

### Example 1: True Positive - Credential Dumping

**Scenario:** Sentinel alert for suspicious PowerShell on domain controller

**Input:**
```json
{
  "alert": {
    "id": "ALT-2024-001234",
    "title": "Suspicious PowerShell Execution on Domain Controller",
    "source_ip": "10.0.50.100",
    "destination_ip": "10.0.1.5",
    "user": "admin-jsmith",
    "process": "powershell.exe",
    "command_line": "Invoke-Mimikatz -DumpCreds",
    "severity_raw": "Medium",
    "timestamp": "2024-12-28T03:15:42Z"
  }
}
```

**Execution:**

```bash
mcp-cli --template alert_enrichment --input-data @alert.json
```

**What Happens:**

```
[03:15:45] Starting alert_enrichment
[03:15:45] Step: extract_iocs
[03:15:46] ✓ Extracted:
  - Source IP: 10.0.50.100
  - Destination IP: 10.0.1.5 (Domain Controller)
  - User: admin-jsmith
  - Process: powershell.exe
  - Command: Invoke-Mimikatz (CREDENTIAL DUMPING TOOL)

[03:15:46] Step: threat_intel_enrichment (parallel)
[03:15:46] → VirusTotal: Checking IPs...
[03:15:46] → AbuseIPDB: Checking abuse history...
[03:15:46] → Active Directory: Fetching user context...
[03:15:49] ✓ Enrichment complete:
  - VirusTotal: IPs clean (internal network)
  - AbuseIPDB: No reports (internal IPs)
  - AD: admin-jsmith = John Smith, IT Admin, created 3 months ago

[03:15:49] Step: historical_correlation
[03:15:51] ✓ Historical data:
  - Same user: 0 previous Mimikatz alerts
  - Same IP: 5 normal admin activities last 30 days
  - Pattern: First time this tool detected

[03:15:51] Step: analyze_threat
[03:15:58] ✓ Analysis complete:
  - Severity: CRITICAL
  - Confidence: HIGH
  - Finding: Mimikatz credential dumping tool detected
  - Evidence: Command line contains "Invoke-Mimikatz -DumpCreds"
  - Context: No legitimate use for this tool
  - Risk: Potential credential theft, privilege escalation

[03:15:58] Step: create_incident
[03:16:02] ✓ Incident created: SEC-2024-4567
  - Priority: P1 (Critical)
  - Assigned to: Security Team Lead
  - SLA: 15 minutes response time

[03:16:02] Step: escalate_critical
[03:16:03] ✓ PagerDuty alert sent
  - Incident triggered for on-call security engineer
  - SMS + Push notification sent

[03:16:03] Step: summary_report
[03:16:05] ✓ Report generated

[03:16:05] ✓ Template completed (20 seconds total)
```

**Output:**

```markdown
# Alert Triage Report

**Alert ID:** ALT-2024-001234
**Severity:** CRITICAL
**Confidence:** HIGH
**Incident:** SEC-2024-4567

## Analysis Summary

CRITICAL credential dumping attempt detected. User admin-jsmith executed
Mimikatz on domain controller 10.0.1.5. This is a well-known credential
theft tool with NO legitimate administrative use.

## Key Findings

**Threat Intelligence:**
- Mimikatz: Known credential dumping tool (MITRE ATT&CK T1003)
- Command: -DumpCreds explicitly dumps credentials from memory
- Target: Domain Controller (highest value target)

**Historical Patterns:**
- First time this user has triggered this alert
- User account created 3 months ago
- Previous activities were normal admin tasks

**IOCs:**
- Source: 10.0.50.100 (admin workstation)
- User: admin-jsmith (IT Admin account)
- Tool: Invoke-Mimikatz
- Target: DC01 (10.0.1.5)

## Immediate Actions Required

1. **ISOLATE:** Quarantine source system 10.0.50.100
2. **DISABLE:** Disable account admin-jsmith immediately
3. **RESET:** Force password reset for all domain admin accounts
4. **INVESTIGATE:** Check for data exfiltration from DC
5. **REVIEW:** Audit all actions by admin-jsmith in last 7 days
6. **CONTAIN:** Check for lateral movement to other systems

## Incident Details

**Ticket:** SEC-2024-4567
**On-call:** Paged via PagerDuty
**Response SLA:** 15 minutes
```

**Cost:**
- Extract IOCs: $0.005
- Parallel enrichment (3 concurrent): $0.015
- Historical correlation: $0.008
- Threat analysis: $0.025
- Incident creation: $0.003
- Escalation: $0.002
- **Total: $0.058**

**Time saved:**
- Manual triage: 25 minutes
- Automated: 20 seconds
- **Time saved: 24 minutes 40 seconds**

---

### Example 2: False Positive - Legitimate Admin Activity

**Scenario:** Alert for RDP from unusual location

**Input:**
```json
{
  "alert": {
    "id": "ALT-2024-001235",
    "title": "RDP Login from Unusual Location",
    "source_ip": "203.0.113.50",
    "destination_ip": "10.0.2.100",
    "user": "alice-ops",
    "activity": "RDP login",
    "location": "Singapore",
    "severity_raw": "Medium",
    "timestamp": "2024-12-28T10:30:00Z"
  }
}
```

**What Happens:**

```
[10:30:03] Starting alert_enrichment
[10:30:04] Step: extract_iocs
  - Source IP: 203.0.113.50 (Singapore)
  - User: alice-ops
  - Activity: RDP login

[10:30:05] Step: threat_intel_enrichment (parallel)
  - VirusTotal: IP clean, no malicious reports
  - AbuseIPDB: No abuse reports for this IP
  - AD: alice-ops = Alice Wong, DevOps Engineer, Singapore office

[10:30:08] Step: historical_correlation
  - Same user from Singapore: 150 RDP logins last 30 days
  - Pattern: Normal working hours (9am-6pm SGT)
  - Location: Consistent with user's assigned office

[10:30:09] Step: analyze_threat
  - Severity: FALSE_POSITIVE
  - Confidence: HIGH
  - Explanation: User is based in Singapore office, this is normal activity
  - Evidence: 150 previous logins from same location, during business hours

[10:30:09] Step: close_false_positive
  - Closed alert ALT-2024-001235
  - Reason: Legitimate user activity from assigned office location

[10:30:10] ✓ Template completed (7 seconds)
```

**Output:**

```markdown
# Alert Triage Report

**Alert ID:** ALT-2024-001235
**Severity:** FALSE_POSITIVE
**Status:** Auto-closed

## Analysis Summary

This alert is a false positive. User alice-ops (Alice Wong, DevOps Engineer)
is based in the Singapore office and regularly accesses systems via RDP from
this location during business hours.

## Evidence

**User Context:**
- Name: Alice Wong
- Role: DevOps Engineer
- Office: Singapore
- Account status: Active, good standing

**Historical Pattern:**
- 150 RDP logins from Singapore in last 30 days
- All during business hours (9am-6pm SGT)
- Consistent IP range (company Singapore office)

**Threat Intel:**
- Source IP: Clean (no malicious reports)
- Location: Matches user's assigned office
- Time: Within normal business hours

## Disposition

**Status:** Closed automatically
**Reason:** Legitimate user activity
**Action Taken:** No incident created
```

**Cost:** $0.042 (saved SOC analyst 15 minutes)

---

## When to Use

### ✅ Appropriate Use Cases

**High Alert Volume:**
- 100+ alerts per day
- Majority are false positives
- Analysts overwhelmed
- Need automated first-pass triage

**Consistent Investigation Process:**
- Want same enrichment steps every time
- Ensure no data sources skipped
- Standardize across team
- Compliance requires documented process

**24/7 Coverage:**
- Need triage during night shift
- Can't afford alert backlog
- Want immediate escalation for critical threats
- Reduce analyst burnout

**Multi-Source Enrichment:**
- Query 5+ threat intel sources
- Check user context in AD
- Correlate historical patterns
- Too time-consuming manually

### ❌ Inappropriate Use Cases

**Novel Threat Types:**
- Never-seen-before attacks
- Template may not recognize patterns
- Human analysis better for unknowns

**Highly Complex Investigations:**
- APT investigations
- Forensic analysis
- Legal/compliance investigations requiring human judgment

**Very Low Alert Volume:**
- <10 alerts per day
- Manual triage feasible
- Automation overhead not worth it

---

## Trade-offs

### Advantages

**Massive Time Savings:**
- Manual triage: 15-30 minutes/alert
- Automated: 10-30 seconds/alert
- 100 alerts/day = **~40 hours saved/day** (across team)

**Consistent Quality:**
- Every alert gets same enrichment
- No steps skipped due to time pressure
- No variation by analyst skill level
- Compliance: documented process

**24/7 Coverage:**
- Works night shift automatically
- No alert backlog in morning
- Critical alerts escalated immediately
- Analyst burnout reduced

**False Positive Reduction:**
- Auto-close obvious false positives (measured 90% reduction in analyst workload)
- Analysts focus on real threats
- Alert fatigue reduced

### Limitations

**Cannot Replace Human Judgment:**
- Complex investigations still need analysts
- Template provides first-pass triage
- Human review required for low-confidence findings

**API Dependencies:**
- Requires MCP servers for tools (Sentinel, VirusTotal, etc.)
- API rate limits apply
- Cost per enrichment (~$0.05)

**Cost at Scale:**
- 100 alerts/day × $0.05 = $5/day = $150/month
- vs. analyst time saved: 40 hours/day × $50/hr = $2,000/day
- **ROI:** 99.25% cost savings

**Setup Complexity:**
- Requires MCP server configuration
- Threat intel API keys needed
- Integration with existing SIEM
- Testing required before production

---

## Integration Requirements

### MCP Servers Needed

**Microsoft Sentinel:**
```yaml
servers:
  sentinel:
    command: "sentinel-mcp-server"
    env:
      TENANT_ID: "${AZURE_TENANT_ID}"
      SUBSCRIPTION_ID: "${AZURE_SUBSCRIPTION_ID}"
      WORKSPACE_ID: "${SENTINEL_WORKSPACE_ID}"
```

**Threat Intelligence:**
```yaml
servers:
  virustotal:
    command: "virustotal-mcp-server"
    env:
      API_KEY: "${VIRUSTOTAL_API_KEY}"
  
  abuseipdb:
    command: "abuseipdb-mcp-server"
    env:
      API_KEY: "${ABUSEIPDB_API_KEY}"
```

**Ticketing:**
```yaml
servers:
  jira:
    command: "jira-mcp-server"
    env:
      URL: "${JIRA_URL}"
      EMAIL: "${JIRA_EMAIL}"
      API_TOKEN: "${JIRA_API_TOKEN}"
```

---

## Customization

### Adjust Severity Thresholds

```yaml
# In analyze_threat step, modify severity criteria:
- Severity: CRITICAL
  Criteria:
    - Known malware hashes
    - Credential dumping tools
    - C2 communication detected
    - Ransomware indicators
    
- Severity: HIGH
  Criteria:
    - Suspicious PowerShell
    - Unusual process injection
    - Lateral movement attempt
    - Data exfiltration signs
```

### Add Custom Enrichment Sources

```yaml
# Add your proprietary threat intel
- name: internal_threat_intel
  servers: [internal-ti-database]
  prompt: "Check internal threat intel for: {{iocs}}"
  output: internal_ti
```

### Customize Escalation

```yaml
# Different escalation paths by severity
- name: escalate
  condition: "{{threat_analysis.severity}} == 'CRITICAL'"
  servers: [pagerduty, slack]  # Page on-call + post to Slack
  
- name: notify_team
  condition: "{{threat_analysis.severity}} == 'HIGH'"
  servers: [slack]  # Just Slack for high
```

---

## Best Practices

### Before Deployment

**✅ Do:**
- Test with 30 days of historical alerts
- Verify all MCP servers configured correctly
- Set appropriate severity thresholds for your environment
- Define escalation policies clearly
- Document automated vs manual review criteria

**❌ Don't:**
- Deploy without testing against known false positives
- Skip verifying threat intel API limits
- Auto-respond to critical alerts without validation
- Forget to configure backup alerting if template fails

### During Operations

**✅ Do:**
- Review auto-closed false positives weekly (spot check)
- Track false negative rate (missed threats)
- Monitor API rate limits and costs
- Adjust thresholds based on feedback
- Maintain audit log of all automated decisions

**❌ Don't:**
- Trust automation blindly
- Ignore low-confidence findings
- Skip reviewing critical escalations
- Forget to update threat intel sources

---

## Metrics to Track

```
Daily Metrics:
- Total alerts processed: 120
- Auto-closed false positives: 108 (90%)
- Incidents created: 10 (8%)
- Critical escalations: 2 (2%)
- Avg processing time: 15 seconds
- Time saved: 40 analyst hours
- Cost: $6 in API calls

Quality Metrics:
- False positive rate: 1% (1 FP incident per 100 alerts)
- False negative rate: 0.5% (missed threats)
- Analyst override rate: 5% (analysts disagree with severity)

Efficiency Metrics:
- Time to triage: 15s (was 20min)
- Alert backlog: 0 (was 50+)
- Night shift coverage: 100% (was 0%)
```

---

## Related Resources

- **[Template File](../templates/alert_enrichment.yaml)** - Download complete template
- **[Phishing Analysis](phishing-analysis.md)** - Similar workflow for email
- **[Consensus Threat Validation](consensus-threat-validation.md)** - Multi-provider for critical alerts
- **[Why Templates Matter](../../../WHY_TEMPLATES_MATTER.md)** - Context management explained

---

**Alert enrichment transforms SOC operations: analysts focus on real threats, automation handles the noise.**

Remember: This template automates first-pass triage, not complete investigation. Human judgment remains essential for complex threats.
