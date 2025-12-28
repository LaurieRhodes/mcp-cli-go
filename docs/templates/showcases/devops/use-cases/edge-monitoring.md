# Edge-Deployed Monitoring

> **Template:** [edge_health_monitor.yaml](../templates/edge_health_monitor.yaml)  
> **Workflow:** Collect Metrics → Analyze Locally → Escalate if Critical  
> **Best For:** Distributed environments, edge locations, air-gapped networks

---

## Problem Description

### The Challenge

**Modern infrastructure is distributed:**
- Edge data centers in remote locations
- Factory IoT devices
- Retail store systems
- Mobile edge computing
- Air-gapped secure environments
- Developer laptops

**Traditional AI frameworks don't work:**
```
Typical Python AI Framework Requirements:
├── Python 3.11 runtime (100MB)
├── Dependencies (pip packages) (400MB+)
├── Vector database (Pinecone/Chroma)
├── Application server (FastAPI)
├── Message queue (RabbitMQ)
└── Total: 1GB+ deployment

Installation process:
1. Install Python
2. Set up virtual environment
3. Install 50+ pip packages
4. Debug dependency conflicts
5. Configure external services
6. Deploy and monitor
```

**Problems for edge deployment:**
- **Limited resources:** Edge devices have constrained CPU/memory
- **Intermittent connectivity:** Can't rely on cloud APIs
- **Complex setup:** Dependencies fail in restricted environments
- **Large footprint:** Bandwidth-limited locations
- **Air-gapped:** Some environments have zero internet access

### Why This Matters

**Edge monitoring requirements:**
- Local intelligence (analyze on-device)
- Minimal resource usage
- Works offline
- Fast deployment
- No dependency hell
- Simple updates

**Use cases:**
- **Remote data centers:** Monitor health without cloud roundtrip
- **Factory floor:** Analyze equipment metrics locally
- **Retail locations:** Detect anomalies in store systems
- **Development:** Full AI capabilities on laptop
- **Air-gapped:** Secure environments with no external access

---

## Template Solution

### What It Does

**mcp-cli enables edge deployment through:**

1. **Single 20MB binary** - No runtime dependencies
2. **Local AI models** - Ollama runs on-device
3. **Minimal memory** - 50MB RAM footprint
4. **Offline capable** - Works without internet
5. **Fast deployment** - Copy binary, run
6. **Simple updates** - Replace binary

### Template Structure

```yaml
name: edge_health_monitor
description: Lightweight health monitoring for edge deployments
version: 1.0.0

config:
  defaults:
    provider: ollama  # Local model - no internet required
    model: qwen2.5:7b  # Smaller model for edge devices

steps:
  # Step 1: Collect local metrics
  - name: collect_metrics
    servers: [local-prometheus]  # Local metrics server
    prompt: |
      Fetch current system metrics:
      - CPU usage
      - Memory usage
      - Disk usage
      - Network throughput
      - Process health
    output: current_metrics

  # Step 2: Analyze metrics locally
  - name: analyze_health
    provider: ollama
    model: qwen2.5:7b
    prompt: |
      Analyze these system metrics:
      {{current_metrics}}
      
      Determine:
      - Overall health status (healthy/degraded/critical)
      - Any anomalies detected
      - Trends (improving/stable/degrading)
      - Resource pressure points
      
      Use these thresholds:
      - CPU > 80%: Warning
      - CPU > 95%: Critical
      - Memory > 85%: Warning
      - Memory > 95%: Critical
      - Disk > 90%: Critical
    output: health_analysis

  # Step 3: Detect anomalies
  - name: anomaly_detection
    provider: ollama
    model: qwen2.5:7b
    prompt: |
      Compare current metrics to baseline:
      
      Current: {{current_metrics}}
      Baseline (historical): {{baseline_metrics}}
      
      Detect anomalies:
      - Sudden spikes (>50% change)
      - Gradual degradation (trend analysis)
      - Pattern changes (normal load patterns)
      
      Return severity: none/low/medium/high/critical
    output: anomaly_report

  # Step 4: Local logging (always happens)
  - name: log_locally
    servers: [local-file]
    prompt: |
      Log to /var/log/edge-monitor.log:
      Timestamp: {{execution.timestamp}}
      Health: {{health_analysis.status}}
      Anomalies: {{anomaly_report.severity}}
      Details: {{health_analysis}}

  # Step 5: Escalate if critical (only if internet available and issue is critical)
  - name: escalate_if_critical
    condition: "{{health_analysis.status}} == 'critical' OR {{anomaly_report.severity}} == 'critical'"
    servers: [pagerduty]  # This step only runs if critical AND internet available
    prompt: |
      Send alert to PagerDuty:
      
      Severity: Critical
      Location: {{device_location}}
      Issue: {{health_analysis.summary}}
      Metrics: {{current_metrics}}
      Anomaly: {{anomaly_report}}
      
      Timestamp: {{execution.timestamp}}
```

