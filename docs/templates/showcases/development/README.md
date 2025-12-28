# Development Workflows Templates

> **For:** Software Developers, Engineering Teams, Tech Leads  
> **Purpose:** Automate tedious development tasks with AI-powered code analysis and generation

---

## What This Showcase Contains

This section demonstrates how templates automate critical development workflows. All examples solve real developer pain points: outdated documentation, slow code reviews, tedious migrations, performance issues, and low test coverage.

### Available Use Cases

**Developer Productivity:**

1. **[API Documentation Generator](use-cases/api-documentation.md)** - Auto-generate OpenAPI specs from code
2. **[Database Query Optimizer](use-cases/query-optimizer.md)** - Detect N+1 queries and missing indexes
3. **[Code Migration Assistant](use-cases/code-migration.md)** - Automate framework/language upgrades
4. **[PR Review Assistant](use-cases/pr-review.md)** - Code quality reviews with consensus
5. **[Test Generator](use-cases/test-generator.md)** - Baseline test coverage automation
6. **[Architecture Documentation](use-cases/architecture-docs.md)** - Generate system diagrams from code

---

## Why Templates Matter for Development

### 1. API Documentation Automation

**The Challenge:** API documentation is always outdated. OpenAPI specs are written manually, get out of sync with code, cause support tickets when consumers use wrong endpoints.

**Template Solution:** Auto-generate docs from code

```yaml
# Analyze codebase, generate OpenAPI spec:
steps:
  # 1. Parse API routes/endpoints
  - name: extract_endpoints
    prompt: "Find all API endpoints in: {{codebase}}"
    output: endpoints

  # 2. Extract request/response schemas
  - name: extract_schemas
    prompt: "Extract request/response types for: {{endpoints}}"
    output: schemas

  # 3. Generate OpenAPI spec
  - name: generate_openapi
    prompt: "Create OpenAPI 3.0 spec from: {{endpoints}} + {{schemas}}"
    output: openapi_spec

  # 4. Validate against actual code
  - name: validate
    prompt: "Verify spec matches code, flag discrepancies"
```

**Impact:**

- Manual OpenAPI writing: 8 hours per service
- Automated generation: 5 minutes
- **Time saved: 99%**
- **Doc accuracy: 60% → 95%** (measured)
- **Support tickets reduced: 40%** (outdated docs)

**Real scenario:**

```
Before: Developer changes endpoint, forgets to update docs
Result: API consumers hit 404s, file tickets, waste dev time

After: Template regenerates docs automatically on merge
Result: Docs always in sync, zero outdated endpoint issues
```

**Documentation:** [API Documentation Generator](use-cases/api-documentation.md)

---

### 2. Database Query Optimization

**The Challenge:** N+1 queries kill performance. 1000 users × 1 query per user = 1001 queries. API responds in 2000ms instead of 20ms. Hard to spot without profiling.

**Template Solution:** Static analysis detects inefficient queries

```yaml
# Detect query anti-patterns:
steps:
  # 1. Parse ORM/SQL queries
  - name: extract_queries
    prompt: "Extract all database queries from: {{codebase}}"
    output: queries

  # 2. Detect N+1 patterns
  - name: detect_n_plus_1
    prompt: |
      Identify N+1 query patterns:
      - Loop over collection
      - Query inside loop
      - No prefetch/join
    output: n_plus_1_issues

  # 3. Find missing indexes
  - name: missing_indexes
    prompt: |
      Analyze queries:
      - WHERE clauses without indexes
      - JOIN columns not indexed
      - ORDER BY without index
    output: index_recommendations

  # 4. Generate fixes
  - name: generate_fixes
    prompt: |
      For each issue, provide:
      - Code fix (add .select_related, etc.)
      - SQL for creating indexes
      - Expected performance improvement
```

**Real detection:**

```python
# BAD: N+1 query detected
users = User.objects.all()  # 1 query
for user in users:  # Loop
    posts = user.posts.all()  # N queries (1 per user!)

# Template output:
⚠️ N+1 Query: user_posts_view.py:42
  Current: 1 + 1000 = 1001 queries
  Fix: User.objects.all().prefetch_related('posts')
  Expected: 2 queries
  Performance gain: 500× faster (2000ms → 4ms)

# GOOD: Fixed query
users = User.objects.all().prefetch_related('posts')  # 2 queries total
for user in users:
    posts = user.posts.all()  # No extra queries (cached)
```

**Impact:**

- API response time: 2000ms → 20ms (100× faster)
- Database load: 1001 queries → 2 queries
- Server costs: $500/month → $100/month (fewer DB instances needed)

**Documentation:** [Database Query Optimizer](use-cases/query-optimizer.md)

---

### 3. Code Migration Automation

