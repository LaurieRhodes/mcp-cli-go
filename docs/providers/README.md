# Provider Documentation

Comprehensive guides for each AI provider supported by mcp-cli.

## Documentation Structure

Each provider guide includes:

- 🟢 **Quick Start**: Get running in minutes
- 📋 **Setup Guide**: Step-by-step instructions
- ⚙️ **Configuration Reference**: All options explained  
- 🔧 **Implementation Details**: Technical architecture
- 💡 **Usage Examples**: Real-world scenarios
- 🐛 **Troubleshooting**: Common issues and solutions
- 💰 **Cost & Limits**: Pricing and quotas

---

## Available Providers

### Local & Free

| Provider                     | Interface         | Skill Level | Best For                  | Tool Calling | Embeddings |
| ---------------------------- | ----------------- | ----------- | ------------------------- | ------------ | ---------- |
| **[Ollama](ollama.md)**      | OpenAI-Compatible | 🟢 Beginner | Local, offline, zero cost | ✅ Yes        | ✅ Yes      |
| **[LM Studio](lmstudio.md)** | OpenAI-Compatible | 🟢 Beginner | Local with GUI            | ✅ Yes        | ✅ Yes      |

### Cloud - OpenAI Compatible

| Provider                              | Interface         | Skill Level     | Best For               | Tool Calling | Embeddings |
| ------------------------------------- | ----------------- | --------------- | ---------------------- | ------------ | ---------- |
| **[OpenAI](openai.md)**               | OpenAI-Compatible | 🟢 Beginner     | GPT-4, latest models   | ✅ Yes        | ✅ Yes      |
| **[DeepSeek](deepseek.md)**           | OpenAI-Compatible | 🟢 Beginner     | Cost-effective, coding | ✅ Yes        | ❌ No       |
| **[OpenRouter](openrouter.md)**       | OpenAI-Compatible | 🟢 Beginner     | Multi-model access     | ✅ Yes        | ✅ Yes      |
| **[Azure Foundry](azure-foundry.md)** | OpenAI-Compatible | 🟡 Intermediate | Azure integration      | ✅ Yes        | ✅ Yes      |

### Cloud - Native Interfaces

| Provider                      | Interface        | Skill Level | Best For              | Tool Calling | Embeddings |
| ----------------------------- | ---------------- | ----------- | --------------------- | ------------ | ---------- |
| **[Anthropic](anthropic.md)** | Anthropic Native | 🟢 Beginner | Claude, reasoning     | ✅ Yes        | ❌ No       |
| **[Gemini](gemini.md)**       | Gemini Native    | 🟢 Beginner | Google AI, multimodal | ✅ Yes        | ✅ Yes      |

### Cloud - Enterprise

| Provider                          | Interface     | Skill Level     | Best For           | Tool Calling | Embeddings |
| --------------------------------- | ------------- | --------------- | ------------------ | ------------ | ---------- |
| **[GCP Vertex AI](vertex-ai.md)** | Custom Hybrid | 🟡 Intermediate | Gemini on GCP      | ✅ Yes        | ✅ Yes      |
| **[AWS Bedrock](bedrock.md)**     | AWS Custom    | 🟡 Intermediate | Multi-model on AWS | ✅ Yes        | ✅ Yes      |

---

## Quick Comparison

### By Use Case

**Best for Beginners**:

1. Ollama - Free, local, no API key
2. OpenAI - Simple setup, great docs
3. Anthropic - Claude is excellent

**Best for Cost**:

1. Ollama - Free (hardware cost only)
2. DeepSeek - Very affordable
3. Gemini - Competitive pricing

**Best for Privacy**:

1. Ollama - Fully local
2. LM Studio - Fully local
3. AWS Bedrock - Enterprise security

**Best for Enterprise**:

1. AWS Bedrock - Multi-model, AWS integration
2. GCP Vertex AI - GCP integration, Gemini
3. Azure Foundry - Azure integration, OpenAI

**Best for Coding**:

1. Claude (Anthropic) - Excellent at code
2. GPT-4 (OpenAI) - Very capable
3. DeepSeek - Cost-effective coding

**Best for Tool Calling**:

1. Claude (Anthropic) - Best tool use
2. GPT-4 (OpenAI) - Excellent
3. Gemini (Vertex AI/Public) - Very good

---

## Interface Types Explained

### OpenAI-Compatible

**Providers**: Ollama, OpenAI, DeepSeek, OpenRouter, LM Studio, Azure Foundry

**What it means**: 

- Uses OpenAI's `/v1/chat/completions` API format
- Drop-in compatible with OpenAI tools
- Easy to switch between providers
- Consistent tool calling format