---

## Usage Examples

### Example 1: Deploy to Edge Device

**Scenario:** Deploy monitoring to remote data center with intermittent connectivity

**Deployment:**

```bash
# 1. Copy 20MB binary to edge device (3-second transfer)
scp mcp-cli edge-device.example.com:/usr/local/bin/
# Transferred: 20.1 MB in 3.2 seconds

# 2. Copy template config
scp edge_health_monitor.yaml edge-device:/etc/mcp-cli/templates/
# Transferred: 2.4 KB in 0.1 seconds

# 3. Install Ollama and model on edge device
ssh edge-device.example.com
edge$ curl https://ollama.ai/install.sh | sh
edge$ ollama pull qwen2.5:7b  # 4.7GB model, one-time download

# 4. Run monitoring
edge$ mcp-cli --template edge_health_monitor --loop 60
# Runs every 60 seconds
```

**Total deployment:**
- Binary: 20MB
- Template: 2KB
- Model: 4.7GB (one-time, can work offline after)
- **No additional dependencies**

**vs. Python Framework:**
```bash
# Python framework deployment (for comparison)
scp -r python_framework/ edge-device:/opt/  # 500MB+
ssh edge-device
edge$ apt-get install python3.11  # Requires package manager
edge$ pip install -r requirements.txt  # 50+ packages, dependency resolution
edge$ # Debug dependency conflicts...
edge$ # Configure vector DB connection...
edge$ # Set up message queue...
edge$ # Total time: 30-60 minutes, many potential failures
```

---

### Example 2: Offline Operation

**Scenario:** Air-gapped secure environment with no internet

**Setup:**
```bash
# On internet-connected machine (one-time setup)
workstation$ ./mcp-cli --help
workstation$ ollama pull qwen2.5:7b

# Transfer to air-gapped device
workstation$ scp mcp-cli airgap-device:/usr/local/bin/
workstation$ scp -r ~/.ollama/models/qwen2.5 airgap-device:/models/

# On air-gapped device
airgap$ export OLLAMA_MODELS=/models
airgap$ mcp-cli --template edge_health_monitor
# Works completely offline
```

**What happens:**
```
[14:32:10] Starting edge_health_monitor
[14:32:10] Step: collect_metrics (local Prometheus)
[14:32:11] ✓ Metrics collected (1.2s)
[14:32:11] Step: analyze_health (local Ollama model)
[14:32:14] ✓ Analysis complete (3.1s) - Status: Healthy
[14:32:14] Step: anomaly_detection (local Ollama model)
[14:32:17] ✓ Anomaly check complete (2.8s) - Severity: None
[14:32:17] Step: log_locally
[14:32:17] ✓ Logged to /var/log/edge-monitor.log
[14:32:17] Step: escalate_if_critical (skipped - condition not met)
[14:32:17] ✓ Monitoring cycle complete (7 seconds total)

# NO INTERNET REQUIRED
# NO EXTERNAL DEPENDENCIES
# NO CLOUD API CALLS
```

**Resource usage:**
- CPU: 15% during analysis
- Memory: 52MB (mcp-cli) + 800MB (Ollama model loaded)
- Disk: 20MB (binary) + 4.7GB (model)
- Network: 0 bytes (completely offline)

---

### Example 3: Developer Laptop

**Scenario:** Full AI capabilities on laptop without complex installation

**Installation:**
```bash
# macOS
brew install mcp-cli
brew install ollama
ollama pull qwen2.5:7b

# Linux
curl -LO https://github.com/LaurieRhodes/mcp-cli/releases/latest/download/mcp-cli-linux-amd64
chmod +x mcp-cli-linux-amd64
sudo mv mcp-cli-linux-amd64 /usr/local/bin/mcp-cli
curl https://ollama.ai/install.sh | sh
ollama pull qwen2.5:7b

# Total time: 5 minutes (mostly model download)
# No Python environment
# No dependency management
# No virtual environments
```

**Usage:**
```bash
# Analyze logs
cat application.log | mcp-cli --template log_analysis

# Review code
cat mycode.py | mcp-cli --template code_review

# Generate tests
mcp-cli --template test_generator --input-data "$(cat mycode.py)"

# All running locally, no API costs, offline capable
```

---

## When to Use

### ✅ Appropriate Use Cases

**Edge Data Centers:**
- Limited bandwidth to cloud
- Need local intelligence
- Intermittent connectivity acceptable
- Resource monitoring and alerting

