# lrok - Expose Local Servers to the Internet

**lrok** (short for *lum-rok*) is a fast, secure tunnel service that exposes your localhost to the internet with HTTPS and human-readable URLs — available as both a **CLI tool** and **programmable SDKs**.

```bash
# CLI: Quick tunnel in seconds
lrok 8000
# → https://happy-dolphin.t.lum.tools
# → http://localhost:4242 (dashboard)

# SDK: Programmatic tunnel control
python -c "
from lrok import TunnelsApi, CreateTunnelRequest
tunnel = TunnelsApi().create_tunnel(CreateTunnelRequest(type='http', local_port=8000))
print(f'Tunnel: {tunnel.public_url}')
"
```

---

## 🎯 Why lrok?

**For Developers:**
- Test webhooks locally (Stripe, GitHub, Twilio, etc.)
- Share dev environments with teammates
- Demo work-in-progress to clients
- Remote access to local services

**For Automation:**
- Create tunnels programmatically in CI/CD
- Dynamic tunnel management in your applications
- Multi-language SDK support (Python, TypeScript, Ruby, Go)

**100% Free**: No credit card. No paid tiers. No usage limits (for now).

---

## 🚀 Quick Start

### CLI Usage

```bash
# Install
curl -fsSL https://platform.lum.tools/install.sh | bash

# Get API key (free, no credit card)
# Visit: https://platform.lum.tools/keys

# Authenticate
lrok login lum_your_api_key_here

# Create tunnel
lrok 8000
# → https://happy-dolphin.t.lum.tools
```

### SDK Usage

**Prerequisites:** Start the lrok daemon first:
```bash
lrok daemon --port 4243
```

**Python SDK:**
```python
from lrok import ApiClient, Configuration, TunnelsApi
from lrok.models import CreateTunnelRequest

# Configure
config = Configuration(host="http://localhost:4243")
api = TunnelsApi(ApiClient(config))

# Create tunnel
tunnel = api.create_tunnel(CreateTunnelRequest(
    type="http",
    local_port=8000,
    subdomain="my-app"
))

print(f"Tunnel URL: {tunnel.public_url}")

# Stop tunnel
api.delete_tunnel(tunnel.id)
```

**TypeScript/Node.js SDK:**
```typescript
import { Configuration, TunnelsApi, CreateTunnelRequest } from 'lrok';

const api = new TunnelsApi(new Configuration({
  basePath: 'http://localhost:4243'
}));

const { data: tunnel } = await api.createTunnel({
  type: 'http',
  localPort: 8000,
  subdomain: 'my-app'
});

console.log(`Tunnel URL: ${tunnel.publicUrl}`);
```

---

## ✨ Features

