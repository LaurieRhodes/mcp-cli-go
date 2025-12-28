# OpenAI

> **Skill Level**: 🟢 Beginner-Friendly  
> **Interface**: OpenAI-Compatible (Native)  
> **Best For**: Latest GPT models, production applications, comprehensive features

## Quick Start

**For Beginners**: Get started with GPT-4 in 3 steps.

```bash
# 1. Get API key from OpenAI
# Sign up at: https://platform.openai.com/signup

# 2. Set your API key
export OPENAI_API_KEY=sk-...your-key-here

# 3. Use with mcp-cli
./mcp-cli query "Explain quantum computing simply" --provider openai
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

### What is OpenAI?

OpenAI is the creator of GPT (Generative Pre-trained Transformer) models, including GPT-4, GPT-4o, and GPT-4o-mini. They provide industry-leading language models through a simple API that powers applications worldwide.

### When to Use

- ✅ **Use when**:
  - You need **state-of-the-art** language models
  - You want **proven reliability** (99.9% uptime SLA)
  - You need **comprehensive features** (vision, function calling, embeddings)
  - You're building **production applications**
  - You want **extensive documentation** and community support

- ❌ **Avoid when**:
  - You need **zero cost** (use Ollama)
  - You require **data privacy** (data sent to OpenAI)
  - You want **longer context** (Claude has 200K, Gemini has 2M)
  - You're testing/learning on a budget (Ollama is free)

- 🤔 **Consider alternatives**:
  - **Anthropic Claude**: Better reasoning, longer context
  - **Ollama**: Free, local, private
  - **DeepSeek**: Much cheaper, good for coding

### Key Features

- **Chat/Completion**: ✅ Industry Standard
- **Streaming**: ✅ Fully Supported
- **Tool Calling (MCP)**: ✅ Fully Supported (Function Calling)
- **Embeddings**: ✅ Best-in-class (text-embedding-3)
- **Vision**: ✅ GPT-4o, GPT-4 Turbo
- **Audio**: ✅ Whisper (transcription), TTS (text-to-speech)

---

## Prerequisites

### Required

- [ ] **OpenAI Account** ([Sign up](https://platform.openai.com/signup))
- [ ] **API Key** ([Get key](https://platform.openai.com/api-keys))
- [ ] **Payment Method** (required after free trial)

### Optional

- [ ] **Organization ID** (for team accounts)
- [ ] **Project Key** (for organization isolation)

### Cost Awareness

⚠️ **OpenAI is pay-per-use**. See [Cost & Limits](#cost--limits) section.

💡 **Free trial**: $5 credit for new accounts (expires after 3 months)

---

## Setup Guide

### Step 1: Create OpenAI Account

**For Beginners**: This is like signing up for any online service.

1. Go to https://platform.openai.com/signup
2. Sign up with email or Google/Microsoft account
3. Verify your email
4. Complete profile

**What this does**: Creates your OpenAI developer account.

### Step 2: Add Payment Method

**For Beginners**: Required to use the API after free trial.

1. Go to [Billing](https://platform.openai.com/account/billing/overview)
2. Click **Add payment method**
3. Enter credit card information
4. Set up **usage limits** to avoid surprises

💡 **Tip**: Set a monthly limit (e.g., $10/month) to control costs.

**What this does**: Enables API access and sets spending limits.

### Step 3: Create API Key

**For Beginners**: This is like a password for your application.

1. Go to [API Keys](https://platform.openai.com/api-keys)
2. Click **Create new secret key**
3. Name it (e.g., "mcp-cli")
4. Copy the key (starts with `sk-`)
5. **Save it securely** - you can't see it again!

⚠️ **Security**: Never share your API key or commit it to git!

**What this does**: Creates credentials for API access.

### Step 4: Set Environment Variable

**For Beginners**: This tells mcp-cli how to authenticate.

```bash
# Add to your .env file
export OPENAI_API_KEY=sk-proj-...your-key-here
```

Or in `.env` file:
```bash
OPENAI_API_KEY=sk-proj-abc123...
```

**What this does**: Makes your API key available to mcp-cli.

### Step 5: Test Connection

```bash
# Initialize configuration
./mcp-cli init
# Select "OpenAI" when prompted

