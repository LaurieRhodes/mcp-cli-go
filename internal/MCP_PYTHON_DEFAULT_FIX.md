# MCP Python Default Problem - Root Cause Analysis

**Date:** January 18, 2026  
**Severity:** CRITICAL - Breaks bash-only skills  
**Issue:** MCP server hardcodes Python as default language

---

## The Problem

### What We Observed

Context-builder skill:
- ✅ SKILL.md says "⚠️ BASH SCRIPTS ONLY"
- ✅ Only .sh files in scripts/ directory
- ✅ No __init__.py file
- ✅ Multiple "DON'T USE PYTHON" warnings

**But LLM still tries Python!**

**Error:**
```
Error: crun: executable file `python` not found in $PATH
```

---

## Root Cause: MCP Tool Schema

### Location: `/internal/services/skills/service.go:1115`

```go
"language": map[string]interface{}{
    "type":        "string",
    "enum":        []string{"python", "bash"},
    "description": "Programming language ('python' or 'bash')",
    "default":     "python",  ← PROBLEM: Hardcoded default
},
```

**This is what the LLM sees** when it loads the `execute_skill_code` tool.

---

### What the LLM Receives

When LLM queries available tools, it gets this schema:

```json
{
  "name": "execute_skill_code",
  "input_schema": {
    "properties": {
      "skill_name": {
        "type": "string",
        "description": "Name of skill..."
      },
      "language": {
        "type": "string",
        "enum": ["python", "bash"],
        "description": "Programming language ('python' or 'bash')",
        "default": "python"  ← LLM SEES THIS
      },
      "code": {
        "type": "string",
        "description": "Code to execute (Python or Bash). Example: doc.save('/outputs/file.docx')"
      }
    },
    "required": ["skill_name", "code"]
  }
}
```

**Key issues:**
1. `"default": "python"` - Explicitly defaults to Python
2. `"required": ["skill_name", "code"]` - language NOT required!
3. Example is Python code: `doc.save('/outputs/file.docx')`
4. `"enum": ["python", "bash"]` - Python listed first

---

## Why This Overrides SKILL.md

### Information Hierarchy (LLM's Perspective)

```
1. Tool Schema (from MCP server)
   ├─ "default": "python"          ← HIGHEST AUTHORITY
   ├─ language not required
   └─ Python example given
   
2. SKILL.md (from skill directory)
   ├─ "⚠️ BASH SCRIPTS ONLY"
   ├─ "DON'T USE PYTHON"
   └─ Shows bash examples
   
3. Directory Structure
   ├─ Only .sh files
   └─ No __init__.py
```

**LLM reasoning:**
- "Tool schema says default is Python"
- "Language parameter is optional"
- "SKILL.md says bash, but tool schema is more authoritative"
- "I'll default to Python as the schema suggests"

**Result:** Python wins even with perfect SKILL.md

---

## Secondary Issue: Fallback Code

### Location: `/internal/services/server/service.go:476`

```go
// Extract language (default to python)
language := "python"  ← PROBLEM: Fallback to Python
if lang, ok := arguments["language"].(string); ok && lang != "" {
    language = lang
}
```

**What this does:**
- If LLM doesn't specify `language` parameter
- Code defaults to `"python"`
- Executes as Python even for bash-only skills

---

## The Complete Problem Chain

```
1. LLM sees tool schema
   └─ "default": "python"
   
2. LLM uses skill
   ├─ SKILL.md says "BASH ONLY"
   └─ Tool schema says "default python"
   
3. LLM makes decision
   ├─ Tool schema is authoritative
   └─ Decides to use default (Python)
   
4. LLM omits language parameter
   └─ Assumes default will be used
   
5. Server receives request
   ├─ No language specified
   └─ Fallback code: language := "python"
   
6. Server executes
   ├─ Tries to run Python in bash-only container
   └─ ERROR: python not found
```

---

## The Fix

### Part 1: Make Language Required

**File:** `/internal/services/skills/service.go`

**Current (line 1111-1116):**
```go
"language": map[string]interface{}{
    "type":        "string",
    "enum":        []string{"python", "bash"},
    "description": "Programming language ('python' or 'bash')",
    "default":     "python",  ← REMOVE THIS
},
```

