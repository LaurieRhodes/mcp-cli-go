# Research Agent Pattern

Build autonomous research workflows that explore topics systematically.

---

## Overview

The **Research Agent Pattern** enables AI to conduct research autonomously through iteration.

**What is "autonomous research"?**
- AI breaks down your question into smaller parts
- Searches for information on each part
- Analyzes what it finds
- Identifies what's still missing
- Searches again to fill gaps
- Repeats until comprehensive
- Synthesizes everything into report

**Like a human researcher:**
1. You ask: "What's the best database for my use case?"
2. AI thinks: "I need to know: performance, scalability, cost, ease of use"
3. Researches each aspect
4. Discovers: "Wait, what about data consistency?"
5. Researches that too
6. Combines all findings
7. Gives you comprehensive answer

**Use when:**
- Need comprehensive understanding (not just quick answer)
- Topic requires multiple sources
- Want to discover related aspects you didn't think of
- Building knowledge base on a subject

**Don't use when:**
- Simple factual question ("What's the capital of France?")
- Time-sensitive (this is slow: 5-15 minutes)
- Budget-sensitive (this is expensive: $1-5 per research)
- Single source is sufficient

**Cost warning:**
- Multiple searches (5-15 web searches)
- Multiple AI calls (10-30 AI analyses)
- **Total: $1-5 per comprehensive research**
- Compare: Single web search = $0.03

**Performance:**
- Quick research: 2-3 minutes, $0.50
- Thorough research: 5-10 minutes, $2.00
- Deep research: 10-15 minutes, $5.00

---

## Pattern Structure

```
Question → Break Down → Search → Analyze → Follow-up → Search More → Synthesize
```

**The iteration cycle:**
1. **Decompose:** Break question into 3-5 sub-questions
2. **Search:** Research each sub-question (web search)
3. **Analyze:** Review findings, identify gaps
4. **Follow-up:** Generate new questions based on gaps
5. **Search again:** Fill the gaps
6. **Synthesize:** Combine everything into comprehensive report

**Why iteration matters:**
- First pass finds obvious information
- Second pass finds details and nuances
- Each iteration discovers new aspects
- Final result is much more comprehensive

### Basic Research Agent

**What it does:** Complete autonomous research with iteration and gap-filling.

**Use when:** Need comprehensive understanding of a topic.

```yaml
name: research_agent
description: Autonomous research on any topic
version: 1.0.0

config:
  defaults:
    provider: anthropic
    model: claude-sonnet-4

steps:
  # Step 1: Break down research question into sub-questions
  - name: decompose_question
    prompt: |
      Break this research question into 3-5 specific sub-questions:

      {{input_data.question}}

      Each sub-question should:
      - Address a specific aspect
      - Be searchable
      - Be answerable with facts
      
      Return as JSON array of strings.
    output: sub_questions

  # Step 2: Research each sub-question independently
  - name: research_sub_questions
    for_each: "{{sub_questions}}"
    item_name: sub_question
    servers: [brave-search]
    prompt: |
      Research this sub-question thoroughly:
      {{sub_question}}

      Search for:
      - Credible sources (official sites, research papers, expert articles)
      - Specific facts and data (numbers, statistics, examples)
      - Expert opinions (quotes from authorities)
      - Recent information (prioritize current data)
      
      Extract key findings with sources.
    output: research_findings

  # Step 3: Analyze findings and identify what's missing
  - name: identify_gaps
    prompt: |
      Review these research findings:
      {{research_findings}}
      
      Original question: {{input_data.question}}

      Identify:
      - What important questions remain unanswered?
      - What needs deeper investigation?
      - What contradictions need clarification?
      - What related aspects should be explored?

      Return top 3 follow-up questions (or empty if comprehensive).
    output: follow_up_questions

  # Step 4: Follow-up research on gaps (only if needed)
  - name: follow_up_research
    condition: "{{follow_up_questions}} not empty"
    for_each: "{{follow_up_questions}}"
    item_name: follow_up
    servers: [brave-search]
    prompt: |
      Deep dive research on:
      {{follow_up}}
      
      Focus on filling the knowledge gap identified.
    output: deep_dive_findings

  # Step 5: Synthesize all findings into comprehensive report
  - name: create_research_report
    prompt: |
      Create comprehensive research report answering:
      {{input_data.question}}

      Based on:
      - Initial Research: {{research_findings}}
      - Deep Dive Research: {{deep_dive_findings}}

      Format as markdown:
      
      # Research Report: {{input_data.question}}

      ## Executive Summary
      [3-4 sentence overview of key findings]

      ## Detailed Findings
      [Organized by sub-topic with evidence]

      ## Evidence and Sources
      [Key facts with context and source citations]

      ## Areas of Uncertainty
      [What remains unknown or debated]

      ## Recommendations
      [Actionable conclusions based on findings]
      
      ## Additional Resources
      [Links to key sources]
    output: research_report
```

