# Skill Writing Checklist

**Based on proven results from ISM assessment project (Jan 2026)**

Use this before creating or updating any skill.

---

## The Proven Formula

**DSC (Success State) + Critical Patterns = Success**

**Evidence:**
- statement-extractor with patterns: **56 seconds, 100% success** ✅
- statement-extractor without patterns: **395 seconds, multiple errors** ❌
- **7x performance improvement** from adding critical patterns

---

## ✅ Structure (Required)

- [ ] **SUCCESS STATE defined** - What will exist when complete
- [ ] **Critical Technical Constraints** - MUST use X library/approach (with why)
- [ ] **Common Patterns** - 2-3 code snippets showing key techniques
- [ ] **Required fields/format** - Data structure examples
- [ ] **Constraints** - 4-5 critical rules
- [ ] **Total length: 50-100 lines** (sweet spot: 90 lines)

---

## ✅ Paths (Non-negotiable)

- [ ] **Always show `/outputs/`** never `/tmp/mcp-outputs/`
- [ ] **Use f-string format:** `Path(f"/outputs/{work_dir}/{file}")`
- [ ] **Show mkdir pattern:** `output_path.parent.mkdir(parents=True, exist_ok=True)`
- [ ] **Explicit constraint:** "Use `/outputs/` not `/tmp/mcp-outputs/`"

**Why:** Small LLMs copy literally. If they see `/tmp/mcp-outputs/`, they use it (wrong path).

---

## ✅ Critical Technical Constraints (NEW - Proven Essential)

**Include a section that specifies:**

```markdown
## Critical Technical Constraints

**MUST use [specific library/tool]:**
```python
from lxml import etree  # NOT xml.etree
```

**Why:** [Explain the technical reason]
lxml supports XPath with ancestor::section[1]. Standard xml.etree does NOT.
```

**Real example from statement-extractor:**
- Without this: LLM chose `xml.etree`, got `AttributeError`, 395 seconds
- With this: LLM chose `lxml`, worked first try, 56 seconds

**When to include:**
- ✅ Specific library required (lxml vs xml.etree)
- ✅ Specific tool flags needed (curl -L for redirects)
- ✅ Critical API usage (append mode vs write mode)
- ✅ Order dependencies (clean BEFORE download)

---

## ✅ Common Patterns (Proven Essential)

**Include 2-4 code patterns showing:**

```markdown
## Common Patterns

### Pattern: [What it does]
```python
# 3-5 lines showing the technique
section_elem = row.xpath('ancestor::section[1]')
if section_elem:
    heading = section_elem[0].find('heading')
```

**Why:** Shows HOW without dictating complete implementation
```

**Guidelines:**
- Show **technique**, not complete function
- 3-5 lines per pattern max
- Explain **when** to use it
- Cover the **tricky parts** (XPath, edge cases, etc.)

**Real patterns that worked:**
```python
# Pattern: Get parent section (XPath)
section_elem = row.xpath('ancestor::section[1]')

# Pattern: Build requirements dict
for i, cell in enumerate(cells[2:]):
    requirements[levels[i]] = cell.text

# Pattern: Apply test limit AFTER extraction
statements = [...]  # Extract all first
if test_mode:
    statements = statements[:test_limit]  # Then limit
```

---

## ✅ Content Essentials

- [ ] **Active skill?** State: "generate code and execute via execute_skill_code"
- [ ] **Variable names** clearly documented (from workflow)
- [ ] **Data structures** shown with JSON examples
- [ ] **Thresholds/values** explicitly stated (0.20, 0.35, etc.)

---

## ❌ Avoid (Anti-patterns)

- [ ] Complete code templates (60+ lines of implementation)
- [ ] "How to" procedural steps without patterns
- [ ] Long explanations or philosophy sections
- [ ] Multiple redundant examples
- [ ] Host paths (`/tmp/mcp-outputs/`)
- [ ] Assuming LLM knows execution context

---

## Size Guidelines

**Proven optimal ranges:**

| Skill Type | Lines | Example |
|------------|-------|---------|
| **Simple** | 50-60 | policy-fetcher (52 lines) |
| **Medium** | 80-100 | statement-extractor (94 lines) |
| **Complex** | 100-130 | odt-parser (129 lines) |

**Red flags:**
- ❌ < 40 lines: Likely missing critical patterns
- ❌ > 150 lines: Likely has procedural bloat

---

## The Balanced Structure

