# Automation & Scripting Guide

Use AI in your scripts, CI/CD, and scheduled tasks - the right way.

**What is automation?** Running AI queries automatically without human intervention.

**Examples:**
- Code review on every pull request (CI/CD)
- Daily report at 5 PM every day (cron job)
- Analyze logs every hour (scheduled task)
- Process files in batch (scripts)

**The Right Way:**
- ✅ **Templates** - Define workflows once, version control them, reuse everywhere
- ✅ **Configuration** - Proper setup with providers and servers
- ✅ **Scripts** - Simple glue code that calls templates

**The Wrong Way:**
- ❌ Hardcode AI prompts in bash scripts (unmaintainable)
- ❌ Use query mode for everything (no reusability)
- ❌ Skip configuration setup (breaks in production)

**Key insight:** Separate concerns!
- **Templates** = AI logic (what to do)
- **Scripts** = Orchestration (when to do it)
- **Config** = Settings (which AI, which tools)

---

## Table of Contents

- [Quick Start Example](#quick-start-example)
- [Setup for Automation](#setup-for-automation)
- [Template-Based Automation](#template-based-automation)
- [CI/CD Integration](#cicd-integration)
- [Scheduled Tasks](#scheduled-tasks)
- [When to Use Query Mode](#when-to-use-query-mode)
- [Best Practices](#best-practices)

---

## Quick Start Example

**See automation in action:**

**1. Create a template:**
```yaml
# config/templates/analyze-code.yaml
name: analyze_code
version: 1.0.0
steps:
  - name: analyze
    prompt: "Review this code for bugs: {{stdin}}"
```

**2. Use in script:**
```bash
#!/bin/bash
# review.sh
cat main.go | mcp-cli --template analyze_code > review.txt
echo "Review saved to review.txt"
```

**3. Run it:**
```bash
chmod +x review.sh
./review.sh
```

**What happens:**
1. Script reads `main.go`
2. Passes content to `analyze_code` template
3. Template sends to AI for review
4. Review saved to `review.txt`

**Cost:** ~$0.001 per file (1/10th of a cent)

**Why this is better than:**
```bash
# DON'T DO THIS - prompt hardcoded in script!
cat main.go | mcp-cli query "Review this code for bugs..."
```

**Problems with above:**
- ❌ Prompt is in bash script (hard to version/test)
- ❌ Can't reuse across projects
- ❌ If you change prompt, must update all scripts
- ❌ No multi-step workflows

---

## Setup for Automation

**Before automating, setup once:**

### Initial Setup Script

**Run this once to set up automation project:**

```bash
#!/bin/bash
# setup-automation.sh - One-time setup

set -euo pipefail

echo "Setting up MCP-CLI automation..."

# 1. Create project directory
PROJECT_DIR="${1:-mcp-automation}"
mkdir -p "$PROJECT_DIR"
cd "$PROJECT_DIR"

echo "✓ Created directory: $PROJECT_DIR"

# 2. Download MCP-CLI binary
echo "Downloading MCP-CLI..."
wget -q https://github.com/LaurieRhodes/mcp-cli-go/releases/latest/download/mcp-cli-linux-amd64
chmod +x mcp-cli-linux-amd64
mv mcp-cli-linux-amd64 mcp-cli

echo "✓ Downloaded mcp-cli"

# 3. Initialize configuration
echo "Initializing configuration..."
./mcp-cli init --quick

echo "✓ Created config/"

# 4. Set up API keys
echo "Setting up API keys..."
cat > .env << EOF
# Add your API keys here
ANTHROPIC_API_KEY=${ANTHROPIC_API_KEY:-your_key_here}
OPENAI_API_KEY=${OPENAI_API_KEY:-your_key_here}
EOF

echo "✓ Created .env (update with real keys!)"

# 5. Create scripts directory
mkdir -p scripts

# 6. Create example template
mkdir -p config/templates
cat > config/templates/example.yaml << 'EOF'
name: example
description: Example automation template
version: 1.0.0

steps:
  - name: process
    prompt: "Process this: {{stdin}}"
EOF

echo "✓ Created example template"

# 7. Test setup
echo "Testing setup..."
if ./mcp-cli query "Test" >/dev/null 2>&1; then
    echo "✓ MCP-CLI working!"
else
    echo "⚠ Warning: Test query failed. Check API keys in .env"
fi

# 8. Create .gitignore
cat > .gitignore << 'EOF'
.env
config.yaml
config/providers/
mcp-cli
EOF

echo "✓ Created .gitignore"

echo ""
echo "Setup complete! 🎉"
echo ""
echo "Next steps:"
echo "1. Edit .env with your real API keys"
echo "2. Create templates in config/templates/"
echo "3. Create automation scripts in scripts/"
echo ""
echo "Try: ./mcp-cli --template example"
```

**Run it:**
```bash
chmod +x setup-automation.sh
./setup-automation.sh
```

**What this creates:**
```
mcp-automation/
├── mcp-cli                    # ✓ Binary ready to use
├── .env                       # ✓ For API keys (update!)
├── .gitignore                 # ✓ Don't commit secrets
├── config.yaml                # ✓ Main config
├── config/
│   ├── settings.yaml          # ✓ Global settings
│   ├── providers/             # ✓ AI provider configs
│   └── templates/             # ✓ Your workflows here!
│       └── example.yaml
└── scripts/                   # ✓ Your automation scripts
```

**Time:** 30 seconds

**Now you're ready to automate!**

---

### Directory Structure Best Practices

**Recommended layout:**

```
my-automation-project/
├── mcp-cli                    # Binary (not in git)
├── .env                       # Secrets (not in git)
├── .gitignore                 # Excludes above
├── config.yaml                # Main config (git if no secrets)
├── config/
│   ├── settings.yaml          # ✓ In git
│   └── templates/             # ✓ In git (this is your code!)
│       ├── code-review.yaml
│       ├── daily-report.yaml
│       └── analyze-logs.yaml
├── scripts/                   # ✓ In git
│   ├── review-pr.sh          # Calls templates
│   ├── daily-report.sh       # Calls templates
│   └── monitor.sh            # Calls templates
└── README.md                  # ✓ In git
```

**What goes in git:**
- ✅ Templates (config/templates/)
- ✅ Scripts (scripts/)
- ✅ README, documentation
- ✅ .gitignore

**What DOESN'T go in git:**
- ❌ .env (has API keys!)
- ❌ mcp-cli binary (download from releases)
- ❌ config.yaml if it has embedded secrets
- ❌ config/providers/ if keys embedded (use .env instead)

**Key principle:** Templates and scripts are code. Binaries and secrets are not.

---

## Template-Based Automation

### Example 1: Code Review Template

**What it does:** Automated code review on every git commit or pull request.

**Create the template:**

```yaml
# config/templates/code-review.yaml
name: code_review
description: Automated code review for quality and security
version: 1.0.0

config:
  defaults:
    provider: openai
    model: gpt-4o
    temperature: 0.3  # Lower = more consistent

steps:
  - name: analyze_changes
    prompt: |
      Review these code changes:
      {{stdin}}
      
      Check for:
      - Logic errors and bugs
      - Security vulnerabilities
      - Code quality issues
      - Best practice violations
      - Performance concerns
      
      Rate severity: CRITICAL, HIGH, MEDIUM, LOW
    output: review

  - name: format_report
    prompt: |
      Format this review as markdown with:
      - Summary (1-2 sentences)
      - Issues found (grouped by severity)
      - Recommendations
      
      Review: {{review}}
    output: report
```

**Create the automation script:**

```bash
#!/bin/bash
# scripts/review-pr.sh

set -euo pipefail

cd "$(dirname "$0")/.."  # Go to project root

# Get code changes
git diff origin/main...HEAD > /tmp/changes.diff

# Run template
./mcp-cli --template code_review < /tmp/changes.diff > /tmp/review.md

# Display results
cat /tmp/review.md

# Exit with error if critical issues found
if grep -qi "CRITICAL" /tmp/review.md; then
    echo ""
    echo "❌ CRITICAL issues found! Review required."
    exit 1
fi

echo "✓ No critical issues"
```

**Usage:**
```bash
# Manually
./scripts/review-pr.sh

# In git hook (.git/hooks/pre-push)
#!/bin/bash
./scripts/review-pr.sh || exit 1

# In CI (see CI/CD section below)
```

**What happens:**
1. Gets code changes (git diff)
2. Passes to code_review template
3. AI analyzes for issues
4. Report generated as markdown
5. Script fails if critical issues found

**Cost:** ~$0.02 per review (depends on code size)

**Benefits:**
- ✅ Catches bugs before merge
- ✅ Enforces code quality
- ✅ Saves review time
- ✅ Consistent standards

---

### Example 2: Daily Report Template

**What it does:** Generates daily development summary (commits, issues, progress).

**Create the template:**

```yaml
# config/templates/daily-report.yaml
name: daily_report
description: Generate daily development report
version: 1.0.0

config:
  defaults:
    provider: anthropic
    model: claude-sonnet-4

steps:
  - name: summarize_commits
    prompt: |
      Summarize today's commits in 2-3 sentences:
      {{input_data.commits}}
      
      Focus on: what was built, what was fixed
    output: commit_summary

  - name: summarize_issues
    prompt: |
      Summarize open issues (highlight urgent ones):
      {{input_data.issues}}
    output: issue_summary

  - name: create_report
    prompt: |
      Create daily report in markdown:
      
      # Daily Development Report
      ## Date: {{input_data.date}}
      
      ## Commits Today
      {{commit_summary}}
      
      ## Open Issues  
      {{issue_summary}}
      
      ## Metrics
      - Commits: {{input_data.commit_count}}
      - Issues closed: {{input_data.closed_count}}
      - Issues opened: {{input_data.opened_count}}
```

**Create the automation script:**

```bash
#!/bin/bash
# scripts/daily-report.sh

set -euo pipefail

cd /path/to/your/project

# Collect data
DATE=$(date +%Y-%m-%d)
COMMITS=$(git log --since="24 hours ago" --pretty=format:"%h %s" | head -20)
COMMIT_COUNT=$(git log --since="24 hours ago" --oneline | wc -l)
ISSUES=$(gh issue list --limit 10 2>/dev/null || echo "No GitHub CLI")
CLOSED_COUNT=$(gh issue list --state closed --search "closed:>=$(date -d '1 day ago' +%Y-%m-%d)" 2>/dev/null | wc -l || echo 0)
OPENED_COUNT=$(gh issue list --search "created:>=$(date -d '1 day ago' +%Y-%m-%d)" 2>/dev/null | wc -l || echo 0)

# Create JSON input
INPUT_JSON=$(cat <<EOF
{
  "date": "$DATE",
  "commits": "$COMMITS",
  "commit_count": $COMMIT_COUNT,
  "issues": "$ISSUES",
  "closed_count": $CLOSED_COUNT,
  "opened_count": $OPENED_COUNT
}
EOF
)

# Run template
./mcp-cli --template daily_report \
  --input-data "$INPUT_JSON" \
  > reports/daily-$DATE.md

echo "✓ Report generated: reports/daily-$DATE.md"

# Email it (optional)
if command -v mail >/dev/null; then
    mail -s "Daily Report - $DATE" \
      team@example.com < reports/daily-$DATE.md
fi
```

**Set up as cron job:**
```bash
# Edit crontab
crontab -e

# Add line (runs daily at 5 PM)
0 17 * * * /path/to/scripts/daily-report.sh
```

**What happens:**
1. Collects git commits from last 24 hours
2. Gets open issues from GitHub
3. Calculates metrics
4. AI generates human-readable summary
5. Saves to reports/daily-YYYY-MM-DD.md
6. Optionally emails team

**Cost:** ~$0.05 per report

**Example output:**
```markdown
# Daily Development Report
## Date: 2024-12-26

## Commits Today
Today's work focused on adding authentication middleware and fixing 
database connection pooling. The team also updated API documentation 
and resolved several security warnings.

## Open Issues
5 issues remain open, with 2 marked urgent:
- #142: Memory leak in background worker (urgent)
- #156: Login timeout on slow connections (urgent)
- #160-162: Documentation improvements

## Metrics
- Commits: 12
- Issues closed: 3
- Issues opened: 2
```

---

### Example 3: Multi-Provider Analysis

**What it does:** Uses multiple AI providers for high-confidence results.

**Why:** Different AIs catch different issues. Using 2-3 increases accuracy.

**Create the template:**

```yaml
# config/templates/thorough-analysis.yaml
name: thorough_analysis
description: Multi-provider analysis for critical decisions
version: 1.0.0

steps:
  # Claude analyzes
  - name: claude_analysis
    provider: anthropic
    model: claude-sonnet-4
    prompt: |
      Analyze this thoroughly:
      {{stdin}}
      
      Focus on: risks, implications, recommendations
    output: claude_view

  # GPT-4 analyzes (different perspective)
  - name: gpt_analysis
    provider: openai
    model: gpt-4o
    prompt: |
      Analyze this thoroughly:
      {{stdin}}
      
      Focus on: risks, implications, recommendations
    output: gpt_view

  # Local model synthesizes (free!)
  - name: synthesize
    provider: ollama
    model: qwen2.5:32b
    prompt: |
      Synthesize these two analyses:
      
      Claude's analysis:
      {{claude_view}}
      
      GPT-4's analysis:
      {{gpt_view}}
      
      Provide:
      - Points where both agree (HIGH CONFIDENCE)
      - Points where they differ (NEEDS VERIFICATION)
      - Balanced conclusion
    output: final_analysis
```

**Usage:**
```bash
# Analyze security proposal
cat security-proposal.txt | \
  ./mcp-cli --template thorough_analysis \
  > analysis.md
```

**Cost:** ~$0.15 per analysis (2 paid + 1 free)

**When to use:**
- ✅ Security decisions
- ✅ Architecture choices
- ✅ High-stakes code reviews
- ✅ Production deployments

**When not worth it:**
- ❌ Routine reviews (use single provider)
- ❌ Low-stakes decisions
- ❌ Cost-sensitive (3x more expensive)

---

## CI/CD Integration

### GitHub Actions - Proper Setup

```yaml
# .github/workflows/ai-review.yml
name: AI Code Review

on:
  pull_request:
    types: [opened, synchronize]

jobs:
  review:
    runs-on: ubuntu-latest
    
    steps:
      - uses: actions/checkout@v3
      
      # 1. Install MCP-CLI
      - name: Install MCP-CLI
        run: |
          wget https://github.com/LaurieRhodes/mcp-cli-go/releases/latest/download/mcp-cli-linux-amd64
          chmod +x mcp-cli-linux-amd64
          sudo mv mcp-cli-linux-amd64 /usr/local/bin/mcp-cli
      
      # 2. Initialize config
      - name: Setup Config
        env:
          OPENAI_API_KEY: ${{ secrets.OPENAI_API_KEY }}
        run: |
          mcp-cli init --quick
          
          # Add API key to .env
          echo "OPENAI_API_KEY=$OPENAI_API_KEY" > .env
          
          # Verify setup
          mcp-cli query "Test"
      
      # 3. Add code-review template
      - name: Add Review Template
        run: |
          cat > config/templates/code-review.yaml << 'EOF'
          name: code_review
          version: 1.0.0
          config:
            defaults:
              provider: openai
              model: gpt-4o
          steps:
            - name: review
              prompt: |
                Review these changes for:
                - Logic errors
                - Security issues  
                - Best practices
                
                Changes:
                {{stdin}}
          EOF
      
      # 4. Run template
      - name: Review Changes
        run: |
          git diff origin/main...HEAD > changes.diff
          mcp-cli --template code_review < changes.diff > review.md
      
      # 5. Post review
      - name: Post Review
        uses: actions/github-script@v6
        with:
          script: |
            const fs = require('fs');
            const review = fs.readFileSync('review.md', 'utf8');
            
            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: '## AI Code Review\n\n' + review
            });
```

**Key improvements:**
- ✅ Proper initialization with `init`
- ✅ Template defined in workflow
- ✅ Clean separation of concerns

### GitLab CI - Template Approach

```yaml
# .gitlab-ci.yml
stages:
  - setup
  - analyze

variables:
  MCP_VERSION: "latest"

setup:
  stage: setup
  script:
    # Install
    - wget https://github.com/LaurieRhodes/mcp-cli-go/releases/$MCP_VERSION/download/mcp-cli-linux-amd64
    - chmod +x mcp-cli-linux-amd64
    - mv mcp-cli-linux-amd64 /usr/local/bin/mcp-cli
    
    # Initialize
    - mcp-cli init --quick
    - echo "ANTHROPIC_API_KEY=$ANTHROPIC_API_KEY" > .env
    
    # Verify
    - mcp-cli query "Test"
  artifacts:
    paths:
      - config/
      - .env

analyze:
  stage: analyze
  dependencies:
    - setup
  script:
    # Create analysis template
    - |
      cat > config/templates/commit-analysis.yaml << 'EOF'
      name: commit_analysis
      version: 1.0.0
      steps:
        - name: analyze
          prompt: "Analyze commit: {{stdin}}"
      EOF
    
    # Run analysis
    - echo "$CI_COMMIT_MESSAGE" | mcp-cli --template commit_analysis > analysis.txt
  artifacts:
    paths:
      - analysis.txt
```

### Docker-Based CI

**Dockerfile:**
```dockerfile
FROM ubuntu:22.04

# Install MCP-CLI
RUN apt-get update && apt-get install -y wget && \
    wget https://github.com/LaurieRhodes/mcp-cli-go/releases/latest/download/mcp-cli-linux-amd64 && \
    chmod +x mcp-cli-linux-amd64 && \
    mv mcp-cli-linux-amd64 /usr/local/bin/mcp-cli

# Initialize config
WORKDIR /workspace
RUN mcp-cli init --quick

# Copy templates
COPY config/templates/ /workspace/config/templates/

# API keys via env vars at runtime
ENV ANTHROPIC_API_KEY=""

ENTRYPOINT ["mcp-cli"]
```

**Use in CI:**
```bash
# Build once
docker build -t mcp-automation .

# Use in CI
docker run -e ANTHROPIC_API_KEY=$KEY \
  mcp-automation --template code_review < changes.diff
```

---

## Scheduled Tasks

### Cron Job Template

```bash
#!/bin/bash
# /etc/cron.d/mcp-automation

# Daily report at 5 PM
0 17 * * * user /opt/mcp-automation/daily-report.sh

# Weekly summary on Sundays
0 8 * * 0 user /opt/mcp-automation/weekly-summary.sh
```

### Daily Report Script

```bash
#!/bin/bash
# /opt/mcp-automation/daily-report.sh

cd /opt/mcp-automation

# Config should already exist from setup
# Just run the template
REPORT_DATE=$(date +%Y-%m-%d)

./mcp-cli --template daily_report \
  --input-data "{\"date\": \"$REPORT_DATE\"}" \
  > reports/daily-$REPORT_DATE.md

# Email it
mail -s "Daily Report - $REPORT_DATE" \
  team@example.com < reports/daily-$REPORT_DATE.md
```

### Monitoring with Templates

**Template:**
```yaml
# config/templates/system-health.yaml
name: system_health
description: Analyze system health metrics
version: 1.0.0

steps:
  - name: analyze_metrics
    prompt: |
      Analyze system health and flag issues:
      
      CPU: {{cpu}}
      Memory: {{memory}}
      Disk: {{disk}}
      
      Alert if critical.
    output: analysis

  - name: recommend
    condition: "{{analysis}} contains 'CRITICAL'"
    prompt: |
      Provide immediate recommendations:
      {{analysis}}
```

**Script:**
```bash
#!/bin/bash
# /opt/mcp-automation/monitor.sh

cd /opt/mcp-automation

# Gather metrics
CPU=$(top -bn1 | grep "Cpu(s)" | awk '{print $2}')
MEMORY=$(free -h | awk '/^Mem:/ {print $3 "/" $2}')
DISK=$(df -h / | awk 'NR==2 {print $5}')

# Run template
RESULT=$(./mcp-cli --template system_health \
  --input-data "{\"cpu\": \"$CPU\", \"memory\": \"$MEMORY\", \"disk\": \"$DISK\"}")

# Alert if critical
if echo "$RESULT" | grep -qi "CRITICAL"; then
    echo "ALERT: $RESULT" | mail -s "CRITICAL: System Health" admin@example.com
fi
```

---

## When to Use Query Mode

**Query mode is for:**
- ✅ One-off questions in scripts
- ✅ Quick tests during development
- ✅ True ad-hoc queries

**Example - One-off question:**

```bash
#!/bin/bash
# Quick check in a script

# Config already exists
STATUS=$(./mcp-cli query "Is this error critical: $ERROR_MSG")

if echo "$STATUS" | grep -qi "yes"; then
    alert_team
fi
```

**But if you do this more than once → Make it a template!**

**Better approach:**

```yaml
# config/templates/error-triage.yaml
name: error_triage
version: 1.0.0
steps:
  - name: classify
    prompt: "Is this error critical? Answer yes or no: {{error}}"
```

```bash
#!/bin/bash
# Now reusable!

RESULT=$(echo "$ERROR_MSG" | ./mcp-cli --template error_triage)
```

---

## Best Practices

### 1. Version Control Templates

```bash
# Project structure
my-project/
├── config/
│   └── templates/          # ← In git!
│       ├── review.yaml
│       ├── analyze.yaml
│       └── report.yaml
├── scripts/
│   ├── ci-review.sh       # ← Calls templates
│   └── daily-report.sh    # ← Calls templates
└── .gitignore             # Exclude .env, config.yaml if has keys
```

**.gitignore:**
```
.env
config.yaml          # If it has embedded secrets
config/providers/    # If providers have embedded keys

# Keep these:
# config/templates/
```

### 2. Initialize Once, Use Everywhere

**Setup script (run once):**

```bash
#!/bin/bash
# setup-automation.sh

./mcp-cli init --quick

# Configure providers
cat > config/providers/anthropic.yaml << EOF
provider_name: anthropic
interface_type: anthropic_native
config:
  api_key: \${ANTHROPIC_API_KEY}
  default_model: claude-sonnet-4
EOF

# Add to version control
git add config/templates/
git commit -m "Add automation templates"
```

**Automation scripts (run many times):**

```bash
#!/bin/bash
# review.sh - Just uses config

./mcp-cli --template code_review < changes.diff
```

### 3. Template Composition

**Build complex workflows from simple templates:**

```yaml
# config/templates/complete-review.yaml
name: complete_review
version: 1.0.0

steps:
  # Call security review template
  - name: security
    template: security_review
    template_input: "{{stdin}}"
    output: security_result

  # Call quality review template
  - name: quality
    template: quality_review
    template_input: "{{stdin}}"
    output: quality_result

  # Synthesize
  - name: final_report
    prompt: |
      Combine reviews:
      Security: {{security_result}}
      Quality: {{quality_result}}
```

### 4. Environment-Specific Config

```bash
# Production
./mcp-cli --config prod-config.yaml --template deploy_check

# Staging
./mcp-cli --config staging-config.yaml --template deploy_check

# Development
./mcp-cli --config dev-config.yaml --template deploy_check
```

### 5. Error Handling

```bash
#!/bin/bash
set -euo pipefail

# Function with error handling
run_template() {
    local template=$1
    local input=$2
    
    if ! result=$(echo "$input" | ./mcp-cli --template "$template" 2>&1); then
        echo "Error running template $template: $result" >&2
        return 1
    fi
    
    echo "$result"
}

# Use it
if ! REVIEW=$(run_template "code_review" "$CODE"); then
    exit 1
fi
```

---

## Migration from Query Mode to Templates

**If you have this:**

```bash
# BAD: Hardcoded prompts in scripts
analyze() {
    ./mcp-cli query "Analyze this code for errors: $1"
}
```

**Convert to:**

```yaml
# config/templates/code-analyze.yaml
name: code_analyze
version: 1.0.0
steps:
  - name: analyze
    prompt: "Analyze this code for errors: {{stdin}}"
```

```bash
# GOOD: Reusable template
analyze() {
    echo "$1" | ./mcp-cli --template code_analyze
}
```

**Benefits:**
- ✅ Prompt changes don't require script changes
- ✅ Can version/test prompt separately
- ✅ Reusable across projects
- ✅ Can add steps without touching scripts

---

## Quick Reference

### Setup
```bash
# One-time setup
mcp-cli init --quick
echo "ANTHROPIC_API_KEY=..." > .env

# Verify
mcp-cli query "Test"
```

### Create Template
```yaml
# config/templates/my-workflow.yaml
name: my_workflow
version: 1.0.0
steps:
  - name: step1
    prompt: "..."
```

### Use Template
```bash
# In scripts
./mcp-cli --template my_workflow < input.txt > output.txt

# In CI/CD
mcp-cli --template my_workflow --input-data "$DATA"
```

### Query Mode (sparingly)
```bash
# Only for true one-offs
./mcp-cli query "Is this critical: $ERROR"
```

---

## Next Steps

- **[Template Authoring](../templates/authoring-guide.md)** - Learn template creation
- **[Template Examples](../templates/examples/)** - Working examples
- **[Query Mode](query-mode.md)** - When to use query mode
- **[Debugging](debugging.md)** - Troubleshoot automation

---

**Remember:** Templates are the automation layer, not scripts with embedded prompts! 🚀
