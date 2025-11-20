# lrok SDKs

Multi-language SDKs for programmatic tunnel management with lrok.

## Overview

lrok provides official SDKs for multiple programming languages, generated from our OpenAPI 3.0 specification. This ensures consistency, type safety, and comprehensive coverage across all supported languages.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     lrok Daemon (Go)                        │
│                   HTTP REST API Server                      │
│                 (Port 4243 by default)                      │
└─────────────────────────────────────────────────────────────┘
                            ▲
                            │ HTTP/JSON
                            │
        ┌───────────────────┼───────────────────┐
        │                   │                   │
    ┌───▼────┐         ┌────▼───┐         ┌────▼───┐
    │ Python │         │ Node.js│         │  Ruby  │
    │  SDK   │         │  SDK   │         │  SDK   │
    └────────┘         └────────┘         └────────┘
```

## Supported Languages

| Language | Status | Package Name | Version |
|----------|--------|--------------|---------|
| **Python** | ✅ Stable | `lrok` | 1.0.0 |
| **Node.js/TypeScript** | ✅ Stable | `lrok` | 1.0.0 |
| **Ruby** | 🚧 Beta | `lrok` | 1.0.0 |
| **Go** | 🚧 Beta | `github.com/lum-tools/lrok-go` | 1.0.0 |
| **Java** | 📋 Planned | `tools.lum:lrok` | - |
| **PHP** | 📋 Planned | `lum/lrok` | - |

## Quick Start

### Prerequisites

1. **Start lrok daemon**:
   ```bash
   lrok daemon --port 4243
   ```

2. **Set API key** (optional):
   ```bash
   export LUM_API_KEY="lum_your_api_key_here"
   ```

### Python

```python
from lrok import ApiClient, Configuration, TunnelsApi
from lrok.models import CreateTunnelRequest

# Configure client
config = Configuration(host="http://localhost:4243")
client = ApiClient(config)
api = TunnelsApi(client)

# Create HTTP tunnel
request = CreateTunnelRequest(
    type="http",
    local_port=8000,
    subdomain="my-app"
)

tunnel = api.create_tunnel(request)
print(f"Tunnel URL: {tunnel.public_url}")

# Use the tunnel...
# Your app is now accessible at tunnel.public_url

# Stop tunnel
api.delete_tunnel(tunnel.id)
```

### Node.js/TypeScript

```typescript
import { Configuration, TunnelsApi, CreateTunnelRequest } from 'lrok';

// Configure client
const config = new Configuration({
  basePath: 'http://localhost:4243'
});

const api = new TunnelsApi(config);

// Create HTTP tunnel
const request: CreateTunnelRequest = {
  type: 'http',
  localPort: 8000,
  subdomain: 'my-app'
};

const { data: tunnel } = await api.createTunnel(request);
console.log(`Tunnel URL: ${tunnel.publicUrl}`);

// Stop tunnel
await api.deleteTunnel(tunnel.id);
```

### Ruby

```ruby
require 'lrok'

# Configure client
Lrok.configure do |config|
  config.host = 'localhost:4243'
  config.scheme = 'http'
end

api = Lrok::TunnelsApi.new

# Create HTTP tunnel
request = Lrok::CreateTunnelRequest.new(
  type: 'http',
  local_port: 8000,
  subdomain: 'my-app'
)

tunnel = api.create_tunnel(request)
puts "Tunnel URL: #{tunnel.public_url}"

# Stop tunnel
api.delete_tunnel(tunnel.id)
```

## Features

### Tunnel Types

All SDKs support the following tunnel types:

- **HTTP/HTTPS** - Expose HTTP services with custom subdomains
- **TCP** - Direct TCP port forwarding
- **STCP** - Secure TCP with pre-shared keys
- **XTCP** - P2P tunneling with fallback
- **Visitor** - Connect to STCP/XTCP tunnels

### Core Operations

- ✅ Create tunnels with full configuration
- ✅ List active tunnels
- ✅ Get tunnel details
- ✅ Stop tunnels
- ✅ Get real-time statistics
- ✅ HTTP request inspection (HTTP tunnels)
- ✅ Subdomain management
- ✅ Authentication management

### Advanced Features

- **Request Inspection**: Capture and analyze HTTP requests
- **Statistics**: Real-time metrics (bytes, connections, uptime)
- **Health Checks**: Configure custom health checks
- **Transport Options**: Encryption, compression, bandwidth limiting
- **Error Handling**: Comprehensive error types and messages

## API Reference

### Tunnels

#### Create Tunnel

```python
# HTTP tunnel
tunnel = api.create_tunnel(CreateTunnelRequest(
    type="http",
    local_port=8000,
    subdomain="my-app",          # Optional custom subdomain
    encryption=True,              # Enable encryption
    compression=True,             # Enable compression
    bandwidth_limit="1MB"         # Limit bandwidth
))

# TCP tunnel
tunnel = api.create_tunnel(CreateTunnelRequest(
    type="tcp",
    local_port=5432,
    remote_port=10001,            # Required for TCP
    encryption=True
))

# STCP tunnel (secure)
tunnel = api.create_tunnel(CreateTunnelRequest(
    type="stcp",
    local_port=5432,
    secret_key="my-secret-key-12345",  # Min 8 chars
    name="secure-db"
))
```

#### List Tunnels

```python
response = api.list_tunnels()
for tunnel in response.tunnels:
    print(f"{tunnel.id}: {tunnel.public_url}")
