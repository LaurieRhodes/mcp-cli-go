# Server Mode Troubleshooting

Common issues specific to MCP server mode and their solutions.

---

## Tool Not Executing

### Symptom
Tool appears in client but fails when called.

### Diagnosis

**1. Check template exists:**
```bash
ls config/templates/your_template.yaml
```

**2. Test template directly:**
```bash
mcp-cli --template your_template --input-data '{"param": "test"}'
```

**3. Check logs:**
```bash
mcp-cli serve --verbose config/runas/server.yaml
# Watch for errors when tool is called
```

### Common Causes

**Template not found:**
```yaml
# runas config
template: code_analyzer  # ← Must match filename exactly

# Must exist:
# config/templates/code_analyzer.yaml  ✓
# NOT: config/templates/code-analyzer.yaml  ✗
```

**Parameter mismatch:**
```yaml
# Client sends: {"code": "..."}
# Template expects: {{input_data.code}}
# Not: {{code}}  ✗
```

**Missing API key:**
```bash
# Template uses provider but key not set
echo $OPENAI_API_KEY  # Should show key, not empty
```

---

## Parameters Not Mapping

### Symptom
Tool receives parameters but template gets empty/wrong values.

### Diagnosis

**Add debug output to template:**
```yaml
steps:
  - name: debug
    prompt: |
      Debug: Received parameters:
      {{input_data}}
```

### Common Causes

**Wrong variable prefix:**
```yaml
# Wrong
prompt: "Process: {{code}}"

# Correct  
prompt: "Process: {{input_data.code}}"
```

**Nested access syntax:**
```yaml
# Parameter is object:
parameters:
  config:
    type: object
    properties:
      mode: {type: string}

# Template access:
# Correct: {{input_data.config.mode}}
# Wrong: {{input_data.config}}  ✗ (gets whole object)
```

**Type mismatch:**
```yaml
# Server expects number
parameters:
  count:
    type: number

# Client sends string "5" instead of number 5
# Fix: Client must send correct JSON type
```

---

## Server Won't Start

### Symptom
`mcp-cli serve` exits immediately or hangs.

### Diagnosis

**1. Check YAML syntax:**
```bash
python3 -c "import yaml; yaml.safe_load(open('config/runas/server.yaml'))"
# Shows syntax errors if any
```

**2. Check template paths:**
```bash
mcp-cli --list-templates
# Verify referenced templates exist
```

**3. Run with verbose:**
```bash
mcp-cli serve --verbose config/runas/server.yaml
# See detailed startup process
```

### Common Causes

**Invalid YAML:**
```yaml
# Wrong indentation
tools:
  - name: tool1
  description: "Bad"  # Should be indented under tool

# Correct
tools:
  - name: tool1
    description: "Good"
```

**Missing required fields:**
```yaml
# Missing template reference
tools:
  - name: my_tool
    parameters: {...}
    # Missing: template: template_name
```

**Circular template reference:**
```yaml
# template_a.yaml
steps:
  - template: template_b

# template_b.yaml
steps:
  - template: template_a  # ← Circular!
```

---

## Client Can't Connect

### Symptom
MCP client doesn't show tools or reports connection error.

### Diagnosis

**1. Verify server path in client config:**
```json
{
  "mcpServers": {
    "server": {
      "command": "mcp-cli",
      "args": ["serve", "/absolute/path/to/config.yaml"]
      //                  ↑ Must be absolute, not relative
    }
  }
}
```

**2. Check mcp-cli in PATH:**
```bash
which mcp-cli
# Should show path, not empty
```

**3. Test server manually:**
```bash
mcp-cli serve /absolute/path/to/config.yaml
# Should show: Server ready! Listening on stdio...
```

### Common Causes

**Relative path in client config:**
```json
// Wrong
"args": ["serve", "config/runas/server.yaml"]

// Correct
"args": ["serve", "/Users/you/project/config/runas/server.yaml"]
```

**mcp-cli not in PATH:**
```bash
# Client can't find command
# Fix: Install to system location
sudo cp mcp-cli /usr/local/bin/
```

**Client config syntax error:**
```json
{
  "mcpServers": {
    "server1": {...},  // ← Trailing comma on last item!
  }
}
```

---

## Tool Returns Error

### Symptom
Tool executes but returns error instead of result.

### Diagnosis

**Check error message from client:**
- Usually shows in client UI
- Or check client logs

**Test template with same input:**
```bash
mcp-cli --template template_name \
  --input-data '{"param": "same value client sent"}'
```

### Common Causes

**Required parameter missing:**
```yaml
# Tool definition
parameters:
  code:
    required: true  # ← Must be provided

# Client sent: {}  ← Empty, missing required parameter
```

**Template execution failure:**
```yaml
# Template references undefined variable
prompt: "Process: {{input_data.undefined}}"
# Client didn't send 'undefined' parameter
```

**AI provider error:**
```bash
# API key invalid or quota exceeded
# Check: echo $OPENAI_API_KEY
# Test: mcp-cli query "test"
```

