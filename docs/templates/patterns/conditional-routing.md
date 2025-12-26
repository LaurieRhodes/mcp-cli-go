# Conditional Routing Pattern

Route requests to appropriate handlers based on classification and conditions.

---

## Overview

The **Conditional Routing Pattern** enables intelligent request routing by:
- Classifying input to understand what kind of request it is
- Routing to the appropriate workflow for that type
- Selecting the best AI provider for the task
- Adapting processing based on complexity

**Use when:**
- Different inputs need different processing (code vs documents vs data)
- Provider selection matters (some AIs better at certain tasks)
- Processing complexity varies (simple vs complex requests)
- Cost optimization is important (use cheap models when possible)

**Real-world benefit:** Route simple questions to fast local models (free), complex analysis to premium models ($$) - saves money and time.

---

## Pattern Structure

```
Input → Classify → Route to Handler → Process → Output
```

**What happens:**
1. User provides input
2. Classification step determines type/complexity
3. Conditional logic routes to appropriate handler
4. Specialized handler processes the request
5. Result returned to user

### Basic Routing Pattern

Here's a complete example that classifies requests and routes them to specialized handlers:

```yaml
name: basic_routing
description: Route requests based on classification
version: 1.0.0

config:
  defaults:
    provider: anthropic
    model: claude-sonnet-4

steps:
  # Step 1: Classify the incoming request
  - name: classify_request
    prompt: |
      Classify this request into exactly one category:
      {{input_data.request}}
      
      Categories:
      - code: Code-related tasks (writing, reviewing, debugging)
      - document: Document processing (analysis, summarization)
      - research: Research and analysis (investigation, fact-finding)
      - creative: Creative writing (stories, marketing copy)
      - data: Data processing (transformation, analysis)
      
      Return only the category name.
    output: category
  
  # Step 2: Route to code handler (if code request)
  - name: handle_code
    condition: "{{category}} == 'code'"
    provider: openai
    model: gpt-4o  # GPT-4 excels at code tasks
    prompt: |
      Handle this code request:
      {{input_data.request}}
    output: result
  
  # Step 3: Route to document handler (if document request)
  - name: handle_document
    condition: "{{category}} == 'document'"
    provider: anthropic
    model: claude-sonnet-4  # Claude excels at long documents
    prompt: |
      Handle this document request:
      {{input_data.request}}
    output: result
  
  # Step 4: Route to research handler (if research request)
  - name: handle_research
    condition: "{{category}} == 'research'"
    template: research_workflow  # Complex workflow for research
    template_input: "{{input_data.request}}"
    output: result
  
  # Step 5: Route to creative handler (if creative request)
  - name: handle_creative
    condition: "{{category}} == 'creative'"
    provider: anthropic
    model: claude-sonnet-4
    temperature: 0.9  # Higher temperature for more creativity
    prompt: |
      Handle this creative request:
      {{input_data.request}}
    output: result
  
  # Step 6: Default handler (catches anything else)
  - name: handle_default
    condition: "{{result}} is empty"
    prompt: |
      Handle this general request:
      {{input_data.request}}
    output: result
```

**Usage:**
```bash
# Code request - routes to GPT-4
mcp-cli --template basic_routing --input-data '{
  "request": "Write a Python function to calculate fibonacci numbers"
}'

# Document request - routes to Claude
mcp-cli --template basic_routing --input-data '{
  "request": "Summarize this 50-page technical specification"
}'

# Creative request - routes to Claude with high temperature
mcp-cli --template basic_routing --input-data '{
  "request": "Write a compelling product description for smart glasses"
}'
```

**What happens:**
1. User provides request
2. Classification step analyzes request → determines it's "code"
3. Only `handle_code` step runs (others skip due to conditions)
4. GPT-4 processes the code request
5. Result returned to user

**Why this works:**
- Each AI model handles what it's best at
- Only one handler runs (saves API costs)
- Easy to add new categories
- Clear separation of concerns

---

## Pattern: Complexity-Based Routing

**What it does:** Routes simple tasks to cheap/fast models, complex tasks to expensive/capable models.

**Use when:** You want to save money by using free local models for easy questions, paid APIs only for hard problems.

**Cost savings:** Can reduce API costs by 80% if most requests are simple.

