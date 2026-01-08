# Usage Guides

Learn how to use MCP-CLI effectively - whether you're chatting with AI, automating tasks, or building search systems.

**What are these guides?** Step-by-step instructions for each way to use MCP-CLI.

**Which guide should I read?**
- Want to **chat with AI**? → [Chat Mode](chat-mode.md)
- Want to **automate tasks**? → [Automation](automation.md)  
- Want to **script with AI**? → [Query Mode](query-mode.md)
- Want to **test tools**? → [Interactive Mode](interactive-mode.md)
- Want to **build search**? → [Embeddings](embeddings.md)
- **Something broken**? → [Debugging](debugging.md)

---

## Available Guides

### For Beginners

**Start here if you're new:**

1. **[Chat Mode](chat-mode.md)** - Talk with AI interactively
   - Like ChatGPT but in terminal
   - See AI use tools in real-time
   - Perfect for exploring and learning
   - **Time:** 10 minutes to learn

2. **[Query Mode](query-mode.md)** - Ask AI questions from scripts
   - One question → one answer → exits
   - Perfect for automation
   - Use in bash scripts, CI/CD
   - **Time:** 15 minutes to learn

3. **[Debugging](debugging.md)** - Fix problems when things break
   - Find and fix issues
   - Understand error messages
   - Get verbose output
   - **Time:** 5 minutes when needed

---

### For Automation

**Read these when building automated workflows:**

4. **[Automation & Scripting](automation.md)** - Build production workflows
   - Create templates (reusable AI workflows)
   - CI/CD integration (GitHub Actions, GitLab)
   - Scheduled tasks (cron jobs)
   - **Time:** 30 minutes to master

5. **[Interactive Mode](interactive-mode.md)** - Test tools without AI
   - Call MCP server tools directly
   - Debug tool issues
   - Inspect tool capabilities
   - **Time:** 10 minutes to learn

---

### For Advanced Use Cases

**For specialized applications:**

6. **[Embeddings](embeddings.md)** - Build semantic search & RAG
   - Generate vectors from text
   - Build similarity search
   - Create RAG systems
   - **Time:** 20 minutes to learn

---

## Quick Mode Comparison

**Which mode should I use?** Depends what you're doing:

| Mode | What It Is | When To Use | Example |
|------|------------|-------------|---------|
| **Chat** | Conversation with AI | Exploring, developing, learning | `mcp-cli chat` |
| **Query** | One question → one answer | Scripts, automation, CI/CD | `mcp-cli query "question"` |
| **Template** | Multi-step workflow | Reusable processes | `mcp-cli --template name` |
| **Interactive** | Test tools directly | Tool debugging, testing | `mcp-cli interactive --server fs` |
| **Embeddings** | Generate vectors | Search, RAG, clustering | `mcp-cli embeddings --input file.txt` |

### Detailed Comparison

| Feature | Chat Mode | Query Mode | Template Mode | Interactive Mode |
|---------|-----------|------------|---------------|------------------|
| **AI involved?** | ✅ Yes | ✅ Yes | ✅ Yes | ❌ No (just tools) |
| **Conversation history?** | ✅ Yes | ❌ No | ❌ No | ❌ No |
| **Can use tools?** | ✅ Auto | ✅ Auto | ✅ Auto | ✅ Manual |
| **Good for scripts?** | ❌ No | ✅ Yes | ✅ Yes | ✅ Yes |
| **Cost per use** | ~$0.001 | ~$0.001 | ~$0.01-0.10 | $0 (no AI) |
| **Best for** | Exploring | Automation | Workflows | Testing |

---

## Choosing the Right Mode

### Use Chat Mode When:
- ✅ **Exploring ideas** - Don't know exactly what you need yet
- ✅ **Multi-turn conversations** - Each answer leads to next question
- ✅ **Need context** - AI should remember what you said before
- ✅ **Developing templates** - Testing prompts interactively
- ✅ **Learning** - New to MCP-CLI, want to see how it works

**Example:**
```bash
mcp-cli chat --provider anthropic

You> What files are in this directory?
AI> [uses filesystem tool, lists files]

You> Read the README.md file
AI> [reads and summarizes README]

You> What's the main purpose of this project?
AI> [answers based on README content]
```

**Why this works:** AI remembers the conversation!

**Cost:** ~$0.001 per message (~1/10th of a penny)

**See:** [Chat Mode Guide](chat-mode.md)

---

### Use Query Mode When:
- ✅ **Automation** - Running in scripts, cron jobs, CI/CD
- ✅ **One-off questions** - Single question, don't need conversation
- ✅ **Processing files** - Piped input (cat file | mcp-cli)
- ✅ **Need JSON output** - Parsing results in scripts
- ✅ **No conversation needed** - Each query independent

**Example:**
```bash
# In a script
SUMMARY=$(cat document.txt | mcp-cli query "Summarize in 3 bullets")
echo "$SUMMARY" > summary.txt

# In CI/CD
git diff | mcp-cli query "Review this code" > review.md

# Cron job (daily report)
./mcp-cli query "Summarize today's logs" < /var/log/app.log | mail -s "Daily Report" team@example.com
```

**Why this works:** Simple, scriptable, no conversation state.

**Cost:** ~$0.001 per query

**See:** [Query Mode Guide](query-mode.md)

---

### Use Templates When:
- ✅ **Multi-step workflows** - Need several AI calls in sequence
- ✅ **Reusable processes** - Same workflow repeatedly
- ✅ **Production automation** - Reliable, tested workflows
- ✅ **Version control** - Workflows as code
- ✅ **Team sharing** - Others need same workflow

