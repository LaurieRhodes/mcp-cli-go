# Debugging Guide

Fix problems and understand what's happening inside MCP-CLI.

**What is debugging?** Finding and fixing issues when things don't work as expected.

**Common problems:**
- "Command not found" (installation issue)
- "Config file not found" (setup issue)
- "API key not found" (configuration issue)
- Empty or wrong responses (provider issue)
- Template not working (template issue)

**This guide shows:** How to diagnose each problem type and fix it.

**Tools you'll use:**
- `--verbose` flag (see what's happening)
- `--noisy` flag (see connections and operations)
- Log files (capture full details)
- Test commands (verify each piece works)

---

## Table of Contents

- [Verbosity Levels](#verbosity-levels)
- [Common Issues](#common-issues)
- [Provider Debugging](#provider-debugging)
- [Server Debugging](#server-debugging)
- [Template Debugging](#template-debugging)
- [Network Issues](#network-issues)
- [Quick Diagnostic Commands](#quick-diagnostic-commands)

---

## Verbosity Levels

**Think of verbosity like zoom levels:**
- **Quiet (default):** Zoomed out - just see final result
- **Noisy:** Zoomed in a bit - see major operations
- **Verbose:** Fully zoomed in - see everything

---

### Quiet Mode (Default)

**What you see:** Just the answer, nothing else.

```bash
mcp-cli query "What is 2+2?"
```

**Output:**
```
The answer is 4.
```

**Use when:**
- ✅ Scripts (clean output)
- ✅ Production (no clutter)
- ✅ Everything working fine

---

### Noisy Mode

**What you see:** Connections, operations, progress.

```bash
mcp-cli --noisy query "What is 2+2?"
```

**Output:**
```
Loading configuration from: config.yaml
Using provider: anthropic
Model: claude-sonnet-4
Connecting to provider...
Sending query...
Received response.

The answer is 4.
```

**What `--noisy` shows:**
- Which config file loaded
- Which AI provider used
- Connection status
- Major operations
- INFO level logs

**Use when:**
- ✅ Want to see what's happening
- ✅ Verifying configuration works
- ✅ Developing automation
- ✅ Something seems slow

**Don't use when:**
- ❌ Scripts that parse output (adds extra lines)
- ❌ Everything working (just adds noise)

---

### Verbose Mode

**What you see:** EVERYTHING - full debug output.

```bash
mcp-cli --verbose query "What is 2+2?"
```

**Output:**
```
[DEBUG] Loading configuration from: /home/user/config.yaml
[DEBUG] Config file parsed successfully
[DEBUG] Loading providers from: config/providers/
[DEBUG] Found provider: anthropic
[DEBUG] Found provider: openai
[DEBUG] Using provider: anthropic
[DEBUG] Model: claude-sonnet-4
[DEBUG] API key loaded from: ANTHROPIC_API_KEY env var
[DEBUG] Creating provider client...
[DEBUG] Connecting to API endpoint: https://api.anthropic.com/v1/messages
[DEBUG] Request headers: {...}
[DEBUG] Request body: {"model":"claude-sonnet-4",...}
[DEBUG] Response status: 200 OK
[DEBUG] Response time: 1.234s
[DEBUG] Tokens used: 15 input, 8 output

The answer is 4.
```

**What `--verbose` shows:**
- EVERYTHING --noisy shows
- DEBUG level logs
- File paths
- API requests/responses
- Timing information
- Token usage
- Internal operations

**Use when:**
- ✅ Something broken (need to see why)
- ✅ Reporting bugs (need full details)
- ✅ Understanding internals
- ✅ Templates not working right

**Don't use when:**
- ❌ Normal operation (too much info)
- ❌ Scripts (pollutes output)

---

### Extra Verbose (Double Verbose)

**For extreme debugging:**

```bash
mcp-cli --verbose --verbose query "Test"
# or
mcp-cli -vv query "Test"
```

**Shows:** Even more internal details, usually only needed when reporting bugs to developers.

---

## Common Issues & Solutions

### Issue 1: "Command not found"

**What this means:** Your terminal can't find the `mcp-cli` program.

**Symptoms:**
```bash
$ mcp-cli query "test"
-bash: mcp-cli: command not found
```

**Why this happens:** Binary not in your system's PATH.

**Solution - Find where it is:**

```bash
# Check if installed at common locations
ls -la /usr/local/bin/mcp-cli       # Common location
ls -la ~/.local/bin/mcp-cli          # User location
ls -la ~/Downloads/mcp-cli-*         # Downloaded but not moved
ls -la ./mcp-cli                     # Current directory
```

**If found in current directory:**
```bash
# Option 1: Run with ./
./mcp-cli query "test"

# Option 2: Move to PATH
sudo mv mcp-cli /usr/local/bin/
mcp-cli query "test"  # Now works!
```

**If in ~/.local/bin but not working:**
```bash
# Add to PATH (add to ~/.bashrc or ~/.zshrc)
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc

# Test
mcp-cli --version
```

**If not found anywhere:**
```bash
# Download it (see installation guide)
wget https://github.com/LaurieRhodes/mcp-cli-go/releases/latest/download/mcp-cli-linux-amd64
chmod +x mcp-cli-linux-amd64
sudo mv mcp-cli-linux-amd64 /usr/local/bin/mcp-cli
```

---

### Issue 2: "Config file not found"

**What this means:** MCP-CLI looking for config.yaml but can't find it.

**Symptoms:**
```bash
$ mcp-cli query "test"
Error: config file not found: config.yaml
```

**Why this happens:** Haven't run `mcp-cli init` yet OR running from wrong directory.

**Solution 1: Initialize config:**
```bash
# Quick setup (Ollama only, no API keys)
mcp-cli init --quick

# Full setup (all providers)
mcp-cli init

# Verify created
ls config.yaml  # Should exist now
```

**Solution 2: Specify config location:**
```bash
# If config is elsewhere
mcp-cli --config /path/to/config.yaml query "test"

# Or CD to config directory
cd /path/to/project
mcp-cli query "test"
```

**Solution 3: Use provider directly (skip config):**
```bash
# Use Ollama without any config
mcp-cli query --provider ollama "test"
```

**Debug which config it's looking for:**
```bash
mcp-cli --verbose query "test" 2>&1 | grep -i "config"

# Output shows:
# Loading configuration from: /home/user/config.yaml
```

---

### Issue 3: "API key not found"

**What this means:** Provider needs API key but MCP-CLI can't find it.

**Symptoms:**
```bash
$ mcp-cli query "test"
Error: OPENAI_API_KEY not found
```

**Why this happens:** API key not in environment or .env file.

**Check if key exists:**

```bash
# Check environment variable
echo $OPENAI_API_KEY
# Should show: sk-proj-...
# If empty, not set!

echo $ANTHROPIC_API_KEY
# Should show: sk-ant-...

# Check .env file
cat .env
# Should contain: OPENAI_API_KEY=sk-...
```

**Solution 1: Add to .env file (recommended):**
```bash
# Create or edit .env (in same directory as config.yaml)
nano .env

# Add your keys (NO quotes, NO spaces around =)
OPENAI_API_KEY=sk-proj-abc123...
ANTHROPIC_API_KEY=sk-ant-xyz789...

# Save and test
mcp-cli query "test"
```

**Solution 2: Set environment variable:**
```bash
# Set for current session
export OPENAI_API_KEY="sk-proj-..."
export ANTHROPIC_API_KEY="sk-ant-..."

# Test
mcp-cli query "test"

# Make permanent (add to ~/.bashrc)
echo 'export OPENAI_API_KEY="sk-proj-..."' >> ~/.bashrc
source ~/.bashrc
```

**Solution 3: Use provider that doesn't need key:**
```bash
# Use local Ollama (free, no API key needed)
mcp-cli query --provider ollama "test"
```

**Verify key is valid:**
```bash
# Test OpenAI key
curl https://api.openai.com/v1/models \
  -H "Authorization: Bearer $OPENAI_API_KEY"
# Should return list of models, not error

# Test Anthropic key
curl https://api.anthropic.com/v1/messages \
  -H "x-api-key: $ANTHROPIC_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"model":"claude-sonnet-4","max_tokens":10,"messages":[{"role":"user","content":"test"}]}'
# Should return response, not "unauthorized"
```

### "Provider not found"

**Problem:** Requested provider not in config

**Check available providers:**
```bash
ls config/providers/

# Or inspect config
cat config.yaml
```

**Solution:**
```bash
# List what's configured
mcp-cli --verbose query --provider test "Test" 2>&1 | grep -i provider

# Initialize missing provider
mcp-cli init
# Choose provider during setup
```

### "No response" or "Empty output"

**Problem:** Query completes but no output

**Debug:**
```bash
# Check what happened
mcp-cli --verbose query "Question"

# Check provider is responding
mcp-cli chat --provider anthropic
# Try a simple question

# Try different provider
mcp-cli query --provider ollama "Test"
```

**Common causes:**
- Provider API down
- Rate limiting
- Invalid API key
- Model not accessible

### "Tool execution failed"

**Problem:** MCP server tool errors

**Debug:**
```bash
# Check available tools
mcp-cli chat
# Then: /tools

# Test specific server
mcp-cli --verbose chat --server filesystem
# Try using a tool

# Check server configuration
cat config/servers/filesystem.yaml
```

**Solutions:**
```bash
# Verify server binary exists
which filesystem-server

# Check server permissions
ls -la /usr/local/bin/*-server

# Test server directly
filesystem-server --version
```

---

## Provider Debugging

### Test Provider Connection

```bash
# Quick test
mcp-cli query --provider anthropic "Hello"

# With debugging
mcp-cli --verbose query --provider anthropic "Hello"
```

### Provider-Specific Issues

#### Anthropic (Claude)

```bash
# Check API key
echo $ANTHROPIC_API_KEY

# Test with verbose
mcp-cli --verbose query --provider anthropic "Test"

# Common errors:
# - "Invalid API key" → Check key format (starts with sk-ant-)
# - "Rate limit" → Wait and retry
# - "Model not found" → Verify model name
```

#### OpenAI (GPT)

```bash
# Check API key
echo $OPENAI_API_KEY

# Test with specific model
mcp-cli query --provider openai --model gpt-4o "Test"

# Common errors:
# - "Invalid API key" → Check key (starts with sk-)
# - "Model not found" → Check model access
# - "Quota exceeded" → Check billing
```

#### Ollama (Local)

```bash
# Check Ollama is running
curl http://localhost:11434/api/tags

# List models
ollama list

# Test model
mcp-cli query --provider ollama --model llama3.2 "Test"

# Common errors:
# - "Connection refused" → Start Ollama: `ollama serve`
# - "Model not found" → Pull model: `ollama pull llama3.2`
```

---

## Server Debugging

### List Connected Servers

```bash
# In chat mode
mcp-cli chat
# Then: /tools

# With verbose
mcp-cli --verbose chat --server filesystem
```

### Test Specific Server

```bash
# Test filesystem server
mcp-cli --verbose query --server filesystem "List files in current directory"

# Check server output
mcp-cli --noisy chat --server brave-search
# Try searching
```

### Server Configuration Issues

```bash
# Check server config exists
ls config/servers/

# Validate YAML
cat config/servers/filesystem.yaml

# Test server binary
which filesystem-server
filesystem-server --version
```

### Server Connection Failures

**Debug:**
```bash
# Full debug output
mcp-cli --verbose chat --server problematic-server 2>&1 | less

# Look for:
# - "Failed to connect"
# - "Server not responding"
# - "Tool execution error"
```

**Common fixes:**
```bash
# Restart server
pkill -9 filesystem-server
# Restart mcp-cli

# Check permissions
chmod +x /usr/local/bin/filesystem-server

# Verify configuration
cat config/servers/filesystem.yaml
```

---

## Template Debugging

### Test Template Step-by-Step

```bash
# Test individual steps
mcp-cli query "Step 1 prompt"
mcp-cli query "Step 2 prompt with mock data"

# Run template with verbose
mcp-cli --verbose --template my_template
```

### Common Template Issues

#### "Template not found"

```bash
# Check template location
ls config/templates/

# Check template name
cat config/templates/my_template.yaml
# Verify "name:" field matches

# List available templates
mcp-cli --list-templates
```

#### "Variable not found"

**Debug:**
```yaml
# Add debug step to template
steps:
  - name: debug
    prompt: "Show me {{variable_name}}"
  
  - name: actual_step
    prompt: "Use {{variable_name}}"
```

#### "Step execution failed"

```bash
# Run with verbose
mcp-cli --verbose --template failing_template

# Check step configuration
cat config/templates/failing_template.yaml

# Test step prompt separately
mcp-cli query "The exact prompt from the step"
```

---

## Network Issues

### Provider API Timeouts

```bash
# Check network
ping api.openai.com
ping api.anthropic.com

# Test with curl
curl https://api.openai.com/v1/models \
  -H "Authorization: Bearer $OPENAI_API_KEY"

# Increase timeout (if supported)
mcp-cli query --max-tokens 1000 "Question"
```

### Proxy Configuration

```bash
# Set proxy
export HTTP_PROXY=http://proxy:8080
export HTTPS_PROXY=http://proxy:8080

# Test
mcp-cli --verbose query "Test"
```

### SSL/TLS Issues

```bash
# Check certificates
curl https://api.anthropic.com

# Skip verification (not recommended for production)
# Currently not supported - file feature request if needed
```

---

## Log Analysis

### Capture Full Logs

```bash
# Redirect all output
mcp-cli --verbose query "Test" > output.txt 2>&1

# View logs
less output.txt

# Search for errors
grep -i error output.txt
grep -i failed output.txt
```

### Important Log Patterns

**Look for:**

```
# Configuration loading
Loading configuration from: /path/to/config.yaml

# Provider initialization
Using provider: anthropic
Model: claude-sonnet-4

# Server connections
Connected to server: filesystem
Available tools: 15

# Tool execution
Executing tool: list_directory
Tool result: {...}

# Errors
Error: failed to connect
Failed to execute tool
API request failed
```

### Save Debugging Session

```bash
#!/bin/bash

# Create debug log
DEBUG_LOG="debug-$(date +%Y%m%d-%H%M%S).log"

echo "=== MCP-CLI Debug Session ===" > "$DEBUG_LOG"
echo "Date: $(date)" >> "$DEBUG_LOG"
echo "Version: $(mcp-cli --version)" >> "$DEBUG_LOG"
echo "" >> "$DEBUG_LOG"

echo "=== Configuration ===" >> "$DEBUG_LOG"
cat config.yaml >> "$DEBUG_LOG"
echo "" >> "$DEBUG_LOG"

echo "=== Environment ===" >> "$DEBUG_LOG"
env | grep -E "MCP|OPENAI|ANTHROPIC" >> "$DEBUG_LOG"
echo "" >> "$DEBUG_LOG"

echo "=== Test Query ===" >> "$DEBUG_LOG"
mcp-cli --verbose query "Test" >> "$DEBUG_LOG" 2>&1

echo "Debug log saved to: $DEBUG_LOG"
```

---

## Quick Diagnostic Commands

```bash
# Version
mcp-cli --version

# Test basic query
mcp-cli query --provider ollama "Test"

# List templates
mcp-cli --list-templates

# Check configuration
cat config.yaml

# Test with full debug
mcp-cli --verbose query "Debug test" 2>&1 | less

# Check API keys
env | grep API_KEY

# List servers
ls config/servers/

# Test server
mcp-cli chat --server filesystem
# Then: /tools
```

---

## Reporting Issues

When reporting a bug, include:

### 1. System Information

```bash
mcp-cli --version
uname -a  # Linux/macOS
# Or
ver       # Windows
```

### 2. Configuration

```bash
# Sanitize API keys first!
cat config.yaml | sed 's/api_key:.*/api_key: REDACTED/'
```

### 3. Debug Output

```bash
mcp-cli --verbose [command] 2>&1 > debug.log
# Attach debug.log (sanitize first!)
```

### 4. Steps to Reproduce

```bash
# Exact commands run
mcp-cli query "Question"

# Expected behavior
# Actual behavior
# Error message
```

---

## Quick Reference

```bash
# Basic debugging
mcp-cli --verbose query "Test"

# Test provider
mcp-cli query --provider ollama "Hello"

# Test server
mcp-cli chat --server filesystem

# Check config
cat config.yaml

# Full diagnostic
mcp-cli --verbose --template test 2>&1 | tee debug.log
```

---

## Next Steps

- **[Chat Mode](chat-mode.md)** - Interactive debugging
- **[Query Mode](query-mode.md)** - Script debugging
- **[GitHub Issues](https://github.com/LaurieRhodes/mcp-cli-go/issues)** - Report bugs

---

**Still stuck?** Ask in [Discussions](https://github.com/LaurieRhodes/mcp-cli-go/discussions)! 🔧