```yaml
name: complexity_routing
description: Route based on task complexity
version: 1.0.0

steps:
  # Step 1: Assess how complex the task is
  - name: assess_complexity
    provider: ollama
    model: llama3.2  # Fast free model for assessment
    prompt: |
      Rate the complexity of this task:
      {{input_data.task}}
      
      Return exactly one word:
      - simple: Can be answered quickly (facts, definitions, basic math)
      - medium: Needs some thought (analysis, comparison)
      - complex: Requires deep reasoning (strategy, nuanced judgment)
    output: complexity
  
  # Route: Simple tasks → Free local model
  - name: handle_simple
    condition: "{{complexity}} == 'simple'"
    provider: ollama
    model: llama3.2  # Fast and free!
    prompt: "{{input_data.task}}"
    output: result
  
  # Route: Medium tasks → Balanced paid model
  - name: handle_medium
    condition: "{{complexity}} == 'medium'"
    provider: openai
    model: gpt-4o-mini  # Fast and affordable
    prompt: "{{input_data.task}}"
    output: result
  
  # Route: Complex tasks → Best paid model
  - name: handle_complex
    condition: "{{complexity}} == 'complex'"
    provider: anthropic
    model: claude-sonnet-4  # Most capable
    prompt: "{{input_data.task}}"
    output: result
```

**Usage:**
```bash
# Simple question - costs $0 (local model)
mcp-cli --template complexity_routing --input-data '{
  "task": "What is the capital of France?"
}'

# Medium question - costs ~$0.01
mcp-cli --template complexity_routing --input-data '{
  "task": "Compare REST vs GraphQL APIs"
}'

# Complex question - costs ~$0.03
mcp-cli --template complexity_routing --input-data '{
  "task": "Design a scalable microservices architecture for e-commerce"
}'
```

**What happens:**
1. Assessment step rates complexity → returns "simple", "medium", or "complex"
2. Only the matching handler runs (others skip due to conditions)
3. Simple: Ollama (free) → costs $0
4. Medium: GPT-4o-mini → costs ~$0.01
5. Complex: Claude Sonnet → costs ~$0.03

**Cost analysis:**
- If 80% of requests are simple: 80% free, 20% paid
- Average cost: 0.80 × $0 + 0.15 × $0.01 + 0.05 × $0.03 = $0.003 per request
- vs. always using Claude: $0.03 per request
- **Savings: 90%**

**Trade-off:** Assessment step adds ~1 second, but saves money.

---

## Pattern: Multi-Stage Classification

**What it does:** Classifies in multiple stages for precise routing (broad category first, then subcategory).

**Use when:** You have many specialized handlers and need precise routing.

**Benefits:** More accurate routing than single classification step.

```yaml
name: multi_stage_classification
description: Classify in stages for precision
version: 1.0.0

steps:
  # Stage 1: Broad classification
  - name: classify_domain
    prompt: |
      What is the broad domain of this request?
      {{input_data.request}}
      
      Domains (choose one):
      - technical: Programming, infrastructure, security
      - business: Analysis, strategy, operations
      - creative: Writing, design, marketing
      - general: Everything else
      
      Return only the domain name.
    output: domain
  
  # Stage 2: Technical sub-classification (only if technical)
  - name: classify_technical
    condition: "{{domain}} == 'technical'"
    prompt: |
      What type of technical request is this?
      {{input_data.request}}
      
      Types (choose one):
      - code: Writing or reviewing code
      - infrastructure: DevOps, deployment, servers
      - security: Security analysis, vulnerabilities
      - architecture: System design, patterns
      
      Return only the type.
    output: tech_type
  
  # Route: Code requests
  - name: handle_code
    condition: "{{tech_type}} == 'code'"
    provider: openai
    model: gpt-4o  # Best for code
    prompt: "{{input_data.request}}"
    output: result
  
  # Route: Infrastructure requests
  - name: handle_infrastructure
    condition: "{{tech_type}} == 'infrastructure'"
    template: infrastructure_workflow
    template_input: "{{input_data.request}}"
    output: result
  
  # Route: Security requests
  - name: handle_security
    condition: "{{tech_type}} == 'security'"
    template: security_workflow
    template_input: "{{input_data.request}}"
    output: result
  
  # Route: Architecture requests
  - name: handle_architecture
    condition: "{{tech_type}} == 'architecture'"
    provider: anthropic
    model: claude-sonnet-4  # Best for system thinking
    prompt: "{{input_data.request}}"
    output: result
  
  # Stage 2: Business sub-classification (only if business)
  - name: classify_business
    condition: "{{domain}} == 'business'"
    prompt: |
      What type of business request is this?
      {{input_data.request}}
      
      Types: analysis, strategy, operations, finance
      Return only the type.
    output: business_type
  
  # Route business analysis
  - name: handle_analysis
    condition: "{{business_type}} == 'analysis'"
    template: business_analysis_workflow
    template_input: "{{input_data.request}}"
    output: result
  
  # Default handler for general/creative
  - name: handle_other
    condition: "{{result}} is empty"
    prompt: "{{input_data.request}}"
    output: result
```

