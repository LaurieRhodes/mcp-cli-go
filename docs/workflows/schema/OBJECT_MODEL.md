# Workflow Object Model

**Version:** workflow/v2.0  
**Purpose:** TypeScript-style interface definitions showing inheritance model

---

## Core Concept

> Workflows are sequences of mcp-cli calls. All objects inherit from a base `MCPQuery` interface that maps directly to CLI arguments.

---

## Base Interface

### MCPQuery

The foundation - all execution contexts inherit these properties.

```typescript
interface MCPQuery {
  // Provider configuration
  provider: string;              // AI provider (anthropic, openai, deepseek, ollama, gemini)
  model: string;                 // Model identifier (claude-sonnet-4, gpt-4o, etc.)

  // Model parameters
  temperature?: number;          // 0.0 (deterministic) to 2.0 (creative), default: 0.7
  max_tokens?: number;           // Maximum tokens in response, default: auto

  // Infrastructure
  servers?: string[];            // MCP servers to enable, default: []
  timeout?: Duration;            // Call timeout, default: 60s
}
```

**CLI Mapping:**

```bash
mcp-cli \
  --provider ${provider} \
  --model ${model} \
  --temperature ${temperature} \
  --max-tokens ${max_tokens} \
  --server ${servers[0]} \
  --server ${servers[1]}
```

---

## Workflow Structure

### WorkflowV2

Root workflow definition.

```typescript
interface WorkflowV2 {
  // Metadata (required)
  $schema: "workflow/v2.0";     // Schema version
  name: string;                  // Unique identifier
  version: string;               // Semantic version (e.g., "1.0.0")
  description: string;           // Human-readable description

  // Execution defaults (required)
  execution: ExecutionContext;   // Workflow-level MCPQuery defaults

  // Environment (optional)
  env?: Record<string, string>;  // Environment variables

  // Execution phases (at least one required)
  steps: Step[];                 // Sequential steps
  loops?: Loop[];                // Iterative loops
}
```

---

### ExecutionContext

Workflow-level execution defaults. Same properties as `MCPQuery` with optional provider failover.

```typescript
interface ExecutionContext extends MCPQuery {
  // Option 1: Single provider (extends MCPQuery)
  provider: string;
  model: string;
  temperature?: number;
  max_tokens?: number;
  servers?: string[];
  skills?: string[];
  timeout?: Duration;

  // Option 2: Failover chain (mutually exclusive with single provider)
  providers?: ProviderFallback[];

  // Additional workflow settings
  logging?: "normal" | "verbose" | "noisy";
  no_color?: boolean;
}

interface ProviderFallback {
  provider: string;
  model: string;
  temperature?: number;
  max_tokens?: number;
  timeout?: Duration;
}
```

**Usage:**

```typescript
// Single provider
const execution: ExecutionContext = {
  provider: "anthropic",
  model: "claude-sonnet-4",
  temperature: 0.7,
  servers: ["filesystem"],
  skills: ["docx", "pdf"]
};

// Failover chain
const execution: ExecutionContext = {
  providers: [
    { provider: "anthropic", model: "claude-sonnet-4" },
    { provider: "openai", model: "gpt-4o" },
    { provider: "ollama", model: "qwen2.5:32b" }
  ],
  temperature: 0.7,
  servers: ["filesystem"],
  skills: ["docx", "xlsx"]
};
```

---

## Step Definitions

### Step

Steps inherit all `MCPQuery` properties and can override them.

```typescript
interface Step extends Partial<MCPQuery> {
  // Identity (required)
  name: string;                  // Unique step identifier

  // Orchestration (optional)
  needs?: string[];              // Step dependencies (wait for these)
  if?: string;                   // Skip if condition false

  // Inherited properties (can override)
  provider?: string;             // Override provider
  model?: string;                // Override model
  temperature?: number;          // Override temperature
  max_tokens?: number;           // Override max_tokens
  servers?: string[];            // Override servers
  timeout?: Duration;            // Override timeout

  // Execution mode (exactly ONE required)
  run?: string;                  // LLM query mode
  template?: TemplateCall;       // Workflow call mode
  embeddings?: EmbeddingsConfig; // Embeddings generation mode
  consensus?: ConsensusConfig;   // Multi-provider validation mode
}
```

