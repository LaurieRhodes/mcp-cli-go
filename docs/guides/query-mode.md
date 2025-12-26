# Query Mode Guide

Run single AI queries from the command line - perfect for scripts, automation, and CI/CD.

**What is Query Mode?** Ask AI a question, get an answer, done. No conversation, no interaction - just one question → one answer.

**Use when:**

- Writing shell scripts
- Automating tasks (cron jobs, CI/CD)
- Processing files in batch
- Need clean, parseable output
- No conversation needed

**Don't use when:**

- Want conversation back-and-forth
- Building on previous context
- Interactive exploration
  → Use [Chat Mode](chat-mode.md) instead

---

## Table of Contents

- [Quick Start](#quick-start)
- [Basic Usage](#basic-usage)
- [Command-Line Options](#command-line-options)
- [Output Formats](#output-formats)
- [Scripting Patterns](#scripting-patterns)
- [Advanced Features](#advanced-features)
- [Error Handling](#error-handling)
- [Best Practices](#best-practices)

---

## Quick Start

**Most basic query:**

```bash
mcp-cli query "What is 2+2?"
```

**Expected output:**

```
The answer is 4.
```

**What happens:**

1. MCP-CLI sends "What is 2+2?" to default AI provider
2. AI generates response
3. Response printed to stdout
4. Program exits (exit code 0 = success)

**That's it!** No conversation history saved, no state kept.

---

## Basic Usage

### Simple Query

```bash
mcp-cli query "What is the capital of France?"
```

**Output:**

```
The capital of France is Paris.
```

---

### Query with Specific Provider

**Why:** Different AIs are better at different things.

```bash
# Use Claude (best for analysis)
mcp-cli query --provider anthropic "Analyze this business proposal"

# Use GPT-4 (best for code)
mcp-cli query --provider openai "Write a Python function for binary search"

# Use local model (free!)
mcp-cli query --provider ollama "What's the weather like?"
```

**What `--provider` does:** Overrides your default AI provider for this one query.

---

### Query with Specific Model

**Why:** Newer/bigger models are smarter but cost more.

```bash
# Use latest Claude model
mcp-cli query --provider anthropic --model claude-sonnet-4 "Complex question"

# Use cheaper GPT model
mcp-cli query --provider openai --model gpt-4o-mini "Simple question"

# Use specific Ollama model
mcp-cli query --provider ollama --model llama3.2 "Question"
```

**Cost comparison:**

- `gpt-4o`: $0.01 per 1K words (expensive, very smart)
- `gpt-4o-mini`: $0.0002 per 1K words (cheap, still good)
- `ollama`: $0 (free, runs on your computer)

---

### Query with Tools (MCP Servers)

**What are tools?** Give AI ability to read files, search web, etc.

```bash
# AI can read files
mcp-cli query --server filesystem "What files are in this directory?"
```

**Expected output:**

```
The current directory contains:
1. README.md - Documentation file
2. main.go - Main program file
3. config.yaml - Configuration file
4. .env - Environment variables
```

**What happens:**

1. AI receives your question
2. AI decides "I need to list files" → uses filesystem tool
3. Filesystem tool executes, returns file list
4. AI formats the list nicely
5. You get formatted response

```bash
# AI can search web
mcp-cli query --server brave-search "What's the latest news on AI?"
```

**Multiple tools:**

```bash
# AI can use both filesystem AND web search
mcp-cli query --server filesystem,brave-search \
    "Search for MCP documentation and save to file"
```



---

## Command-Line Options

### Provider Options

```bash
# Specific provider
--provider anthropic

# Specific model
--model claude-sonnet-4

# Both
--provider openai --model gpt-4o
```

### Server Options

```bash
# Single server
--server filesystem

# Multiple servers
--server filesystem,brave-search

# Disable filesystem (when auto-loaded)
--disable-filesystem
```

### Output Options

```bash
# JSON output
--json

# Output to file
--output result.txt

# JSON to file
--json --output result.json

# Raw tool data (bypass AI summarization)
--raw-data
```

### Context Options

```bash
# Add context from file
--context background.txt

# Custom system prompt
--system-prompt "You are a senior developer"

# Max response tokens
--max-tokens 1000
```

### Verbosity Options

```bash
# Show detailed logs
--noisy

# Debug mode
--verbose

# Quiet (default)
# (suppresses connection messages)
```

### Error Handling

```bash
# Exit codes only (no error messages)
--error-code-only
```

---

## Output Formats

### Plain Text (Default)

```bash
mcp-cli query "List 3 colors"
```

**Output:**

```
Here are three colors:
1. Red
2. Blue
3. Green
```

### JSON Output

```bash
mcp-cli query --json "List 3 colors"
```

**Output:**

```json
{
  "response": "Here are three colors:\n1. Red\n2. Blue\n3. Green",
  "tool_calls": [],
  "server_connections": [],
  "execution_time_ms": 1234
}
```

### Raw Tool Data

```bash
mcp-cli query --raw-data --server filesystem "What files are here?"
```

**Output:**

```
RAW TOOL DATA:
------------------------

Tool Call #1: list_directory
Result:
  files: 
    [0]: README.md
    [1]: main.go
    [2]: config.yaml
```

**Use when:**

- You want tool output without AI summarization
- Parsing structured data
- Debugging tool calls

---

## Scripting Patterns

Query mode shines in scripts! Here are proven patterns.

---

### Pattern 1: Capture AI Response in Variable

**What it does:** Store AI's answer in a bash variable for later use.

```bash
#!/bin/bash

# Ask AI, store answer in variable
ANSWER=$(mcp-cli query "What is the capital of France?")

# Use the answer
echo "The answer is: $ANSWER"

# Use in conditional logic
if [[ "$ANSWER" =~ "Paris" ]]; then
    echo "✓ Correct!"
else
    echo "✗ Wrong answer"
fi
```

**What happens:**

1. `$(...)` runs command and captures output
2. Output stored in `ANSWER` variable
3. Can use `$ANSWER` anywhere in script

**Real-world use:**

```bash
# Analyze git commits
COMMIT_MSG=$(git log -1 --pretty=%B)
SENTIMENT=$(mcp-cli query "Rate sentiment 1-10: $COMMIT_MSG")

if [[ "$SENTIMENT" =~ [7-9]|10 ]]; then
    echo "Good commit message!"
fi
```

---

### Pattern 2: Process Piped Input

**What it does:** Feed data to AI from pipes or files.

```bash
# Process a file
cat document.txt | mcp-cli query "Summarize this document"

# Process command output  
git log --oneline -10 | mcp-cli query "Summarize these commits"

# Process and transform
cat data.txt | \
    mcp-cli query "Extract all email addresses" | \
    sort | \
    uniq > emails.txt
```

**What happens:**

1. `cat document.txt` outputs file contents
2. `|` (pipe) sends output to mcp-cli as stdin
3. AI receives: "Summarize this document" + file contents
4. Response printed to stdout

**Real-world use:**

```bash
#!/bin/bash
# Analyze log files for errors

cat /var/log/app.log | \
    grep ERROR | \
    mcp-cli query "Categorize these errors and suggest fixes" \
    > error-analysis.txt

echo "Analysis saved to error-analysis.txt"
```

---

### Pattern 3: Parse JSON Output

**What it does:** Get structured data you can parse programmatically.

```bash
#!/bin/bash

# Get JSON response
RESULT=$(mcp-cli query --json "List top 3 programming languages")

# Parse response field
RESPONSE=$(echo "$RESULT" | jq -r '.response')
echo "$RESPONSE"

# Check for errors
ERROR=$(echo "$RESULT" | jq -r '.error // empty')
if [ -n "$ERROR" ]; then
    echo "Error occurred: $ERROR" >&2
    exit 1
fi

# Get execution time
TIME_MS=$(echo "$RESULT" | jq -r '.execution_time_ms')
echo "Took ${TIME_MS}ms"
```

**JSON structure:**

```json
{
  "response": "1. Python\n2. JavaScript\n3. Go",
  "tool_calls": [],
  "server_connections": [],
  "execution_time_ms": 1234,
  "error": null
}
```

**What `jq` does:** Command-line JSON parser

- `.response` extracts response field
- `.error // empty` gets error or empty string
- `-r` removes quotes from output

**Install jq:**

```bash
# macOS
brew install jq

# Ubuntu/Debian
sudo apt install jq

# Windows
choco install jq
```

**Real-world use:**

```bash
#!/bin/bash
# CI/CD: Check if code review is positive

DIFF=$(git diff main)
RESULT=$(echo "$DIFF" | mcp-cli query --json "Rate this code change 1-10")
RATING=$(echo "$RESULT" | jq -r '.response' | grep -oE '[0-9]+' | head -1)

if [ "$RATING" -lt 7 ]; then
    echo "Code quality too low (${RATING}/10). Review needed."
    exit 1
fi
```

---

### Pattern 4: Batch Process Multiple Files

**What it does:** Run same AI query on many files.

```bash
#!/bin/bash

# Process each .txt file
for file in *.txt; do
    echo "Processing $file..."

    OUTPUT_FILE="${file%.txt}_summary.txt"

    cat "$file" | \
        mcp-cli query "Create 3-sentence summary" \
        > "$OUTPUT_FILE"

    echo "  → Saved to $OUTPUT_FILE"
done

echo "Done! Processed $(ls *.txt | wc -l) files"
```

**What `${file%.txt}` does:** Removes `.txt` extension

- `report.txt` → `report`
- Then adds `_summary.txt` → `report_summary.txt`

**Real-world use:**

```bash
#!/bin/bash
# Generate README files for all project directories

for dir in */; do
    if [ -f "$dir/main.py" ] || [ -f "$dir/main.go" ]; then
        echo "Generating README for $dir..."

        # List files and analyze
        ls -la "$dir" | \
            mcp-cli query \
                "Generate README.md for this project structure" \
                > "$dir/README.md"
    fi
done
```

**Performance:** Processing 100 files takes ~5 minutes at ~3 seconds per file

**Cost:** 100 files × $0.001 = ~$0.10

---

### Pattern 5: Add Context to Queries

**What it does:** Give AI background information for better answers.

```bash
#!/bin/bash

# Create context file with project info
cat > context.txt << EOF
Project: $(basename $(pwd))
Language: $(find . -name "*.go" | wc -l) Go files
Files: $(ls -1 | wc -l) total
Git branch: $(git branch --show-current)
Last commit: $(git log -1 --pretty=%B | head -1)
EOF

# Query with context
mcp-cli query \
    --context context.txt \
    --system-prompt "You are a senior software architect" \
    "Analyze this project structure and suggest improvements"

# Clean up
rm context.txt
```

**What `--context` does:** Includes file contents in AI prompt automatically

**What happens:**

1. Context file created with project details
2. AI receives: system prompt + context + your question
3. AI has full picture, gives better answer

**Real-world use:**

```bash
#!/bin/bash
# Generate release notes from git commits

# Get commits since last tag
LAST_TAG=$(git describe --tags --abbrev=0)
COMMITS=$(git log $LAST_TAG..HEAD --pretty=format:"%s")

# Create context
echo "$COMMITS" > commits.txt

# Generate release notes
mcp-cli query \
    --context commits.txt \
    --system-prompt "You write clear, concise release notes" \
    "Create release notes from these commits" \
    > RELEASE_NOTES.md

rm commits.txt
echo "Release notes saved to RELEASE_NOTES.md"
```

---

### Pattern 6: Retry Logic for Reliability

**What it does:** Retry failed queries (network issues, rate limits).

```bash
#!/bin/bash

MAX_RETRIES=3
RETRY_DELAY=2
QUERY="Analyze this data: $(cat data.txt)"

for attempt in $(seq 1 $MAX_RETRIES); do
    echo "Attempt $attempt/$MAX_RETRIES..."

    if RESULT=$(mcp-cli query "$QUERY" 2>&1); then
        echo "$RESULT"
        echo "✓ Success!"
        exit 0
    fi

    # Failed - check if should retry
    if [ $attempt -lt $MAX_RETRIES ]; then
        echo "✗ Failed, retrying in ${RETRY_DELAY}s..."
        sleep $RETRY_DELAY
        RETRY_DELAY=$((RETRY_DELAY * 2))  # Exponential backoff
    else
        echo "✗ Failed after $MAX_RETRIES attempts"
        echo "Last error: $RESULT" >&2
        exit 1
    fi
done
```

**What exponential backoff does:**

- Attempt 1: Wait 2s
- Attempt 2: Wait 4s
- Attempt 3: Wait 8s
- Gives API time to recover

**When to use:**

- ✅ Production scripts
- ✅ Critical automation
- ✅ Network-dependent queries

**When not needed:**

- ❌ Local Ollama (no network)
- ❌ One-off scripts
- ❌ Development/testing

---

### Pattern 7: Error-Safe Pipelines

**What it does:** Handle errors gracefully without breaking pipeline.

```bash
#!/bin/bash
set -euo pipefail  # Exit on error

# Function to safely query AI
safe_query() {
    local question="$1"
    local default="$2"

    if result=$(mcp-cli query "$question" 2>/dev/null); then
        echo "$result"
    else
        echo "$default"
    fi
}

# Use in pipeline
echo "Processing data..." >&2

cat data.txt | \
    safe_query "Extract key points" "Error: Could not analyze" | \
    tee analysis.txt | \
    safe_query "Create executive summary" "Error: Could not summarize" | \
    tee summary.txt

echo "Done! Files created: analysis.txt, summary.txt" >&2
```

**What `set -euo pipefail` does:**

- `set -e`: Exit if any command fails
- `set -u`: Error if using undefined variable
- `set -o pipefail`: Fail if any pipe command fails

**What `2>/dev/null` does:** Suppresses error messages (sends stderr to /dev/null)

**What `tee` does:** Copies output to file AND passes to next command

---

## Advanced Features

### Multiple Tool Calls

Query mode can automatically make multiple tool calls:

```bash
mcp-cli query --server filesystem,brave-search \
    "Search for 'MCP protocol' and save results to mcp-info.txt"
```

**Flow:**

1. AI uses brave-search to find information
2. AI uses filesystem to write file
3. Returns confirmation

### Context from Files

```bash
# Technical background
cat > tech-context.txt << EOF
Stack: Go, PostgreSQL, Redis
Architecture: Microservices
Cloud: AWS EKS
EOF

# Query with context
mcp-cli query \
    --context tech-context.txt \
    "How should we implement caching?"
```

### System Prompt Override

```bash
# Default behavior
mcp-cli query "Write code for binary search"

# With custom system prompt
mcp-cli query \
    --system-prompt "You are a Go expert. Always use Go idioms." \
    "Write code for binary search"
```

### Token Limit Control

```bash
# Brief response (500 tokens max)
mcp-cli query --max-tokens 500 "Explain Docker"

# Detailed response (4000 tokens)
mcp-cli query --max-tokens 4000 "Explain Docker architecture in detail"
```

---

## Error Handling

Query mode provides detailed exit codes for reliable automation.

**What are exit codes?** Numbers programs return when they finish:

- `0` = Success (everything worked)
- `1-255` = Various failures (something went wrong)

**Why they matter:** Your scripts can detect and handle specific failures.

---

### Exit Code Reference

| Code | What It Means             | Common Cause                  | How to Fix                              |
| ---- | ------------------------- | ----------------------------- | --------------------------------------- |
| 0    | **Success**               | Query completed               | No action needed                        |
| 1    | **General error**         | Unknown problem               | Check verbose output                    |
| 2    | **Config not found**      | Missing config.yaml           | Run `mcp-cli init`                      |
| 3    | **Provider not found**    | Provider not configured       | Check config/providers/                 |
| 4    | **Context file missing**  | --context file doesn't exist  | Verify file path                        |
| 5    | **Initialization failed** | Can't start AI provider       | Check API key, network                  |
| 6    | **Query failed**          | AI request failed             | Network issue, rate limit, or bad query |
| 7    | **Output format error**   | Invalid --json output         | Bug, report it                          |
| 8    | **Can't write output**    | Permission denied on --output | Check file permissions                  |

---

### Using Exit Codes in Scripts

**Basic error detection:**

```bash
#!/bin/bash

if mcp-cli query "What is 2+2?"; then
    echo "✓ Query succeeded"
else
    echo "✗ Query failed (exit code: $?)"
    exit 1
fi
```

**What `$?` does:** Contains exit code of last command

---

**Handling specific errors:**

```bash
#!/bin/bash

mcp-cli query "Test query" 2>/dev/null
EXIT_CODE=$?

case $EXIT_CODE in
    0)
        echo "✓ Success"
        ;;
    2)
        echo "✗ Config file missing - run: mcp-cli init"
        exit 1
        ;;
    3)
        echo "✗ Provider not configured"
        echo "  Check: config/providers/"
        exit 1
        ;;
    5)
        echo "✗ Can't connect to AI provider"
        echo "  Check: API key, network connection"
        exit 1
        ;;
    6)
        echo "✗ Query execution failed"
        echo "  Possible causes: network timeout, rate limit, bad query"
        exit 1
        ;;
    *)
        echo "✗ Unknown error (code: $EXIT_CODE)"
        exit 1
        ;;
esac
```

---

**Silent mode (exit codes only):**

```bash
#!/bin/bash

# Suppress all output, only check exit code
if mcp-cli query --error-code-only "Test" 2>/dev/null; then
    echo "AI is working"
else
    echo "AI is down (code: $?)"
fi
```

**What `--error-code-only` does:** 

- No success messages printed
- Only errors go to stderr
- Clean for scripting

**What `2>/dev/null` does:**

- `2>` redirects stderr (error messages)
- `/dev/null` is like trash can (discards output)
- Combined: silences all error messages

---

### Real-World Error Handling Example

**Production-grade script with full error handling:**

```bash
#!/bin/bash
set -euo pipefail

# Configuration
MAX_RETRIES=3
LOG_FILE="query.log"

# Logging function
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

# Safe query function with retries
safe_query() {
    local query="$1"
    local retry_count=0

    while [ $retry_count -lt $MAX_RETRIES ]; do
        log "Attempt $((retry_count + 1))/$MAX_RETRIES: $query"

        if result=$(mcp-cli query "$query" 2>&1); then
            echo "$result"
            log "✓ Success"
            return 0
        fi

        exit_code=$?
        retry_count=$((retry_count + 1))

        case $exit_code in
            2|3|4)
                log "✗ Configuration error (code $exit_code) - Not retrying"
                return $exit_code
                ;;
            5|6)
                if [ $retry_count -lt $MAX_RETRIES ]; then
                    log "✗ Transient error (code $exit_code) - Retrying..."
                    sleep $((retry_count * 2))
                else
                    log "✗ Failed after $MAX_RETRIES attempts"
                    return $exit_code
                fi
                ;;
            *)
                log "✗ Unknown error (code $exit_code)"
                return $exit_code
                ;;
        esac
    done

    return 6
}

# Main script
log "Starting AI query pipeline"

if result=$(safe_query "Analyze system status"); then
    echo "$result" > report.txt
    log "✓ Report generated: report.txt"
    exit 0
else
    log "✗ Pipeline failed"
    exit 1
fi
```

**What this does:**

1. Logs every action with timestamp
2. Retries transient failures (network, API)
3. Doesn't retry config errors (they won't fix themselves)
4. Returns appropriate exit codes
5. Keeps full audit trail in log file

**Production features:**

- ✅ Logging with timestamps
- ✅ Smart retry logic
- ✅ Distinguishes config vs runtime errors
- ✅ Audit trail
- ✅ Clean exit codes

---

### Testing Error Handling

**Test your error handling:**

```bash
#!/bin/bash

# Test 1: Success case
echo "Test 1: Normal query"
if mcp-cli query "What is 2+2?" >/dev/null; then
    echo "✓ Success case works"
fi

# Test 2: Missing config
echo "Test 2: Missing config"
if mcp-cli --config /nonexistent/config.yaml query "test" 2>/dev/null; then
    echo "✗ Should have failed!"
else
    [ $? -eq 2 ] && echo "✓ Correct exit code (2)"
fi

# Test 3: Bad provider
echo "Test 3: Bad provider"
if mcp-cli query --provider nonexistent "test" 2>/dev/null; then
    echo "✗ Should have failed!"
else
    [ $? -eq 3 ] && echo "✓ Correct exit code (3)"
fi

echo "Error handling tests complete"
```

---

## Best Practices

### 1. Use Appropriate Verbosity

```bash
# Production scripts (quiet, clean output)
mcp-cli query "Process data" > result.txt

# Development/debugging (show what's happening)
mcp-cli --noisy query "Test query"
```

### 2. Handle Errors Gracefully

```bash
#!/bin/bash

# Good: Check for errors
if ! RESULT=$(mcp-cli query "Get data" 2>&1); then
    echo "Error: $RESULT" >&2
    exit 1
fi
```

### 3. Use JSON for Parsing

```bash
# Good: Structured output
RESULT=$(mcp-cli query --json "List items" | jq -r '.response')
```

---

## Quick Reference

```bash
# Basic
mcp-cli query "question"

# With provider
mcp-cli query --provider anthropic "question"

# JSON output
mcp-cli query --json "question"

# To file
mcp-cli query --output result.txt "question"

# With context
mcp-cli query --context bg.txt "question"

# Verbose
mcp-cli --noisy query "question"
```

---

## Next Steps

- **[Automation Guide](automation.md)** - Advanced scripting patterns
- **[Chat Mode](chat-mode.md)** - Interactive alternative
- **[Debugging](debugging.md)** - Troubleshooting guide

---

**Ready to automate?** Try your first query! 🚀