**Usage:**
```bash
# Example 1: Technology research
mcp-cli --template research_agent --input-data '{
  "question": "What are the best practices for securing Kubernetes clusters in production?"
}'

# Example 2: Business research
mcp-cli --template research_agent --input-data '{
  "question": "How do successful SaaS companies handle pricing strategy?"
}'

# Example 3: Medical research
mcp-cli --template research_agent --input-data '{
  "question": "What are the current evidence-based treatments for chronic lower back pain?"
}'
```

**What happens (real example):**

Input: "What database should I use for a high-traffic e-commerce site?"

1. **Decompose:**
   - What are database performance requirements for e-commerce?
   - What databases handle high transaction volumes?
   - What are cost considerations for databases at scale?
   - What are maintenance and operational requirements?

2. **Initial Research (4 searches):**
   - Postgres: Good for relational, ACID compliance, up to 100K requests/sec
   - MongoDB: Good for flexible schema, horizontal scaling, 200K+ ops/sec
   - MySQL: Mature, reliable, good ecosystem, 80K requests/sec
   - Redis: For caching layer, sub-millisecond latency

3. **Identify Gaps:**
   - "What about data consistency needs for e-commerce?"
   - "What are actual costs at 1M daily active users?"
   - "What's the learning curve for each?"

4. **Follow-up Research (3 searches):**
   - Strong consistency needed for inventory, payments
   - Cost comparison: Postgres ~$500/mo, MongoDB ~$800/mo, MySQL ~$400/mo
   - Postgres has largest talent pool, easiest to hire for

5. **Synthesize:**
   - **Recommendation:** Postgres + Redis
   - **Rationale:** Strong consistency, proven at scale, cost-effective, easy to hire for
   - **Architecture:** Postgres for transactional data, Redis for caching
   - **Expected cost:** ~$600/month at 1M DAU

**Cost breakdown:**
- Decompose: 1 AI call = $0.03
- Initial research: 4 searches + 4 analyses = $0.40
- Gap identification: 1 AI call = $0.03
- Follow-up research: 3 searches + 3 analyses = $0.30
- Synthesis: 1 AI call = $0.03
- **Total: ~$0.80 for this research**

**Time:**
- ~5 minutes total

**When to use this vs simple web search:**
- **Use research agent:** Complex decisions, need comprehensive view
- **Use simple search:** Quick facts, single answer needed

---

## Pattern: Multi-Source Research

**What it does:** Researches same topic across different types of sources, then cross-references.

**Use when:** Need balanced view from multiple source types (web, academic, news).

**Why multiple source types:**
- Web: Practical guides, tutorials, current info
- Academic: Research papers, rigorous studies
- News: Recent developments, industry trends
- Cross-referencing finds contradictions and builds confidence