**Fixed:**
```go
"language": map[string]interface{}{
    "type":        "string",
    "enum":        []string{"bash", "python"},  // Bash first = no implicit preference
    "description": "Programming language: 'bash' for bash skills, 'python' for Python skills. Check skill's SKILL.md for required language.",
    // NO DEFAULT - force LLM to specify
},
```

**And update required array (line 1127):**
```go
"required": []string{"skill_name", "language", "code"},  // Add "language"
```

---

### Part 2: Remove Fallback Default

**File:** `/internal/services/server/service.go`

**Current (line 475-479):**
```go
// Extract language (default to python)
language := "python"
if lang, ok := arguments["language"].(string); ok && lang != "" {
    language = lang
}
```

**Fixed:**
```go
// Extract language (REQUIRED - no default)
language, ok := arguments["language"].(string)
if !ok || language == "" {
    return nil, fmt.Errorf("language parameter is required (must be 'bash' or 'python')")
}
if language != "bash" && language != "python" {
    return nil, fmt.Errorf("language must be 'bash' or 'python', got: %s", language)
}
```

---

### Part 3: Update Description to be Neutral

**Current:**
```go
"description": "Code to execute (Python or Bash). IMPORTANT: Save all files to /outputs/ directory only. Example: doc.save('/outputs/file.docx')",
```

**Fixed:**
```go
"description": "Code to execute in the specified language. IMPORTANT: Save all files to /outputs/ directory only. " +
    "Bash example: bash /skill/scripts/process.sh /outputs/input.dat /outputs/output.dat | " +
    "Python example: doc.save('/outputs/file.docx')",
```

---

## Impact of Fix

### Before (Broken)

**Tool schema advertises:**
- Default: Python
- Language: Optional
- Example: Python code

**LLM behavior:**
- Sees Python as default
- Doesn't specify language
- Server defaults to Python
- **Result:** Fails on bash-only skills

---

### After (Fixed)

**Tool schema advertises:**
- Default: NONE (must specify)
- Language: **REQUIRED**
- Example: Both bash and Python

**LLM behavior:**
- Must read SKILL.md to know language
- Must specify language explicitly
- Server validates language parameter
- **Result:** Uses correct language for each skill

---

## Testing the Fix

### Step 1: Apply Code Changes

```bash
cd /media/laurie/Data/Github/mcp-cli-go

# Edit internal/services/skills/service.go
# - Remove "default": "python" line
# - Add "language" to required array
# - Update description

# Edit internal/services/server/service.go
# - Remove language := "python" default
# - Add validation

# Rebuild
go build -o mcp-cli .
```

---

### Step 2: Test with context-builder

```bash
# Should now work because LLM must specify language
./mcp-cli --workflow rlm_poc/workflows/test_skill_pattern_minimal_v2
```

**Expected:**
- LLM reads SKILL.md
- Sees "BASH SCRIPTS ONLY" 
- **Now required to specify language**
- Calls with `language: 'bash'`
- ✅ Success

---

### Step 3: Test with Python skill

```bash
# Test with a Python skill (docx, pdf, etc.)
./mcp-cli query --server skills "Use docx skill to create a document"
```

**Expected:**
- LLM reads SKILL.md
- Sees Python libraries mentioned
- Specifies `language: 'python'`
- ✅ Success

---

## Why This Fix Works

### Current (Broken) Flow

```
LLM → Tool schema shows default=python
    → Doesn't specify language
    → Server defaults to python
    → Fails on bash skills
```

---

### Fixed Flow

```
LLM → Tool schema requires language
    → Must read SKILL.md
    → Sees "BASH SCRIPTS ONLY"
    → Specifies language='bash'
    → Server validates and uses bash
    → ✅ Success
```

---

## Alternative Solutions (Rejected)

### Option 1: Skill-Specific Defaults

**Idea:** Set default based on skill name

**Rejected because:**
- Still defaults without reading SKILL.md
- Brittle (must maintain skill→language mapping)
- Doesn't force LLM to read documentation

---

### Option 2: Auto-Detect from Directory

**Idea:** If scripts/ has only .sh files, assume bash

**Rejected because:**
- Magic behavior (implicit)
- Doesn't work for mixed skills
- Hides the requirement from LLM

---

### Option 3: Separate Tools

**Idea:** execute_bash_code and execute_python_code

**Rejected because:**
- Duplicates tool definitions
- More complex for LLM
- Doesn't solve the core issue