# Test query
./mcp-cli query "What is the capital of France?" --provider openai

# Expected output: "The capital of France is Paris."
```

---

## Configuration Reference

### Provider Configuration

**File**: `config/providers/openai.yaml`

```yaml
interface_type: openai_compatible
provider_name: openai
config:
  # API credentials (required)
  api_key: ${OPENAI_API_KEY}
  
  # API endpoint (default OpenAI)
  api_endpoint: https://api.openai.com/v1
  
  # Default model
  default_model: gpt-4o-mini
  
  # Request settings
  timeout_seconds: 300
  max_retries: 2
  
  # Context management
  context_window: 128000
  reserve_tokens: 4000
```

#### Configuration Options

| Option | Type | Required | Default | Description |
|--------|------|----------|---------|-------------|
| `api_key` | string | Yes | - | Your OpenAI API key |
| `api_endpoint` | string | No | OpenAI API | API endpoint URL |
| `default_model` | string | Yes | `gpt-4o-mini` | Default model to use |
| `timeout_seconds` | int | No | 300 | Request timeout (5 min) |
| `max_retries` | int | No | 2 | Retry attempts on failure |
| `context_window` | int | No | 128000 | Max input tokens |
| `reserve_tokens` | int | No | 4000 | Reserved for output |
| `organization` | string | No | - | Organization ID (optional) |

#### Environment Variables

| Variable | Required | Description | Example |
|----------|----------|-------------|---------|
| `OPENAI_API_KEY` | Yes | Your OpenAI API key | `sk-proj-...` |
| `OPENAI_ORG_ID` | No | Organization ID | `org-...` |

#### Available Models

**Recommended Models** (Dec 2024):

**General Purpose**:
- `gpt-4o-mini` - Fast, affordable, very capable (recommended)
- `gpt-4o` - Most capable, multimodal
- `gpt-4-turbo` - Previous generation flagship

**Specialized**:
- `gpt-3.5-turbo` - Legacy, very fast and cheap
- `o1-preview` - Advanced reasoning (limited availability)
- `o1-mini` - Reasoning optimized for coding

### Embedding Configuration

**File**: `config/embeddings/openai.yaml`

```yaml
interface_type: openai_compatible
provider_name: openai
config:
  api_key: ${OPENAI_API_KEY}
  api_endpoint: https://api.openai.com/v1
  default_model: text-embedding-3-small
  embedding_models:
    text-embedding-3-small:
      description: Most capable embedding model
      dimensions: 1536
      max_tokens: 8191
    text-embedding-3-large:
      description: Higher performance, larger embeddings
      dimensions: 3072
      max_tokens: 8191
    text-embedding-ada-002:
      description: Legacy embedding model
      dimensions: 1536
      max_tokens: 8191
```

---

## Implementation Details

### Interface Type

**Type**: `openai_compatible`

**What this means**:
- **For Beginners**: This is OpenAI's own API - the standard that others copy
- **For Developers**: Uses OpenAI's native `/v1/chat/completions` endpoint format

### How It Works

```
User → mcp-cli → OpenAI Client → OpenAI API → GPT Model
                     ↓
              OpenAI format (native)
```

**Architecture**:
1. mcp-cli sends requests in OpenAI format
2. OpenAI API receives at `https://api.openai.com/v1/chat/completions`
3. Request routed to appropriate GPT model
4. Model generates response
5. Response returned in OpenAI format

### API Endpoints

- **Chat/Completion**: `https://api.openai.com/v1/chat/completions`
- **Embeddings**: `https://api.openai.com/v1/embeddings`
- **Images**: `https://api.openai.com/v1/images/generations`
- **Audio**: `https://api.openai.com/v1/audio/transcriptions`

### Authentication

**Method**: Bearer Token (API Key)

