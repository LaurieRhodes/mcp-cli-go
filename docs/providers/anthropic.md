# Anthropic (Claude)

> **Skill Level**: 🟢 Beginner-Friendly  
> **Interface**: Anthropic Native  
> **Best For**: Complex reasoning, coding, analysis, long-form content

## Quick Start

**For Beginners**: Access Claude in 3 steps.

```bash
# 1. Get API key from Anthropic
# Sign up at: https://console.anthropic.com/

# 2. Set your API key
export ANTHROPIC_API_KEY=sk-ant-...your-key-here

# 3. Use with mcp-cli
./mcp-cli query "Explain quantum computing" --provider anthropic
```

---

## Table of Contents

- [Overview](#overview)
- [Prerequisites](#prerequisites)
- [Setup Guide](#setup-guide)
- [Configuration Reference](#configuration-reference)
- [Implementation Details](#implementation-details)
- [Features](#features)
- [Usage Examples](#usage-examples)
- [Troubleshooting](#troubleshooting)
- [Cost & Limits](#cost--limits)
- [Related Resources](#related-resources)

---

## Overview

### What is Anthropic?

Anthropic is the creator of Claude, an AI assistant focused on being helpful, harmless, and honest. Claude is known for exceptional reasoning, coding ability, and following complex instructions accurately.

### When to Use

- ✅ **Use when**:
  - You need **best-in-class reasoning** and analysis
  - You want **superior code generation** and review
  - You need **200K context window** (entire codebases)
  - You require **careful, nuanced responses**
  - You want **excellent tool calling** (function use)
  - You need **ethical AI** with strong safety guardrails

- ❌ **Avoid when**:
  - You need **zero cost** (use Ollama)
  - You need **embeddings** (Claude doesn't offer them - use OpenAI)
  - You want **vision** on a budget (GPT-4o is cheaper)
  - You need **very fast responses** (GPT-4o-mini is faster)

- 🤔 **Consider alternatives**:
  - **OpenAI GPT-4o**: Faster, has embeddings, vision
  - **Ollama**: Free, local, private
  - **DeepSeek**: Much cheaper for coding

### Key Features

- **Chat/Completion**: ✅ Industry-leading quality
- **Streaming**: ✅ Fully Supported
- **Tool Calling (MCP)**: ✅ Excellent (best-in-class)
- **Embeddings**: ❌ Not Available
- **Vision**: ✅ Claude 3.5 Sonnet/Opus (image analysis)
- **Long Context**: ✅ 200K tokens (huge context window)

---

## Prerequisites

### Required

- [ ] **Anthropic Account** ([Sign up](https://console.anthropic.com/))
- [ ] **API Key** ([Get key](https://console.anthropic.com/settings/keys))
- [ ] **Payment Method** (after free trial)

### Optional

- [ ] **Organization settings** (for teams)

### Cost Awareness

⚠️ **Claude is pay-per-use**. See [Cost & Limits](#cost--limits) section.

💡 **Free trial**: $5 credit for new accounts

---

## Setup Guide

### Step 1: Create Anthropic Account

1. Go to https://console.anthropic.com/
2. Sign up with email or Google
3. Verify email
4. Complete setup

### Step 2: Add Payment Method

1. Go to [Billing](https://console.anthropic.com/settings/billing)
2. Click **Add payment method**
3. Enter credit card
4. Set **usage limits** (recommended)

💡 **Tip**: Start with a $20/month limit

### Step 3: Create API Key

1. Go to [API Keys](https://console.anthropic.com/settings/keys)
2. Click **Create Key**
3. Name it (e.g., "mcp-cli")
4. Copy key (starts with `sk-ant-`)
5. **Save securely** - can't view again!

⚠️ **Security**: Never commit API key to git!

### Step 4: Set Environment Variable

```bash
# Add to .env file
export ANTHROPIC_API_KEY=sk-ant-api03-...your-key-here
```

Or in `.env`:
```bash
ANTHROPIC_API_KEY=sk-ant-api03-...
```

### Step 5: Test Connection

```bash
# Initialize configuration
./mcp-cli init
# Select "Anthropic" when prompted

# Test query
./mcp-cli query "What is the capital of France?" --provider anthropic

# Expected: "The capital of France is Paris."
```

---

## Configuration Reference

### Provider Configuration

**File**: `config/providers/anthropic.yaml`

```yaml
interface_type: anthropic_native
provider_name: anthropic
config:
  # API credentials (required)
  api_key: ${ANTHROPIC_API_KEY}
  
  # Default model
  default_model: claude-3-5-sonnet-20241022
  
  # Request settings
  timeout_seconds: 300
  max_retries: 5
```

#### Configuration Options

| Option | Type | Required | Default | Description |
|--------|------|----------|---------|-------------|
| `api_key` | string | Yes | - | Anthropic API key |
| `default_model` | string | Yes | `claude-3-5-sonnet-20241022` | Default Claude model |
| `timeout_seconds` | int | No | 300 | Request timeout |
| `max_retries` | int | No | 5 | Retry attempts |

#### Environment Variables

| Variable | Required | Description | Example |
|----------|----------|-------------|---------|
| `ANTHROPIC_API_KEY` | Yes | Your Anthropic API key | `sk-ant-api03-...` |

#### Available Models

**Current Models** (Dec 2024):

**Claude 3.5 Family** (Latest):
- `claude-3-5-sonnet-20241022` - Best balance (recommended)
- `claude-3-5-haiku-20241022` - Fast, affordable

**Claude 3 Family** (Legacy):
- `claude-3-opus-20240229` - Most capable (expensive)
- `claude-3-sonnet-20240229` - Previous version
- `claude-3-haiku-20240307` - Previous fast version

**Recommended**: Use `claude-3-5-sonnet-20241022` for best results.

---

## Implementation Details

### Interface Type

**Type**: `anthropic_native`

**What this means**:
- **For Beginners**: Uses Anthropic's own API format (not OpenAI-compatible)
- **For Developers**: Native Messages API with Anthropic-specific features

### How It Works

```
User → mcp-cli → Anthropic Client → Anthropic API → Claude Model
                     ↓
              Anthropic format (native)
```

**Architecture**:
1. mcp-cli converts to Anthropic Messages format
2. Request sent to `https://api.anthropic.com/v1/messages`
3. Anthropic routes to Claude model
4. Claude generates response
5. Response in Anthropic format, converted to domain format

### API Endpoints

- **Messages**: `https://api.anthropic.com/v1/messages`
- **Streaming**: Same endpoint with `stream: true`

### Authentication

**Method**: API Key Header

```
x-api-key: sk-ant-api03-...
anthropic-version: 2023-06-01
```

### Model Name Format

**Format**: `claude-{version}-{size}-{date}`

Examples:
- `claude-3-5-sonnet-20241022` - Claude 3.5 Sonnet, Oct 22 2024
- `claude-3-5-haiku-20241022` - Claude 3.5 Haiku, Oct 22 2024
- `claude-3-opus-20240229` - Claude 3 Opus, Feb 29 2024

---

## Features

### Chat & Completions

```bash
# Basic query
./mcp-cli query "Explain machine learning" --provider anthropic

# With system prompt
./mcp-cli query "Review this code" \
  --system "You are a senior software architect" \
  --provider anthropic

# Specific model
./mcp-cli query "Complex analysis task" \
  --provider anthropic \
  --model claude-3-5-sonnet-20241022
```

**Claude's Strengths**:
- Detailed, thorough responses
- Follows complex instructions precisely
- Excellent at code generation and review
- Strong reasoning for multi-step problems

### Streaming

```bash
# Stream responses in real-time
./mcp-cli query "Write a technical article about microservices" \
  --provider anthropic \
  --stream
```

**Status**: ✅ Fully Supported

### Tool Calling (MCP Tools)

**Status**: ✅ Excellent (Best-in-class)

```bash
# Claude excels at tool use
./mcp-cli query "Search for Python best practices for async programming, then analyze the top 3 recommendations" \
  --provider anthropic \
  --verbose

# Complex multi-tool workflows
./mcp-cli query "Research current AI trends, compare with last year, and create a summary report" \
  --provider anthropic
```

**Why Claude is best at tools**:
- Decides when to use tools intelligently
- Combines multiple tool results effectively
- Understands tool context deeply
- Minimal hallucination with tool outputs

### Vision

**Status**: ✅ Supported (Claude 3.5 Sonnet, Opus)

```bash
# Image analysis (when image support added)
./mcp-cli query "Analyze this diagram and explain the architecture" \
  --provider anthropic \
  --model claude-3-5-sonnet-20241022 \
  --image architecture.png
```

**Capabilities**:
- Technical diagram analysis
- Code in screenshots (OCR)
- Chart/graph interpretation
- Document understanding

### Long Context

**Status**: ✅ 200K tokens

```bash
# Analyze entire codebases
./mcp-cli query "Review this entire codebase for security issues" \
  --provider anthropic \
  --model claude-3-5-sonnet-20241022 \
  --input large-codebase.txt
```

**Use cases**:
- Analyze complete projects
- Review long documents
- Multi-file code review
- Legal document analysis

---

## Usage Examples

### Example 1: Code Review with Deep Analysis

**Scenario**: Comprehensive code review

```bash
./mcp-cli query "Review this Python code for:
1. Security vulnerabilities
2. Performance issues
3. Best practices
4. Potential bugs
5. Improvement suggestions

\`\`\`python
def process_user_data(data):
    results = []
    for user in data:
        if user['age'] > 18:
            results.append(user)
    return results
\`\`\`" --provider anthropic --model claude-3-5-sonnet-20241022
```

**Why Claude**: Thorough, catches subtle issues, great explanations

### Example 2: Complex Problem Solving

**Scenario**: Multi-step reasoning

```bash
./mcp-cli query "I need to design a distributed caching system for an e-commerce platform with these requirements:
- 100K requests/second
- Sub-10ms latency
- Multi-region deployment
- High availability

Analyze the requirements, research current solutions, compare Redis vs Memcached, and provide a detailed architecture recommendation." \
  --provider anthropic
```

**Why Claude**: Best-in-class reasoning, thorough analysis

### Example 3: Technical Writing

**Scenario**: Create documentation

```bash
./mcp-cli query "Write comprehensive API documentation for a REST endpoint that creates user accounts. Include:
- Endpoint details
- Request/response examples
- Error codes
- Rate limits
- Authentication
- Best practices" \
  --provider anthropic
```

**Why Claude**: Excellent at structured, detailed content

### Example 4: Debugging Assistance

**Scenario**: Debug complex issue

```bash
./mcp-cli query "I'm getting this error in my Node.js app:
'UnhandledPromiseRejectionWarning: Error: connect ETIMEDOUT'

The app connects to MongoDB. It works locally but fails in production. Help me debug this systematically." \
  --provider anthropic
```

**Why Claude**: Systematic debugging approach, asks good questions

### Example 5: Learning Complex Topics

**Scenario**: Understand difficult concepts

```bash
./mcp-cli query "Explain how blockchain consensus mechanisms work, specifically comparing:
- Proof of Work
- Proof of Stake
- Byzantine Fault Tolerance

Use analogies for each and explain trade-offs." \
  --provider anthropic
```

**Why Claude**: Clear explanations, good analogies, thorough

---

## Troubleshooting

### Common Issues

#### Issue: Invalid API Key

**Symptoms**:
```
Error: 401 Unauthorized - Invalid API Key
```

**Cause**: API key incorrect or expired

**Solution**:
```bash
# Verify key is set
echo $ANTHROPIC_API_KEY

# Should start with "sk-ant-"
# Get new key at: https://console.anthropic.com/settings/keys
```

#### Issue: Rate Limit Exceeded

**Symptoms**:
```
Error: 429 Rate limit exceeded
```

**Cause**: Too many requests

**Solution**:
```bash
# Wait and retry (automatic with max_retries)
# Check rate limits at: https://console.anthropic.com/settings/limits

# Limits vary by tier (see Cost & Limits section)
```

#### Issue: Overloaded API

**Symptoms**:
```
Error: 529 Service temporarily overloaded
```

**Cause**: Anthropic API experiencing high load

**Solution**:
```bash
# Automatic retry with exponential backoff
# Configured in max_retries setting

# If persistent, check status:
# https://status.anthropic.com/
```

#### Issue: Context Too Long

**Symptoms**:
```
Error: prompt is too long: 205000 tokens > 200000 maximum
```

**Cause**: Input exceeds 200K token limit

**Solution**:
```bash
# Reduce input size
# Or split into multiple requests
# Or summarize parts of the input first
```

#### Issue: Model Not Found

**Symptoms**:
```
Error: model not found: claude-4
```

**Cause**: Invalid model name

**Solution**:
```bash
# Use correct model names:
# ✅ claude-3-5-sonnet-20241022
# ✅ claude-3-5-haiku-20241022
# ✅ claude-3-opus-20240229
# ❌ claude-4 (doesn't exist)
```

---

## Cost & Limits

### Pricing (Dec 2024)

**Claude 3.5 Sonnet** (Recommended):
- Input: $3.00 per 1M tokens
- Output: $15.00 per 1M tokens
- **~$0.003 per query** (typical)

**Claude 3.5 Haiku** (Fast & Affordable):
- Input: $0.80 per 1M tokens
- Output: $4.00 per 1M tokens
- **~$0.001 per query** (typical)

**Claude 3 Opus** (Most Capable):
- Input: $15.00 per 1M tokens
- Output: $75.00 per 1M tokens
- **~$0.015 per query** (typical)

**Vision** (additional cost):
- Images count as ~1000-2000 tokens each

### Rate Limits

**By Tier**:

| Tier | Requests/min | Tokens/min | How to Qualify |
|------|--------------|------------|----------------|
| Free | 5 | 50K | New accounts |
| Tier 1 | 50 | 50K | $5+ spent |
| Tier 2 | 1,000 | 100K | $40+ spent |
| Tier 3 | 2,000 | 200K | $200+ spent |
| Tier 4 | 4,000 | 400K | $1,000+ spent |

### Context Windows

| Model | Context | Output | Best For |
|-------|---------|--------|----------|
| `claude-3-5-sonnet` | 200K | 8K | Complex tasks, analysis |
| `claude-3-5-haiku` | 200K | 8K | Fast, cost-effective |
| `claude-3-opus` | 200K | 4K | Highest quality |

### Cost Optimization Tips

1. **Use Claude 3.5 Haiku** for simple tasks (4x cheaper)
2. **Use Sonnet** for complex reasoning (best balance)
3. **Reserve Opus** for critical tasks only
4. **Enable streaming** for better UX (same cost)
5. **Monitor usage** in console dashboard
6. **Set spending limits** to avoid surprises
7. **Cache system prompts** when possible

**Example costs for 1000 queries**:
- Haiku: ~$1
- Sonnet: ~$3
- Opus: ~$15

---

## Related Resources

- **Anthropic Console**: https://console.anthropic.com/
- **API Documentation**: https://docs.anthropic.com/
- **Pricing**: https://www.anthropic.com/pricing
- **Model Comparison**: https://docs.anthropic.com/claude/docs/models-overview
- **Prompt Engineering**: https://docs.anthropic.com/claude/docs/prompt-engineering
- **Status Page**: https://status.anthropic.com/
- **Discord Community**: https://www.anthropic.com/discord

### Best Practices

- **Prompt Library**: https://docs.anthropic.com/claude/prompt-library
- **Tool Use Guide**: https://docs.anthropic.com/claude/docs/tool-use
- **Safety Best Practices**: https://docs.anthropic.com/claude/docs/safety-best-practices

---

## Provider Comparison

**vs OpenAI (GPT-4o)**:
- ✅ Pros: Better reasoning, longer context (200K vs 128K), superior tool use, more careful/thorough
- ❌ Cons: More expensive, no embeddings, slower, no audio features

**vs Ollama**:
- ✅ Pros: Cloud-scale quality, official support, latest features
- ❌ Cons: Costs money, requires internet, data sent to Anthropic

**vs Gemini**:
- ✅ Pros: Better reasoning, superior coding, excellent tool use
- ❌ Cons: Shorter context (200K vs 2M), more expensive, no multimodal (yet)

**Best Use Cases**:
- **Complex reasoning**: ✅ Choose Claude
- **Code generation**: ✅ Choose Claude
- **Fast queries**: ❌ Use GPT-4o-mini
- **Embeddings**: ❌ Use OpenAI
- **Budget**: ❌ Use Ollama

---

**Last Updated**: December 28, 2024  
**Verified Against**: mcp-cli v1.0.0, Claude API Dec 2024  
**Code Reference**: `internal/providers/ai/clients/anthropic.go`

**Implementation Notes**:
- Native Anthropic Messages API
- Tool calling: ✅ Excellent quality verified
- Streaming: ✅ Verified
- 200K context: ✅ Tested with large inputs
- No embeddings: ❌ Use OpenAI for this
