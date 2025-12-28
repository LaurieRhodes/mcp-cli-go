# OpenRouter

> **Skill Level**: 🟢 Beginner-Friendly  
> **Interface**: OpenAI-Compatible  
> **Best For**: Access to multiple models through one API, model comparison, flexibility

## Quick Start

```bash
# 1. Get API key
# Sign up at: https://openrouter.ai/

# 2. Set API key
export OPENROUTER_API_KEY=sk-or-...

# 3. Use it
./mcp-cli query "Hello from OpenRouter!" --provider openrouter
```

---

## Overview

### What is OpenRouter?

OpenRouter is a unified API that provides access to dozens of AI models from different providers (OpenAI, Anthropic, Meta, Google, etc.) through a single interface. Pay-as-you-go pricing with automatic routing to available models.

### When to Use

- ✅ **Use when**:
  - You want **multiple model access** through one API
  - You need **model flexibility** (easy switching)
  - You want to **compare models** easily
  - You need **fallback options** (if one model is down)
  - You want **simple billing** (one invoice)
  - You don't want to manage multiple API keys

- ❌ **Avoid when**:
  - You want **lowest possible cost** (direct APIs are cheaper)
  - You need **enterprise features** (use direct providers)
  - You want **guaranteed SLAs** (depends on upstream)
  - You need **data residency** guarantees

### Key Features

- **Chat/Completion**: ✅ 50+ models available
- **Streaming**: ✅ Fully Supported
- **Tool Calling (MCP)**: ✅ Model-dependent
- **Embeddings**: ✅ Multiple providers
- **Multi-Provider**: ✅ GPT-4, Claude, Llama, Gemini, and more
- **Auto-Fallback**: ✅ Switch models if primary fails

---

## Prerequisites

