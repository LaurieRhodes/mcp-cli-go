# Tasks SEP-1686 Quick Reference

## TL;DR

Tasks enable **non-blocking execution** of long-running tools:

```
Old: tools/call → [blocks 30 mins] → result
New: tools/call + task → taskId [1ms] → poll → result
```

---

## Quick Start

### 1. Create Task (Non-Blocking)

```json
{
  "method": "tools/call",
  "params": {
    "name": "long_workflow",
    "arguments": {"input": "data"},
    "task": {"ttl": 1800000}  // 30 minutes
  }
}
```

**Response (immediate):**
```json
{
  "result": {
    "task": {
      "taskId": "abc-123",
      "status": "working",
      "pollInterval": 5000
    }
  }
}
```

### 2. Poll Status

```json
{
  "method": "tasks/get",
  "params": {"taskId": "abc-123"}
}
```

**Response:**
```json
{
  "result": {
    "task": {
      "status": "working",  // or "completed", "failed"
      "lastUpdatedAt": "2026-01-24T13:05:00Z"
    }
  }
}
```

### 3. Get Result (Blocks Until Complete)

```json
{
  "method": "tasks/result",
  "params": {"taskId": "abc-123"}
}
```

**Response:**
```json
{
  "result": {
    "content": [{"type": "text", "text": "Success!"}]
  }
}
```

---

## All Operations

| Operation | Method | Params | Blocks | Use Case |
|-----------|--------|--------|--------|----------|
| **Create** | `tools/call` | `{name, arguments, task}` | No | Start task |
| **Status** | `tasks/get` | `{taskId}` | No | Check progress |
| **Result** | `tasks/result` | `{taskId}` | Yes* | Get output |
| **List** | `tasks/list` | `{cursor?}` | No | View all tasks |
| **Cancel** | `tasks/cancel` | `{taskId}` | No | Stop task |

*Blocks until task reaches terminal state

---

## Task States

```
┌─────────┐
│ working │ (initial)
└────┬────┘
     │
     v
┌─────────┐   ┌────────┐   ┌──────────┐
│completed│   │ failed │   │cancelled │
└─────────┘   └────────┘   └──────────┘
(terminal)    (terminal)    (terminal)
```

---

## Configuration (mcp-cli serve)

**Defaults:**
```go
Default TTL:  30 minutes
Max TTL:      2 hours
Poll Interval: 5 seconds
Cleanup:      Every 1 minute
```

**Customize:**
```go
// cmd/serve.go
taskManager := tasks.NewManager(
    1*time.Hour,   // Default TTL
    4*time.Hour,   // Max TTL
    10000,         // Poll interval (ms)
)
```

---

## Capability Check

**During initialize:**
```json
{
  "result": {
    "capabilities": {
      "tasks": {
        "requests": {"tools/call": true},
        "list": true,
        "cancel": true
      }
    }
  }
}
```

---

## Error Handling

### Task Not Found
```json
{
  "error": {
    "code": -32603,
    "message": "task not found: abc-123"
  }
}
```

**Causes:** Invalid ID, expired TTL, wrong server

### Tool Error
```json
{
  "result": {
    "task": {
      "status": "failed",
      "statusMessage": "Tool execution failed: ..."
    }
  }
}
```

---

## Best Practices

### When to Use Tasks

✅ **Use tasks for:**
- Workflows > 2 minutes
- User-initiated long operations
- Concurrent execution
- Resumable operations

❌ **Don't use tasks for:**
- Quick lookups (< 5 seconds)
- Real-time interactions
- Operations requiring immediate feedback

### Polling Strategy

```javascript
// Good: Exponential backoff
async function pollTask(taskId) {
  let interval = 5000; // Start with suggested interval
  
  while (true) {
    const status = await tasks_get(taskId);
    
    if (status.task.status !== "working") {
      return status;
    }
    
    await sleep(interval);
    interval = Math.min(interval * 1.2, 30000); // Max 30s
  }
}
```

### TTL Selection

```javascript
// Workflow duration estimates
quick_analysis:   ttl: 60000     // 1 minute
medium_research:  ttl: 600000    // 10 minutes
deep_research:    ttl: 1800000   // 30 minutes
data_migration:   ttl: 3600000   // 1 hour
```

---

## Common Patterns

### Fire-and-Forget
```javascript
// Start task, don't wait
const {task} = await tools_call({
  name: "batch_processor",
  task: {ttl: 3600000}
});

// Do other work...
// Check later
const result = await tasks_result(task.taskId);
```

