# Interactive Mode Guide

Test MCP tools directly - no AI involved, just you calling tools manually.

**What is Interactive Mode?** A command-line interface where YOU call tools directly, without AI.

**Think of it like:** A debugging console for MCP servers.

**The difference:**
- **Chat Mode:** You ask AI → AI decides to use tools → AI calls tools → AI shows you results
- **Interactive Mode:** You directly call tools → Tool executes → You see raw results

**Use when:**
- ✅ Testing if MCP server works
- ✅ Learning what tools do
- ✅ Debugging tool problems
- ✅ Understanding tool parameters
- ✅ Exploring new MCP servers

**Don't use when:**
- ❌ Want AI to help you (use chat mode)
- ❌ Automating tasks (use query/template mode)
- ❌ Building workflows (use templates)

**Cost:** $0 (no AI calls!)

---

## Table of Contents

- [Quick Start](#quick-start)
- [Starting Interactive Mode](#starting-interactive-mode)
- [Available Commands](#available-commands)
- [Calling Tools](#calling-tools)
- [Use Cases](#use-cases)
- [Best Practices](#best-practices)

---

## Quick Start

**Start interactive mode with filesystem server:**

```bash
mcp-cli interactive --server filesystem
```

**You'll see:**
```
Starting interactive mode with server: filesystem

Connected to servers:
  • filesystem

Interactive Mode - Type '/help' for commands or '/exit' to quit
>
```

**List available tools:**
```
> /tools
Tools from server filesystem:
  • read_file - Read contents of a file
  • list_directory - List directory contents
  • write_file - Write content to a file
```

**Call a tool:**
```
> /call filesystem list_directory {"path": "."}
Calling tool 'list_directory' on server 'filesystem'...

Tool execution successful!
Result:
{
  files: [
    "README.md",
    "main.go",
    "config.yaml"
  ]
}
```

**What just happened:**
1. YOU called the tool (not AI)
2. Tool executed on MCP server
3. Raw results returned (no AI interpretation)

**That's interactive mode!**

---

## Starting Interactive Mode

### With Single Server

```bash
# Filesystem server (file operations)
mcp-cli interactive --server filesystem

# Brave search (web search)
mcp-cli interactive --server brave-search

# Any server you have configured
mcp-cli interactive --server my-server
```

**What happens:**
1. MCP-CLI connects to specified server
2. Interactive prompt appears
3. You can now call tools manually

---

### With Multiple Servers

```bash
# Connect to multiple servers
mcp-cli interactive --server filesystem,brave-search
```

**You'll see:**
```
Connected to servers:
  • filesystem (5 tools)
  • brave-search (1 tool)

Interactive Mode
>
```

**Now can call tools from either server:**
```
> /call filesystem list_directory {"path": "."}
> /call brave-search search {"query": "MCP"}
```

---

### No Server Specified

```bash
# Start without server
mcp-cli interactive
```

**Result:**
```
Error: No MCP servers specified
Use: mcp-cli interactive --server <server_name>
```

**Why:** Interactive mode needs at least one server to interact with.

---

## Available Commands

All commands start with `/`.

### /help - Show All Commands

```
> /help
```

**Output:**
```
Available Commands:
  /help         - Show this help message
  /ping         - Test if server is responsive  
  /tools        - List all available tools
  /tools-all    - Show detailed tool information
  /tools-raw    - Show raw JSON tool definitions
  /call <server> <tool> <args> - Call a tool
  /clear, /cls  - Clear the screen
  /exit, /quit  - Exit interactive mode
```

---

### /ping - Test Connection

**What it does:** Checks if MCP server is responding.

```
> /ping
```

**Output if working:**
```
Pong! Server is responsive.
```

**Output if broken:**
```
Error: Server not responding
Connection timeout after 5s
```

**Use when:**
- Server seems frozen
- Want to verify connection
- Debugging connectivity issues

---

### /tools - List Available Tools

**What it does:** Shows all tools from connected servers.

```
> /tools
```

**Example output:**
```
Tools from server filesystem:
  • read_file - Read contents of a file
  • write_file - Write content to a file
  • list_directory - List directory contents
  • search_files - Search for files matching pattern
  • get_file_info - Get file metadata
  • create_directory - Create a new directory
  • move_file - Move or rename a file
  • delete_file - Delete a file

Tools from server brave-search:
  • search - Search the web
```

**Use this:** To see what tools you can call.

---

### /tools-all - Detailed Tool Information

**What it does:** Shows parameters each tool accepts.

```
> /tools-all
```

**Example output:**
```
Tools from server filesystem:

  read_file:
    Description: Read contents of a file
    Parameters:
      • path (string, required): Path to the file to read
        
  write_file:
    Description: Write content to a file
    Parameters:
      • path (string, required): Path to the file
      • content (string, required): Content to write
      
  list_directory:
    Description: List directory contents
    Parameters:
      • path (string, required): Directory path to list
```

**Use this:** To understand what arguments each tool needs.

---

### /tools-raw - Raw JSON Schema

**What it does:** Shows complete tool definitions as JSON.

```
> /tools-raw
```

**Example output:**
```json
{
  "tools": [
    {
      "name": "read_file",
      "description": "Read contents of a file",
      "inputSchema": {
        "type": "object",
        "properties": {
          "path": {
            "type": "string",
            "description": "Path to the file to read"
          }
        },
        "required": ["path"]
      }
    }
  ]
}
```

**Use this:** 
- Understanding exact types (string vs number vs boolean)
- Copying tool definitions
- Debugging schema validation issues

---

### /clear or /cls - Clear Screen

```
> /clear
```

**Result:** Terminal cleared, fresh prompt.

**Use when:** Screen gets cluttered with output.

---

### /exit or /quit - Exit Interactive Mode

```
> /exit
Exiting interactive mode.
```

**Returns to:** Normal terminal prompt.

---

## Calling Tools

### Single-Line Format

**Syntax:**
```
/call <server_name> <tool_name> <json_arguments>
```

**Example - List directory:**
```
> /call filesystem list_directory {"path": "."}
```

**What happens:**
1. Parses server name: `filesystem`
2. Parses tool name: `list_directory`
3. Parses arguments: `{"path": "."}`
4. Validates arguments against tool schema
5. Calls tool on MCP server
6. Returns raw result

**Output:**
```
Calling tool 'list_directory' on server 'filesystem'...

Tool execution successful!
Result:
{
  files: [
    "README.md",
    "main.go",
    "config.yaml"
  ]
}
```

---

### More Examples

**Read a file:**
```
> /call filesystem read_file {"path": "README.md"}

Result:
{
  content: "# My Project\n\nThis is a README..."
}
```

**Get file info:**
```
> /call filesystem get_file_info {"path": "config.yaml"}

Result:
{
  size: 1234,
  modified: "2024-12-26T10:30:00Z",
  type: "file",
  permissions: "644"
}
```

**Web search:**
```
> /call brave-search search {"query": "MCP protocol"}

Result:
{
  results: [
    {
      title: "Model Context Protocol",
      url: "https://...",
      snippet: "MCP is an open protocol..."
    }
  ]
}
```

---

### Multi-Line Format (For Complex Arguments)

**Use when:** Arguments are complex or have multiple lines.

**Syntax:**
```
> /call <server_name>
Enter tool name: <tool_name>
Enter JSON arguments (end with a line containing only '###'):
<your json here>
<can be multiple lines>
###
```

**Example:**
```
> /call filesystem
Enter tool name: write_file
Enter JSON arguments (end with a line containing only '###'):
{
  "path": "test.txt",
  "content": "Hello, world!\nThis is line 2.\nThis is line 3."
}
###

Tool execution successful!
Result:
{
  success: true,
  message: "File written successfully"
}
```

**When to use multi-line:**
- ✅ Content with newlines
- ✅ Complex nested JSON
- ✅ Large parameter values
- ✅ Better readability

---

### Argument Validation

**Interactive mode validates your arguments before calling the tool.**

**Missing required parameter:**
```
> /call filesystem read_file {}

Error: Argument validation failed:
  • Missing required parameter: path
```

**Wrong type:**
```
> /call filesystem read_file {"path": 123}

Error: Argument validation failed:
  • Parameter 'path' must be a string (got number)
```

**Invalid JSON:**
```
> /call filesystem read_file {path: "test.txt"}

Error: Invalid JSON:
  • Syntax error at position 1
  • Hint: Use double quotes for keys and strings
```

**Correct format:**
```
> /call filesystem read_file {"path": "test.txt"}
✓ Validation passed
[Tool executes]
```

---

## Use Cases

### Use Case 1: Testing New MCP Server

```bash
# Start interactive mode
mcp-cli interactive --server my-new-server

# Explore available tools
> /tools

# Check detailed parameters
> /tools-all

# Test each tool
> /call my-new-server tool_name {"param": "value"}

# Verify responses
# Debug issues
```

### Use Case 2: Understanding Tool Behavior

```
# See exact parameter requirements
> /tools-all

# Try different inputs
> /call filesystem search_files {"pattern": "*.go"}
> /call filesystem search_files {"pattern": "*.txt", "path": "./docs"}

# Observe output formats
```

### Use Case 3: Debugging Tool Issues

```
# Get raw schema
> /tools-raw

# Verify parameter types
# Test edge cases
> /call filesystem read_file {"path": "/nonexistent/file.txt"}

# Check error messages
Tool execution failed: file not found

# Validate server responses
```

### Use Case 4: Manual File Operations

```
# Check what's in directory
> /call filesystem list_directory {"path": "."}

# Read configuration
> /call filesystem read_file {"path": "config.yaml"}

# Create file
> /call filesystem write_file {"path": "test.txt", "content": "Test"}

# Verify creation
> /call filesystem get_file_info {"path": "test.txt"}
```

### Use Case 5: Exploring API Capabilities

```
# Start with brave-search server
mcp-cli interactive --server brave-search

# List search tools
> /tools

# Test search
> /call brave-search search {"query": "MCP protocol"}

# Inspect results
# Understand data structure
```

---

## Best Practices

### 1. Use for Development & Testing

**Do:**
- ✅ Test new MCP servers
- ✅ Verify tool implementations
- ✅ Debug parameter issues
- ✅ Explore tool capabilities

**Don't:**
- ❌ Use for production workflows (use templates)
- ❌ Automate with interactive mode (use query/templates)

### 2. Start with /tools

```
# Always start by listing tools
> /tools

# Then check details
> /tools-all

# Then test
> /call ...
```

### 3. Use Multi-Line for Complex JSON

**For simple arguments:**
```
> /call filesystem read_file {"path": "file.txt"}
```

**For complex arguments:**
```
> /call filesystem
Enter tool name: complex_operation
Enter JSON arguments (end with a line containing only '###'):
{
  "config": {
    "mode": "detailed",
    "options": ["a", "b", "c"]
  },
  "filters": {
    "type": "include",
    "patterns": ["*.go", "*.yaml"]
  }
}
###
```

### 4. Validate Before Production

**Development flow:**

1. **Interactive mode** - Test tool manually
2. **Chat mode** - Test with AI calling tool
3. **Template** - Capture workflow
4. **Automation** - Run in production

**Example:**

```bash
# 1. Test tool in interactive mode
mcp-cli interactive --server filesystem
> /call filesystem list_directory {"path": "."}

# 2. Test with AI in chat mode
mcp-cli chat --server filesystem
You> List files in current directory

# 3. Create template
cat > config/workflows/list-files.yaml << EOF
name: list_files
steps:
  - name: list
    servers: [filesystem]
    prompt: "List files in: {{path}}"
EOF

# 4. Automate
./mcp-cli --template list_files --input-data '{"path": "."}'
```

### 5. Keep Sessions Focused

**Good:**
```
# Test one server at a time
mcp-cli interactive --server filesystem
# Test filesystem tools
# Exit, then test next server
```

**Bad:**
```
# Too many servers
mcp-cli interactive --server filesystem,brave-search,database,github
# Confusing, hard to track which server has which tools
```

---

## Syntax Highlighting

Interactive mode provides syntax highlighting for JSON output:

```
Result:
{
  "success": true,          # Green for strings
  "count": 42,              # Yellow for numbers
  "enabled": true,          # Cyan for booleans
  "items": [                # Arrays with indices
    "item1",
    "item2"
  ]
}
```

**Colors:**
- 🟢 **Green** - Strings
- 🟡 **Yellow** - Numbers
- 🔵 **Blue** - Keys
- 🔵 **Cyan** - Booleans, null

---

## Error Handling

### Invalid Tool Name

```
> /call filesystem nonexistent_tool {}
Error: Tool 'nonexistent_tool' not found on server
```

### Missing Required Parameters

```
> /call filesystem read_file {}
Argument validation failed:
  - Missing required parameter: path
```

### Wrong Parameter Type

```
> /call filesystem list_directory {"path": 123}
Argument validation failed:
  - Parameter 'path' must be a string
```

### Server Not Found

```
> /call nonexistent_server tool {}
Error: Server 'nonexistent_server' not found
```

### Tool Execution Failure

```
> /call filesystem read_file {"path": "/nonexistent/file.txt"}
Tool execution failed: file not found: /nonexistent/file.txt
```

---

## Comparison with Other Modes

| Feature | Interactive | Chat | Query |
|---------|------------|------|-------|
| **AI involved** | No | Yes | Yes |
| **Tool calls** | Manual | Automatic | Automatic |
| **Use case** | Testing | Exploration | Automation |
| **History** | No | Yes | No |
| **Scripting** | No | Limited | Yes |

**When to use each:**

- **Interactive**: Testing tools, debugging servers
- **Chat**: Multi-turn conversations with AI
- **Query**: One-shot automation, scripting

---

## Quick Reference

```bash
# Start interactive mode
mcp-cli interactive --server filesystem

# Commands
/help          # Show help
/tools         # List tools
/tools-all     # Detailed info
/tools-raw     # JSON schema
/call srv tool {"arg": "val"}  # Call tool
/clear         # Clear screen
/exit          # Exit
```

---

## Examples

### Example 1: File Operations

```
> /call filesystem list_directory {"path": "."}
[Lists files]

> /call filesystem read_file {"path": "README.md"}
[Shows README content]

> /call filesystem get_file_info {"path": "config.yaml"}
{
  "size": 1234,
  "modified": "2024-12-26T10:30:00Z",
  "type": "file"
}
```

### Example 2: Web Search

```
mcp-cli interactive --server brave-search

> /tools
Tools from server brave-search:
  - search: Search the web

> /call brave-search search {"query": "MCP protocol documentation"}
{
  "results": [
    {
      "title": "Model Context Protocol",
      "url": "https://...",
      "snippet": "..."
    }
  ]
}
```

### Example 3: Multi-Line Complex Query

```
> /call database
Enter tool name: query
Enter JSON arguments (end with a line containing only '###'):
{
  "sql": "SELECT * FROM users WHERE active = true",
  "parameters": {
    "limit": 10,
    "offset": 0
  },
  "options": {
    "format": "json",
    "include_metadata": true
  }
}
###
[Query results]
```

---

## Next Steps

- **[Chat Mode](chat-mode.md)** - AI-powered conversations
- **[Query Mode](query-mode.md)** - Automation
- **[Debugging](debugging.md)** - Troubleshooting tools

---

**Ready to explore?** Start interactive mode and test your MCP servers! 🔧