**Usage:**
```bash
# Technical code request
mcp-cli --template multi_stage --input-data '{
  "request": "Review this Python function for bugs"
}'
# → domain: technical → tech_type: code → handle_code (GPT-4)

# Technical security request  
mcp-cli --template multi_stage --input-data '{
  "request": "Audit this authentication flow"
}'
# → domain: technical → tech_type: security → handle_security (security workflow)

# Business analysis request
mcp-cli --template multi_stage --input-data '{
  "request": "Analyze Q4 sales trends"
}'
# → domain: business → business_type: analysis → handle_analysis (analysis workflow)
```

**What happens:**
1. Broad classification: "technical" or "business" or "creative"
2. Sub-classification (only for that domain): "code" or "infrastructure" etc.
3. Precise routing to specialized handler
4. Only 2-3 steps run (not all branches)

**Why multi-stage:**
- More accurate than trying to classify into 10+ categories at once
- Each stage is simpler (fewer options)
- Easy to add new subcategories
- Clear hierarchy

---

## Pattern: Intent-Based Routing

Route based on user intent.

```yaml
name: intent_routing
steps:
  # Detect intent
  - name: detect_intent
    prompt: |
      Detect user intent:
      {{user_input}}
      
      Intents:
      - question: User asking a question
      - task: User wants something done
      - feedback: User providing feedback
      - command: User giving a command
      
      Also extract:
      - urgency: low/medium/high
      - requires_tools: yes/no
    output: intent
  
  # Route: Questions
  - name: answer_question
    condition: "{{intent.type}} == 'question'"
    prompt: |
      Answer this question:
      {{user_input}}
      
      {% if intent.requires_tools == 'yes' %}
      Use available tools to find accurate information.
      {% endif %}
  
  # Route: Tasks
  - name: execute_task
    condition: "{{intent.type}} == 'task'"
    prompt: |
      Execute this task:
      {{user_input}}
      
      {% if intent.urgency == 'high' %}
      Prioritize speed and reliability.
      {% endif %}
  
  # Route: High-urgency to faster model
  - name: urgent_response
    condition: "{{intent.urgency}} == 'high'"
    provider: openai
    model: gpt-4o-mini  # Faster
  
  # Route: Low-urgency to better model
  - name: thorough_response
    condition: "{{intent.urgency}} == 'low'"
    provider: anthropic
    model: claude-sonnet-4  # More thorough
```

---

## Pattern: Language-Based Routing

Route based on input language or format.

```yaml
name: language_routing
steps:
  # Detect language
  - name: detect_language
    prompt: |
      Detect language of:
      {{input}}
      
      Return: language code (en, es, fr, de, zh, etc.)
    output: language
  
  # Route to native language model if available
  - name: native_model
    condition: "{{language}} == 'zh'"
    provider: deepseek  # Chinese model
    prompt: "{{input}}"
  
  # Route to multilingual model
  - name: multilingual_model
    condition: "{{language}} != 'zh' and {{language}} != 'en'"
    provider: anthropic
    model: claude-sonnet-4  # Good multilingual
    prompt: "{{input}}"
  
  # Route to English-optimized
  - name: english_model
    condition: "{{language}} == 'en'"
    provider: openai
    model: gpt-4o
    prompt: "{{input}}"
```

---

## Pattern: Content-Type Routing

Route based on content type (code, text, data, etc.).

```yaml
name: content_type_routing
steps:
  # Detect content type
  - name: detect_content_type
    prompt: |
      Detect content type:
      {{content}}
      
      Types:
      - code: Programming code
      - prose: Natural language text
      - data: Structured data (JSON, CSV, etc.)
      - mixed: Combination of types
      
      Also detect:
      - language: Programming language if code
      - format: Data format if data
    output: content_type
  
  # Route: Code
  - name: handle_code
    condition: "{{content_type.type}} == 'code'"
    provider: openai
    model: gpt-4o
    system_prompt: |
      You are an expert in {{content_type.language}}.
      Focus on code quality, best practices, and clarity.
    prompt: "{{task}} for this code: {{content}}"
  
  # Route: Data
  - name: handle_data
    condition: "{{content_type.type}} == 'data'"
    prompt: |
      Process {{content_type.format}} data:
      {{content}}
      
      Task: {{task}}
  
  # Route: Prose
  - name: handle_prose
    condition: "{{content_type.type}} == 'prose'"
    provider: anthropic
    model: claude-sonnet-4
    prompt: "{{task}} for this text: {{content}}"
```