**Inheritance hierarchy:**

```
workflow.execution (MCPQuery defaults)
         ↓ inherits
    Step (can override any MCPQuery property)
```

---

## Execution Modes

### Mode 1: LLM Query

```typescript
interface Step {
  name: string;
  run: string;                   // Prompt with {{variable}} interpolation

  // Optional overrides
  temperature?: number;
  // ... any MCPQuery property
}
```

**Example:**

```typescript
{
  name: "analyze",
  run: "Analyze this code: {{input}}",
  temperature: 0.3
}
```

---

### Mode 2: Template Call

Call another workflow.

```typescript
interface TemplateCall {
  name: string;                  // Workflow name to call
  with?: Record<string, any>;    // Input data (key-value pairs)
}

interface Step {
  name: string;
  template: TemplateCall;

  // MCPQuery properties passed through to called workflow
}
```

**Example:**

```typescript
{
  name: "review",
  template: {
    name: "code_reviewer",
    with: {
      code: "{{input}}",
      language: "go",
      strict: true
    }
  }
}
```

---

### Mode 3: Embeddings Generation

Generate vector embeddings from text.

```typescript
interface EmbeddingsConfig {
  // Input source (one required)
  input?: string | string[];     // Inline text(s)
  input_file?: string;           // File path (alternative to input)

  // Provider override (optional - inherits from step/execution)
  provider?: string;
  model?: string;

  // Chunking configuration (optional)
  chunk_strategy?: "sentence" | "paragraph" | "fixed";
  max_chunk_size?: number;       // Default: 512 tokens
  overlap?: number;              // Default: 0 tokens

  // Model configuration (optional)
  dimensions?: number;           // For supported models

  // Output configuration (optional)
  encoding_format?: "float" | "base64";
  include_metadata?: boolean;    // Default: true
  output_format?: "json" | "csv" | "compact";
  output_file?: string;          // File path for output
}

interface Step {
  name: string;
  embeddings: EmbeddingsConfig;

  // Can override provider/model at step level too
  provider?: string;
  model?: string;
}
```

**Example:**

```typescript
{
  name: "embed_docs",
  embeddings: {
    model: "text-embedding-3-small",
    input: ["doc1", "doc2", "doc3"],
    chunk_strategy: "sentence",
    max_chunk_size: 512,
    overlap: 50,
    output_file: "embeddings.json"
  }
}
```

---

### Mode 4: Consensus Validation

Execute across multiple providers and require agreement.

```typescript
interface ConsensusConfig {
  // Prompt (required)
  prompt: string;                // Sent to all providers (supports {{variables}})

  // Executions (required)
  executions: ConsensusExecution[];

  // Agreement threshold (required)
  require: "unanimous" | "majority" | "2/3";

  // Timeout (optional)
  timeout?: Duration;            // Default: 60s
}

interface ConsensusExecution extends Partial<MCPQuery> {
  // Identity (required)
  provider: string;
  model: string;

  // Overrides (optional)
  temperature?: number;
  max_tokens?: number;
  timeout?: Duration;
}

interface Step {
  name: string;
  consensus: ConsensusConfig;

  // Step-level properties inherited by all executions
  temperature?: number;
}
```

**Inheritance chain for consensus:**

```
workflow.execution (defaults)
         ↓
    Step (can override)
         ↓
consensus.executions[] (can override per-execution)
```

**Example:**

```typescript
{
  name: "validate",
  temperature: 0.2,              // Inherited by all executions
  consensus: {
    prompt: "Is this safe? Answer YES or NO: {{code}}",
    executions: [
      {
        provider: "anthropic",
        model: "claude-sonnet-4"
        // Inherits temperature: 0.2
      },
      {
        provider: "openai",
        model: "gpt-4o",
        temperature: 0.1         // Override for this execution
      },
      {
        provider: "deepseek",
        model: "deepseek-chat"
        // Inherits temperature: 0.2
      }
    ],
    require: "2/3"
  }
}
```