### Progress Display
```javascript
// Show progress to user
const {task} = await tools_call({...});

while (true) {
  const status = await tasks_get(task.taskId);
  console.log(`Status: ${status.task.status}`);
  
  if (status.task.status !== "working") break;
  await sleep(5000);
}
```

### Multiple Concurrent Tasks
```javascript
// Start multiple tasks
const tasks = await Promise.all([
  tools_call({name: "analyze_a", task: {ttl: 600000}}),
  tools_call({name: "analyze_b", task: {ttl: 600000}}),
  tools_call({name: "analyze_c", task: {ttl: 600000}})
]);

// Wait for all
const results = await Promise.all(
  tasks.map(t => tasks_result(t.task.taskId))
);
```

---

## Troubleshooting

### Task Stuck in "working"

**Debug:**
```bash
# Check server logs
grep "task abc-123" /var/log/mcp-cli.log

# List all tasks
echo '{"method":"tasks/list"}' | mcp-cli serve config.yaml
```

**Fix:**
- Check TTL hasn't expired
- Verify tool isn't actually running
- Use tasks/cancel to force cleanup

### High Memory Usage

**Cause:** Too many tasks in memory

**Fix:**
```go
// Reduce TTL defaults
taskManager := tasks.NewManager(
    10*time.Minute,  // Shorter default
    30*time.Minute,  // Lower max
    5000,
)
```

---

## CLI Examples

### Start Task
```bash
echo '{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "workflow_name",
    "arguments": {},
    "task": {"ttl": 1800000}
  }
}' | mcp-cli serve workflows.yaml
```

### Check Status
```bash
echo '{
  "jsonrpc": "2.0",
  "id": 2,
  "method": "tasks/get",
  "params": {"taskId": "abc-123"}
}' | mcp-cli serve workflows.yaml
```

### Get Result
```bash
echo '{
  "jsonrpc": "2.0",
  "id": 3,
  "method": "tasks/result",
  "params": {"taskId": "abc-123"}
}' | mcp-cli serve workflows.yaml
```

---

## Go SDK Example (Hypothetical)

```go
// Create task
task, err := client.CallToolWithTask("workflow", args, 30*time.Minute)
if err != nil {
    log.Fatal(err)
}

// Poll for completion
for {
    status, err := client.GetTask(task.TaskID)
    if err != nil {
        log.Fatal(err)
    }
    
    if status.Status != "working" {
        break
    }
    
    time.Sleep(5 * time.Second)
}

// Get result
result, err := client.GetTaskResult(task.TaskID)
```

---

## Comparison

### Before Tasks
```
User: "Analyze this dataset"
Claude: [sends tools/call]
         [waits 30 minutes]
         [timeout - failure]
User: "Did it finish?"
Claude: "I don't know, it timed out"
```

### With Tasks
```
User: "Analyze this dataset"
Claude: [sends task-augmented tools/call]
         [receives taskId immediately]
         "I started the analysis (task abc-123)"
         [polls every 5 seconds]
         "Still working... (5 minutes elapsed)"
         [task completes]
         "Analysis complete! Here are the results..."
```

---

## Quick Reference Card

```
╔══════════════════════════════════════════════╗
║            TASKS SEP-1686                    ║
╠══════════════════════════════════════════════╣
║ CREATE:  tools/call + task → taskId (1ms)   ║
║ STATUS:  tasks/get(id) → status             ║
║ RESULT:  tasks/result(id) → output (blocks) ║
║ LIST:    tasks/list() → all tasks           ║
║ CANCEL:  tasks/cancel(id) → cancelled       ║
╠══════════════════════════════════════════════╣
║ States: working → completed/failed/cancelled ║
║ TTL: 30min default, 2hr max                 ║
║ Poll: Every 5 seconds (suggested)           ║
╚══════════════════════════════════════════════╝
```

---

## Further Reading

- **Full Docs:** [tasks-sep-1686.md](tasks-sep-1686.md)
- **SEP Spec:** [github.com/modelcontextprotocol/specification/issues/1686](https://github.com/modelcontextprotocol/specification/issues/1686)
- **Implementation:** [TASKS-IMPLEMENTATION-SUMMARY.md](TASKS-IMPLEMENTATION-SUMMARY.md)

---

**Quick Start:** Add `"task": {"ttl": 1800000}` to any tools/call → Get taskId → Poll with tasks/get → Retrieve with tasks/result ✅
