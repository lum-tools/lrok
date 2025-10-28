# UDP Tunnel Examples

UDP tunnels are perfect for DNS servers, game servers, and other UDP-based services. Note: UDP tunnels don't support encryption or compression due to UDP's stateless nature.

## DNS Server

```bash
# Start DNS server (running on port 53)
sudo systemctl start bind9

# Create UDP tunnel
lrok udp 53 --remote-port 10001

# Query from anywhere:
dig @frp.lum.tools -p 10001 google.com
nslookup google.com frp.lum.tools -port=10001
```

## Game Server

```bash
# Start game server (running on port 7777)
./game-server --port 7777

# Create UDP tunnel
lrok udp 7777 --remote-port 10002

# Players connect to: frp.lum.tools:10002
```

## NTP Server

```bash
# Start NTP server (running on port 123)
sudo systemctl start ntp

# Create UDP tunnel
lrok udp 123 --remote-port 10003

# Sync time from anywhere:
ntpdate -p 1 frp.lum.tools
```

## Custom UDP Service

```bash
# Start your custom UDP service (running on port 8080)
./my-udp-server --port 8080

# Create UDP tunnel with bandwidth limit
lrok udp 8080 --remote-port 10004 --bandwidth 500KB

# Connect from anywhere:
nc -u frp.lum.tools 10004
```

## VoIP Server

```bash
# Start VoIP server (running on port 5060)
./voip-server --port 5060

# Create UDP tunnel
lrok udp 5060 --remote-port 10005

# SIP clients connect to: frp.lum.tools:10005
```

## Performance Considerations

- **Latency**: UDP tunnels have lower latency than TCP
- **Bandwidth**: Set appropriate limits with `--bandwidth` flag
- **Packet Loss**: UDP doesn't guarantee delivery
- **NAT Traversal**: May require additional configuration

## Connection Examples

```bash
# DNS queries
dig @frp.lum.tools -p 10001 google.com
nslookup google.com frp.lum.tools -port=10001

# Game clients
# Configure game client to connect to frp.lum.tools:10002

# NTP sync
ntpdate -p 1 frp.lum.tools
chrony sources -v

# Generic UDP
nc -u frp.lum.tools 10004
socat UDP:frp.lum.tools:10004 -
```

## Troubleshooting

- **No response**: Check if your local UDP service is running
- **Firewall**: Ensure UDP packets can reach your service
- **NAT**: Some NAT configurations may block UDP
- **Packet loss**: Monitor network quality
