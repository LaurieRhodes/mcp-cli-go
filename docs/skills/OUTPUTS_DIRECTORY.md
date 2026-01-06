# Output Directory

## Purpose

Files created by skills persist after containers exit by mounting `/outputs` from the host filesystem.

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

Last updated: January 6, 2026