---

## Performance Issues

### Symptom
Tool takes very long to execute.

### Diagnosis

**Time template directly:**
```bash
time mcp-cli --template template_name
# See how long template takes
```

**Check template complexity:**
```yaml
# Count steps
grep "^  - name:" config/templates/template.yaml | wc -l
# Many steps = longer execution
```

### Solutions

**Use faster model:**
```yaml
# Slow
config:
  defaults:
    provider: anthropic
    model: claude-opus-4

# Faster
config:
  defaults:
    provider: openai
    model: gpt-4o-mini
```

**Combine steps:**
```yaml
# Instead of 5 separate steps:
steps:
  - name: step1
    prompt: "..."
  - name: step2
    prompt: "..."
  # ... 3 more steps

# Combine into 2:
steps:
  - name: analyze
    prompt: "Do steps 1-3..."
  - name: format
    prompt: "Do steps 4-5..."
```

**Cache common results:**
```yaml
# If template frequently processes same input
# Consider caching at application level
```

---

## Environment Variables Not Working

### Symptom
Template references `${VAR}` but gets empty value.

### Diagnosis

**1. Check variable is set:**
```bash
echo $OPENAI_API_KEY
# Should show value
```

**2. Check .env file location:**
```bash
ls -la .env config.yaml
# Both should be in same directory
```

**3. Test variable expansion:**
```bash
mcp-cli query "test"  # Uses same env loading
# If this works, env vars are loaded
```

### Solutions

**Set in .env file:**
```bash
# .env (in same dir as config.yaml)
OPENAI_API_KEY=sk-...
ANTHROPIC_API_KEY=sk-ant-...
```

**Or pass via client config:**
```json
{
  "mcpServers": {
    "server": {
      "command": "mcp-cli",
      "args": ["serve", "/path/to/config.yaml"],
      "env": {
        "OPENAI_API_KEY": "sk-...",
        "ANTHROPIC_API_KEY": "sk-ant-..."
      }
    }
  }
}
```

**Or set system-wide:**
```bash
# In ~/.bashrc or ~/.zshrc
export OPENAI_API_KEY="sk-..."
```

---

## Multiple Tools Interfering

### Symptom
One tool works, another doesn't, but both use same template structure.

### Diagnosis

**Test each tool independently:**
```bash
# Test tool 1
mcp-cli --template template1

# Test tool 2
mcp-cli --template template2
```

**Check for name conflicts:**
```yaml
tools:
  - name: analyze  # ← Must be unique
    template: template1
  
  - name: analyze  # ← CONFLICT! Same name
    template: template2
```

### Solution

**Ensure unique tool names:**
```yaml
tools:
  - name: analyze_code      # ← Unique
  - name: analyze_data      # ← Unique
  - name: analyze_security  # ← Unique
```

---

## Debugging Workflow

**When something's broken:**

1. **Isolate the problem:**
   ```bash
   # Does template work directly?
   mcp-cli --template template_name
   
   # Does server start?
   mcp-cli serve config/runas/server.yaml
   
   # Can client connect?
   # Check client logs
   ```

2. **Add logging:**
   ```yaml
   # Template debug step
   steps:
     - name: debug_input
       prompt: "Echo input: {{input_data}}"
   ```

3. **Simplify:**
   ```yaml
   # Reduce to minimal example
   tools:
     - name: simple_test
       template: simple_template
       parameters:
         input:
           type: string
           required: true
   ```

4. **Verify assumptions:**
   ```bash
   # Template exists?
   ls config/templates/template.yaml
   
   # YAML valid?
   python3 -c "import yaml; yaml.safe_load(open('...'))"
   
   # Path absolute?
   realpath config/runas/server.yaml
   ```

---

## Quick Diagnostic Commands

```bash
# Verify mcp-cli installation
which mcp-cli && mcp-cli --version

# Test server startup
timeout 5 mcp-cli serve config/runas/server.yaml &
ps aux | grep mcp-cli

# Test template
mcp-cli --template template_name --input-data '{"test": "value"}'

# Validate YAML
python3 -c "import yaml; yaml.safe_load(open('config/runas/server.yaml'))"

# List available templates
mcp-cli --list-templates

# Check environment
env | grep -i api

# Test with verbose logging
mcp-cli serve --verbose config/runas/server.yaml 2>&1 | tee debug.log
```

---

## Still Stuck?

**If these solutions don't help:**

1. **Simplify to minimal reproduction:**
   - One tool
   - One template
   - One parameter
   - Test if that works

2. **Compare with working example:**
   - Use example from `examples/code-analysis.md`
   - Verify that works
   - Identify differences

3. **Check assumptions:**
   - Is template in right location?
   - Are paths absolute?
   - Is YAML syntax valid?
   - Are API keys set?

**Most server-mode issues are:**
- Configuration mistakes (90%)
- Template problems (5%)
- Client setup issues (5%)

**Start with configuration, work backwards from there.**