---

## Loop Definitions

### Loop

Iteratively execute a workflow until a condition is met.

```typescript
interface Loop {
  // Identity (required)
  name: string;                  // Loop identifier

  // Workflow reference (required)
  workflow: string;              // Workflow name to call repeatedly
  with: Record<string, any>;     // Input data (can use {{loop.*}} variables)

  // Exit conditions (required)
  max_iterations: number;        // Safety limit
  until: string;                 // LLM-evaluated condition (e.g., "Tests pass")

  // Error handling (optional)
  on_failure?: "continue" | "fail";

  // History tracking (optional)
  accumulate?: string;           // Variable name to store all iteration results
}
```

**Special loop variables:**

- `{{loop.iteration}}` - Current iteration number (1-based)
- `{{loop.last.output}}` - Previous iteration result
- `{{loop.history}}` - All iteration results (if accumulate set)

**Example:**

```typescript
{
  name: "fix_code",
  workflow: "test_and_fix",
  with: {
    code: "{{input}}",
    previous_error: "{{loop.last.output.error}}"
  },
  max_iterations: 10,
  until: "All tests pass",
  on_failure: "continue",
  accumulate: "fix_history"
}
```

---

## Type Hierarchy

Visual representation of the inheritance model:

```typescript
// Base interface
interface MCPQuery {
  provider: string;
  model: string;
  temperature?: number;
  max_tokens?: number;
  servers?: string[];
  timeout?: Duration;
}

// Workflow level
interface ExecutionContext extends MCPQuery {
  providers?: ProviderFallback[];
  logging?: string;
  no_color?: boolean;
}

// Step level (can override any MCPQuery property)
interface Step extends Partial<MCPQuery> {
  name: string;
  needs?: string[];
  if?: string;

  // Execution modes
  run?: string;
  template?: TemplateCall;
  embeddings?: EmbeddingsConfig;
  consensus?: ConsensusConfig;
}

// Consensus level (can override any MCPQuery property)
interface ConsensusExecution extends Partial<MCPQuery> {
  provider: string;
  model: string;
}

// Embeddings (can override provider/model)
interface EmbeddingsConfig {
  provider?: string;
  model?: string;
  input?: string | string[];
  // ... chunking and output config
}
```

---

## Property Inheritance Flow

```typescript
// Example showing inheritance in action

const workflow: WorkflowV2 = {
  $schema: "workflow/v2.0",
  name: "example",
  version: "1.0.0",
  description: "Inheritance example",

  execution: {
    provider: "anthropic",       // Default for all steps
    model: "claude-sonnet-4",    // Default for all steps
    temperature: 0.7,            // Default for all steps
    servers: ["filesystem"],     // Default for all steps
    skills: ["docx", "xlsx"]     // Default for all steps
  },

  steps: [
    {
      name: "step1",
      run: "Analyze"
      // Inherits: provider, model, temperature, servers, skills
    },

    {
      name: "step2",
      temperature: 0.3,          // Override temperature only
      run: "Refine"
      // Inherits: provider, model, servers, skills
      // Overrides: temperature
    },

    {
      name: "step3",
      consensus: {
        prompt: "Validate",
        executions: [
          {
            provider: "anthropic",
            model: "claude-sonnet-4"
            // Inherits temperature from step3 (0.7 from execution)
          },
          {
            provider: "openai",
            model: "gpt-4o",
            temperature: 0.1     // Override for this execution
          }
        ],
        require: "unanimous"
      }
    }
  ]
};
```

**Effective configuration for each execution:**

