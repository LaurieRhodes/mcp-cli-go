# Frequently Asked Questions (FAQ)

Common questions and answers about MCP-CLI-Go.

---

## Installation & Setup

### Q: Which binary should I download?

**A:** Match your operating system and CPU architecture.

**How to check your system:**

**Linux/macOS:**

```bash
uname -m
# Output tells you:
# x86_64 or amd64 → Intel/AMD 64-bit
# aarch64 or arm64 → ARM 64-bit (Raspberry Pi, newer Macs)
```

**Windows PowerShell:**

```powershell
$env:PROCESSOR_ARCHITECTURE
# Output: AMD64 → use windows-amd64
```

**Download the right binary:**

| Your System                               | Download This               |
| ----------------------------------------- | --------------------------- |
| Linux on Intel/AMD (most common)          | `mcp-cli-linux-amd64`       |
| Linux on ARM (Raspberry Pi, some servers) | `mcp-cli-linux-arm64`       |
| macOS Intel (before 2020)                 | `mcp-cli-darwin-amd64`      |
| macOS Apple Silicon (M1/M2/M3 - 2020+)    | `mcp-cli-darwin-arm64`      |
| Windows (almost all PCs)                  | `mcp-cli-windows-amd64.exe` |

**Still not sure?**

- Mac: Apple menu → About This Mac → look for "Chip" (Intel vs M1/M2/M3)
- Windows: Almost certainly AMD64
- Linux: Run `uname -m`

---

### Q: Do I need API keys to use MCP-CLI?

**A:** It depends on which AI providers you want to use.

**No API keys needed (FREE):**

- ✅ **Ollama** - Run AI models on your own computer
  
  - Download from: https://ollama.com
  - Models run locally, completely free
  - No internet needed after model download
  - Privacy: data never leaves your computer

- ✅ **LM Studio** - Another local AI option
  
  - Download from: https://lmstudio.ai
  - Also runs models locally

**API keys required (PAID - charges per use):**

- 💰 **OpenAI** - GPT-4, GPT-4o, etc.
  
  - ~$0.01-0.06 per 1000 words processed
  - Get key at: https://platform.openai.com/api-keys

- 💰 **Anthropic** - Claude models
  
  - ~$0.01-0.08 per 1000 words processed  
  - Get key at: https://console.anthropic.com/

- 💰 **DeepSeek** - Cost-effective alternative
  
  - ~$0.001 per 1000 words (very cheap!)
  - Get key at: https://platform.deepseek.com/

- 💰 **Google Gemini** - Large context windows
  
  - ~$0.01 per 1000 words
  - Get key at: https://makersuite.google.com/app/apikey

- 💰 **OpenRouter** - Access to many models
  
  - Various pricing
  - Get key at: https://openrouter.ai/keys

**Recommendation for beginners:**

1. Start with `mcp-cli init --quick` (sets up Ollama)
2. Install Ollama: https://ollama.com
3. Pull a model: `ollama pull llama3.2`
4. Use for free: `mcp-cli query "test"`
5. Add paid APIs later if needed

**Cost comparison for 1000 queries:**

- Ollama (local): $0 (free!)
- DeepSeek: ~$1
- Claude/GPT-4: ~$10-60

---

### Q: Where do I put API keys?

**A:** There are three ways. Use method #1 (most secure).

**Method 1: .env file (RECOMMENDED)**

Create a file named `.env` in the same directory as your mcp-cli executable:

```bash
# Create .env file
nano .env
# or
code .env
# or any text editor
```

Add your keys (NO quotes, NO spaces around =):

```bash
OPENAI_API_KEY=sk-proj-abc123...
ANTHROPIC_API_KEY=sk-ant-xyz789...
DEEPSEEK_API_KEY=sk-def456...
```

**What this means:**

- Keys are in one file
- File is hidden (starts with .)
- Easy to update
- Add `.env` to `.gitignore` to keep keys private

**Why this is best:**

- ✅ Secure (not in code)
- ✅ Easy to update
- ✅ Won't accidentally commit to git (if .gitignore set)

---

**Method 2: Environment variables**

Set in your shell (Linux/macOS):

```bash
export OPENAI_API_KEY=sk-proj-...
export ANTHROPIC_API_KEY=sk-ant-...

# Add to ~/.bashrc or ~/.zshrc to make permanent
echo 'export OPENAI_API_KEY=sk-proj-...' >> ~/.bashrc
```

Windows PowerShell:

