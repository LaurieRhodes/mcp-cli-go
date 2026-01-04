# Chat Mode Guide

Have conversations with AI - like ChatGPT but in your terminal, with access to tools.

![](./img/Chat%20Mode.jpeg)



**What is Chat Mode?** Interactive back-and-forth conversation with AI that remembers what you said.

**Think of it like:** Texting with AI. You say something, AI responds, you continue the conversation.

**Difference from Query Mode:**

- **Query Mode:** One question → one answer → exits (like a function call)
- **Chat Mode:** Ongoing conversation → remembers context → keeps going (like messaging)

**Use when:**

- ✅ Exploring ideas (don't know exactly what you need yet)
- ✅ Multi-step problem solving (each answer leads to next question)
- ✅ Testing templates before automation
- ✅ Need AI to use tools (file access, web search)
- ✅ Interactive development

---

## Table of Contents

- [Quick Start](#quick-start)
- [Starting Chat](#starting-chat)
- [Chat Commands](#chat-commands)
- [Using Tools](#using-tools)
- [Context Management](#context-management)
- [Advanced Features](#advanced-features)
- [Tips & Best Practices](#tips--best-practices)

---

## Quick Start

**Start a chat:**

```bash
mcp-cli chat
```

**You'll see:**

```
Starting chat mode...
Provider: anthropic, Model: claude-sonnet-4

Welcome to MCP CLI Chat!
Type your message and press Enter.
Commands: /help, /exit, /clear

You>
```

**Type and press Enter:**

```
You> Hello! What can you help me with?

Hello! I can help you with many things:
• Answer questions
• Write code
• Analyze data
• Work with files on your computer (via tools)
• Search the web
• And much more!

What would you like to work on?

You>
```

**The conversation continues!** AI remembers everything you said.

**To exit:** Type `/exit` or press Ctrl+C

---

## Starting Chat

### Basic Start (Default Settings)

```bash
mcp-cli chat
```

**What happens:**

1. Loads config.yaml to find default provider
2. Connects to AI provider (e.g., Anthropic, OpenAI etc)
3. Starts conversation (all exposed MCP tools unless configured)
4. Waits for your input

**Cost:** Free until you type something! Then ~$0.001 per message.

---

### Start with Specific Provider

**Why:** Different AIs have different personalities and strengths.

```bash
# Use Claude (best for analysis, long documents)
mcp-cli chat --provider anthropic

# Use GPT-4 (best for code, structured output)
mcp-cli chat --provider openai

# Use local model (free!)
mcp-cli chat --provider ollama
```

**What `--provider` changes:**

- Which AI you're talking to
- Cost per message
- Context window size (how much it remembers)
- Capabilities

---

### Start with Specific Model

**Why:** Newer/bigger models are smarter but cost more.

```bash
# Latest Claude (smartest)
mcp-cli chat --provider anthropic --model claude-sonnet-4

# Cheaper GPT (60% cheaper)
mcp-cli chat --provider openai --model gpt-4o-mini

# Specific local model
mcp-cli chat --provider ollama --model llama3.2
```

---

### Start with Tools (MCP Servers)

**What are tools?** Give AI ability to DO things (read files, search web, etc.).  All configured tools will be loaded by default or specific singular tools may be enabled from the commandline.

```bash
# AI can access files
mcp-cli chat --server filesystem

# AI can search web
mcp-cli chat --server brave-search

# AI can use BOTH
mcp-cli chat --server filesystem,brave-search
```

**What you'll see:**

```
Starting chat mode...
Connecting to servers: filesystem, brave-search

Connected servers:
  • filesystem - File operations
  • brave-search - Web search

You>
```

**Now AI can:**

```
You> What's in my Documents folder?

[AI uses filesystem tool to list directory]

You have 23 files in Documents:
• proposal.pdf
• budget.xlsx
...

You> Search the web for "MCP protocol" and save results to a file

[AI uses brave-search tool, then filesystem tool]

I found information about MCP protocol and saved it to mcp-info.txt.
```

---

## Chat Commands

Commands let you control the chat session. All start with `/`.

**Quick reference:**

- `/help` - Show all commands
- `/exit` - Leave chat
- `/clear` - Start fresh (forget history)
- `/tools` - See what AI can do
- `/history` - See conversation so far
- `/context` - Check token usage

---

### /help - Show Available Commands

**What it does:** Lists all commands and what they do.

```
You> /help
```

**Output:**

```
Available commands:
  /help     - Show this help message
  /exit     - Exit chat mode
  /quit     - Exit chat mode (same as /exit)
  /clear    - Clear conversation history
  /tools    - List available MCP tools
  /history  - Show full conversation
  /context  - Show context statistics
  /model    - Switch AI model
```

**When to use:** Forgot what commands exist

---

### /exit or /quit - Leave Chat

**What it does:** Closes chat mode, returns to terminal.

```
You> /exit
Exiting chat mode. Goodbye!
```

**Keyboard shortcut:** Press `Ctrl+C` (same effect)

**What happens to history:** Gone! (not saved anywhere)

**If you want to save conversation:**

```
You> Summarize our conversation and save to summary.txt

[AI saves summary]

You> /exit
```

---

### /clear - Forget Everything

**What it does:** Erases conversation history, starts fresh.

```
You> /clear
Chat history cleared. Starting fresh!
```

**What's forgotten:**

- All previous messages
- All context about what you discussed
- Tool results from earlier

**What's kept:**

- Which AI provider/model
- Which tools are connected
- Your session (chat stays open)

**When to use:**

- ✅ Switching to completely different topic
- ✅ Context getting too long (slowing down responses)
- ✅ AI seems "confused" from earlier context
- ✅ Starting a new task

**When NOT to use:**

- ❌ Still working on same problem (you'll lose progress!)
- ❌ Just to "reset" AI behavior (doesn't work that way)

**Real example:**

```
You> Help me debug this Python code
[20 messages back and forth about Python]

You> /clear
Chat history cleared.

You> Now help me write a SQL query
[Fresh start - AI doesn't remember Python context]
```

**Cost saving:** Clearing frequently can reduce costs (less context = cheaper messages)

---

### /tools - See What AI Can Do

**What it does:** Lists all MCP tools AI has access to.

```
You> /tools
```

**Example output:**

```
Available tools from connected servers:

Server: filesystem
  • read_file - Read contents of a file
    Parameters: path (string)

  • write_file - Write content to a file
    Parameters: path (string), content (string)

  • list_directory - List files in directory
    Parameters: path (string)

  • search_files - Search for files by name
    Parameters: pattern (string), path (string)

Server: brave-search
  • web_search - Search the web
    Parameters: query (string)
```

**When to use:**

- Curious what AI can do with tools
- Forgot exact tool capabilities
- Debugging ("why isn't AI using X tool?")

**Note:** AI automatically decides when to use tools. You can't call them directly.

---

### /history - See Full Conversation

**What it does:** Shows all messages in current session.

```
You> /history
```

**Example output:**

```
Chat History (5 messages):

[1] User:
What files are in this directory?

[2] Assistant:
Let me check that for you.
[Used tool: list_directory]

[3] Tool Result (call_abc123):
{"files": ["README.md", "main.go", "config.yaml"]}

[4] Assistant:
The current directory contains:
• README.md
• main.go  
• config.yaml

[5] User:
What's in README.md?
```

**When to use:**

- Review what you discussed
- Find something AI said earlier
- Copy-paste previous output
- Debug why AI is confused

**Long conversations:** History gets truncated to fit terminal

---

### /context - Check Your Usage

**What it does:** Shows detailed statistics about conversation context.

```
You> /context
```

**Example output:**

```
Context Statistics:
  Provider: anthropic
  Model: claude-sonnet-4

  Conversation:
    Messages: 12
    Your messages: 6
    AI messages: 6
    Tool calls: 3

  Tokens:
    Current usage: 2,450 tokens
    Maximum limit: 200,000 tokens
    Reserved: 4,000 tokens (for response)
    Effective limit: 196,000 tokens
    Utilization: 1.2%

  Status: ✓ Plenty of room remaining
```

**What tokens are:** Words/pieces of text. Roughly: 1 token ≈ 0.75 words

**Why it matters:**

- Each AI model has a token limit
- When you hit limit, old messages get dropped
- More tokens = higher cost

**Token limits by provider:**

- Claude Sonnet 4: 200,000 tokens (~150,000 words)
- GPT-4o: 128,000 tokens (~96,000 words)
- Ollama Qwen: 32,768 tokens (~24,000 words)

**When to /clear:**

- Utilization > 50% and switching topics
- Responses getting slow (too much context)
- Want to save costs (less context = cheaper)

**Real example:**

```
You> /context
Context Statistics:
  Current usage: 185,000 tokens
  Maximum limit: 200,000 tokens
  Utilization: 92.5%

Status: ⚠️ Approaching limit! Consider /clear

You> /clear
You> /context
Context Statistics:
  Current usage: 0 tokens
  Utilization: 0%

Status: ✓ Fresh start!
```

---

## Using Tools

Chat mode automatically uses tools when the AI decides they're needed.

### Automatic Tool Execution

**You don't call tools directly** - the AI decides when to use them.

**Example:**

```
You> What files are in the current directory?

Thinking...
Executing tool calls...
Tool: list_directory (filesystem)
  Result: Found 15 files

The current directory contains the following files:
1. README.md
2. config.yaml
3. main.go
...
```

### Tool Execution Flow

```
1. User asks question
   ↓
2. AI decides tool is needed
   ↓
3. MCP-CLI executes tool on server
   ↓
4. Tool result returned to AI
   ↓
5. AI generates response using result
```

### Multi-Step Tool Usage

AI can use multiple tools in sequence:

```
You> Search for recent AI news and save the results to a file

Thinking...
Executing tool calls...
Tool: search (brave-search)
  Result: Found 10 articles about AI

Executing additional tool calls...
Tool: write_file (filesystem)
  Path: ai_news.txt
  Result: File written successfully

I've searched for recent AI news and saved the results to ai_news.txt.
```

---

## Context Management

Chat mode maintains conversation context automatically.

### How Context Works

```
Message 1: "What's the weather?"
   ↓ (stored in context)
Message 2: "How about tomorrow?"
   ↓ (AI knows "How about" refers to weather)
AI understands: "What's the weather tomorrow?"
```

### Token Limits

Different providers have different context limits:

| Provider  | Model           | Max Tokens |
| --------- | --------------- | ---------- |
| Anthropic | Claude Sonnet 4 | 200,000    |
| OpenAI    | GPT-4o          | 128,000    |
| Ollama    | Qwen 2.5 32B    | 32,768     |

# 

---

## Advanced Features

### Streaming Responses

Responses stream in real-time (like ChatGPT):

```
You> Write a story about AI

Thinking...

Once upon a time, in a world not too different from our own,
[response continues streaming...]
```

**Benefits:**

- See progress immediately
- Faster perceived response time

---

## Quick Reference

```bash
# Start chat
mcp-cli chat

# With options
mcp-cli chat --provider anthropic --model claude-sonnet-4

# Commands
/help      # Show help
/tools     # List tools
/history   # Show history
/context   # Show stats
/clear     # Clear history
/exit      # Exit
```

---

## Next Steps

- **[Query Mode](query-mode.md)** - Learn one-shot automation
- **[Automation Guide](automation.md)** - Script with chat mode
- **[Debugging](debugging.md)** - Troubleshoot issues

---

**Ready to explore?** Start chat mode and try asking questions! 🚀
