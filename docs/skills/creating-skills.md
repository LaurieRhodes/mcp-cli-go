# Creating Skills

**Quick reference for skill creation. For comprehensive guide, see [COMPLETE_GUIDE.md](COMPLETE_GUIDE.md#creating-skills).**

## Skill Structure

```
skill-name/
├── SKILL.md           # Required: documentation
├── scripts/           # Optional: helper libraries
│   ├── __init__.py
│   └── helpers.py
└── examples/          # Optional: usage examples
    └── example1.py
```

## Step 1: Create Directory

```bash
cd config/skills
mkdir my-skill
cd my-skill
```

## Step 2: Write SKILL.md

**Required frontmatter:**
```markdown
---
name: my-skill
description: "Brief description"
---

# My Skill

Detailed documentation...

## Helper Functions

\`\`\`python
from scripts.helpers import process_data
result = process_data("input")
\`\`\`
```

**Frontmatter fields:**
- `name` - Skill identifier (lowercase-with-hyphens)
- `description` - What it does, when to use it

**Documentation should include:**
- Purpose and capabilities
- Available functions/classes
- Import statements
- Usage examples
- Parameter documentation

## Step 3: Add Helper Libraries (Optional)

**scripts/helpers.py:**
```python
"""Helper functions for my-skill."""

def process_data(input_data):
    """
    Process input data.
    
    Args:
        input_data (str): Data to process
        
    Returns:
        str: Processed result
    """
    return f"Processed: {input_data}"
```

**scripts/__init__.py:**
```python
"""My Skill package."""
from .helpers import process_data

__all__ = ['process_data']
```

## Step 4: Test the Skill

**Via mcp-cli:**
```bash
./mcp-cli serve config/runasMCP/mcp_skills_stdio.yaml
```

Check logs for:
```
Loaded skill 'my-skill'
```

**Via LLM:**
Ask it to use your skill and verify it can access the documentation and helpers.

## Custom Container Image (Optional)

If your skill needs additional packages:

**1. Create Dockerfile:**
```dockerfile
# docker/skills/Dockerfile.myskill
FROM python:3.11-slim

RUN pip install --break-system-packages \
    my-required-package \
    another-package

WORKDIR /workspace
```

**2. Add to build script:**

Edit `docker/skills/build-skills-images.sh`:
```bash
IMAGES[myskill]="Dockerfile.myskill:mcp-skills-myskill:My skill"
```

**3. Build:**
```bash
cd docker/skills
./build-skills-images.sh myskill
```

**4. Register in skill-images.yaml:**

⚠️ **CRITICAL:** Add your skill to `config/skills/skill-images.yaml`:

```yaml
skills:
  my-skill:
    image: python:3.11-slim  # Use default Python
    language: python         # Advertised via MCP
    description: "My custom skill"
```

Or for a custom container:

```yaml
skills:
  my-skill:
    image: mcp-skills-my-skill
    language: python
    dockerfile: docker/skills/Dockerfile.my-skill
    description: "My custom skill with special packages"
```

**Without this:** Skill won't work properly!

See [SKILL_IMAGES_YAML.md](SKILL_IMAGES_YAML.md) for complete reference.

## Best Practices

1. **Clear documentation** - LLMs need to understand usage
2. **Simple helpers** - Focus on reusable functions
3. **Type hints** - Help LLMs use functions correctly
4. **Docstrings** - Document all functions
5. **Examples** - Show actual usage patterns

## Example: Data Processing Skill

**SKILL.md:**
```markdown
---
name: data-processor
description: "Process and transform CSV data"
---

# Data Processor

Process CSV files with custom transformations.

## Functions

\`\`\`python
from scripts.csv_tools import read_csv, transform_data, write_csv

# Read CSV
data = read_csv('/workspace/input.csv')

# Transform
transformed = transform_data(data, operation='normalize')

# Write result
write_csv(transformed, '/outputs/result.csv')
\`\`\`
```

**scripts/csv_tools.py:**
```python
import csv

def read_csv(filepath):
    """Read CSV file into list of dicts."""
    with open(filepath) as f:
        return list(csv.DictReader(f))

def transform_data(data, operation='normalize'):
    """Transform data based on operation."""
    # Implementation
    return transformed_data

def write_csv(data, filepath):
    """Write data to CSV file."""
    if not data:
        return
    with open(filepath, 'w', newline='') as f:
        writer = csv.DictWriter(f, fieldnames=data[0].keys())
        writer.writeheader()
        writer.writerows(data)
```

## Troubleshooting

**Skill not loading:**
- Check SKILL.md has required frontmatter
- Verify directory is in `config/skills/`
- Check server logs for errors

**Imports failing:**
- Ensure `scripts/__init__.py` exists
- Check PYTHONPATH includes `/skill`
- Verify container image has required packages

**Container errors:**
- Build custom image if packages needed
- Check `skill-images.yaml` mapping
- Verify image exists: `docker images | grep mcp-skills`

---

Last updated: January 20, 2026