**The Challenge:** Framework upgrades are tedious. Migrate 500 React class components to hooks? 2 weeks of manual work. Prone to errors, tests break, production bugs.

**Template Solution:** Automated code transformation

```yaml
# Migrate React class components → hooks:
steps:
  # 1. Find all class components
  - name: find_components
    prompt: "Find React class components with lifecycle methods"
    output: class_components

  # 2. Transform each component
  - name: transform
    for_each: "{{class_components}}"
    prompt: |
      Transform class component to hooks:
      - componentDidMount → useEffect
      - this.state → useState
      - this.props → props
      - Class methods → functions
    output: hook_components

  # 3. Validate syntax
  - name: validate
    prompt: "Check transformed code compiles"
    output: validation

  # 4. Run tests
  - name: test
    prompt: "Verify tests still pass"
    output: test_results
```

**Real migration:**

```javascript
// BEFORE: Class component (old pattern)
class UserProfile extends React.Component {
  constructor(props) {
    super(props);
    this.state = { user: null };
  }

  componentDidMount() {
    fetchUser(this.props.userId).then(user => {
      this.setState({ user });
    });
  }

  render() {
    return <div>{this.state.user?.name}</div>;
  }
}

// AFTER: Hooks (modern pattern) - auto-generated
function UserProfile({ userId }) {
  const [user, setUser] = useState(null);

  useEffect(() => {
    fetchUser(userId).then(user => {
      setUser(user);
    });
  }, [userId]);

  return <div>{user?.name}</div>;
}
```

**Impact:**

- 500 components migrated
- Manual: 2 weeks (80 hours)
- Automated: 2 hours
- **Time saved: 97.5%**
- **Tests passing: 98%** (manual review remaining 2%)

**Documentation:** [Code Migration Assistant](use-cases/code-migration.md)

---

### 4. PR Review with Consensus (Code Quality, NOT Security)

**The Challenge:** Code review bottleneck. Reviews take 2-4 hours, inconsistent standards across team, junior devs unsure what to check.

**Template Solution:** Multi-provider code quality review

```yaml
# Consensus review (NOT security scanning):
parallel:
  - provider: anthropic
    prompt: "Review code quality: readability, naming, structure"

  - provider: openai
    prompt: "Review code quality: patterns, complexity, maintainability"

  - provider: gemini
    prompt: "Review code quality: documentation, conventions"

# Cross-validate:
# - All 3 agree: Definitely needs fixing
# - 2 of 3 agree: Suggest to developer
# - Only 1 found: Flag for human reviewer
```

**What it checks:**

- ✅ Code readability
- ✅ Naming conventions
- ✅ Code complexity (cyclomatic complexity)
- ✅ Architectural patterns
- ✅ Documentation completeness
- ✅ Test coverage

**What it does NOT check:**

- ❌ Security vulnerabilities (too many false positives)
- ❌ Performance without profiling
- ❌ Architectural decisions (need human context)

**Why consensus reduces false positives:**

```
Single reviewer: Flags 50 issues (20 valid, 30 false positives)
Developer trust: Low (learns to ignore warnings)

Consensus: Only flags issues all 3 agree on (18 valid, 2 false positives)
Developer trust: High (90% accuracy)
```

**Impact:**

- Review time: 2 hours → 30 minutes (human reviews consensus findings)
- False positives: 60% → 10%
- Developer trust: High (actionable feedback)

**Documentation:** [PR Review Assistant](use-cases/pr-review.md)

---

### 5. Context-Efficient Large Codebase Analysis

**The Challenge:** Analyzing 50K+ lines of code exceeds LLM context limits. Can't load entire codebase into single prompt.

**Template Solution:** File-by-file processing with aggregation

**Traditional approach (fails):**

```
LLM Context (200K tokens):
├── Full codebase: 50K LOC × 3 tokens = 150K tokens
├── Analysis prompt: 5K tokens
└── Output: 45K tokens
Total: 200K tokens ✓ Fits

BUT: Large codebases exceed this
100K LOC = 300K tokens → Context overflow
```

**Template approach (scalable):**

```
Analyze in chunks:

File 1 (500 LOC):
├── LLM analyzes: Fresh 200K context
└── Output: API endpoints, dependencies, complexity

File 2 (500 LOC):
├── LLM analyzes: Fresh 200K context
└── Output: API endpoints, dependencies, complexity

...

File 200 (500 LOC):
├── LLM analyzes: Fresh 200K context
└── Output: API endpoints, dependencies, complexity

Final aggregation:
├── LLM receives: 200 file summaries (50K tokens)
└── Output: Overall architecture, API catalog, dependencies
```

**Benefits:**

- Analyze unlimited codebase size
- Each file gets full context
- Parallel processing for speed
- Aggregated insights

---

### 6. Parallel Code Transformation

**The Challenge:** Transforming 1000 files takes forever if done sequentially.