---

## The Correct Solution: Required Parameter

**Why this is best:**

1. **Forces LLM to read SKILL.md**
   - Must know language to call tool
   - Can't rely on defaults
   - Explicit is better than implicit

2. **No ambiguity**
   - Tool schema is clear
   - SKILL.md is authoritative
   - No conflicts

3. **Works for all skills**
   - Bash-only ✓
   - Python-only ✓
   - Mixed skills ✓

4. **Fails fast**
   - Missing language → immediate error
   - Wrong language → clear message
   - Easy to debug

---

## Implementation Checklist

- [ ] Edit `/internal/services/skills/service.go`
  - [ ] Remove `"default": "python"` line
  - [ ] Change enum order to `["bash", "python"]`
  - [ ] Update description to be neutral
  - [ ] Add "language" to required array
  
- [ ] Edit `/internal/services/server/service.go`
  - [ ] Remove `language := "python"` default
  - [ ] Add validation for missing language
  - [ ] Add validation for invalid language
  
- [ ] Rebuild mcp-cli
  - [ ] `go build -o mcp-cli .`
  
- [ ] Test with bash skill
  - [ ] context-builder should work
  
- [ ] Test with Python skill
  - [ ] docx/pdf/xlsx should still work
  
- [ ] Update documentation
  - [ ] Note that language is now required
  - [ ] Update examples

---

## Code Changes

### File 1: `/internal/services/skills/service.go`

**Lines 1111-1116 (BEFORE):**
```go
"language": map[string]interface{}{
    "type":        "string",
    "enum":        []string{"python", "bash"},
    "description": "Programming language ('python' or 'bash')",
    "default":     "python",
},
```

**Lines 1111-1115 (AFTER):**
```go
"language": map[string]interface{}{
    "type":        "string",
    "enum":        []string{"bash", "python"},
    "description": "Programming language: 'bash' for bash skills, 'python' for Python skills. Check skill's SKILL.md for required language.",
},
```

**Line 1127 (BEFORE):**
```go
"required": []string{"skill_name", "code"},
```

**Line 1127 (AFTER):**
```go
"required": []string{"skill_name", "language", "code"},
```

**Lines 1117-1120 (BEFORE):**
```go
"code": map[string]interface{}{
    "type":        "string",
    "description": "Code to execute (Python or Bash). IMPORTANT: Save all files to /outputs/ directory only. Example: doc.save('/outputs/file.docx')",
},
```

**Lines 1117-1120 (AFTER):**
```go
"code": map[string]interface{}{
    "type":        "string",
    "description": "Code to execute in the specified language. IMPORTANT: Save all files to /outputs/ directory only. " +
        "Bash example: bash /skill/scripts/process.sh /outputs/input.dat /outputs/output.dat | " +
        "Python example: doc.save('/outputs/file.docx')",
},
```

---

### File 2: `/internal/services/server/service.go`

**Lines 475-479 (BEFORE):**
```go
// Extract language (default to python)
language := "python"
if lang, ok := arguments["language"].(string); ok && lang != "" {
    language = lang
}
```

**Lines 475-481 (AFTER):**
```go
// Extract language (REQUIRED - no default)
language, ok := arguments["language"].(string)
if !ok || language == "" {
    return nil, fmt.Errorf("language parameter is required (must be 'bash' or 'python')")
}
if language != "bash" && language != "python" {
    return nil, fmt.Errorf("language must be 'bash' or 'python', got: %s", language)
}
```

---

## Summary

**Root Cause:**
- MCP server hardcodes `"default": "python"` in tool schema
- LLM sees this and defaults to Python
- Overrides SKILL.md documentation

**Fix:**
- Remove default from tool schema
- Make language REQUIRED parameter
- Force LLM to read SKILL.md to know language
- Add validation in server code

**Impact:**
- Bash-only skills will work ✓
- Python skills still work ✓
- Mixed skills possible ✓
- Clear error messages ✓
- No implicit defaults ✓

**Status:** Ready to implement  
**Risk:** Low (just removes bad default)  
**Testing:** Required before deployment

---

**Discovered:** January 18, 2026  
**Category:** MCP Server Configuration  
**Severity:** Critical (breaks bash-only skills)  
**Fix Complexity:** Simple (4 line changes + validation)
