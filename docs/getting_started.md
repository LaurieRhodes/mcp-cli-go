# Getting Started with MCP CLI

Welcome to MCP CLI - a powerful command-line interface for the Model Context Protocol that connects Large Language Models with external tools and data sources.

## 🚀 Quick Start (5 Minutes)

### Step 1: Installation

#### Build from Source

```bash
git clone https://github.com/LaurieRhodes/mcp-cli-go.git
cd mcp-cli-go
go build -o mcp-cli
```

### Step 2: Set Up Your First AI Provider

Choose one of these popular options:

#### OpenAI (Recommended for Beginners)

```bash
export OPENAI_API_KEY="your-api-key-here"
```

#### Anthropic Claude

```bash
export ANTHROPIC_API_KEY="your-api-key-here"  
```

#### Google Gemini

```bash
export GEMINI_API_KEY="your-api-key-here"
```

#### Local with Ollama (No API Key Required)

```bash
# Install Ollama first: https://ollama.ai
ollama pull llama3.1:8b
```

### Step 3: Your First Query

```bash
# Simple question
mcp-cli query "What is 2 + 2?"

# With specific provider
mcp-cli query --provider openai "What is the current time?"
```

**🎉 Success!** You just made your first MCP CLI query.

## 📋 Common First Steps

### Chat Mode - Interactive Conversations

```bash
# Start chatting with default provider
mcp-cli chat

# Chat with specific provider and model
mcp-cli chat --provider anthropic --model claude-3-5-sonnet-20240620
```

Chat commands you can use:

- Type naturally to chat with the AI
- `/help` - Show available commands
- `/tools` - List available tools
- `/exit` - Leave chat mode

### Query Mode - Perfect for Automation

```bash
# Get JSON output for scripts
mcp-cli query --json "List the files in the current directory"

# Save output to file
mcp-cli query --output report.txt "Analyze the current project structure"

# Quiet mode for clean automation
mcp-cli query --provider gemini "What's the weather like?"
```

### Interactive Mode - Direct Server Control

```bash
# Direct server interaction
mcp-cli interactive

# Then use slash commands like:
# /tools - List available tools
# /call filesystem list_directory {"path": "."}
```

## 🛠️ Setting Up Tools (MCP Servers)

MCP CLI becomes much more powerful when connected to servers that provide tools. Here are some popular options:

### Filesystem Access

```bash
# Install filesystem server
npm install -g @modelcontextprotocol/server-filesystem

# Create basic configuration
cat > server_config.json << 'EOF'
{
  "servers": {
    "filesystem": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "/path/to/your/directory"]
    }
  },
  "ai": {
    "default_provider": "openai"
  }
}
EOF
```

### Web Search with Brave

```bash
# Get a free API key from https://brave.com/search/api/
# Install brave search server
npm install -g @modelcontextprotocol/server-brave-search

# Add to your server_config.json
{
  "servers": {
    "brave-search": {
      "command": "npx", 
      "args": ["-y", "@modelcontextprotocol/server-brave-search"],
      "env": {"BRAVE_API_KEY": "your-brave-api-key"}
    }
  }
}
```

### Test Your Setup

```bash
# Query with tools
mcp-cli query "What files are in the current directory and what do they contain?"

# The AI will automatically use the filesystem tool to answer!
```

## 🎯 Choose Your Path

### For Automation & Scripting

Start with **Query Mode**:

```bash
# Script-friendly examples
mcp-cli query --json --output results.json "Analyze project files"
mcp-cli query --error-code-only "Check system status"
```

📖 **Next**: Read the [Query Mode Guide](docs/features/QUERY_MODE.md)

### For Interactive Conversations

Start with **Chat Mode**:

```bash
mcp-cli chat --provider anthropic
```

📖 **Next**: Read the [Chat Mode Guide](docs/features/CHAT_MODE.md)

### For System Administration

Start with **Interactive Mode**:

```bash
mcp-cli interactive --verbose
```

📖 **Next**: Read the [Interactive Mode Guide](docs/features/INTERACTIVE_MODE.md)

## 📚 Complete Configuration Example

Create `server_config.json` with multiple providers and servers:

```json
{
  "servers": {
    "filesystem": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "."]
    },
    "brave-search": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-brave-search"],
      "env": {"BRAVE_API_KEY": "your-brave-key"}
    }
  },
  "ai": {
    "default_provider": "openai",
    "interfaces": {
      "openai_compatible": {
        "providers": {
          "openai": {
            "api_key": "your-openai-key",
            "default_model": "gpt-4o"
          }
        }
      },
      "anthropic_native": {
        "providers": {
          "anthropic": {
            "api_key": "your-anthropic-key", 
            "default_model": "claude-3-5-sonnet-20240620"
          }
        }
      }
    }
  }
}
```

Test your complete setup:

```bash
mcp-cli query "Search for recent news about AI and save a summary to ai-news.md"
```

## 🆘 Need Help?

### Quick Troubleshooting

**Problem**: `command not found: mcp-cli`
**Solution**: Make sure the binary is in your PATH or use `./mcp-cli`

**Problem**: `API key is required`  
**Solution**: Set environment variable or add to config file

**Problem**: `no servers configured`
**Solution**: Create `server_config.json` with at least one server

**Problem**: Tool execution fails
**Solution**: Verify server command paths and ensure dependencies are installed

### Get More Help

- 📖 [Complete User Guide](docs/USER_GUIDE.md) - Comprehensive usage documentation
- ⚙️ [Configuration Guide](docs/CONFIGURATION.md) - Detailed configuration options  
- 🤖 [Provider Guides](docs/providers/) - AI provider specific documentation
- 🐛 [Troubleshooting Guide](docs/TROUBLESHOOTING.md) - Common issues and solutions
- 🏗️ [Architecture Guide](docs/ARCHITECTURE.md) - Technical deep dive

### Community

- 🐛 [Report Issues](https://github.com/LaurieRhodes/mcp-cli-go/issues)
- 💬 [Discussions](https://github.com/LaurieRhodes/mcp-cli-go/discussions)  
- 📖 [Documentation](https://github.com/LaurieRhodes/mcp-cli-go/tree/main/docs)

## 🚀 What's Next?

You're now ready to explore the full power of MCP CLI! Here are some popular next steps:

1. **Set up multiple AI providers** for different use cases
2. **Add more MCP servers** for expanded tool capabilities  
3. **Create automation scripts** using query mode
4. **Explore advanced configuration** options
5. **Contribute to the project** by adding new features

**Happy automating!** 🎉

---

*Need something specific? Jump to the [Complete User Guide](docs/USER_GUIDE.md) or [Configuration Reference](docs/CONFIGURATION.md).*
