# Deployment Guide

**Version:** 1.0.0  
**Last Updated:** 2025-11-22

## Overview

This guide covers deploying Steri-Connect in various environments, from development to production.

## Deployment Options

### Option 1: Standalone Executable (Recommended)

The application is distributed as a single executable file with no external dependencies.

#### Windows

1. Download `steri-connect-go.exe`
2. Create deployment directory:
   ```powershell
   mkdir C:\SteriConnect
   cd C:\SteriConnect
   ```
3. Copy executable and configuration:
   ```powershell
   copy steri-connect-go.exe C:\SteriConnect\
   copy config\config.yaml C:\SteriConnect\config.yaml
   ```
4. Create data directory:
   ```powershell
   mkdir data
   ```
5. Run application:
   ```powershell
   .\steri-connect-go.exe
   ```

#### Linux

1. Download `steri-connect-go`
2. Create deployment directory:
   ```bash
   mkdir -p /opt/stericonnect
   cd /opt/stericonnect
   ```
3. Copy executable and configuration:
   ```bash
   cp steri-connect-go /opt/stericonnect/
   cp config/config.yaml /opt/stericonnect/config.yaml
   chmod +x steri-connect-go
   ```
4. Create data directory:
   ```bash
   mkdir -p data
   ```
5. Run application:
   ```bash
   ./steri-connect-go
   ```

### Option 2: Systemd Service (Linux)

Create systemd service file:

```ini
# /etc/systemd/system/stericonnect.service
[Unit]
Description=Steri-Connect Middleware Service
After=network.target

[Service]
Type=simple
User=stericonnect
WorkingDirectory=/opt/stericonnect
ExecStart=/opt/stericonnect/steri-connect-go
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

Enable and start service:

```bash
sudo systemctl daemon-reload
sudo systemctl enable stericonnect
sudo systemctl start stericonnect
sudo systemctl status stericonnect
```

### Option 3: Windows Service

Use NSSM (Non-Sucking Service Manager) or similar tool:

```powershell
# Install NSSM
# Download from https://nssm.cc/download

# Install service
nssm install SteriConnect "C:\SteriConnect\steri-connect-go.exe"
nssm set SteriConnect AppDirectory "C:\SteriConnect"
nssm set SteriConnect AppStdout "C:\SteriConnect\logs\service.log"
nssm set SteriConnect AppStderr "C:\SteriConnect\logs\service-error.log"

# Start service
nssm start SteriConnect
```

## Configuration

### Production Configuration

Edit `config/config.yaml` for production:

```yaml
server:
  port: 8080
  bind_address: "127.0.0.1"  # Use "0.0.0.0" only if needed

database:
  path: "/var/lib/stericonnect/data/steri-connect.db"

logging:
  level: "INFO"
  format: "json"
  output: "/var/log/stericonnect/app.log"
  max_file_size_mb: 50
  max_backups: 10
  compress: true

auth:
  api_key_required: true  # Enable for network access
  api_key: "${API_KEY}"   # Use environment variable

test_ui:
  enabled: false  # Disable in production
  require_auth: true
