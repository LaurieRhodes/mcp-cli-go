# Skills Container Images

Docker/Podman images for skills requiring Python packages beyond the standard library.  These are required for exposing Anthropic Skills to LLMs which have an active scripting component associated.

For security, Anthropic deliberately sandboxes Skills rather than allowing execution on a host machine.

Following this concept, Skills sandboxing allows us to create custom containers for any type of specific activity we may wish to expose in the environment.

## Quick Start

```bash
# Build all images
./build-skills-images.sh

# Build specific image
./build-skills-images.sh pptx
```

## Available Images

Built by `build-skills-images.sh`:

### Individual Skills

**mcp-skills-docx** (~170MB)

- Packages: `defusedxml`, `lxml`
- For: Word document manipulation

**mcp-skills-pptx** (~190MB)

- Packages: `python-pptx`, `Pillow`, `lxml`
- For: PowerPoint presentations

**mcp-skills-xlsx** (~175MB)

- Packages: `openpyxl`, `lxml`
- For: Excel spreadsheets (basic)

**mcp-skills-pdf** (~220MB)

- Packages: `pypdf`, `pdf2image`, `Pillow`, `pdfplumber`, `pytesseract`, `poppler-utils`, `tesseract-ocr`
- For: PDF manipulation, forms, text extraction, OCR

### Combined Images

**mcp-skills-office** (~195MB) - Recommended

- Packages: All of docx + pptx + xlsx
- For: Word, PowerPoint, and Excel in one image

**mcp-skills-xlsx-libreoffice** (~350MB)

- Packages: openpyxl + LibreOffice Calc
- For: Excel with formula recalculation support

## Default Image

Skills without a custom image mapping use: `python:3.11-alpine`

## Image Mapping

Configure in `config/skills/skill-images.yaml in the configuration subdirectory of mcp-cli`:

```yaml
skills:
  docx: mcp-skills-docx
  pptx: mcp-skills-pptx
  xlsx: mcp-skills-xlsx
  pdf: mcp-skills-pdf
```

These container definitions align with the public skills demonstrated by Anthropic in their archive at:  [skills/skills at main · anthropics/skills · GitHub](https://github.com/anthropics/skills/tree/main/skills)

Configuring  `config/skills/skill-images.yaml in the configuration subdirectory of mcp-cli allows selective containers to be run when invoked to support the particular skill being used.

Only skills that have an execution component (active) require containers.  Many (passive) skills simply provide context to the LLM for completing a task.

## Building Images

```bash
cd docker/skills

# Build all
./build-skills-images.sh

# Build specific
./build-skills-images.sh office

# List available images
docker images | grep mcp-skills
```

## Testing

```bash
# Test image has required packages
docker run --rm mcp-skills-pptx python -c "from pptx import Presentation; print('OK')"

# Test Excel
docker run --rm mcp-skills-xlsx python -c "from openpyxl import Workbook; print('OK')"
```

## Image Details

All images based on `python:3.11-slim`.

**Build time:**

- First build: 5-10 minutes (downloading packages)
- Subsequent: 1-2 minutes (cached layers)

**Maintenance:**

- Images are rebuilt locally as needed
- No external registry dependencies
- Package versions specified in Dockerfiles

---

Last updated: January 6, 2026
