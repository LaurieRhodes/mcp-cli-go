# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [2.2.0] - 2026-01-24

### Added

- **Unix Socket Support for Nested MCP Execution**
  
  - **Dual-mode MCP server operation**: Servers can now listen on both stdio (for Claude Desktop) and Unix sockets (for nested workflows) simultaneously
  - **Automatic nested context detection**: mcp-cli workflows automatically detect when running in nested MCP contexts via `MCP_NESTED` environment variable
  - **Unix socket client**: New transport layer for connecting to MCP servers via Unix domain sockets
  - **Zero-configuration auto-detection**: No manual configuration needed - works automatically when environment variables are set
  - **Resolves workflow deadlocks**: Workflows executed via bash tool no longer hang due to stdio conflicts
  - **Secure by default**: Unix sockets created with 0600 permissions (owner-only access)
  
  **Problem solved:** When Claude Desktop used the bash tool to execute `mcp-cli --workflow`, both the bash server and skills server tried to use stdin/stdout simultaneously, causing a deadlock. Workflows would hang indefinitely.
  
  **Solution:** Server mode now supports dual listeners (stdio + Unix socket). When workflows detect nested execution via `MCP_NESTED=1`, they automatically connect via Unix socket instead of stdio, eliminating the conflict.
  
  **Performance improvement:**
  
  - Before: Workflows hung indefinitely (∞)
  - After: Workflows complete successfully (~46 seconds)
  - Success rate: 0% → 100%
  
  **Configuration:**
  
  Enable Unix socket in Claude Desktop config:
  
  ```json
  {
    "mcpServers": {
      "skills": {
        "command": "/path/to/mcp-cli",
        "args": ["serve", "config.yaml"],
        "env": {
          "MCP_SOCKET_PATH": "/tmp/mcp-sockets/skills.sock"
        }
      }
    }
  }
  ```

## [2.2.0] - 2026-01-24

### Added - Tasks SEP-1686 Implementation

mcp-cli is now one of the **first MCP servers with full Tasks SEP-1686 support**, enabling call-now, fetch-later execution for long-running workflows.

- **Complete Tasks SEP-1686 Implementation**
  
  - **Task-augmented tool calls**: Tools can be called with task metadata for non-blocking execution
  - **tasks/get**: Poll task status at any time
  - **tasks/result**: Retrieve results when task completes (blocks until ready)
  - **tasks/list**: List all active tasks with cursor-based pagination
  - **tasks/cancel**: Cancel running tasks
  - **Automatic TTL management**: Tasks expire after configurable duration (default: 30min, max: 2hr)
  - **Background execution**: Tools execute in goroutines without blocking protocol
  - **Thread-safe task storage**: Concurrent-safe task access with mutex protection
  - **Automatic cleanup**: Expired tasks removed every minute
  - **Status tracking**: 5 task states (working, completed, failed, cancelled, input_required)
  - **Cryptographic task IDs**: 128-bit random UUIDs for security

- **Capability Negotiation**
  
  - Server declares task support during initialization
  - Granular control: tasks.requests.tools/call, tasks.list, tasks.cancel
  - Follows SEP-1686 capability structure exactly

- **Documentation**
  
  - Complete implementation guide: `docs/tasks-sep-1686.md`
  - Usage examples for API/MCP clients
  - Architecture documentation with diagrams
  - Troubleshooting guide
  - Standards compliance checklist

**Problem Solved:** Long-running workflows (30+ minutes) previously failed due to client timeouts. With Tasks SEP, clients receive task ID immediately and poll for completion.

**Real-World Performance:**
- RLM extraction workflow: 0% success → 100% success
- Execution time: ∞ (timeout) → 13 minutes (complete)
- User experience: blocking → non-blocking

**Technical Details:**
- Task manager: `internal/infrastructure/tasks/manager.go`
- Domain types: `internal/domain/task.go`  
- Server handlers: `internal/services/server/service.go`
- Transport integration: stdio + Unix socket servers
- Default TTL: 30 minutes, Max TTL: 2 hours, Poll interval: 5 seconds

**Standards Compliance:** Fully compliant with Tasks SEP-1686 (accepted 2025-10-20)

See [Tasks Documentation](docs/tasks-sep-1686.md) for complete details.

## [2.1.0] - 2026-01-20

### Added

- **Workflow System**: 
  
  - Iterative looping
  - Parrallelism with Loops
  - Parallelism (eg fan-out_ with workflows)
  - Skills filtering on tasks
  - Bash skill executor support