---

## Real-World Examples

### Example 1: Customer Support Routing

```yaml
name: support_routing
version: 1.0.0

steps:
  # Classify support ticket
  - name: classify_ticket
    prompt: |
      Classify support ticket:
      {{ticket}}
      
      Categories:
      - bug: Technical issue
      - feature: Feature request
      - question: General question
      - account: Account/billing issue
      
      Priority: low/medium/high/critical
      Department: tech/sales/billing/general
    output: classification
  
  # Route critical to immediate response
  - name: critical_response
    condition: "{{classification.priority}} == 'critical'"
    provider: openai
    model: gpt-4o
    system_prompt: "Provide immediate, actionable response"
    prompt: |
      Critical ticket:
      {{ticket}}
      
      Provide:
      - Immediate workaround if possible
      - Clear next steps
      - Escalation path
  
  # Route bugs to technical team template
  - name: handle_bug
    condition: "{{classification.category}} == 'bug'"
    template: bug_triage_template
    template_input: "{{ticket}}"
  
  # Route features to product team
  - name: handle_feature
    condition: "{{classification.category}} == 'feature'"
    template: feature_request_template
    template_input: "{{ticket}}"
  
  # Route billing to billing system
  - name: handle_billing
    condition: "{{classification.department}} == 'billing'"
    servers: [billing-system]
    prompt: |
      Look up account and address:
      {{ticket}}
```

### Example 2: Code Analysis Routing

```yaml
name: code_analysis_routing
steps:
  # Detect programming language
  - name: detect_language
    prompt: |
      Detect programming language:
      {{code}}
      
      Return: language name
    output: language
  
  # Route Go code to specialized handler
  - name: analyze_go
    condition: "{{language}} == 'Go'"
    provider: openai
    model: gpt-4o
    system_prompt: |
      You are a Go expert.
      Know Go idioms, standard library, and best practices.
    prompt: "Analyze this Go code: {{code}}"
  
  # Route Python to different handler
  - name: analyze_python
    condition: "{{language}} == 'Python'"
    provider: anthropic
    system_prompt: |
      You are a Python expert.
      Know PEP 8, type hints, and modern Python practices.
    prompt: "Analyze this Python code: {{code}}"
  
  # Detect code issues
  - name: detect_issues
    prompt: |
      Analyze the review:
      {{analyze_go || analyze_python}}
      
      Categorize issues:
      - security: Security vulnerabilities
      - performance: Performance problems
      - style: Style violations
      - bugs: Logical errors
    output: issues
  
  # Route security issues to security workflow
  - name: security_deep_dive
    condition: "{{issues.security | length}} > 0"
    template: security_analysis_template
    template_input: |
      Code: {{code}}
      Issues: {{issues.security}}
```

### Example 3: Document Processing Router

```yaml
name: document_router
steps:
  # Classify document
  - name: classify_document
    prompt: |
      Classify document:
      {{document}}
      
      Type: contract, report, email, article, technical, other
      Length: short (<1000 words), medium, long (>5000 words)
      Language: en, es, fr, de, zh, other
    output: doc_info
  
  # Route long documents to chunking workflow
  - name: process_long_document
    condition: "{{doc_info.length}} == 'long'"
    template: long_document_template
    template_input: "{{document}}"
  
  # Route contracts to legal workflow
  - name: process_contract
    condition: "{{doc_info.type}} == 'contract'"
    template: legal_review_template
    template_input: "{{document}}"
  
  # Route technical docs to technical workflow
  - name: process_technical
    condition: "{{doc_info.type}} == 'technical'"
    provider: openai  # Good for technical content
    prompt: "Analyze technical document: {{document}}"
  
  # Route short documents to simple processing
  - name: process_short
    condition: "{{doc_info.length}} == 'short'"
    provider: ollama  # Fast local model for short docs
    prompt: "Summarize: {{document}}"
```

---

## Best Practices

### 1. Clear Classification

