# Anthropic Skills Archive Analysis

**Date:** December 29, 2024  
**Location:** `/media/laurie/Data/Github/mcp-cli-go/config/skills/`  
**Analysis Purpose:** Understand real-world skill patterns for MCP integration

---

## 📊 Skills Inventory

The Anthropic skills archive contains **17 skills** covering various domains:

### Production Skills from Anthropic

1. **algorithmic-art** - Generative art and algorithmic design
2. **brand-guidelines** - Brand identity and style guide enforcement
3. **canvas-design** - Canvas-based design work
4. **doc-coauthoring** - Document collaboration workflows
5. **docx** - Word document creation, editing, analysis
6. **frontend-design** - Production-grade web interface design
7. **internal-comms** - Internal communication templates
8. **mcp-builder** - Model Context Protocol server building
9. **pdf** - PDF manipulation and creation
10. **pptx** - PowerPoint creation and editing
11. **slack-gif-creator** - Slack GIF creation automation
12. **skill-creator** - Meta-skill for creating new skills
13. **theme-factory** - Design system and theme generation
14. **web-artifacts-builder** - Web artifact construction
15. **webapp-testing** - Web application testing workflows
16. **xlsx** - Excel spreadsheet operations

### Custom Skills

17. **python-best-practices** - Our test skill (created earlier)

---

## 🏗️ Skill Structure Patterns

### Pattern Analysis from Real Skills

After examining the archive, here are the actual patterns used:

#### 1. Simple Skills (Single SKILL.md)

**Example:** `frontend-design`

```
frontend-design/
├── LICENSE.txt
└── SKILL.md
```

**Characteristics:**
- Pure guidance and instructions
- No bundled resources
- Self-contained in SKILL.md
- Focused on workflows and patterns

**When to use:** When the skill provides procedural knowledge or guidance that doesn't require additional files.

#### 2. Skills with Scripts

**Example:** `web-artifacts-builder`

```
web-artifacts-builder/
├── LICENSE.txt
├── SKILL.md
└── scripts/
    └── [helper scripts]
```

**Characteristics:**
- Python/Bash scripts for deterministic operations
- SKILL.md references scripts
- Scripts are executable, not just examples
- Token-efficient (execute without loading into context)

**When to use:** When tasks require reliable, repeatable code execution.

#### 3. Skills with References

**Example:** `docx`

```
docx/
├── LICENSE.txt
├── SKILL.md
└── references/
    ├── docx-js.md
    ├── ooxml.md
    └── [other references]
```

**Characteristics:**
- Progressive disclosure pattern
- SKILL.md has overview and workflow
- References loaded only when needed
- Keeps SKILL.md lean (<500 lines)

**When to use:** When comprehensive documentation would bloat SKILL.md.

#### 4. Complex Skills (Multiple Resource Types)

**Example:** `skill-creator`

```
skill-creator/
├── LICENSE.txt
├── SKILL.md
├── references/
│   └── [documentation]
└── scripts/
    ├── init_skill.py
    └── package_skill.py
```

**Characteristics:**
- Both scripts AND references
- Multi-layered progressive disclosure
- Scripts for automation
- References for detailed guidance

**When to use:** For complex workflows requiring both automation and extensive documentation.

---

## 📋 YAML Frontmatter Patterns

### Required Fields

All skills use **only two required fields**:

```yaml
---
name: skill-name
description: What the skill does and when to use it
---
```

### Optional Fields Observed

```yaml
---
name: skill-name
description: ...
license: Complete terms in LICENSE.txt  # Common in Anthropic skills
---
```

**Key Finding:** No skills in the archive use `allowed-tools` field. This is an optional feature for restricting tool usage, but not commonly used in these production skills.

### Description Patterns

Excellent descriptions follow this pattern:

```yaml
description: "[What it does]. [When to use]. [Specific triggers]."
```

**Examples:**

