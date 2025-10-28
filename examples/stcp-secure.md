# STCP Tunnel Examples

STCP (Secret TCP) tunnels provide secure access without exposing the service publicly. They require a pre-shared secret key and a visitor on the other end.

## Secure Database Access

```bash
# Server side: Expose PostgreSQL securely
lrok stcp 5432 --secret-key my-secret-key --encrypt --compress

# Client side: Connect as visitor
lrok visitor tunnel-name --type stcp --secret-key my-secret-key --bind-port 5432

# Now connect locally:
psql -h 127.0.0.1 -p 5432 -U myuser mydb
```

## Secure SSH Access

```bash
# Server side: Expose SSH securely
lrok stcp 22 --secret-key ssh-secret-123

# Client side: Connect as visitor
lrok visitor tunnel-name --type stcp --secret-key ssh-secret-123 --bind-port 2222

# SSH locally:
ssh -p 2222 user@127.0.0.1
```

## Secure API Access

```bash
# Server side: Expose API securely
lrok stcp 8000 --secret-key api-secret-456 --encrypt

# Client side: Connect as visitor
lrok visitor tunnel-name --type stcp --secret-key api-secret-456 --bind-port 8000

# Access locally:
curl http://127.0.0.1:8000/api/health
```

## Secure File Server

```bash
# Server side: Expose file server securely
lrok stcp 8080 --secret-key file-secret-789 --compress

# Client side: Connect as visitor
lrok visitor tunnel-name --type stcp --secret-key file-secret-789 --bind-port 8080

# Access locally:
curl http://127.0.0.1:8080/files/
```

## Secure Redis Access

```bash
# Server side: Expose Redis securely
lrok stcp 6379 --secret-key redis-secret-abc

# Client side: Connect as visitor
lrok visitor tunnel-name --type stcp --secret-key redis-secret-abc --bind-port 6379

# Connect locally:
redis-cli -h 127.0.0.1 -p 6379
```

## Security Features

- **Pre-shared Secret**: Only clients with the secret can connect
- **Encryption**: Use `--encrypt` flag for additional security
- **Compression**: Use `--compress` flag for better performance
- **Bandwidth Limits**: Set limits with `--bandwidth` flag

## Use Cases

- **Internal Services**: Expose internal APIs securely
- **Database Access**: Secure database connections
- **Remote Administration**: Secure SSH/RDP access
- **File Sharing**: Secure file server access
- **Development**: Share dev environments securely

## Connection Examples

```bash
# PostgreSQL
psql -h 127.0.0.1 -p 5432 -U myuser mydb

# MySQL
mysql -h 127.0.0.1 -P 5432 -u myuser -p mydb

# Redis
redis-cli -h 127.0.0.1 -p 6379

# SSH
ssh -p 2222 user@127.0.0.1

# HTTP API
curl http://127.0.0.1:8000/api/health

# Generic TCP
telnet 127.0.0.1 5432
nc 127.0.0.1 5432
```

## Troubleshooting

- **Secret Key**: Ensure both sides use the same secret key
- **Visitor Connection**: Make sure visitor connects after server
- **Port Conflicts**: Use different bind ports for multiple visitors
- **Firewall**: Check local firewall settings
