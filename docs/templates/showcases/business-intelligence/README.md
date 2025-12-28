# Business Intelligence & Analytics Templates

> **For:** Business Analysts, Data Analysts, Market Researchers, Product Managers  
> **Purpose:** Automate recurring BI workflows with multi-stage template composition

---

## What This Showcase Contains

This section demonstrates how templates automate critical business intelligence workflows through **template composition** - where complex multi-stage analyses are built by chaining smaller, reusable templates. All examples solve real BI challenges: customer feedback overload, manual monthly reports, time-consuming market research, and competitive intelligence gathering.

### Available Use Cases

**BI Automation:**
1. **[Customer Feedback Intelligence](use-cases/customer-feedback.md)** - Synthesize 5,000+ feedback items into actionable insights
2. **[Executive Report Automation](use-cases/executive-reports.md)** - Auto-generate monthly KPI decks from multiple data sources
3. **[Market Opportunity Analysis](use-cases/market-analysis.md)** - Size markets in hours instead of weeks
4. **[Competitive Intelligence Dashboard](use-cases/competitive-intelligence.md)** - Automated weekly competitor monitoring
5. **[Trend Detection & Early Warning](use-cases/trend-detection.md)** - Monitor emerging market trends automatically

---

## Why Templates Matter for Business Intelligence

### 1. Template Composition for Multi-Stage Workflows

**The Challenge:** BI workflows have multiple stages (gather data → clean → analyze → visualize → report). Building these as monolithic templates is brittle and hard to maintain.

**Template Solution:** Compose complex workflows from reusable building blocks

```yaml
# Master template that orchestrates sub-workflows
name: comprehensive_market_analysis

steps:
  # Stage 1: Research phase (calls web_research template)
  - name: gather_market_data
    template: web_research_workflow
    template_input:
      query: "{{input_data.market}} market size"
      sources: ["industry_reports", "competitors", "news"]
    output: research_data
  
  # Stage 2: Data extraction (calls extraction template)
  - name: extract_metrics
    template: metric_extraction_workflow
    template_input: "{{research_data}}"
    output: structured_metrics
  
  # Stage 3: Analysis (calls TAM/SAM/SOM template)
  - name: size_market
    template: tam_analysis_workflow
    template_input: "{{structured_metrics}}"
    output: market_sizing
  
  # Stage 4: Report generation (calls presentation template)
  - name: create_executive_deck
    template: executive_presentation_workflow
    template_input:
      data: "{{market_sizing}}"
      format: "powerpoint"
    output: final_presentation
```

**Benefits:**
- **Reusability:** Research template used for market sizing, competitor analysis, trend detection
- **Maintainability:** Update one sub-template, all workflows benefit
- **Testability:** Test each stage independently
- **Flexibility:** Swap TAM analysis for different methodology without changing research

**Impact:**
- Manual research: 2 weeks (80 hours)
- Automated workflow: 2 hours
- **Time saved: 97.5%**

**Documentation:** [Market Opportunity Analysis](use-cases/market-analysis.md)

---

### 2. Recurring Automation for Monthly/Weekly Reports

**The Challenge:** Monthly executive reports take 4 hours of manual data gathering every month. Quarterly competitive analysis takes 8 hours. Wastes analyst time on repetitive work.

**Template Solution:** Schedule recurring workflows that run automatically

