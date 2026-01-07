# Customer Cohort Analyzer

> **Workflow:** [customer_cohort_analyzer.yaml](../workflows/customer_cohort_analyzer.yaml)  
> **Pattern:** Systematic Analysis Pipeline  
> **Best For:** Retention insights and LTV calculations in 8 minutes

---

## Problem Description

### The Retention Black Box

**Without cohort analysis:**

```
CEO: "What's our retention rate?"
Team: "Uh... let me check"
      (3 hours of SQL queries)
      "About 75%?"
CEO: "How does it vary by cohort?"
Team: "...we don't track that"
```

**Problems:**
- No retention visibility
- LTV unknown
- Churn reasons unclear
- Can't optimize

---

## Workflow Solution

### What It Does

**Complete cohort analysis:**

1. **Parse → Retention → LTV → Behavior → Insights**
2. **3 LTV calculation methods**
3. **RFM segmentation**
4. **Retention heatmap**
5. **Action recommendations**

**Value:**
- Time: 3 hours → 8 minutes (96%)
- Insights: Actionable recommendations
- ROI: $25K revenue impact (5% retention improvement)

---

## Usage Example

```bash
./mcp-cli --workflow customer_cohort_analyzer \
  --input-data "$(cat transactions.csv)"
```

**Output: Cohort Analysis Report**

```markdown
# Customer Cohort Analysis

## Summary

**Total Customers:** 10,000
**Average LTV:** $500
**6-Month Retention:** 65%
**Churn Rate:** 35%

## Retention Heatmap

| Cohort | M0 | M1 | M3 | M6 | M12 |
|--------|----|----|----|----|-----|
| Jan 24 | 100% | 75% | 65% | 55% | 45% |
| Feb 24 | 100% | 80% | 70% | 60% | - |
| Mar 24 | 100% | 82% | 72% | - | - |

**Trend:** Retention improving in recent cohorts

## RFM Segments

| Segment | Count | Avg LTV | Strategy |
|---------|-------|---------|----------|
| Champions | 1,200 | $1,200 | Retain & reward |
| At-Risk | 800 | $600 | Win-back campaign |
| Churned | 2,000 | $200 | Reactivate |

## Recommendations

**Improve Retention by 5%:**
- Launch win-back for at-risk (800 customers)
- Expected: 40 customers retained
- Revenue impact: $24,000

**Action Plan:**
1. Email campaign to at-risk customers
2. Special offers for champions
3. Survey churned for feedback
```

---

## Trade-offs

**Advantages:**
- 96% time savings
- 3 LTV methods
- RFM segmentation
- Revenue impact quantified

**Limitations:**
- Requires transaction history
- Projections are estimates
- Human judgment needed

---

## Related Resources

- **[Workflow File](../workflows/customer_cohort_analyzer.yaml)**
- **[Financial Report Generator](financial-report-generator.md)**
- **[Business Metrics Dashboard](business-metrics-dashboard.md)**

---

**Cohort analysis: Know your retention, optimize your growth.**
