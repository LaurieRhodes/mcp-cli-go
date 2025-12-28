# GCP Vertex AI

> **Skill Level**: 🟡 Intermediate (GCP knowledge helpful)  
> **Interface**: Custom Hybrid (OpenAI-compatible for chat, Native for embeddings)  
> **Best For**: Enterprise GCP users, Gemini models with tool calling, regulated industries

## Quick Start

**For Beginners**: Use Google's Gemini models on Vertex AI in 4 steps.

```bash
# 1. Create GCP project and enable Vertex AI
# Go to: https://console.cloud.google.com/

# 2. Create service account and download key
gcloud iam service-accounts create vertex-ai-client
gcloud iam service-accounts keys create ~/vertex-ai-key.json \
  --iam-account=vertex-ai-client@YOUR_PROJECT.iam.gserviceaccount.com

# 3. Set environment variables
export GCP_PROJECT_ID=your-project-id
export GOOGLE_APPLICATION_CREDENTIALS=~/vertex-ai-key.json
export GCP_LOCATION=us-central1

# 4. Use with mcp-cli
./mcp-cli query "Hello from Vertex AI!" --provider vertex-ai
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

### What is GCP Vertex AI?

Vertex AI is Google Cloud's unified AI platform. It provides access to Google's Gemini models with enterprise features like VPC Service Controls, customer-managed encryption keys (CMEK), data residency guarantees, and IAM-based access control.

### When to Use

- ✅ **Use when**:
  - You're already using **Google Cloud Platform**
  - You need **enterprise security** (VPC-SC, CMEK, DLP)
  - You require **data residency** guarantees
  - You want **IAM-based access control**
  - You need **unified logging** with Cloud Logging
  - You want **MCP tool calling** with Gemini

- ❌ **Avoid when**:
  - You want **simplest setup** (use Gemini public API instead)
  - You're **not on GCP** (use native Gemini API or other providers)
  - You need **non-Google models** (use AWS Bedrock for multi-model)

- 🤔 **Consider alternatives**:
  - **Gemini API**: Simpler setup, same models, no GCP needed
  - **OpenAI**: Broader model selection
  - **AWS Bedrock**: Multi-provider (Claude, Llama, etc.)

### Key Features

- **Chat/Completion**: ✅ Fully Supported (via OpenAI-compatible endpoint)
- **Streaming**: ✅ Fully Supported
- **Tool Calling (MCP)**: ✅ Fully Supported (OpenAI format)
- **Embeddings**: ✅ Fully Supported (via native endpoint)
- **Vision**: ✅ Supported (Gemini Pro Vision)
- **Long Context**: ✅ Up to 2M tokens (Gemini 1.5 Pro)

---

## Prerequisites

### Required

- [ ] **Google Cloud Account** ([Sign up](https://cloud.google.com/))
- [ ] **GCP Project** created
- [ ] **Billing enabled** on project
- [ ] **Vertex AI API** enabled
- [ ] **Service Account** with "Vertex AI User" role
- [ ] **Service Account JSON key** downloaded

### Optional

- [ ] **gcloud CLI** installed ([Install](https://cloud.google.com/sdk/docs/install))
- [ ] **VPC Service Controls** (for enhanced security)
- [ ] **Organization policies** (for enterprise governance)

### Cost Awareness

⚠️ **Vertex AI charges for API usage**. See [Cost & Limits](#cost--limits) section.

---

## Setup Guide

### Step 1: Create GCP Project

**For Beginners**: A project is like a container for your Google Cloud resources.

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Click **project dropdown** → **NEW PROJECT**
3. Enter project name (e.g., "vertex-ai-mcp")
4. Note your **Project ID** (e.g., `gen-lang-client-0905788261`)
5. Click **Create**

**What this does**: Creates a billing container for your AI usage.

### Step 2: Enable Vertex AI API

**For Beginners**: This turns on the AI features for your project.

**Option A: Via Console (Easiest)**:
1. Go to https://console.cloud.google.com/apis/library/aiplatform.googleapis.com
2. Select your project
3. Click **ENABLE**

**Option B: Via gcloud CLI**:
```bash
gcloud services enable aiplatform.googleapis.com --project=YOUR_PROJECT_ID
```

**What this does**: Activates the Vertex AI API for your project.

### Step 3: Enable Billing

1. Go to [Billing](https://console.cloud.google.com/billing)
2. Link a billing account to your project
3. ⚠️ **Set up budget alerts** to avoid surprises

**What this does**: Required for API calls (free tier available).

### Step 4: Create Service Account

**For Beginners**: A service account is like a "robot user" that your application uses to access Google Cloud.

**Option A: Via Console**:
1. Go to **IAM & Admin** → **Service Accounts**
2. Click **CREATE SERVICE ACCOUNT**
3. Name: `vertex-ai-client`
4. Click **CREATE AND CONTINUE**
5. Role: Select **Vertex AI User** (`roles/aiplatform.user`)
6. Click **DONE**

**Option B: Via gcloud CLI**:
```bash
# Create service account
gcloud iam service-accounts create vertex-ai-client \
  --display-name="Vertex AI Client" \
  --project=YOUR_PROJECT_ID

