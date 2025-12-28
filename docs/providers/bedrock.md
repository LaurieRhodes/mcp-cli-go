# AWS Bedrock

> **Skill Level**: 🟡 Intermediate (AWS knowledge required)  
> **Interface**: AWS Custom (Bedrock API)  
> **Best For**: Enterprise AWS users, multi-model access, regulated industries

## Quick Start

**For Beginners**: Access multiple AI models through AWS in 4 steps.

```bash
# 1. Set up AWS credentials
export AWS_ACCESS_KEY_ID=AKIA...
export AWS_SECRET_ACCESS_KEY=...
export AWS_REGION=us-east-1

# 2. Enable Bedrock models in AWS Console
# Go to: https://console.aws.amazon.com/bedrock

# 3. Initialize configuration
./mcp-cli init
# Select "AWS Bedrock" when prompted

# 4. Test it
./mcp-cli query "Hello from Bedrock!" --provider bedrock
```

---

## Overview

### What is AWS Bedrock?

AWS Bedrock is Amazon's fully managed service that provides access to foundation models from leading AI companies through a single API. It offers Claude (Anthropic), Llama (Meta), Titan (Amazon), and more with enterprise security.

### When to Use

- ✅ **Use when**:
  - You're already on **AWS infrastructure**
  - You need **multiple model providers** (Claude, Llama, etc.)
  - You require **enterprise security** (VPC, KMS, IAM)
  - You want **AWS integrations** (Lambda, SageMaker, S3)
  - You need **data residency** guarantees
  - You require **compliance** (SOC, HIPAA, etc.)

- ❌ **Avoid when**:
  - You want **simplest setup** (use direct provider APIs)
  - You're **not on AWS** (higher complexity)
  - You need **latest models** (Bedrock may lag behind direct APIs)

- 🤔 **Consider alternatives**:
  - **Direct APIs**: OpenAI, Anthropic (simpler, sometimes newer models)
  - **GCP Vertex AI**: If you're on Google Cloud
  - **Azure OpenAI**: If you're on Azure

### Key Features

- **Chat/Completion**: ✅ Multi-model support
- **Streaming**: ✅ Fully Supported
- **Tool Calling (MCP)**: ✅ Supported (Claude, Llama 3.1+)
- **Embeddings**: ✅ Multiple providers
- **Vision**: ✅ Model-dependent (Claude Sonnet/Opus)
- **Multi-Model**: ✅ Claude, Llama, Titan, Cohere, and more

---

## Prerequisites

### Required

