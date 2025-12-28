# LM Studio

> **Skill Level**: 🟢 Beginner-Friendly  
> **Interface**: OpenAI-Compatible  
> **Best For**: Local models with GUI, beginners who want visual interface

## Quick Start

```bash
# 1. Install LM Studio
# Download from: https://lmstudio.ai/

# 2. Download a model in LM Studio GUI
# Click "Search" → Download model (e.g., Qwen 2.5)

# 3. Start local server in LM Studio
# Click "Local Server" tab → Start Server

# 4. Use with mcp-cli
./mcp-cli query "Hello from LM Studio!" --provider lmstudio
```

**No API key needed!** ✅

---

## Overview

### What is LM Studio?

LM Studio is a desktop application for running large language models locally on your computer. It provides a user-friendly GUI for downloading models, chatting, and serving models via an OpenAI-compatible API.

### When to Use

- ✅ **Use when**:
  - You want **local AI** with **GUI** (easier than Ollama)
  - You're a **beginner** who prefers visual interfaces
  - You want **zero API costs**
  - You need **offline capability**
  - You want to **compare models** easily
  - You need **privacy** (data stays local)

- ❌ **Avoid when**:
  - You prefer **command-line** only (use Ollama)
  - You want **server deployment** (Ollama is better for this)
  - You need **absolute latest models** (cloud APIs update faster)

### Key Features

- **Chat/Completion**: ✅ Any GGUF model
- **Streaming**: ✅ Fully Supported
- **Tool Calling (MCP)**: ✅ Model-dependent
- **Embeddings**: ✅ Supported
- **Vision**: ✅ With vision-capable models
- **GUI**: ✅ Easy model management
- **Offline**: ✅ Fully offline after download

---

## Prerequisites

