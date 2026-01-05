# Anthropic Skills & MCP Integration

Skills are Exposed to LLMs through Model Context Protocol!

![](./img/Anthropic%20Skills.jpeg)

---

## 🎯 What are Anthropic Skills?

**Skills** are modular, self-contained packages that extend an LLM's capabilities by providing specialized knowledge, workflows, and tools. Think of them as "onboarding guides" for specific domains or tasks—they transform Claude from a general-purpose agent into a specialized agent equipped with procedural knowledge.

### Key Characteristics:

- **Modular**: Each skill focuses on a specific domain or task
- **Self-contained**: Includes everything needed to execute the skill
- **Progressive**: Loads content only when needed to save context
- **Discoverable**: Clear descriptions trigger skill activation

---

## 🏗️ Skill Structure (4 Types)

### 1. **Simple Skills** (Single SKILL.md)

```
skill-name/
├── LICENSE.txt
└── SKILL.md
```

**Use**: Pure guidance and instructions, no bundled resources

### 2. **Skills with Scripts**

```
skill-name/
├── LICENSE.txt
├── SKILL.md
└── scripts/
    └── helper_script.py
```

**Use**: Tasks requiring reliable, repeatable code execution

### 3. **Skills with References**

```
skill-name/
├── LICENSE.txt
├── SKILL.md
└── references/
    ├── api_docs.md
    └── schemas.md
```

**Use**: Comprehensive documentation loaded only when needed

### 4. **Complex Skills** (Multiple Resources)

```
skill-name/
├── LICENSE.txt
├── SKILL.md
├── references/
│   └── documentation.md
└── scripts/
    └── automation.py
```

**Use**: Complex workflows requiring both automation and extensive documentation

---

## 🔄 Progressive Disclosure (3-Level Loading)

### **Level 1: Metadata** (Always in context)

- **Name**: Skill identifier
- **Description**: What it does + when to use it
- **Size**: ~100 words
- **Purpose**: Skill discovery and triggering

### **Level 2: SKILL.md Body** (When skill triggers)

- **Content**: Core instructions and workflows
- **Size**: <5k words (ideally <500 lines)
- **Purpose**: Main skill execution guidance

### **Level 3: Bundled Resources** (As needed)

- **Scripts**: Executable code (Python/Bash)
- **References**: Detailed documentation
- **Assets**: Templates, images, fonts
- **Size**: Unlimited (scripts execute without context)
- **Purpose**: Support complex operations

---

## 🛠️ MCP Integration Architecture

### **Two Integration Modes:**

#### **Mode 1: Passive Loading** (Context Injection)

```
User Request → MCP Tool → Load SKILL.md → Return as Context → LLM Uses Guidance
```

**Use Case**: Skills providing guidance, patterns, best practices
**Benefits**: Simple, works like native Claude Skills

#### **Mode 2: Active Execution** (Workflow Execution)

```
User Request → MCP Tool → Execute workflow.yaml → Run Scripts → Return Results
```

**Use Case**: Skills with scripts, deterministic operations
**Benefits**: Can execute scripts, multi-step orchestration

---

## 🔗 How Skills Become MCP Tools

### **Auto-Discovery Process:**

1. **Scan Directory**: MCP-CLI in serve mode scans `config/skills/`
2. **Parse Frontmatter**: Extracts name + description from each SKILL.md
3. **Generate Tools**: Creates one MCP tool per skill
4. **Expose via MCP**: Tools available to any MCP client

### **Tool Definition Example:**

```yaml
tools:
  - name: frontend_design_skill
    description: "Create distinctive, production-grade frontend interfaces..."
    input_schema:
      properties:
        mode:
          type: string
          enum: [passive, active]
          default: passive
```

---

## 📊 Real-World Skill Examples

### **From Anthropic Skills Archive:**

1. **frontend-design** - Production-grade web interfaces
2. **docx** - Word document creation/editing
3. **mcp-builder** - MCP server development
4. **skill-creator** - Meta-skill for creating skills
5. **web-artifacts-builder** - Web artifact construction
6. **pdf** - PDF manipulation and creation
7. **xlsx** - Excel spreadsheet operations
8. **algorithmic-art** - Generative art and design

### **Skill Description Pattern:**

```
[CAPABILITY STATEMENT] + [USAGE CONTEXTS] + [SPECIFIC TRIGGERS/EXAMPLES]
```

**Example (frontend-design):**

> "Create distinctive, production-grade frontend interfaces with high design quality. Use this skill when the user asks to build web components, pages, artifacts, posters, or applications (examples include websites, landing pages, dashboards, React components, HTML/CSS layouts, or when styling/beautifying any web UI)."

---

## 🚀 MCP Server Implementation

### **Directory Mapping:**

```
config/skills/
├── skill-name/
│   ├── SKILL.md          → Passive: Return content
│   │                     → Active: Load + execute
│   ├── references/       → Progressive: Load on demand
│   ├── scripts/          → Active: Execute via workflow
│   └── assets/           → Active: Copy to output
```