- [ ] **AWS Account** ([Sign up](https://aws.amazon.com/))
- [ ] **AWS IAM User** with Bedrock permissions
- [ ] **Model Access** enabled in Bedrock console
- [ ] **AWS Credentials** (Access Key ID + Secret)

### Optional

- [ ] **AWS CLI** installed ([Install](https://aws.amazon.com/cli/))
- [ ] **VPC Configuration** (for private deployments)
- [ ] **KMS Keys** (for encryption)

### Cost Awareness

⚠️ **Bedrock charges per token**. Prices vary by model. See [Cost & Limits](#cost--limits).

---

## Setup Guide

### Step 1: Create AWS Account

1. Go to https://aws.amazon.com/
2. Click **Create an AWS Account**
3. Follow signup process
4. Add payment method

### Step 2: Create IAM User

**For Beginners**: IAM users are like sub-accounts for your AWS account.

1. Go to [IAM Console](https://console.aws.amazon.com/iam/)
2. Click **Users** → **Add users**
3. Username: `bedrock-user`
4. Select **Access key - Programmatic access**
5. Click **Next**

### Step 3: Attach Bedrock Permissions

1. Select **Attach policies directly**
2. Search for `AmazonBedrockFullAccess`
3. Check the policy
4. Click **Next** → **Create user**

Or create custom policy:
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "bedrock:InvokeModel",
        "bedrock:InvokeModelWithResponseStream"
      ],
      "Resource": "*"
    }
  ]
}
```

### Step 4: Create Access Keys

1. Click on the user you created
2. Go to **Security credentials** tab
3. Click **Create access key**
4. Choose **Application running outside AWS**
5. Click **Next** → **Create access key**
6. **Save both keys securely**!

### Step 5: Enable Model Access

**For Beginners**: You must request access to each model.

1. Go to [Bedrock Console](https://console.aws.amazon.com/bedrock/)
2. Click **Model access** in left menu
3. Click **Manage model access**
4. Check models you want:
   - **Anthropic Claude 3.5 Sonnet**
   - **Meta Llama 3.1**
   - **Amazon Titan**
   - **Cohere**
5. Click **Request model access**
6. Wait for approval (usually instant for standard models)

### Step 6: Set Environment Variables

```bash
# Add to .env file
export AWS_ACCESS_KEY_ID=AKIA...
export AWS_SECRET_ACCESS_KEY=...
export AWS_REGION=us-east-1

# Optional: for temporary credentials
# export AWS_SESSION_TOKEN=...
```

### Step 7: Test Connection

```bash
# Initialize configuration
./mcp-cli init
# Select "AWS Bedrock"

# Test with Claude
./mcp-cli query "What is the capital of France?" --provider bedrock

# Test with different model
./mcp-cli query "Hello" --provider bedrock --model anthropic.claude-3-haiku-20240307-v1:0
```

---

## Configuration Reference

### Provider Configuration

**File**: `config/providers/aws-bedrock.yaml`

```yaml
interface_type: aws_bedrock
provider_name: bedrock
config:
  # AWS credentials (required)
  aws_region: ${AWS_REGION}
  aws_access_key_id: ${AWS_ACCESS_KEY_ID}
  aws_secret_access_key: ${AWS_SECRET_ACCESS_KEY}
  # aws_session_token: ${AWS_SESSION_TOKEN}  # Optional
  
  # Default model
  default_model: anthropic.claude-3-5-sonnet-20241022-v2:0
  
  # Request settings
  timeout_seconds: 300
  max_retries: 3
```

#### Configuration Options

| Option | Type | Required | Default | Description |
|--------|------|----------|---------|-------------|
| `aws_region` | string | Yes | - | AWS region (e.g., us-east-1) |
| `aws_access_key_id` | string | Yes | - | AWS access key ID |
| `aws_secret_access_key` | string | Yes | - | AWS secret access key |
| `aws_session_token` | string | No | - | Session token (for temp creds) |
| `default_model` | string | Yes | Claude 3.5 Sonnet | Default model ID |
| `timeout_seconds` | int | No | 300 | Request timeout |
| `max_retries` | int | No | 3 | Retry attempts |

#### Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `AWS_ACCESS_KEY_ID` | Yes | AWS access key |
| `AWS_SECRET_ACCESS_KEY` | Yes | AWS secret key |
| `AWS_REGION` | Yes | AWS region |
| `AWS_SESSION_TOKEN` | No | Temporary session token |

#### Available Models

**Anthropic Claude** (Recommended):
- `anthropic.claude-3-5-sonnet-20241022-v2:0` - Best quality
- `anthropic.claude-3-5-haiku-20241022-v1:0` - Fast, affordable
- `anthropic.claude-3-opus-20240229-v1:0` - Most capable
- `anthropic.claude-3-sonnet-20240229-v1:0` - Previous gen
- `anthropic.claude-3-haiku-20240307-v1:0` - Fast previous gen

**Meta Llama**:
- `meta.llama3-1-405b-instruct-v1:0` - Largest, most capable
- `meta.llama3-1-70b-instruct-v1:0` - Good balance
- `meta.llama3-1-8b-instruct-v1:0` - Fast, efficient

**Amazon Titan**:
- `amazon.titan-text-premier-v1:0` - Latest flagship
- `amazon.titan-text-express-v1` - Fast, cost-effective

**Cohere**:
- `cohere.command-r-plus-v1:0` - Latest, RAG-optimized
- `cohere.command-r-v1:0` - Balanced

### Embedding Configuration

**File**: `config/embeddings/aws-bedrock.yaml`

```yaml
interface_type: aws_bedrock
provider_name: bedrock
config:
  aws_region: ${AWS_REGION}
  aws_access_key_id: ${AWS_ACCESS_KEY_ID}
  aws_secret_access_key: ${AWS_SECRET_ACCESS_KEY}
  default_model: cohere.embed-english-v3
  timeout_seconds: 30
  max_retries: 3
  embedding_models:
    cohere.embed-english-v3:
      description: Cohere English embeddings (serverless)
      dimensions: 1024
      max_tokens: 512
    cohere.embed-multilingual-v3:
      description: Cohere multilingual (100+ languages)
      dimensions: 1024
      max_tokens: 512
    amazon.titan-embed-text-v2:0:
      description: Amazon Titan Embeddings V2
      dimensions: 1024
      max_tokens: 8192
```

---

## Implementation Details

### Interface Type

**Type**: `aws_bedrock`

**What this means**:
- **For Beginners**: Custom AWS authentication using access keys
- **For Developers**: Uses AWS Signature V4 signing for requests

### How It Works

```
User → mcp-cli → AWS Bedrock Client → AWS SigV4 Auth → Bedrock API → Model
                     ↓
              AWS request signing
```

**Architecture**:
1. mcp-cli creates Bedrock-format request
2. AWS SDK signs request with SigV4 (access key + secret)
3. Request sent to Bedrock regional endpoint
4. Bedrock routes to selected model provider
5. Model generates response
6. Response in Bedrock format, converted to domain format

### API Endpoints

Regional endpoints:
```
https://bedrock-runtime.{region}.amazonaws.com
```

Examples:
- `us-east-1`: N. Virginia
- `us-west-2`: Oregon
- `eu-central-1`: Frankfurt
- `ap-southeast-1`: Singapore

### Authentication

**Method**: AWS Signature Version 4

**Flow**:
1. Create canonical request
2. Create string to sign (algorithm + timestamp + scope + hash)
3. Calculate signature using secret key
4. Add Authorization header with signature

**Handled automatically by AWS SDK**

### Model Name Format

**Format**: `provider.model-name-version`

Examples:
- `anthropic.claude-3-5-sonnet-20241022-v2:0`
- `meta.llama3-1-70b-instruct-v1:0`
- `amazon.titan-text-premier-v1:0`

---

## Features

### Multi-Model Support

```bash
# Use Claude
./mcp-cli query "Complex reasoning task" \
  --provider bedrock \
  --model anthropic.claude-3-5-sonnet-20241022-v2:0

# Use Llama
./mcp-cli query "Code generation" \
  --provider bedrock \
  --model meta.llama3-1-70b-instruct-v1:0

# Use Titan
./mcp-cli query "Simple query" \
  --provider bedrock \
  --model amazon.titan-text-premier-v1:0
```

### Tool Calling (MCP Tools)

**Status**: ✅ Supported (Claude, Llama 3.1+)

```bash
# Works with Claude models
./mcp-cli query "Search for AI news and summarize" \
  --provider bedrock \
  --model anthropic.claude-3-5-sonnet-20241022-v2:0
```

**Note**: Tool calling support varies by model:
- ✅ Anthropic Claude: Full support
- ✅ Llama 3.1+: Good support
- ❌ Titan: No tool calling

### Embeddings

```bash
# Cohere embeddings
./mcp-cli embed "Your text here" \
  --provider bedrock \
  --model cohere.embed-english-v3

# Amazon Titan embeddings
./mcp-cli embed "Your text" \
  --provider bedrock \
  --model amazon.titan-embed-text-v2:0
```

---

## Usage Examples

### Example 1: Multi-Model Comparison

**Scenario**: Test same query across models

```bash
# Test with Claude
./mcp-cli query "Explain REST APIs" --provider bedrock --model anthropic.claude-3-5-sonnet-20241022-v2:0

# Test with Llama
./mcp-cli query "Explain REST APIs" --provider bedrock --model meta.llama3-1-70b-instruct-v1:0

# Test with Titan
./mcp-cli query "Explain REST APIs" --provider bedrock --model amazon.titan-text-premier-v1:0
```

### Example 2: Cost Optimization

**Scenario**: Use cheaper models for simple tasks

```bash
# Simple queries: Use Haiku (cheapest)
./mcp-cli query "What is Python?" --provider bedrock --model anthropic.claude-3-5-haiku-20241022-v1:0

# Complex tasks: Use Sonnet
./mcp-cli query "Design a microservices architecture" --provider bedrock --model anthropic.claude-3-5-sonnet-20241022-v2:0
```

### Example 3: Enterprise RAG System

**Scenario**: Build retrieval system

```bash
# Generate embeddings with Cohere
./mcp-cli embed "$(cat document.txt)" \
  --provider bedrock \
  --model cohere.embed-english-v3

# Query with Claude
./mcp-cli query "Answer based on retrieved context: ..." \
  --provider bedrock
```

---

## Troubleshooting

### Common Issues

#### Issue: Access Denied (403)

**Symptoms**:
```
Error: AccessDeniedException: User is not authorized
```

**Cause**: IAM permissions insufficient

**Solution**:
```bash
# Add Bedrock permissions to IAM user
# Policy: AmazonBedrockFullAccess

# Or custom policy (see Setup Guide)
```

#### Issue: Model Not Found

**Symptoms**:
```
Error: Model not found or not accessible
```

**Cause**: Model access not enabled

**Solution**:
1. Go to [Bedrock Console](https://console.aws.amazon.com/bedrock/)
2. Click **Model access**
3. Enable required models
4. Wait for approval

#### Issue: Throttling

**Symptoms**:
```
Error: ThrottlingException: Rate exceeded
```

**Cause**: Too many requests

**Solution**:
```bash
# Automatic retry with exponential backoff
# Check limits: https://docs.aws.amazon.com/bedrock/latest/userguide/quotas.html

# Request quota increase if needed
```

---

## Cost & Limits

### Pricing (Dec 2024)

**Claude 3.5 Sonnet**:
- Input: $3.00 per 1M tokens
- Output: $15.00 per 1M tokens

**Claude 3.5 Haiku**:
- Input: $0.80 per 1M tokens
- Output: $4.00 per 1M tokens

**Llama 3.1 70B**:
- Input: $0.99 per 1M tokens
- Output: $0.99 per 1M tokens

**Titan Premier**:
- Input: $0.50 per 1M tokens
- Output: $1.50 per 1M tokens

**Embeddings**:
- Cohere: $0.10 per 1M tokens
- Titan: $0.02 per 1M tokens

### Rate Limits

Default quotas (per region, per account):
- **Requests per minute**: 100-1000 (model-dependent)
- **Tokens per minute**: 100K-500K

**Increase limits**: Request via AWS Support

---

## Related Resources

- **AWS Bedrock Console**: https://console.aws.amazon.com/bedrock/
- **Documentation**: https://docs.aws.amazon.com/bedrock/
- **Pricing**: https://aws.amazon.com/bedrock/pricing/
- **Model Catalog**: https://aws.amazon.com/bedrock/models/
- **IAM Guide**: https://docs.aws.amazon.com/bedrock/latest/userguide/security-iam.html

---

**Last Updated**: December 28, 2024  
**Verified Against**: mcp-cli v1.0.0, AWS Bedrock Dec 2024  
**Code Reference**: `internal/providers/ai/clients/aws_bedrock.go`