```
Authorization: Bearer sk-proj-...
```

Simple and secure - just include your API key in the header.

### Model Name Format

**Format**: `model-name`

Examples:
- `gpt-4o-mini` - GPT-4 Optimized Mini
- `gpt-4o` - GPT-4 Optimized
- `gpt-4-turbo` - GPT-4 Turbo

**List all models**:
```bash
curl https://api.openai.com/v1/models \
  -H "Authorization: Bearer $OPENAI_API_KEY"
```

---

## Features

### Chat & Completions

```bash
# Basic query
./mcp-cli query "Explain photosynthesis" --provider openai

# With system prompt
./mcp-cli query "Review this code" \
  --system "You are a senior software engineer" \
  --provider openai

# Specific model
./mcp-cli query "Complex reasoning task" \
  --provider openai \
  --model gpt-4o
```

**Model Selection**:
- `gpt-4o-mini`: Fast, cheap, 95% as good as GPT-4
- `gpt-4o`: Best quality, multimodal
- `gpt-4-turbo`: Previous flagship

### Streaming

```bash
# Stream responses in real-time
./mcp-cli query "Write a short story" \
  --provider openai \
  --stream
```

**Status**: ✅ Fully Supported
**Use case**: Better UX for long responses

### Tool Calling (MCP Tools)

**Status**: ✅ Fully Supported (Function Calling)

```bash
# Automatic tool use
./mcp-cli query "Search for latest AI news and summarize" \
  --provider openai \
  --verbose

# Multiple tool calls
./mcp-cli query "What's the weather in Paris and convert 20°C to Fahrenheit" \
  --provider openai
```

**How it works**:
1. mcp-cli advertises available tools
2. GPT decides when to call tools
3. mcp-cli executes tool
4. Results sent back to GPT
5. GPT synthesizes final answer

**Best models for tools**:
- `gpt-4o` - Excellent tool use
- `gpt-4o-mini` - Good, more affordable
- `gpt-4-turbo` - Very reliable

### Embeddings

**Status**: ✅ Best-in-class

```bash
# Generate embeddings
./mcp-cli embed "Your text here" \
  --provider openai \
  --model text-embedding-3-small

# Higher performance
./mcp-cli embed "Your text here" \
  --provider openai \
  --model text-embedding-3-large
```

**Models**:
- `text-embedding-3-small`: 1536 dims, excellent quality, affordable
- `text-embedding-3-large`: 3072 dims, highest quality, more expensive
- `text-embedding-ada-002`: Legacy, still good

**Use cases**:
- Semantic search
- RAG (Retrieval-Augmented Generation)
- Clustering
- Recommendations

### Vision

**Status**: ✅ Supported (GPT-4o, GPT-4 Turbo)

```bash
# Analyze image (when image support added to mcp-cli)
./mcp-cli query "Describe this image in detail" \
  --provider openai \
  --model gpt-4o \
  --image path/to/image.jpg
```

**Capabilities**:
- Image understanding
- OCR (text extraction)
- Object detection
- Scene description
- Chart/diagram analysis

---

## Usage Examples

### Example 1: Content Generation

**Scenario**: Generate blog post outline

```bash
./mcp-cli query "Create a detailed outline for a blog post about AI in healthcare. Include 5 main sections with 3 subsections each." \
  --provider openai \
  --model gpt-4o-mini
```

### Example 2: Code Review with Tools

**Scenario**: Review code and search for best practices

```bash
./mcp-cli query "Review this Python function and search for modern Python best practices:

def process_data(data):
    result = []
    for item in data:
        if item > 0:
            result.append(item * 2)
    return result
" --provider openai --model gpt-4o
```

GPT will search for best practices and provide recommendations.

### Example 3: Semantic Search Setup

**Scenario**: Build document search system

```bash
# Generate embeddings for documents
for doc in docs/*.txt; do
  ./mcp-cli embed "$(cat $doc)" \
    --provider openai \
    --model text-embedding-3-small \
    > "embeddings/$(basename $doc .txt).json"
done

# Use in vector database (Pinecone, Weaviate, etc.)
```

