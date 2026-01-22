# Output Directory

## Purpose

Files created by skills persist after containers exit by mounting `/outputs` from the host filesystem.

## Automatic LLM Guidance

**Good news:** The LLM is automatically taught to use `/outputs/`!

When skills are enabled in **chat mode** or **query mode**, mcp-cli automatically enhances the system prompt to guide the LLM:

```
When writing code, save output files to /outputs/ directory:
   output.save('/outputs/result.docx')  ✅ CORRECT - File persists to host
   output.save('/workspace/result.docx') ❌ WRONG - File deleted when container exits
   output.save('result.docx') ❌ WRONG - Defaults to /workspace/
```

This means in normal usage, the LLM will automatically save files to `/outputs/` without you needing to specify it in your prompts.

## Configuration

Edit `config/settings.yaml`:

```yaml
skills:
  outputs_dir: "/path/to/your/outputs"
```

Default: `/tmp/mcp-outputs`

## Usage

In skill code:

```python
# Persists to host
prs.save('/outputs/file.pptx')

# Deleted when container exits
prs.save('/workspace/file.pptx')
```

Files at `/outputs/file.pptx` in container appear at the configured path on host.

## Verification

```bash
# Check current configuration
grep -A 2 "^skills:" config/settings.yaml

# Create directory if needed
mkdir -p /path/to/your/outputs
```

---

Last updated: January 20, 2026
