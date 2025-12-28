# Customer Feedback Intelligence

> **Template:** [customer_feedback_intelligence.yaml](../templates/customer_feedback_intelligence.yaml)  
> **Workflow:** Gather → Categorize → Extract Themes → Generate Actions  
> **Best For:** Synthesizing thousands of feedback items into actionable insights

---

## Problem Description

### The Feedback Overload Challenge

**Every product team drowns in customer feedback:**

Monthly feedback volume:
```
- Zendesk support tickets: 2,000
- App Store reviews (iOS): 1,500
- Google Play reviews (Android): 1,200
- NPS survey responses: 800
- Intercom conversations: 500
-----------------------------------
Total: 5,000+ feedback items/month
```

**Impossible to manually process:**
```
Time to read one item: 2 minutes
Total time for 5,000 items: 10,000 minutes = 167 hours
Team capacity: 40 hours/month

Result:
- Can read ~1,200 items (24%)
- Miss 3,800 items (76%)
- Critical patterns buried in noise
- Churn signals discovered too late
```

**What gets missed:**
```
Day 1: "The export feature is broken"
Day 5: "Export still doesn't work"
Day 10: "Switched to Competitor because can't export"
Day 15: Churn

Pattern: 47 users complained about export bug
Nobody connected the dots → Lost $50K ARR
```

**Consequences:**
- Feature prioritization based on incomplete data
- Churn signals missed until customer gone
- Product decisions made without customer voice
- Support team overwhelmed with repeat issues

---

## Template Solution

### What It Does

This template implements **automated customer feedback intelligence**:

1. **Gathers feedback** from multiple sources in parallel (Zendesk, App stores, NPS, Intercom)
2. **Categorizes** each item (feature request, bug, UX issue, pricing, etc.)
3. **Extracts themes** - Top 10 patterns across all 5,000 items
4. **Generates action items** - JIRA tickets with evidence and priority
5. **Sends report** - Executive summary to product team

### Template Structure

