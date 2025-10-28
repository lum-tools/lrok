# XTCP Tunnel Examples

XTCP (P2P) tunnels establish direct client-to-client connections when possible, bypassing the server for data transmission. This provides better performance for high-bandwidth applications.

## P2P File Transfer

```bash
# Server side: Expose file server via P2P
lrok xtcp 8080 --secret-key p2p-file-transfer

# Client side: Connect as visitor
lrok visitor tunnel-name --type xtcp --secret-key p2p-file-transfer --bind-port 8080

# Access locally: http://127.0.0.1:8080
# If P2P fails, falls back to server relay automatically
```

## P2P Web Development

```bash
# Developer side: Expose dev server via P2P
lrok xtcp 3000 --secret-key dev-p2p-key

# Client side: Connect as visitor
lrok visitor tunnel-name --type xtcp --secret-key dev-p2p-key --bind-port 3000

# Access locally: http://127.0.0.1:3000
# Direct connection for better performance
```

## P2P Video Streaming

```bash
# Server side: Expose video stream via P2P
lrok xtcp 8080 --secret-key video-stream-p2p

# Client side: Connect as visitor
lrok visitor tunnel-name --type xtcp --secret-key video-stream-p2p --bind-port 8080

# Stream locally: http://127.0.0.1:8080/stream
# Direct connection reduces latency
```

## P2P Database Access

```bash
# Server side: Expose database via P2P
lrok xtcp 5432 --secret-key db-p2p-secret

# Client side: Connect as visitor
lrok visitor tunnel-name --type xtcp --secret-key db-p2p-secret --bind-port 5432

# Connect locally:
psql -h 127.0.0.1 -p 5432 -U myuser mydb
```

## P2P Gaming

```bash
# Server side: Expose game server via P2P
lrok xtcp 7777 --secret-key game-p2p-key

# Client side: Connect as visitor
lrok visitor tunnel-name --type xtcp --secret-key game-p2p-key --bind-port 7777

# Game clients connect to: 127.0.0.1:7777
# Direct connection for lower latency
```

## P2P Performance Benefits

- **Lower Latency**: Direct connection reduces round-trip time
- **Higher Bandwidth**: No server bottleneck
- **Better Performance**: Especially for real-time applications
- **Automatic Fallback**: Falls back to server relay if P2P fails

## NAT Traversal

XTCP uses STUN servers to traverse NAT:
- **Automatic**: No manual configuration needed
- **STUN Servers**: Uses public STUN servers by default
- **Custom STUN**: Can specify custom STUN server if needed

## Connection Examples

```bash
# HTTP/Web
curl http://127.0.0.1:8080/
wget http://127.0.0.1:3000/

# Database
psql -h 127.0.0.1 -p 5432 -U myuser mydb
mysql -h 127.0.0.1 -P 5432 -u myuser -p mydb

# Game clients
# Configure to connect to 127.0.0.1:7777

# Generic TCP
telnet 127.0.0.1 8080
nc 127.0.0.1 8080
```

## Troubleshooting

- **P2P Failure**: Check NAT/firewall configuration
- **Fallback**: Server relay should work if P2P fails
- **STUN**: Ensure STUN servers are accessible
- **Secret Key**: Both sides must use the same secret key
- **Timing**: Connect visitor after server is running