**IoT/Industrial:**
- Factory equipment monitoring
- Retail store systems
- Remote site operations
- Field device diagnostics

**Secure Environments:**
- Air-gapped networks
- Classified systems
- Financial trading floors
- Healthcare systems (HIPAA)

**Development:**
- Developer laptops/workstations
- No cloud costs during dev
- Offline coding (flights, remote work)
- Fast iteration without API latency

**Cost-Sensitive:**
- High-volume monitoring (100+ devices)
- Budget constraints
- Avoid per-query API costs
- Hardware is cheaper than cloud APIs

### ❌ Inappropriate Use Cases

**Need Latest Models:**
- Edge models lag behind cloud (GPT-4o, Claude Opus)
- Complex reasoning tasks better on cloud
- Specialized domains (medical, legal) need latest training

**Resource-Constrained:**
- IoT devices with <1GB RAM
- Embedded systems
- Very old hardware
- Can't fit model (4.7GB minimum)

**Highest Quality Required:**
- Critical decision-making
- Compliance where best model needed
- Customer-facing responses
- Brand reputation risk

---

## Trade-offs

### Advantages

**Deployment Simplicity:**
- **20MB binary** vs 500MB+ framework
- **No dependencies** vs 50+ pip packages
- **3-second transfer** vs 30-minute installation
- **Copy and run** vs complex setup

**Resource Efficiency:**

| Metric | mcp-cli | Python Framework |
|--------|---------|------------------|
| Binary | 20MB | 500MB+ |
| Runtime memory | 50MB | 512MB+ |
| Startup time | 100ms | 2-5s |
| Dependencies | 0 | 50+ packages |

**Offline Capability:**
- **100% offline** with local models
- **No internet required** for operation
- **No API costs** (free after hardware)
- **Low latency** (no network roundtrip)

**MIT License Benefits:**
- **On-premise deployment** allowed
- **Source modification** permitted
- **Commercial use** unrestricted
- **No vendor lock-in**

### Limitations

**Model Quality:**
- **Qwen 2.5 7B** good but not GPT-4o/Claude Opus
- **Simpler reasoning** than cloud models
- **Less specialized knowledge** in niche domains
- **Acceptable for:** monitoring, logs, patterns, metrics
- **Not ideal for:** complex analysis, novel domains

**Resource Requirements:**
```
Minimum for local models:
- RAM: 8GB (16GB recommended)
- Disk: 10GB free (for models)
- CPU: 4 cores (8+ better)
```

**Model Size:**
- 7B model: 4.7GB
- 13B model: 7.8GB
- 32B model: 19GB
- Must fit on device storage

**Initial Setup:**
- Model download: 4.7GB (one-time)
- Requires internet for initial pull
- Subsequent operation fully offline

---

## Resource Comparison

### Edge Device Example

**Specifications:**
- CPU: 4 cores @ 2.4 GHz
- RAM: 8GB
- Disk: 128GB SSD
- Network: 100 Mbps (intermittent)

**mcp-cli Deployment:**
```
Binary: 20MB
Model: 4.7GB (qwen2.5:7b)
Runtime memory: 850MB (model loaded)
CPU usage: 10-30% during analysis
Queries/hour: ~1,200 (3s per query)
Cost: $0 (hardware already there)
```

**Python Framework (hypothetical):**
```
Framework: 500MB
Dependencies: 400MB
Runtime memory: 1.2GB
CPU usage: 15-40%
Queries/hour: ~800 (4.5s per query)
Cost: $0 + complexity overhead
```

**Cloud API (hypothetical):**
```
Deployment: N/A (cloud)
Runtime memory: N/A (cloud)
Cost: $0.001/query × 1,200/hour × 24h × 30d = $864/month
Latency: 200ms network + 1s processing = 1.2s
Availability: Requires internet
```

**Decision matrix:**
- **Cost:** Edge wins (no recurring costs)
- **Latency:** Edge wins (no network)
- **Quality:** Cloud wins (better models)
- **Availability:** Edge wins (offline capable)
- **Simplicity:** mcp-cli wins (no framework complexity)

---

## Deployment Patterns

### Pattern 1: Kubernetes DaemonSet

**Deploy to every node in cluster:**