**Good (docx):**
```yaml
description: "Comprehensive document creation, editing, and analysis with support for tracked changes, comments, formatting preservation, and text extraction. When Claude needs to work with professional documents (.docx files) for: (1) Creating new documents, (2) Modifying or editing content, (3) Working with tracked changes, (4) Adding comments, or any other document tasks"
```

**Good (frontend-design):**
```yaml
description: "Create distinctive, production-grade frontend interfaces with high design quality. Use this skill when the user asks to build web components, pages, artifacts, posters, or applications (examples include websites, landing pages, dashboards, React components, HTML/CSS layouts, or when styling/beautifying any web UI). Generates creative, polished code and UI design that avoids generic AI aesthetics."
```

**Pattern Elements:**
1. **What:** Clear statement of capabilities
2. **When:** Explicit usage contexts
3. **Triggers:** Specific keywords/phrases that should activate skill
4. **Examples:** Concrete use cases

---

## 💡 Key Insights from skill-creator

The `skill-creator` skill is meta - it teaches how to create skills. Key principles:

### 1. Progressive Disclosure (3-Level Loading)

```
Level 1: Metadata (name + description)
  ↓ Always in context (~100 words)

Level 2: SKILL.md body
  ↓ When skill triggers (<5k words, ideally <500 lines)

Level 3: Bundled resources
  ↓ As needed (unlimited - scripts can execute without context)
```

### 2. Conciseness is Critical

> "The context window is a public good. Skills share the context window with everything else Claude needs."

**Guidelines:**
- Default assumption: Claude is already very smart
- Only add context Claude doesn't have
- Challenge each piece of information
- Prefer concise examples over verbose explanations
- Keep SKILL.md under 500 lines

### 3. Resource Organization