```

### Environment Variables

Override configuration with environment variables:

```bash
export STERI_CONNECT_API_KEY="your-secure-api-key"
export STERI_CONNECT_SERVER_PORT=8080
export STERI_CONNECT_DATABASE_PATH="/var/lib/stericonnect/data/steri-connect.db"
```

## Security Considerations

### Network Access

**Default:** Application binds to `127.0.0.1` (localhost only) for security.

**For Network Access:**

1. Set `bind_address: "0.0.0.0"` in config
2. **REQUIRED:** Enable API key authentication:
   ```yaml
   auth:
     api_key_required: true
     api_key: "strong-random-api-key-here"
   ```
3. Use firewall rules to restrict access
4. Consider reverse proxy (nginx, Apache) for HTTPS

### API Key Security

- Generate strong random API keys (minimum 32 characters)
- Store API keys securely (environment variables, secrets manager)
- Rotate API keys regularly
- Never commit API keys to version control

### Test UI

- **Disable in production:** `test_ui.enabled: false`
- If enabled, use authentication: `test_ui.require_auth: true`
- Restrict access via firewall rules

## Reverse Proxy Setup

### Nginx Configuration

```nginx
server {
    listen 80;
    server_name stericonnect.example.com;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location /ws {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}
```

### HTTPS with Let's Encrypt

```nginx
server {
    listen 443 ssl;
    server_name stericonnect.example.com;

    ssl_certificate /etc/letsencrypt/live/stericonnect.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/stericonnect.example.com/privkey.pem;

    location / {
        proxy_pass http://127.0.0.1:8080;
        # ... proxy settings
    }
}
```

## Database Management

### Backup

```bash
# Manual backup
cp /var/lib/stericonnect/data/steri-connect.db \
   /backup/steri-connect-$(date +%Y%m%d-%H%M%S).db

# Automated backup script
#!/bin/bash
BACKUP_DIR="/backup/stericonnect"
DB_PATH="/var/lib/stericonnect/data/steri-connect.db"
DATE=$(date +%Y%m%d-%H%M%S)

mkdir -p $BACKUP_DIR
cp $DB_PATH "$BACKUP_DIR/steri-connect-$DATE.db"

# Keep only last 30 days
find $BACKUP_DIR -name "*.db" -mtime +30 -delete
```

### Restore

```bash
# Stop service
sudo systemctl stop stericonnect

# Restore database
cp /backup/steri-connect-20251122-120000.db \
   /var/lib/stericonnect/data/steri-connect.db

# Set permissions
chown stericonnect:stericonnect /var/lib/stericonnect/data/steri-connect.db

# Start service
sudo systemctl start stericonnect
```

## Monitoring

### Health Checks

Monitor health endpoint:

```bash
# Simple health check script
#!/bin/bash
HEALTH_URL="http://localhost:8080/api/health"
RESPONSE=$(curl -s $HEALTH_URL)
STATUS=$(echo $RESPONSE | jq -r '.status')

if [ "$STATUS" != "ok" ]; then
    echo "Health check failed: $STATUS"
    exit 1
fi
```

### Log Monitoring

Monitor application logs:

```bash
# View logs
tail -f /var/log/stericonnect/app.log

# Search for errors
grep ERROR /var/log/stericonnect/app.log

# Log rotation handled automatically by application
```

### Metrics Monitoring

Monitor metrics endpoint:

```bash
curl http://localhost:8080/api/metrics | jq
```

Key metrics to monitor:
- `uptime_seconds` - Service uptime
- `active_device_connections` - Connected devices
- `total_api_requests` - API usage
- `database_size_mb` - Database growth

## Performance Tuning

### Database Optimization

SQLite WAL mode is enabled by default for better concurrency. For high-write scenarios:

```sql
-- Increase cache size
PRAGMA cache_size = -64000;  -- 64MB

-- Enable synchronous NORMAL (faster, still safe)
PRAGMA synchronous = NORMAL;
```

### Polling Intervals

Adjust device polling intervals based on requirements:

```yaml
devices:
  melag:
    status_poll_interval: 2    # Lower = more frequent updates, higher CPU
  getinge:
    ping_interval: 15          # Lower = more frequent checks, higher network usage
```

### Log Rotation

Configure log rotation to manage disk space:

```yaml
logging:
  max_file_size_mb: 50         # Rotate when file reaches 50MB
  max_backups: 10              # Keep 10 backup files
  compress: true               # Compress old logs
```

## Troubleshooting

### Service Won't Start

1. Check logs: `journalctl -u stericonnect -n 50`
2. Verify configuration syntax: `yamllint config/config.yaml`
3. Check port availability: `netstat -tuln | grep 8080`
4. Verify file permissions

### Database Locked

1. Check for multiple instances running
2. Verify database file permissions
3. Check disk space
4. Review SQLite WAL files

### High Memory Usage

1. Monitor with metrics endpoint
2. Review log buffer size
3. Check for memory leaks in device adapters
4. Consider reducing polling intervals

## Updates and Upgrades

### Upgrading Application

1. **Backup database:**
   ```bash
   cp data/steri-connect.db backup/
   ```

2. **Stop service:**
   ```bash
   sudo systemctl stop stericonnect
   ```

3. **Replace executable:**
   ```bash
   cp new-version/steri-connect-go /opt/stericonnect/
   ```

4. **Start service:**
   ```bash
   sudo systemctl start stericonnect
   ```

5. **Verify:**
   ```bash
   curl http://localhost:8080/api/health
   ```

### Database Migrations

Database migrations run automatically on startup. Ensure backups are current before upgrading.

## Disaster Recovery

### Recovery Plan

1. **Database Corruption:**
   - Restore from latest backup
   - Verify database integrity: `sqlite3 database.db "PRAGMA integrity_check;"`

2. **Service Failure:**
   - Check logs for errors
   - Verify configuration
   - Restart service

3. **Data Loss:**
   - Restore from backup
   - Review audit logs for data reconstruction

## Support

For deployment issues:
- Review logs: `/var/log/stericonnect/app.log`
- Check health endpoint: `GET /api/health`
- Review diagnostics: `GET /api/diagnostics/{deviceId}`
- Consult Troubleshooting Guide: `docs/Troubleshooting-Guide.md`