```yaml
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: edge-health-monitor
  namespace: monitoring
spec:
  selector:
    matchLabels:
      app: health-monitor
  template:
    metadata:
      labels:
        app: health-monitor
    spec:
      containers:
      - name: monitor
        image: alpine:latest
        command: ["/bin/sh", "-c"]
        args:
        - |
          while true; do
            /usr/local/bin/mcp-cli --template edge_health_monitor
            sleep 60
          done
        resources:
          limits:
            memory: "1Gi"   # Enough for binary + model
            cpu: "500m"
          requests:
            memory: "512Mi"
            cpu: "250m"
        volumeMounts:
        - name: mcp-cli
          mountPath: /usr/local/bin/mcp-cli
          subPath: mcp-cli
        - name: ollama-models
          mountPath: /root/.ollama/models
      volumes:
      - name: mcp-cli
        hostPath:
          path: /opt/mcp-cli/mcp-cli
          type: File
      - name: ollama-models
        hostPath:
          path: /opt/ollama/models
          type: Directory
```

**Result:** Monitoring runs on every Kubernetes node with minimal overhead.

---

### Pattern 2: Systemd Service

**Run as background service:**

```ini
# /etc/systemd/system/edge-monitor.service
[Unit]
Description=Edge Health Monitor
After=network.target ollama.service

[Service]
Type=simple
User=monitor
ExecStart=/usr/local/bin/mcp-cli --template edge_health_monitor --loop 60
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
```

```bash
# Enable and start
sudo systemctl enable edge-monitor
sudo systemctl start edge-monitor

# Check status
sudo systemctl status edge-monitor
# ● edge-monitor.service - Edge Health Monitor
#    Loaded: loaded (/etc/systemd/system/edge-monitor.service; enabled)
#    Active: active (running) since...
```

---

### Pattern 3: Docker Container (Minimal Image)

```dockerfile
# Minimal Docker image (25MB total)
FROM alpine:latest

# Copy binary (20MB)
COPY mcp-cli /usr/local/bin/mcp-cli

# Copy templates
COPY templates/ /etc/mcp-cli/templates/

# Copy Ollama models (mounted as volume in production)
# VOLUME /root/.ollama/models

# Run monitor
CMD ["mcp-cli", "--template", "edge_health_monitor", "--loop", "60"]
```

```bash
# Build
docker build -t edge-monitor:latest .

# Run
docker run -d \
  -v /var/run/ollama:/var/run/ollama \
  -v ollama-models:/root/.ollama/models \
  --name edge-monitor \
  edge-monitor:latest

# Image size: 25MB (Alpine 5MB + mcp-cli 20MB)
# vs Python framework: 1GB+ image
```

---

## Best Practices

### Edge Deployment

**✅ Do:**
- Use smaller models for resource-constrained devices (7B vs 32B)
- Test offline operation before deployment
- Monitor resource usage (memory, CPU)
- Set up local logging (survive restarts)
- Plan for model updates

**❌ Don't:**
- Require internet for core functionality
- Use models too large for device RAM
- Skip testing in offline mode
- Forget about log rotation
- Deploy without rollback plan

### Model Selection

**7B models** (qwen2.5:7b, llama3.2:7b):
- RAM: 8GB minimum
- Good for: monitoring, logs, patterns, metrics
- Speed: 3-5 tokens/second
- Quality: Acceptable for operational tasks

**13B models** (qwen2.5:13b):
- RAM: 16GB recommended
- Good for: more complex analysis, better reasoning
- Speed: 2-3 tokens/second
- Quality: Better, still not cloud-level

**32B+ models** (qwen2.5:32b):
- RAM: 32GB required
- Good for: desktop workstations, high-end edge
- Speed: 1-2 tokens/second
- Quality: Approaches cloud models

---

## Troubleshooting

### Issue: Model Too Large for Device

**Symptoms:**
```
Error: Failed to load model - insufficient memory
```

**Solution:**
```bash
# Use smaller model
ollama pull qwen2.5:7b  # Instead of 32b

# Or use quantized model (lower quality, smaller size)
ollama pull qwen2.5:7b-q4_0  # 4-bit quantization
```

---

### Issue: Slow Performance

**Symptoms:**
```
Analysis taking 30+ seconds per query
```

**Solution:**
```bash
# Check CPU/RAM usage
htop

# If CPU saturated: Use smaller model
# If RAM swapping: Use smaller model or add RAM
# If both okay: Model might be too complex for hardware

# Optimize:
ollama pull qwen2.5:7b  # Smaller model
# Or use GPU if available
```

---

## Related Resources

- **[Template File](../templates/edge_health_monitor.yaml)** - Download complete template
- **[Resilient Incident Analysis](resilient-incident-analysis.md)** - Failover for reliability
- **[Why Templates Matter](../../../WHY_TEMPLATES_MATTER.md)** - Lightweight deployment explained
- **[Ollama Documentation](https://ollama.ai/docs)** - Local model setup

---

**Edge deployment: Full AI capabilities where you need them, without cloud dependency.**

Remember: Edge models trade some quality for independence, offline capability, and zero recurring costs.
