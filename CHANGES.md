# mcp-cli Changes

## [Unreleased]

### Fixed
- **Standard PATH initialization**: mcp-cli now automatically adds standard system directories to PATH on startup
  - Linux/macOS: `/usr/local/bin`, `/usr/bin`, `/bin`, `/usr/local/sbin`, `/usr/sbin`, `/sbin`
  - Windows: `C:\Windows\system32`, `C:\Windows`, etc.
  - Fixes ENOENT errors when running from non-interactive shells (Claude Desktop, systemd, cron)
  - Ensures `docker`/`podman` are accessible even with minimal PATH
  - **No wrapper script needed** - PATH expansion happens internally at runtime
  - Cross-platform compatible

### Implementation
- Added `env.EnsureStandardPaths()` function in `internal/infrastructure/env/loader.go`
- Called unconditionally in `cmd/root.go` init() function (runs regardless of .env file presence)
- Only adds paths that aren't already present (idempotent)
- Standard paths added to beginning of PATH (take precedence)

### Migration
**Before** (required wrapper):
```json
{
  "command": "/path/to/wrapper.sh",
  "args": ["serve", "config.yaml"]
}
```

**After** (direct binary):
```json
{
  "command": "/path/to/mcp-cli",
  "args": ["serve", "config.yaml"]
}
```

No configuration changes needed - existing configs work without modification.
