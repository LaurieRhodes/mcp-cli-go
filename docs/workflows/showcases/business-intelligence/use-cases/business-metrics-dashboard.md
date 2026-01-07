# Business Metrics Dashboard

> **Workflow:** [business_metrics_dashboard.yaml](../workflows/business_metrics_dashboard.yaml)  
> **Pattern:** Multi-Dimensional Analysis  
> **Best For:** Comprehensive executive dashboard in 7 minutes

---

## Problem Description

### The Scattered Metrics Problem

**Weekly executive dashboard:**

```
Monday: Collect revenue data (1 hour)
Tuesday: Pull customer metrics (1 hour)
Wednesday: Calculate KPIs (1 hour)
Thursday: Create dashboard (1 hour)
Friday: Present to executives

Total: 4 hours every week
```

**Problems:**
- Data scattered across systems
- Manual calculation errors
- Always rush to deadline
- Inconsistent format

---

## Workflow Solution

### What It Does

**Comprehensive metrics analysis:**

1. **Parse → Growth → Efficiency → Health → Dashboard**
2. **Growth metrics** (revenue, customers, ARR)
3. **Efficiency metrics** (CAC, LTV:CAC, margins)
4. **Health metrics** (retention, NPS, burn)
5. **KPI tracking** with targets

**Value:**
- Time: 4 hours → 7 minutes (97%)
- Consistency: Same format every week
- Insights: Automated trend analysis

---

## Usage Example

```bash
./mcp-cli --workflow business_metrics_dashboard \
  --input-data "$(cat business_data.json)"
```

**Output: Executive Dashboard**

```markdown
# Business Metrics Dashboard

**Period:** Q4 2025
**Status:** 🟢 HEALTHY

## Executive Summary

**Revenue:** $1.2M (↑15%)
**Customers:** 5,000 (↑12%)
**Growth Rate:** 15% MoM
**Profitability:** On track

## Growth Scorecard

| Metric | Current | Prior | Change | Target | Status |
|--------|---------|-------|--------|--------|--------|
| Revenue | $1.2M | $1.0M | +15% | +10% | ✓ |
| Customers | 5,000 | 4,500 | +12% | +10% | ✓ |
| ARR | $14.4M | $12.5M | +15% | +12% | ✓ |

## Efficiency Scorecard

| Metric | Value | Benchmark | Status |
|--------|-------|-----------|--------|
| CAC | $100 | <$150 | ✓ |
| LTV:CAC | 4.2 | >3.0 | ✓ |
| Gross Margin | 70% | >65% | ✓ |

## Health Scorecard

| Category | Status | Key Metrics |
|----------|--------|-------------|
| Financial | 🟢 | Margin: 70%, Runway: 24mo |
| Customer | 🟢 | Retention: 85%, NPS: 45 |
| Operational | 🟡 | Efficiency: Good, hiring needed |

## Key Insights

**Positive:**
- Revenue growing faster than target
- Customer acquisition efficient
- Margins healthy and improving

**Areas of Concern:**
- Operational capacity at 90%
- Consider hiring in Q1
- Monitor churn in enterprise segment

## Action Items

**This Week:**
- Review hiring plan for Q1
- Monitor enterprise churn

**This Month:**
- Optimize operational processes
- Prepare for scale

**This Quarter:**
- Strategic hiring
- Infrastructure investments
```

---

## Trade-offs

**Advantages:**
- 97% time savings
- All metrics in one view
- Automated insights
- Professional format

**Limitations:**
- Requires data integration
- Historical data needed for trends
- Human context important

---

## Integration

**Weekly Automation:**
```bash
#!/bin/bash
# Run every Monday at 9 AM
0 9 * * 1 /usr/local/bin/generate-dashboard.sh
```

**Slack Integration:**
```bash
./mcp-cli --workflow business_metrics_dashboard \
  --input-data "$(fetch_metrics.py)" | \
  post_to_slack.sh #executive-dashboard
```

---

## Related Resources

- **[Workflow File](../workflows/business_metrics_dashboard.yaml)**
- **[Financial Report Generator](financial-report-generator.md)**
- **[Customer Cohort Analyzer](customer-cohort-analyzer.md)**

---

**Executive dashboards: 7 minutes vs 4 hours, every week.**