**Example:**
```yaml
# config/workflows/code-review.yaml
name: code_review
steps:
  - name: analyze
    prompt: "Review code: {{stdin}}"
  - name: format
    prompt: "Format as markdown: {{analyze}}"
```

```bash
# Use it
git diff | mcp-cli --template code_review > review.md

# Reuse it
cat feature.py | mcp-cli --template code_review
cat bugfix.go | mcp-cli --template code_review
```

**Why this works:** Write once, use everywhere!

**Cost:** ~$0.01-0.10 depending on complexity

**See:** [Automation Guide](automation.md) and [Workflow Authoring](../workflows/authoring-guide.md)

---

### Use Interactive Mode When:
- ✅ **Testing MCP servers** - Is tool working correctly?
- ✅ **Debugging tools** - Why is tool failing?
- ✅ **Learning tools** - What can this tool do?
- ✅ **No AI needed** - Just want to call tools directly
- ✅ **Exploring new servers** - What tools does this server provide?

**Example:**
```bash
mcp-cli interactive --server filesystem

> /tools
Available tools:
  - read_file
  - write_file
  - list_directory

> /call filesystem list_directory {"path": "."}
Result: ["README.md", "main.go", "config.yaml"]

> /call filesystem read_file {"path": "README.md"}
Result: "# My Project\n..."
```

**Why this works:** Direct tool access, no AI in the way.

**Cost:** $0 (no AI calls!)

**See:** [Interactive Mode Guide](interactive-mode.md)

---

### Use Embeddings Mode When:
- ✅ **Building search** - Semantic search, not keyword search
- ✅ **RAG systems** - Retrieval Augmented Generation
- ✅ **Document clustering** - Group similar documents
- ✅ **Recommendations** - "More like this" features
- ✅ **Similarity detection** - Find duplicates, similar content

**Example:**
```bash
# Generate embeddings for documents
mcp-cli embeddings \
  --input-file documents.txt \
  --output-file vectors.json \
  --provider openai

# Now use in search system
cat vectors.json | python search-system.py "query"
```

**Why this works:** Vectors enable semantic understanding.

**Cost:** ~$0.0001 per 1000 words embedded

**See:** [Embeddings Guide](embeddings.md)

---

## Common Workflows

### Development Workflow
1. **Interactive mode** - Test MCP server tools manually
2. **Chat mode** - Explore AI + tool interactions
3. **Templates** - Capture reusable patterns
4. **Query mode** - Automate in scripts

### Production Workflow
1. **Templates** - Define workflows in YAML
2. **Query mode** - Execute via scripts/CI/CD
3. **Server mode** - Expose as MCP tools (coming soon)

### Search/RAG Workflow
1. **Embeddings** - Generate vectors from documents
2. **Vector DB** - Store and index vectors
3. **Query mode** - Search + RAG with templates
4. **Chat mode** - Interactive Q&A with context

---

## Quick Start by Use Case

### I want to...

**...have a conversation with AI**
```bash
mcp-cli chat --provider anthropic
```
→ See [Chat Mode Guide](chat-mode.md)

**...automate a task**
```bash
# Create template first
cat > config/workflows/my-task.yaml << EOF
name: my_task
steps:
  - name: step1
    prompt: "..."
EOF

# Then use it
./mcp-cli --template my_task
```
→ See [Automation Guide](automation.md)

**...test an MCP server**
```bash
mcp-cli interactive --server my-server
> /tools
> /call my-server tool_name {"arg": "value"}
```
→ See [Interactive Mode Guide](interactive-mode.md)

**...build semantic search**
```bash
# Generate embeddings
mcp-cli embeddings --input-file docs.txt --output-file vectors.json

# Use in RAG
cat vectors.json | python rag-system.py
```
→ See [Embeddings Guide](embeddings.md)

**...debug an issue**
```bash
mcp-cli --verbose query "Test"
```
→ See [Debugging Guide](debugging.md)

---

## Next Steps

Choose your guide based on your goal:

**New to MCP-CLI?**
1. [Chat Mode](chat-mode.md) - Start here
2. [Core Concepts](../getting-started/concepts.md) - Understand the system
3. [Templates](../workflows/authoring-guide.md) - Create workflows

**Building automation?**
1. [Automation Guide](automation.md) - Best practices
2. [Query Mode](query-mode.md) - Scripting reference
3. [Workflow Examples](../workflows/examples/) - Working examples

**Testing or developing?**
1. [Interactive Mode](interactive-mode.md) - Test tools
2. [Debugging](debugging.md) - Troubleshoot issues
3. [Chat Mode](chat-mode.md) - Explore interactively

**Building search/RAG?**
1. [Embeddings Guide](embeddings.md) - Generate vectors
2. [Automation Guide](automation.md) - Batch processing
3. [Query Mode](query-mode.md) - RAG queries

---

## Contributing

Found a useful pattern or workflow? Share it in [Discussions](https://github.com/LaurieRhodes/mcp-cli-go/discussions/categories/show-and-tell)!

---

## All Guides

1. **[Chat Mode](chat-mode.md)** - Interactive AI conversations
2. **[Query Mode](query-mode.md)** - One-shot automation
3. **[Interactive Mode](interactive-mode.md)** - Direct tool testing
4. **[Automation](automation.md)** - Templates, CI/CD, workflows
5. **[Embeddings](embeddings.md)** - Vector generation & search
6. **[Debugging](debugging.md)** - Troubleshooting

---

**Ready to get started?** Pick a guide and dive in! 🚀