---

## 💡 Key Design Principles

### **1. Conciseness is Critical**

> "The context window is a public good."

- Default assumption: LLMs are already very smart
- Only add context LLMs don't have
- Challenge each piece of information
- Keep SKILL.md under 500 lines

### **2. Progressive Disclosure**

- Three-level loading keeps context efficient
- Metadata always in context (100 words)
- Main content when skill triggers (<5k words)
- Resources as needed (unlimited)

### **3. Skills are Workflows, Not Just Docs**

Real skills contain:

- Procedural knowledge (instructions)
- Executable scripts (automation)
- Reference material (specifications)
- Assets (templates, boilerplate)

### **4. No Auxiliary Files**

Skills should NOT contain:

- README.md
- Installation guides
- Changelogs
- User documentation

> "Only what the AI agent needs to execute."

---

## 🎯 Benefits of MCP Integration

### **For LLM Developers:**

- **Universal Access**: Skills work with ANY LLM via MCP
- **Standardized Interface**: Consistent tool discovery and execution
- **Progressive Loading**: Efficient context management
- **Hot Reload**: Skills update without restart

### **For Skill Authors:**

- **Write Once, Run Anywhere**: Skills work across all MCP clients
- **Automatic Discovery**: Skills auto-register as MCP tools
- **Flexible Execution**: Both passive and active modes
- **Community Sharing**: Distribute skills via MCP servers

### **For End Users:**

- **Seamless Experience**: Skills feel like native capabilities
- **Context-Aware**: Right information at the right time
- **Reliable Execution**: Scripts ensure deterministic results
- **Discoverable**: Clear descriptions help find needed skills

---

## 🔄 Workflow Example: DOCX Skill

### **Decision Tree Pattern:**

```
## Workflow Decision Tree

### Reading/Analyzing Content
Use "Text extraction" or "Raw XML access"

### Creating New Document
Use "Creating a new Word document" workflow

### Editing Existing Document
- Your own document + simple changes
  Use "Basic OOXML editing"

- Someone else's document
  Use "Redlining workflow" (recommended)

- Legal/academic/business/gov docs
  Use "Redlining workflow" (required)
```

### **Mandatory Reading Patterns:**

```markdown
1. **MANDATORY - READ ENTIRE FILE**: Read [`ooxml.md`](ooxml.md) (~600 lines) 
   completely from start to finish. **NEVER set any range limits when reading 
   this file.** Read the full file content for the Document library API...
```

---

## 🎓 Summary

**Anthropic Skills** are discoverable via description, load progressively, and can be passive OR active. **MCP integration** provides a perfect mapping: descriptions → tool discovery, content → passive mode, workflows → active mode.

**The path is clear**: Anthropic's skill format maps beautifully to MCP tools with two modes (passive loading + active execution), enabling universal skills that work with ANY LLM via MCP.

---

## 📚 Documentation

### Getting Started
- **[Quick Start](quick-start.md)** - Get up and running in 5 minutes
- **[Skills Auto-Loading Guide](auto-loading.md)** - Complete guide to automatic skill discovery and MCP exposure
- **[Quick Reference](quick-reference.md)** - Fast lookup for common tasks

### Understanding Skills
- **[Overview](overview.md)** - How skills work and why they're different
- **[Why Skills Matter](WHY_SKILLS_MATTER.md)** - The philosophy behind skills

### Execution & Security
- **[Docker/Podman Execution](docker-podman-execution.md)** - Complete guide to containerized code execution

### Creating Skills
- **[Creating Skills](creating-skills.md)** - Build your own skills

### Reference
- **Skills Archive**: `/config/skills/`
- **MCP Specification**: https://modelcontextprotocol.io/

---

## 🎯 Quick Links

**New to skills?** Start with [Quick Start](quick-start.md)

**Want auto-discovery?** See [Auto-Loading Guide](auto-loading.md)

**Need Docker/Podman help?** Read [Docker/Podman Execution](docker-podman-execution.md)

**Building skills?** Read [Creating Skills](creating-skills.md)

**Understanding the system?** Check [Overview](overview.md)

---

## ⚠️ Production Status

**Production-Ready:**
- ✅ Skills auto-discovery and MCP tool generation
- ✅ Passive mode (load documentation)
- ✅ `execute_skill_code` with Docker/Podman sandboxing
- ✅ PYTHONPATH configuration for skill libraries

**Experimental/Not Implemented:**
- ⚠️ Active mode with `workflow.yaml` (stub only)
- ⚠️ Workflow-based orchestration

**Recommended:** Use `execution_mode: auto` for full functionality.

---

**Last Updated**: January 4, 2026  
**Based on Analysis of**: 17 real-world Anthropic skills  
**MCP Integration Status**: ✅ Fully implemented with auto-loading