```typescript
// step1 effective config
{
  provider: "anthropic",         // from execution
  model: "claude-sonnet-4",      // from execution
  temperature: 0.7,              // from execution
  servers: ["filesystem"]        // from execution
}

// step2 effective config
{
  provider: "anthropic",         // from execution
  model: "claude-sonnet-4",      // from execution
  temperature: 0.3,              // overridden at step
  servers: ["filesystem"]        // from execution
}

// step3 consensus execution 1 effective config
{
  provider: "anthropic",         // from consensus execution
  model: "claude-sonnet-4",      // from consensus execution
  temperature: 0.7,              // from execution (inherited through step3)
  servers: ["filesystem"]        // from execution (inherited through step3)
}

// step3 consensus execution 2 effective config
{
  provider: "openai",            // from consensus execution
  model: "gpt-4o",               // from consensus execution
  temperature: 0.1,              // overridden at consensus execution
  servers: ["filesystem"]        // from execution (inherited through step3)
}
```

---

## Validation Rules

### Required Fields

```typescript
// Workflow level
✓ $schema must be "workflow/v2.0"
✓ name must be unique string
✓ version must be semantic version
✓ description must be non-empty
✓ execution must have provider + model OR providers array
✓ steps must have at least one element

// Step level
✓ name must be unique within workflow
✓ Exactly ONE execution mode (run, template, embeddings, consensus)
✓ needs array can only reference previous steps

// Consensus level
✓ prompt required
✓ executions array must have at least 2 elements
✓ require must be valid threshold
✓ Each execution must have provider + model

// Loop level
✓ name must be unique within workflow
✓ workflow must reference existing workflow
✓ max_iterations must be positive integer
✓ until must be non-empty string
```

### Type Constraints

```typescript
type Provider = "anthropic" | "openai" | "deepseek" | "ollama" | "gemini" | "openrouter";
type Temperature = number; // 0.0 to 2.0
type ChunkStrategy = "sentence" | "paragraph" | "fixed";
type EncodingFormat = "float" | "base64";
type OutputFormat = "json" | "csv" | "compact";
type ConsensusThreshold = "unanimous" | "majority" | "2/3";
type LogLevel = "normal" | "verbose" | "noisy";
type Duration = string; // e.g., "30s", "5m", "1h"
```

---

## Complete Example

```typescript
const workflow: WorkflowV2 = {
  $schema: "workflow/v2.0",
  name: "ai_code_review",
  version: "1.0.0",
  description: "Multi-stage code review with consensus validation",

  execution: {
    provider: "anthropic",
    model: "claude-sonnet-4",
    temperature: 0.7,
    servers: ["filesystem"]
  },

  env: {
    REVIEW_STYLE: "strict"
  },

  steps: [
    {
      name: "initial_review",
      run: "Review this code for bugs and style issues: {{input}}"
    },

    {
      name: "security_scan",
      needs: ["initial_review"],
      temperature: 0.2,
      run: "Scan for security vulnerabilities: {{input}}"
    },

    {
      name: "consensus_validation",
      needs: ["initial_review", "security_scan"],
      consensus: {
        prompt: "Is this code safe to deploy? Answer YES or NO based on: {{initial_review}} {{security_scan}}",
        executions: [
          { provider: "anthropic", model: "claude-sonnet-4" },
          { provider: "openai", model: "gpt-4o" },
          { provider: "deepseek", model: "deepseek-chat" }
        ],
        require: "unanimous"
      }
    },

    {
      name: "generate_report",
      needs: ["consensus_validation"],
      template: {
        name: "report_generator",
        with: {
          review: "{{initial_review}}",
          security: "{{security_scan}}",
          validation: "{{consensus_validation}}"
        }
      }
    }
  ]
};
```

---

## See Also

- **[Quick Reference](QUICK_REFERENCE.md)** - One-page overview
- **[CLI Mapping](CLI_MAPPING.md)** - Property → CLI argument mapping
- **[Inheritance Guide](INHERITANCE_GUIDE.md)** - Visual inheritance diagrams
- **[Steps Reference](STEPS_REFERENCE.md)** - Detailed step modes
- **[Consensus Reference](CONSENSUS_REFERENCE.md)** - Consensus validation
- **[Loops Reference](LOOPS_REFERENCE.md)** - Iterative execution

---

**Key Takeaway:** The entire workflow system is built on a single `MCPQuery` interface that maps to mcp-cli arguments. Everything else is composition, inheritance, and orchestration of these base properties.