```yaml
# Good: Specific categories
- prompt: |
    Classify as exactly one:
    - bug_report
    - feature_request
    - question
    - feedback

# Bad: Vague categories
- prompt: "What kind of thing is this?"
```

### 2. Handle Edge Cases

```yaml
# Good: Default route
- name: specific_handler
  condition: "{{category}} == 'specific'"
  
- name: default_handler
  condition: "{{result}} is empty"  # Catch-all

# Bad: No default
- condition: "{{category}} == 'specific'"
# What if category doesn't match?
```

### 3. Optimize for Common Cases

```yaml
# Good: Most common first
- name: handle_common_case  # 80% of requests
  condition: "{{type}} == 'common'"
  provider: ollama  # Fast and cheap
  
- name: handle_rare_case  # 20% of requests
  condition: "{{type}} == 'rare'"
  provider: anthropic  # Best quality

# Bad: Expensive model for everything
- provider: anthropic  # $$$
```

### 4. Validate Routing

```yaml
# Good: Log routing decisions
- name: classify
  output: category
  
- name: log_routing
  prompt: |
    Log routing decision:
    Input: {{input | truncate: 100}}
    Category: {{category}}
    Handler: {{next_step}}

# Bad: Silent routing
# (Hard to debug)
```

---

## Performance Optimization

### Fast Classification

```yaml
# Fast: Simple local model for classification
- name: classify
  provider: ollama
  model: llama3.2  # Fast
  prompt: "Classify: {{input}}"

# Then route to appropriate provider
- name: handle_code
  condition: "{{classify}} == 'code'"
  provider: openai  # Specialized
```

### Parallel Classification

```yaml
# Multiple classifiers in parallel
- parallel:
    - name: classify_type
      prompt: "Type?"
    - name: classify_complexity
      prompt: "Complexity?"
    - name: classify_language
      prompt: "Language?"
  max_concurrent: 3
  aggregate: merge

# Route based on combined classification
- condition: "{{classify_type}} == 'code' and {{classify_complexity}} == 'high'"
  provider: anthropic
```

### Early Exit

```yaml
# Exit early for simple cases
- name: check_simple
  prompt: "Is this simple? yes/no"

- name: quick_response
  condition: "{{check_simple}} == 'yes'"
  provider: ollama
  output: final_result

# Skip expensive processing if done
- name: complex_processing
  condition: "{{final_result}} is empty"
  provider: anthropic
```

---

## Complete Example

```yaml
name: intelligent_router
version: 1.0.0

config:
  defaults:
    provider: anthropic
    model: claude-sonnet-4

steps:
  # 1. Multi-factor classification
  - name: classify
    provider: ollama  # Fast classification
    prompt: |
      Classify request:
      {{request}}
      
      Return JSON:
      {
        "type": "code|document|data|question",
        "complexity": "simple|medium|complex",
        "urgency": "low|high",
        "language": "detected language"
      }
    output: classification
  
  # 2. Route simple + low urgency to cheap model
  - name: simple_handler
    condition: "{{classification.complexity}} == 'simple'"
    provider: ollama
    prompt: "{{request}}"
    output: result
  
  # 3. Route code to specialized model
  - name: code_handler
    condition: "{{classification.type}} == 'code' and {{result}} is empty"
    provider: openai
    model: gpt-4o
    prompt: "{{request}}"
    output: result
  
  # 4. Route urgent to fast model
  - name: urgent_handler
    condition: "{{classification.urgency}} == 'high' and {{result}} is empty"
    provider: openai
    model: gpt-4o-mini
    prompt: "{{request}}"
    output: result
  
  # 5. Route complex to best model
  - name: complex_handler
    condition: "{{classification.complexity}} == 'complex' and {{result}} is empty"
    provider: anthropic
    model: claude-sonnet-4
    prompt: "{{request}}"
    output: result
  
  # 6. Default handler
  - name: default_handler
    condition: "{{result}} is empty"
    prompt: "{{request}}"
    output: result
```

---

## Quick Reference

```yaml
# Basic routing
classify → route_A | route_B | route_C → output

# Complexity-based
assess_complexity → simple | medium | complex

# Multi-stage
broad_classify → sub_classify → route

# Intent-based
detect_intent → question | task | command

# Content-based
detect_type → code | text | data → handler
```

---

## Next Steps

- **[Validation Pattern](validation.md)** - Multi-provider validation
- **[Research Pattern](research-agent.md)** - Deep research workflows
- **[Examples](../examples/)** - Working templates

---

**Route intelligently, process efficiently!** 🎯