```yaml
name: multi_source_research
description: Research across multiple information sources
version: 1.0.0

steps:
  # Parallel research across different source types
  - name: gather_information
    parallel:
      # Web search (practical information)
      - name: web_research
        servers: [brave-search]
        prompt: |
          Search web for practical information on:
          {{input_data.topic}}
          
          Focus on: how-tos, best practices, case studies
        output: web_findings

      # Academic search (research papers)
      - name: academic_research
        servers: [academic-search]
        prompt: |
          Find research papers on:
          {{input_data.topic}}
          
          Focus on: peer-reviewed studies, systematic reviews
        output: academic_findings

      # News search (recent developments)
      - name: news_research
        servers: [news-api]
        prompt: |
          Find recent news and industry coverage of:
          {{input_data.topic}}
          
          Focus on: last 6 months, credible sources
        output: news_findings
    max_concurrent: 3
    aggregate: merge
    output: all_sources

  # Cross-reference findings from different sources
  - name: cross_reference
    prompt: |
      Cross-reference findings from multiple source types:

      Web Sources: {{all_sources.web_findings}}
      Academic Sources: {{all_sources.academic_findings}}
      News Sources: {{all_sources.news_findings}}

      Analyze:
      1. **Consensus points:** Where all sources agree (HIGH CONFIDENCE)
      2. **Contradictions:** Where sources disagree (FLAG FOR REVIEW)
      3. **Unique insights:** Information from only one source type
      4. **Reliability assessment:** Which sources are most credible
      5. **Recency:** Which information is most current
      
      Return structured analysis.
    output: verified_findings

  # Generate balanced report
  - name: final_report
    prompt: |
      Create balanced, well-sourced research report:
      {{verified_findings}}

      Structure:
      # Research Report: {{input_data.topic}}
      
      ## High-Confidence Findings
      [Where multiple source types agree]
      
      ## Areas of Active Debate
      [Where sources contradict - explain both sides]
      
      ## Recent Developments
      [From news sources - what's new]
      
      ## Academic Consensus
      [What research says - peer-reviewed evidence]
      
      ## Practical Guidance
      [From web sources - actionable advice]
      
      ## Limitations and Caveats
      [What we don't know, conflicts unresolved]
```

**Usage:**
```bash
# Research with multiple source types
mcp-cli --template multi_source_research --input-data '{
  "topic": "Effects of intermittent fasting on metabolic health"
}'
```

**What happens:**
1. Three searches run in parallel:
   - Web: Finds practical guides, diet plans, success stories
   - Academic: Finds clinical studies, meta-analyses
   - News: Finds recent findings, new research announcements
2. Cross-reference identifies:
   - **Consensus:** Weight loss benefit confirmed by all sources (high confidence)
   - **Contradiction:** Muscle loss risk debated (web says yes, studies say minimal if protein adequate)
   - **Unique:** News reveals new study on cognitive benefits
3. Report balances all perspectives with confidence levels

**Performance:**
- 3 parallel searches: ~15 seconds
- Cross-reference: ~10 seconds
- Report: ~5 seconds
- **Total: ~30 seconds**

**Cost:**
- 3 searches + 2 AI syntheses = ~$0.20

---

## Pattern: Iterative Deepening

**What it does:** Starts with broad overview, then progressively goes deeper into key areas.

**Use when:** Don't know what's important yet, want to discover focus areas first.

**Research depth levels:**
- **Level 1:** Broad overview (10,000 foot view)
- **Level 2:** Identify key areas (what matters most)
- **Level 3:** Deep dive on key areas (details and nuances)
- **Level 4:** Synthesis (comprehensive understanding)

```yaml
name: iterative_deepening
description: Research in expanding circles of depth
version: 1.0.0

steps:
  # Level 1: Get broad overview first
  - name: overview_research
    servers: [brave-search]
    prompt: |
      High-level overview of: {{input_data.topic}}
      
      Cover:
      - What is it?
      - Why does it matter?
      - Main concepts and categories
      - Key players or technologies
    output: overview

  # Level 2: Identify what to research deeper
  - name: identify_key_areas
    prompt: |
      From this overview:
      {{overview}}

      Identify the 3 most important sub-topics that need deeper research.
      
      Consider:
      - What's most relevant to user needs
      - What's most complex or misunderstood
      - What's most impactful
      
      Return 3 sub-topics as JSON array.
    output: key_areas

  # Level 3: Deep dive on each key area
  - name: deep_research
    for_each: "{{key_areas}}"
    item_name: area
    servers: [brave-search, academic-search]
    prompt: |
      Comprehensive deep dive research on:
      {{area}}

      Find:
      - Key concepts and terminology
      - Current state of the art
      - Recent developments and trends
      - Expert perspectives and best practices
      - Common challenges and solutions
      - Real-world examples
    output: deep_findings

  # Level 4: Synthesize into comprehensive guide
  - name: synthesize
    prompt: |
      Create comprehensive guide on: {{input_data.topic}}

      Using:
      - Overview: {{overview}}
      - Deep Research: {{deep_findings}}

      Structure logically from basics to advanced:
      
      # Complete Guide: {{input_data.topic}}
      
      ## Introduction
      [From overview - what and why]
      
      ## Fundamentals
      [Basic concepts everyone should know]
      
      ## Deep Dive Areas
      [Detailed coverage of key areas]
      
      ## Advanced Topics
      [Nuances and complexities]
      
      ## Practical Application
      [How to use this knowledge]
      
      ## Common Pitfalls
      [What to avoid]
```

