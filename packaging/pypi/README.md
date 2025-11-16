# lrok - Tunnel Service by lum.tools

Expose your local services to the internet with HTTPS and readable URLs.

```bash
lrok 8000
# → https://happy-dolphin.t.lum.tools
```

## Installation

```bash
pip install lrok
```

## Quick Start

1. Get your API key from [platform.lum.tools/keys](https://platform.lum.tools/keys)
2. Login with your API key:
   ```bash
   lrok login lum_your_api_key
   ```
3. Start a tunnel:
   ```bash
   lrok 8000
   ```

## Features

- **HTTP/TCP/STCP/XTCP tunnels** - Multiple protocols for different use cases
- **Reserved subdomains** - Claim permanent URLs that only you can use
- **Built-in request inspector** - Web dashboard at localhost:4242
- **Automatic HTTPS** - All tunnels secured with valid SSL certificates

## Subdomain Management

Reserve permanent tunnel URLs (up to 5 per user):

```bash
# List your reserved subdomains
lrok subdomains list

# Reserve a subdomain
lrok subdomains reserve my-app

# Release a subdomain
lrok subdomains delete my-app
```

## Documentation

Full documentation: [github.com/lum-tools/lrok](https://github.com/lum-tools/lrok)

## Links

- Platform: [platform.lum.tools](https://platform.lum.tools)
- Dashboard: [platform.lum.tools/tunnels](https://platform.lum.tools/tunnels)
- Subdomain Management: [lrok.lum.tools/subdomains](https://lrok.lum.tools/subdomains)
- Blog: [blog.lum.tools](https://blog.lum.tools)
- GitHub: [github.com/lum-tools/lrok](https://github.com/lum-tools/lrok)

## License

MIT

