# RAG Troubleshooting Guide

## Common Issues

### "Found 0 results"

**Possible Causes:**
- Empty database
- Invalid embeddings (dummy 0.1 vectors)
- Threshold too strict
- Wrong vector column name

**Solutions:**
```bash
# Verify API key
echo $OPENAI_API_KEY

# Test database has data
mcp-cli query --server pgvector "Count rows"

# Try broader search
mcp-cli rag search "test" --top-k 10
```

### "embedding service not available"

**Solution:**
```bash
# Set API key
export OPENAI_API_KEY="sk-your-key-here"

# Verify config uses "service" type
cat config/rag/pgvector.yaml
# Should have: type: service
```

### Slow Performance

**Solutions:**
- Reduce `--top-k` (use 3 instead of 20)
- Use single strategy
- Make queries more specific

## Debug Commands

```bash
# Check configuration
mcp-cli rag config

# Enable debug logging
mcp-cli rag search "query" --log-level debug

# Test components
mcp-cli servers list | grep pgvector
```

## Validation Checklist

- [ ] `$OPENAI_API_KEY` is set
- [ ] `mcp-cli rag config` shows configuration
- [ ] Database has data
- [ ] Direct search works: `mcp-cli rag search "test"`

## Getting Help

Include when reporting issues:
1. Error message
2. Debug logs: `--log-level debug`
3. Configuration: `mcp-cli rag config`
4. What you tried

## Quick Fix

```bash
# Reset to working state
export OPENAI_API_KEY="sk-your-key"
mcp-cli rag search "test" --server pgvector --top-k 3
```