**Usage:**
```bash
mcp-cli --template iterative_deepening --input-data '{
  "topic": "GraphQL API design"
}'
```

**What happens (example):**

Input: "GraphQL API design"

1. **Level 1 - Overview:**
   - GraphQL is a query language for APIs
   - Benefits: flexible queries, strongly typed, single endpoint
   - Main concepts: schemas, resolvers, queries, mutations

2. **Level 2 - Identify Key Areas:**
   - Schema design patterns (most critical)
   - Resolver optimization (performance bottleneck)
   - Error handling strategies (often overlooked)

3. **Level 3 - Deep Dive:**
   - **Schema design:** Federation, interface types, naming conventions
   - **Resolvers:** N+1 problem, DataLoader, caching strategies
   - **Error handling:** Partial errors, error extensions, client handling

4. **Level 4 - Synthesis:**
   - Complete guide from basics to advanced
   - Structured progression
   - Practical examples at each level

**Performance:**
- Overview: 15 seconds
- Identify areas: 5 seconds  
- Deep research (3 areas): 45 seconds
- Synthesis: 10 seconds
- **Total: ~75 seconds (~1.25 minutes)**

**Cost:**
- 1 overview search + 3 deep searches + 2 AI analyses = ~$0.30

**When to use:**
- Learning new complex topic
- Building internal knowledge base
- Creating training materials
- Don't know enough to ask specific questions yet

---

## Pattern: Comparative Research

Compare and contrast multiple topics.

```yaml
name: comparative_research
steps:
  # Research each topic independently
  - name: research_topics
    for_each: "{{topics}}"
    item_name: topic
    servers: [brave-search]
    prompt: |
      Research {{topic}} focusing on:
      - Key features
      - Strengths
      - Weaknesses
      - Use cases
      - Costs
    output: topic_research

  # Comparative analysis
  - name: compare
    prompt: |
      Create comparison matrix from:
      {{topic_research}}

      Format as markdown table comparing:
      - Features
      - Performance
      - Cost
      - Ease of use
      - Best for
    output: comparison

  # Recommendations
  - name: recommend
    prompt: |
      Based on comparison:
      {{comparison}}

      And user needs:
      {{requirements}}

      Recommend best option with rationale.
```

---

## Real-World Examples

### Example 1: Technology Stack Research

**Scenario:** Choosing technology stack for new project.

**Usage:**
```bash
mcp-cli --template research_agent --input-data '{
  "question": "What is the best technology stack for building a real-time chat application in 2024?"
}'
```

**What the research discovers:**

1. **Initial sub-questions:**
   - What are real-time communication requirements?
   - What technologies handle WebSocket connections at scale?
   - What databases work well with real-time data?
   - What deployment considerations exist?

2. **Initial findings:**
   - Frontend: React/Vue + Socket.io client
   - Backend: Node.js/Go for WebSocket handling
   - Database: Redis for presence, PostgreSQL for history
   - Infrastructure: Kubernetes for scaling

3. **Follow-up questions discovered:**
   - "How to handle message delivery guarantees?"
   - "What about offline message queuing?"
   - "Security considerations for WebSockets?"

4. **Final recommendation:**
   - **Stack:** Next.js + Socket.io + Redis + PostgreSQL + AWS
   - **Rationale:** Proven at scale, good ecosystem, manageable complexity
   - **Expected cost:** $200/month for 10K users
   - **Scaling limit:** Up to 1M users before architecture changes

**Research cost:** ~$1.20
**Research time:** ~8 minutes
**Value:** Saved weeks of trial and error

---

### Example 2: Market Research

**Scenario:** Evaluating market opportunity for SaaS product.

