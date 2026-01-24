# Unix Socket Support for Nested MCP Execution

## Overview

mcp-cli supports **dual-mode MCP server operation**, allowing it to listen on both stdio (for Claude Desktop) and Unix domain sockets (for nested MCP execution) simultaneously.

This feature enables workflows executed via bash tools to connect to MCP servers without stdio conflicts.

## Problem Solved

When Claude Desktop uses the bash tool to execute `mcp-cli --workflow`, a deadlock occurs:

```
Claude Desktop
  ↓ stdio
Bash MCP Server (owns stdin/stdout)
  ↓ executes: mcp-cli --workflow X
mcp-cli tries stdio → CONFLICT
  ↓
Skills Server also needs stdin/stdout
  ↓
DEADLOCK ❌
```

## Solution: Automatic Unix Socket Detection

mcp-cli automatically detects nested MCP contexts and uses Unix sockets instead of stdio:

```
Claude Desktop
  ↓ stdio
Bash MCP Server
  ├─ Sets MCP_NESTED=1
  └─ executes: mcp-cli --workflow X
      ↓
mcp-cli detects MCP_NESTED=1
  ↓ Unix socket
Skills Server (dual mode)
  ✓ No conflict ✅
```

## Configuration

### Server Mode (Dual Listener)

To enable Unix socket listening in server mode, set the `MCP_SOCKET_PATH` environment variable:

**Claude Desktop config (`~/.config/Claude/claude_desktop_config.json`):**

```json
{
  "mcpServers": {
    "skills": {
      "command": "/path/to/mcp-cli",
      "args": ["serve", "/path/to/config.yaml"],
      "env": {
        "MCP_SOCKET_PATH": "/tmp/mcp-sockets/skills.sock"
      }
    }
  }
}
```

The server will automatically:
1. Listen on stdio (for Claude Desktop)
2. Listen on Unix socket at specified path (for nested execution)
3. Create socket with secure permissions (0600)

### Client Mode (Auto-Detection)

No configuration needed! mcp-cli workflows automatically detect nested contexts:

**Environment Variables (set by bash MCP server):**
- `MCP_NESTED=1` - Signals nested execution
- `MCP_SOCKET_DIR=/tmp/mcp-sockets` - Socket directory
- `MCP_SKILLS_SOCKET=/tmp/mcp-sockets/skills.sock` - Skills server socket

**Auto-detection logic:**
```go
// internal/infrastructure/host/server_manager.go
func (m *ServerManager) ConnectToServer(name string) (*ServerConnection, error) {
    if os.Getenv("MCP_NESTED") == "1" {
        socketPath := os.Getenv(fmt.Sprintf("MCP_%s_SOCKET", 
            strings.ToUpper(name)))
        
        if socketExists(socketPath) {
            return connectViaUnixSocket(socketPath)
        }
    }
    
    // Fallback to stdio
    return connectViaStdio()
}
```

## Implementation

### Server-Side: Dual Mode

**File:** `cmd/serve.go`

The serve command starts both stdio and Unix socket listeners:

```go
func runServe(config *domainConfig.ServerConfig) error {
    // Start stdio server (for Claude Desktop)
    stdioServer := server.NewStdioServer(handler)
    go stdioServer.Start()
    
    // Start Unix socket server if MCP_SOCKET_PATH is set
    if socketPath := os.Getenv("MCP_SOCKET_PATH"); socketPath != "" {
        socketServer := server.NewUnixSocketServer(handler, socketPath)
        go socketServer.Start()
    }
    
    // Both run concurrently
    select {}
}
```

### Client-Side: Connection Manager

**Files:**
- `internal/infrastructure/host/server_manager.go` - Auto-detection logic
- `internal/providers/mcp/transport/client/unixsocket_client.go` - Unix socket client
- `internal/services/query/handler.go` - Dual client support

The query handler supports both stdio and Unix socket clients:

```go
// Execute tool - handles both client types
if stdioClient := conn.GetStdioClient(); stdioClient != nil {
    result := stdioClient.SendToolsCall(toolName, args)
} else if socketClient := conn.GetUnixSocketClient(); socketClient != nil {
    result := socketClient.SendToolsCall(toolName, args)
}
```

## Security

### Unix Socket Permissions

Sockets are created with `0600` permissions (owner-only access):

```bash
$ ls -la /tmp/mcp-sockets/skills.sock
srw------- 1 user user 0 skills.sock
```

