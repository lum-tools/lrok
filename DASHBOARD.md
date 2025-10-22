# lrok Built-in Dashboard

Every lrok tunnel includes a beautiful web dashboard at `http://localhost:4242`.

## What You See

When you run `lrok 8000`, open http://localhost:4242 to see:

```
┌───────────────────────────────────────────────────────────────┐
│                                                               │
│                            lrok                               │
│                      ● Connected                              │
│                                                               │
├───────────────────────────────────────────────────────────────┤
│  🌐 Public URL                                                │
│                                                               │
│  https://happy-dolphin.t.lum.tools                            │
│                                                               │
│  📍 Forwarding to: localhost:8000                             │
├───────────────────────────────────────────────────────────────┤
│  📊 Statistics                                                │
│                                                               │
│  ┌──────────────┬──────────────┬──────────────┬────────────┐ │
│  │ Data Recv    │ Data Sent    │ Connections  │ Uptime     │ │
│  │ 1.2 MB       │ 856 KB       │ 42           │ 0h 5m 23s  │ │
│  └──────────────┴──────────────┴──────────────┴────────────┘ │
│                                                               │
├───────────────────────────────────────────────────────────────┤
│  💡 Tips                                                      │
│                                                               │
│  • Share the public URL with anyone                          │
│  • All traffic is encrypted with HTTPS                       │
│  • View full stats at platform.lum.tools/tunnels             │
│  • Press Ctrl+C in terminal to stop                          │
│                                                               │
└───────────────────────────────────────────────────────────────┘
```

## Features

- **Real-time Updates**: Stats refresh every second
- **Beautiful Design**: Matches lum.tools platform branding
- **Mobile Responsive**: Works on any device
- **Zero Config**: Starts automatically
- **Smart Port**: Prefers 4242, finds available port if occupied

## API Endpoint

The dashboard also exposes a JSON API:

```bash
curl http://localhost:4242/api/stats
```

Returns:
```json
{
  "tunnel_name": "happy-dolphin",
  "public_url": "https://happy-dolphin.t.lum.tools",
  "local_port": 8000,
  "status": "Connected",
  "start_time": "2025-10-22T18:00:00Z",
  "bytes_in": 1234567,
  "bytes_out": 876543,
  "connections": 42
}
```

This allows you to:
- Build custom monitoring tools
- Integrate with your CI/CD
- Create browser extensions
- Display stats in other apps

## Technical Details

- **Language**: Pure Go
- **Dependencies**: stdlib only (net/http, encoding/json)
- **Code Size**: ~200 lines
- **Binary Impact**: Negligible (~0.5%)
- **Memory**: Minimal (single goroutine)
- **Performance**: Handles hundreds of concurrent connections

## Customization

While the default port is 4242, lrok will automatically find an available port if needed. The actual port is always shown in the terminal output.

To disable the dashboard (not recommended):
```bash
# Future flag: lrok 8000 --no-dashboard
```

## Comparison

| Tool        | Web Dashboard | Port    |
|-------------|---------------|---------|
| ngrok       | ✅            | 4040    |
| localtunnel | ❌            | -       |
| bore        | ❌            | -       |
| **lrok**    | ✅            | **4242** |