- **RAG**
  
  - `RAG query support with new RAG config

## [2.0.0] - 2026-01-08

### Added

- **Skills System**: Cross-LLM document creation capability
  
  - PowerPoint, Excel, Word, and PDF creation via any LLM (GPT-4, DeepSeek, Gemini, Claude)
  - Container-based secure execution (Docker/Podman)
  - Skills from Anthropic's repository exposed as MCP tools
  - Helper libraries (python-pptx, openpyxl, pypdf, etc.) pre-installed in containers
  - Automatic skill discovery and tool generation
  - Configurable outputs directory for file persistence
  - Container image mapping system (skill-images.yaml)

- **YAML Configuration for Skills**
  
  - `settings.yaml` now includes `skills.outputs_dir` configuration
  - `skill-images.yaml` for container image mapping
  - Skills configuration in init command

- **New Providers**
  
  - Kimi K2 (Moonshot AI) - 128K context window support
  - AWS Bedrock embeddings support
  - Azure Foundry embeddings support
  - GCP Vertex AI provider

- **HTTP Proxy System**
  
  - Expose MCP servers as REST APIs
  - Authentication via API keys
  - CORS support
  - Optional HTTPS/TLS configuration
  - Proxy configurations for bash, filesystem, skills, and workflow templates

- **Documentation**
  
  - Complete skills documentation (13 files in docs/skills/)
  - Container setup guides
  - Skills creation guide
  - Updated README with skills section
  - Comprehensive examples directory with all config types

- **Init Command Enhancements**
  
  - Auto-creates skill-images.yaml
  - Auto-creates skills directory structure
  - Generates complete proxy configurations
  - Creates all runasMCP examples

### Changed

- **Repository Structure**
  
  - Removed root-level Dockerfile and docker-compose.yml (contradicted single-binary design)
  - Skills container images now built via `docker/skills/build-skills-images.sh`
  - .gitignore updated to properly ignore /config/ directory
  - Examples directory fully synchronized with production configurations

- **Configuration Architecture**
  
  - Skills configuration moved from hardcoded Go to YAML (settings.yaml)
  - No code recompilation needed for configuration changes
  - Consistent with existing configuration system

- **Gemini Provider**
  
  - Updated to use `gemini_native` interface type
  - Improved reliability and performance

- **Documentation**
  
  - Skills documentation reduced from ~4,200 lines to 1,305 lines (69% reduction)
  - Removed speculative content, kept only verified information
  - All statements validated against actual code and defaults
  
  **Templates**
  
  - New Semantic workflows replace previous templates

### Fixed

- **DeepSeek Tool Calling** (3 Critical Bugs)
  
  - Fixed tool error response handling (missing `error: "..."` in tool results)
  - Fixed duplicate tool IDs causing API rejection
  - Fixed empty/null tool arguments causing parsing errors
  - DeepSeek now fully functional with MCP tools

- **Container Execution**
  
  - Verified Docker/Podman detection and execution
  - Container security settings validated
  - Image mapping system tested across all skills

- **Examples Directory**
  
  - Removed hardcoded credentials from examples
  - All API keys now use ${ENV_VAR} format
  - Generic paths instead of user-specific absolute paths

### Removed

- Root-level Dockerfile (containerizing the binary contradicted design)
- Root-level docker-compose.yml (DooD deployment moved to optional/advanced)
- ~3,000 lines of unverified/speculative documentation
- Redundant session notes from docker/skills/ directory (89% size reduction)

### Security

- Container-based execution with strict security:
  
  - No network access (--network=none)
  - Read-only root filesystem
  - Memory limits (256MB default)
  - CPU limits (0.5 cores default)
  - Process limits (100 max)
  - Automatic cleanup after execution

- Configuration separation:
  
  - Production configs in /config/ (gitignored)
  - Example configs in /examples/config/ (tracked, sanitized)
  - No sensitive data in repository

## [2.0.0-rc.1]

### Added

- YAML Workflow template system (V2)
- Embeddings support
- Template composition support
- GitHub Actions workflow for cross-platform builds
- MCP Client to MCP Server support: native capability to expose multi-LLM provider workflows as distinct MCP Server tools
- Template chaining: supporting "shell out" chaining of workflows from workflows for resilience and token efficiency
- Local Model Support: Integration with LM Studio for local model deployment
- Glamour (charmbracelet) support for chat mode

### Changed

- Improved help menu formatting for better terminal compatibility
- Enhanced documentation structure

### Fixed

- TTY output support for query mode on Linux to resolve intermittent failures with STDOUT

## [1.0.0] - 2025-06-15

### Added

- Initial release
- Chat mode for interactive AI conversations
- Query mode for single-shot interactions
- Interactive mode with MCP servers
- Multi-provider AI support (OpenAI, Anthropic, Ollama, DeepSeek, Gemini, OpenRouter)
- MCP server mode
- Configuration management

---

## Release Types

### Major (x.0.0)

- Breaking changes
- Major feature additions
- Architecture changes

### Minor (0.x.0)

- New features
- Non-breaking enhancements
- New provider support

### Patch (0.0.x)

- Bug fixes
- Documentation updates
- Performance improvements

[2.0.0]: https://github.com/LaurieRhodes/mcp-cli-go/releases/tag/v2.0.0
[2.0.0-rc.1]: https://github.com/LaurieRhodes/mcp-cli-go/releases/tag/v2.0.0-rc.1
[1.0.0]: https://github.com/LaurieRhodes/mcp-cli-go/releases/tag/v1.0.0