```yaml
name: monthly_executive_report

# Runs automatically on 1st of each month
schedule:
  frequency: monthly
  day: 1
  time: "06:00"

steps:
  # Gather metrics from multiple platforms in parallel
  - name: collect_metrics
    parallel:
      - name: web_analytics
        servers: [google-analytics]
        prompt: "Get last 30 days: users, sessions, conversions"
        output: web_metrics
      
      - name: product_analytics
        servers: [mixpanel]
        prompt: "Get last 30 days: DAU, MAU, retention"
        output: product_metrics
      
      - name: sales_metrics
        servers: [salesforce]
        prompt: "Get last 30 days: ARR, new deals, pipeline"
        output: sales_metrics
    
    aggregate: merge
    output: all_metrics
  
  # Calculate KPIs and trends
  - name: calculate_kpis
    prompt: |
      Calculate KPIs with MoM and YoY comparisons:
      
      Data: {{all_metrics}}
      
      For each metric:
      - Current value
      - MoM change (% and absolute)
      - YoY change (% and absolute)
      - Trend (up/down/flat)
      - Status vs goal (on track/at risk/off track)
    output: kpis
  
  # Generate insights
  - name: generate_insights
    prompt: |
      Analyze KPIs and generate executive insights:
      
      {{kpis}}
      
      Identify:
      - Top 3 wins (biggest positive changes)
      - Top 3 concerns (negative trends or at-risk goals)
      - Key drivers (what's causing the changes)
      - Recommended actions
    output: insights
  
  # Build PowerPoint deck
  - name: create_presentation
    servers: [powerpoint]
    template: executive_deck_generator
    template_input:
      title: "Monthly Business Review - {{current_month}}"
      kpis: "{{kpis}}"
      insights: "{{insights}}"
      template: "executive_monthly.pptx"
    output: deck
  
  # Distribute to stakeholders
  - name: distribute
    servers: [slack, email]
    prompt: |
      Send monthly report:
      - Slack: #executive-team with summary
      - Email: execs@company.com with deck attached
      
      Message: "Monthly report ready: {{insights.summary}}"
```

**Impact:**
- Manual effort: 4 hours/month = 48 hours/year
- Automated: Runs while you sleep
- **Time saved: 100%** (48 hours/year @ $100/hr = $4,800)

**Documentation:** [Executive Report Automation](use-cases/executive-reports.md)

---

### 3. Multi-Source Data Integration

**The Challenge:** Customer feedback scattered across Zendesk, App Store reviews, NPS surveys, Intercom conversations. Can't manually process 5,000 items/month to find patterns.

**Template Solution:** Parallel data gathering from multiple sources

```yaml
name: customer_feedback_intelligence

steps:
  # Gather from all sources in parallel (fast)
  - name: collect_feedback
    parallel:
      - name: support_tickets
        servers: [zendesk]
        prompt: "Get last 30 days tickets with customer feedback"
        output: tickets
      
      - name: app_reviews
        servers: [app-store-api, google-play-api]
        prompt: "Get last 30 days app reviews (iOS + Android)"
        output: reviews
      
      - name: nps_surveys
        servers: [delighted]
        prompt: "Get last 30 days NPS responses with comments"
        output: nps
      
      - name: chat_conversations
        servers: [intercom]
        prompt: "Get conversations tagged 'feedback' last 30 days"
        output: chats
    
    max_concurrent: 4
    aggregate: merge
    output: all_feedback  # ~5,000 items combined
  
  # Categorize all feedback
  - name: categorize
    for_each: "{{all_feedback}}"
    prompt: |
      Categorize feedback:
      
      "{{item.text}}"
      
      Categories:
      - feature_request
      - bug_report
      - ux_issue
      - pricing_concern
      - customer_success
      - churn_risk
      
      Sentiment: positive/neutral/negative
      Priority: critical/high/medium/low
    output: categorized_feedback
  
  # Extract themes
  - name: identify_themes
    prompt: |
      Analyze {{categorized_feedback.length}} feedback items:
      
      {{categorized_feedback}}
      
      Extract top 10 themes:
      - Theme name
      - Description
      - Frequency (how many mentions)
      - Sentiment distribution
      - Example quotes (3-5)
      - Recommended action
      
      Rank by: (frequency × severity)
    output: themes
  
  # Generate action items
  - name: create_actions
    servers: [jira]
    for_each: "{{themes.top_themes}}"
    prompt: |
      Create JIRA ticket for theme: {{item.name}}
      
      Title: [Customer Feedback] {{item.name}}
      Description: {{item.description}}
      Evidence: {{item.frequency}} mentions
      Examples: {{item.example_quotes}}
      Priority: {{item.priority}}
      
      Assign to: Product team
    output: action_tickets
```

**Impact:**
- Before: Can't process 5,000 items, miss critical patterns
- After: All feedback analyzed, top 10 themes with evidence, action items created
- **Value:** Catch churn signals before losing customers

**Documentation:** [Customer Feedback Intelligence](use-cases/customer-feedback.md)

---

### 4. Context-Efficient Large Dataset Processing

**The Challenge:** Analyzing 5,000 customer feedback items exceeds context limits. Can't load everything into single prompt.

**Template Solution:** Batch processing with aggregation