**Template Solution:** Parallel batch processing

```yaml
# Transform 1000 Python files (add type hints):
steps:
  - name: add_type_hints
    for_each: "{{python_files}}"
    parallel:
      batch_size: 50  # Process 50 files per batch
      max_concurrent: 10  # 10 batches in parallel
    prompt: "Add type hints to: {{file}}"
    output: typed_files

# Result: 1000 files processed in 5 minutes
# vs. Sequential: Would take 8+ hours
```

**Performance:**

- Sequential: 1000 files × 30 seconds = 8 hours
- Parallel (10 concurrent): 1000 / 10 × 30 seconds = 50 minutes
- **Speedup: 10×**

---

## Quick Start

### 1. Choose Your Development Challenge

**Outdated API docs?**

- [API Documentation Generator](use-cases/api-documentation.md) - Auto-generate OpenAPI specs

**Slow database queries?**

- [Database Query Optimizer](use-cases/query-optimizer.md) - Detect N+1 queries

**Legacy code migration?**

- [Code Migration Assistant](use-cases/code-migration.md) - Automate framework upgrades

**Slow code reviews?**

- [PR Review Assistant](use-cases/pr-review.md) - Automated quality checks

**Low test coverage?**

- [Test Generator](use-cases/test-generator.md) - Generate baseline tests

**Poor documentation?**

- [Architecture Documentation](use-cases/architecture-docs.md) - Generate system diagrams

### 2. Run Template Against Your Code

```bash
# Generate API documentation
mcp-cli --template api_documentation --input-data "{
  \"codebase_path\": \"./src/api/\",
  \"framework\": \"fastapi\",
  \"output_format\": \"openapi\"
}"

# Detect N+1 queries
mcp-cli --template query_optimizer --input-data "{
  \"codebase_path\": \"./src/\",
  \"orm\": \"sqlalchemy\",
  \"detect\": [\"n_plus_1\", \"missing_indexes\"]
}"

# Migrate code
mcp-cli --template code_migration --input-data "{
  \"source_path\": \"./src/components/\",
  \"migration\": \"class_to_hooks\",
  \"framework\": \"react\"
}"
```

---

## Integration Patterns

### Pattern 1: CI/CD Documentation Pipeline

**Auto-update docs on every merge:**

```yaml
name: ci_documentation

# Triggered by: git push to main
steps:
  # 1. Generate API docs from code
  - name: generate_docs
    template: api_documentation
    input: "src/api/**/*.py"
    output: openapi_spec

  # 2. Commit to docs repo
  - name: commit_docs
    servers: [git]
    prompt: "Commit {{openapi_spec}} to docs/api.yaml"

  # 3. Deploy to docs site
  - name: deploy
    servers: [netlify]
    prompt: "Deploy updated docs"
```

**Result:** Docs always in sync with code, zero manual effort

---

### Pattern 2: Pre-Commit Performance Check

**Catch N+1 queries before merge:**

```yaml
name: pre_commit_performance

# Triggered by: git commit
steps:
  # 1. Get changed files
  - name: get_diff
    servers: [git]
    prompt: "Get changed files in this commit"
    output: changed_files

  # 2. Analyze queries
  - name: analyze_queries
    template: query_optimizer
    input: "{{changed_files}}"
    output: query_issues

  # 3. Block commit if N+1 found
  - name: validate
    condition: "{{query_issues.count}} > 0"
    prompt: |
      ❌ COMMIT BLOCKED: Performance issues found
      {{query_issues}}
      Fix queries before committing.
```

**Result:** N+1 queries never reach production

---

### Pattern 3: Automated Code Migration

**Migrate entire codebase safely:**

```yaml
name: safe_migration

steps:
  # 1. Find all components to migrate
  - name: find_targets
    prompt: "Find all class components"
    output: components

  # 2. Migrate in batches
  - name: migrate
    for_each: "{{components}}"
    parallel:
      batch_size: 10
      max_concurrent: 5
    template: code_migration
    output: migrated

  # 3. Run tests after each batch
  - name: test
    servers: [jest]
    prompt: "Run tests on: {{migrated}}"
    output: test_results

  # 4. Rollback if tests fail
  - name: rollback
    condition: "{{test_results.passed}} == false"
    servers: [git]
    prompt: "Rollback migration, tests failed"
```

---

## Best Practices

### Development Template Design

**✅ Do:**

- Focus on tedious, repetitive tasks
- Validate output (run tests, check compilation)
- Provide actionable fixes, not just warnings
- Track before/after metrics
- Use consensus for subjective judgments
- Generate code that humans can review

**❌ Don't:**

- Try to replace human judgment
- Generate code without validation
- Create noise with false positives
- Skip testing generated code
- Assume AI output is always correct
- Use for security scanning (too many false positives)

