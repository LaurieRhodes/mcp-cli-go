# Creating Skills Guide

Learn how to create custom skills for mcp-cli.

## Skill Anatomy

A skill is a directory containing:

```
my-skill/
├── SKILL.md              # Documentation (required)
├── scripts/              # Helper libraries (optional)
│   ├── __init__.py
│   └── helpers.py
└── references/           # Additional docs (optional)
    └── examples.md
```

## Creating Your First Skill

### Step 1: Create Directory

```bash
cd config/skills
mkdir my-skill
cd my-skill
```

### Step 2: Write SKILL.md

This is what Claude reads to learn about your skill.

**Minimal SKILL.md:**
```markdown
---
name: my-skill
description: Brief description of what this skill does
---

# My Skill

Detailed description of the skill's purpose and capabilities.

## Helper Functions

\`\`\`python
from scripts.helpers import process_data

result = process_data("input")
print(result)
\`\`\`

## Usage Example

\`\`\`python
from scripts.helpers import process_data, format_output

# Process the data
processed = process_data(raw_data)

# Format for output
formatted = format_output(processed)
print(formatted)
\`\`\`
```

**YAML Frontmatter (required):**
- `name`: Skill identifier (use hyphens: `my-skill`)
- `description`: One-line description

**Documentation Body:**
- Explain what the skill does
- Show import examples
- List available functions/classes
- Provide usage examples
- Document parameters and return values

### Step 3: Create Helper Libraries

```bash
mkdir scripts
```

**scripts/helpers.py:**
```python
"""Helper functions for my-skill."""

def process_data(input_data):
    """
    Process input data and return result.
    
    Args:
        input_data (str): Data to process
        
    Returns:
        str: Processed result
    """
    # Your processing logic
    return f"Processed: {input_data}"

def format_output(data):
    """
    Format data for display.
    
    Args:
        data (str): Data to format
        
    Returns:
        str: Formatted output
    """
    return f"=== {data} ==="

class DataProcessor:
    """Process data with configurable options."""
    
    def __init__(self, prefix=">>"):
        self.prefix = prefix
        self.count = 0
    
    def process(self, item):
        """Process a single item."""
        self.count += 1
        return f"{self.prefix} {item}"
    
    def get_stats(self):
        """Get processing statistics."""
        return {"processed": self.count, "prefix": self.prefix}
```

**scripts/__init__.py:**
```python
"""My skill helper libraries."""
from .helpers import process_data, format_output, DataProcessor

__all__ = ['process_data', 'format_output', 'DataProcessor']
```

### Step 4: Test Your Skill

**Test imports:**
```bash
cd /path/to/config/skills/my-skill
python3 -c "from scripts import helpers; print(helpers.process_data('test'))"
```

**Test in sandbox:**

Create test script `test_skill.py`:
```python
import sys
sys.path.insert(0, '/skill')

from scripts.helpers import process_data, DataProcessor

# Test function
result = process_data("test data")
print(f"Function result: {result}")

# Test class
processor = DataProcessor(prefix="→")
items = ["item1", "item2", "item3"]
for item in items:
    print(processor.process(item))

stats = processor.get_stats()
print(f"Stats: {stats}")
```

Execute with execute_skill_code:
```javascript
const result = await execute_skill_code({
    skill_name: "my-skill",
    code: fs.readFileSync('test_skill.py', 'utf8')
})
```

### Step 5: Document Well

**Good documentation example:**

```markdown
---
name: text-analyzer
description: Analyze text for sentiment, entities, and keywords
---

# Text Analyzer Skill

Provides text analysis capabilities including sentiment analysis,
entity extraction, and keyword identification.

## Available Functions

### analyze_sentiment(text)

Analyze the sentiment of text.

**Parameters:**
- `text` (str): Text to analyze

**Returns:**
- dict: `{"sentiment": "positive|negative|neutral", "score": 0.0-1.0}`

**Example:**
\`\`\`python
from scripts.analyzer import analyze_sentiment

result = analyze_sentiment("This is great!")
print(result)  # {"sentiment": "positive", "score": 0.95}
\`\`\`

### extract_entities(text)

Extract named entities from text.

**Parameters:**
- `text` (str): Text to process

**Returns:**
- list: List of entities with types

**Example:**
\`\`\`python
from scripts.analyzer import extract_entities

entities = extract_entities("Apple Inc. was founded by Steve Jobs")
# [{"text": "Apple Inc.", "type": "ORGANIZATION"},
#  {"text": "Steve Jobs", "type": "PERSON"}]
\`\`\`

## Complete Example

\`\`\`python
from scripts.analyzer import analyze_sentiment, extract_entities, get_keywords

text = "Your text here..."

# Analyze sentiment
sentiment = analyze_sentiment(text)
print(f"Sentiment: {sentiment['sentiment']} ({sentiment['score']})")

# Extract entities
entities = extract_entities(text)
for entity in entities:
    print(f"  {entity['text']} ({entity['type']})")

# Get keywords
keywords = get_keywords(text, top_n=5)
print(f"Keywords: {', '.join(keywords)}")
\`\`\`

## Notes

- Uses standard library only (no external dependencies)
- All functions work with English text
- Results are deterministic for same input
```