**Traditional approach (fails):**
```
LLM Context (200K tokens):
├── 5,000 feedback items × 100 tokens = 500K tokens
└── ERROR: Context overflow
```

**Template approach (scales):**
```
Process in batches:

Batch 1 (500 items):
├── LLM categorizes: Fresh 200K context
└── Output: Categorized batch 1

Batch 2 (500 items):
├── LLM categorizes: Fresh 200K context
└── Output: Categorized batch 2

...

Batch 10 (500 items):
├── LLM categorizes: Fresh 200K context
└── Output: Categorized batch 10

Theme extraction:
├── LLM receives: 10 batch summaries (25K tokens)
└── Output: Top 10 themes across all 5,000 items
```

**Benefits:**
- Process unlimited feedback volume
- Each batch gets full context
- Parallel batches for speed
- Aggregated insights

---

### 5. Scheduled Workflows for Competitive Intelligence

**The Challenge:** Manually check competitor pricing/features every week. Always outdated, takes 4 hours.

**Template Solution:** Automated recurring monitoring

```yaml
name: weekly_competitor_intelligence

schedule:
  frequency: weekly
  day: monday
  time: "06:00"

steps:
  # Scrape competitor websites
  - name: scrape_competitors
    for_each: "{{input_data.competitors}}"
    parallel:
      max_concurrent: 5
    servers: [web-scraping]
    prompt: |
      Scrape competitor site: {{item.url}}
      
      Extract:
      - Pricing tiers and features
      - Product updates/announcements
      - Customer testimonials
      - Marketing messaging
    output: competitor_data
  
  # Compare with historical data
  - name: detect_changes
    servers: [database]
    prompt: |
      Compare current vs historical:
      
      Current: {{competitor_data}}
      Historical: {{load_from_database('competitors', 'last_week')}}
      
      Detect:
      - Price changes
      - New features launched
      - Messaging shifts
      - New customers won
    output: changes
  
  # Generate insights
  - name: analyze_changes
    prompt: |
      Analyze competitive changes:
      
      {{changes}}
      
      Assess:
      - Threats: What puts us at risk?
      - Opportunities: Where can we differentiate?
      - Recommended responses
    output: insights
  
  # Update dashboard
  - name: update_dashboard
    servers: [tableau]
    prompt: "Update competitor intelligence dashboard: {{competitor_data}}"
  
  # Alert team on significant changes
  - name: send_alerts
    condition: "{{changes.significant}} == true"
    servers: [slack]
    prompt: |
      🚨 Competitive Intelligence Alert
      
      {{insights.threats}}
      
      Dashboard: {{dashboard_url}}
```

**Impact:**
- Manual: 4 hours/week = 208 hours/year
- Automated: Runs automatically every Monday
- **Savings: $20,800/year** (208 hours @ $100/hr)

**Documentation:** [Competitive Intelligence Dashboard](use-cases/competitive-intelligence.md)

---

### 6. MCP Integration: Business Data Sources

**The Challenge:** BI data scattered across Google Analytics, Salesforce, Zendesk, Mixpanel, databases.

**Template Solution:** MCP servers expose business tools as data sources

```yaml
# Template can query multiple business systems:
steps:
  # Web analytics
  - name: get_web_data
    servers: [google-analytics]
    prompt: "Get traffic, conversions, bounce rate for last 30 days"
  
  # Product analytics
  - name: get_product_data
    servers: [mixpanel]
    prompt: "Get DAU, MAU, retention cohorts"
  
  # Sales data
  - name: get_sales_data
    servers: [salesforce]
    prompt: "Get ARR, pipeline, win rate"
  
  # Customer support
  - name: get_support_data
    servers: [zendesk]
    prompt: "Get ticket volume, CSAT, response time"
  
  # Survey data
  - name: get_survey_data
    servers: [typeform, surveymonkey]
    prompt: "Get NPS score and responses"
```

**What this enables:**
- Unified data access across platforms
- No manual CSV exports
- Templates become portable (same template, different company)
- Version control for business queries

---

## Quick Start

### 1. Choose Your BI Challenge

**Too much customer feedback?**
- [Customer Feedback Intelligence](use-cases/customer-feedback.md) - Synthesize 5,000+ items into top 10 themes