```yaml
name: market_research
version: 1.0.0

steps:
  # Industry analysis
  - name: analyze_industry
    servers: [brave-search]
    prompt: |
      Analyze {{input_data.industry}} market:
      
      Find:
      - Total addressable market (TAM) size
      - Growth rate and trends
      - Key players and market share
      - Market dynamics and forces
    output: industry_analysis

  # Competitor research
  - name: research_competitors
    for_each: "{{input_data.competitors}}"
    item_name: competitor
    servers: [brave-search]
    prompt: |
      Research competitor: {{competitor}}
      
      Analyze:
      - Product offerings
      - Pricing strategy
      - Customer segments
      - Strengths and weaknesses
      - Market positioning
    output: competitor_analysis

  # Customer insights
  - name: customer_research
    servers: [brave-search]
    prompt: |
      Research customer needs for:
      {{input_data.product_category}}
      
      Find:
      - Common pain points
      - Buying criteria
      - Price sensitivity
      - Feature priorities
    output: customer_insights

  # Synthesis
  - name: market_report
    prompt: |
      Create market analysis report:
      
      Industry: {{industry_analysis}}
      Competitors: {{competitor_analysis}}
      Customers: {{customer_insights}}
      
      Include:
      - Market opportunity assessment
      - Competitive landscape
      - Customer needs analysis
      - Go-to-market recommendations
      - Risk factors
```

**Usage:**
```bash
mcp-cli --template market_research --input-data '{
  "industry": "project management software",
  "competitors": ["Asana", "Monday.com", "ClickUp"],
  "product_category": "team collaboration tools"
}'
```

**Output includes:**
- TAM: $6.8B, growing at 12% annually
- 3 competitor deep-dives with pricing and features
- Customer needs: simplicity > features, mobile crucial
- Recommendation: Focus on ease-of-use for small teams

**Cost:** ~$2.50 (comprehensive multi-competitor research)
**Time:** ~12 minutes

---

### Example 3: Academic Literature Review

**Scenario:** Understanding current research on a topic.

```yaml
name: academic_research
version: 1.0.0

steps:
  # Literature search
  - name: find_papers
    servers: [academic-search]
    prompt: |
      Find recent papers (last 3 years) on:
      {{input_data.topic}}
      
      Prioritize:
      - High citation count
      - Peer-reviewed
      - Recent publication
      - Systematic reviews or meta-analyses
    output: papers

  # Analyze methodology
  - name: analyze_methods
    for_each: "{{papers}}"
    item_name: paper
    prompt: |
      Summarize paper:
      {{paper}}
      
      Extract:
      - Research question
      - Methodology
      - Sample size and demographics
      - Key findings
      - Limitations
      - Conclusion
    output: paper_summaries

  # Identify trends
  - name: identify_trends
    prompt: |
      Analyze trends across papers:
      {{paper_summaries}}
      
      Find:
      - Common methodologies
      - Consistent findings (what most studies agree on)
      - Contradictions (where studies disagree)
      - Evolution of thought (how understanding changed)
      - Current consensus
    output: research_trends

  # Research gaps
  - name: find_gaps
    prompt: |
      Identify research gaps:
      {{paper_summaries}}
      {{research_trends}}
      
      What hasn't been studied yet?
      What questions remain unanswered?
      Where are there contradictions needing resolution?
    output: research_gaps
```

**Usage:**
```bash
mcp-cli --template academic_research --input-data '{
  "topic": "effectiveness of spaced repetition for language learning"
}'
```

**Output:**
- 15 papers analyzed
- **Consensus:** Spaced repetition improves retention by 50-200%
- **Optimal intervals:** Expanding intervals more effective than fixed
- **Contradiction:** Debate on optimal initial interval (1 day vs 1 week)
- **Gap identified:** Limited research on advanced learners

**Cost:** ~$1.80 (15 papers + analysis)
**Time:** ~10 minutes

---

## Best Practices

### 1. Start Broad, Then Deep

**Why:** You don't know what's important until you see the landscape.

```yaml
# Good: Hierarchical research
steps:
  - name: overview          # Broad scan (what exists?)
    output: landscape
  
  - name: identify_focus    # Find what matters
    prompt: "From {{landscape}}, what's most important?"
    output: key_areas
  
  - name: deep_dive         # Go deep on key areas
    for_each: "{{key_areas}}"

# Bad: Too specific too soon
steps:
  - name: very_specific_question
    prompt: "Research X's performance under Y conditions with Z constraints"
    # Might miss that X is deprecated, Y is obsolete, Z doesn't matter
```

**Example:**
- ✅ Good: "Research JavaScript frameworks" → identifies React/Vue/Angular are top 3 → deep dives on those
- ❌ Bad: "Research Ember.js performance" → misses that Ember usage declining, React/Vue more relevant

