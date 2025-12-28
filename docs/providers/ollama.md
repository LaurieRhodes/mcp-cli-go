# Ollama

> **Skill Level**: 🟢 Beginner-Friendly  
> **Interface**: OpenAI-Compatible  
> **Best For**: Local development, privacy-focused use, no API costs

## Quick Start

**For Beginners**: Run AI models locally in 3 steps.

```bash
# 1. Install Ollama
# macOS/Linux: https://ollama.ai/download
curl -fsSL https://ollama.com/install.sh | sh

# 2. Pull a model
ollama pull qwen2.5:32b

# 3. Use with mcp-cli
./mcp-cli query "Hello, how are you?" --provider ollama
```

**No API key needed!** ✅

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

### What is Ollama?

Ollama is a local AI model runner that lets you run large language models on your own machine. Think of it as "Docker for AI models" - you pull models and run them locally without needing API keys or internet connectivity (after download).

### When to Use

- ✅ **Use when**:
  
  - You want **zero API costs**
  - You need **offline/air-gapped** environments
  - You have **privacy/data sensitivity** concerns
  - You're **learning/experimenting** with AI
  - You want **full control** over model versions

- ❌ **Avoid when**:
  
  - You need the **absolute latest/most capable** models (GPT-4, Claude Opus)
  - Your hardware is **limited** (need 8GB+ RAM for good models)
  - You want **zero setup** (cloud APIs are simpler)

- 🤔 **Consider alternatives**:
  
  - **OpenAI**: More capable models, pay-per-use
  - **Anthropic**: Best-in-class reasoning (Claude)
  - **LM Studio**: Alternative local UI

### Key Features

- **Chat/Completion**: ✅ Fully Supported
- **Streaming**: ✅ Fully Supported
- **Tool Calling (MCP)**: ✅ Fully Supported (via OpenAI-compatible API)
- **Embeddings**: ✅ Fully Supported
- **Vision**: ✅ Supported (with vision models like llava)

---

## Prerequisites

### Required