**Manual monthly reports?**
- [Executive Report Automation](use-cases/executive-reports.md) - Auto-generate KPI decks every month

**Need to size a market?**
- [Market Opportunity Analysis](use-cases/market-analysis.md) - Research → Size → Report in hours

**Track competitors manually?**
- [Competitive Intelligence Dashboard](use-cases/competitive-intelligence.md) - Automated weekly monitoring

**Miss emerging trends?**
- [Trend Detection & Early Warning](use-cases/trend-detection.md) - Daily monitoring and alerts

### 2. Set Up MCP Integrations

BI templates integrate with business tools:

```yaml
# Analytics platforms
servers:
  google-analytics:
    command: "ga-mcp-server"
    env:
      PROPERTY_ID: "${GA_PROPERTY_ID}"
      CREDENTIALS: "${GA_CREDENTIALS}"
  
  mixpanel:
    command: "mixpanel-mcp-server"
    env:
      PROJECT_ID: "${MIXPANEL_PROJECT}"
      API_SECRET: "${MIXPANEL_SECRET}"

# CRM and sales
servers:
  salesforce:
    command: "salesforce-mcp-server"
    env:
      INSTANCE_URL: "${SF_INSTANCE_URL}"
      ACCESS_TOKEN: "${SF_ACCESS_TOKEN}"

# Customer feedback
servers:
  zendesk:
    command: "zendesk-mcp-server"
    env:
      SUBDOMAIN: "${ZENDESK_SUBDOMAIN}"
      API_TOKEN: "${ZENDESK_API_TOKEN}"
  
  app-store-api:
    command: "appstore-mcp-server"
    env:
      APP_ID: "${IOS_APP_ID}"
```

### 3. Run Template

```bash
# Analyze customer feedback
mcp-cli --template customer_feedback_intelligence --input-data "{
  \"date_range\": \"last_30_days\"
}"

# Generate monthly report
mcp-cli --template monthly_executive_report

# Size a market
mcp-cli --template market_opportunity_analysis --input-data "{
  \"market\": \"AI code assistants\",
  \"geography\": \"North America\"
}"
```

---

## Integration Patterns

### Pattern 1: Template Composition (Multi-Stage Workflow)

**Build complex workflows from reusable components:**

```yaml
name: comprehensive_competitor_analysis

steps:
  # Stage 1: Research (reusable template)
  - name: research_competitors
    template: web_research_workflow
    template_input:
      targets: "{{input_data.competitors}}"
      data_points: ["pricing", "features", "customers"]
    output: research_data
  
  # Stage 2: Extraction (reusable template)
  - name: extract_structured_data
    template: data_extraction_workflow
    template_input: "{{research_data}}"
    output: competitor_matrix
  
  # Stage 3: Analysis (reusable template)
  - name: swot_analysis
    template: competitive_analysis_workflow
    template_input: "{{competitor_matrix}}"
    output: strategic_insights
  
  # Stage 4: Presentation (reusable template)
  - name: create_deck
    template: presentation_generator_workflow
    template_input:
      data: "{{strategic_insights}}"
      template: "competitor_analysis.pptx"
    output: final_deck
```

**Benefits:**
- Each sub-template is reusable across multiple master templates
- Test and maintain stages independently
- Swap implementations (different research methods) without changing workflow
- Clear separation of concerns

---

### Pattern 2: Recurring Scheduled Analysis

**Automate weekly/monthly reports:**

```yaml
name: quarterly_business_review

schedule:
  frequency: quarterly
  month_day: 1

steps:
  # Pull 90 days of data
  - name: gather_quarterly_data
    template: multi_source_data_collection
    template_input:
      date_range: "last_90_days"
      sources: ["analytics", "sales", "support", "finance"]
    output: quarterly_data
  
  # Calculate QoQ trends
  - name: calculate_trends
    template: trend_analysis
    template_input: "{{quarterly_data}}"
    output: trends
  
  # Generate insights
  - name: generate_insights
    template: insight_generation
    template_input: "{{trends}}"
    output: insights
  
  # Create board deck
  - name: create_board_deck
    template: board_presentation
    template_input: "{{insights}}"
    output: presentation
  
  # Distribute
  - name: send_to_board
    servers: [email, slack]
```

**Result:** Quarterly board deck auto-generated on schedule

---

### Pattern 3: Event-Triggered Workflows