# Grant Vertex AI User role
gcloud projects add-iam-policy-binding YOUR_PROJECT_ID \
  --member="serviceAccount:vertex-ai-client@YOUR_PROJECT_ID.iam.gserviceaccount.com" \
  --role="roles/aiplatform.user"
```

**What this does**: Creates an identity for mcp-cli to authenticate with Vertex AI.

### Step 5: Download Service Account Key

**For Beginners**: This is like a password file for your service account.

1. In **Service Accounts**, click your service account
2. Go to **KEYS** tab
3. Click **ADD KEY** → **Create new key**
4. Choose **JSON**
5. Click **CREATE**
6. Save file securely (e.g., `~/vertex-ai-key.json`)

**Or via gcloud**:
```bash
gcloud iam service-accounts keys create ~/vertex-ai-key.json \
  --iam-account=vertex-ai-client@YOUR_PROJECT_ID.iam.gserviceaccount.com
```

⚠️ **Security**: Never commit this file to git! Keep it secure.

**What this does**: Downloads credentials for authentication.

### Step 6: Set Environment Variables

**For Beginners**: These tell mcp-cli how to connect to your GCP project.

```bash
# Add to your .env file
export GCP_PROJECT_ID=your-project-id
export GCP_LOCATION=us-central1
export GOOGLE_APPLICATION_CREDENTIALS=/absolute/path/to/vertex-ai-key.json
```

Or in `.env` file:
```bash
GCP_PROJECT_ID=gen-lang-client-0905788261
GCP_LOCATION=us-central1
GOOGLE_APPLICATION_CREDENTIALS=/home/user/vertex-ai-key.json
```

**What this does**: Configures connection parameters.

### Step 7: Test Connection

```bash
# Initialize configuration
./mcp-cli init
# Select "GCP Vertex AI" when prompted

# Test query
./mcp-cli query "What is the capital of Australia?" --provider vertex-ai

# Expected output: "The capital of Australia is Canberra."
```

---

## Configuration Reference

### Provider Configuration

**File**: `config/providers/gcp-vertex-ai.yaml`

```yaml
interface_type: gcp_vertex_ai
provider_name: vertex-ai
config:
  # GCP Project ID (required)
  project_id: ${GCP_PROJECT_ID}
  
  # GCP Region (required)
  location: ${GCP_LOCATION:-us-central1}
  
  # Path to service account JSON key (required)
  credentials_path: ${GOOGLE_APPLICATION_CREDENTIALS}
  
  # Default model (required)
  default_model: gemini-2.5-flash
  
  # Request settings
  timeout_seconds: 60
  max_retries: 3
  context_window: 1000000
  reserve_tokens: 4000
  
  # Embedding models (optional)
  embedding_models:
    text-embedding-004:
      max_tokens: 3072
      dimensions: 768
      default: true
    text-multilingual-embedding-002:
      max_tokens: 3072
      dimensions: 768
    textembedding-gecko@003:
      max_tokens: 3072
      dimensions: 768