```

#### Get Tunnel Details

```python
tunnel = api.get_tunnel(tunnel_id)
print(f"Status: {tunnel.status}")
print(f"Created: {tunnel.created_at}")
```

#### Stop Tunnel

```python
api.delete_tunnel(tunnel_id)
```

### Statistics

```python
stats = api.get_tunnel_stats(tunnel_id)
print(f"Bytes In: {stats.bytes_in}")
print(f"Bytes Out: {stats.bytes_out}")
print(f"Connections: {stats.connections}")
print(f"Uptime: {stats.uptime}s")
```

### Request Inspection (HTTP only)

```python
# Get request history
response = api.get_tunnel_requests(tunnel_id, limit=100)
for req in response.requests:
    print(f"{req.method} {req.path} - {req.status_code}")
    print(f"Duration: {req.duration}ms")
```

### Subdomain Management

```python
from lrok import SubdomainsApi

subdomains_api = SubdomainsApi(client)

# Reserve subdomain
subdomain = subdomains_api.reserve_subdomain({"name": "my-app"})

# List reserved subdomains
response = subdomains_api.list_subdomains()

# Delete subdomain
subdomains_api.delete_subdomain("my-app")
```

## Development

### Generating SDKs

SDKs are automatically generated from the OpenAPI specification:

```bash
# Generate all SDKs
./scripts/generate-sdks.sh

# Generated SDKs will be in:
# - sdk/python/
# - sdk/nodejs/
# - sdk/ruby/
# - sdk/go/
```

### Running Tests

```bash
# Run all SDK tests
./scripts/test-sdks.sh

# Run specific SDK tests
docker-compose -f docker-compose.test.yml run --rm sdk-test-python
docker-compose -f docker-compose.test.yml run --rm sdk-test-nodejs
```

### Test Coverage

Tests cover:

- ✅ All tunnel types (HTTP, TCP, STCP, XTCP, Visitor)
- ✅ All API endpoints
- ✅ Error handling and validation
- ✅ Concurrent operations
- ✅ Request inspection
- ✅ Statistics tracking
- ✅ Authentication flows

Target coverage: **100%** of OpenAPI endpoints

## CI/CD

The project uses GitHub Actions for continuous testing:

- **OpenAPI Validation**: Validates spec on every commit
- **SDK Generation**: Auto-generates SDKs from spec
- **E2E Tests**: Full integration tests for each SDK
- **Contract Tests**: Validates API matches OpenAPI spec
- **Coverage Reports**: Aggregated coverage across all SDKs
- **Performance Tests**: Load testing with k6

## OpenAPI Specification

The source of truth for all SDKs is the OpenAPI 3.0 specification:

- **Location**: `api/openapi.yaml`
- **Version**: 3.0.3
- **Format**: YAML

View the spec:
```bash
cat api/openapi.yaml
```

Validate the spec:
```bash
openapi-generator-cli validate -i api/openapi.yaml
```

## Examples

### Web Development

```python
# Test webhooks locally
import lrok
from flask import Flask

app = Flask(__name__)

@app.route('/webhook', methods=['POST'])
def webhook():
    return {'status': 'received'}

# Create tunnel
tunnel = api.create_tunnel(CreateTunnelRequest(
    type="http",
    local_port=5000
))

print(f"Webhook URL: {tunnel.public_url}/webhook")

# Run Flask app
app.run(port=5000)
```

### Database Access

```python
# Secure PostgreSQL tunnel
tunnel = api.create_tunnel(CreateTunnelRequest(
    type="stcp",
    local_port=5432,
    secret_key=os.environ["DB_SECRET"],
    name="prod-db"
))

# On client machine, create visitor
visitor = api.create_tunnel(CreateTunnelRequest(
    type="visitor",
    server_name="prod-db",
    secret_key=os.environ["DB_SECRET"],
    bind_port=5432
))

# Now connect to localhost:5432
```

### Testing/CI

```python
# Expose test server in CI
import pytest

@pytest.fixture(scope="session")
def tunnel():
    tunnel = api.create_tunnel(CreateTunnelRequest(
        type="http",
        local_port=8080,
        subdomain=f"ci-test-{os.environ['CI_RUN_ID']}"
    ))
    yield tunnel.public_url
    api.delete_tunnel(tunnel.id)

def test_external_api(tunnel):
    # Run tests that require external URL
    run_external_tests(tunnel)
```

## Troubleshooting

### Connection Refused

Ensure lrok daemon is running:
```bash
lrok daemon --port 4243
```

### Authentication Error

Set your API key:
```bash
export LUM_API_KEY="lum_your_key"
```

Or configure programmatically:
```python
config.api_key = {"Authorization": "Bearer lum_your_key"}
```

### Port Already in Use

Change daemon port:
```bash
lrok daemon --port 4244
```

Update client configuration:
```python
config = Configuration(host="http://localhost:4244")
```

## Contributing

### Adding a New Language

1. Add generator config in `api/generator-config-<language>.yaml`
2. Update `scripts/generate-sdks.sh`
3. Create test Dockerfile in `test/Dockerfile.<language>-tests`
4. Add test service to `docker-compose.test.yml`
5. Write E2E tests in `test/<language>/`
6. Update CI pipeline in `.github/workflows/sdk-tests.yml`

### Updating the API

1. Modify `api/openapi.yaml`
2. Validate: `openapi-generator-cli validate -i api/openapi.yaml`
3. Regenerate SDKs: `./scripts/generate-sdks.sh`
4. Run tests: `./scripts/test-sdks.sh`
5. Commit changes

## Resources

- **Main Repository**: https://github.com/lum-tools/lrok
- **Documentation**: https://docs.lum.tools/lrok
- **OpenAPI Spec**: `api/openapi.yaml`
- **Issues**: https://github.com/lum-tools/lrok/issues

## License

MIT License - see LICENSE file for details