**React to business events:**

```yaml
name: churn_risk_alert

# Triggered by: NPS score < 6
trigger:
  event: nps_response
  condition: "score < 6"

steps:
  # Gather customer context
  - name: get_customer_data
    parallel:
      - servers: [salesforce]
        prompt: "Get account details for: {{trigger.customer_id}}"
      
      - servers: [mixpanel]
        prompt: "Get product usage for: {{trigger.customer_id}}"
      
      - servers: [zendesk]
        prompt: "Get support history for: {{trigger.customer_id}}"
    output: customer_context
  
  # Analyze churn risk
  - name: assess_risk
    template: churn_risk_assessment
    template_input: "{{customer_context}}"
    output: risk_assessment
  
  # Create action plan
  - name: generate_action_plan
    prompt: |
      Customer at churn risk:
      
      Context: {{customer_context}}
      Risk: {{risk_assessment}}
      
      Recommend:
      - Immediate actions for CSM
      - Escalation path
      - Retention offer suggestions
    output: action_plan
  
  # Alert customer success
  - name: alert_csm
    servers: [slack]
    prompt: |
      🚨 Churn Risk Alert
      
      Customer: {{customer_context.name}}
      ARR: {{customer_context.arr}}
      Risk Score: {{risk_assessment.score}}
      
      Action Plan: {{action_plan}}
```

---

## Best Practices

### BI Template Design

**✅ Do:**
- Build reusable sub-templates (research, analysis, reporting)
- Schedule recurring reports to run automatically
- Process large datasets in batches
- Use parallel execution for multi-source data gathering
- Generate actionable insights, not just data dumps
- Include data sources in reports (for auditability)

**❌ Don't:**
- Build monolithic templates (hard to maintain)
- Manually run monthly reports (automate them)
- Load entire dataset into single context (use batching)
- Skip validation of data quality
- Generate 50-slide decks (keep insights concise)
- Forget to version control templates

### Data Quality

**✅ Do:**
- Validate data ranges (catch missing days)
- Check for anomalies (10× spike = data issue?)
- Include data freshness timestamps
- Handle API rate limits gracefully
- Log data source and collection time

**❌ Don't:**
- Trust data blindly without validation
- Mix data from different time periods
- Skip checking for null values
- Ignore API errors silently

---

## Measuring Success

### Customer Feedback Analysis

**Before templates:**
- Feedback items: 5,000/month
- Can read: ~500 (10%)
- Patterns found: Ad-hoc, inconsistent
- Time to insights: 2 weeks
- Action items created: Manual, incomplete

**After templates:**
- Feedback analyzed: 5,000 (100%)
- Top 10 themes identified automatically
- Time to insights: 2 hours
- Action items: Auto-created in JIRA
- **Coverage: 10× better** (500 → 5,000 items)

### Executive Report Automation

**Before templates:**
- Manual effort: 4 hours/month
- Data freshness: Updated manually
- Consistency: Varies by analyst
- Distribution: Manual email
- Annual cost: 48 hours @ $100/hr = $4,800

**After templates:**
- Manual effort: 0 hours (runs automatically)
- Data freshness: Always current (scheduled)
- Consistency: Identical format every month
- Distribution: Automated (Slack + email)
- **Savings: $4,800/year**

### Market Research Automation

**Before templates:**
- Time to size market: 2 weeks (80 hours)
- Data sources: Manually googled
- Analysis: Spreadsheets
- Report: Manual PowerPoint
- Cost: $8,000 per analysis

**After templates:**
- Time to size market: 2 hours
- Data sources: Automated web research
- Analysis: Template-driven
- Report: Auto-generated deck
- Cost: $200 per analysis
- **Savings: 97.5%** ($7,800 per analysis)

---

## Cost Analysis

### Customer Feedback Intelligence

**Per month:**
- AI cost: $5 (categorizing 5,000 items)
- MCP API costs: $2 (Zendesk, app stores, etc.)
- Total: $7/month

**Value delivered:**
- Analyze 100% of feedback (was 10%)
- Catch churn signals early (save $50K ARR loss)
- **ROI: 7,000×** ($7 cost vs $50K value)

### Executive Reports

**Per month:**
- AI cost: $1 (KPI analysis + insights)
- MCP API costs: $1 (analytics platforms)
- Total: $2/month