```

#### Configuration Options

| Option | Type | Required | Default | Description |
|--------|------|----------|---------|-------------|
| `project_id` | string | Yes | - | GCP Project ID |
| `location` | string | Yes | `us-central1` | GCP region for API calls |
| `credentials_path` | string | Yes | - | Path to service account JSON |
| `default_model` | string | Yes | `gemini-2.5-flash` | Default Gemini model |
| `timeout_seconds` | int | No | 60 | Request timeout |
| `max_retries` | int | No | 3 | Retry attempts |
| `context_window` | int | No | 1000000 | Max input tokens |
| `reserve_tokens` | int | No | 4000 | Reserved for output |

#### Environment Variables

| Variable | Required | Description | Example |
|----------|----------|-------------|---------|
| `GCP_PROJECT_ID` | Yes | Your GCP project ID | `gen-lang-client-0905788261` |
| `GCP_LOCATION` | Yes | GCP region | `us-central1` |
| `GOOGLE_APPLICATION_CREDENTIALS` | Yes | Path to service account JSON | `/home/user/vertex-key.json` |

#### Available Regions

Recommended regions:
- `us-central1` - Iowa, USA (most reliable, all features)
- `us-east1` - South Carolina, USA
- `europe-west2` - London, UK
- `asia-southeast1` - Singapore
- `australia-southeast1` - Sydney, Australia (closest to Melbourne)

**Note**: Some regions may have limited model availability.

### Embedding Configuration

**File**: `config/embeddings/gcp-vertex-ai.yaml`

```yaml
interface_type: gcp_vertex_ai
provider_name: vertex-ai
config:
  project_id: ${GCP_PROJECT_ID}
  location: ${GCP_LOCATION:-us-central1}
  credentials_path: ${GOOGLE_APPLICATION_CREDENTIALS}
  default_model: text-embedding-004
  timeout_seconds: 30
  max_retries: 3
  embedding_models:
    text-embedding-004:
      description: Latest Google embedding model
      dimensions: 768
      max_tokens: 3072
    text-multilingual-embedding-002:
      description: Multilingual embedding model
      dimensions: 768
      max_tokens: 3072
    textembedding-gecko@003:
      description: Gecko embedding model v3
      dimensions: 768
      max_tokens: 3072
```

---

## Implementation Details

### Interface Type

**Type**: `gcp_vertex_ai` (Custom Hybrid)

**What this means**:
- **For Beginners**: We use a special setup that combines two Google endpoints for best results
- **For Developers**: Hybrid implementation using OpenAI-compatible endpoint for chat (tool calling support) and native endpoint for embeddings

### Architecture: Hybrid Approach

```
┌─────────────────────────────────────────────────────┐
│ Chat/Completions (Tool Calling Supported)          │
│                                                      │
│ mcp-cli → OAuth2 Token → OpenAI-Compatible Endpoint │
│                          ↓                           │
│              Gemini Model (google/gemini-2.5-flash) │
└─────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────┐
│ Embeddings (Native Format)                          │
│                                                      │
│ mcp-cli → OAuth2 Token → Native Vertex AI Endpoint  │
│                          ↓                           │
│                 Gecko/Text Embedding Models         │
└─────────────────────────────────────────────────────┘
```

**Why hybrid?**
- **OpenAI endpoint** supports tool calling but NOT embeddings
- **Native endpoint** supports embeddings but tool calling uses different format
- **Solution**: Use OpenAI endpoint for chat, native for embeddings

### API Endpoints

**Chat/Completion** (OpenAI-compatible):
```
https://{location}-aiplatform.googleapis.com/v1beta1/projects/{project}/locations/{location}/endpoints/openapi
```

**Embeddings** (Native):
```
https://{location}-aiplatform.googleapis.com/v1/projects/{project}/locations/{location}/publishers/google/models/{model}:predict
```

### Authentication

**Method**: OAuth2 with JWT (Service Account)

**Flow**:
1. Load service account JSON (private key, client email)
2. Create JWT with claims (issuer, scope, audience, expiry)
3. Sign JWT with RSA private key (SHA256)
4. Exchange JWT for access token at Google's token URI
5. Use access token in `Authorization: Bearer` header
6. Token auto-refreshes with 5-minute buffer before expiry

**Scopes**: `https://www.googleapis.com/auth/cloud-platform`