### Example 4: Multi-Step Reasoning

**Scenario**: Complex problem solving

```bash
./mcp-cli query "I have \$10,000 to invest. Research current market conditions, compare stocks vs bonds, and recommend an allocation strategy for a moderate-risk investor." \
  --provider openai \
  --model gpt-4o
```

GPT will use web search tool to get current data.

### Example 5: Data Analysis

**Scenario**: Analyze CSV data

```bash
./mcp-cli query "Analyze this sales data and provide insights:
Q1: \$45K
Q2: \$52K
Q3: \$48K
Q4: \$61K

Identify trends and make predictions for next quarter." \
  --provider openai \
  --model gpt-4o-mini
```

---

## Troubleshooting

### Common Issues

#### Issue: Invalid API Key

**Symptoms**:
```
Error: 401 Unauthorized - Incorrect API key provided
```

**Cause**: API key is wrong, expired, or not set

**Solution**:
```bash
# Verify API key is set
echo $OPENAI_API_KEY

# Should start with "sk-proj-" or "sk-"
# If not set:
export OPENAI_API_KEY=sk-proj-...

# Test with curl
curl https://api.openai.com/v1/models \
  -H "Authorization: Bearer $OPENAI_API_KEY"
```

#### Issue: Rate Limit Exceeded

**Symptoms**:
```
Error: 429 Rate limit exceeded
```

**Cause**: Too many requests in short time

**Solution**:
```bash
# Wait a minute and retry
# Or upgrade your tier at:
# https://platform.openai.com/account/limits

# Check your limits
curl https://api.openai.com/v1/usage \
  -H "Authorization: Bearer $OPENAI_API_KEY"
```

**Rate limits by tier**:
- Free: 3 requests/min
- Tier 1: 3,500 requests/min
- Tier 2+: Higher limits

#### Issue: Insufficient Quota

**Symptoms**:
```
Error: 429 You exceeded your current quota
```

**Cause**: Used all credits or hit spending limit