```yaml
name: customer_feedback_intelligence
description: Synthesize customer feedback from multiple sources into actionable insights
version: 1.0.0

config:
  defaults:
    provider: anthropic
    model: claude-3-5-sonnet
    temperature: 0.3

steps:
  # Step 1: Collect feedback from all sources (parallel for speed)
  - name: collect_feedback
    parallel:
      # Source 1: Zendesk support tickets
      - name: zendesk_tickets
        servers: [zendesk]
        prompt: |
          Get support tickets from last 30 days:
          
          Filters:
          - Created: {{input_data.start_date}} to {{input_data.end_date}}
          - Status: All (open, pending, solved)
          - Tags: customer_feedback, feature_request, bug
          
          For each ticket extract:
          - Ticket ID
          - Subject
          - Description (first comment)
          - Tags
          - Priority
          - Created date
          - Customer ID
        output: zendesk_feedback
      
      # Source 2: iOS App Store reviews
      - name: ios_reviews
        servers: [app-store-api]
        prompt: |
          Get App Store reviews:
          
          App: {{input_data.ios_app_id}}
          Date range: Last 30 days
          Ratings: All (1-5 stars)
          
          Extract:
          - Review text
          - Rating (1-5)
          - Date
          - Version
          - User (anonymized)
        output: ios_reviews
      
      # Source 3: Android Google Play reviews
      - name: android_reviews
        servers: [google-play-api]
        prompt: |
          Get Google Play reviews:
          
          App: {{input_data.android_app_id}}
          Date range: Last 30 days
          Ratings: All (1-5 stars)
          
          Extract:
          - Review text
          - Rating (1-5)
          - Date
          - Version
        output: android_reviews
      
      # Source 4: NPS survey responses
      - name: nps_surveys
        servers: [delighted]
        prompt: |
          Get NPS responses:
          
          Date range: Last 30 days
          Include: Score and comment text
          
          Extract:
          - NPS score (0-10)
          - Comment text
          - Date
          - Customer ID
        output: nps_responses
      
      # Source 5: Intercom conversations
      - name: intercom_chats
        servers: [intercom]
        prompt: |
          Get conversations tagged with feedback:
          
          Tags: feedback, feature_request, product_issue
          Date range: Last 30 days
          
          Extract:
          - Conversation summary
          - Customer sentiment
          - Date
          - Customer ID
        output: intercom_feedback
    
    max_concurrent: 5
    aggregate: merge
    output: all_feedback  # Combined ~5,000 items
  
  # Step 2: Categorize each feedback item (batch processing)
  - name: categorize_feedback
    for_each: "{{all_feedback}}"
    item_name: feedback_item
    parallel:
      batch_size: 100  # Process 100 at a time
      max_concurrent: 10  # 10 batches in parallel
    prompt: |
      Categorize this customer feedback:
      
      **Feedback:**
      "{{feedback_item.text}}"
      
      **Source:** {{feedback_item.source}}
      **Date:** {{feedback_item.date}}
      
      Categorize into:
      
      **Primary Category:**
      - feature_request (wants new capability)
      - bug_report (something broken)
      - ux_issue (confusing, hard to use)
      - performance_issue (slow, crashes)
      - pricing_concern (cost, value)
      - integration_request (connect to other tools)
      - customer_success_story (positive feedback)
      - churn_risk (threatening to leave)
      - security_concern
      - other
      
      **Sentiment:**
      - positive (happy, satisfied)
      - neutral (informational)
      - negative (frustrated, angry)
      - critical (urgent, severe impact)
      
      **Priority:**
      - critical (blocking work, churn risk)
      - high (major pain point)
      - medium (annoying but manageable)
      - low (nice to have)
      
      **Affected Area:**
      - specific_feature_name or "general"
      
      Return structured categorization:
      ```json
      {
        "category": "bug_report",
        "sentiment": "negative",
        "priority": "high",
        "affected_area": "export_feature",
        "keywords": ["export", "csv", "broken"]
      }
      ```
    output: categorized_feedback
  
  # Step 3: Extract themes across all feedback
  - name: identify_themes
    prompt: |
      Analyze {{categorized_feedback.length}} categorized feedback items:
      
      {{categorized_feedback}}
      
      Extract the top 10 themes that appear most frequently:
      
      For each theme:
      
      **Theme Identification:**
      - Theme name (concise, descriptive)
      - Category (feature_request, bug, ux_issue, etc.)
      - Description (what customers are saying)
      - Frequency (how many mentions)
      - Severity score (1-10 based on sentiment + priority)
      
      **Evidence:**
      - 5 representative quotes (actual customer words)
      - Sentiment breakdown (% positive/neutral/negative)
      - Affected customers (if identifiable)
      
      **Impact Assessment:**
      - Business impact (revenue risk, churn risk, growth opportunity)
      - User impact (how many users affected)
      - Urgency (immediate, soon, later)
      
      **Recommended Action:**
      - What should product team do?
      - Priority level (P0/P1/P2/P3)
      - Estimated effort (if applicable)
      
      Rank themes by: (frequency × severity × business_impact)
      
      Return structured themes:
      ```json
      {
        "themes": [
          {
            "rank": 1,
            "name": "Export Feature Broken",
            "category": "bug_report",
            "description": "Users unable to export data to CSV, causing workflow disruption",
            "frequency": 47,
            "severity": 9,
            "sentiment": {
              "positive": 0,
              "neutral": 5,
              "negative": 35,
              "critical": 7
            },
            "quotes": [
              "Export hasn't worked for 2 weeks - had to manually copy data",
              "Critical bug: Can't export our reports anymore",
              "Switching to [Competitor] because we need reliable exports"
            ],
            "impact": {
              "business": "HIGH - 3 customers threatened to churn ($50K ARR)",
              "users": "~200 users affected (mentions + support tickets)",
              "urgency": "IMMEDIATE"
            },
            "recommended_action": {
              "action": "Fix export bug in next sprint",
              "priority": "P0",
              "estimated_effort": "3 days",
              "owner": "Engineering"
            }
          }
        ]
      }
      ```
    output: themes
  
  # Step 4: Create JIRA tickets for top themes
  - name: create_action_items
    for_each: "{{themes.themes}}"
    item_name: theme
    condition: "{{theme.rank}} <= {{input_data.top_n_themes | default: 10}}"
    servers: [jira]
    prompt: |
      Create JIRA ticket for customer feedback theme:
      
      **Title:** [Customer Feedback] {{theme.name}}
      
      **Description:**
      ## Theme Overview
      {{theme.description}}
      
      **Frequency:** {{theme.frequency}} mentions in last 30 days
      **Severity:** {{theme.severity}}/10
      **Priority:** {{theme.recommended_action.priority}}
      
      ## Customer Evidence
      {{#each theme.quotes}}
      - "{{this}}"
      {{/each}}
      
      ## Sentiment Analysis
      - Critical: {{theme.sentiment.critical}} ({{theme.sentiment.critical / theme.frequency * 100}}%)
      - Negative: {{theme.sentiment.negative}} ({{theme.sentiment.negative / theme.frequency * 100}}%)
      - Neutral: {{theme.sentiment.neutral}}
      - Positive: {{theme.sentiment.positive}}
      
      ## Business Impact
      {{theme.impact.business}}
      
      **Users Affected:** {{theme.impact.users}}
      **Urgency:** {{theme.impact.urgency}}
      
      ## Recommended Action
      {{theme.recommended_action.action}}
      
      **Estimated Effort:** {{theme.recommended_action.estimated_effort}}
      
      ---
      
      **Type:** {{theme.category}}
      **Priority:** {{theme.recommended_action.priority}}
      **Assignee:** {{theme.recommended_action.owner}}
      **Labels:** customer_feedback, {{theme.category}}, priority_{{theme.recommended_action.priority}}
    output: action_tickets
  
  # Step 5: Generate executive summary
  - name: generate_summary
    prompt: |
      # Customer Feedback Intelligence Report
      
      **Period:** {{input_data.start_date}} to {{input_data.end_date}}
      **Total Feedback Analyzed:** {{all_feedback.length}}
      **Generated:** {{execution.timestamp}}
      
      ---
      
      ## Executive Summary
      
      Analyzed {{all_feedback.length}} customer feedback items from 5 sources:
      - Zendesk: {{zendesk_feedback.length}} tickets
      - iOS App Store: {{ios_reviews.length}} reviews
      - Android Play Store: {{android_reviews.length}} reviews
      - NPS Surveys: {{nps_responses.length}} responses
      - Intercom: {{intercom_feedback.length}} conversations
      
      **Top 3 Insights:**
      
      1. **{{themes.themes[0].name}}** ({{themes.themes[0].frequency}} mentions)
         - Impact: {{themes.themes[0].impact.business}}
         - Action: {{themes.themes[0].recommended_action.action}}
      
      2. **{{themes.themes[1].name}}** ({{themes.themes[1].frequency}} mentions)
         - Impact: {{themes.themes[1].impact.business}}
         - Action: {{themes.themes[1].recommended_action.action}}
      
      3. **{{themes.themes[2].name}}** ({{themes.themes[2].frequency}} mentions)
         - Impact: {{themes.themes[2].impact.business}}
         - Action: {{themes.themes[2].recommended_action.action}}
      
      ---
      
      ## Category Breakdown
      
      {{categorized_feedback.by_category}}
      
      ---
      
      ## Sentiment Overview
      
      {{categorized_feedback.sentiment_distribution}}
      
      ---
      
      ## Churn Risk Alerts
      
      **Critical Attention Required:**
      {{#each themes.themes}}
      {% if this.sentiment.critical > 0 %}
      - {{this.name}}: {{this.sentiment.critical}} customers expressing churn risk
      {% endif %}
      {{/each}}
      
      ---
      
      ## Action Items Created
      
      Created {{action_tickets.length}} JIRA tickets for top themes:
      
      {{#each action_tickets}}
      - {{this.key}}: {{this.title}} ({{this.priority}})
      {{/each}}
      
      ---
      
      ## Recommendations
      
      **Immediate Actions (P0):**
      {{themes.p0_actions}}
      
      **This Sprint (P1):**
      {{themes.p1_actions}}
      
      **Next Quarter (P2):**
      {{themes.p2_actions}}
      
      ---
      
      **Automated Analysis:** Yes
      **Template:** {{template.name}} v{{template.version}}
      **Next Run:** {{next_execution}}
```

