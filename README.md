# MCP CLI Go

A Go implementation of the Model Context Protocol (MCP) CLI client that provides a multi-capable tool for integrating AI into automation processes.  The significant innovation is this tools ability to process multi-step agentic workflows locally from a single Go binary.

## Features

- **Chat Mode**: Natural language interaction with AI models that can use tools through MCP servers

- **Query Mode**: Single-shot interaction with LLMs for scripting and automation

- **Interactive Mode**: Execute commands directly on MCP servers, bypassing LLM with validation and formatted output****

- **Embedding**: Distinct support for embeddings with supporting LLM Providers

- **YAML Workflow Templating**: execution support for multi-step YAML templates for Agentic operation

- **MCP Client to MCP Server support**: native capability to expose multi-LLM provider workflows as distinct MCP Server tools

- **Template chaining**: supporting "shell out" chaining of workflows from workflows for resilience and token efficiency

- **Local Model Support**: Integration with Ollama and LM Studio for local model deployment

- **Golang**: code base for small, fast, secure use with portability. 

## **Unique Capabilities**

This project demonstrates unique capabilities that no other MCP implementation offers:

| Capability             | Project                   | TypeScript MCP | Python MCP |
| ---------------------- | ------------------------- | -------------- | ---------- |
| Template Composition   | ✅ Full support            | ❌              | ❌          |
| Multi-Step Workflows   | ✅ Full support            | Limited        | Limited    |
| Multi-Provider Support | ✅ 6 provider examples     | 1-2            | 1-2        |
| Context Optimization   | ✅ **Significant savings** | ❌              | ❌          |
| Recursion Control      | ✅ 10 levels deep          | ❌              | ❌          |
| Reusable Primitives    | ✅ Template library        | ❌              | ❌          |
| Variable Isolation     | ✅ Per-template scope      | ❌              | ❌          |

## MCP Template Examples

**Recursive Multi-LLM Workflows**

```yaml
name: research_with_verification
steps:
  - name: initial_research
    template: web_research  # Claude with web search
    output: findings

  - name: fact_check
    template: verification  # GPT-4 verifies Claude's work
    input: "{{findings}}"
    output: verified

  - name: final_synthesis
    template: synthesis     # Back to Claude for writing
    input: "{{verified}}"
```

**Conditional Template Routing**

```yaml
steps:
  - name: classify_request
    prompt: "Is this: technical | sales | support?"
    output: request_type

  - name: route_to_specialist
    condition: "{{request_type}} == 'technical'"
    template: technical_analysis

  - name: route_to_sales
    condition: "{{request_type}} == 'sales'"
    template: sales_workflow
```

**Template Composition Chain Example:**

```
document_intelligence template
  ├─> Calls summarization template (depth 1)
  │     └─> Executes → Returns summary
  ├─> Calls entity_extraction template (depth 1)
  │     └─> Executes → Returns entities
  ├─> Calls sentiment_analysis template (depth 1)
  │     └─> Executes → Returns sentiment
  └─> Synthesizes all results into final intelligence report
```

### Core Documentation

[Index](./docs/index.md)

- [Architecture](./docs/architecture.md) - Overall architecture documentation
- [Template System Guide](./docs/development.md) - Development guide
- [AI Context](./docs/ai_context.md) - AI development context

## 🤝 Contributing

[](https://github.com/LaurieRhodes/mcp-cli-go#-contributing)

This project is shared as example code for your own development and alteration. I'm not certain there would be a lot of interest or value in turning this into a maintained project. If you think I'm wrong - contact me through details at [https://laurierhodes.info](https://laurierhodes.info/)

## 🙏 Acknowledgments

[](https://github.com/LaurieRhodes/mcp-cli-go#-acknowledgments)

This project started in February 2025 as a Golang fork of Chris Hay's ([GitHub - chrishayuk/mcp-cli](https://github.com/chrishayuk/mcp-cli)) as I needed a Go MCP server for use with Go and Function Apps as I experimented with MCP Server development. That project has contined to grow and is well supported by a team of talented coders. I'm grateful for the generous sharing of code under MIT License and encourage everyone to look at and support that project as it really is awesome!

## 📄 License

[](https://github.com/LaurieRhodes/mcp-cli-go#-license)

This project is licensed under the MIT License - see the [LICENSE](https://github.com/LaurieRhodes/mcp-cli-go/blob/main/LICENSE) file for details.
