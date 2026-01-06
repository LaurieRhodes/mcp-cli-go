# Skills Directory

This directory contains Anthropic-compatible Skills for use with mcp-cli-go.

## What are Skills?

Skills are modular capabilities that extend Claude's functionality through organized folders containing:

- **SKILL.md** (required): Main skill instructions with YAML frontmatter
- **Supporting files** (optional): Reference documentation, examples, templates, scripts

Skills follow the [Anthropic Skills specification](https://code.claude.com/docs/en/skills) and are compatible with:

- Claude Code
- Claude Desktop (via MCP)
- Any MCP client
- **mcp-cli-go** (via Skills-as-MCP-Tools)

## Directory Structure

```
skills/
├── README.md                          # This file
│
└── python-best-practices/             # Example skill
    ├── SKILL.md                       # Main skill file (required)
    ├── reference.md                   # Advanced patterns
    ├── examples.md                    # Complete examples
    └── templates/                     # Code templates
        └── python_class_template.py
```

## Creating a Skill

### 1. Create Skill Directory

```bash
mkdir -p config/skills/my-skill-name
```

**Naming requirements:**

- Lowercase letters, numbers, and hyphens only
- Maximum 64 characters
- Descriptive and specific (e.g., `pdf-form-filling`, not `documents`)

### 2. Create SKILL.md

Every skill must have a `SKILL.md` file with YAML frontmatter:

```markdown
---
name: my-skill-name
description: Brief description of what this skill does and when to use it. Include specific triggers and use cases.
---

# My Skill Name

## Instructions
Clear, step-by-step instructions for Claude to follow.

## Examples
Concrete examples of using this skill.

## Best Practices
Guidelines for effective use.
```

**Required fields:**

- `name`: Must match directory name (lowercase, hyphens, max 64 chars)
- `description`: What the skill does and when to use it (max 1024 chars)

**Optional fields:**

- `allowed-tools`: Restrict which tools Claude can use (e.g., `Read, Grep, Glob`)

### 3. Add Supporting Files (Optional)

Supporting files are loaded progressively (only when needed):

```
my-skill/
├── SKILL.md              # Main instructions
├── reference.md          # Detailed reference
├── examples.md           # Complete examples
├── scripts/              # Helper scripts
│   └── helper.py
└── templates/            # Templates
    └── template.txt
```

Reference these from SKILL.md:

```markdown
For advanced usage, see [reference.md](reference.md).

Use this template:
\`\`\`bash
python scripts/helper.py input.txt
\`\`\`
```

## Skill Description Best Practices

The `description` field is **critical** for skill discovery. It should:

1. **Explain what the skill does**
2. **Specify when to use it**
3. **Include trigger keywords**

**Good description:**

```yaml
description: Extract text and tables from PDF files, fill forms, merge documents. Use when working with PDF files or when the user mentions PDFs, forms, or document extraction. Requires pypdf and pdfplumber packages.
```

**Bad description:**

```yaml
description: Helps with documents
```

### Trigger Keywords

Include terms users would naturally use:

**For a Python skill:**

- "Python", ".py files", "pytest", "Django", "Flask", "FastAPI"

**For a PDF skill:**

- "PDF", "forms", "document extraction", "merge", "split"

**For a database skill:**

- "SQL", "database", "queries", "PostgreSQL", "MySQL"

## Using Skills with mcp-cli-go

### Method 1: Via MCP Server (Recommended)

Skills can be exposed as MCP tools that any LLM can use.

**1. Create a runas configuration:**

```yaml
# config/runas/skills-server.yaml
runas_type: mcp
version: "1.0"

server_info:
  name: skills_engine
  version: 1.0.0
  description: "Universal Skills engine"

tools:
  - template: load_skill
    name: load_skill
    description: |
      LOAD SKILL

      Loads a skill into context.

      → Use when: Need domain expertise, coding patterns, best practices

    input_schema:
      type: object
      properties:
        skill_name:
          type: string
          description: "Skill to load (e.g., 'python-best-practices')"
      required:
        - skill_name
```

**2. Create skill loader template:**

```yaml
# config/templates/load_skill.yaml
name: load_skill
description: "Load and return a skill"
version: 1.0.0

steps:
  - name: load_skill_content
    prompt: |
      Load skill from: config/skills/{{input_data.skill_name}}/SKILL.md

      Return the complete skill content.
    output: skill_content
```

**3. Start MCP server:**

```bash
mcp-cli serve config/runas/skills-server.yaml
```

**4. Skills work with ANY LLM:**

- Claude (via Claude Desktop)
- GPT-4 (via custom client)
- Ollama (via MCP)
- Any MCP-compatible client

### Method 2: Direct CLI Access (Future)

```bash
# Load skill directly
mcp-cli skill load python-best-practices

# List available skills
mcp-cli skill list

# Install from community
mcp-cli skill install react-best-practices
```

## Progressive Disclosure

Skills use progressive disclosure to manage context efficiently:

1. **Metadata First**: Claude sees name and description
2. **Main Content**: Loads SKILL.md if skill is relevant
3. **Supporting Files**: Loads reference.md, examples.md only when needed

This means you can create comprehensive skills without overwhelming context.

## Example Skills

### 1. Simple Skill (Single File)

```
commit-helper/
└── SKILL.md
```

**SKILL.md:**

```markdown
---
name: commit-helper
description: Generates clear commit messages from git diffs. Use when writing commits or reviewing staged changes.
---

# Commit Helper

## Instructions
1. Run `git diff --staged` to see changes
2. Generate commit message with:
   - Summary under 50 characters
   - Detailed description
   - Affected components
```

### 2. Multi-File Skill

```
pdf-processing/
├── SKILL.md
├── reference.md
├── examples.md
└── scripts/
    └── fill_form.py
```

See `python-best-practices/` for a complete example.

## Skill Quality Checklist

Before finalizing a skill:

- [ ] Name is descriptive and follows naming rules
- [ ] Description includes what skill does AND when to use it
- [ ] Description includes trigger keywords
- [ ] Instructions are clear and step-by-step
- [ ] Examples are concrete and realistic
- [ ] Supporting files are referenced from SKILL.md
- [ ] Tested with relevant queries
- [ ] Documentation is complete

## Testing Skills

Test that Claude discovers and uses your skill:

**For python-best-practices:**

```
"Can you review this Python code for best practices?"
"Help me write Python code following PEP 8"
"What are Python naming conventions?"
```

**For a PDF skill:**

```
"Extract text from this PDF"
"Fill out this PDF form"
"How do I merge PDF files?"
```

Claude should autonomously use the skill when questions match the description.

## Sharing Skills

Skills can be shared:

1. **Via Git**: Commit to project repository
2. **Via MCP**: Expose as MCP tools (works across all LLMs)
3. **Via Plugins**: Package as Claude Code plugin
4. **Via Community**: Contribute to skill marketplace

## Skill Versioning

Document versions in your SKILL.md:

```markdown
## Version History
- v2.0.0 (2025-01-15): Breaking changes to API
- v1.1.0 (2025-01-01): Added new features
- v1.0.0 (2024-12-01): Initial release
```

## Best Practices Summary

1. **Keep skills focused**: One skill = one capability
2. **Write clear descriptions**: Include what + when + keywords
3. **Use progressive disclosure**: Main content → supporting files
4. **Test thoroughly**: Verify skill activates correctly
5. **Document well**: Clear instructions and examples
6. **Version carefully**: Track changes for team awareness

## Resources

- [Anthropic Skills Documentation](https://code.claude.com/docs/en/skills)
- [Skills Best Practices](https://docs.claude.com/en/docs/agents-and-tools/agent-skills/best-practices)
- [MCP Protocol](https://modelcontextprotocol.io)
- [mcp-cli-go Documentation](../../docs/)

## Support

For issues or questions:

1. Check this README
2. Review example skills (python-best-practices/)
3. See main documentation in `docs/`
4. Test skills with relevant queries

---

**Happy skill building!** 🚀