```powershell
$env:OPENAI_API_KEY = "sk-proj-..."
$env:ANTHROPIC_API_KEY = "sk-ant-..."

# Make permanent (run as Administrator):
[Environment]::SetEnvironmentVariable("OPENAI_API_KEY", "sk-proj-...", "User")
```

**Why you might use this:**

- ✅ Good for Docker containers
- ✅ Good for CI/CD pipelines
- ❌ Must set in every terminal session (unless made permanent)

---

**Method 3: Config files (NOT RECOMMENDED)**

```yaml
# config/providers/openai.yaml
config:
  api_key: sk-proj-...  # AVOID THIS!
```

**Why avoid this:**

- ❌ Key stored in plain text
- ❌ Easy to accidentally commit to git
- ❌ Harder to update
- ❌ Less secure

**Only use for:**

- Testing
- Keys that are meant to be shared
- Non-production environments

---

### Q: How do I verify my installation?

**A:** Run these checks:

```bash
# 1. Check version
mcp-cli --version
# Should show: mcp-cli version vX.Y.Z

# 2. Check it's static (Linux only)
ldd mcp-cli
# Should show: not a dynamic executable

# 3. Test a simple query
mcp-cli query "What is 2+2?"
```

---

## Configuration

### Q: What's the difference between config.yaml and config/settings.yaml?

**A:**

**config.yaml** (root level):

- Just has `includes` directives
- Points to modular config files
- Minimal, rarely changes

**config/settings.yaml**:

- Global settings (AI, embeddings, logging)
- Default provider and prompts
- Configuration that changes

Think of `config.yaml` as an index, `settings.yaml` as preferences.

---

### Q: Do I need to create all the config directories?

**A:** No! `mcp-cli init` creates everything:

```bash
# Quick setup (just Ollama)
mcp-cli init --quick

# Full setup (all providers)
mcp-cli init
```

This creates:

- `config.yaml`
- `.env`
- `config/` directory with subdirectories
- Example configurations

---

### Q: Can I have multiple config files?

**A:** Yes! Specify which config to use:

```bash
# Use specific config
mcp-cli --config production.yaml query "..."

# Use config in different directory
mcp-cli --config /path/to/config.yaml query "..."

# Default: looks for config.yaml next to executable
```

---

### Q: My template isn't found. Why?

**A:** Check these:

1. **Location**: Templates must be in `config/templates/`
2. **Extension**: Must be `.yaml` or `.yml`
3. **Name**: Use the `name` field in YAML, not filename

```bash
# List available templates
mcp-cli --list-templates

# Check template location
ls config/templates/

# Verify template syntax
cat config/templates/your-template.yaml
```

---

## Usage

### Q: What's the difference between query, chat, and template modes?

**A:**

**Query Mode**: One-shot questions

```bash
mcp-cli query "What is 2+2?"
```

- Best for: Scripts, automation, CI/CD
- Returns: Single answer, then exits

**Chat Mode**: Interactive conversation

```bash
mcp-cli chat
```

- Best for: Exploration, development
- Returns: Ongoing conversation with history

**Template Mode**: Multi-step workflows

```bash
mcp-cli --template analyze
```

- Best for: Repeatable processes, production
- Returns: Final result of workflow

---

### Q: How do I pass data to a template?

**A:** Three ways:

**1. Stdin (pipe):**

```bash
echo "data" | mcp-cli --template analyze
cat file.txt | mcp-cli --template process
```

**2. --input-data flag:**

```bash
mcp-cli --template analyze --input-data "text to analyze"
```

**3. File input:**

```bash
mcp-cli --template analyze --input-data "$(cat file.txt)"
```

Access in template with `{{stdin}}` or `{{input_data}}`.

---

### Q: Can I use mcp-cli in scripts?

**A:** Absolutely! That's what query mode is for:

```bash
#!/bin/bash

# Get answer
ANSWER=$(mcp-cli query "What is 2+2?")
echo "Answer: $ANSWER"

# JSON parsing
RESULT=$(mcp-cli query --json "List 3 colors")
FIRST_COLOR=$(echo "$RESULT" | jq -r '.response')

# Conditional logic
if mcp-cli query "Is the sky blue?" | grep -q "yes"; then
    echo "Correct!"
fi

# Template execution
cat data.txt | mcp-cli --template analyze > report.txt
```

---

## Templates

### Q: What's template composition?

**A:** Templates calling other templates:

```yaml
# parent.yaml
steps:
  - name: call_child
    template: child_template  # Calls child_template.yaml
    template_input: "{{data}}"
    output: result
```