### Model Name Format

**For Chat**:
- Config format: `gemini-2.5-flash`
- API format: `google/gemini-2.5-flash` (auto-converted)

**For Embeddings**:
- Config format: `text-embedding-004`
- API format: `text-embedding-004` (used as-is)

**Why different?**
- OpenAI-compatible endpoint requires `publisher/model` format
- Native endpoint uses model name directly

### Code Implementation

**File**: `internal/providers/ai/clients/gcp_vertex_ai_openai.go`

Key features:
- Wraps `OpenAICompatibleClient` for chat
- Direct native API calls for embeddings
- OAuth2 token management with auto-refresh
- Model name format conversion

---

## Features

### Chat & Completions

```bash
# Basic query
./mcp-cli query "What is the capital of France?" --provider vertex-ai

# With system prompt
./mcp-cli query "Analyze this code" \
  --system "You are a senior software architect" \
  --provider vertex-ai

# Specific model
./mcp-cli query "Explain quantum computing" \
  --provider vertex-ai \
  --model gemini-2.5-pro
```

**Available Models**:
- `gemini-2.5-flash` - Advanced with thinking (recommended)
- `gemini-2.5-pro` - Most capable, complex reasoning
- `gemini-2.0-flash-001` - Fast, efficient
- `gemini-2.0-flash-lite-001` - Ultra-efficient

### Streaming

```bash
# Stream responses in real-time
./mcp-cli query "Write a short story about AI" \
  --provider vertex-ai \
  --stream
```

**Status**: ✅ Fully Supported

### Tool Calling (MCP Tools)

**Status**: ✅ Fully Supported

```bash
# Automatic tool use
./mcp-cli query "Search for latest AI developments and summarize top 3" \
  --provider vertex-ai \
  --verbose

# Verbose shows tool calls
```

**How it works**:
1. mcp-cli advertises available tools in OpenAI format
2. Gemini decides when to call tools
3. mcp-cli executes tool (e.g., web search)
4. Results sent back to Gemini
5. Gemini synthesizes final answer

**Supported tools**:
- ✅ Web search
- ✅ File system operations
- ✅ Database queries
- ✅ Any MCP server tool

### Embeddings

**Status**: ✅ Fully Supported (via native endpoint)

```bash
# Generate embeddings
./mcp-cli embed "Your text here" \
  --provider vertex-ai \
  --model text-embedding-004

# With specific model
./mcp-cli embed "Multilingual text" \
  --provider vertex-ai \
  --model text-multilingual-embedding-002
```

**Available Models**:
- `text-embedding-004` - Latest, best performance
- `text-multilingual-embedding-002` - 100+ languages
- `textembedding-gecko@003` - Stable, proven

### Vision

**Status**: ✅ Supported (with Gemini Pro Vision)

```bash
# Analyze image (when image support added to mcp-cli)
./mcp-cli query "Describe this image" \
  --provider vertex-ai \
  --model gemini-pro-vision \
  --image path/to/image.jpg
```

### Long Context

**Status**: ✅ Up to 2M tokens

Gemini 1.5 Pro supports massive context windows:
- Process entire codebases
- Analyze long documents
- Multi-document reasoning

---

## Usage Examples

### Example 1: Enterprise Chat with Tool Calling

**Scenario**: Internal company assistant with web search

