# Installation Guide

This guide will get MCP-CLI-Go running on your system in 5 minutes.

**What you're installing:** A single, self-contained binary (program) with no dependencies.

**Why it's easy:**

- Just one file to download
- No Python, Node.js, or other runtimes needed
- No package managers required
- Works on any Linux/macOS/Windows system

**What you'll get:**

- `mcp-cli` command available system-wide
- Ready to use immediately
- ~20MB disk space used

---

## Quick Install (Copy-Paste)

Choose your operating system and copy-paste the commands:

### Linux

```bash
# Download latest release
wget https://github.com/LaurieRhodes/mcp-cli-go/releases/latest/download/mcp-cli-linux-amd64

# Make executable (tells Linux this is a program to run)
chmod +x mcp-cli-linux-amd64

# Move to system path (makes 'mcp-cli' command available everywhere)
sudo mv mcp-cli-linux-amd64 /usr/local/bin/mcp-cli

# Verify installation (should show version number)
mcp-cli --version
```

**What happens:**

1. `wget` downloads the file (~20MB)
2. `chmod +x` marks it as executable
3. `sudo mv` moves it to `/usr/local/bin` (where system-wide programs live)
4. Now you can type `mcp-cli` from any directory

**Alternative (no root access):**

```bash
# Install to your home directory
mkdir -p ~/.local/bin
mv mcp-cli-linux-amd64 ~/.local/bin/mcp-cli
chmod +x ~/.local/bin/mcp-cli

# Add to PATH (makes command available)
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc

# Verify
mcp-cli --version
```

### macOS (Intel)

```bash
# Download
wget https://github.com/LaurieRhodes/mcp-cli-go/releases/latest/download/mcp-cli-darwin-amd64

# Make executable
chmod +x mcp-cli-darwin-amd64

# Remove quarantine (macOS Gatekeeper blocks unsigned downloads)
xattr -d com.apple.quarantine mcp-cli-darwin-amd64

# Move to system path
sudo mv mcp-cli-darwin-amd64 /usr/local/bin/mcp-cli

# Verify
mcp-cli --version
```

**What happens:**

1. Downloads the Intel Mac binary (~20MB)
2. `chmod +x` makes it executable
3. `xattr -d` tells macOS "this download is safe" (bypasses Gatekeeper)
4. Moves to `/usr/local/bin` for system-wide access
5. Verification shows it's working

**If you see "command not found":** Make sure `/usr/local/bin` is in your PATH

### macOS (Apple Silicon)

```bash
# Download (for M1/M2/M3/M4 Macs)
wget https://github.com/LaurieRhodes/mcp-cli-go/releases/latest/download/mcp-cli-darwin-arm64

# Make executable
chmod +x mcp-cli-darwin-arm64

# Remove quarantine
xattr -d com.apple.quarantine mcp-cli-darwin-arm64

# Move to system path
sudo mv mcp-cli-darwin-arm64 /usr/local/bin/mcp-cli

# Verify
mcp-cli --version
```

**How to know which Mac you have:**

- Apple Silicon (M1/M2/M3): Use `darwin-arm64`
- Intel Mac: Use `darwin-amd64`
- Check: Apple menu → About This Mac → look for "Chip"

### Windows

**PowerShell (recommended):**

```powershell
# Download
Invoke-WebRequest -Uri "https://github.com/LaurieRhodes/mcp-cli-go/releases/latest/download/mcp-cli-windows-amd64.exe" -OutFile "mcp-cli.exe"

# Create bin directory if it doesn't exist
New-Item -ItemType Directory -Force -Path $env:USERPROFILE\bin

# Move to bin directory
Move-Item mcp-cli.exe $env:USERPROFILE\bin\

# Add to PATH (makes command available everywhere)
$env:Path += ";$env:USERPROFILE\bin"
[Environment]::SetEnvironmentVariable("Path", $env:Path, "User")

# Verify (open NEW PowerShell window first!)
mcp-cli --version
```

**What happens:**