**Code**: `internal/providers/ai/clients/openai_compatible.go`

### Native Interfaces

**Providers**: Anthropic, Gemini

**What it means**:

- Uses provider's own API format
- Provider-specific features available
- May have unique capabilities
- Requires provider-specific code

**Code**: 

- Anthropic: `internal/providers/ai/clients/anthropic.go`
- Gemini: `internal/providers/ai/clients/gemini.go`

### Custom/Hybrid

**Providers**: AWS Bedrock, GCP Vertex AI

**What it means**:

- Custom authentication (IAM, OAuth2)
- May combine multiple endpoint types
- Enterprise features (VPC, CMEK, etc.)
- Provider-specific request/response handling

**Code**:

- Bedrock: `internal/providers/ai/clients/aws_bedrock.go`
- Vertex AI: `internal/providers/ai/clients/gcp_vertex_ai_openai.go`

---

## Feature Matrix

| Feature          | Ollama | OpenAI  | Anthropic | Gemini  | Vertex AI | Bedrock | OpenRouter | DeepSeek | Azure   | LM Studio |
| ---------------- | ------ | ------- | --------- | ------- | --------- | ------- | ---------- | -------- | ------- | --------- |
| **Chat**         | ✅      | ✅       | ✅         | ✅       | ✅         | ✅       | ✅          | ✅        | ✅       | ✅         |
| **Streaming**    | ✅      | ✅       | ✅         | ✅       | ✅         | ✅       | ✅          | ✅        | ✅       | ✅         |
| **Tools**        | ✅      | ✅       | ✅         | ✅       | ✅         | ✅       | ✅          | ✅        | ✅       | ✅         |
| **Embeddings**   | ✅      | ✅       | ❌         | ✅       | ✅         | ✅       | ✅          | ❌        | ✅       | ✅         |
| **Vision**       | ✅*     | ✅       | ✅         | ✅       | ✅         | ✅       | ✅          | ❌        | ✅       | ✅*        |
| **Long Context** | ✅      | ✅       | ✅         | ✅✅      | ✅✅        | ✅       | ✅          | ✅        | ✅       | ✅         |
| **Offline**      | ✅      | ❌       | ❌         | ❌       | ❌         | ❌       | ❌          | ❌        | ❌       | ✅         |
| **Free Tier**    | ✅      | Limited | Limited   | Limited | Limited   | Limited |            | Limited  | Limited | ✅         |

*With vision-capable models

---

## Getting Started

### 1. Choose Your Provider

Consider:

- **Budget**: Ollama (free) vs OpenAI (pay-per-use)
- **Privacy**: Local (Ollama) vs Cloud (others)
- **Features**: Tool calling, embeddings, vision
- **Existing infrastructure**: AWS, GCP, Azure

### 2. Follow Setup Guide

Each provider doc includes complete setup instructions for all skill levels.

### 3. Initialize Configuration

```bash
# Interactive setup
./mcp-cli init

# Select your providers when prompted
```

### 4. Test Connection

```bash
./mcp-cli query "Hello world" --provider your-provider
```

---

## Provider Status

| Provider      | Status   | Last Verified | Notes                 |
| ------------- | -------- | ------------- | --------------------- |
| Ollama        | ✅ Stable | 2024-12-28    | Fully tested          |
| OpenAI        | ✅ Stable | 2024-12-28    | Production ready      |
| Anthropic     | ✅ Stable | 2024-12-28    | Claude 3.5 tested     |
| Gemini        | ✅ Stable | 2024-12-28    | Public API            |
| Vertex AI     | ✅ Stable | 2024-12-28    | Hybrid implementation |
| Bedrock       | ✅ Stable | 2024-12-27    | Multi-model support   |
| Azure Foundry | ✅ Stable | 2024-12-27    | OpenAI on Azure       |
| OpenRouter    | ✅ Stable | 2024-12-28    | Proxy to many models  |
| DeepSeek      | ✅ Stable | 2024-12-28    | Chat only             |
| LM Studio     | ✅ Stable | 2024-12-28    | Local server          |

---

## Support & Feedback

- **Issues**: [GitHub Issues](https://github.com/LaurieRhodes/mcp-cli-go/issues)
- **Discussions**: [GitHub Discussions](https://github.com/LaurieRhodes/mcp-cli-go/discussions)
- **Contributing**: See `CONTRIBUTING.md`

---

## Updates

This documentation is actively maintained. Each provider doc includes:

- Last updated date
- Verified version
- Code references for verification

To verify information, check the referenced code files in the repository.