```markdown
---
name: skill-name
description: One sentence
---

# Skill Name

Brief description.

## Success State

When complete:
1. File exists at `/outputs/{work_dir}/output.json`
2. Valid format (JSON array, JSONL, etc.)
3. Required fields present

## Critical Technical Constraints

**MUST use X:**
```python
from lxml import etree  # NOT xml.etree
```

**Why:** [Technical reason]

## Common Patterns

### Pattern: Key Technique
```python
# 3-5 lines showing the approach
```

### Pattern: Tricky Part
```python
# Show how to handle edge case
```

## Required Fields

```json
{
  "field": "value"
}
```

## Constraints

- Use `/outputs/` not `/tmp/mcp-outputs/`
- Format: `Path(f"/outputs/{var}/{file}")`
- [2-3 more critical rules]
```

**Total: ~90 lines**

---

## Before/After Example

### ❌ Before (No Patterns)

```markdown
## Success State
Extract table rows to JSON

## Constraints
- Use /outputs/ paths
- Include all fields
```

**Result:** LLM chose wrong library, 395 seconds, errors

### ✅ After (With Patterns)

```markdown
## Success State
Extract table rows to JSON

## Critical Technical Constraints
MUST use lxml (supports XPath)

## Common Patterns
### Pattern: Get parent section
section = row.xpath('ancestor::section[1]')

## Constraints
- Use /outputs/ paths
```

**Result:** LLM chose correct library, 56 seconds, success

---

## Testing Checklist

After writing skill, verify:

```bash
# 1. Size check
wc -l config/skills/YOUR_SKILL/SKILL.md
# Should be 50-100 lines

# 2. Path check
grep -c "/outputs/" config/skills/YOUR_SKILL/SKILL.md
# Should be 3+ occurrences

grep -c "/tmp/mcp" config/skills/YOUR_SKILL/SKILL.md
# Should be 0 (or only in constraints as "don't do this")

# 3. Pattern check
grep -c "## Common Patterns" config/skills/YOUR_SKILL/SKILL.md
# Should be 1

grep -c "Pattern:" config/skills/YOUR_SKILL/SKILL.md
# Should be 2-4
```

---

## Performance Indicators

**Good skill design shows:**
- ✅ Completes in < 60 seconds for simple tasks
- ✅ Minimal iterations (2-3 tool calls)
- ✅ No library/tool choice errors
- ✅ Correct paths on first try

**Poor skill design shows:**
- ❌ Takes 300+ seconds
- ❌ Multiple error recovery attempts (5+ iterations)
- ❌ Wrong library/tool chosen
- ❌ Path translation errors

---

## Small LLM Principles (Validated)

**What we proved:**

1. **Small LLMs are pattern matchers** ✅
   - Show them `lxml` → they use `lxml`
   - Show them `xml.etree` → they use `xml.etree`
   
2. **Small LLMs don't reason about context** ✅
   - Don't assume they know container paths
   - Don't assume they know library differences
   - Be explicit

3. **Critical patterns accelerate success** ✅
   - Without patterns: Trial and error (slow)
   - With patterns: Copy proven approach (fast)

4. **Success state + patterns = optimal** ✅
   - Success state alone: Too vague (395s)
   - Patterns alone: No goal (directionless)
   - Both together: Fast and correct (56s)

---

## Real-World Evidence

**ISM Assessment Project:**

| Aspect | Without Patterns | With Patterns | Improvement |
|--------|-----------------|---------------|-------------|
| **Execution time** | 395 seconds | 56 seconds | **7x faster** |
| **Library choice** | Wrong (xml.etree) | Correct (lxml) | **Fixed** |
| **XPath usage** | Missing | Correct | **Fixed** |
| **Iterations** | 8+ | 2 | **4x fewer** |
| **Errors** | AttributeError | None | **100% success** |

**Conclusion:** Patterns are not optional bloat—they're performance critical.

---

## Final Checklist

Before marking skill complete:

- [ ] Success state is clear (what exists when done)
- [ ] Critical technical constraints specified (MUST use X)
- [ ] 2-4 common patterns shown (key techniques)
- [ ] All paths use `/outputs/` format
- [ ] Required fields documented with examples
- [ ] 4-5 constraints listed
- [ ] Total 50-100 lines
- [ ] Tested and validated

**If all checked:** Skill is production-ready ✅

---

**Last Updated:** January 11, 2026  
**Evidence:** ISM assessment extraction skill (7x performance improvement)
