# Skill Design Principles (Learned from Experience)

## Core Principle: Simplicity Over Cleverness

**Lesson Learned:** Small LLMs perform best with simple, straightforward code templates.

---

## ✅ DO: Simple Code Templates

### Example: statement-extractor

**Good (Simple):**
```python
# Skip if both fields nearly empty
if len(control_id) < 2 and len(description) < 5:
    continue
```

**Bad (Over-engineered):**
```python
def is_table_header(control_id, description):
    """Complex detection with keyword matching"""
    header_patterns = ["control", "security level", ...]
    for pattern in header_patterns:
        if pattern == control_id.lower():
            return True
    return False
```

**Why simple wins:**
- ✅ LLM can copy-paste template directly
- ✅ No confusion about function calls
- ✅ Easy to debug when it fails
- ✅ Obvious what it does

---

## ✅ DO: Generic Logic (Reusable)

### Example: Header Detection

**Good (Works with any document):**
```python
# Generic length check
if len(text) < 15:
    status = "not-applicable"
```

**Bad (Dataset-specific):**
```python
# Hard-coded for one document format
if control_id == "Control" or description == "Security Level":
    continue
```

**Why generic wins:**
- ✅ Works with different policy documents
- ✅ Works with different table structures
- ✅ No maintenance when document format changes
- ✅ Truly reusable across projects

---

## ✅ DO: Handle Edge Cases Downstream

### Example: Headers in Assessment

**Good (Flexible):**
```python
# Extract everything
statements.append({...})  # Include headers

# Later: Assessment marks headers
if len(text) < 15:
    status = "not-applicable"
```

**Bad (Rigid):**
```python
# Try to filter during extraction
if is_header(row):  # Hard-coded logic
    continue  # Skip it
```

**Why downstream wins:**
- ✅ Extraction stays simple and generic
- ✅ Assessment can use context (ISM matches)
- ✅ Easy to change thresholds
- ✅ Preserves data for debugging

---

## ✅ DO: Minimal Helper Functions

### Proven Pattern from policy-fetcher

**What worked:**
```python
# Variables from workflow (clearly labeled)
work_dir = "test"
policy_url = "https://..."

# Simple path construction
output_path = Path(f"/outputs/{work_dir}/{output_file}")

# Direct code - no helpers
response = requests.get(policy_url)
output_path.write_text(response.text)
```

**What we removed:**
```python
# Complex helper modules
from scripts.fetch import download_policy  # ❌ Too abstract
from scripts.paths import validate_path   # ❌ Adds complexity

download_policy(url, path)  # ❌ LLM confused about imports
```

**Why minimal helpers win:**
- ✅ LLM sees all logic in one place
- ✅ No "magic" - everything explicit
- ✅ Fewer import errors
- ✅ Copy-paste friendly

---

## ❌ DON'T: Complex Conditionals

### Bad Example

```python
def should_process(control_id, desc, section, metadata):
    if section in CRITICAL_SECTIONS:
        if control_id.startswith(("A.", "B.", "C.")):
            if len(desc) > 10 and metadata.get('priority') == 'high':
                return True
    return check_fallback_rules(control_id, desc)
```

### Good Example

```python
# Simple, flat logic
if len(control_id) < 2 and len(desc) < 5:
    continue  # Skip empty rows

# Process everything else
statements.append({...})
```

---

## ❌ DON'T: Hard-code Domain Knowledge

### Bad: Keyword Lists

```python
# Specific to University of Minnesota policy
HEADER_KEYWORDS = ["Control", "Security Level", "AAAM.A"]
CRITICAL_SECTIONS = ["Authentication", "Access Control"]
ISM_CONTROL_PATTERN = r"ISM-\d{4}"
```

### Good: Pattern Recognition

```python
# Generic patterns that work anywhere
has_control_format = bool(re.match(r'[A-Z]+\.[A-Z]+\.\d+', control_id))
has_content = len(text) > 50
```

---

## Pattern: Simple Code Template Structure

```python
# 1. Variables from workflow (clearly labeled)
input_file = "..."   # From workflow
output_file = "..."  # From workflow

# 2. Simple path construction
input_path = Path(f"/outputs/{work_dir}/{input_file}")
output_path = Path(f"/outputs/{work_dir}/{output_file}")

# 3. Basic validation
if not input_path.exists():
    raise FileNotFoundError(f"Not found: {input_path}")

# 4. Core logic (straightforward, no helpers)
data = process_data(input_path)

# 5. Save output
output_path.write_text(data)

# 6. Simple status message
print(f"Processed {input_path} → {output_path}")
```

**Total lines:** 20-60 (not 200+)

---

## When Complexity is Needed

**Sometimes you DO need complex logic:**
- Parsing XML with namespaces
- Vector embedding calculations
- Statistical analysis

**Solution: Pre-build and containerize**
- Include complex libraries in Docker image
- Keep skill template simple (calls libraries)
- Document which container provides what

**Example:**
```python
# Simple skill code
from lxml import etree  # ✅ Available in container

tree = etree.parse(file)  # ✅ Simple call
# Complex XML parsing happens in lxml (not our code)
```

---

## Testing Principle

**Before adding complexity, ask:**

1. Does this work with other documents? (Reusability)
2. Can a small LLM generate this code? (Simplicity)
3. Is this the simplest solution? (Minimalism)
4. Could we handle this downstream? (Flexibility)

**If 2+ answers are "no" → simplify!**

---

## Real Examples from This Project

### ✅ Success: policy-fetcher
- 40 lines total
- One clear code template
- No helper functions
- LLM generated it perfectly first try

### ❌ Initial Failure: statement-extractor
- 150+ lines with helper functions
- Hard-coded header keywords
- Complex conditional logic
- LLM struggled with complexity

### ✅ After Simplification: statement-extractor
- 60 lines total
- Simple length checks only
- No helper functions
- Works with any policy document

---

## Summary

| Principle | Why It Matters |
|-----------|----------------|
| **Simple templates** | Small LLMs copy-paste, don't reason |
| **Generic logic** | Reusable across documents |
| **Downstream handling** | Flexibility for edge cases |
| **Minimal helpers** | Reduces confusion and errors |
| **Flat conditionals** | Easy to understand and debug |
| **No hard-coding** | Domain-agnostic skills |

**Golden Rule:** If you can't explain it in 3 sentences, it's too complex for a skill template.
