# Template Examples

Working template examples demonstrating real-world use cases.

---

## Available Examples

### Basic Templates

- **[hello.yaml](hello.yaml)** - Simple greeting template
- **[summarize.yaml](summarize.yaml)** - Document summarization
- **[translate.yaml](translate.yaml)** - Multi-step translation

### Code & Development

- **[code-review.yaml](code-review.yaml)** - Comprehensive code review
- **[bug-analysis.yaml](bug-analysis.yaml)** - Bug report analysis
- **[docs-generator.yaml](docs-generator.yaml)** - Documentation generation

### Document Processing

- **[extract-summary.yaml](extract-summary.yaml)** - Extract and summarize
- **[report-generator.yaml](report-generator.yaml)** - Multi-section reports
- **[content-analysis.yaml](content-analysis.yaml)** - Content classification

### Advanced Workflows

- **[parallel-analysis.yaml](parallel-analysis.yaml)** - Parallel execution
- **[loop-processing.yaml](loop-processing.yaml)** - Array processing
- **[multi-provider.yaml](multi-provider.yaml)** - Multiple AI providers
- **[composed-workflow.yaml](composed-workflow.yaml)** - Template composition

### Automation

- **[daily-report.yaml](daily-report.yaml)** - Daily development report
- **[ci-review.yaml](ci-review.yaml)** - CI/CD code review
- **[batch-processing.yaml](batch-processing.yaml)** - Batch file processing

---

## Using Examples

### Copy to Your Config

```bash
# Copy example to your templates directory
cp docs/templates/examples/code-review.yaml config/templates/

# Use it
mcp-cli --template code_review --input-data '{"code": "..."}'
```

### Test Examples

```bash
# Test with sample data
echo '{"name": "Alice"}' | mcp-cli --template hello

# Test with file
mcp-cli --template summarize --input-file document.txt
```

### Modify Examples

```bash
# Copy and customize
cp docs/templates/examples/code-review.yaml config/templates/my-review.yaml

# Edit for your needs
vim config/templates/my-review.yaml

# Use custom version
mcp-cli --template my_review
```

---

## Example Categories

### 🟢 **Beginner**
Simple templates for learning:
- hello.yaml
- summarize.yaml
- translate.yaml

### 🟡 **Intermediate**
Real-world applications:
- code-review.yaml
- docs-generator.yaml
- report-generator.yaml

### 🔴 **Advanced**
Complex workflows:
- parallel-analysis.yaml
- multi-provider.yaml
- composed-workflow.yaml

---

## Creating Your Own

### Start Simple

```yaml
name: my_template
version: 1.0.0

steps:
  - name: process
    prompt: "Process: {{input}}"
```

### Add Features Gradually

1. Start with one step
2. Add outputs and variables
3. Add conditions or loops
4. Compose with other templates
5. Add error handling

### Test Frequently

```bash
# Test after each change
mcp-cli --template my_template --input-data '{"test": "data"}'
```

---

## Contributing Examples

Have a useful template? Share it!

1. Test thoroughly
2. Add clear documentation
3. Submit PR to examples/
4. Share in [Discussions](https://github.com/LaurieRhodes/mcp-cli-go/discussions)

---

## Quick Reference

```bash
# Copy example
cp docs/templates/examples/NAME.yaml config/templates/

# Use example
mcp-cli --template NAME --input-data '{...}'

# Customize
vim config/templates/NAME.yaml
```

---

## All Examples

Browse all available templates below.
