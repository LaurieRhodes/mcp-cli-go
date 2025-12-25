# MCP Servers Configuration

Place your MCP server configuration files here.

## Example Server Configuration

Create a file like `filesystem.yaml`:

```yaml
server_name: filesystem
config:
  command: /path/to/filesystem-server
  args: []
  env: {}
```

Each server gets its own YAML file for easy management.