**Value delivered:**
- Saves 4 hours analyst time ($400)
- Always fresh data (prevent bad decisions)
- **ROI: 200×** ($2 cost vs $400 value)

### Market Research

**Per analysis:**
- AI cost: $10 (research + analysis)
- MCP API costs: $5 (web scraping)
- Total: $15 per analysis

**Value delivered:**
- Saves 78 hours analyst time ($7,800)
- Faster time to market (2 hours vs 2 weeks)
- **ROI: 520×** ($15 cost vs $7,800 value)

---

## Template Library

All templates available in [templates/](templates/):

**Multi-Stage Workflows (Composition):**
- `market_opportunity_analysis.yaml` - Research → Size → Report
- `customer_feedback_intelligence.yaml` - Gather → Categorize → Themes → Actions
- `competitive_analysis.yaml` - Research → Extract → Analyze → Present

**Recurring Reports:**
- `monthly_executive_report.yaml` - Scheduled monthly KPI deck
- `weekly_competitor_monitor.yaml` - Automated competitor tracking
- `daily_trend_detection.yaml` - Emerging trend monitoring

**Sub-Templates (Reusable Components):**
- `web_research_workflow.yaml` - Multi-source research
- `data_extraction_workflow.yaml` - Structured data extraction
- `tam_analysis_workflow.yaml` - Market sizing methodology
- `presentation_generator.yaml` - Auto-create PowerPoint

---

## Example: Complete BI Pipeline

```yaml
name: complete_market_entry_analysis

# Composition: Multiple sub-templates chained together
steps:
  # Stage 1: Market research (template composition)
  - name: research_market
    template: web_research_workflow
    template_input:
      market: "{{input_data.market}}"
      geography: "{{input_data.region}}"
      depth: "comprehensive"
    output: research_findings
  
  # Stage 2: Competitor analysis (template composition)
  - name: analyze_competitors
    template: competitive_analysis_workflow
    template_input: "{{research_findings.competitors}}"
    output: competitive_landscape
  
  # Stage 3: Market sizing (template composition)
  - name: size_opportunity
    template: tam_analysis_workflow
    template_input:
      research: "{{research_findings}}"
      competitors: "{{competitive_landscape}}"
    output: market_sizing
  
  # Stage 4: Customer feedback analysis (template composition)
  - name: analyze_customer_pain
    template: customer_feedback_intelligence
    template_input:
      segment: "{{input_data.target_segment}}"
    output: customer_insights
  
  # Stage 5: Generate go-to-market plan
  - name: create_gtm_plan
    prompt: |
      Generate go-to-market strategy:
      
      Market Size: {{market_sizing}}
      Competitors: {{competitive_landscape}}
      Customer Pain Points: {{customer_insights}}
      
      Create:
      - Target customer profile
      - Value proposition
      - Pricing strategy
      - Channel strategy
      - Success metrics
    output: gtm_strategy
  
  # Stage 6: Build executive presentation (template composition)
  - name: create_presentation
    template: presentation_generator_workflow
    template_input:
      title: "Market Entry Analysis: {{input_data.market}}"
      sections:
        - market_sizing
        - competitive_landscape
        - customer_insights
        - gtm_strategy
      template: "market_entry.pptx"
    output: final_deck
```

**This workflow demonstrates:**
- Template composition (6 stages, 4 use sub-templates)
- Multi-source data gathering
- Comprehensive analysis
- Auto-generated deliverable

---

## Next Steps

1. **Review use cases** - Read detailed documentation for each workflow
2. **Set up MCP servers** - Configure analytics, CRM, feedback tools
3. **Test with real data** - Run templates on actual business questions
4. **Schedule recurring reports** - Automate monthly/weekly workflows
5. **Build template library** - Create reusable sub-templates for your org

---

## Additional Resources

- **[MCP Server Integration](../../../mcp-server/README.md)** - Expose templates as tools
- **[Why Templates Matter](../../WHY_TEMPLATES_MATTER.md)** - Template composition explained
- **[Template Authoring Guide](../../authoring-guide.md)** - Create custom BI templates

---

**Business intelligence with AI: Automate recurring reports, synthesize feedback, size markets in hours.**

Templates transform BI from manual data gathering to automated, scheduled, multi-stage workflows through template composition.