1. Downloads Windows executable (~20MB)
2. Creates `C:\Users\YourName\bin` directory
3. Moves `mcp-cli.exe` there
4. Adds that directory to your PATH
5. After opening new terminal, `mcp-cli` command works

**Important:** You must open a NEW PowerShell window after adding to PATH!

**Manual alternative (if PowerShell fails):**

1. Download from [Releases Page](https://github.com/LaurieRhodes/mcp-cli-go/releases/latest)
2. Click `mcp-cli-windows-amd64.exe`
3. Rename to `mcp-cli.exe`
4. Move to `C:\Users\YourName\bin\`
5. Add `C:\Users\YourName\bin` to PATH:
   - Right-click "This PC" → Properties
   - Advanced system settings → Environment Variables
   - Under "User variables", edit "Path"
   - Click "New", add: `C:\Users\YourName\bin`
   - Click OK on all dialogs
6. Open new PowerShell, run: `mcp-cli --version`

---

## Platform-Specific Installation

### Linux

#### Debian/Ubuntu (APT)

Currently no official APT repository. Use binary installation or build from source.

#### RHEL/CentOS/Fedora (DNF/YUM)

Currently no official RPM repository. Use binary installation or build from source.

#### Arch Linux (AUR)

Currently no AUR package. Use binary installation or build from source.

#### Manual Installation (All Distros)

```bash
# 1. Download binary
cd /tmp
wget https://github.com/LaurieRhodes/mcp-cli-go/releases/latest/download/mcp-cli-linux-amd64

# 2. Verify checksum (optional but recommended)
wget https://github.com/LaurieRhodes/mcp-cli-go/releases/latest/download/mcp-cli-linux-amd64.sha256
sha256sum -c mcp-cli-linux-amd64.sha256

# 3. Install
chmod +x mcp-cli-linux-amd64
sudo mv mcp-cli-linux-amd64 /usr/local/bin/mcp-cli

# 4. Verify
mcp-cli --version
```

#### User Installation (No Root)

```bash
# Install to user directory
mkdir -p ~/.local/bin
mv mcp-cli-linux-amd64 ~/.local/bin/mcp-cli
chmod +x ~/.local/bin/mcp-cli

# Add to PATH (add to ~/.bashrc or ~/.zshrc)
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc

# Verify
mcp-cli --version
```

---

### macOS

#### Homebrew

Currently no Homebrew tap. Use binary installation or build from source.

Future:

```bash
# Coming soon
brew tap LaurieRhodes/mcp-cli
brew install mcp-cli
```

#### Manual Installation

```bash
# Determine your architecture
if [[ $(uname -m) == "arm64" ]]; then
    ARCH="arm64"
else
    ARCH="amd64"
fi

# Download
cd /tmp
wget https://github.com/LaurieRhodes/mcp-cli-go/releases/latest/download/mcp-cli-darwin-${ARCH}

# Verify checksum
wget https://github.com/LaurieRhodes/mcp-cli-go/releases/latest/download/mcp-cli-darwin-${ARCH}.sha256
shasum -a 256 -c mcp-cli-darwin-${ARCH}.sha256

# Install
chmod +x mcp-cli-darwin-${ARCH}
xattr -d com.apple.quarantine mcp-cli-darwin-${ARCH}
sudo mv mcp-cli-darwin-${ARCH} /usr/local/bin/mcp-cli

# Verify
mcp-cli --version
```

#### Troubleshooting macOS

**"mcp-cli cannot be opened because the developer cannot be verified"**

Solution:

```bash
xattr -d com.apple.quarantine /usr/local/bin/mcp-cli
```

Or: System Preferences → Security & Privacy → Allow

**"Permission denied"**

Solution:

```bash
chmod +x /usr/local/bin/mcp-cli
```

---

### Windows

#### Manual Installation

1. **Download Binary:**
   
   - Visit [Releases Page](https://github.com/LaurieRhodes/mcp-cli-go/releases/latest)
   - Download `mcp-cli-windows-amd64.exe`

2. **Create Installation Directory:**
   
   ```powershell
   New-Item -ItemType Directory -Force -Path $env:USERPROFILE\bin
   ```

3. **Move Binary:**
   
   ```powershell
   Move-Item .\mcp-cli-windows-amd64.exe $env:USERPROFILE\bin\mcp-cli.exe
   ```

4. **Add to PATH:**
   
   **PowerShell (User):**
   
   ```powershell
   $env:Path += ";$env:USERPROFILE\bin"
   [Environment]::SetEnvironmentVariable("Path", $env:Path, "User")
   ```
   
   **Or manually:**
   
   - Right-click "This PC" → Properties
   - Advanced system settings
   - Environment Variables
   - Edit "Path" under User variables
   - Add: `C:\Users\YourUsername\bin`

5. **Verify:**
   
   ```powershell
   # Open new PowerShell window
   mcp-cli --version
   ```

#### Windows Subsystem for Linux (WSL)

Follow Linux installation instructions inside WSL.



---

## Build from Source

### Prerequisites

- Go 1.21 or higher
- Git

### Build Steps

```bash
# 1. Clone repository
git clone https://github.com/LaurieRhodes/mcp-cli-go.git
cd mcp-cli-go

# 2. Build
go build -o mcp-cli .

# 3. Install (optional)
sudo mv mcp-cli /usr/local/bin/

# 4. Verify
mcp-cli --version
```

### Build with Version Info

```bash
# Build with embedded version information
VERSION="v1.0.0"
BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT=$(git rev-parse --short HEAD)

go build \
  -ldflags="-s -w \
    -X 'main.Version=${VERSION}' \
    -X 'main.BuildTime=${BUILD_TIME}' \
    -X 'main.GitCommit=${GIT_COMMIT}'" \
  -o mcp-cli .
```

### Build for Multiple Platforms

```bash
# Use the build script
./scripts/build-all.sh

# Or manually
GOOS=linux GOARCH=amd64 go build -o mcp-cli-linux-amd64 .
GOOS=darwin GOARCH=amd64 go build -o mcp-cli-darwin-amd64 .
GOOS=darwin GOARCH=arm64 go build -o mcp-cli-darwin-arm64 .
GOOS=windows GOARCH=amd64 go build -o mcp-cli-windows-amd64.exe .
```

---

## Post-Installation Setup

After installation, you need to configure MCP-CLI. This takes 2-5 minutes.

### 1. Verify Installation Works

```bash
# Check version (should show vX.Y.Z)
mcp-cli --version

# Check help (should show command options)
mcp-cli --help
```

**If these work, installation succeeded!**

**If you get "command not found":**

- Linux/Mac: Binary not in PATH. Re-run installation, ensure `/usr/local/bin` is in PATH
- Windows: Must open NEW terminal after PATH changes
- All platforms: Try `./mcp-cli --version` in download directory

---

### 2. Initialize Configuration

MCP-CLI needs a configuration directory. Choose one option:

**Option A: Quick Setup (No API Keys - Uses Free Local AI)**

```bash
mcp-cli init --quick
```

**What this does:**

- Creates `config/` directory
- Sets up Ollama provider (free local AI)
- Creates `.env` file (empty, no API keys needed)
- Creates `config.yaml`
- **No cloud AI providers configured**

**Use when:** Want to try MCP-CLI immediately without API keys

**Next step:** Install Ollama from https://ollama.com to run local AI

---

**Option B: Full Setup (Interactive - Multiple Providers)**

```bash
mcp-cli init
```

**What this does:**

- Interactive prompts guide you through setup
- Configure multiple AI providers (OpenAI, Anthropic, etc.)
- Creates full directory structure
- Asks for API keys

**Use when:** Want to use cloud AI providers (Claude, GPT-4, etc.)

**You'll be asked:**

- Which providers to configure (OpenAI, Anthropic, Ollama, etc.)
- API keys for each provider
- Default settings

---

### Directory Structure Created

After `mcp-cli init`, you'll have:

```
your-directory/
├── mcp-cli                  # The program you installed
├── .env                     # API keys (keep this private!)
├── config.yaml              # Main config (points to other configs)
└── config/
    ├── settings.yaml        # Global settings (default AI, logging)
    ├── providers/           # AI provider configs (OpenAI, Claude, etc.)
    │   ├── openai.yaml
    │   ├── anthropic.yaml
    │   └── ollama.yaml
    ├── embeddings/          # Embedding configs (optional, for RAG)
    ├── servers/             # MCP server configs (tools for AI)
    ├── templates/           # Your workflow templates
    │   └── examples/        # Example templates
    └── runas/               # Server mode configs (advanced)
```

**What each part does:**

- **`.env`**: Stores API keys securely (never commit to git!)
- **`config.yaml`**: Index file pointing to all other configs
- **`settings.yaml`**: Your preferences (which AI to use by default)
- **`providers/`**: Each AI provider's configuration
- **`templates/`**: Your multi-step AI workflows

---

### 3. Add API Keys (If Using Cloud Providers)

Edit the `.env` file:

```bash
# Open in text editor
nano .env

# Or
code .env
```

Add your API keys:

```bash
# OpenAI (for GPT-4, GPT-4o, etc.)
OPENAI_API_KEY=sk-proj-...

# Anthropic (for Claude)
ANTHROPIC_API_KEY=sk-ant-...

# DeepSeek (cost-effective alternative)
DEEPSEEK_API_KEY=sk-...

# Google Gemini
GEMINI_API_KEY=...

# OpenRouter (access to many models)
OPENROUTER_API_KEY=sk-or-...
```

**Where to get API keys:**

- OpenAI: https://platform.openai.com/api-keys
- Anthropic: https://console.anthropic.com/
- DeepSeek: https://platform.deepseek.com/
- Google: https://makersuite.google.com/app/apikey
- OpenRouter: https://openrouter.ai/keys

**Important security notes:**

- Never commit `.env` to git (add to `.gitignore`)
- Keep API keys private (they charge your account)
- Use environment variables in production

**Don't want cloud APIs?** Use Ollama (local, free):

1. Install Ollama: https://ollama.com
2. Pull a model: `ollama pull llama3.2`
3. No API key needed!

---

### 4. Test Your Setup

**Test with Ollama (if installed):**

```bash
mcp-cli query --provider ollama "What is 2+2?"
```

**Expected output:**

```
2 + 2 equals 4.
```

**Test with cloud provider:**

```bash
mcp-cli query --provider openai "Hello, world!"
```

**Expected output:**

```
Hello! How can I help you today?
```

**List available templates:**

```bash
mcp-cli --list-templates
```

**Expected output:**

```
Available templates:
- analyze
- summarize
- extract
...
```

**If everything works, you're ready to go!**

---

## What's Next?

Now that MCP-CLI is installed:

1. **[Quick Start Guide](../quick_start.md)** - Run your first queries and templates
2. **[Core Concepts](concepts.md)** - Understand how MCP-CLI works
3. **[Template Examples](../templates/examples/)** - See what you can build

**Try this right now:**

```bash
# Ask a question
mcp-cli query "What are the top 3 programming languages?"

# Try a template
echo "Hello world" | mcp-cli --template analyze
```

---

## Verification

### Verify Static Linking (Linux)

```bash
ldd /usr/local/bin/mcp-cli

# Should output:
# not a dynamic executable
```

This means the binary has **no dependencies** and will run on any Linux system.

### Verify Binary Integrity

```bash
# Download checksum file
wget https://github.com/LaurieRhodes/mcp-cli-go/releases/latest/download/mcp-cli-linux-amd64.sha256

# Verify
sha256sum -c mcp-cli-linux-amd64.sha256

# Should output:
# mcp-cli-linux-amd64: OK
```

---

## Upgrading

### Binary Installation

```bash
# 1. Download new version
wget https://github.com/LaurieRhodes/mcp-cli-go/releases/latest/download/mcp-cli-linux-amd64

# 2. Replace existing binary
sudo mv mcp-cli-linux-amd64 /usr/local/bin/mcp-cli
chmod +x /usr/local/bin/mcp-cli

# 3. Verify new version
mcp-cli --version
```

# 

## Uninstallation

### Remove Binary

```bash
# Linux/macOS
sudo rm /usr/local/bin/mcp-cli

# Windows
Remove-Item $env:USERPROFILE\bin\mcp-cli.exe
```

### Remove Configuration

```bash
# Remove config directory
rm -rf config/

# Remove main config
rm config.yaml

# Remove environment file
rm .env
```

---

## Troubleshooting Common Issues

### "Command not found" after installation

**What this means:** Your terminal can't find the `mcp-cli` program.

**Cause:** Binary not in your system's PATH.

**Solution:**

**Linux/macOS:**

```bash
# Check where binary is
which mcp-cli
ls -la /usr/local/bin/mcp-cli
ls -la ~/.local/bin/mcp-cli

# If it's there but still doesn't work, add to PATH
echo 'export PATH="/usr/local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc

# Try again
mcp-cli --version
```

**Windows:**

- Must open NEW PowerShell window after installation
- PATH changes only apply to new terminals
- Close all PowerShell windows, open fresh one

**Still not working?**
Run from the directory where you downloaded it:

```bash
./mcp-cli --version
```

---

### "Permission denied" when running mcp-cli

**What this means:** File isn't marked as executable.

**Solution:**

**Linux/macOS:**

```bash
chmod +x /usr/local/bin/mcp-cli
```

**Windows:**
Right-click `mcp-cli.exe` → Properties → Unblock → OK

---

### macOS: "cannot be opened because developer cannot be verified"

**What this means:** macOS Gatekeeper is blocking unsigned app.

**Solution:**

```bash
# Remove quarantine flag
xattr -d com.apple.quarantine /usr/local/bin/mcp-cli

# Or do it manually:
# System Settings → Privacy & Security → scroll down → click "Open Anyway"
```

**Why this happens:** The binary isn't signed with Apple Developer certificate (costs $99/year).

**Is it safe?** Yes - you can verify the checksum from GitHub releases.

---

### Windows SmartScreen blocks exe

**What this means:** Windows Defender SmartScreen blocking unsigned executable.

**Solution:**

1. Click "More info"
2. Click "Run anyway"

**Is it safe?** Yes - verify checksum if concerned:

```powershell
# Download checksum file
Invoke-WebRequest -Uri "https://github.com/LaurieRhodes/mcp-cli-go/releases/latest/download/mcp-cli-windows-amd64.exe.sha256" -OutFile "checksum.txt"

# Verify
Get-FileHash mcp-cli.exe -Algorithm SHA256 | Select-Object Hash
# Compare with checksum.txt
```

---

### mcp-cli works, but "config.yaml not found"

**What this means:** You haven't run initialization yet.

**Solution:**

```bash
# Run init in the directory where mcp-cli lives
mcp-cli init --quick
```

**Or specify config location:**

```bash
mcp-cli --config /path/to/config.yaml query "test"
```

---

### "API key not found" or "invalid API key"

**What this means:** MCP-CLI can't find or use your API key.

**Checklist:**

1. **Does `.env` file exist?**
   
   ```bash
   cat .env
   ```

2. **Is the key name correct?**
   
   ```bash
   # Should be exactly these names:
   OPENAI_API_KEY=sk-...
   ANTHROPIC_API_KEY=sk-ant-...
   # NOT: "OPENAI_KEY" or "OPENAI_API_KEY=sk-..."
   ```

3. **Are there quotes around the key?**
   
   ```bash
   # Wrong:
   OPENAI_API_KEY="sk-..."
   
   # Right:
   OPENAI_API_KEY=sk-...
   ```

4. **Is the key actually valid?**
   
   - Test at https://platform.openai.com/playground
   - Keys can expire or be revoked

---

### "Cannot execute binary file" (Linux)

**What this means:** Downloaded wrong architecture or file corrupted.

**Solution:**

1. **Check your architecture:**
   
   ```bash
   uname -m
   # x86_64 → use mcp-cli-linux-amd64
   # aarch64 → use mcp-cli-linux-arm64
   ```

2. **Re-download:**
   
   ```bash
   rm mcp-cli-linux-amd64
   wget https://github.com/LaurieRhodes/mcp-cli-go/releases/latest/download/mcp-cli-linux-amd64
   ```

3. **Verify checksum:**
   
   ```bash
   wget https://github.com/LaurieRhodes/mcp-cli-go/releases/latest/download/mcp-cli-linux-amd64.sha256
   sha256sum -c mcp-cli-linux-amd64.sha256
   # Should say: OK
   ```

---

### Installation worked, but queries timeout

**What this means:** Network issues or provider problem.

**Debugging steps:**

1. **Test network:**
   
   ```bash
   # Can you reach the AI provider?
   ping api.openai.com
   curl https://api.openai.com/v1/models -H "Authorization: Bearer $OPENAI_API_KEY"
   ```

2. **Try different provider:**
   
   ```bash
   # If OpenAI times out, try Anthropic
   mcp-cli query --provider anthropic "test"
   ```

3. **Check verbose output:**
   
   ```bash
   mcp-cli --verbose query "test"
   # Shows exactly where it's hanging
   ```

4. **Firewall/proxy:**
   
   - Corporate firewall might block AI APIs
   - Try from different network
   - Check if proxy configuration needed

---

## Still Having Issues?

**Get help:**

1. **Check logs:** Run with `--verbose --verbose` to see detailed output
2. **Search Issues:** [GitHub Issues](https://github.com/LaurieRhodes/mcp-cli-go/issues)
3. **Ask Community:** [GitHub Discussions](https://github.com/LaurieRhodes/mcp-cli-go/discussions)
4. **Report Bug:** [New Issue](https://github.com/LaurieRhodes/mcp-cli-go/issues/new)

**Include in bug report:**

- Operating system and architecture
- mcp-cli version (`mcp-cli --version`)
- Full error message
- Output of `mcp-cli --verbose --verbose query "test"`

---

## Platform-Specific Notes

### Linux

- **Static binary**: No dependencies, runs on any Linux distro
- **No root required**: Can install to `~/.local/bin`
- **Works in Docker**: Use `FROM scratch` base image

### macOS

- **Gatekeeper**: Requires removing quarantine attribute
- **Architecture**: Separate binaries for Intel and Apple Silicon
- **Rosetta**: Intel binary works on Apple Silicon (slower)

### Windows

- **PATH**: Must add to PATH manually
- **Antivirus**: May flag unsigned binary (false positive)
- **WSL**: Use Linux version inside WSL for better compatibility

---

## Next Steps

After installation:

1. **[Quick Start Guide](quick-start.md)** - Get running in 5 minutes
2. **[Core Concepts](concepts.md)** - Understand how MCP-CLI works
3. **[Configuration Guide](../../CLI-REFERENCE.md)** - Configure providers and servers
4. **[Template Guide](../templates/authoring-guide.md)** - Create your first template

---

## Getting Help

- **Documentation:** [docs/](../)
- **Issues:** [GitHub Issues](https://github.com/LaurieRhodes/mcp-cli-go/issues)
- **Discussions:** [GitHub Discussions](https://github.com/LaurieRhodes/mcp-cli-go/discussions)

---

## Alternative Installation Methods

### Go Install

```bash
# Install latest from source
go install github.com/LaurieRhodes/mcp-cli-go@latest

# This installs to $GOPATH/bin (usually ~/go/bin)
```

### Nix

Currently no Nix package. Coming soon.

### Snap

Currently no Snap package. Coming soon.

---

**Installation complete!** 🎉

Continue to [Quick Start Guide](quick-start.md) to run your first query.
