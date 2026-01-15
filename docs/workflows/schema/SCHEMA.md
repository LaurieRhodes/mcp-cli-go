# Workflow Schema Reference

⚠️ **This document has been replaced with modular documentation**

---

## New Documentation Structure

The workflow schema documentation has been reorganized for better clarity:

### 📚 Start Here

**[Quick Reference →](QUICK_REFERENCE.md)**

- One-page cheatsheet
- Complete object schemas
- Common patterns
- Best for: Getting started quickly

---

### 📖 Detailed References

**[Object Model →](OBJECT_MODEL.md)**

- TypeScript-style interfaces
- Type hierarchy
- Validation rules
- Best for: Understanding the type system

**[Inheritance Guide →](INHERITANCE_GUIDE.md)**

- Visual property flow diagrams
- Override precedence
- Provider failover
- Best for: Understanding configuration inheritance

**[Steps Reference →](STEPS_REFERENCE.md)**

- All execution modes (run, template, embeddings, consensus)
- Variable interpolation
- Dependencies and conditions
- Best for: Learning step capabilities

**[CLI Mapping →](CLI_MAPPING.md)**

- Property → CLI argument mapping
- Execution mode equivalents
- Best for: Understanding YAML ↔ CLI relationship

---

## Schema Version

**Current:** `workflow/v2.0`

All documentation covers the v2.0 workflow system only. Legacy template system (v1) is no longer supported.

---

## Quick Example

```yaml
$schema: "workflow/v2.0"
name: hello_world
version: 1.0.0
description: Simple workflow example

execution:
  provider: anthropic
  model: claude-sonnet-4
  temperature: 0.7

steps:
  - name: greet
    run: "Say hello to: {{input}}"
```

For more examples, see [Quick Reference](QUICK_REFERENCE.md).

---

**Last Updated:** 2026-01-08  
**Status:** Redirected to modular documentation
