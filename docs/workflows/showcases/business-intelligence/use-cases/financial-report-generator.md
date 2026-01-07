# Financial Report Generator

> **Workflow:** [financial_report_generator.yaml](../workflows/financial_report_generator.yaml)  
> **Pattern:** Systematic 5-Step Analysis  
> **Best For:** Executive-ready financial reports in 8 minutes vs 4 hours

---

## Problem Description

### The Manual Reporting Burden

**Monthly financial reporting:**

```
Day 25: Finance team starts month-end
Day 26: Pull data from systems (2 hours)
Day 27: Calculate metrics (4 hours)
Day 28: Create report (3 hours)
Day 29: Review and revise (2 hours)
Day 30: Board meeting

Total: 11 hours of manual work
```

**Problems:**
- Inconsistent calculations
- Manual errors
- Always rushing
- No time for insights

---

## Workflow Solution

### What It Does

**Automated financial analysis:**

1. **Parse → Calculate → Analyze → Insights → Report**
2. **GAAP-aligned calculations**
3. **Benchmark comparisons**
4. **Trend analysis**
5. **Executive summary**

**Value:**
- Time: 4 hours → 8 minutes (97%)
- Consistency: Same every time
- Quality: Professional output

---

## Usage Example

```bash
./mcp-cli --workflow financial_report_generator \
  --input-data "$(cat financials.csv)"
```

**Output: Executive Financial Report**

```markdown
# Financial Report - Q4 2025

## Executive Summary

**Revenue:** $1.2M (↑ 15% YoY)
**Net Income:** $240K (20% margin)
**Cash Position:** $800K
**Status:** HEALTHY

## Key Metrics

| Metric | Value | Target | Status |
|--------|-------|--------|--------|
| Revenue Growth | 15% | 10% | ✓ |
| Gross Margin | 65% | 60% | ✓ |
| Operating Margin | 25% | 20% | ✓ |
| Current Ratio | 2.5 | >1.5 | ✓ |

## Trends

Revenue growing steadily, margins improving,
cash position strong. On track for profitability.

## Action Items

- Maintain current growth trajectory
- Monitor operating expenses
- Consider strategic investments
```

---

## Trade-offs

**Advantages:**
- 97% time savings
- Consistent quality
- Professional format
- Board-ready

**Limitations:**
- Requires clean data
- Human review recommended
- Context may be missing

---

## Related Resources

- **[Workflow File](../workflows/financial_report_generator.yaml)**
- **[Customer Cohort Analyzer](customer-cohort-analyzer.md)**
- **[Business Metrics Dashboard](business-metrics-dashboard.md)**

---

**Automated financial reports: 8 minutes vs 4 hours.**
