# Troubleshooting

## Files Not Appearing

Check:

```bash
# Directory exists
ls -ld ~/outputs

# Configuration matches
grep OutputsDir internal/sandbox/executor.go

# mcp-cli rebuilt after config change
```

Verify skill code uses `/outputs/` not `/workspace/`.

## Images Not Built

```bash
cd docker/skills
./build-skills-images.sh
```

Verify:

```bash
docker images | grep mcp-skills
```

## Permission Denied

```bash
chmod 755 ~/outputs
```

---

Last updated: January 6, 2026
