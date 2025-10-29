# lrok - Expose Local Servers to the Internet

**lrok** (short for *lum-rok*) is a fast, secure tunnel service that exposes your localhost to the internet with HTTPS and human-readable URLs — like ngrok, but built on [platform.lum.tools](https://platform.lum.tools) infrastructure.

```bash
lrok 8000
# → https://happy-dolphin.t.lum.tools (public URL)
# → http://localhost:4242 (local dashboard)
```

## TL;DR - Quick Start

```bash
# Install
curl -fsSL https://platform.lum.tools/install.sh | bash

# Get API key (free, no credit card)
# Visit: https://platform.lum.tools/keys

# Set key
export LUM_API_KEY='lum_your_key'

# Expose port 8000
lrok 8000
```

Done! Your local server is now accessible at a public HTTPS URL. 🎉

---

**Perfect for:**
- Testing webhooks locally (Stripe, GitHub, etc.)
- Sharing dev environments with teammates
- Demoing work-in-progress to clients
- Remote access to local services
- Mobile app testing

**100% Free**: No credit card required. No paid tiers. No usage limits (for now).

## Features

- **🎯 Readable URLs**: Get memorable names like `happy-dolphin.t.lum.tools` instead of random hashes
- **📊 Built-in Request Inspector**: Beautiful web UI at `http://localhost:4242` with:
  - Real-time request/response viewer
  - HTTP headers inspection
  - Request/response body viewer
  - Status code tracking
  - Request timing metrics
- **🔒 HTTPS by Default**: All tunnels automatically secured with valid SSL certificates
- **📈 Traffic Tracking**: Monitor your tunnel usage at [platform.lum.tools/tunnels](https://platform.lum.tools/tunnels)
- **⚡ Zero Configuration**: Works out of the box with lum.tools infrastructure
- **🌍 Cross-Platform**: Single binary for macOS, Linux, and Windows
- **📦 Self-Contained**: No additional dependencies to install
- **🔌 Multiple Protocols**: Support for HTTP, TCP, STCP, and XTCP tunnels
- **🔐 Secure Tunnels**: STCP and XTCP for encrypted, private connections
- **⚡ P2P Mode**: XTCP for direct client-to-client connections
- **📊 Advanced Monitoring**: Protocol-specific statistics and Prometheus metrics

## Installation

### Quick Install (macOS/Linux) - Recommended

```bash
curl -fsSL https://platform.lum.tools/install.sh | bash
```

This automatically detects your platform and installs lrok to `~/.local/bin`.

### npm (Cross-platform)

```bash
npm install -g lrok
```

### PyPI (Cross-platform)

```bash
pip install lrok
```

### Direct Download

Download the latest binary from [GitHub Releases](https://github.com/lum-tools/lrok/releases).

## Quick Start

### 1. Get Your Free API Key

1. Visit [platform.lum.tools](https://platform.lum.tools)
2. Sign in with Google, GitHub, or email (no credit card required)
3. Navigate to [API Keys](https://platform.lum.tools/keys)
4. Click "Create New Key" and give it a name
5. Copy your key (starts with `lum_`)

**Note**: lum.tools is 100% free with no usage limits. We don't ask for credit cards.

### 2. Login with Your API Key

**Recommended:** Save your API key to config file (like `docker login`, `gh auth login`):

```bash
lrok login lum_your_api_key_here
```

This saves your key to `~/.lrok/config.toml` so you never have to set it again!

**Alternative:** Use environment variable (still supported):

```bash
export LUM_API_KEY='lum_your_api_key_here'

# Make it permanent (optional)
echo 'export LUM_API_KEY="lum_your_key"' >> ~/.bashrc  # or ~/.zshrc
```

**Check your auth status:**

```bash
lrok whoami
# → ✅ Logged in
#    API Key: lum_abc123...xyz
#    Source: config file (~/.lrok/config.toml)
```

### 3. Start Tunneling

```bash
# Expose port 8000 with a random name
lrok 8000
# → https://happy-dolphin.t.lum.tools

# Use a custom name
lrok http 3000 --name my-app
# → https://my-app.t.lum.tools

# Shorthand is also supported
lrok 3000 -n my-app
```

Your terminal will show:
- 📍 Local address
- 🌐 Public URL
- 🏷️ Tunnel name
- 📊 Dashboard URL (http://localhost:4242)

## Usage

### Basic Syntax

```bash
# Shorthand (recommended)
lrok <port>

# Explicit protocol
lrok http <port>

# With options
lrok <port> [flags]
```

### Available Commands

```
lrok - Expose local servers to the internet

Usage:
  lrok [port]                 Quick HTTP tunnel with random name
  lrok http [port] [flags]    HTTP tunnel with options
  lrok tcp <port> [flags]     TCP tunnel for direct port forwarding
  lrok stcp <port> [flags]    Secret TCP tunnel (requires visitor)
  lrok xtcp <port> [flags]    P2P tunnel for direct client connections
  lrok visitor <name> [flags] Connect to STCP/XTCP tunnel as visitor
  lrok version                Show version information
  lrok help                   Show help

Examples:
  lrok 8000                   Expose port 8000 with random name
  lrok 3000 -n my-app         Expose port 3000 as my-app.t.lum.tools
  lrok tcp 5432 --remote-port 10001    Expose PostgreSQL on port 10001
  lrok stcp 22 --secret-key my-secret  Secure SSH tunnel
  lrok xtcp 8080 --secret-key p2p-key  P2P web server tunnel

Flags:
  -n, --name string        Custom tunnel name (generates random if not provided)
      --subdomain string   Alias for --name
  -k, --api-key string     API key (or set LUM_API_KEY env var)
      --ip string          Local IP address (default: 127.0.0.1)
      --remote-port int     Remote port on server (TCP only)
      --secret-key string   Pre-shared secret key (STCP/XTCP only)
      --encrypt            Enable encryption (TCP/STCP only)
      --compress           Enable compression (TCP/STCP only)
      --bandwidth string    Bandwidth limit (e.g., 1MB, 500KB)
      --health-check        Enable health checks (TCP only)
  -h, --help               Show help
```

## Examples

### HTTP Tunnels (Web Services)

#### Expose a Development Server
```bash
# Start your dev server
npm run dev  # Running on port 3000

# Create tunnel with custom name
lrok 3000 -n my-project

# Share: https://my-project.t.lum.tools
# Dashboard: http://localhost:4242
```

#### Webhook Testing
```bash
# Start local webhook server
python -m http.server 8000

# Create tunnel (random name)
lrok 8000

# Output shows your public URL:
# 🌐 https://clever-fox.t.lum.tools
# 📊 http://localhost:4242

# Use the public URL in your webhook provider
# Watch requests live in the dashboard!
```

### TCP Tunnels (Direct Port Forwarding)

#### Expose PostgreSQL Database
```bash
# Start PostgreSQL (running on port 5432)
sudo systemctl start postgresql

# Create TCP tunnel
lrok tcp 5432 --remote-port 10001

# Connect from anywhere:
# psql -h frp.lum.tools -p 10001 -U myuser mydb
```

#### Expose SSH Server
```bash
# Start SSH server (running on port 22)
sudo systemctl start ssh

# Create TCP tunnel with encryption
lrok tcp 22 --remote-port 10002 --encrypt --compress

# SSH from anywhere:
# ssh -p 10002 user@frp.lum.tools
```

#### Expose Redis Server
```bash
# Start Redis (running on port 6379)
redis-server

# Create TCP tunnel
lrok tcp 6379 --remote-port 10003

# Connect from anywhere:
# redis-cli -h frp.lum.tools -p 10003
```

### STCP Tunnels (Secret TCP - Secure Access)

#### Secure Database Access
```bash
# Server side: Expose PostgreSQL securely
lrok stcp 5432 --secret-key my-secret-key --encrypt --compress

# Client side: Connect as visitor
lrok visitor tunnel-name --type stcp --secret-key my-secret-key --bind-port 5432

# Now connect locally:
# psql -h 127.0.0.1 -p 5432 -U myuser mydb
```

#### Secure SSH Access
```bash
# Server side: Expose SSH securely
lrok stcp 22 --secret-key ssh-secret-123

# Client side: Connect as visitor
lrok visitor tunnel-name --type stcp --secret-key ssh-secret-123 --bind-port 2222

# SSH locally:
# ssh -p 2222 user@127.0.0.1
```

### XTCP Tunnels (P2P - Direct Connection)

#### P2P File Transfer
```bash
# Server side: Expose file server via P2P
lrok xtcp 8080 --secret-key p2p-file-transfer

# Client side: Connect as visitor
lrok visitor tunnel-name --type xtcp --secret-key p2p-file-transfer --bind-port 8080

# Access locally: http://127.0.0.1:8080
# If P2P fails, falls back to server relay automatically
```

#### P2P Web Development
```bash
# Developer side: Expose dev server via P2P
lrok xtcp 3000 --secret-key dev-p2p-key

# Client side: Connect as visitor
lrok visitor tunnel-name --type xtcp --secret-key dev-p2p-key --bind-port 3000

# Access locally: http://127.0.0.1:3000
# Direct connection for better performance
```

### Multiple Tunnels (Different Terminals)

```bash
# Terminal 1: HTTP Frontend
lrok 3000 -n frontend
# → https://frontend.t.lum.tools

# Terminal 2: TCP Backend API
lrok tcp 8000 --remote-port 10001
# → frp.lum.tools:10001

# Terminal 3: Secure Database
lrok stcp 5432 --secret-key db-secret
# → Requires visitor connection
```

### Inspect HTTP Traffic

Every tunnel includes a local dashboard at `http://localhost:4242`:

- **Real-time request list**: See each request as it happens
- **Request inspector**: Click any request to view:
  - Full headers (in/out)
  - Request/response bodies
  - Status codes & timing
  - Copy as cURL command

Perfect for debugging webhooks, API integrations, or understanding what your app is doing!

## Platform Dashboard

Track all your tunnel activity at [platform.lum.tools/tunnels](https://platform.lum.tools/tunnels):

- **Active tunnels**: See what's currently running
- **Traffic stats**: Bytes in/out per tunnel
- **Connection history**: Past tunnel sessions
- **Total uptime**: Cumulative connection time

All tracked automatically — no extra configuration needed.

## How It Works

lrok uses [frp](https://github.com/fatedier/frp) (Fast Reverse Proxy) under the hood, enhanced with:

1. **Pre-configured** to connect to `frp.lum.tools:7000` (no setup required)
2. **API Key Authentication** for secure tunnel creation
3. **Automatic HTTPS** via Let's Encrypt wildcard certificates
4. **Traffic Tracking** logged to your account automatically
5. **Embedded binaries** - frpc is bundled, nothing to install separately

### Architecture

```
Your App (localhost:8000)
    ↓
lrok CLI (local proxy + frpc)
    ↓ (secure tunnel)
frp.lum.tools (FRP server)
    ↓ (HTTPS with SSL)
Public Internet → https://your-tunnel.t.lum.tools
```

All traffic is encrypted end-to-end. The local dashboard intercepts requests for inspection without breaking the tunnel.

## Security

- All tunnels use HTTPS with valid SSL certificates
- API keys authenticate and authorize tunnel creation
- All activity is logged and trackable
- Rotate API keys anytime at [platform.lum.tools/keys](https://platform.lum.tools/keys)

## Authentication Commands

Manage your API key credentials:

```bash
# Login (saves key to ~/.lrok/config.toml)
lrok login lum_your_api_key_here

# Check authentication status
lrok whoami

# Logout (removes saved key)
lrok logout
```

**API Key Priority:**
1. `--api-key` flag (highest priority, allows temporary override)
2. `LUM_API_KEY` environment variable
3. `~/.lrok/config.toml` file (saved via `lrok login`)

**Security:** Config file is created with `0600` permissions (owner-only read/write).

## Troubleshooting

### "No API key configured"

Login with your API key:
```bash
lrok login lum_your_key
```

Or use environment variable:
```bash
export LUM_API_KEY='lum_your_key'
```

### "Invalid API key"

- Verify your key at [platform.lum.tools/keys](https://platform.lum.tools/keys)
- Ensure it starts with `lum_`
- Check for extra spaces or quotes
- Try running `lrok whoami` to check current auth

### Connection Issues

- Verify `frp.lum.tools` is reachable
- Check firewall isn't blocking port 7000
- Try a different network

## Contributing

Issues and PRs welcome at [github.com/lum-tools/lrok](https://github.com/lum-tools/lrok)

## License

MIT License - see LICENSE file

## Links

- Platform: [platform.lum.tools](https://platform.lum.tools)
- Tunnels Dashboard: [platform.lum.tools/tunnels](https://platform.lum.tools/tunnels)
- API Keys: [platform.lum.tools/keys](https://platform.lum.tools/keys)
- Blog: [blog.lum.tools](https://blog.lum.tools)
- GitHub: [github.com/lum-tools/lrok](https://github.com/lum-tools/lrok)

---

**Made with ❤️ by [platform.lum.tools](https://platform.lum.tools)**