**Solution**:
1. Go to [Billing](https://platform.openai.com/account/billing/overview)
2. Add payment method or increase limit
3. Check usage to understand costs

```bash
# Check usage
curl https://api.openai.com/v1/usage?date=2024-12-28 \
  -H "Authorization: Bearer $OPENAI_API_KEY"
```

#### Issue: Model Not Found

**Symptoms**:
```
Error: 404 Model 'gpt-5' does not exist
```

**Cause**: Model name is incorrect or not available

**Solution**:
```bash
# Use correct model names
# ✅ gpt-4o-mini
# ✅ gpt-4o
# ✅ gpt-4-turbo
# ❌ gpt-5 (doesn't exist)

# List available models
curl https://api.openai.com/v1/models \
  -H "Authorization: Bearer $OPENAI_API_KEY" \
  | jq '.data[].id'
```

#### Issue: Context Length Exceeded

**Symptoms**:
```
Error: maximum context length is 128000 tokens
```

**Cause**: Input + output would exceed model's context window

**Solution**:
```bash
# Reduce input length
# Or use model with larger context:
# - gpt-4-turbo: 128K tokens
# - gpt-4o: 128K tokens

# Or split into multiple requests
```

#### Issue: Timeout

**Symptoms**:
```
Error: Request timeout after 300 seconds
```

**Cause**: Request took too long

**Solution**:
```yaml
# Increase timeout in config
# config/providers/openai.yaml
config:
  timeout_seconds: 600  # 10 minutes
```

---

## Cost & Limits

### Pricing (Dec 2024)

**GPT-4o Mini** (Recommended for most use cases):
- Input: $0.150 per 1M tokens
- Output: $0.600 per 1M tokens
- **~$0.001 per query** (typical)

**GPT-4o**:
- Input: $2.50 per 1M tokens
- Output: $10.00 per 1M tokens
- **~$0.015 per query** (typical)

**GPT-4 Turbo**:
- Input: $10.00 per 1M tokens
- Output: $30.00 per 1M tokens

**Embeddings**:
- text-embedding-3-small: $0.020 per 1M tokens
- text-embedding-3-large: $0.130 per 1M tokens

**Vision** (GPT-4o):
- Same as text pricing
- Images count as ~85-170 tokens (low detail)
- Images count as ~255-765 tokens (high detail)

### Rate Limits

**By Tier**:

| Tier | Requests/min | Tokens/min | How to Qualify |
|------|--------------|------------|----------------|
| Free | 3 | 40K | New accounts |
| Tier 1 | 3,500 | 800K | $5+ spent |
| Tier 2 | 3,500 | 800K | $50+ spent, 7+ days |
| Tier 3 | 5,000 | 800K | $100+ spent, 7+ days |
| Tier 4 | 10,000 | 800K | $250+ spent, 14+ days |
| Tier 5 | 10,000 | 2M | $1,000+ spent, 30+ days |

### Context Windows

| Model | Context | Output | Best For |
|-------|---------|--------|----------|
| `gpt-4o-mini` | 128K | 16K | General use, cost-effective |
| `gpt-4o` | 128K | 16K | Complex tasks, vision |
| `gpt-4-turbo` | 128K | 4K | Reliability, JSON mode |
| `gpt-3.5-turbo` | 16K | 4K | Simple tasks, very fast |

### Cost Optimization Tips

1. **Start with gpt-4o-mini** - 95% as good, much cheaper
2. **Use streaming** - Better UX, same cost
3. **Cache system prompts** - Reduce repeated context
4. **Set token limits** - Use `reserve_tokens` config
5. **Monitor usage** - Check dashboard regularly
6. **Batch requests** - Group similar queries
7. **Use appropriate models** - Don't use GPT-4o for simple tasks

**Example costs for 1000 queries**:
- GPT-4o-mini: ~$1
- GPT-4o: ~$15
- GPT-4-turbo: ~$40

---

## Related Resources

- **OpenAI Platform**: https://platform.openai.com/
- **API Documentation**: https://platform.openai.com/docs/api-reference
- **Pricing**: https://openai.com/pricing
- **Models**: https://platform.openai.com/docs/models
- **Usage Dashboard**: https://platform.openai.com/usage
- **Status Page**: https://status.openai.com/
- **Community Forum**: https://community.openai.com/
- **Cookbook**: https://cookbook.openai.com/

### Best Practices

- **Prompt Engineering**: https://platform.openai.com/docs/guides/prompt-engineering
- **Function Calling**: https://platform.openai.com/docs/guides/function-calling
- **Safety Best Practices**: https://platform.openai.com/docs/guides/safety-best-practices
- **Rate Limits**: https://platform.openai.com/docs/guides/rate-limits

---

## Provider Comparison

**vs Anthropic (Claude)**:
- ✅ Pros: Faster, cheaper (mini), better embeddings, more features (vision, audio)
- ❌ Cons: Shorter context (128K vs 200K), Claude may reason better for complex tasks

**vs Ollama**:
- ✅ Pros: Latest models, no hardware needed, official support
- ❌ Cons: Costs money, requires internet, data sent to OpenAI

**vs Gemini**:
- ✅ Pros: Better embeddings, more mature API, better docs
- ❌ Cons: Shorter context (128K vs 2M), more expensive

**vs DeepSeek**:
- ✅ Pros: Better quality, official support, proven track record
- ❌ Cons: More expensive (but gpt-4o-mini is competitive)

---

**Last Updated**: December 28, 2024  
**Verified Against**: mcp-cli v1.0.0, OpenAI API Dec 2024  
**Code Reference**: `internal/providers/ai/clients/openai_compatible.go`

**Implementation Notes**:
- Native OpenAI format (reference implementation)
- Full tool calling support: ✅ Verified
- Streaming: ✅ Verified
- Embeddings: ✅ Verified
- Rate limit handling: ✅ Implemented with retries