**scripts/** - Executable code (Python/Bash)
- When: Same code repeatedly rewritten, or deterministic reliability needed
- Benefits: Token efficient, can execute without loading into context
- May still be read for patching or environment adjustments

**references/** - Documentation loaded as needed
- When: Information Claude should reference while working
- Examples: Database schemas, API docs, domain knowledge, policies
- Best practice: Include grep patterns in SKILL.md if files >10k words
- Avoid duplication: Info should be in SKILL.md OR references, not both

**assets/** - Files used in output (not loaded into context)
- When: Files needed in final output
- Examples: Templates, images, icons, boilerplate, fonts
- Use cases: Files that get copied or modified
- Benefits: Separates output resources from documentation

### 4. What NOT to Include

**Do NOT create:**
- README.md
- INSTALLATION_GUIDE.md
- QUICK_REFERENCE.md
- CHANGELOG.md
- Other auxiliary documentation

> "The skill should only contain information needed for an AI agent to do the job."

### 5. Reference Organization Patterns

**Pattern 1: Domain-specific organization**

```
bigquery-skill/
├── SKILL.md
└── reference/
    ├── finance.md
    ├── sales.md
    ├── product.md
    └── marketing.md
```

When user asks about sales metrics, Claude only reads `sales.md`.

**Pattern 2: Framework/variant organization**

```
cloud-deploy/
├── SKILL.md
└── references/
    ├── aws.md
    ├── gcp.md
    └── azure.md
```

User chooses AWS → Claude only reads `aws.md`.

**Pattern 3: Conditional details**

```markdown
# DOCX Processing

## Creating documents
Use docx-js. See [DOCX-JS.md](DOCX-JS.md).

## Editing documents
For simple edits, modify XML directly.
**For tracked changes**: See [REDLINING.md](REDLINING.md)
```

**Guidelines:**
- Avoid deeply nested references (keep one level from SKILL.md)
- For files >100 lines, include table of contents at top

---

## 🎯 Description Writing Best Practices

### The Formula (From Observation)

```
[CAPABILITY STATEMENT] + [USAGE CONTEXTS] + [SPECIFIC TRIGGERS/EXAMPLES]
```

### Elements to Include

1. **Capabilities** - What the skill can do
2. **Use cases** - When to use it (explicit contexts)
3. **Triggers** - Specific keywords, file types, or scenarios
4. **Examples** - Concrete examples of queries that should activate it

### Examples from Archive

**docx skill:**
```
"Comprehensive document creation, editing, and analysis with support for 
tracked changes, comments, formatting preservation, and text extraction. 
When Claude needs to work with professional documents (.docx files) for: 
(1) Creating new documents, (2) Modifying or editing content, 
(3) Working with tracked changes, (4) Adding comments, or any other document tasks"
```

**Breakdown:**
- Capabilities: "creation, editing, analysis", "tracked changes, comments, formatting, extraction"
- Use cases: "professional documents (.docx files)"
- Triggers: ".docx files", "creating", "modifying", "tracked changes", "comments"
- Specific contexts: Numbered list of when to use

**frontend-design skill:**
```
"Create distinctive, production-grade frontend interfaces with high design quality. 
Use this skill when the user asks to build web components, pages, artifacts, 
posters, or applications (examples include websites, landing pages, dashboards, 
React components, HTML/CSS layouts, or when styling/beautifying any web UI). 
Generates creative, polished code and UI design that avoids generic AI aesthetics."
```

**Breakdown:**
- Capabilities: "production-grade frontend", "high design quality"
- Use cases: "build web components, pages, artifacts, posters, applications"
- Triggers: "websites", "landing pages", "dashboards", "React", "HTML/CSS", "styling"
- Value proposition: "avoids generic AI aesthetics"

### What Makes Descriptions Effective

**✅ Good:**
- Specific capabilities and features
- Explicit "when to use" contexts
- Multiple trigger keywords
- Concrete examples
- Clear scope boundaries

**❌ Bad:**
- Vague ("helps with documents")
- No usage context
- Missing trigger keywords
- No examples
- Unclear boundaries

---

## 🔍 Workflow Patterns from DOCX Skill

The `docx` skill demonstrates sophisticated workflow design:

### Decision Tree Pattern

```markdown
## Workflow Decision Tree

### Reading/Analyzing Content
Use "Text extraction" or "Raw XML access" sections

### Creating New Document
Use "Creating a new Word document" workflow

### Editing Existing Document
- **Your own document + simple changes**
  Use "Basic OOXML editing" workflow
  
- **Someone else's document**
  Use **"Redlining workflow"** (recommended)
  
- **Legal/academic/business/gov docs**
  Use **"Redlining workflow"** (required)
```

**Pattern:** Start with decision tree, then provide detailed workflows for each path.

### Mandatory Reading Patterns

```markdown
1. **MANDATORY - READ ENTIRE FILE**: Read [`ooxml.md`](ooxml.md) (~600 lines) 
   completely from start to finish. **NEVER set any range limits when reading 
   this file.** Read the full file content for the Document library API...
```

**Pattern:** Explicit instructions to read supporting files when needed, with emphasis on reading completely.

### Step-by-Step Workflows

```markdown
### Tracked changes workflow

1. **Get markdown representation**: Convert document...
2. **Identify and group changes**: Review document...
3. **Read documentation and unpack**: MANDATORY READ...
4. **Implement changes in batches**: Group changes...
5. **Pack the document**: Convert back...
6. **Final verification**: Comprehensive check...
```

**Pattern:** Numbered steps with clear actions and validation points.

---

## 🚀 MCP Integration Strategy

Based on analysis of these real-world skills, here's the optimal MCP integration approach:

### Architecture: Two Modes

#### Mode 1: Passive Loading (Context Injection)

**Use case:** Skills that provide guidance, patterns, best practices

**How it works:**
1. MCP tool receives skill name
2. Loads SKILL.md content
3. Returns markdown to LLM as context
4. LLM uses guidance in conversation

**Example:**
```yaml
# MCP tool
tools:
  - name: load_skill
    description: Load a skill's instructions into context
    template: load_skill_passive
```

**Benefits:**
- Simple implementation
- Works like native Claude Skills
- Progressive disclosure supported
- No workflow execution

#### Mode 2: Active Execution (Workflow Execution)

**Use case:** Skills with scripts, deterministic operations, multi-step processes

**How it works:**
1. MCP tool receives skill name + input data
2. Executes workflow.yaml (if present)
3. Runs scripts, calls tools, orchestrates steps
4. Returns results to LLM

**Example:**
```yaml
# MCP tool
tools:
  - name: execute_skill_workflow
    description: Execute a skill's workflow
    template: execute_skill_workflow
```

**Benefits:**
- Can execute scripts
- Multi-step orchestration
- Multi-provider optimization
- Template composition

### Progressive Disclosure Implementation

**Level 1: Metadata (Always)**
```yaml
# MCP tool exposes
name: skill-name
description: "full description for discovery"
```

**Level 2: SKILL.md Body (When skill triggers)**
```yaml
steps:
  - name: load_main_content
    prompt: |
      Load: config/skills/{{skill_name}}/SKILL.md
```

**Level 3: Resources (As needed)**
```yaml
steps:
  - name: load_reference
    condition: "{{needs_reference}}"
    prompt: |
      Load: config/skills/{{skill_name}}/references/{{ref_file}}
```

### Directory Mapping

**Anthropic Format → MCP Server:**

```
config/skills/
├── skill-name/
│   ├── SKILL.md          → Passive: Return content
│   │                     → Active: Load + execute
│   ├── references/       → Progressive: Load on demand
│   ├── scripts/          → Active: Execute via workflow
│   └── assets/           → Active: Copy to output
```

### Tool Discovery

Each skill becomes an MCP tool:

```yaml
# Auto-generated from skills directory
tools:
  - name: docx_skill
    description: "[SKILL.md description field]"
    template: skill_loader
    input_schema:
      properties:
        mode:
          type: string
          enum: [passive, active]
          default: passive
```

**Discovery mechanism:**
- Scan `config/skills/` directory
- Read each SKILL.md YAML frontmatter
- Generate MCP tool definition
- Expose via runas config

---

## 📦 Implementation Plan

### Phase 1: Passive Skills (Week 1)

**Goal:** Load skills as context (like native Claude Skills)

**Implementation:**
1. Create skill scanner
   - Scan `config/skills/`
   - Parse YAML frontmatter
   - Extract name + description
   
2. Create `load_skill_passive` template
   - Load SKILL.md content
   - Support progressive disclosure
   - Return markdown to LLM

3. Create `skills-server.yaml` runas config
   - Auto-generate tools from scanned skills
   - One tool per skill
   - Use descriptions for discovery

**Deliverable:**
```bash
mcp-cli serve config/runas/skills-server.yaml
# Exposes all skills as MCP tools
# Works with any MCP client
# Skills load into context
```

### Phase 2: Progressive Disclosure (Week 2)

**Goal:** Implement 3-level loading

**Implementation:**
1. Enhance skill loader template
   - Level 1: Always expose name/description
   - Level 2: Load SKILL.md when triggered
   - Level 3: Load references when requested

2. Add reference detection
   - Parse SKILL.md for links like `[file.md](file.md)`
   - Track which references exist
   - Load on demand

3. Implement smart loading
   - Detect when Claude requests more detail
   - Load appropriate reference file
   - Return just what's needed

**Deliverable:**
```yaml
# Enhanced loader supports:
- Main content loading
- Reference file loading
- Conditional loading based on user needs
```

### Phase 3: Active Skills with Scripts (Week 3)

**Goal:** Execute workflows and scripts

**Implementation:**
1. Detect workflow.yaml presence
   - Check for `skills/skill-name/workflow.yaml`
   - Parse workflow specification
   - Determine execution mode

2. Create `execute_skill_workflow` template
   - Load workflow.yaml
   - Execute steps sequentially
   - Run scripts when needed
   - Return results

3. Script execution support
   - Python script runner
   - Bash script runner
   - Environment isolation
   - Error handling

**Deliverable:**
```yaml
# Skills can now:
- Execute Python/Bash scripts
- Run multi-step workflows
- Return execution results
```

### Phase 4: Tool Auto-Discovery (Week 4)

**Goal:** Automatic MCP tool generation

**Implementation:**
1. Create skill indexer
   - Scans skills directory on startup
   - Generates tool definitions
   - Updates when skills change

2. Auto-generate runas config
   - One tool per skill
   - Use skill descriptions
   - Support both passive/active modes

3. Hot reload support
   - Watch skills directory for changes
   - Regenerate tools on skill add/update
   - Notify connected clients

**Deliverable:**
```bash
# Add new skill:
mkdir config/skills/new-skill
# Write SKILL.md
# Skill automatically available as MCP tool
```

---

## 🎓 Critical Learnings

### 1. Description is Everything

The description field is THE critical discovery mechanism. It must:
- Explain what the skill does
- Specify when to use it
- Include trigger keywords
- Provide concrete examples

**This maps perfectly to MCP tool descriptions.**

### 2. Progressive Disclosure is Essential

Three-level loading keeps context efficient:
1. Metadata always in context (100 words)
2. Main content when skill triggers (<5k words)
3. Resources as needed (unlimited)

**MCP tools can implement this via conditional loading.**

### 3. Skills are Workflows, Not Just Docs

Real skills contain:
- Procedural knowledge (instructions)
- Executable scripts (automation)
- Reference material (specifications)
- Assets (templates, boilerplate)

**MCP workflows can execute all of these.**

### 4. Conciseness Matters

"The context window is a public good."

Skills must be:
- Lean and focused
- Challenge every piece of information
- Prefer examples over explanations
- Split content when >500 lines

**This aligns with efficient MCP tool design.**

### 5. No Auxiliary Files

Skills should NOT contain:
- README.md
- Installation guides
- Changelogs
- User documentation

**Only what the AI agent needs to execute.**

---

## 🔄 Next Steps

### Immediate (This Week)

1. **Create skill scanner script**
   ```go
   // internal/services/skill_scanner.go
   - Scan config/skills/
   - Parse YAML frontmatter
   - Generate tool definitions
   ```

2. **Create passive loader template**
   ```yaml
   # config/templates/load_skill_passive.yaml
   - Load SKILL.md
   - Support progressive disclosure
   - Return content
   ```

3. **Create skills-server runas config**
   ```yaml
   # config/runas/skills-server.yaml
   - Auto-generated tool list
   - One tool per skill
   ```

4. **Test with real skills**
   ```bash
   # Test with Anthropic skills
   mcp-cli serve config/runas/skills-server.yaml
   # Verify discovery
   # Test loading
   ```

### Short Term (Next Month)

1. Implement progressive disclosure
2. Add script execution support
3. Create active workflow mode
4. Build hot reload system

### Long Term (Next Quarter)

1. CLI commands (`mcp-cli skill list/install/create`)
2. Skill marketplace integration
3. Community skill repository
4. Skill authoring tools

---

## 📊 Summary

**Skills Analyzed:** 17 (16 Anthropic + 1 custom)  
**Patterns Identified:** 4 main structure types  
**Key Insight:** Skills are discoverable via description, load progressively, can be passive OR active  
**MCP Fit:** Perfect mapping - descriptions → tool discovery, content → passive mode, workflows → active mode  
**Implementation:** 4 phases over 4 weeks  
**Impact:** Universal skills that work with ANY LLM via MCP  

**The path is clear: Anthropic's skill format maps beautifully to MCP tools with two modes (passive loading + active execution).** 🚀

---

## 📁 Reference Files

All Anthropic skills available in:
```
/media/laurie/Data/Github/mcp-cli-go/config/skills/
```

Key skills to study:
- `skill-creator/` - Meta-skill with best practices
- `docx/` - Complex skill with references
- `frontend-design/` - Simple but effective
- `web-artifacts-builder/` - Skills with scripts

**Next:** Build the skill scanner and passive loader to prove the concept.
