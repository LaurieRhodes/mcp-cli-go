# Container Selection Quick Reference

**Choose the right container for your skill**

---

## Decision Tree

```
What does your skill need?
│
├─ Only bash + text tools?
│  ├─ xmlstarlet? ────────────┐
│  ├─ jq? ─────────────────────┤
│  ├─ grep/sed/awk? ───────────┤──→ mcp-skills-bash-tools
│  └─ bc/curl? ───────────────┘    (~15MB, 64MB RAM)
│
├─ Python + document processing?
│  ├─ Pandoc? ────────────────┐
│  ├─ pdfplumber? ────────────┤
│  ├─ requests? ──────────────┤──→ mcp-skills-document-parsing
│  └─ lxml? ──────────────────┘    (~300MB, 256MB RAM)
│
├─ Office document formats?
│  ├─ .docx? ─────────────────┬──→ mcp-skills-docx
│  ├─ .pptx? ─────────────────┼──→ mcp-skills-pptx
│  ├─ .xlsx? ─────────────────┼──→ mcp-skills-xlsx
│  └─ .pdf? ──────────────────┴──→ mcp-skills-pdf
│                                   (~200-500MB, 256-512MB RAM)
│
└─ General Python?
   └─ Not sure yet? ──────────────→ python:3.11-slim
                                    (~150MB, 256MB RAM)
```

---

## Container Comparison

| Container | Size | RAM | Tools | Use Case |
|-----------|------|-----|-------|----------|
| **bash-tools** | 15MB | 64MB | bash, xmlstarlet, jq, grep, sed, awk, bc | Bash-only text processing |
| **python:3.11-slim** | 150MB | 256MB | Python 3.11, pip | General Python scripts |
| **document-parsing** | 300MB | 256MB | Python, pandoc, pdfplumber, requests, lxml | Document conversion & analysis |
| **docx/pptx/xlsx** | 200-500MB | 256-512MB | Python, office format libraries | Office document manipulation |

---

## Quick Examples

### ✅ Use bash-tools

**context-builder:**
```yaml
context-builder:
  image: mcp-skills-bash-tools
  dockerfile: docker/skills/Dockerfile.bash-tools
  memory: 64MB
  cpu: "0.25"
```

**Why:** Only uses bash + xmlstarlet + jq

---

**xml-transformer:**
```yaml
xml-transformer:
  image: mcp-skills-bash-tools
  dockerfile: docker/skills/Dockerfile.bash-tools
  memory: 64MB
```

**Why:** xmlstarlet for XML, jq for output

---

### ✅ Use document-parsing

**policy-fetcher:**
```yaml
policy-fetcher:
  image: mcp-skills-document-parsing
  dockerfile: docker/skills/Dockerfile.document-parsing
```

**Why:** Uses Python requests library

---

**odt-parser:**
```yaml
odt-parser:
  image: mcp-skills-document-parsing
  dockerfile: docker/skills/Dockerfile.document-parsing
```

**Why:** Uses Python + lxml for ODT parsing

---

### ✅ Use office format containers

**docx-editor:**
```yaml
docx-editor:
  image: mcp-skills-docx
  dockerfile: docker/skills/Dockerfile.docx
```

**Why:** Needs python-docx for .docx manipulation

---

## How to Check Your Skill

### Is it bash-only?

```bash
cd /path/to/skill

# 1. Check SKILL.md
grep -i "bash.*only\|bash scripts only" SKILL.md
# Found? → bash-tools

# 2. Check scripts/
ls scripts/
# Only .sh files? → bash-tools
# Has .py files? → NOT bash-tools

# 3. Check for Python imports
grep -r "import " scripts/
# None found? → bash-tools

# 4. Check for __init__.py
find scripts/ -name "__init__.py"
# None found AND only .sh? → bash-tools
```

---

## Container Selection Checklist

- [ ] Skill only uses bash scripts (.sh files)?
- [ ] No Python imports in code?
- [ ] Only needs: xmlstarlet, jq, grep, sed, awk, bc?
- [ ] SKILL.md says "BASH ONLY" or similar?
- [ ] No __init__.py in scripts/ directory?

**If all YES → Use bash-tools**

---

## Benefits by Container

### bash-tools

**Pros:**
- Smallest (15MB)
- Fastest startup (<1s)
- Lowest memory (64MB)
- Minimal attack surface

**Cons:**
- No Python
- No complex libraries
- Text processing only

**Best for:**
- XML/JSON transformations
- Text filtering/processing
- Simple data extraction
- Bash-based workflows

---

### document-parsing

**Pros:**
- Has pandoc for conversion
- Python for complex logic
- PDF/HTML processing
- Many useful libraries

**Cons:**
- Larger (300MB)
- More memory (256MB)
- Slower startup (~3s)

**Best for:**
- Document conversion
- PDF extraction
- Web scraping
- Complex parsing

---

### Office formats

**Pros:**
- Format-specific libraries
- Proper structure handling
- Full feature support

**Cons:**
- Largest (200-500MB)
- Most memory needed
- Slowest startup

**Best for:**
- Creating/editing .docx/.pptx/.xlsx
- Preserving formatting
- Office automation

---

## Migration Guide

### From document-parsing to bash-tools

**When to migrate:**
- Skill only uses bash + text tools
- No Python dependencies
- Want faster/lighter execution

**Steps:**
1. Verify bash-only (use checklist above)
2. Update skill-images.yaml:
   ```yaml
   your-skill:
     image: mcp-skills-bash-tools
     dockerfile: docker/skills/Dockerfile.bash-tools
     memory: 64MB
     cpu: "0.25"
   ```
3. Test thoroughly
4. Measure improvement

---

## Summary

**Choose based on needs, not defaults**

- Bash-only? → bash-tools (lightest)
- Need Python? → Check what for
- Document processing? → document-parsing
- Office formats? → format-specific container
- Not sure? → python:3.11-slim (can optimize later)

**Optimize later is better than never optimize!**

---

**Created:** January 18, 2026  
**Updated:** As new containers added  
**Reference:** `/media/laurie/Data/Github/mcp-cli-go/config/skills/skill-images.yaml`