## Best Practices

### Documentation

✅ **DO:**
- Provide complete function signatures
- Include parameter and return types
- Show concrete examples
- Explain edge cases
- Document any limitations

❌ **DON'T:**
- Assume Claude knows your API
- Use vague descriptions
- Skip examples
- Leave parameters undocumented

### Code Style

✅ **DO:**
- Use clear function names
- Add docstrings
- Keep functions focused
- Return predictable types
- Handle errors gracefully

❌ **DON'T:**
- Use cryptic abbreviations
- Create god classes
- Mix concerns
- Return inconsistent types
- Let exceptions propagate silently

### Dependencies

✅ **DO:**
- Use standard library when possible
- Document required dependencies
- Provide installation instructions
- Consider creating custom Docker image

❌ **DON'T:**
- Assume dependencies are installed
- Use version-specific features without noting
- Import heavy libraries unnecessarily

### Testing

✅ **DO:**
- Test imports work
- Test each function independently
- Test with various inputs
- Test error cases
- Verify in sandbox environment

❌ **DON'T:**
- Skip testing helper functions
- Only test happy path
- Assume sandbox matches local environment

## Advanced Patterns

### Multiple Helper Modules

```
my-skill/
├── SKILL.md
└── scripts/
    ├── __init__.py
    ├── analyzers.py      # Analysis functions
    ├── formatters.py     # Formatting functions
    └── utilities.py      # Utility functions
```

**scripts/__init__.py:**
```python
"""My skill helper libraries."""
from .analyzers import analyze_text, extract_data
from .formatters import format_report, format_table
from .utilities import validate_input, normalize_data

__all__ = [
    'analyze_text', 'extract_data',
    'format_report', 'format_table',
    'validate_input', 'normalize_data'
]
```

### Configuration Classes

```python
class ProcessorConfig:
    """Configuration for data processing."""
    
    def __init__(self, mode='standard', threshold=0.5):
        self.mode = mode
        self.threshold = threshold
        self.validate()
    
    def validate(self):
        """Validate configuration."""
        if self.mode not in ['standard', 'aggressive', 'conservative']:
            raise ValueError(f"Invalid mode: {self.mode}")
        if not 0 <= self.threshold <= 1:
            raise ValueError("Threshold must be between 0 and 1")

def process_with_config(data, config):
    """Process data using configuration."""
    # Use config.mode and config.threshold
    return processed_data
```

### State Management

```python
class StatefulProcessor:
    """Processor that maintains state across calls."""
    
    def __init__(self):
        self.history = []
        self.stats = {'total': 0, 'errors': 0}
    
    def process(self, item):
        """Process item and update state."""
        try:
            result = self._do_process(item)
            self.history.append(result)
            self.stats['total'] += 1
            return result
        except Exception as e:
            self.stats['errors'] += 1
            return None
    
    def get_summary(self):
        """Get processing summary."""
        return {
            'total': self.stats['total'],
            'errors': self.stats['errors'],
            'success_rate': (self.stats['total'] - self.stats['errors']) / max(self.stats['total'], 1)
        }
```

## Packaging Skills

### For Distribution

If sharing your skill:

1. **Add LICENSE** - Choose appropriate license
2. **Add README.md** - Installation and usage instructions
3. **Version your skill** - Use semantic versioning
4. **Test thoroughly** - On different systems
5. **Document dependencies** - Including Python version

### Directory Structure for Distribution

```
my-skill/
├── LICENSE
├── README.md
├── SKILL.md
├── scripts/
│   ├── __init__.py
│   └── helpers.py
├── tests/
│   └── test_helpers.py
└── examples/
    └── basic_usage.py
```

## Troubleshooting

### "No module named 'scripts'"

Ensure `scripts/__init__.py` exists and exports functions.

### "Permission denied" writing files

Remember code runs in `/workspace` (writable) but skill is at `/skill` (read-only).

### Functions not available

Check `__all__` in `__init__.py` includes your functions.

### Import errors

Verify PYTHONPATH includes `/skill` (automatic in execute_skill_code).

## Next Steps

- **[Best Practices](best-practices.md)** - Design patterns and conventions
- **[Available Skills](available-skills.md)** - Study existing skills for examples
- **[Reference](reference.md)** - Technical details on execute_skill_code
- **[Troubleshooting](troubleshooting.md)** - Common issues and solutions
