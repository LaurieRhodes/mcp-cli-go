# Azure AI Foundry

> **Skill Level**: 🟡 Intermediate (Azure knowledge helpful)  
> **Interface**: OpenAI-Compatible  
> **Best For**: Azure users, enterprise Azure integration, OpenAI on Azure

## Quick Start

```bash
# 1. Create Azure AI Foundry resource
# Go to: https://portal.azure.com/

# 2. Get endpoint and key
export AZURE_FOUNDRY_API_KEY=...
export AZURE_FOUNDRY_ENDPOINT=https://your-resource.openai.azure.com/openai/v1/

# 3. Use it
./mcp-cli query "Hello from Azure!" --provider azure-foundry
```

---

## Overview

### What is Azure AI Foundry?

Azure AI Foundry (formerly Azure OpenAI Service) provides access to OpenAI models (GPT-4, GPT-4o, embeddings) through Microsoft Azure with enterprise security, compliance, and regional deployment options.

### When to Use

- ✅ **Use when**:
  - You're on **Microsoft Azure** infrastructure
  - You need **enterprise compliance** (SOC, HIPAA, GDPR)
  - You want **data residency** in Azure regions
  - You need **Azure AD integration**
  - You require **private networking** (VNet)
  - You need **content filtering** built-in

- ❌ **Avoid when**:
  - You want **simplest setup** (use OpenAI directly)
  - You're **not on Azure** (more complex)
  - You need **latest models immediately** (Azure lags behind OpenAI)
  - You don't need enterprise features

### Key Features

- **Chat/Completion**: ✅ GPT-4o, GPT-4, GPT-3.5
- **Streaming**: ✅ Fully Supported
- **Tool Calling (MCP)**: ✅ Supported
- **Embeddings**: ✅ OpenAI embeddings
- **Vision**: ✅ GPT-4o vision
- **Enterprise**: ✅ VNet, Azure AD, compliance

---

## Prerequisites

