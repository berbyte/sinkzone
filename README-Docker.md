# Sinkzone with Unbound Docker Setup

A Docker Compose setup for Sinkzone with Unbound as the upstream DNS resolver.

## Quick Start

1. **Create environment file:**
   ```bash
   cat > .env << 'EOF'
   DOCKER_NETWORK_SUBNET=172.30.0.0/16
   SINKZONE_IP=172.30.0.1
   UNBOUND_IP=172.30.0.2
   SINKZONE_UPSTREAM_NAMESERVERS=unbound
   EOF
   ```

2. **Start services:**
   ```bash
   docker compose up -d
   ```

3. **Test DNS resolution:**
   ```bash
   dig @127.0.0.1 -p 5353 google.com
   ```

## Usage

### DNS Resolution
- **Sinkzone**: `dig @127.0.0.1 -p 5353 google.com`
- **Direct Unbound**: `dig @127.0.0.1 -p 5335 google.com`

### API Access
```bash
# Health check
curl http://localhost:8080/health

# View DNS queries
curl http://localhost:8080/api/queries

# Enable focus mode
curl -X POST http://localhost:8080/api/focus \
  -H "Content-Type: application/json" \
  -d '{"enabled": true, "duration": "1h"}'
```

### Sinkzone Commands via Docker Exec

```bash
# Terminal UI
docker compose exec -it sinkzone ./sinkzone tui --api-url http://localhost:8080

# Monitor DNS requests
docker compose exec -it sinkzone ./sinkzone monitor --api-url http://localhost:8080

# Add domain to allowlist
docker compose exec -it sinkzone ./sinkzone allowlist add google.com

# Remove domain from allowlist
docker compose exec -it sinkzone ./sinkzone allowlist remove google.com

# List allowlist
docker compose exec -it sinkzone ./sinkzone allowlist list

# Enable focus mode
docker compose exec -it sinkzone ./sinkzone focus start

# Disable focus mode
docker compose exec -it sinkzone ./sinkzone focus --disable

# Check status
docker compose exec -it sinkzone ./sinkzone status
```

## Ports

- **5353**: Sinkzone DNS server
- **8080**: Sinkzone API server  
- **5335**: Unbound DNS server (direct access)

## Architecture

```
Client → Sinkzone (5353) → Unbound (172.30.0.2) → Internet
```

## Troubleshooting

```bash
# View logs
docker compose logs -f

# Restart services
docker compose restart

# Rebuild and restart
docker compose up -d --build
``` 