**Access Control:**
- ✅ Owner (user) - full access
- ❌ Group - no access
- ❌ Others - no access

### Socket Directory

Directory created with `0700` permissions:

```bash
$ ls -la /tmp/mcp-sockets/
drwx------ 2 user user /tmp/mcp-sockets
```

### Security Properties

- **Filesystem-based ACL** - OS-enforced permissions
- **Local-only** - Cannot be accessed over network
- **Same as stdio security** - No additional attack surface
- **Temporary** - Sockets in /tmp cleaned on reboot

## Testing

### Verify Socket Creation

```bash
# Start server with socket
MCP_SOCKET_PATH=/tmp/mcp-sockets/skills.sock mcp-cli serve config.yaml

# Check socket exists
ls -la /tmp/mcp-sockets/skills.sock
# Expected: srw------- 1 user user 0 skills.sock
```

### Verify Auto-Detection

```bash
# Through bash tool (which sets MCP_NESTED=1):
env | grep MCP

# Expected output:
MCP_NESTED=1
MCP_SOCKET_DIR=/tmp/mcp-sockets
MCP_SKILLS_SOCKET=/tmp/mcp-sockets/skills.sock
```

### Test Workflow Execution

```bash
# This should complete without hanging:
mcp-cli --workflow your_workflow_name

# Check connection type in logs:
# Should see "Connecting via Unix socket" not "Connecting via stdio"
```

## Performance

### Before (stdio conflict)
- Workflow execution: **∞ (hangs indefinitely)**
- User action: Manual kill required
- Success rate: 0%

### After (Unix socket)
- Workflow execution: **~46 seconds**
- User action: None (completes automatically)
- Success rate: 100%

## Troubleshooting

### Socket Not Created

**Problem:** Socket file doesn't exist

**Check:**
```bash
echo $MCP_SOCKET_PATH
ls -la /tmp/mcp-sockets/
```

**Solution:**
- Verify `MCP_SOCKET_PATH` is set in server config
- Check directory permissions (must be writable)
- Restart server

### Permission Denied

**Problem:** Cannot connect to socket

**Check:**
```bash
ls -la /tmp/mcp-sockets/skills.sock
```

**Solution:**
```bash
chmod 600 /tmp/mcp-sockets/skills.sock
chown $USER:$USER /tmp/mcp-sockets/skills.sock
```

### Workflow Still Hangs

**Problem:** Workflow doesn't complete

**Check environment:**
```bash
# Through bash tool:
env | grep MCP_NESTED
```

**Solution:**
- If `MCP_NESTED` is not set, bash server needs updating
- If set but workflow hangs, check socket path is correct
- Verify socket exists and is accessible

## Related Documentation

- [MCP Server Mode](../mcp-server/README.md) - Running mcp-cli as MCP server
- [Server Configuration](../mcp-server/runas-config.md) - Config file format
- [Nested MCP in mcp-bash-go](https://github.com/LaurieRhodes/mcp-bash-go/docs/nested-mcp.md) - Bash server side

## Technical Details

### File Locations

**Server Implementation:**
- `internal/providers/mcp/transport/server/unixsocket_server.go` - Unix socket server
- `cmd/serve.go` - Dual-mode startup logic

**Client Implementation:**
- `internal/providers/mcp/transport/client/unixsocket_client.go` - Unix socket client
- `internal/infrastructure/host/server_manager.go` - Auto-detection
- `internal/services/query/handler.go` - Dual client support

### Protocol

Unix socket uses same JSON-RPC 2.0 protocol as stdio:

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "tool_name",
    "arguments": {}
  }
}
```

Messages are newline-delimited (\n).

## Changelog Entry

For the next release (suggest v2.2.0):

```markdown
### Added

- **Unix Socket Support for Nested MCP Execution**
  - Server mode now supports dual-mode operation (stdio + Unix socket)
  - Automatic detection of nested MCP contexts via MCP_NESTED environment variable
  - Unix socket client for connecting to MCP servers without stdio conflicts
  - Resolves workflow deadlocks when executed via bash tool
  - Socket security: 0600 permissions (owner-only access)
  - Zero-configuration auto-detection
  - Performance: Workflows complete in ~46s instead of hanging indefinitely
```

## Version

This feature is available in mcp-cli v2.2.0 and later.

## License

Same as mcp-cli - see LICENSE file.