- [ ] **Ollama installed** on your system ([Download](https://ollama.com/download))
- [ ] **8GB+ RAM** recommended (16GB+ for larger models)
- [ ] **Disk space**: 4-10GB per model

### Optional

- [ ] **GPU** (NVIDIA/AMD/Metal): Dramatically faster inference
- [ ] **NVMe SSD**: Faster model loading

### System Requirements

| Component | Minimum         | Recommended                  |
| --------- | --------------- | ---------------------------- |
| RAM       | 8GB             | 16GB+                        |
| Disk      | 20GB free       | 50GB+ free                   |
| CPU       | 4 cores         | 8+ cores                     |
| GPU       | None (CPU only) | NVIDIA RTX/AMD/Apple Silicon |

---

## Setup Guide

### Step 1: Install Ollama

**For Beginners**: Ollama is a program that runs on your computer.

**macOS**:

```bash
# Download from https://ollama.com/download
# Or use Homebrew
brew install ollama
```

**Linux**:

```bash
curl -fsSL https://ollama.com/install.sh | sh
```

**Windows**:

- Download installer from https://ollama.com/download

**What this does**: Installs the Ollama service that will run models locally.

### Step 2: Start Ollama Service

```bash
# The service usually starts automatically
# Check if it's running:
ollama list

# If not running, start it:
ollama serve
```

### Step 3: Pull a Model

**For Beginners**: "Pulling" downloads a model to your computer.

```bash
# Recommended starter model (32B parameters, very capable)
ollama pull qwen2.5:32b

# Or a smaller, faster model (7B parameters)
ollama pull qwen2.5:7b

# Or llama3 (Meta's model)
ollama pull llama3.1:8b
```

**What this does**: Downloads the AI model files to your computer (one-time download).

### Step 4: Test It

```bash
# Direct Ollama test
ollama run qwen2.5:32b "What is the capital of France?"

# Or via mcp-cli
./mcp-cli query "What is the capital of France?" --provider ollama
```

### Step 5: Configure mcp-cli (Optional)

**For Beginners**: This lets you set your preferred model.

```bash
# Initialize configuration
./mcp-cli init
# Select "Ollama" when prompted
```

Or manually edit `config/providers/ollama.yaml`:

```yaml
interface_type: openai_compatible
provider_name: ollama
config:
  api_endpoint: http://localhost:11434
  default_model: qwen2.5:32b  # Change to your preferred model
  timeout_seconds: 300
  max_retries: 5
```

---

## Configuration Reference

### Provider Configuration

**File**: `config/providers/ollama.yaml`

```yaml
interface_type: openai_compatible
provider_name: ollama
config:
  api_endpoint: http://localhost:11434  # Ollama default endpoint
  default_model: qwen2.5:32b            # Your default model
  timeout_seconds: 300                   # 5 minutes for large responses
  max_retries: 5                         # Retry on failures
```

#### Configuration Options

| Option            | Type   | Required | Default                  | Description                         |
| ----------------- | ------ | -------- | ------------------------ | ----------------------------------- |
| `api_endpoint`    | string | Yes      | `http://localhost:11434` | Ollama server URL                   |
| `default_model`   | string | Yes      | `qwen2.5:32b`            | Default model to use                |
| `timeout_seconds` | int    | No       | 300                      | Request timeout                     |
| `max_retries`     | int    | No       | 5                        | Number of retries on failure        |
| `context_window`  | int    | No       | Auto                     | Max context tokens (model-specific) |
| `reserve_tokens`  | int    | No       | 2000                     | Tokens reserved for response        |

#### Environment Variables

**None required!** Ollama runs locally without authentication.

If running Ollama on a different machine:

```bash
export OLLAMA_HOST=http://192.168.1.100:11434
```

### Embedding Configuration

**File**: `config/embeddings/ollama.yaml`

```yaml
interface_type: openai_compatible
provider_name: ollama
config:
  api_endpoint: http://localhost:11434
  default_model: nomic-embed-text
  embedding_models:
    nomic-embed-text:
      description: High-performance open embedding model
      dimensions: 768
      max_tokens: 8192
    mxbai-embed-large:
      description: Large multilingual embedding model
      dimensions: 1024
      max_tokens: 512
```

**Popular embedding models**:

```bash
# Pull embedding models
ollama pull nomic-embed-text      # Recommended
ollama pull mxbai-embed-large     # Multilingual
```

---

## Implementation Details

### Interface Type

**Type**: `openai_compatible`

**What this means**:

- **For Beginners**: Ollama speaks the same language as OpenAI's API, so it works with many tools
- **For Developers**: Uses OpenAI's `/v1/chat/completions` endpoint format, enabling drop-in compatibility

### How It Works

```
User → mcp-cli → OpenAI-Compatible Client → Ollama Server (localhost:11434) → Local Model
                     ↓
              Converts to OpenAI format
```

**Architecture**:

1. mcp-cli sends requests in OpenAI's API format
2. Ollama server receives request at `http://localhost:11434/v1/chat/completions`
3. Ollama loads model into memory (first request is slower)
4. Model generates response locally on your hardware
5. Response sent back in OpenAI format

### API Endpoints

- **Chat/Completion**: `http://localhost:11434/v1/chat/completions`
- **Embeddings**: `http://localhost:11434/v1/embeddings`
- **Native API**: `http://localhost:11434/api/*` (alternative non-OpenAI endpoints)

### Authentication

**Method**: None (local access)

Ollama runs on localhost and doesn't require authentication by default. For remote access, you can set up authentication separately.

### Model Name Format

**Format**: `model-name:tag`

Examples:

- `qwen2.5:32b` - Qwen 2.5 with 32 billion parameters
- `llama3.1:8b` - Llama 3.1 with 8 billion parameters
- `mistral:latest` - Latest Mistral model

**List available models**:

```bash
ollama list
```

**Browse models**: https://ollama.com/library

---

## Features

### Chat & Completions

```bash
# Basic query
./mcp-cli query "What is the capital of France?" --provider ollama

# With system prompt
./mcp-cli query "Analyze this code" --system "You are a code reviewer" --provider ollama

# Longer conversation
./mcp-cli query "Explain quantum computing" --provider ollama --model qwen2.5:32b
```

### Streaming

```bash
# Stream responses in real-time (like ChatGPT typing)
./mcp-cli query "Tell me a story" --provider ollama --stream
```

**Great for**:

- Long responses
- Real-time feedback
- Better UX

### Tool Calling (MCP Tools)

**Status**: ✅ Fully Supported

```bash
# Ollama supports function/tool calling via OpenAI-compatible API
./mcp-cli query "Search for recent AI news and summarize it" --provider ollama

# Tools are automatically available
./mcp-cli query "What's the weather in Paris?" --provider ollama
```

**Compatible models**: Most modern models support tool calling (qwen2.5, llama3.1, mistral, etc.)

### Embeddings

**Status**: ✅ Fully Supported

```bash
# Pull embedding model first
ollama pull nomic-embed-text

# Generate embeddings
./mcp-cli embed "Your text here" --provider ollama --model nomic-embed-text

# Use in RAG applications
./mcp-cli embed "Document chunk 1" --provider ollama
```

**Popular embedding models**:

- `nomic-embed-text`: 768 dims, 8K tokens, excellent performance
- `mxbai-embed-large`: 1024 dims, multilingual
- `all-minilm`: 384 dims, very fast

### Vision

**Status**: ✅ Supported (with vision models)

```bash
# Pull a vision model
ollama pull llava:13b

# Use with images
./mcp-cli query "Describe this image" --provider ollama --model llava:13b --image path/to/image.jpg
```

---

## Usage Examples

### Example 1: Local Development

**Scenario**: Testing AI features without API costs

```bash
# Quick test
./mcp-cli query "Generate 5 creative app names for a fitness tracker" --provider ollama

# No API charges!
```

### Example 2: Privacy-Sensitive Data

**Scenario**: Analyzing confidential documents

```bash
# Your data never leaves your machine
./mcp-cli query "Summarize this confidential report" --provider ollama --system "You are a business analyst"
```

### Example 3: Offline Work

**Scenario**: Working in an air-gapped environment

```bash
# Pull models while online
ollama pull qwen2.5:32b

# Use offline later
./mcp-cli query "Help me debug this code" --provider ollama
```

### Example 4: Model Comparison

**Scenario**: Testing different models

```bash
# Try different models
./mcp-cli query "Explain recursion" --provider ollama --model qwen2.5:7b
./mcp-cli query "Explain recursion" --provider ollama --model llama3.1:8b
./mcp-cli query "Explain recursion" --provider ollama --model qwen2.5:32b
```

### Example 5: Embeddings for RAG

**Scenario**: Building a semantic search system

```bash
# Generate embeddings for documents
./mcp-cli embed "Chapter 1 content" --provider ollama --model nomic-embed-text > embeddings1.json
./mcp-cli embed "Chapter 2 content" --provider ollama --model nomic-embed-text > embeddings2.json
```

---

## Troubleshooting

### Common Issues

#### Issue: "Connection refused" Error

**Symptoms**:

```
Error: failed to connect to ollama: connection refused
```

**Cause**: Ollama service isn't running

**Solution**:

```bash
# Check if Ollama is running
ollama list

# If not, start it
ollama serve

# Or on macOS/Linux
brew services start ollama  # macOS
systemctl start ollama      # Linux
```

#### Issue: Slow First Response

**Symptoms**: First query takes 30+ seconds

**Cause**: Ollama loads model into memory on first use

**Solution**: This is normal! Subsequent requests will be much faster. To keep model loaded:

```bash
# Keep model in memory
ollama run qwen2.5:32b
# Ctrl+D to exit but keep loaded
```

#### Issue: Out of Memory

**Symptoms**:

```
Error: failed to load model: insufficient memory
```

**Cause**: Model too large for your RAM

**Solution**:

```bash
# Use a smaller model
ollama pull qwen2.5:7b     # Instead of 32b
ollama pull llama3.1:8b    # Instead of larger models

# Or upgrade RAM
```

#### Issue: Model Not Found

**Symptoms**:

```
Error: model 'qwen2.5:32b' not found
```

**Cause**: Model not pulled yet

**Solution**:

```bash
# Pull the model first
ollama pull qwen2.5:32b

# List available models
ollama list
```

#### Issue: Slow Inference

**Symptoms**: Responses take 10+ seconds per sentence

**Cause**: Running on CPU without GPU acceleration

**Solution**:

- **Get a GPU**: NVIDIA RTX, AMD, or Apple Silicon
- **Use smaller models**: 7B instead of 32B
- **Reduce context**: Shorter prompts
- **Quantization**: Use `q4` quantized models (smaller, faster)

```bash
# Quantized models (faster, slightly less accurate)
ollama pull qwen2.5:7b-q4
```

#### Issue: Wrong Endpoint

**Symptoms**:

```
Error: failed to connect to localhost:11434
```

**Cause**: Ollama running on different port

**Solution**:

```bash
# Check Ollama settings
echo $OLLAMA_HOST

# Update config
# Edit config/providers/ollama.yaml
api_endpoint: http://localhost:YOUR_PORT
```

---

## Cost & Limits

### Pricing

**Cost**: **FREE** ✅

- No API charges
- No token limits
- Unlimited requests

**Your costs**:

- Electricity to run your computer
- Initial time to download models
- Hardware (if upgrading)

### Rate Limits

**None!** It's your local machine.

- Run as many queries as your hardware can handle
- No daily/monthly quotas
- No API keys to manage

### Context Windows

| Model            | Context Window | Speed  | RAM Needed |
| ---------------- | -------------- | ------ | ---------- |
| `qwen2.5:7b`     | 32K tokens     | Fast   | 8GB        |
| `qwen2.5:32b`    | 32K tokens     | Medium | 16GB+      |
| `llama3.1:8b`    | 128K tokens    | Fast   | 8GB        |
| `llama3.1:70b`   | 128K tokens    | Slow   | 64GB+      |
| `mistral:latest` | 32K tokens     | Fast   | 8GB        |

**Note**: Larger context = more RAM needed

---

## Related Resources

- **Ollama Official Site**: https://ollama.com
- **Model Library**: https://ollama.com/library
- **GitHub**: https://github.com/ollama/ollama
- **Discord Community**: https://discord.gg/ollama
- **Documentation**: https://github.com/ollama/ollama/tree/main/docs

### Recommended Models

**Best for Coding**:

- `qwen2.5-coder:32b` - Excellent code generation
- `deepseek-coder:33b` - Strong coding assistant

**Best for Chat**:

- `qwen2.5:32b` - All-around excellent
- `llama3.1:8b` - Fast, good quality
- `mistral:latest` - Fast, efficient

**Best for Embeddings**:

- `nomic-embed-text` - Best quality/speed balance
- `mxbai-embed-large` - Multilingual support

---

## Provider Comparison

**vs OpenAI**:

- ✅ Pros: Free, private, offline, no limits
- ❌ Cons: Needs hardware, setup required, not quite as capable as GPT-4

**vs Anthropic**:

- ✅ Pros: Free, private, no rate limits
- ❌ Cons: Local hardware needed, Claude Opus is superior for complex reasoning

**vs Cloud Providers**:

- ✅ Pros: Zero ongoing costs, full control, no API dependencies
- ❌ Cons: Setup complexity, hardware investment, model updates manual

---

**Last Updated**: December 28, 2024  
**Verified Against**: mcp-cli v1.0.0, Ollama v0.3.x  
**Code Reference**: `internal/providers/ai/clients/openai_compatible.go`