---

### 2. Use Follow-up Questions

**Why:** First pass finds obvious info, follow-ups find the nuances.

```yaml
# Good: Iterative discovery
steps:
  - name: initial_research
    output: findings
  
  - name: identify_gaps
    prompt: "What's missing from {{findings}}?"
    output: gaps
  
  - name: follow_up
    condition: "{{gaps}} not empty"
    for_each: "{{gaps}}"
    # Fills the gaps

# Bad: One-shot research
steps:
  - name: single_research
    prompt: "Research everything about X"
    # Likely to miss important aspects
```

**Real example:**
- Research "best database for e-commerce"
- Initial: Finds Postgres, MySQL, MongoDB
- Gap identified: "What about high availability?"
- Follow-up: Discovers clustering, replication strategies
- Better final recommendation

---

### 3. Cross-Reference Sources

**Why:** Single source might be biased, outdated, or wrong.

```yaml
# Good: Multiple independent sources
parallel:
  - servers: [brave-search]     # Web perspectives
  - servers: [academic-search]  # Academic rigor
  - servers: [news-api]         # Recent developments
aggregate: merge

- name: cross_reference
  prompt: "Where do sources agree? Disagree?"

# Bad: Single source
- servers: [brave-search]
  # What if top results are sponsored content or outdated?
```

**Confidence levels:**
- All 3 sources agree: **HIGH CONFIDENCE**
- 2 of 3 agree: **MEDIUM CONFIDENCE**
- All disagree: **LOW CONFIDENCE - needs more research**

---

### 4. Handle Contradictions Explicitly

**Why:** Contradictions are valuable information, not problems.

```yaml
# Good: Identify and explain contradictions
- name: find_contradictions
  prompt: |
    Find where sources disagree:
    {{findings}}

    For each contradiction:
    - Source A says: [position]
    - Source B says: [different position]
    - Possible reasons: [why they differ]
    - Current consensus: [if any]
    - Recommendation: [what to believe or do]

# Bad: Ignore contradictions
- name: just_summarize
  prompt: "Summarize: {{findings}}"
  # Hides disagreements, gives false confidence
```

**Example contradiction:**
- Web articles say: "NoSQL is faster than SQL"
- Academic papers say: "Depends on workload characteristics"
- Explanation: Web articles oversimplify, reality is nuanced
- Recommendation: Choose based on specific use case, not blanket statement

---

### 5. Track Confidence Levels

**Why:** Not all information is equally reliable.

```yaml
# Good: Explicit confidence tracking
- prompt: |
    For each finding, rate confidence:
    
    HIGH: Multiple credible sources agree, recent data, expert consensus
    MEDIUM: Single credible source or older data
    LOW: Anecdotal, contradictory, or unverified
    
    Format: [Finding] (Confidence: HIGH/MEDIUM/LOW)

# Bad: All findings treated equally
- prompt: "List findings"
  # User doesn't know what to trust
```

**Confidence indicators:**
- ✅ **HIGH:** Peer-reviewed studies, official documentation, multiple sources
- ⚠️ **MEDIUM:** Single expert opinion, company blog, older data
- ❌ **LOW:** Forum posts, unverified claims, contradicted by others

---

### 6. Set Depth Limits

**Why:** Research can spiral infinitely, eating time and money.

```yaml
# Good: Bounded research
config:
  variables:
    max_follow_ups: 3      # Stop after 3 iterations
    max_sub_questions: 5   # Don't decompose into 20 questions

steps:
  - name: follow_up
    condition: "{{iteration}} < {{max_follow_ups}}"
    # Prevents infinite loops

# Bad: Unbounded research
- name: follow_up
  condition: "{{gaps}} not empty"
  # Could iterate forever, costing $$$
```

**Depth levels:**
- **Quick:** 1 iteration, 3 sub-questions = ~$0.50, 2 minutes
- **Standard:** 2 iterations, 5 sub-questions = ~$1.50, 5 minutes  
- **Deep:** 3 iterations, 7 sub-questions = ~$3.00, 10 minutes
- **Exhaustive:** 4+ iterations = $5+, 15+ minutes (rarely needed)

---

### 7. Cache and Reuse

**Why:** Don't repeat expensive searches.