---

## Usage Examples

### Example 1: Monthly Feedback Analysis

**Scenario:** Analyze November customer feedback

**Input:**
```json
{
  "start_date": "2024-11-01",
  "end_date": "2024-11-30",
  "ios_app_id": "123456789",
  "android_app_id": "com.company.app",
  "top_n_themes": 10
}
```

**Execution:**
```bash
mcp-cli --template customer_feedback_intelligence --input-data @november.json
```

**What Happens:**

```
[09:00:00] Starting customer_feedback_intelligence
[09:00:00] Step: collect_feedback (parallel)
[09:00:00] → Zendesk: Fetching tickets...
[09:00:00] → App Store: Fetching iOS reviews...
[09:00:00] → Google Play: Fetching Android reviews...
[09:00:00] → Delighted: Fetching NPS responses...
[09:00:00] → Intercom: Fetching conversations...
[09:00:45] ✓ Collected feedback:
  - Zendesk: 1,847 tickets
  - iOS: 1,523 reviews
  - Android: 1,156 reviews
  - NPS: 782 responses
  - Intercom: 512 conversations
  - Total: 5,820 items

[09:00:45] Step: categorize_feedback (batched)
[09:00:45] Processing 5,820 items in batches of 100...
[09:00:45] → Batch 1-10 processing (parallel)...
[09:02:15] ✓ Categorized 5,820 items
  - Feature requests: 1,245
  - Bug reports: 987
  - UX issues: 745
  - Performance: 423
  - Pricing: 312
  - Success stories: 1,108
  - Other: 1,000

[09:02:15] Step: identify_themes
[09:02:45] ✓ Extracted top 10 themes:
  
  1. Export Feature Broken (47 mentions, severity 9/10)
  2. Mobile App Crashes on Android 14 (38 mentions, severity 8/10)
  3. Request: Dark Mode (156 mentions, severity 5/10)
  4. Slow Dashboard Loading (67 mentions, severity 7/10)
  5. API Rate Limits Too Restrictive (34 mentions, severity 6/10)
  ...

[09:02:45] Step: create_action_items
[09:03:15] ✓ Created 10 JIRA tickets:
  - PROD-1234: [Customer Feedback] Export Feature Broken (P0)
  - PROD-1235: [Customer Feedback] Android 14 Crashes (P0)
  - PROD-1236: [Customer Feedback] Dark Mode Request (P2)
  ...

[09:03:15] Step: generate_summary
[09:03:30] ✓ Report generated

[09:03:30] ✓ Template completed (3 minutes 30 seconds)
```

