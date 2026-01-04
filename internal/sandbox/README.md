# Sandbox Package

Provides cross-platform Docker/Podman-based sandboxing for executing skill scripts.

## Architecture

```
Executor Interface
├── NativeExecutor    (for native deployments)
│   └── Uses: os/exec + docker/podman CLI
└── DooDockerExecutor (for containerized deployments)
    └── Uses: Docker/Podman API client + socket mount
```

## Auto-Detection

The `DetectExecutor()` function automatically chooses the right executor:

1. **Running in container** → Use `DooDockerExecutor` (Docker-out-of-Docker)
2. **Running natively** → Use `NativeExecutor` (Docker/Podman CLI)

### Podman Support

**Native mode:** Automatically detects and uses Podman if Docker isn't available
**Socket mode:** Tries multiple socket locations:
- `/var/run/docker.sock` (Docker)
- `/run/user/$UID/podman/podman.sock` (Podman rootless)
- `/run/podman/podman.sock` (Podman rootful)

## Usage

```go
import "github.com/LaurieRhodes/mcp-cli-go/internal/sandbox"

// Create executor with default config
config := sandbox.DefaultConfig()
executor, err := sandbox.DetectExecutor(config)
if err != nil {
    log.Fatal("Docker/Podman not available:", err)
}

// Check availability
if !executor.IsAvailable() {
    log.Warn("Docker/Podman not available")
    return
}

// Execute Python script
ctx := context.Background()
output, err := executor.ExecutePython(
    ctx,
    "/path/to/skill",
    "scripts/document.py",
    []string{"input.docx"},
)
```

## Security Features

All executors apply strict security constraints:

- **Filesystem:** Read-only root, read-only skill mount
- **Network:** Disabled (`--network=none`)
- **Resources:** Memory limit (256MB), CPU limit (0.5 cores)
- **Processes:** Limit 100 processes
- **Privileges:** No privilege escalation, all capabilities dropped
- **Timeout:** 30 seconds default

## Deployment Modes

### Native Deployment

```bash
# User runs mcp-cli directly on their machine
mcp-cli serve config/runas/skills-auto.yaml

# Uses NativeExecutor
# → Spawns Docker containers as siblings
```

### Containerized Deployment

```bash
# User runs mcp-cli in Docker/Podman
docker-compose up

# Uses DooDockerExecutor  
# → Talks to host Docker/Podman via socket mount
# → Spawns containers as siblings on host
```

### Podman-Specific Setup

**For native deployment (easiest):**
```bash
# No setup needed! Code automatically detects Podman
go build -o mcp-cli .
./mcp-cli serve config/runas/skills-auto.yaml
```

**For containerized deployment:**
```bash
# Enable Podman socket
systemctl --user enable --now podman.socket

# Run with docker-compose (podman-compose also works)
docker-compose up  # or: podman-compose up
```

**Optional: Create docker alias:**
```bash
# Add to ~/.bashrc
alias docker='podman'
```

## Configuration

```go
config := sandbox.ExecutorConfig{
    PythonImage: "python:3.11-alpine",
    Timeout:     30 * time.Second,
    MemoryLimit: "256m",
    CPULimit:    "0.5",
}
```

## Supported Platforms

- ✅ Windows (Docker Desktop)
- ✅ Linux (Docker or Podman)
- ✅ macOS (Docker Desktop)
- ✅ Container (DooD with Docker or Podman)

## Dependencies

- `github.com/fsouza/go-dockerclient` - Docker/Podman API client (for DooD executor)
- Standard library `os/exec` - For native executor
- **No Docker required if Podman is installed**
