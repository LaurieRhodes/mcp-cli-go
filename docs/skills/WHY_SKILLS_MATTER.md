# Why Skills Matter

## The Problem

LLMs have knowledge but lack:
- Specialized libraries (python-pptx, openpyxl)
- Safe execution environments
- File persistence mechanisms

## The Solution

Skills provide:
1. **Helper libraries** - Pre-installed packages in containers
2. **Documentation** - Instructions for using libraries
3. **Safe execution** - Isolated containers with no network
4. **File persistence** - `/outputs` directory mounted from host

## How It Works

```
User: "Create a PowerPoint about sales"
    ↓
LLM reads pptx skill documentation
    ↓
LLM writes code using python-pptx library
    ↓
Code executes in isolated container
    ↓
File persists in /outputs on host
```

## Benefits

### For Users
- Natural language requests
- No command-line syntax to learn
- Files appear on their filesystem

### For LLMs
- Access to specialized libraries
- Clear usage documentation
- Standardized execution environment

### For Developers
- Reusable helper libraries
- Secure execution
- Cross-LLM compatibility (GPT-4, DeepSeek, Gemini, Claude)

## Example: Creating a Presentation

**Without skills:**
```
User: "Create a PowerPoint"
LLM: "I can't create files, but here's the content..."
```

**With skills:**
```
User: "Create a PowerPoint about Q4 sales"
LLM: 
1. Reads pptx skill documentation
2. Writes code:
   from pptx import Presentation
   prs = Presentation()
   slide = prs.slides.add_slide(...)
   prs.save('/outputs/q4-sales.pptx')
3. Executes in container
4. File appears at ~/outputs/q4-sales.pptx
```

## Architecture

```
Documentation → LLM → Custom Code → Container → Persistent File
```

Not:
```
Fixed Script → Limited Options → Temporary Result
```

## Cross-LLM Compatibility

The same skills work with:
- OpenAI GPT-4
- DeepSeek
- Google Gemini
- Anthropic Claude
- Moonshot Kimi

All via the Model Context Protocol (MCP).

---

Last updated: January 6, 2026