- [ ] **LM Studio installed** ([Download](https://lmstudio.ai/))
- [ ] **8GB+ RAM** (16GB+ recommended)
- [ ] **Disk space** for models (4-20GB per model)

### System Requirements

| Component | Minimum | Recommended |
|-----------|---------|-------------|
| RAM | 8GB | 16GB+ |
| Disk | 20GB free | 100GB+ |
| CPU | 4 cores | 8+ cores |
| GPU | None (CPU works) | NVIDIA/AMD/Apple Silicon |

---

## Setup Guide

### Step 1: Install LM Studio

**Windows/Mac/Linux**:
1. Go to https://lmstudio.ai/
2. Download for your OS
3. Install application
4. Launch LM Studio

### Step 2: Download a Model

**For Beginners**: "Downloading" means copying the AI to your computer.

1. Click **Search** icon (🔍)
2. Search for model (try "Qwen 2.5")
3. Select a model:
   - **7B** - Fast, good for most tasks
   - **32B** - Slower, excellent quality
4. Click **Download**
5. Wait for download (can be large!)

**Recommended starter models**:
- `Qwen/Qwen2.5-7B-Instruct-GGUF`
- `TheBloke/Llama-3.1-8B-Instruct-GGUF`
- `TheBloke/Mistral-7B-Instruct-v0.2-GGUF`

### Step 3: Load Model

1. Click **Chat** icon (💬)
2. Click **Select a model to load**
3. Choose downloaded model
4. Wait for loading (first time is slower)

### Step 4: Start Local Server

**For Beginners**: This makes the model available to mcp-cli.

1. Click **Local Server** tab (🌐)
2. Click **Start Server**
3. Server starts on `http://localhost:1234`
4. Note: Model must be loaded in Chat tab first!

### Step 5: Test with mcp-cli

```bash
# Initialize configuration
./mcp-cli init
# Select "LM Studio"

# Test query
./mcp-cli query "What is the capital of France?" --provider lmstudio

# Expected: "The capital of France is Paris."
```

---

## Configuration Reference

**File**: `config/providers/lmstudio.yaml`

```yaml
interface_type: openai_compatible
provider_name: lmstudio
config:
  api_endpoint: http://localhost:1234/v1
  default_model: local-model
  timeout_seconds: 300
  max_retries: 2
```

### Configuration Options

| Option | Type | Required | Default | Description |
|--------|------|----------|---------|-------------|
| `api_endpoint` | string | Yes | `http://localhost:1234/v1` | LM Studio server URL |
| `default_model` | string | Yes | `local-model` | Model identifier |
| `timeout_seconds` | int | No | 300 | Request timeout |

**Note**: `default_model` can be any string - LM Studio uses whichever model is loaded.

---

## Implementation Details

**Interface**: OpenAI-Compatible  
**Authentication**: None (local access)  
**Endpoint**: `http://localhost:1234/v1/chat/completions`

**How it works**:
1. LM Studio loads model into memory
2. Local server exposes OpenAI-compatible API
3. mcp-cli connects to localhost
4. No internet needed for inference!

---

## Features

### GUI Model Management

**Advantages over Ollama**:
- ✅ Visual model browser
- ✅ Easy model comparison
- ✅ Built-in chat interface
- ✅ Performance monitoring
- ✅ No command-line needed

**Chat Features**:
- Model comparison side-by-side
- Export conversations
- System prompt templates
- Temperature/parameter controls

### Multiple Models

```bash
# Switch models in LM Studio GUI
# 1. Stop current server
# 2. Load different model in Chat tab
# 3. Restart server
# 4. Query will use new model

./mcp-cli query "Test with new model" --provider lmstudio
```

### Embeddings

```bash
# Load embedding model in LM Studio first!
./mcp-cli embed "Your text" --provider lmstudio --model nomic-embed-text
```

**Popular embedding models**:
- `nomic-ai/nomic-embed-text-v1.5-GGUF`
- `sentence-transformers/all-MiniLM-L6-v2-GGUF`

---

## Usage Examples

### Example 1: Beginner's First Local AI

```bash
# 1. Download Qwen 2.5 7B in LM Studio
# 2. Load model in Chat tab
# 3. Start server
# 4. Use mcp-cli

./mcp-cli query "Explain Python loops simply" --provider lmstudio
```

**Why LM Studio**: GUI makes it approachable for beginners!

### Example 2: Model Comparison

**In LM Studio GUI**:
1. Chat with Model A
2. Switch to Model B (Chat → Load different model)
3. Compare responses

**With mcp-cli**:
```bash
# Load Model A in LM Studio → Start server
./mcp-cli query "Explain REST APIs" --provider lmstudio > model-a.txt

# Load Model B in LM Studio → Restart server
./mcp-cli query "Explain REST APIs" --provider lmstudio > model-b.txt

# Compare outputs
diff model-a.txt model-b.txt
```

### Example 3: Offline Development

```bash
# Download models while online
# (In LM Studio, download 3-4 models)

# Work offline later
# No internet needed!
./mcp-cli query "Help me debug this code" --provider lmstudio
```

---

## Troubleshooting

### Issue: Connection Refused

**Symptoms**:
```
Error: connection refused: localhost:1234
```

**Cause**: Server not started

**Solution**:
1. Open LM Studio
2. Go to **Local Server** tab
3. Click **Start Server**
4. Make sure a model is loaded first!

### Issue: Slow Responses

**Symptoms**: Takes 30+ seconds per response

**Cause**: Model too large for your hardware

**Solution**:
```bash
# Use smaller model
# Instead of 32B → try 7B
# Instead of 70B → try 13B or 7B

# Or use quantized models:
# q4_K_M (smaller, faster)
# q8_0 (larger, better quality)
```

### Issue: Out of Memory

**Symptoms**:
```
Error: Failed to load model - insufficient memory
```

**Cause**: Model too large for RAM

**Solution**:
- Close other applications
- Use smaller model (7B instead of 13B)
- Use lower quantization (q4 instead of q8)
- Upgrade RAM

### Issue: Wrong Model Loaded

**Symptoms**: Unexpected responses

**Cause**: Different model loaded than expected

**Solution**:
1. Check LM Studio Chat tab
2. Verify which model is loaded
3. Reload correct model
4. Restart server

---

## Cost & Limits

### Pricing

**Cost**: **FREE** ✅

- No API charges
- No subscriptions
- Unlimited usage

**Your costs**:
- Electricity
- Hardware (if upgrading)
- Time to download models

### Rate Limits

**None!** Run as many queries as your hardware can handle.

### Performance

**Typical speed** (depends on hardware):

| Hardware | Tokens/sec | Use Case |
|----------|------------|----------|
| CPU only | 2-5 | Light usage |
| Integrated GPU | 5-15 | Regular use |
| Dedicated GPU (NVIDIA) | 30-100+ | Heavy use |
| Apple Silicon (M1/M2/M3) | 20-60 | Great performance |

---

## Related Resources

- **LM Studio**: https://lmstudio.ai/
- **Model Hub**: Built into LM Studio (Search tab)
- **Discord**: LM Studio community
- **Documentation**: https://lmstudio.ai/docs

### Model Sources

LM Studio downloads from:
- **Hugging Face**: Main model repository
- **TheBloke**: Popular quantized models

### Recommended Models

**Best All-Around**:
- Qwen 2.5 (7B or 32B)
- Llama 3.1 (8B or 70B)
- Mistral (7B)

**Best for Coding**:
- DeepSeek Coder (6.7B or 33B)
- Qwen 2.5 Coder

**Best for Chat**:
- Llama 3.1 Instruct
- Qwen 2.5 Instruct

---

## Provider Comparison

**vs Ollama**:
- ✅ Pros: GUI, easier for beginners, visual model comparison
- ❌ Cons: Less suitable for servers, fewer community models

**vs Cloud APIs**:
- ✅ Pros: Free, private, offline, no rate limits
- ❌ Cons: Needs good hardware, slower than cloud, setup required

**vs OpenAI**:
- ✅ Pros: Zero cost, privacy, offline capability
- ❌ Cons: Lower quality models, needs hardware, manual updates

**Best for**: 
- Beginners who want local AI
- Visual learners
- Privacy-conscious users
- Budget-conscious developers

---

**Last Updated**: December 28, 2024  
**Verified Against**: mcp-cli v1.0.0, LM Studio 0.2.x  
**Code Reference**: `internal/providers/ai/clients/openai_compatible.go`

**Key Difference from Ollama**: LM Studio = GUI-first, Ollama = CLI-first. Both run local models!
