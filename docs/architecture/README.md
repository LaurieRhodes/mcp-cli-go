# Architecture Documentation

Comprehensive technical architecture for MCP-CLI-Go, designed for architects, senior engineers, and system integrators who need deep understanding of system internals, design decisions, and operational characteristics.

---

## Documentation Organization

This architecture documentation is organized for **progressive disclosure** - start broad, drill down to specifics:

### [Overview](overview.md) - System Architecture

**Start here.** High-level architecture, design philosophy, and key architectural decisions.

**Critical sections:**

- System architecture with C4 model context
- Design philosophy and architectural constraints
- Layered architecture and dependency rules
- Operational modes and their architectural implications
- Concurrency model and process coordination
- Configuration hierarchy and resolution strategy

**Read this to:** Understand the system's structure, design principles, and how components interact.

---

### [Components](components.md) - Deep Component Analysis

Detailed examination of each architectural component, implementation patterns, and internal structure.

**Critical sections:**

- Command layer: CLI framework and routing
- Service layer: Orchestration and workflow management
- Core layer: Business logic implementations
- Provider layer: External system adapters
- Infrastructure layer: Cross-cutting concerns
- Domain layer: Core types and contracts

**Read this to:** Understand component responsibilities, implementation details, and extension points.

---

### [Data Flow](data-flow.md) - Request/Response Lifecycles

How data moves through the system in different operational modes and scenarios.

**Critical sections:**

- Complete request flows for each mode
- Tool execution and MCP protocol flows
- Error propagation and handling
- Streaming data processing
- State transitions and side effects

**Read this to:** Debug issues, optimize performance, or understand system behavior under different conditions.

---

###[API & Domain](api-domain.md) - Contracts and Interfaces
Core types, interfaces, schemas, and API contracts.

**Critical sections:**

- Domain model and type definitions
- Interface specifications
- Provider abstractions and implementations
- Configuration schema and validation
- Extension points and plugin architecture

**Read this to:** Extend the system, integrate external components, or understand the contract surface.

---

## Architectural Principles

### 1. Modularity with Bounded Contexts

**Principle:** System divided into cohesive modules with clear boundaries and minimal coupling.

**Implementation:**

- **Command Context** - CLI interaction and routing
- **Execution Context** - Mode-specific business logic
- **Provider Context** - External system integration
- **Infrastructure Context** - Cross-cutting services

**Constraint:** Modules communicate through well-defined interfaces only. No cross-module implementation dependencies.

**Trade-off:** More initial design overhead for long-term maintainability and parallel development.

---

### 2. Interface-Based Design with Dependency Inversion

**Principle:** Depend on abstractions (interfaces), not concrete implementations.

**Implementation:**

```go
// High-level module depends on abstraction
type ChatService struct {
    provider domain.LLMProvider  // Interface, not *OpenAIClient
}

// Low-level modules implement abstractions
type OpenAIClient struct { ... }
func (c *OpenAIClient) CreateCompletion(...) (*CompletionResponse, error)
```

**Benefit:** Easy mocking, runtime polymorphism, provider swapping without code changes.

**Constraint:** All external dependencies must have interface wrappers.

---

### 3. Configuration-Driven Behavior

**Principle:** Behavior controlled through configuration, not code compilation.

**Implementation:**

- Provider selection via config
- Model parameters via config
- Template workflows via YAML
- Server connections via config

**Benefits:**

- Zero-downtime provider switching
- A/B testing different models
- Environment-specific configurations
- User customization without code access

**Constraint:** Configuration changes must be validated at load time to fail fast.

---

### 4. Fail-Fast with Graceful Degradation

**Principle:** Validate early, fail fast on unrecoverable errors. Degrade gracefully on transient failures.

**Implementation:**

- Configuration validation at startup (fail-fast)
- API call retries with exponential backoff (graceful degradation)
- Circuit breakers for external dependencies (protection)
- Fallback to Ollama when remote providers unavailable (degradation)

**Trade-off:** More complex error handling logic for better user experience.

---

### 5. Explicit Over Implicit

**Principle:** Make operations visible and traceable. No hidden magic.

**Implementation:**

```bash
# Explicit provider selection
mcp-cli query --provider anthropic "question"

# Explicit template execution
mcp-cli --template code_review --input-data '{"code": "..."}'

# Explicit server connection
mcp-cli chat --server filesystem,brave_search
```

**Benefit:** Users understand what's happening. Easier debugging and troubleshooting.

**Constraint:** More verbose commands, but clarity trumps brevity.

---

### 6. Concurrent by Design

**Principle:** Leverage Go's concurrency primitives for parallelism and responsiveness.

**Implementation:**

- Each MCP server runs in separate goroutine
- Tool calls execute concurrently when possible
- Streaming responses processed asynchronously
- User input handled in separate goroutine

**Constraint:** Must handle synchronization correctly. All shared state protected by mutexes.

---

- **[Overview](overview.md)** - System architecture and design
- **[Components](components.md)** - Component deep-dive
- **[Data Flow](data-flow.md)** - Request/response flows
- **[API & Domain](api-domain.md)** - Interfaces and contracts

---