```bash
# Query with automatic tool use
./mcp-cli query "What are the latest developments in quantum computing? \
Search recent papers and summarize the top 3 breakthroughs" \
  --provider vertex-ai \
  --verbose

# Gemini will:
# 1. Call web_search tool
# 2. Receive search results
# 3. Synthesize answer
```

**Expected**: Comprehensive answer with citations from web search.

### Example 2: Semantic Search with Embeddings

**Scenario**: Build internal document search

```bash
# Generate embeddings for documents
for doc in docs/*.txt; do
  ./mcp-cli embed "$(cat $doc)" \
    --provider vertex-ai \
    --model text-embedding-004 \
    > "embeddings/$(basename $doc .txt).json"
done

# Store in vector database, then search!
```

### Example 3: Multi-Region Deployment

**Scenario**: Data residency in Australia

```bash
# Use Sydney region for data residency
export GCP_LOCATION=australia-southeast1

./mcp-cli query "Analyze customer data" \
  --provider vertex-ai

# Data stays in Australia region
```

### Example 4: Long Document Analysis

**Scenario**: Analyze entire codebase

```bash
# Gemini 1.5 Pro can handle massive context
./mcp-cli query "Analyze this entire codebase for security issues" \
  --provider vertex-ai \
  --model gemini-1.5-pro \
  --input codebase.txt
```

### Example 5: Multilingual Embeddings

**Scenario**: Multi-language support system

```bash
# Use multilingual model
./mcp-cli embed "Hello world" \
  --provider vertex-ai \
  --model text-multilingual-embedding-002

./mcp-cli embed "Bonjour le monde" \
  --provider vertex-ai \
  --model text-multilingual-embedding-002

# Embeddings are comparable across languages!
```

---

## Troubleshooting

### Common Issues

#### Issue: Permission Denied (403)

**Symptoms**:
```
Error: Vertex AI API error (403): Permission denied
```

**Cause**: Service account lacks required permissions

**Solution**:
```bash
# Grant Vertex AI User role
gcloud projects add-iam-policy-binding YOUR_PROJECT_ID \
  --member="serviceAccount:vertex-ai-client@YOUR_PROJECT_ID.iam.gserviceaccount.com" \
  --role="roles/aiplatform.user"

# Verify role
gcloud projects get-iam-policy YOUR_PROJECT_ID \
  --flatten="bindings[].members" \
  --filter="bindings.members:vertex-ai-client@YOUR_PROJECT_ID.iam.gserviceaccount.com"
```

#### Issue: Model Not Found (404)

**Symptoms**:
```
Error: model 'gemini-3-flash-preview' not found in 'us-central1'
```

**Cause**: Model not available in your region or model retired

**Solution**:
```bash
# Use stable models
# ✅ gemini-2.5-flash (recommended)
# ✅ gemini-2.5-pro
# ✅ gemini-2.0-flash-001

# Update config
default_model: gemini-2.5-flash

# Or use global region for preview models
export GCP_LOCATION=global
```

#### Issue: Credentials Not Found

**Symptoms**:
```
Error: failed to load service account: no such file
```

**Cause**: `GOOGLE_APPLICATION_CREDENTIALS` path incorrect

**Solution**:
```bash
# Use absolute path
export GOOGLE_APPLICATION_CREDENTIALS=/home/user/vertex-ai-key.json

# Verify file exists
ls -la $GOOGLE_APPLICATION_CREDENTIALS

# Check file permissions
chmod 600 $GOOGLE_APPLICATION_CREDENTIALS
```

#### Issue: API Not Enabled

**Symptoms**:
```
Error: Vertex AI API has not been enabled
```

**Cause**: Vertex AI API not activated

**Solution**:
```bash
# Enable via gcloud
gcloud services enable aiplatform.googleapis.com \
  --project=YOUR_PROJECT_ID

# Or enable via console
# https://console.cloud.google.com/apis/library/aiplatform.googleapis.com
```

#### Issue: Billing Not Enabled

**Symptoms**:
```
Error: billing must be enabled
```

**Cause**: GCP project doesn't have billing account