- [ ] **OpenRouter Account** ([Sign up](https://openrouter.ai/))
- [ ] **API Key** ([Get key](https://openrouter.ai/keys))
- [ ] **Credits** (add funds or use free models)

---

## Setup Guide

### Step 1: Create Account

1. Go to https://openrouter.ai/
2. Sign up with email or social login
3. Verify email

### Step 2: Add Credits

1. Go to [Credits](https://openrouter.ai/credits)
2. Add payment method
3. Add credits (minimum $5)

💡 **Tip**: Some models are free! Check model pricing.

### Step 3: Create API Key

1. Go to [API Keys](https://openrouter.ai/keys)
2. Click **Create Key**
3. Copy key (starts with `sk-or-`)

### Step 4: Set Environment Variable

```bash
export OPENROUTER_API_KEY=sk-or-v1-...
```

### Step 5: Test

```bash
./mcp-cli init
# Select "OpenRouter"

./mcp-cli query "What is the capital of France?" --provider openrouter
```

---

## Configuration Reference

**File**: `config/providers/openrouter.yaml`

```yaml
interface_type: openai_compatible
provider_name: openrouter
config:
  api_key: ${OPENROUTER_API_KEY}
  api_endpoint: https://openrouter.ai/api/v1
  default_model: anthropic/claude-3.5-sonnet
  timeout_seconds: 300
  max_retries: 2
  context_window: 200000
  reserve_tokens: 2000
```

### Available Models (Popular)

**Anthropic Claude**:
- `anthropic/claude-3.5-sonnet` - Best reasoning
- `anthropic/claude-3.5-haiku` - Fast, affordable
- `anthropic/claude-3-opus` - Most capable

**OpenAI GPT**:
- `openai/gpt-4o` - Latest GPT-4
- `openai/gpt-4o-mini` - Fast, cheap
- `openai/gpt-4-turbo` - Previous flagship

**Meta Llama**:
- `meta-llama/llama-3.1-405b-instruct` - Largest
- `meta-llama/llama-3.1-70b-instruct` - Good balance
- `meta-llama/llama-3.1-8b-instruct` - Fast

**Google Gemini**:
- `google/gemini-pro-1.5` - 2M context
- `google/gemini-flash-1.5` - Fast

**Free Models**:
- `meta-llama/llama-3.1-8b-instruct:free`
- `google/gemini-flash-1.5:free`

**Full list**: https://openrouter.ai/models

---

## Implementation Details

**Interface**: OpenAI-Compatible  
**Authentication**: Bearer token  
**Endpoint**: `https://openrouter.ai/api/v1/chat/completions`

**Model format**: `provider/model-name`

---

## Features

### Multi-Model Access

```bash
# Use Claude
./mcp-cli query "Complex reasoning" \
  --provider openrouter \
  --model anthropic/claude-3.5-sonnet

# Use GPT-4
./mcp-cli query "Same query" \
  --provider openrouter \
  --model openai/gpt-4o

# Use Llama
./mcp-cli query "Same query" \
  --provider openrouter \
  --model meta-llama/llama-3.1-70b-instruct
```

### Free Models

```bash
# Use free models (no credit needed)
./mcp-cli query "Test query" \
  --provider openrouter \
  --model meta-llama/llama-3.1-8b-instruct:free

# Or Gemini Flash free
./mcp-cli query "Test query" \
  --provider openrouter \
  --model google/gemini-flash-1.5:free
```

### Model Comparison

```bash
# Compare responses easily
for model in "anthropic/claude-3.5-sonnet" "openai/gpt-4o" "meta-llama/llama-3.1-70b-instruct"
do
  echo "=== $model ==="
  ./mcp-cli query "Explain recursion" --provider openrouter --model $model
done
```

---

## Usage Examples

### Example 1: Flexible Development

```bash
# Start with free model for testing
./mcp-cli query "test" --provider openrouter --model meta-llama/llama-3.1-8b-instruct:free

# Switch to paid for production
./mcp-cli query "production query" --provider openrouter --model anthropic/claude-3.5-sonnet
```

### Example 2: Cost Optimization

```bash
# Simple tasks: Use free or cheap models
./mcp-cli query "What is Python?" --provider openrouter --model google/gemini-flash-1.5:free

# Complex tasks: Use premium models
./mcp-cli query "Design system architecture" --provider openrouter --model anthropic/claude-3.5-sonnet
```

### Example 3: Model Experimentation

```bash
# Test different models without managing multiple API keys
./mcp-cli query "Generate a sorting algorithm" \
  --provider openrouter \
  --model meta-llama/llama-3.1-70b-instruct

./mcp-cli query "Generate a sorting algorithm" \
  --provider openrouter \
  --model anthropic/claude-3.5-sonnet
```

---

## Cost & Limits

### Pricing

**Pay upstream cost + small markup**:
- Claude 3.5 Sonnet: $3.00-$3.50 input / $15.00-$17.50 output (per 1M tokens)
- GPT-4o: $2.50-$3.00 input / $10.00-$12.00 output
- Llama 3.1 70B: $0.50-$0.80 input/output

**Free Models**: $0 (rate limited)

💡 **Pricing**: https://openrouter.ai/models (shows current prices)

### Rate Limits

- **Free tier**: 20 requests/min (free models)
- **Paid**: Varies by model
- **Credits**: Pre-pay, no monthly minimums

### Cost Comparison

**OpenRouter vs Direct**:
- Direct API: $3.00/M (Claude Sonnet input)
- OpenRouter: $3.50/M (10-15% markup)

**Trade-off**: Pay slightly more for convenience and flexibility.

---

## Troubleshooting

### Issue: Model Not Available

**Symptoms**:
```
Error: Model temporarily unavailable
```

**Cause**: Upstream provider issues

**Solution**:
```bash
# Try fallback model
./mcp-cli query "..." --provider openrouter --model anthropic/claude-3.5-haiku

# Check status: https://openrouter.ai/models
```

### Issue: Insufficient Credits

**Symptoms**:
```
Error: Insufficient credits
```

**Solution**:
1. Go to https://openrouter.ai/credits
2. Add credits
3. Or use free models

### Issue: Rate Limit (Free Models)

**Symptoms**:
```
Error: Rate limit exceeded
```

**Solution**:
- Wait for rate limit reset
- Add credits to use paid models
- Free models have strict limits

---

## Related Resources

- **OpenRouter**: https://openrouter.ai/
- **Models**: https://openrouter.ai/models
- **Pricing**: https://openrouter.ai/models (live pricing)
- **API Docs**: https://openrouter.ai/docs
- **Discord**: https://discord.gg/openrouter

---

## Provider Comparison

**vs Direct APIs**:
- ✅ Pros: One API for all models, easy switching, simple billing
- ❌ Cons: 10-15% markup, depends on upstream, no SLAs

**vs Individual Providers**:
- ✅ Pros: Flexibility, no vendor lock-in, easy comparison
- ❌ Cons: Slightly more expensive, extra abstraction layer

**Best for**: 
- Multi-model experimentation
- Avoiding vendor lock-in
- Simplified billing
- Backup/fallback options

---

**Last Updated**: December 28, 2024  
**Verified Against**: mcp-cli v1.0.0  
**Code Reference**: `internal/providers/ai/clients/openai_compatible.go`

**Note**: OpenRouter acts as a proxy to multiple providers - perfect for flexibility!