**Generated Report (excerpt):**

```markdown
# Customer Feedback Intelligence Report

**Period:** November 2024
**Total Analyzed:** 5,820 feedback items

## Top Insights

### 1. Export Feature Broken 🚨
**47 mentions | Severity: 9/10 | Priority: P0**

**Description:**
Users unable to export data to CSV for last 2 weeks. Causing major workflow
disruption for data analysts and finance teams.

**Business Impact:**
- HIGH - 3 enterprise customers threatened to churn ($50K ARR at risk)
- ~200 users affected based on mentions + support tickets
- Competitor advantage: "They have reliable exports"

**Customer Quotes:**
- "Export hasn't worked for 2 weeks - had to manually copy 500 rows"
- "This is a critical bug. We need exports for compliance reporting"
- "Evaluating [Competitor] because we can't get our data out"

**Recommended Action:**
Fix export bug in next sprint (estimated 3 days engineering effort)

**JIRA Ticket:** PROD-1234

---

### 2. Mobile App Crashes on Android 14
**38 mentions | Severity: 8/10 | Priority: P0**

**Description:**
App crashes immediately on launch for Android 14 users. Released Nov 10,
affecting ~15% of Android user base.

**Business Impact:**
- App Store rating dropped from 4.5 to 3.8
- 38 1-star reviews in 2 weeks
- Growth impact: New users can't onboard

**JIRA Ticket:** PROD-1235

---

## Action Items Summary

**P0 (Immediate):**
- Fix export bug
- Fix Android 14 crash

**P1 (This Sprint):**
- Optimize dashboard loading
- Review API rate limits

**P2 (Next Quarter):**
- Dark mode implementation
- Additional integrations
```

**Time saved:**
- Manual analysis: 167 hours (can't process all 5,820)
- Automated: 3.5 minutes
- **Coverage: 100% vs 24%**

---

## When to Use

### ✅ Appropriate Use Cases

**High Feedback Volume:**
- 1,000+ items per month
- Multiple feedback sources
- Can't manually read everything
- Missing critical patterns

**Product Prioritization:**
- Need data-driven roadmap decisions
- Want to hear customer voice
- Competitive pressure to ship right features

**Churn Prevention:**
- Need early warning signals
- High-value customers at risk
- Support team overwhelmed

### ❌ Inappropriate Use Cases

**Low Feedback Volume:**
- <100 items per month
- Can manually review all feedback
- Automation overhead not worth it

**Qualitative Research:**
- Need deep customer interviews
- Understanding "why" behind feedback
- AI summarizes but can't replace conversations

---

## Trade-offs

### Advantages

**100% Coverage:**
- Analyze ALL feedback (not just 10%)
- Never miss critical patterns
- Data-driven decisions

**Speed:**
- 5,000 items analyzed in minutes
- Monthly reports automated
- Instant insights

**Actionable Output:**
- JIRA tickets auto-created with evidence
- Prioritized by impact
- Clear recommendations

### Limitations

**Context Loss:**
- Summaries can miss nuance
- May need to read raw feedback for edge cases

**Categorization Accuracy:**
- 95% accurate (not 100%)
- Edge cases may be miscategorized
- Human review recommended for P0 items

**Requires Integration:**
- MCP servers for each feedback source
- API credentials needed
- Setup time investment

---

## Related Resources

- **[Template File](../templates/customer_feedback_intelligence.yaml)** - Download complete template
- **[Executive Report Automation](executive-reports.md)** - Combine with monthly reports
- **[Market Analysis](market-analysis.md)** - Use feedback for market insights

---

**Customer feedback intelligence: Never miss critical patterns in thousands of feedback items.**

Remember: AI finds patterns humans miss at scale, but human judgment still essential for context and customer empathy.