**Solution**:
1. Go to [Billing](https://console.cloud.google.com/billing)
2. Link billing account to project
3. Set up budget alerts

#### Issue: Embedding Model Not Supported

**Symptoms**:
```
Error: OpenMaaS model 'google/text-embedding-004' not supported
```

**Cause**: Tried to use embedding via OpenAI endpoint (doesn't work)

**Solution**: This is handled automatically! The hybrid client uses native endpoint for embeddings.

If you see this error, update to latest mcp-cli:
```bash
git pull
go build -o mcp-cli
```

---

## Cost & Limits

### Pricing (as of Dec 2024)

**Gemini 2.5 Flash**:
- Input: $0.075 per million tokens
- Output: $0.30 per million tokens

**Gemini 2.5 Pro**:
- Input: $1.25 per million tokens
- Output: $5.00 per million tokens

**Embeddings (text-embedding-004)**:
- $0.025 per million tokens

**Free Tier**:
- Limited free usage per month (check GCP console)

### Rate Limits

Default quotas per project:
- **Requests per minute**: 300
- **Tokens per minute**: 4M (Gemini Flash)
- **Requests per day**: 15,000

**Increase limits**: Request quota increase in GCP Console

### Context Windows

| Model | Input | Output | Best For |
|-------|-------|--------|----------|
| `gemini-2.5-flash` | 1M tokens | 8K tokens | Fast, efficient |
| `gemini-2.5-pro` | 2M tokens | 8K tokens | Complex reasoning |
| `gemini-2.0-flash-001` | 1M tokens | 8K tokens | Production |
| `gemini-1.5-pro` | 2M tokens | 8K tokens | Long documents |

### Cost Optimization Tips

1. **Use Flash over Pro** when possible (4x cheaper)
2. **Batch embeddings** to reduce API calls
3. **Cache prompts** for repeated queries
4. **Monitor usage** via Cloud Console
5. **Set budget alerts** to avoid surprises
6. **Use regional models** (sometimes cheaper)

---

## Related Resources

- **Vertex AI Console**: https://console.cloud.google.com/vertex-ai
- **Official Docs**: https://cloud.google.com/vertex-ai/docs
- **Pricing**: https://cloud.google.com/vertex-ai/pricing
- **Model Garden**: https://cloud.google.com/vertex-ai/docs/start/explore-models
- **Quotas**: https://cloud.google.com/vertex-ai/quotas
- **Status Page**: https://status.cloud.google.com/
- **Support**: https://cloud.google.com/support

### GCP Resources

- **IAM Best Practices**: https://cloud.google.com/iam/docs/best-practices
- **Service Accounts**: https://cloud.google.com/iam/docs/service-account-overview
- **VPC Service Controls**: https://cloud.google.com/vpc-service-controls
- **Cloud Logging**: https://cloud.google.com/logging

---

## Provider Comparison

**vs Gemini Public API**:
- ✅ Pros: Enterprise features, VPC-SC, IAM control, data residency
- ❌ Cons: More setup, requires GCP, slightly more expensive

**vs OpenAI**:
- ✅ Pros: Longer context (2M), multimodal, cheaper (Flash)
- ❌ Cons: Setup complexity, GCP lock-in

**vs AWS Bedrock**:
- ✅ Pros: Google models, simpler auth, unified platform
- ❌ Cons: Google-only models, less provider choice

**vs Anthropic**:
- ✅ Pros: Longer context, multimodal, cheaper (for long inputs)
- ❌ Cons: Claude may be better at reasoning/coding

---

**Last Updated**: December 28, 2024  
**Verified Against**: mcp-cli v1.0.0, Vertex AI API v1  
**Code Reference**: `internal/providers/ai/clients/gcp_vertex_ai_openai.go`

**Implementation Notes**:
- Hybrid architecture verified in code
- OpenAI-compatible endpoint for chat: ✅ Working
- Native endpoint for embeddings: ✅ Working  
- OAuth2 token refresh: ✅ Implemented
- Model name conversion: ✅ Automatic
