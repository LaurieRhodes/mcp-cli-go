# GraphRAG Showcase - Quick Reference Index

**Production-Ready GraphRAG for Australian ISM Policy**

## 🎯 Start Here

| What do you want to do? | Go to |
|-------------------------|-------|
| 📖 **Learn what this is** | [README.md](README.md) |
| 🏗️ **Understand the architecture** | [ARCHITECTURE.md](ARCHITECTURE.md) |
| 📊 **See results and metrics** | [RESULTS.md](RESULTS.md) |
| 🚀 **Start using the system** | [USAGE.md](USAGE.md) |
| 📋 **See what's included** | [MANIFEST.md](MANIFEST.md) |
| 📚 **Browse all contents** | [SHOWCASE_CONTENTS.md](SHOWCASE_CONTENTS.md) |

## 🔍 Find Specific Content

### Workflows
- **All workflows** → [workflows/](workflows/) (8 YAML files)
- **Production workflow** → [workflows/main_production_graphrag.yaml](workflows/main_production_graphrag.yaml)
- **Workflow docs** → [workflows/README.md](workflows/README.md)

### Scripts & Code
- **Python scripts** → [skills/](skills/) (3 scripts)
- **Helper scripts** → [scripts/](scripts/) (3 bash scripts)
- **Script docs** → [skills/README.md](skills/README.md), [scripts/README.md](scripts/README.md)

### Examples
- **Query examples** → [examples/queries/](examples/queries/) (3 queries)
- **Output examples** → [example_outputs/](example_outputs/) (6 files)
- **Real graph data** → [example_outputs/knowledge_graph_structure.json](example_outputs/knowledge_graph_structure.json)

### Documentation
- **Extended guides** → [documentation/](documentation/) (5 guides)
- **Validation example** → [validation/CHUNK-305_validation.md](validation/CHUNK-305_validation.md)

## ⚡ Quick Actions

### Query the Graph
```bash
./scripts/ask_graphrag_direct.sh "password requirements"
```

### Explore Graph Structure
```bash
./scripts/explore_graphrag.sh stats
```

### Validate Extraction
```bash
./scripts/validate_extraction.sh CHUNK-305
```

### View Example Output
```bash
cat example_outputs/knowledge_graph_structure.json | jq '.metadata'
```

## 📊 Key Statistics

| Metric | Value |
|--------|-------|
| Chunks Processed | 531 |
| Entities Extracted | 2,337 |
| Relationships Mapped | 2,313 |
| Entity Types | 100+ |
| Query Response Time | <1 second |
| Total Cost | ~$8-10 |
| Source Preservation | 100% |

## 🎓 Learning Paths

### Beginner (30 minutes)
1. Read [README.md](README.md)
2. Browse [example_outputs/](example_outputs/)
3. Read [validation/CHUNK-305_validation.md](validation/CHUNK-305_validation.md)

### Intermediate (2 hours)
1. Study [ARCHITECTURE.md](ARCHITECTURE.md)
2. Review [workflows/](workflows/)
3. Inspect [skills/](skills/)
4. Read [RESULTS.md](RESULTS.md)

### Advanced (1 day)
1. Deep dive [USAGE.md](USAGE.md)
2. Run all workflows
3. Customize for your data
4. Build integrations

## 📁 File Count

- **Documentation:** 11 markdown files
- **Workflows:** 8 YAML files
- **Python Scripts:** 3 files
- **Bash Scripts:** 3 files
- **Example Outputs:** 6 JSON files
- **Example Queries:** 3 JSON files

**Total:** 34 files, ~240KB

## 🎯 Use Cases Demonstrated

- ✅ Policy document analysis
- ✅ Entity extraction from technical docs
- ✅ Knowledge graph construction
- ✅ Semantic search and querying
- ✅ Multi-hop relationship traversal
- ✅ Source traceability
- ✅ Quality validation
- ✅ Production deployment

## 🔧 Requirements

**To Run:**
- Python 3.8+
- jq (JSON processor)
- Data files in `/tmp/mcp-outputs/rlm_poc/`

**To Build:**
- DeepSeek API key
- MCP-CLI-Go framework
- Source ISM document

## ✨ Highlights

**Architecture:**
- Three-layer design (source/extraction/graph)
- Complete traceability
- 100% source preservation

**Performance:**
- <1 second queries
- Parallel extraction (20 workers)
- In-memory graph loading

**Quality:**
- Validated extraction accuracy
- No hallucinations detected
- Full audit trail

**Cost:**
- One-time build: $8-10
- Unlimited queries: $0
- 60% savings vs alternatives

## 📞 Support

**Questions about the showcase?**
- Check [USAGE.md](USAGE.md) for how-to guides
- Review [ARCHITECTURE.md](ARCHITECTURE.md) for design decisions
- See [RESULTS.md](RESULTS.md) for metrics and analysis

**Want to adapt this?**
- Workflows are in [workflows/](workflows/)
- Scripts are in [skills/](skills/) and [scripts/](scripts/)
- Examples are in [examples/](examples/) and [example_outputs/](example_outputs/)

---

**Quick Reference Index**  
**Version:** 1.0  
**Updated:** February 2026