### CLI Features
- **🎯 Readable URLs**: `happy-dolphin.t.lum.tools` instead of random hashes
- **📊 Built-in Dashboard**: Beautiful request inspector at `http://localhost:4242`
- **🔒 HTTPS by Default**: Automatic SSL certificates
- **📈 Traffic Tracking**: Monitor usage at [platform.lum.tools/tunnels](https://platform.lum.tools/tunnels)
- **🌍 Cross-Platform**: Single binary for macOS, Linux, Windows
- **🔌 Multiple Protocols**: HTTP, TCP, STCP (secure), XTCP (P2P)

### SDK Features
- **🐍 Python SDK**: Full-featured with type hints
- **📦 TypeScript SDK**: Type-safe with Axios
- **💎 Ruby SDK**: Idiomatic Ruby interface
- **🐹 Go SDK**: Native Go implementation
- **🔄 REST API**: HTTP/JSON API on port 4243
- **📡 Real-time Stats**: Bytes, connections, uptime
- **🔍 Request Inspection**: Capture and analyze HTTP requests

---

## 📦 Installation

### CLI Installation

**Quick Install (macOS/Linux):**
```bash
curl -fsSL https://platform.lum.tools/install.sh | bash
```

**npm (Cross-platform):**
```bash
npm install -g lrok
```

**PyPI (Cross-platform):**
```bash
pip install lrok
```

**Direct Download:**
- [GitHub Releases](https://github.com/lum-tools/lrok/releases)

### SDK Installation

**Python:**
```bash
pip install lrok
```

**Node.js/TypeScript:**
```bash
npm install lrok
```

**Ruby:**
```bash
gem install lrok
```

**Go:**
```bash
go get github.com/lum-tools/lrok-go
```

---

## 🔐 Authentication

### 1. Get Your Free API Key

1. Visit [platform.lum.tools](https://platform.lum.tools)
2. Sign in (Google, GitHub, or email)
3. Navigate to [API Keys](https://platform.lum.tools/keys)
4. Create new key (starts with `lum_`)

### 2. Configure CLI

**Recommended:** Save to config file:
```bash
lrok login lum_your_api_key_here
```

**Alternative:** Use environment variable:
```bash
export LUM_API_KEY='lum_your_api_key_here'
```

**Check auth status:**
```bash
lrok whoami
# → ✅ Logged in as: user@example.com
```

---

## 🎮 CLI Usage

### Basic Commands

```bash
# Quick tunnel with random name
lrok 8000

# Custom subdomain
lrok 8000 -n my-app
# → https://my-app.t.lum.tools

# TCP tunnel (databases, SSH, etc.)
lrok tcp 5432 --remote-port 10001
# → frp.lum.tools:10001

# Secure tunnel with pre-shared key
lrok stcp 22 --secret-key my-secret

# P2P tunnel (direct connection)
lrok xtcp 3000 --secret-key p2p-key
```

### Available Commands

```
lrok [port]                  Quick HTTP tunnel
lrok http [port] [flags]     HTTP tunnel with options
lrok tcp <port> [flags]      TCP tunnel
lrok stcp <port> [flags]     Secure TCP tunnel
lrok xtcp <port> [flags]     P2P tunnel
lrok visitor <name> [flags]  Connect to secure/P2P tunnel
lrok daemon [flags]          Start API daemon
lrok subdomains              Manage reserved subdomains
lrok login <api-key>         Save API key
lrok logout                  Remove API key
lrok whoami                  Check auth status
lrok version                 Show version
```

### Common Flags

```
-n, --name string         Custom tunnel name
--subdomain string        Alias for --name
-k, --api-key string      Override API key
--ip string               Local IP (default: 127.0.0.1)
--remote-port int         Remote port (TCP only)
--secret-key string       Secret key (STCP/XTCP)
--encrypt                 Enable encryption
--compress                Enable compression
--bandwidth string        Limit bandwidth (e.g., "1MB")
```

---

## 💻 SDK Documentation

### Architecture

```
┌─────────────────────────────────────────────────┐
│         Your Application (Python/TS/etc)       │
│                                                  │
│  ┌────────────────────────────────────────┐   │
│  │         lrok SDK Client                 │   │
│  │  (HTTP REST API calls via port 4243)   │   │
│  └────────────────────────────────────────┘   │
└─────────────────────────────────────────────────┘
                      ↓ HTTP/JSON
┌─────────────────────────────────────────────────┐
│            lrok Daemon (Go)                      │
│       REST API Server (port 4243)               │
│  ┌────────────────────────────────────────┐   │
│  │  Tunnel Manager + FRP Client           │   │
│  └────────────────────────────────────────┘   │
└─────────────────────────────────────────────────┘
                      ↓ Secure Tunnel
┌─────────────────────────────────────────────────┐
│     frp.lum.tools (FRP Server)                   │
│          HTTPS + SSL Termination                │
└─────────────────────────────────────────────────┘
                      ↓
            Public Internet
    https://your-tunnel.t.lum.tools
```

### Python SDK Examples

**Create HTTP Tunnel:**
```python
from lrok import ApiClient, Configuration, TunnelsApi
from lrok.models import CreateTunnelRequest

config = Configuration(host="http://localhost:4243")
api = TunnelsApi(ApiClient(config))

tunnel = api.create_tunnel(CreateTunnelRequest(
    type="http",
    local_port=8000,
    subdomain="my-app",
    encryption=True,
    compression=True
))

print(f"Public URL: {tunnel.public_url}")
print(f"Tunnel ID: {tunnel.id}")
```

**List Active Tunnels:**
```python
response = api.list_tunnels()
for tunnel in response.tunnels:
    print(f"{tunnel.name}: {tunnel.public_url}")
```

**Get Tunnel Statistics:**
```python
stats = api.get_tunnel_stats(tunnel_id)
print(f"Bytes In: {stats.bytes_in}")
print(f"Bytes Out: {stats.bytes_out}")
print(f"Connections: {stats.connections}")
```

**Inspect HTTP Requests:**
```python
response = api.get_tunnel_requests(tunnel_id, limit=100)
for req in response.requests:
    print(f"{req.method} {req.path} - {req.status_code} ({req.duration}ms)")
```

### TypeScript SDK Examples

**Create HTTP Tunnel:**
```typescript
import { Configuration, TunnelsApi, CreateTunnelRequest } from 'lrok';

const config = new Configuration({ basePath: 'http://localhost:4243' });
const api = new TunnelsApi(config);

const { data: tunnel } = await api.createTunnel({
  type: 'http',
  localPort: 8000,
  subdomain: 'my-app',
  encryption: true,
  compression: true
});

console.log(`Public URL: ${tunnel.publicUrl}`);
```

**TCP Tunnel:**
```typescript
const { data: tunnel } = await api.createTunnel({
  type: 'tcp',
  localPort: 5432,
  remotePort: 10001,
  encryption: true
});

console.log(`Connect at: frp.lum.tools:${tunnel.remotePort}`);
```

**Manage Subdomains:**
```typescript
import { SubdomainsApi } from 'lrok';

const subdomainsApi = new SubdomainsApi(config);

// Reserve subdomain
await subdomainsApi.reserveSubdomain({ name: 'my-app' });

// List reserved
const { data } = await subdomainsApi.listSubdomains();
console.log(data.subdomains);
```

### Supported Operations

All SDKs support:
- ✅ Create tunnels (HTTP, TCP, STCP, XTCP, Visitor)
- ✅ List active tunnels
- ✅ Get tunnel details
- ✅ Delete tunnels
- ✅ Get real-time statistics
- ✅ Inspect HTTP requests
- ✅ Reserve/manage subdomains
- ✅ Authentication management

---

## 📖 Examples

### Webhook Testing (Flask + Python SDK)

```python
from flask import Flask, request
from lrok import TunnelsApi, ApiClient, Configuration
from lrok.models import CreateTunnelRequest

app = Flask(__name__)

# Create tunnel
config = Configuration(host="http://localhost:4243")
api = TunnelsApi(ApiClient(config))
tunnel = api.create_tunnel(CreateTunnelRequest(
    type="http",
    local_port=5000,
    subdomain="webhook-test"
))

print(f"Webhook URL: {tunnel.public_url}/webhook")

@app.route('/webhook', methods=['POST'])
def webhook():
    data = request.json
    print(f"Received webhook: {data}")
    return {'status': 'received'}

app.run(port=5000)
```

### CI/CD Testing (TypeScript)

```typescript
import { TunnelsApi, Configuration } from 'lrok';

async function setupTestTunnel() {
  const api = new TunnelsApi(
    new Configuration({ basePath: 'http://localhost:4243' })
  );

  const { data: tunnel } = await api.createTunnel({
    type: 'http',
    localPort: 8080,
    subdomain: `ci-test-${process.env.CI_RUN_ID}`
  });

  console.log(`Test at: ${tunnel.publicUrl}`);

  return async () => {
    await api.deleteTunnel(tunnel.id);
  };
}

// In your test
const cleanup = await setupTestTunnel();
try {
  // Run tests...
} finally {
  await cleanup();
}
```

### Database Access (Secure STCP)

```bash
# Server: Expose PostgreSQL securely
lrok stcp 5432 --secret-key db-secret-key-123 --name prod-db

# Client: Create visitor connection
lrok visitor prod-db --type stcp --secret-key db-secret-key-123 --bind-port 5432

# Now connect locally
psql -h 127.0.0.1 -p 5432 -U postgres
```

---

## 🏗️ Architecture

lrok is built on [frp](https://github.com/fatedier/frp) (Fast Reverse Proxy) with enhancements:

### Components

1. **lrok CLI**: Command-line interface for quick tunneling
2. **lrok Daemon**: HTTP REST API server (port 4243) for SDK access
3. **FRP Client**: Embedded frp client for tunnel connections
4. **FRP Server**: `frp.lum.tools:7000` (managed by lum.tools)
5. **HTTPS Gateway**: Automatic SSL with Let's Encrypt

### How It Works

```
┌──────────────────────────────────────────────────────────┐
│                   Your Local Machine                      │
│                                                            │
│  ┌─────────────┐    ┌──────────────┐   ┌─────────────┐  │
│  │   Your App  │←───│  lrok Daemon │───│  lrok CLI   │  │
│  │ (port 8000) │    │  (port 4243) │   │   or SDK    │  │
│  └─────────────┘    └──────────────┘   └─────────────┘  │
│                            ↓                               │
│                     ┌──────────────┐                      │
│                     │  FRP Client  │                      │
│                     │  (embedded)  │                      │
│                     └──────────────┘                      │
└──────────────────────────┼───────────────────────────────┘
                           ↓ Encrypted Tunnel
                    ┌──────────────┐
                    │  FRP Server  │
                    │ frp.lum.tools│
                    │   (port 7000)│
                    └──────────────┘
                           ↓
                    ┌──────────────┐
                    │ HTTPS Gateway│
                    │  SSL + DNS   │
                    └──────────────┘
                           ↓
              https://your-tunnel.t.lum.tools
```

### Key Features

- **Pre-configured**: Connects to `frp.lum.tools:7000` automatically
- **API Key Auth**: Secure tunnel creation with your lum.tools account
- **Automatic HTTPS**: Let's Encrypt wildcard certificates
- **Traffic Tracking**: All activity logged to your account
- **Request Inspector**: Built-in HTTP request/response viewer (port 4242)

---

## 🔧 Development & Publishing

### For Maintainers: SDK Generation

SDKs are automatically generated from the OpenAPI specification (`api/openapi.yaml`):

```bash
# Generate all SDKs
./scripts/generate-sdks.sh

# Run SDK tests
./scripts/test-sdks.sh

# Generated in:
# - sdk/python/
# - sdk/nodejs/
# - sdk/ruby/
# - sdk/go/
```

### Publishing SDKs

#### Python SDK (PyPI)

```bash
# Build
cd sdk/python
python setup.py sdist bdist_wheel

# Test upload
twine upload --repository testpypi dist/*

# Production upload
twine upload dist/*
```

**Required credentials:**
```bash
# ~/.pypirc
[pypi]
username = __token__
password = pypi-your-token-here
```

#### TypeScript SDK (npm)

```bash
# Build
cd sdk/nodejs
npm install
npm run build

# Test publish
npm publish --dry-run

# Production publish
npm publish --access public
```

**Required credentials:**
```bash
# Login to npm
npm login

# Or use token
echo "//registry.npmjs.org/:_authToken=\${NPM_TOKEN}" > ~/.npmrc
```

**Note:** Only Python and TypeScript SDKs are published to package registries. Other language SDKs (Ruby, Go) are generated but distributed via direct download or included in the main repository.

### CI/CD Integration

GitHub Actions automatically:
- Validates OpenAPI spec
- Generates SDKs
- Runs E2E tests (Python, TypeScript)
- Publishes to PyPI and npm on release tags

See `.github/workflows/sdk-tests.yml` for details.

---

## 📊 Request Inspector Dashboard

Every HTTP tunnel includes a beautiful web dashboard at `http://localhost:4242`:

**Features:**
- Real-time request list
- Click to view full request/response
- Headers inspection (in/out)
- Body viewer (JSON, HTML, text)
- Status codes & timing
- Copy as cURL command

Perfect for debugging webhooks and API integrations!

---

## 🎯 Reserved Subdomains

Reserve permanent tunnel URLs exclusively for your use:

```bash
# List reserved subdomains
lrok subdomains list

# Reserve a subdomain
lrok subdomains reserve my-app
# → Only YOU can use my-app.t.lum.tools

# Release subdomain
lrok subdomains delete my-app
```

**Benefits:**
- Consistent URLs across sessions
- Exclusive access
- Professional appearance
- Up to 5 free reservations per account

---

## 🐛 Troubleshooting

### "No API key configured"

```bash
# Login
lrok login lum_your_key

# Or export
export LUM_API_KEY='lum_your_key'

# Check status
lrok whoami
```

### "Connection refused" (SDK)

Make sure daemon is running:
```bash
lrok daemon --port 4243
```

### "Invalid API key"

- Verify at [platform.lum.tools/keys](https://platform.lum.tools/keys)
- Ensure it starts with `lum_`
- No extra spaces or quotes

### Port Already in Use

Change daemon port:
```bash
# Start daemon on different port
lrok daemon --port 4244

# Update SDK config
config = Configuration(host="http://localhost:4244")
```

---

## 🤝 Contributing

Issues and PRs welcome at [github.com/lum-tools/lrok](https://github.com/lum-tools/lrok)

### Adding New Features

1. Update `api/openapi.yaml` with new endpoints
2. Regenerate SDKs: `./scripts/generate-sdks.sh`
3. Add tests
4. Submit PR

---

## 📄 License

MIT License - see [LICENSE](LICENSE) file

---

## 🔗 Links

- **lrok Interface**: [lrok.lum.tools](https://lrok.lum.tools) - Dedicated web interface for lrok
- **Platform**: [platform.lum.tools](https://platform.lum.tools) - Account management
- **Tunnels Dashboard**: [platform.lum.tools/tunnels](https://platform.lum.tools/tunnels) - Track your tunnels
- **API Keys**: [platform.lum.tools/keys](https://platform.lum.tools/keys) - Manage API keys
- **Blog**: [blog.lum.tools](https://blog.lum.tools) - Updates and tutorials
- **GitHub**: [github.com/lum-tools/lrok](https://github.com/lum-tools/lrok) - Source code
- **OpenAPI Spec**: [`api/openapi.yaml`](api/openapi.yaml) - API specification
- **SDK Documentation**: [`SDK_README.md`](SDK_README.md) - Detailed SDK docs

---

**Made with ❤️ by [lum.tools](https://platform.lum.tools)**

*Expose localhost. Build faster. Ship with confidence.*
