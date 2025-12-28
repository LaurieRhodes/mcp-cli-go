# DeepSeek

> **Skill Level**: 🟢 Beginner-Friendly  
> **Interface**: OpenAI-Compatible  
> **Best For**: Cost-effective coding, budget-conscious development, learning

## Quick Start

```bash
# 1. Get API key
# Sign up at: https://platform.deepseek.com/

# 2. Set API key
export DEEPSEEK_API_KEY=sk-...

# 3. Use it
./mcp-cli query "Write a Python function to sort a list" --provider deepseek
```

---

## Overview

### What is DeepSeek?

DeepSeek is a Chinese AI company offering highly capable coding models at very competitive prices. Their models excel at programming tasks while being significantly cheaper than alternatives.

### When to Use

- ✅ **Use when**:
  - You need **low-cost coding assistance**
  - You're on a **tight budget**
  - You want **good quality** at affordable prices
  - You're **learning/experimenting**
  - You need **solid reasoning** without premium cost

- ❌ **Avoid when**:
  - You need **embeddings** (DeepSeek doesn't offer them)
  - You need **vision** capabilities
  - You require **fastest response times**
  - You need **strict data residency** (servers in China)

### Key Features

- **Chat/Completion**: ✅ Excellent for coding
- **Streaming**: ✅ Fully Supported
- **Tool Calling (MCP)**: ✅ Supported
- **Embeddings**: ❌ Not Available
- **Vision**: ❌ Not Available
- **Cost**: ✅✅✅ Very affordable

---

## Prerequisites

- [ ] **DeepSeek Account** ([Sign up](https://platform.deepseek.com/))
- [ ] **API Key** ([Get key](https://platform.deepseek.com/api_keys))
- [ ] **Payment Method** (after free trial)

---

## Setup Guide

### Step 1: Create Account

1. Go to https://platform.deepseek.com/
2. Sign up with email
3. Verify email

### Step 2: Get API Key

1. Go to [API Keys](https://platform.deepseek.com/api_keys)
2. Click **Create API Key**
3. Copy key (starts with `sk-`)

### Step 3: Set Environment Variable

```bash
export DEEPSEEK_API_KEY=sk-...
```

### Step 4: Test

```bash
./mcp-cli init
# Select "DeepSeek"

./mcp-cli query "Write a quicksort in Python" --provider deepseek
```

---

## Configuration Reference

**File**: `config/providers/deepseek.yaml`

```yaml
interface_type: openai_compatible
provider_name: deepseek
config:
  api_key: ${DEEPSEEK_API_KEY}
  api_endpoint: https://api.deepseek.com/v1
  default_model: deepseek-chat
  timeout_seconds: 300
  max_retries: 2
  context_window: 32000
  reserve_tokens: 2000
```

### Available Models

- `deepseek-chat` - Main chat model (recommended)
- `deepseek-coder` - Optimized for coding

---

## Implementation Details

**Interface**: OpenAI-Compatible  
**Authentication**: Bearer token (API key)  
**Endpoint**: `https://api.deepseek.com/v1/chat/completions`

---

## Features

### Coding Excellence

```bash
# Generate code
./mcp-cli query "Write a REST API in Python using FastAPI" --provider deepseek

# Code review
./mcp-cli query "Review this code for bugs" --provider deepseek

# Debug
./mcp-cli query "Fix this error: IndexError" --provider deepseek
```

**Why DeepSeek**: Excellent code generation at fraction of cost.

### Tool Calling

```bash
# Works with MCP tools
./mcp-cli query "Search for Python async best practices" --provider deepseek
```

---

## Usage Examples

### Example 1: Budget Coding Assistant

```bash
# Use DeepSeek for all coding tasks
./mcp-cli query "Create a binary search tree class in Python with insert, delete, and search methods" --provider deepseek

# Cost: ~$0.0002 per query (vs $0.003+ for GPT-4o-mini)
```

### Example 2: Learning Programming

```bash
./mcp-cli query "Explain list comprehensions in Python with 3 examples" --provider deepseek
```

---

## Cost & Limits

### Pricing (Dec 2024)

**DeepSeek Chat**:
- Input: $0.14 per 1M tokens
- Output: $0.28 per 1M tokens
- **~$0.0002 per typical query**

**10x-20x cheaper than GPT-4o-mini!**

### Comparison

| Model | Input | Output | Typical Query |
|-------|-------|--------|---------------|
| DeepSeek | $0.14/M | $0.28/M | $0.0002 |
| GPT-4o-mini | $0.15/M | $0.60/M | $0.001 |
| Claude Sonnet | $3.00/M | $15.00/M | $0.003 |

---

## Troubleshooting

### Issue: Slower Responses

**Cause**: DeepSeek servers in China

**Solution**: Normal - latency is trade-off for cost

### Issue: No Embeddings

**Cause**: DeepSeek doesn't offer embedding models

**Solution**: Use OpenAI or Ollama for embeddings

---

## Related Resources

- **DeepSeek Platform**: https://platform.deepseek.com/
- **Documentation**: https://platform.deepseek.com/docs
- **Pricing**: https://platform.deepseek.com/pricing

---

**Last Updated**: December 28, 2024  
**Verified Against**: mcp-cli v1.0.0  
**Code Reference**: `internal/providers/ai/clients/openai_compatible.go`
