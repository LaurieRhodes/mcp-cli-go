# Server Examples

Production-ready MCP server configurations demonstrating real-world patterns.

---

## Available Examples

### [Code Analysis Server](code-analysis.md)

**Demonstrates:**
- Multi-tool server (quality, security, improvements)
- Parameter schemas (enums, ranges, defaults)
- Template composition patterns
- Multi-provider optimization
- Integration patterns (git hooks, CI/CD, IDE)

**Tools exposed:**
- `analyze_quality` - Code quality analysis
- `scan_security` - Security vulnerability scanning  
- `suggest_improvements` - Improvement recommendations

**Use this example to learn:**
- How to structure production servers
- Parameter validation patterns
- Tool orchestration
- Cost optimization strategies
- Real-world integration

---

## Creating Your Own

**Use the code analysis example as a template:**

1. **Define your domain** - What workflows do you want to expose?
2. **Create templates** - Build the AI workflows
3. **Map to tools** - Create runas configuration
4. **Test locally** - Verify tool execution
5. **Deploy** - Add to MCP client configs

**Key pattern:**

```yaml
# Your server
name: your_domain
tools:
  - name: your_tool
    template: your_workflow
    parameters: {...}
```

**Each tool:**
- Solves a specific problem
- Has clear parameters
- Maps to a tested template
- Provides value independently

---

## Additional Example Ideas

**Not included but follow same pattern:**

### Research Server
```yaml
tools:
  - name: research_topic
    template: multi_source_research
  - name: verify_claims
    template: fact_checker
  - name: synthesize_findings
    template: research_synthesis
```

### Data Analysis Server
```yaml
tools:
  - name: analyze_dataset
    template: statistical_analyzer
  - name: detect_anomalies
    template: anomaly_detector
  - name: forecast
    template: time_series_forecaster
```

### Documentation Server
```yaml
tools:
  - name: generate_api_docs
    template: api_doc_generator
  - name: create_guide
    template: guide_writer
  - name: explain_code
    template: code_explainer
```

**The pattern is universal - domain knowledge → templates → tools.**

---

## See Also

- **[runas Configuration](../runas-config.md)** - Complete specification
- **[Integration Guide](../integration.md)** - Deployment patterns
- **[Template Documentation](../../templates/)** - Building workflows