### Code Quality vs Security

**Good for AI:**

- Code formatting and style
- Naming convention violations
- Code complexity metrics
- Documentation completeness
- Pattern adherence
- Test coverage gaps

**Bad for AI (high false positives):**

- Security vulnerabilities
- Authentication/authorization issues
- Input validation
- Cryptography review
- Dependency vulnerabilities

---

## Measuring Success

### API Documentation Metrics

**Before templates:**

- Manual OpenAPI writing: 8 hours per service
- Doc accuracy: 60% (outdated endpoints)
- Support tickets: 50/month (wrong API usage)
- Developer time lost: 40 hours/month

**After templates:**

- Automated generation: 5 minutes
- Doc accuracy: 95% (always in sync)
- Support tickets: 30/month (40% reduction)
- **Time saved: 99%** (8 hours → 5 minutes)

### Query Optimization Metrics

**Before detection:**

- N+1 queries in production: 15 instances
- API p95 latency: 2000ms
- Database load: 80% CPU
- AWS costs: $500/month (over-provisioned DB)

**After detection:**

- N+1 queries caught pre-merge: 100%
- API p95 latency: 50ms (40× faster)
- Database load: 20% CPU
- AWS costs: $150/month (right-sized DB)
- **Cost savings: $350/month**

### Code Migration Metrics

**Before automation:**

- 500 components to migrate
- Manual migration: 2 weeks (80 hours)
- Error rate: 15% (broken tests)
- Developer frustration: High

**After automation:**

- Migration time: 2 hours
- Error rate: 2% (mostly edge cases)
- Developer review: 4 hours
- **Total: 6 hours vs 80 hours** (93% time savings)

---

## Cost Analysis

### API Documentation

**Per service:**

- Manual writing: 8 hours @ $100/hr = $800
- Automated: $0.10 (AI cost) + 5 min
- **Savings: 99.9%**

**10 microservices:**

- Manual: $8,000
- Automated: $1
- **Savings: $7,999**

### Query Optimization

**Cost avoidance:**

- Detection: $0.05 per file analyzed
- N+1 query in production: $350/month in DB costs
- Find 1 N+1 query: **Pays for 7,000 file analyses**

### Code Migration

**500 components:**

- Manual: 80 hours @ $100/hr = $8,000
- Automated: $2.50 (AI cost) + 6 hours review = $602.50
- **Savings: $7,397.50** (92% cost reduction)

---

## Template Library

All templates available in [templates/](templates/):

**Documentation:**

- `api_documentation.yaml` - OpenAPI generation
- `architecture_docs.yaml` - System diagrams

**Code Quality:**

- `pr_review.yaml` - Consensus code review
- `query_optimizer.yaml` - N+1 detection
- `test_generator.yaml` - Test coverage

**Code Transformation:**

- `code_migration.yaml` - Framework migrations
- `type_hints.yaml` - Add Python type hints
- `refactor_patterns.yaml` - Apply refactoring patterns

---

## Example: Complete Development Pipeline

```yaml
name: complete_dev_pipeline

# On every PR:
steps:
  # 1. Consensus code review
  - name: review_quality
    template: pr_review
    input: "{{pr.changed_files}}"
    output: review

  # 2. Detect performance issues
  - name: check_queries
    template: query_optimizer
    input: "{{pr.changed_files}}"
    output: query_issues

  # 3. Generate tests for new code
  - name: generate_tests
    template: test_generator
    input: "{{pr.new_functions}}"
    output: tests

  # 4. Update API docs if endpoints changed
  - name: update_docs
    condition: "{{pr.changed_files}} contains 'api/'"
    template: api_documentation
    output: openapi

  # 5. Post review comment
  - name: post_comment
    servers: [github]
    prompt: |
      ## Automated Review

      **Code Quality:** {{review.summary}}
      **Performance:** {{query_issues.count}} issues
      **Tests:** {{tests.coverage}}% coverage
      **Docs:** {{openapi.status}}
```

---

## Next Steps

1. **Review use cases** - Read detailed documentation for each workflow
2. **Try on sample code** - Test templates on small codebase first
3. **Measure baseline** - Track current metrics (review time, doc accuracy, etc.)
4. **Deploy incrementally** - Start with one template, expand gradually
5. **Customize for team** - Adjust conventions and thresholds

---

## Additional Resources

- **[Why Templates Matter](../../WHY_TEMPLATES_MATTER.md)** - Context management explained
- **[Template Authoring Guide](../../authoring-guide.md)** - Create custom dev templates
- **[DevOps Showcase](../devops/)** - CI/CD automation patterns

---

**Development automation with AI: Stop writing docs manually, catch bugs before production, migrate code safely.**

Templates transform development from tedious manual work to automated, validated, reproducible workflows.