**Benefits:**

- Reuse existing templates
- Modular workflows
- 50-87% token savings (context isolation)

See [Concepts: Template Composition](concepts.md#template-composition)

---

### Q: How deep can templates nest?

**A:** Up to 10 levels:

```
Level 1: parent
  └─► Level 2: child
       └─► Level 3: grandchild
            └─► ... up to Level 10
```

This prevents infinite loops while allowing complex workflows.

---

### Q: Can I use multiple AI providers in one template?

**A:** Yes! That's a key feature:

```yaml
steps:
  - name: research
    provider: anthropic  # Claude for research
    prompt: "Research: {{topic}}"

  - name: verify
    provider: openai  # GPT-4 for verification
    prompt: "Verify: {{research}}"

  - name: format
    provider: ollama  # Local model for formatting (free!)
    prompt: "Format: {{verify}}"
```

See [Concepts: Multi-Provider Workflows](concepts.md#multi-provider-workflows)

---

### Q: How do I debug a template?

**A:** Use these techniques:

**1. Verbose mode:**

```bash
mcp-cli --verbose --template my_template
```

**2. Test steps individually:**

```bash
# Test step 1 alone
mcp-cli query "step 1 prompt"

# Test step 2 with mock data
mcp-cli query "step 2 prompt with mock {{data}}"
```

**3. Add intermediate outputs:**

```yaml
steps:
  - name: step1
    prompt: "..."
    output: step1_result  # Save for inspection

  - name: debug
    prompt: "Show me: {{step1_result}}"  # See what step1 produced
```

**4. Check template syntax:**

```bash
# Make sure YAML is valid
cat config/templates/my_template.yaml | yaml-lint
```

---

## Providers

### Q: Which provider should I use?

**A:** Depends on your needs:

| Need                | Recommended Provider        |
| ------------------- | --------------------------- |
| **Best overall**    | Anthropic (Claude Sonnet 4) |
| **Best for code**   | OpenAI (GPT-4o)             |
| **Cheapest**        | Ollama (free, local)        |
| **Largest context** | Google Gemini (1M+ tokens)  |
| **Best value**      | DeepSeek                    |
| **Most models**     | OpenRouter                  |

Or mix them in one workflow!

---

### Q: Do I need Ollama installed?

**A:** Only if you want local models:

**With Ollama:**

- Free inference
- Privacy (stays local)
- No API keys needed
- Works offline

**Without Ollama:**

- Use cloud providers (OpenAI, Anthropic, etc.)
- Requires API keys
- Costs money per request

Install Ollama: https://ollama.com

---

### Q: Can I use my own OpenAI-compatible API?

**A:** Yes! Configure any OpenAI-compatible endpoint:

```yaml
# config/providers/custom.yaml
interface_type: openai_compatible
provider_name: custom
config:
  api_endpoint: https://your-api.com/v1
  api_key: ${YOUR_API_KEY}
  default_model: your-model-name
```

Works with:

- LM Studio
- Text Generation WebUI
- vLLM
- Any OpenAI-compatible API

---

## MCP Servers

### Q: What's an MCP server?

**A:** A tool that gives AI access to external systems:

- **Filesystem**: Read/write files
- **Brave Search**: Web search
- **Database**: Query databases
- **GitHub**: Repository operations
- **Custom**: Your own tools

See [Concepts: MCP Servers vs Providers](concepts.md#mcp-servers-vs-providers)

---

### Q: Do I need MCP servers?

**A:** No, they're optional:

**Without MCP servers:**

- AI answers from training knowledge
- No external data access
- Simpler setup

**With MCP servers:**

- AI can read files
- AI can search web
- AI can query databases
- More powerful workflows

Start without them, add later when needed.

---

### Q: How do I add an MCP server?

**A:** Three steps:

**1. Install the server binary:**

```bash
# Example: Brave search server
npm install -g @modelcontextprotocol/server-brave-search
```

**2. Create config file:**

```yaml
# config/servers/brave-search.yaml
server_name: brave-search
config:
  command: brave-search-server
  env:
    BRAVE_API_KEY: ${BRAVE_API_KEY}
```

**3. Use in query/template:**

```bash
mcp-cli query --server brave-search "Search for AI news"
```

---

## Performance

### Q: How much do tokens really save with composition?

**A:** Real-world examples:

| Workflow                 | Without Composition | With Composition | Savings |
| ------------------------ | ------------------- | ---------------- | ------- |
| 3-step document analysis | 12,500 tokens       | 6,200 tokens     | 50%     |
| 5-step research pipeline | 28,000 tokens       | 4,100 tokens     | 85%     |
| 4-step validation        | 15,800 tokens       | 6,800 tokens     | 57%     |

See [Concepts: Context Isolation](concepts.md#context-isolation)

---

### Q: How can I reduce costs?

**A:** Strategies:

**1. Use local models for cheap tasks:**

```yaml
steps:
  - provider: anthropic  # Expensive, for complex work
    prompt: "Deep analysis..."

  - provider: ollama  # Free, for simple formatting
    prompt: "Format as markdown..."
```

**2. Use template composition** (automatic savings)

**3. Use cheaper models when possible:**

- GPT-4o-mini instead of GPT-4o
- Claude Haiku instead of Claude Opus

**4. Cache repeated queries** (coming soon)

---

## Errors & Troubleshooting

### Q: "Command not found" error

**A:** Binary not in PATH:

```bash
# Find where it is
which mcp-cli

# Add to PATH (in ~/.bashrc or ~/.zshrc)
export PATH="/usr/local/bin:$PATH"

# Or move to PATH location
sudo mv mcp-cli /usr/local/bin/
```

---

### Q: "API key not found" error

**A:** Check these:

1. **.env file exists** next to config.yaml
2. **Environment variable** is set: `echo $OPENAI_API_KEY`
3. **Variable name** matches provider config
4. **No quotes** around key in .env: `OPENAI_API_KEY=sk-...` not `"sk-..."`

---

### Q: macOS "cannot be opened because developer cannot be verified"

**A:** Remove quarantine attribute:

```bash
xattr -d com.apple.quarantine /usr/local/bin/mcp-cli
```

Or: System Preferences → Security & Privacy → Allow

---

### Q: Template execution hangs / times out

**A:** Possible causes:

**1. Model is thinking** (normal for complex prompts)

- Wait longer
- Try simpler prompt
- Use faster model

**2. Network issues**

- Check internet connection
- Try `--verbose` to see where it stops

**3. Server timeout**

```yaml
# Increase timeout in provider config
config:
  timeout_seconds: 600  # 10 minutes
```

**4. MCP server not responding**

- Check server is running
- Check server logs
- Try without server

---

### Q: How do I see what's happening?

**A:** Use verbose mode:

```bash
# See all operations
mcp-cli --verbose query "..."

# See even more detail
mcp-cli --verbose --verbose query "..."

# Template execution
mcp-cli --verbose --template my_template
```

This shows:

- Configuration loading
- Server connections
- API calls
- Variable values
- Step execution

---

## Advanced

### Q: Can I run mcp-cli as a service?

**A:** Yes! See [Deployment Guide](../deployment/):

**Linux (systemd):**

```bash
# Create service file
sudo systemctl enable mcp-cli.service
sudo systemctl start mcp-cli
```

**Docker:**

```bash
docker run -d --restart=always \
  -v ./config:/app/config \
  mcp-cli serve config/runas/agent.yaml
```

---

### Q: Can I use mcp-cli in CI/CD?

**A:** Absolutely! Common patterns:

**GitHub Actions:**

```yaml
- name: Analyze commit
  run: |
    mcp-cli query "Review this commit: $(git log -1)" > review.txt
```

**GitLab CI:**

```yaml
analyze:
  script:
    - mcp-cli --template code_review --input-data "$CI_COMMIT_MESSAGE"
```

See [Automation Guide](../guides/automation.md)

---

### Q: Is there a way to cache results?

**A:** Not yet, but coming soon:

**Planned features:**

- Response caching
- Embedding caching
- Template result caching

Track: [Issue #XX](https://github.com/LaurieRhodes/mcp-cli-go/issues)

---

### Q: Can I contribute templates?

**A:** Yes! Share in [Show & Tell](https://github.com/LaurieRhodes/mcp-cli-go/discussions/categories/show-and-tell)

Or submit to template library (coming soon).

---

## Still Need Help?

- **Documentation**: [docs/](../)
- **Discussions**: [GitHub Discussions](https://github.com/LaurieRhodes/mcp-cli-go/discussions)
- **Issues**: [GitHub Issues](https://github.com/LaurieRhodes/mcp-cli-go/issues)
- **Examples**: [Template Examples](../templates/examples/)

---

**Don't see your question?** [Ask in Discussions](https://github.com/LaurieRhodes/mcp-cli-go/discussions/new) 💬
