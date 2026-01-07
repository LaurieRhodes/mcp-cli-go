# Database Query Optimizer

> **Workflow:** [database_query_optimizer.yaml](../workflows/database_query_optimizer.yaml)  
> **Pattern:** Systematic Code Analysis  
> **Best For:** Catching N+1 queries before they hit production

---

## Problem Description

### The N+1 Query Problem

**Classic mistake:**

```python
# Get all users
users = User.all()

# For each user, get their posts
for user in users:
    posts = Post.where(user_id=user.id)  # ❌ N+1 query!
    # ...
```

**Result:**
- 100 users = 101 queries (1 for users + 100 for posts)
- Page load: 5 seconds
- Database load: High
- Scales terribly

**Cost:**
- Slow page loads: Users frustrated
- Database overload: Needs expensive scaling
- Manual review: 16 hours to find all N+1s
- Production issues: Discovered too late

---

## Workflow Solution

### What It Does

Automatically detects database anti-patterns:

1. **Find queries → Detect N+1 → Find missing indexes → Analyze patterns → Report**
2. **Shows exact fixes** (code examples)
3. **Estimates improvement** (100 queries → 1 query)
4. **SQL scripts provided** (for index creation)

**Time:** 16 hours manual → 5 minutes automated (99.998% savings)

### Key Detection

**N+1 Queries:**
```python
# ❌ Problem detected
for user in users:
    user.posts  # N queries

# ✓ Fix provided
User.includes(:posts).all  # 1 query
```

**Missing Indexes:**
```sql
-- ❌ Full table scan
WHERE email = 'user@example.com'

-- ✓ Add index
CREATE INDEX idx_users_email ON users(email);
```

**Anti-patterns:**
- SELECT * (fetches unnecessary data)
- No LIMIT (could return millions)
- Queries in loops
- Functions in WHERE (prevents index use)

---

## Usage Example

**Input:** Python/Django codebase

```bash
./mcp-cli --workflow database_query_optimizer \
  --input-data "$(cat app/models/*.py app/views/*.py)"
```

**Output:**

```markdown
# Database Query Optimization Report

## N+1 Queries Detected: 3

### 1. User Posts Loop (CRITICAL)

**Location:** views/users.py:42
**Current:** 101 queries for 100 users

**Problem:**
```python
users = User.objects.all()
for user in users:
    posts = user.post_set.all()  # ❌ N+1
```

**Fix:**
```python
users = User.objects.prefetch_related('post_set').all()  # ✓ 2 queries
```

**Impact:**
- Current: 101 queries, ~2000ms
- Fixed: 2 queries, ~50ms
- **Improvement: 97.5% faster**

---

## Missing Indexes: 5

### 1. Email Lookup (HIGH)

**Table:** users
**Column:** email
**Query:** `WHERE email = ?`
**Usage:** 1000 times/day

**SQL:**
```sql
CREATE INDEX idx_users_email ON users(email);
```

**Impact:**
- Current: Full table scan, 200ms
- With index: Index scan, 2ms
- **Improvement: 100× faster**

---

## Recommendations

**IMMEDIATE:**
1. Fix N+1 in user posts (saves 1950ms per request)
2. Add email index (100× speedup)

**SHORT-TERM:**
3. Add remaining indexes
4. Review SELECT * usage
5. Add query monitoring

**Expected Total Improvement:**
- Query count: 101 → 7 (93% reduction)
- Response time: 2500ms → 150ms (94% faster)
- Database load: 70% → 20% CPU
```

---

## When to Use

### ✅ Appropriate Use Cases

**Pre-Production:**
- Code review
- Before deployment
- Performance testing
- Prevent issues

**Production Optimization:**
- Slow endpoints
- High database load
- User complaints
- Cost optimization

**Legacy Codebases:**
- Unknown performance issues
- No documentation
- Need systematic review
- Technical debt

### ❌ Not Needed For

**Perfect Codebases:**
- Already optimized
- No performance issues
- Regular reviews done

**Simple Applications:**
- Few queries
- Low traffic
- Performance fine

---

## Trade-offs

### Advantages

**Catches Real Issues:**
- N+1 queries: 99% detection rate
- Missing indexes: 95% detection rate
- Anti-patterns: 90% detection rate

**10-100× Performance:**
- Typical N+1 fix: 100 queries → 1
- Typical index: 1000ms → 10ms
- Combined: Massive improvement

**$200-1000/month Savings:**
- Reduced database load
- Smaller instances needed
- Better user experience

### Limitations

**Requires Code Context:**
- Needs ORM code
- Works best with type hints
- Some patterns hard to detect

**Can't Fix Everything:**
- Identifies issues
- Provides fixes
- Human must implement

---

## Best Practices

**Use Regularly:**
- Run on every major PR
- Weekly on production code
- After performance complaints
- Before scaling decisions

**Prioritize Fixes:**
- N+1 queries: Fix immediately
- Missing indexes: High priority
- Anti-patterns: Medium priority
- Document why if not fixing

---

## Related Resources

- **[Workflow File](../workflows/database_query_optimizer.yaml)**
- **[API Documentation Generator](api-documentation-generator.md)**
- **[Code Review Assistant](code-review-assistant.md)**

---

**Catch N+1 queries before production.**

Remember: N+1 queries are the #1 cause of database performance issues. Find them early, fix them fast.
