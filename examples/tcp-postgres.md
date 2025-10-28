# TCP Tunnel Examples

TCP tunnels provide direct port forwarding without HTTP/HTTPS overhead. Perfect for databases, SSH, Redis, and other TCP-based services.

## PostgreSQL Database

```bash
# Start PostgreSQL (running on port 5432)
sudo systemctl start postgresql

# Create TCP tunnel
lrok tcp 5432 --remote-port 10001

# Connect from anywhere:
psql -h frp.lum.tools -p 10001 -U myuser mydb
```

## SSH Server

```bash
# Start SSH server (running on port 22)
sudo systemctl start ssh

# Create TCP tunnel with encryption and compression
lrok tcp 22 --remote-port 10002 --encrypt --compress

# SSH from anywhere:
ssh -p 10002 user@frp.lum.tools
```

## Redis Server

```bash
# Start Redis (running on port 6379)
redis-server

# Create TCP tunnel
lrok tcp 6379 --remote-port 10003

# Connect from anywhere:
redis-cli -h frp.lum.tools -p 10003
```

## MySQL Database

```bash
# Start MySQL (running on port 3306)
sudo systemctl start mysql

# Create TCP tunnel
lrok tcp 3306 --remote-port 10004

# Connect from anywhere:
mysql -h frp.lum.tools -P 10004 -u myuser -p mydb
```

## MongoDB

```bash
# Start MongoDB (running on port 27017)
sudo systemctl start mongod

# Create TCP tunnel
lrok tcp 27017 --remote-port 10005

# Connect from anywhere:
mongo frp.lum.tools:10005
```

## Custom TCP Service

```bash
# Start your custom TCP service (running on port 8080)
./my-tcp-server --port 8080

# Create TCP tunnel with bandwidth limit
lrok tcp 8080 --remote-port 10006 --bandwidth 1MB

# Connect from anywhere:
telnet frp.lum.tools 10006
```

## Security Considerations

- **Encryption**: Use `--encrypt` flag for sensitive data
- **Compression**: Use `--compress` flag for better performance
- **Bandwidth**: Set limits with `--bandwidth` flag
- **Health Checks**: Enable with `--health-check` flag

## Connection Examples

```bash
# PostgreSQL
psql -h frp.lum.tools -p 10001 -U myuser mydb

# MySQL
mysql -h frp.lum.tools -P 10004 -u myuser -p mydb

# Redis
redis-cli -h frp.lum.tools -p 10003

# MongoDB
mongo frp.lum.tools:10005

# SSH
ssh -p 10002 user@frp.lum.tools

# Generic TCP
telnet frp.lum.tools 10006
nc frp.lum.tools 10006
```

## Troubleshooting

- **Connection refused**: Check if your local service is running
- **Port conflicts**: Try different remote ports
- **Firewall**: Ensure local service accepts connections
- **Authentication**: Configure your service's authentication properly