```yaml
# Good: Cache initial research
- name: baseline_research
  output: cached_baseline
  
# Reuse for multiple analyses
- name: analysis_1
  prompt: "Analyze for security: {{cached_baseline}}"
  
- name: analysis_2  
  prompt: "Analyze for performance: {{cached_baseline}}"
  # Uses same research, different lens

# Bad: Research twice
- name: research_for_security
- name: research_for_performance  # Repeats searches unnecessarily
```

---

### 8. Validate Critical Claims

**Why:** Research findings used for decisions should be verified.

```yaml
# Good: Validation for high-stakes
- name: research
  output: findings
  
- name: extract_critical_claims
  prompt: "Which findings are critical to our decision?"
  
- name: validate_critical
  condition: "{{critical}} true"
  template: multi_provider_validation
  # Use multi-provider validation pattern for critical claims

# Bad: No validation
- name: research
- name: make_decision  # Based on unvalidated research
```

---

## Cost Optimization

Research agents can be expensive. Here's how to control costs:

### 1. Set Research Depth Based on Stakes

```yaml
# For high-stakes decisions ($10K+)
config:
  variables:
    max_iterations: 3
    max_sources: 10
# Cost: ~$3-5, Worth it

# For medium-stakes decisions ($1K-10K)
config:
  variables:
    max_iterations: 2
    max_sources: 5
# Cost: ~$1-2, Reasonable

# For low-stakes decisions (<$1K)
config:
  variables:
    max_iterations: 1
    max_sources: 3
# Cost: ~$0.30-0.50, Appropriate
```

### 2. Use Conditional Deep Research

```yaml
# Only deep dive if initial research is insufficient
steps:
  - name: quick_research
    output: quick_findings
  
  - name: assess_sufficiency
    prompt: "Is {{quick_findings}} enough for {{requirements}}?"
    output: is_sufficient
  
  - name: deep_research
    condition: "{{is_sufficient}} contains 'no'"
    # Only runs if needed, saves money
```

**Savings:** 60-70% of queries answered by quick research alone.

### 3. Parallel vs Sequential

```yaml
# Fast but expensive: All searches in parallel
parallel:
  - search1
  - search2
  - search3
# All run even if first answers question

# Slow but cheap: Sequential with early exit
- search1
  output: result1
- condition: "{{result1}} insufficient"
  search2
  # Only runs if needed
```

**Trade-off:** Speed vs cost. Use parallel for time-sensitive, sequential for budget-sensitive.

### 4. Cache Common Research

```yaml
# Cache frequently researched topics
- name: check_cache
  servers: [redis]
  prompt: "Check if {{topic}} researched in last 7 days"
  
- name: use_cached
  condition: "{{cache}} exists"
  prompt: "Use cached: {{cache}}"
  
- name: fresh_research
  condition: "{{cache}} empty"
  # Only research if cache miss
```

**Savings:** Eliminate redundant research, 80-90% cost reduction for repeat topics.

### 5. Tiered Research Strategy

```yaml
# Tier 1: Free local model for common questions
- name: try_local
  provider: ollama
  output: local_answer

# Tier 2: Single search if local insufficient  
- condition: "{{local_answer.confidence}} < 0.7"
  servers: [brave-search]
  output: search_answer

# Tier 3: Full research agent only if critical
- condition: "{{stakes}} == 'high' AND {{search_answer}} insufficient"
  template: full_research_agent
```

**Cost ladder:**
- Tier 1: $0 (free local)
- Tier 2: $0.03 (one search)
- Tier 3: $1-5 (full research)

Most queries (80%) resolved at Tier 1 or 2.

---

## Performance Optimization

### 1. Parallel Research (Faster)

```yaml
# Parallel: All sub-questions researched at once
- name: research_all
  parallel:
    - servers: [brave-search]
      for_each: "{{sub_questions}}"
  max_concurrent: 5

# Time: max(search1, search2, search3, search4, search5)
# If each search takes 10s: Total = 10s
```

vs

```yaml
# Sequential: One after another
- for_each: "{{sub_questions}}"
  servers: [brave-search]

# Time: search1 + search2 + search3 + search4 + search5
# If each search takes 10s: Total = 50s
```

**Speedup:** 5x faster with parallel!

**Trade-off:** All searches run even if first few answer the question. Pay for speed.

### 2. Smart Caching