- [ ] **Azure Subscription** ([Sign up](https://azure.microsoft.com/))
- [ ] **Azure AI Foundry Resource** created
- [ ] **Model Deployment** (GPT-4o, etc.)
- [ ] **API Key** from resource

---

## Setup Guide

### Step 1: Create Azure AI Foundry Resource

1. Go to [Azure Portal](https://portal.azure.com/)
2. Search "Azure OpenAI" or "AI Foundry"
3. Click **Create**
4. Fill details:
   - Resource group: Create or select
   - Region: Choose closest
   - Name: `my-ai-foundry`
   - Pricing tier: Standard
5. Click **Review + create**

### Step 2: Deploy Models

**For Beginners**: You must "deploy" models to use them.

1. Open your AI Foundry resource
2. Go to **Model deployments**
3. Click **Create new deployment**
4. Select model: `gpt-4o`
5. Deployment name: `gpt-4o`
6. Click **Create**

Repeat for other models you need.

### Step 3: Get API Key and Endpoint

1. In your resource, go to **Keys and Endpoint**
2. Copy:
   - **KEY 1** (API key)
   - **Endpoint** URL

### Step 4: Set Environment Variables

```bash
export AZURE_FOUNDRY_API_KEY=abc123...
export AZURE_FOUNDRY_ENDPOINT=https://my-ai-foundry.openai.azure.com/openai/v1/
```

### Step 5: Test

```bash
./mcp-cli init
# Select "Azure Foundry"

./mcp-cli query "What is the capital of France?" --provider azure-foundry
```

---

## Configuration Reference

**File**: `config/providers/azure-foundry.yaml`

```yaml
interface_type: openai_compatible
provider_name: azure-foundry
config:
  api_key: ${AZURE_FOUNDRY_API_KEY}
  api_endpoint: ${AZURE_FOUNDRY_ENDPOINT}
  default_model: gpt-4o
  timeout_seconds: 60
  max_retries: 3
  context_window: 128000
  reserve_tokens: 4000
```

### Configuration Options

| Option | Type | Required | Default | Description |
|--------|------|----------|---------|-------------|
| `api_key` | string | Yes | - | Azure AI Foundry API key |
| `api_endpoint` | string | Yes | - | Your resource endpoint |
| `default_model` | string | Yes | `gpt-4o` | Deployment name (not model name!) |

⚠️ **Important**: `default_model` is your **deployment name**, not the model ID!

### Available Models

**Depends on your deployments**:
- `gpt-4o` - Latest GPT-4 Optimized
- `gpt-4o-mini` - Fast, affordable
- `gpt-4-turbo` - GPT-4 Turbo
- `gpt-35-turbo` - GPT-3.5 Turbo
- `text-embedding-3-small` - Embeddings
- `text-embedding-3-large` - Embeddings

**Note**: You must deploy each model you want to use!

---

## Implementation Details

**Interface**: OpenAI-Compatible  
**Authentication**: API Key header  
**Endpoint**: `https://{resource-name}.openai.azure.com/openai/v1/`

**Difference from OpenAI**: 
- Model name = deployment name (you choose)
- Must deploy models before use
- API version in URL

---

## Features

### OpenAI Models on Azure

```bash
# Use GPT-4o (if deployed)
./mcp-cli query "Explain microservices" --provider azure-foundry --model gpt-4o

# Use GPT-4o-mini (if deployed)
./mcp-cli query "Simple query" --provider azure-foundry --model gpt-4o-mini
```

### Enterprise Security

- **VNet Integration**: Private endpoints
- **Azure AD**: Identity-based access
- **Managed Identity**: No API keys needed
- **Content Filtering**: Built-in safety
- **Compliance**: SOC, HIPAA, GDPR certified
- **Data Residency**: Choose Azure region

### Embeddings

```bash
# Generate embeddings (if deployment exists)
./mcp-cli embed "Your text" --provider azure-foundry --model text-embedding-3-small
```

---

## Usage Examples

### Example 1: Enterprise Chat

```bash
# All data stays in Azure (compliance)
./mcp-cli query "Analyze this sensitive data" \
  --provider azure-foundry \
  --model gpt-4o
```

### Example 2: Regional Deployment

```bash
# Use European deployment (data residency)
export AZURE_FOUNDRY_ENDPOINT=https://eu-ai-foundry.openai.azure.com/openai/v1/

./mcp-cli query "EU data processing" --provider azure-foundry
```

---

## Cost & Limits

### Pricing (Dec 2024)

**Same as OpenAI, billed through Azure**:

- GPT-4o: $2.50/M input, $10.00/M output
- GPT-4o-mini: $0.150/M input, $0.600/M output
- GPT-4 Turbo: $10.00/M input, $30.00/M output

**Plus Azure fees** (minimal)

### Quota Management

**Per deployment**:
- Tokens per minute (TPM)
- Requests per minute (RPM)

**Increase quotas**: Submit request in Azure Portal

---

## Troubleshooting

### Issue: Deployment Not Found

**Symptoms**:
```
Error: Deployment 'gpt-4o' not found
```

**Cause**: Model not deployed

**Solution**:
1. Go to Azure Portal
2. Open AI Foundry resource
3. Model deployments → Create deployment
4. Use deployment name in config

### Issue: Quota Exceeded

**Symptoms**:
```
Error: 429 Rate limit exceeded
```

**Cause**: Exceeded deployment quota

**Solution**:
- Wait for quota reset
- Request quota increase
- Create additional deployments

### Issue: Endpoint Invalid

**Symptoms**:
```
Error: 404 Not Found
```

**Cause**: Wrong endpoint URL

**Solution**:
```bash
# Verify endpoint format:
# ✅ https://your-resource.openai.azure.com/openai/v1/
# ❌ https://your-resource.openai.azure.com/ (missing /openai/v1/)
```

---

## Related Resources

- **Azure Portal**: https://portal.azure.com/
- **Documentation**: https://learn.microsoft.com/azure/ai-services/openai/
- **Pricing**: https://azure.microsoft.com/pricing/details/cognitive-services/openai-service/
- **Quotas**: https://learn.microsoft.com/azure/ai-services/openai/quotas-limits

---

## Provider Comparison

**vs OpenAI Direct**:
- ✅ Pros: Azure integration, compliance, VNet, content filtering
- ❌ Cons: More complex setup, may lag on new models

**vs GCP Vertex AI**:
- ✅ Pros: OpenAI models (GPT-4o)
- ❌ Cons: Only OpenAI (Vertex has Gemini), Azure lock-in

**Best for**: Enterprise Azure users needing OpenAI models with compliance

---

**Last Updated**: December 28, 2024  
**Verified Against**: mcp-cli v1.0.0  
**Code Reference**: `internal/providers/ai/clients/openai_compatible.go`

**Note**: Use deployment names, not model IDs! Each model must be deployed first.
