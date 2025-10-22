# lrok Features

## 🚀 Built-in Dashboard

Every lrok tunnel automatically starts a local web dashboard at `http://localhost:4242`.

### Dashboard Features:

- **Live Status**: Real-time connection status with animated indicator
- **Public URL**: Shows your tunnel's public HTTPS URL
- **Traffic Stats**: 
  - Data received (bytes)
  - Data sent (bytes)
  - Active connections
  - Uptime counter
- **Auto-Update**: Stats refresh every second
- **Zero Config**: Starts automatically, no setup needed
- **Mobile Responsive**: Works on any device
- **lum.tools Branding**: Matches the platform's beautiful design

### Access the Dashboard:

When you start a tunnel:
```bash
lrok 8000
```

The output shows:
```
🚀 Starting lrok tunnel...
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  📍 Local:      http://127.0.0.1:8000
  🌐 Public URL: https://happy-dolphin.t.lum.tools
  🏷️  Name:       happy-dolphin
  📊 Dashboard:  http://localhost:4242  ← Open this!
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

Just open `http://localhost:4242` in your browser!

### Technical Details:

- **Pure Go**: Minimal HTTP server, no external dependencies
- **Tiny Footprint**: ~200 lines of code, negligible binary size impact
- **Auto Port**: Prefers port 4242, finds available port if occupied
- **Graceful Shutdown**: Stops cleanly when tunnel closes

