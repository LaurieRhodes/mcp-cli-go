# Template Showcases - Documentation Structure

This document defines the structure and standards for industry-specific template showcases.

## Principles

### No Speculative Claims
❌ "Saves 2 hours per incident"
❌ "Reduces costs by 50%"
❌ "10x faster than manual process"

✅ "Template executes in 45 seconds with these 4 steps"
✅ "Costs approximately $0.045 per execution (Claude 3.5 Sonnet)"
✅ "Creates consistent documentation structure across all incidents"

### Focus on What Templates DO
- Demonstrate actual capabilities
- Show real execution flow
- Provide verifiable metrics (execution time, API cost)
- Explain technical behavior

### Educational Approach
Follow the methodology:
1. **Clarity:** Use `{{input_data}}`, explain terms, show examples
2. **Education:** Explain "what happens" and "why"
3. **Practical:** Real commands, performance metrics, actual costs
4. **Completeness:** When to use, trade-offs, troubleshooting

---

## Directory Structure

```
showcases/
├── STRUCTURE.md                    # This file
├── devops/
│   ├── README.md                   # Showcase overview
│   ├── use-cases/
│   │   ├── incident-response.md
│   │   ├── log-analysis.md
│   │   ├── infrastructure-docs.md
│   │   ├── runbook-automation.md
│   │   └── on-call-workflow.md
│   └── templates/
│       ├── incident_response.yaml
│       ├── log_analysis.yaml
│       ├── infra_documentation.yaml
│       ├── runbook_executor.yaml
│       └── on_call_triage.yaml
├── development/
├── data-engineering/
├── security/
├── business-intelligence/
└── content-marketing/
```

---

## Quality Checklist

Before publishing use case documentation, verify:

- [ ] No speculative time/cost savings claims
- [ ] All examples use real, runnable commands
- [ ] Cost calculations based on actual token counts
- [ ] Execution times measured (not estimated)
- [ ] "What happens" explains actual processing
- [ ] Trade-offs honestly assessed
- [ ] Limitations clearly stated
- [ ] Template file tested and working
- [ ] Examples produce documented output
- [ ] Integration examples are realistic
- [ ] Troubleshooting from real issues
- [ ] No marketing language or hype

---

## Metrics to Include

### Always Include (Verifiable)

✅ **Execution Time:** "Template executes in ~45 seconds (4 sequential steps)"
✅ **API Cost:** "Approximately $0.045 per execution (Claude 3.5 Sonnet, 2K input, 1.5K output)"
✅ **Provider/Model:** "Uses Anthropic Claude 3.5 Sonnet for analysis"
✅ **Token Counts:** "Average input: 2000 tokens, output: 1500 tokens"
✅ **Step Count:** "Workflow has 4 steps: analyze → assess → diagnose → report"

### Never Include (Speculative)

❌ "Saves 2 hours compared to manual process"
❌ "Reduces incident documentation time by 80%"
❌ "10x faster than traditional methods"
❌ "ROI of 300% within first month"
❌ "Pays for itself after 5 incidents"

---

## Writing Guidelines

### Voice and Tone

**Educational, not promotional:**
- Explain what templates DO
- Show how they work
- Demonstrate capabilities
- Acknowledge limitations

**Specific, not vague:**
- ✅ "Template parses logs and identifies error patterns"
- ❌ "Template quickly analyzes logs"

**Honest, not hyped:**
- ✅ "Requires human review before publishing"
- ❌ "Produces production-ready reports automatically"

### Language to Avoid

❌ "Revolutionary"
❌ "Game-changing"
❌ "Dramatically improves"
❌ "Effortlessly automates"
❌ "Instantly generates"
❌ "Saves countless hours"

### Language to Use

✅ "Enables consistent analysis"
✅ "Creates structured documentation"
✅ "Automates multi-step workflow"
✅ "Provides starting point for review"
✅ "Codifies best practices"
✅ "Executes in [specific time]"