```yaml
# Cache expensive computations
- name: baseline_research
  servers: [brave-search, academic-search]
  output: research_cache  # Store for reuse

# Multiple analyses use same cache
- name: security_analysis
  prompt: "Security implications: {{research_cache}}"
  
- name: cost_analysis
  prompt: "Cost implications: {{research_cache}}"
  
# Saves: 2 searches by reusing research_cache
```

### 3. Incremental Deepening (Efficient)

```yaml
# Stop when you have enough
- name: quick_overview
  output: overview
  
- name: check_if_sufficient
  prompt: "Is {{overview}} sufficient?"
  output: sufficient
  
# Only continue if needed
- name: deeper_research
  condition: "{{sufficient}} contains 'no'"
  
- name: even_deeper
  condition: "{{deeper_research.confidence}} < 0.8"
```

**Efficiency:** Only goes as deep as needed, doesn't waste effort.

### 4. Batch Processing

```yaml
# Process multiple topics efficiently
- name: research_all_topics
  for_each: "{{topics}}"
  parallel: true
  max_concurrent: 3
  servers: [brave-search]
  
# Processes 9 topics in same time as 3 topics
# With max_concurrent=3: 
# - Batch 1: topics 1,2,3 (parallel)
# - Batch 2: topics 4,5,6 (parallel)
# - Batch 3: topics 7,8,9 (parallel)
```

**Performance comparison:**
- Sequential: 9 topics × 10s = 90 seconds
- Parallel (max 3): 3 batches × 10s = 30 seconds
- **3x faster**

---

## Integration with MCP Servers

### Web Search

```yaml
- servers: [brave-search]
  prompt: "Search for: {{query}}"
```

### Academic Databases

```yaml
- servers: [academic-search]
  prompt: "Find papers on: {{topic}}"
```

### News APIs

```yaml
- servers: [news-api]
  prompt: "Recent news on: {{topic}}"
```

### Social Media

```yaml
- servers: [social-media]
  prompt: "Public sentiment on: {{topic}}"
```

---

## Error Handling

```yaml
steps:
  - name: search
    servers: [brave-search]
    prompt: "Research: {{topic}}"
    error_handling:
      on_failure: continue
      default_output: "SEARCH_UNAVAILABLE"

  - name: fallback
    condition: "{{search}} contains 'UNAVAILABLE'"
    prompt: "Use general knowledge on: {{topic}}"
```

---

## Complete Example

```yaml
name: comprehensive_research_agent
version: 1.0.0

config:
  defaults:
    provider: anthropic
    model: claude-sonnet-4

steps:
  # 1. Decompose question
  - name: decompose
    prompt: "Break down: {{question}}"
    output: sub_questions

  # 2. Parallel research
  - name: research_all
    parallel:
      - name: web
        servers: [brave-search]
        for_each: "{{sub_questions}}"
        prompt: "Web search: {{item}}"
        output: web_results

      - name: academic
        servers: [academic-search]
        for_each: "{{sub_questions}}"
        prompt: "Academic search: {{item}}"
        output: academic_results
    max_concurrent: 2
    aggregate: merge

  # 3. Identify gaps
  - name: gaps
    prompt: "What's missing? {{research_all}}"
    output: knowledge_gaps

  # 4. Follow-up
  - name: follow_up
    condition: "{{knowledge_gaps}} not empty"
    for_each: "{{knowledge_gaps}}"
    servers: [brave-search]
    prompt: "Deep dive: {{item}}"
    output: follow_up_findings

  # 5. Cross-reference
  - name: verify
    prompt: |
      Cross-reference all findings.
      Mark confidence levels.
      {{research_all}}
      {{follow_up_findings}}
    output: verified

  # 6. Report
  - name: report
    prompt: "Create comprehensive report: {{verified}}"
```

---

## Quick Reference

```yaml
# Basic research
decompose → research → synthesize

# Iterative research
overview → deep_dive → follow_up → report

# Multi-source research
parallel_search → cross_reference → verify → report

# Comparative research
research_each → compare → recommend
```

---

## Next Steps

- **[Data Pipeline Pattern](data-pipeline.md)** - ETL workflows
- **[Validation Pattern](validation.md)** - Multi-provider verification
- **[Examples](../examples/)** - Working templates

---

**Build powerful research agents!** 🔍
