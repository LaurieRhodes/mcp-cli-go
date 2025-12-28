# Gemini (Public API)

> **Skill Level**: 🟢 Beginner-Friendly  
> **Interface**: OpenAI-Compatible  
> **Best For**: Massive context windows (2M tokens), multimodal, Google integration

## Quick Start

```bash
# 1. Get API key
# Go to: https://makersuite.google.com/app/apikey

# 2. Set API key
export GEMINI_API_KEY=AIza...

# 3. Use it
./mcp-cli query "Explain quantum computing" --provider gemini
```

---

## Overview

### What is Gemini?

Gemini is Google's flagship AI model family, offering massive 2M token context windows and strong multimodal capabilities. The public API provides simple access without needing Google Cloud Platform.

### When to Use

- ✅ **Use when**:
  - You need **2M token context** (entire codebases, long documents)
  - You want **multimodal** (text, images, video)
  - You need **Google integration** (simpler than Vertex AI)
  - You want **competitive pricing**
  - You need **free tier** for experimentation

- ❌ **Avoid when**:
  - You need **enterprise features** (use Vertex AI instead)
  - You need **best reasoning** (Claude is better)
  - You need **VPC/compliance** (use Vertex AI)

### Key Features

- **Chat/Completion**: ✅ Excellent
- **Streaming**: ✅ Fully Supported
- **Tool Calling (MCP)**: ✅ Supported
- **Embeddings**: ✅ Available
- **Vision**: ✅ Native multimodal
- **Long Context**: ✅✅ 2M tokens!

---

## Prerequisites

- [ ] **Google Account**
- [ ] **Gemini API Key** ([Get key](https://makersuite.google.com/app/apikey))

---

## Setup Guide

### Step 1: Get API Key

1. Go to https://makersuite.google.com/app/apikey
2. Click **Create API Key**
3. Select or create project
4. Copy API key (starts with `AIza`)

### Step 2: Set Environment Variable

```bash
export GEMINI_API_KEY=AIzaSy...
```

### Step 3: Test

```bash
./mcp-cli init
# Select "Gemini"

./mcp-cli query "What is the capital of France?" --provider gemini
```

---

## Configuration Reference

**File**: `config/providers/gemini.yaml`

```yaml
interface_type: openai_compatible
provider_name: gemini
config:
  api_key: ${GEMINI_API_KEY}
  api_endpoint: https://generativelanguage.googleapis.com
  default_model: gemini-2.0-flash-exp
  timeout_seconds: 300
  max_retries: 2
  context_window: 1000000
  reserve_tokens: 8000
```

### Available Models

**Gemini 2.0** (Latest):
- `gemini-2.0-flash-exp` - Fast, multimodal (recommended)

**Gemini 1.5** (Stable):
- `gemini-1.5-pro` - 2M context, most capable
- `gemini-1.5-flash` - Fast, 1M context

**Gemini 1.0** (Legacy):
- `gemini-pro` - Previous generation

---

## Implementation Details

**Interface**: OpenAI-Compatible (via Google AI API)  
**Authentication**: API Key in URL parameter  
**Endpoint**: `https://generativelanguage.googleapis.com/v1beta/...`

**Note**: Public Gemini API is simpler than Vertex AI - no GCP project needed!

---

## Features

### Massive Context Windows

```bash
# Analyze entire codebases (2M tokens = ~1.5M words)
./mcp-cli query "Review this entire repository for security issues" \
  --provider gemini \
  --model gemini-1.5-pro \
  --input huge-codebase.txt
```

**Use cases**:
- Analyze complete books
- Review full codebases
- Process long transcripts
- Multi-document analysis

### Multimodal

```bash
# Image + text (when image support added)
./mcp-cli query "Describe this architecture diagram" \
  --provider gemini \
  --image diagram.png
```

### Tool Calling

```bash
# Gemini can use MCP tools
./mcp-cli query "Search for AI news and create a summary" --provider gemini
```

---

## Usage Examples

### Example 1: Document Analysis

```bash
# Process entire PDF/book
./mcp-cli query "Summarize all key points from this 500-page document" \
  --provider gemini \
  --model gemini-1.5-pro
```

### Example 2: Codebase Review

```bash
# Analyze full project
./mcp-cli query "Find all SQL injection vulnerabilities in this codebase" \
  --provider gemini \
  --model gemini-1.5-pro
```

---

## Cost & Limits

### Pricing (Dec 2024)

**Gemini 2.0 Flash**:
- Input: $0.00 per 1M tokens (FREE!)
- Output: $0.00 per 1M tokens (FREE!)
- Rate limit: 15 requests/min (free tier)

**Gemini 1.5 Pro** (after free tier):
- Input: $1.25 per 1M tokens (128K+ context)
- Output: $5.00 per 1M tokens
- Under 128K: $3.50/$10.50

**Gemini 1.5 Flash**:
- Input: $0.075 per 1M tokens (128K+ context)
- Output: $0.30 per 1M tokens
- Under 128K: $0.19/$0.38

### Free Tier

- **15 requests per minute**
- **1 million tokens per day**
- **1,500 requests per day**

Great for learning and experimentation!

### Context Window Pricing

| Model | <128K | 128K+ |
|-------|-------|-------|
| Pro | Cheaper | Standard |
| Flash | Cheaper | Standard |

---

## Troubleshooting

### Issue: Rate Limit (Free Tier)

**Symptoms**:
```
Error: 429 Resource exhausted
```

**Solution**:
- Free tier: 15 req/min
- Upgrade to paid tier
- Add billing to project

### Issue: API Key Invalid

**Symptoms**:
```
Error: 400 API key not valid
```

**Solution**:
```bash
# Verify key starts with "AIza"
echo $GEMINI_API_KEY

# Generate new key at:
# https://makersuite.google.com/app/apikey
```

---

## Related Resources

- **Google AI Studio**: https://makersuite.google.com/
- **Documentation**: https://ai.google.dev/docs
- **Pricing**: https://ai.google.dev/pricing
- **Models**: https://ai.google.dev/models

---

## Provider Comparison

**vs Vertex AI**:
- ✅ Pros: Simpler setup, free tier, no GCP needed
- ❌ Cons: No enterprise features, no VPC, no fine-tuning

**vs Claude**:
- ✅ Pros: 10x longer context (2M vs 200K), cheaper
- ❌ Cons: Claude better at reasoning, no guaranteed quality

**vs OpenAI**:
- ✅ Pros: Longer context (2M vs 128K), free tier
- ❌ Cons: GPT-4o may be better quality

**Best for**: Long documents, large codebases, budget users

---

**Last Updated**: December 28, 2024  
**Verified Against**: mcp-cli v1.0.0  
**Code Reference**: `internal/providers/ai/clients/openai_compatible.go`

**Note**: For enterprise use with Google Cloud, see [Vertex AI](vertex-ai.md)
