# Core Concepts

This guide explains the fundamental concepts that make MCP-CLI-Go powerful and unique.

---

## Table of Contents

- [What is MCP?](#what-is-mcp)
- [What is a Template?](#what-is-a-template)
- [Template Composition](#template-composition)
- [Multi-Provider Workflows](#multi-provider-workflows)
- [Variables and Data Flow](#variables-and-data-flow)
- [Execution Modes](#execution-modes)
- [Context Isolation](#context-isolation)
- [MCP Servers vs Providers](#mcp-servers-vs-providers)

---

## What is MCP?

**MCP (Model Context Protocol)** is an open standard created by Anthropic for connecting AI assistants to tools and data.

**In plain English:** It's a way for AI to use tools (like reading files, searching web, querying databases) through a standard interface.

**Think of it like USB:**

- USB is a standard plug that works with any computer
- MCP is a standard protocol that works with any AI

### The Problem MCP Solves

**Before MCP (Traditional AI):**

```
┌─────────────┐
│   AI Model  │  ← Isolated in its own bubble
│  (Claude)   │  ← Only knows training data
│             │  ← Can't access real world
└─────────────┘
```

**What AI CANNOT do:**

- ❌ Read your files
- ❌ Search the web for current info
- ❌ Access your database
- ❌ Use any tools or APIs
- ❌ Get information after training cutoff date

**Result:** AI is limited to what it learned during training. Like talking to someone from the past who can't see the present.

---

### With MCP (Tool-Enabled AI)

```
┌─────────────┐         MCP Protocol        ┌──────────────────┐
│   AI Model  │ ←─────────────────────────→ │   MCP Servers    │
│  (Claude)   │                              │  (Tools/Data)    │
│             │                              ├──────────────────┤
│ "I need to  │                              │ • Filesystem     │
│  read       │                              │ • Web Search     │
│  file.txt"  │ ─────────request────────────→ • Database       │
│             │ ←────────content─────────────│ • GitHub         │
│ "Now I can  │                              │ • Slack          │
│  answer!"   │                              │ • Custom Tools   │
└─────────────┘                              └──────────────────┘
```

**What AI CAN do with MCP:**

- ✅ Read and write files on your computer
- ✅ Search the web for current information
- ✅ Query databases for real data
- ✅ Use any tool you connect
- ✅ Access real-time information

**Result:** AI becomes a capable assistant that can actually DO things, not just talk about them.

---

### Real-World Example

**Without MCP:**

```
You: "What files are in my Documents folder?"
AI: "I can't access your filesystem. I can only provide information 
     from my training data."
```

**With MCP (Filesystem Server):**

```
You: "What files are in my Documents folder?"
AI: [Uses MCP filesystem server to list files]
    "You have 23 files in Documents:
     - proposal.pdf (modified yesterday)
     - budget.xlsx (modified 2 days ago)
     - notes.txt (modified today)
     ..."
```

---

### MCP-CLI-Go's Dual Role

MCP-CLI-Go works both ways:

**1. As MCP Client (Using Tools):**

```yaml
# Your template
steps:
  - name: search_web
    servers: [brave-search]  # ← MCP server
    prompt: "Find recent AI news"
```

AI connects TO MCP servers to use tools.

**2. As MCP Server (Being a Tool):**

```bash
# Run as server
mcp-cli serve config/runas/agent.yaml
```

Now OTHER apps (like Claude Desktop) can use YOUR workflows as tools!

**Example:** Claude Desktop can call your custom "analyze_logs" template.

---

### Why MCP Matters

**Before:** Every AI tool needs custom integrations.

```
Claude ←→ Custom code ←→ Your database
GPT-4 ←→ Different custom code ←→ Your database
Gemini ←→ Yet another custom code ←→ Your database
```

**With MCP:** Write once, works with any AI.

```
Claude  ─┐
GPT-4   ─┼→ MCP Standard ←→ Your database (one integration)
Gemini  ─┘
```

**Benefit:** Build tool once, any AI can use it.

---

## What is a Template?

A **template** is a YAML file that defines a multi-step AI workflow.

**In plain English:** Instead of manually running multiple AI queries and passing results between them, you write down all the steps once, and MCP-CLI executes them automatically.

**Like a recipe:**

- Recipe: "Mix flour and eggs → knead dough → bake 20 minutes"
- Template: "Analyze text → extract key points → create summary"

---

### Simple Template Example

```yaml
name: analyze_and_summarize
description: Analyze text then create bullet summary
version: 1.0.0

steps:
  # Step 1: Analyze the input text
  - name: analyze
    prompt: "Analyze this text for key themes and sentiment: {{stdin}}"
    output: analysis

  # Step 2: Create bullet summary from analysis
  - name: summarize
    prompt: |
      Create a 3-bullet summary from this analysis:
      {{analysis}}

      Format:
      • Bullet 1
      • Bullet 2
      • Bullet 3
    output: summary
```

---

### How It Executes

**Input:**

```
"Sales were up 15% in Q4. New customer acquisition exceeded 
targets by 23%. However, churn rate increased slightly to 5.2%."
```

**Execution Flow:**

```
User provides input
    ↓
Step 1: analyze
├─ AI receives: "Analyze this text... Sales were up 15%..."
├─ AI analyzes and generates detailed analysis
└─ Stores in {{analysis}} variable

Step 2: summarize  
├─ AI receives: "Create 3-bullet summary from: [detailed analysis]"
├─ AI creates bullets
└─ Returns final result

Final Output:
• Revenue growth strong at 15% in Q4
• Customer acquisition exceeded targets (23% over goal)
• Churn rate increased slightly to 5.2%, needs monitoring
```

**What the user sees:**

```bash
echo "Sales were up 15%..." | mcp-cli --template analyze_and_summarize

# Output:
• Revenue growth strong at 15% in Q4
• Customer acquisition exceeded targets (23% over goal)  
• Churn rate increased slightly to 5.2%, needs monitoring
```

**Behind the scenes:**

- 2 AI calls made
- Variables passed automatically
- Error handling built in
- Results combined

---

### Why Templates vs Manual Queries?

**Without Templates (Manual):**

```bash
# Step 1: Analyze (manual)
result1=$(mcp-cli query "Analyze: $data")

# Step 2: Extract (manual, must pass result1)
result2=$(mcp-cli query "Extract key points from: $result1")

# Step 3: Summarize (manual, must pass result2)
result3=$(mcp-cli query "Summarize: $result2")

# Step 4: Format (manual, must pass result3)
result4=$(mcp-cli query "Format as bullets: $result3")
```

**Problems:**

- ❌ Must manually copy-paste results between steps
- ❌ No error handling (if step 2 fails, step 3 runs anyway)
- ❌ Can't reuse (must retype every time)
- ❌ Hard to maintain (if you change step 2, must update step 3)
- ❌ No version control
- ❌ Complex logic difficult (conditions, loops)

---

**With Templates:**

```bash
echo "$data" | mcp-cli --template analyze
```

**Benefits:**

- ✅ Automatic variable flow ({{analysis}} passes automatically)
- ✅ Built-in error handling (stops if step fails)
- ✅ Reusable (save once, use forever)
- ✅ Easy to maintain (change step 2, everything else updates)
- ✅ Version controlled (commit to git)
- ✅ Complex logic easy (conditions, loops, parallel execution)
- ✅ Composable (templates can call other templates)

---

### Real-World Analogy

**Manual queries** = Cooking without a recipe:

- You remember "add flour, then eggs..."
- If you forget a step, you're stuck
- Hard to teach someone else
- Different every time

**Templates** = Following a recipe:

- Recipe written down once
- Steps always in right order
- Anyone can follow it
- Same result every time
- Can share with others

---

### Workflow Features

Templates support:

**Basic:**

- Multiple steps in sequence
- Variable passing between steps
- Different AI providers per step

**Advanced:**

- Conditions ("if category is X, do Y")
- Loops ("for each item, do this")
- Parallel execution ("do these 3 things at once")
- Template composition ("call this other template")
- Error handling ("if this fails, do fallback")
- MCP server integration ("use these tools")

**Example with conditions:**

```yaml
steps:
  - name: classify
    prompt: "Classify as: urgent, normal, low {{input_data}}"
    output: priority

  - name: urgent_handler
    condition: "{{priority}} == 'urgent'"
    prompt: "Handle urgent: {{input_data}}"

  - name: normal_handler
    condition: "{{priority}} == 'normal'"
    prompt: "Handle normal: {{input_data}}"
```

**Only ONE handler runs**, based on classification!

---

### When to Use Templates

**Use templates when:**

- ✅ You run the same workflow repeatedly
- ✅ Multiple steps needed
- ✅ Need consistent results
- ✅ Want to share workflow with team
- ✅ Production use (reliability matters)

**Don't use templates when:**

- ❌ One-off question
- ❌ Exploratory (don't know steps yet)
- ❌ Interactive conversation
- ❌ Experimenting

**For those cases, use:**

- Query mode: `mcp-cli query "question"`
- Chat mode: `mcp-cli chat`

---

## Workflow Composition

**Template composition** allows templates to call other templates, creating modular, reusable workflows.

### The Power of Composition

```yaml
name: document_intelligence
description: Complete document analysis pipeline
version: 1.0.0

steps:
  # Step 1: Call sentiment template
  - name: get_sentiment
    template: sentiment_analysis
    template_input: "{{input_data}}"
    output: sentiment

  # Step 2: Call entity extraction template
  - name: get_entities
    template: entity_extraction
    template_input: "{{input_data}}"
    output: entities

  # Step 3: Call summarization template
  - name: get_summary
    template: summarization
    template_input: "{{input_data}}"
    output: summary

  # Step 4: Combine results
  - name: final_report
    prompt: |
      Create intelligence report:
      Sentiment: {{sentiment}}
      Entities: {{entities}}
      Summary: {{summary}}
```

### Visual Representation

```
document_intelligence
├─► sentiment_analysis (executes → returns result)
├─► entity_extraction (executes → returns result)
├─► summarization (executes → returns result)
└─► Combines all results
```

### Context Isolation

Each sub-template execution is **isolated**:

```
Parent Template Context:
┌─────────────────────────────────┐
│ Variables: input_data,          │
│  sentiment,entities, summary    │
│                                 │
│ Calls: sentiment_analysis       │
│   ┌─────────────────────────┐  │
│   │ ISOLATED CONTEXT        │  │
│   │ Input: document text    │  │
│   │ Steps: analyze → score  │  │
│   │ Output: "positive"      │  │
│   └─────────────────────────┘  │
│         ↓                       │
│   Returns only: "positive"      │
│   (Not full conversation)       │
└─────────────────────────────────┘
```

**Benefits:**

- **Token Efficiency**: Only final results passed up, not full conversation
- **Modularity**: Sub-templates are reusable black boxes
- **Testing**: Test each template independently
- **Maintainability**: Update templates without affecting callers

### Recursion Depth

Templates can nest **up to 10 levels deep**:

```
Level 1: research_workflow
  └─► Level 2: web_search
       └─► Level 3: query_expansion
            └─► Level 4: generate_synonyms
                 └─► ...up to Level 10
```

This prevents infinite loops while allowing complex workflows.

---

## Multi-Provider Workflows

Use **different AI providers** for different tasks in the same workflow.

### Why Mix Providers?

Different models excel at different tasks:

- **Claude (Anthropic)**: Best for research, analysis, long documents
- **GPT-4 (OpenAI)**: Strong at structured output, code generation
- **Ollama**: Free local models for synthesis, formatting
- **DeepSeek**: Cost-effective for bulk processing
- **Gemini**: Large context windows (1M+ tokens)

### Multi-Provider Example

```yaml
name: research_pipeline
version: 1.0.0

steps:
  # Step 1: Claude researches (best for comprehensive analysis)
  - name: research
    provider: anthropic
    model: claude-sonnet-4
    prompt: "Research this topic: {{input_data}}"
    output: findings

  # Step 2: GPT-4 fact-checks (different perspective)
  - name: verify
    provider: openai
    model: gpt-4o
    prompt: "Fact-check this research: {{findings}}"
    output: verified

  # Step 3: Local model summarizes (free!)
  - name: synthesize
    provider: ollama
    model: qwen2.5:32b
    prompt: "Create summary: {{verified}}"
```

### Visual Flow

```
Input: "Impact of AI on healthcare"
    ↓
┌──────────────────────────────┐
│ Provider: Anthropic          │
│ Model: Claude Sonnet 4       │
│ Task: Deep research          │
└──────────────────────────────┘
    ↓ {{findings}}
┌──────────────────────────────┐
│ Provider: OpenAI             │
│ Model: GPT-4o                │
│ Task: Fact verification      │
└──────────────────────────────┘
    ↓ {{verified}}
┌──────────────────────────────┐
│ Provider: Ollama (Local)     │
│ Model: Qwen 2.5 32B          │
│ Task: Synthesis (no cost!)   │
└──────────────────────────────┘
    ↓
Output: Executive summary
```

### Cost Optimization

```yaml
# Example: Expensive task → Cheap verification

steps:
  # Use expensive model for complex analysis
  - name: complex_analysis
    provider: anthropic
    model: claude-opus-4  # Most capable
    prompt: "Deep analysis: {{input_data}}"
    output: analysis

  # Use cheap model for formatting
  - name: format
    provider: ollama  # Free!
    model: llama3.2
    prompt: "Format as markdown: {{analysis}}"
```

**Result:** High quality at lower cost!

---

## Variables and Data Flow

Variables connect steps together and flow data through your workflow.

### Variable Basics

```yaml
steps:
  # Step 1: Creates {{analysis}} variable
  - name: analyze
    prompt: "Analyze: {{input_data}}"
    output: analysis

  # Step 2: Reads {{analysis}} variable
  - name: summarize
    prompt: "Summarize: {{analysis}}"
    output: summary

  # Step 3: Reads {{summary}} variable
  - name: format
    prompt: "Format: {{summary}}"
```

### Built-in Variables

| Variable         | Description                     | Example                      |
| ---------------- | ------------------------------- | ---------------------------- |
| `{{stdin}}`      | Input from pipe or --input-data | `echo "text" \| mcp-cli ...` |
| `{{input_data}}` | Same as stdin                   | `--input-data "text"`        |
| `{{step_name}}`  | Output from previous step       | `{{analysis}}`               |
| `{{item}}`       | Current item in loop            | In `for_each` loops          |

### Variable Scope

```yaml
config:
  variables:
    global_var: "Available everywhere"

steps:
  - name: step1
    variables:
      local_var: "Only in this step"
    prompt: "Use {{global_var}} and {{local_var}}"

  - name: step2
    prompt: "Can use {{global_var}} but not {{local_var}}"
```

### Variable Types

Variables can be:

- **Strings**: `"hello world"`
- **Numbers**: `42`, `3.14`
- **Booleans**: `true`, `false`
- **Objects**: `{key: value}`
- **Arrays**: `[1, 2, 3]`

### Complex Variables

```yaml
steps:
  # Step creates structured output
  - name: extract_data
    prompt: "Extract name and email from: {{input_data}}"
    output: person  # Stores: {name: "John", email: "..."}

  # Access nested data
  - name: send_email
    prompt: "Send email to {{person.email}}"
```

### Conditional Variables

```yaml
steps:
  - name: classify
    prompt: "Classify as: technical, sales, support"
    output: category

  - name: route_technical
    condition: "{{category}} == 'technical'"
    prompt: "Technical response..."

  - name: route_sales
    condition: "{{category}} == 'sales'"
    prompt: "Sales response..."
```

---

## Execution Modes

MCP-CLI-Go supports multiple execution modes for different use cases.

### 1. Query Mode

**One-shot queries**, perfect for scripting and automation.

```bash
# Simple query
mcp-cli query "What is 2+2?"

# With provider
mcp-cli query --provider anthropic "Explain quantum physics"

# JSON output for parsing
mcp-cli query --json "List top 5 languages" > result.json

# In scripts
ANSWER=$(mcp-cli query "Calculate: $X + $Y")
```

**Use cases:**

- Shell scripts
- CI/CD pipelines
- Cron jobs
- API integrations

### 2. Chat Mode

**Interactive conversations** with AI and tools.

```bash
# Start chat
mcp-cli chat

# With specific provider
mcp-cli chat --provider anthropic

# With MCP servers
mcp-cli chat --server filesystem,brave-search
```

**Features:**

- Conversation history
- Multi-turn interactions
- Tool use (MCP servers)
- Interactive commands (/help, /clear, /exit)

**Use cases:**

- Exploratory analysis
- Development
- Interactive debugging

### 3. Template Mode

**Execute predefined workflows**.

```bash
# Run template
mcp-cli --template analyze

# With input
echo "data" | mcp-cli --template analyze

# List templates
mcp-cli --list-templates
```

**Use cases:**

- Repeatable workflows
- Production pipelines
- Standardized processes

### 4. Server Mode

**Run as MCP server** to expose workflows as tools.

```bash
# Start server
mcp-cli serve config/runas/agent.yaml
```

**Use cases:**

- Claude Desktop integration
- IDE integration
- Tool ecosystems

---

## Context Isolation

**Context isolation** is key to efficiency and modularity.

### Without Isolation (Traditional)

```
Step 1: 1000 tokens
    ↓ (sends full conversation)
Step 2: 1000 + 1000 = 2000 tokens
    ↓ (sends full conversation)
Step 3: 2000 + 1000 = 3000 tokens
    ↓
Total: 6000 tokens sent
```

**Problem:** Exponential token growth!

### With Isolation (MCP-CLI)

```
Step 1: 1000 tokens → Output: 50 tokens
    ↓ (only 50-token result passed)
Step 2: 50 + 1000 = 1050 tokens → Output: 50 tokens
    ↓ (only 50-token result passed)
Step 3: 50 + 1000 = 1050 tokens
    ↓
Total: 3100 tokens sent
```

**Benefit:** 50% token savings!

### Workflow Composition Isolation

```yaml
# Parent template
steps:
  - name: call_child
    template: child_template
    template_input: "{{data}}"
    output: result  # Only gets final result
```

**Child template execution:**

```
Child Template (Isolated Context)
├── Step 1: Process input (1000 tokens)
├── Step 2: Analyze (1000 tokens)
├── Step 3: Format (1000 tokens)
└── Returns: Final 100-token result

Parent Template receives: 100 tokens
Parent does NOT receive: 3000 tokens of conversation
```

**Real-world savings:** 50-87% reduction in token usage!

---

## MCP Servers vs Providers

Understanding the difference is crucial.

### AI Providers

**What they are:** AI model APIs (Claude, GPT-4, Ollama)

**What they do:**

- Generate text
- Answer questions
- Analyze content
- Create completions

**Configuration:**

```yaml
# config/providers/openai.yaml
provider_name: openai
config:
  api_key: ${OPENAI_API_KEY}
  default_model: gpt-4o
```

**Usage:**

```yaml
steps:
  - name: analyze
    provider: openai  # Which AI to use
    model: gpt-4o
    prompt: "Analyze this..."
```

### MCP Servers

**What they are:** Tools that give AI access to external systems

**What they do:**

- Read/write files
- Search the web
- Query databases
- Run code
- Custom integrations

**Configuration:**

```yaml
# config/servers/filesystem.yaml
server_name: filesystem
config:
  command: /usr/local/bin/filesystem-server
  args: []
```

**Usage:**

```yaml
steps:
  - name: search_files
    prompt: "Find Python files in /src"
    servers: [filesystem]  # Which tools AI can use
```

### How They Work Together

```
┌─────────────────────────────────┐
│         Your Template           │
├─────────────────────────────────┤
│                                 │
│  Provider: openai  ←────────┐  │
│  Model: gpt-4o             │  │
│  Servers: [filesystem]     │  │
│                           │  │
│         ↓                 │  │
│  ┌──────────────┐         │  │
│  │  GPT-4 AI    │ ←───────┘  │
│  │  "I need to  │            │
│  │   read a     │            │
│  │   file..."   │            │
│  └──────┬───────┘            │
│         │                    │
│         ↓                    │
│  ┌──────────────┐            │
│  │  MCP Server  │            │
│  │  Filesystem  │            │
│  │  reads file  │            │
│  └──────┬───────┘            │
│         │                    │
│         ↓                    │
│  Returns content to AI       │
│         ↓                    │
│  AI summarizes and returns   │
└─────────────────────────────────┘
```

**Example:**

```yaml
steps:
  - name: research
    provider: anthropic        # AI provider
    model: claude-sonnet-4
    servers: [brave-search]    # Tool server
    prompt: "Search for recent AI news"
```

**Flow:**

1. Claude receives prompt
2. Claude decides to use brave-search tool
3. MCP server executes web search
4. Results returned to Claude
5. Claude synthesizes and responds

---

## Key Takeaways

### 1. Templates are Workflows

- Multi-step AI processes in YAML
- Reusable and version-controlled
- Automatic variable flow

### 2. Composition Creates Modularity

- Templates can call templates
- Build complex workflows from simple primitives
- 50-87% token savings through context isolation

### 3. Multi-Provider Flexibility

- Use best model for each task
- Mix cloud and local models
- Optimize cost vs. quality

### 4. Variables Connect Everything

- Data flows between steps
- Scoped appropriately
- Support complex types

### 5. Multiple Modes, One Tool

- Query: One-shot automation
- Chat: Interactive exploration
- Template: Production workflows
- Server: MCP tool integration

### 6. Providers vs Servers

- **Providers**: Which AI brain to use
- **Servers**: Which tools AI can access
- Both work together seamlessly

---

## Next Steps

Now that you understand the concepts:

1. **[Quick Start Guide](../quick-start.md)** - Run your first template
2. **[Workflow Authoring Guide](../workflows/authoring-guide.md)** - Create templates
3. **[Workflow Examples](../workflows/examples/)** - Real-world examples
4. **[Provider Guide](../providers/)** - Configure AI providers

---

## Questions?

- **Need help?** [GitHub Discussions](https://github.com/LaurieRhodes/mcp-cli-go/discussions)
- **Found a bug?** [Issues](https://github.com/LaurieRhodes/mcp-cli-go/issues)
- **Want examples?** [Workflow Examples](../workflows/examples/)

Understanding these concepts unlocks the full power of MCP-CLI-Go! 🚀
