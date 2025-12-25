# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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

- TTY output support for query mode on Linux to resolve intermittant failures with STDOUT

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

[Unreleased]: https://github.com/LaurieRhodes/mcp-cli-go/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/LaurieRhodes/mcp-cli-go/releases/tag/v0.1.0
