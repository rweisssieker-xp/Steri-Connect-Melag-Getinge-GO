# Troubleshooting Guide

**Version:** 1.0.0  
**Last Updated:** 2025-11-22

## Common Issues

### Application Won't Start

#### Port Already in Use

**Symptoms:**
- Error: "failed to listen on 127.0.0.1:8080: bind: address already in use"

**Solution:**
1. Find process using port:
   ```bash
   # Linux/Mac
   lsof -i :8080
   # Windows
   netstat -ano | findstr :8080
   ```
2. Stop the process or change port in `config/config.yaml`:
   ```yaml
   server:
     port: 8081  # Use different port
   ```

#### Configuration File Error

**Symptoms:**
- Error: "Failed to load configuration"

**Solution:**
1. Verify YAML syntax:
   ```bash
   yamllint config/config.yaml
   ```
2. Check file encoding (must be UTF-8)
3. Verify all required fields are present
4. Check indentation (YAML is sensitive to spaces)

#### Database Permission Error

**Symptoms:**
- Error: "Failed to initialize database: permission denied"

**Solution:**
1. Check directory permissions:
   ```bash
   ls -la data/
   ```
2. Create data directory with proper permissions:
   ```bash
   mkdir -p data
   chmod 755 data
   ```
3. Verify write permissions:
   ```bash
   touch data/test.txt && rm data/test.txt
   ```

---

### Device Connection Issues

#### Device Not Connecting

**Symptoms:**
- Device status shows "offline" or "disconnected"
- Connection attempts fail

**Diagnosis:**
1. Check device diagnostics:
   ```bash
   curl http://localhost:8080/api/diagnostics/{deviceId}
   ```
2. Verify network connectivity:
   ```bash
   ping <device-ip>
   ```
3. Check device IP address in configuration
4. Review connection test results in diagnostics

**Solutions:**

**Network Connectivity:**
- Verify device is on same network
- Check firewall rules
- Verify IP address is correct
- Test with ping/telnet

**Melag Devices:**
- Verify MELAnet Box is accessible
- Check FTP port (usually 21)
- Verify MELAnet Box IP configuration
- Review FTP connection logs

**Getinge Devices:**
- Verify ICMP ping is not blocked
- Check device is powered on
- Review ping history in diagnostics
- Verify ping interval configuration

#### Device Keeps Disconnecting

**Symptoms:**
- Device connects but disconnects frequently
- Connection status fluctuates

**Solutions:**
1. Check network stability:
   ```bash
   ping -c 100 <device-ip>  # Monitor packet loss
   ```
2. Review connection retry settings
3. Check device power supply
4. Verify network cable connections
5. Review device logs for errors

---

### Cycle Issues

#### Cycle Won't Start

**Symptoms:**
- POST to `/api/melag/{id}/start` returns error
- Cycle status remains "IDLE"

**Diagnosis:**
1. Check device status:
   ```bash
   curl http://localhost:8080/api/melag/{id}/status
   ```
2. Verify device is connected
3. Check if device is already running a cycle
4. Review device diagnostics

**Solutions:**

**Device Not Connected:**
- Ensure device is online
- Check device connection status
- Review connection logs

**Device Busy:**
- Wait for current cycle to complete
- Check for stuck cycles
- Restart device if necessary

**Invalid Parameters:**
- Verify cycle parameters are valid
- Check temperature/pressure ranges
- Review program name

#### Cycle Status Not Updating

**Symptoms:**
- Cycle started but status doesn't update
- Progress percentage stuck

**Solutions:**
1. Check cycle polling is active:
   ```bash
   curl http://localhost:8080/api/cycles/running
   ```
2. Verify WebSocket connection (if using frontend)
3. Check polling interval configuration
4. Review device adapter logs
5. Restart cycle if necessary

#### Cycle Results Not Retrieved

**Symptoms:**
- Cycle completes but result is missing
- Error codes not populated

**Solutions:**
1. Check cycle completion handler logs
2. Verify device adapter result retrieval
3. Review database for cycle record
4. Check audit logs for completion events

---

### Database Issues

#### Database Locked

**Symptoms:**
- Error: "database is locked"
- Operations timeout

**Solutions:**
1. Check for multiple application instances:
   ```bash
   ps aux | grep steri-connect-go
   ```
2. Verify WAL files exist:
   ```bash
   ls -la data/steri-connect.db*
   ```
3. Check for long-running transactions
4. Restart application if necessary

#### Database Corruption

**Symptoms:**
- Errors reading from database
- Inconsistent data

**Diagnosis:**
```bash
sqlite3 data/steri-connect.db "PRAGMA integrity_check;"
```

**Solutions:**
1. Restore from backup
2. Export data before corruption:
   ```bash
   sqlite3 data/steri-connect.db .dump > backup.sql
   ```
3. Recreate database:
   ```bash
   rm data/steri-connect.db
   # Application will recreate on startup
   ```
4. Import data if possible

#### Database Size Growing Rapidly

**Symptoms:**
- Database file size increases quickly
- Disk space concerns

**Solutions:**
1. Review cycle data retention:
   ```bash
   sqlite3 data/steri-connect.db "SELECT COUNT(*) FROM cycles;"
   ```
2. Archive old cycles:
   ```sql
   DELETE FROM cycles WHERE start_ts < date('now', '-90 days');
   ```
3. Vacuum database:
   ```sql
   VACUUM;
   ```
4. Review audit log retention

---

### API Issues

#### 401 Unauthorized

**Symptoms:**
- API requests return 401 status
- "Unauthorized" error message

**Solutions:**
1. Check API key configuration:
   ```yaml
   auth:
     api_key_required: true
     api_key: "your-api-key"
   ```
2. Verify API key in request header:
   ```bash
   curl -H "X-API-Key: your-api-key" http://localhost:8080/api/devices
   ```
3. Check API key matches configuration
4. Verify authentication middleware is enabled

#### 404 Not Found

**Symptoms:**
- Endpoint returns 404
- Route not found

**Solutions:**
1. Verify endpoint path is correct
2. Check route registration in `router.go`
3. Verify API base path: `/api/...`
4. Check for typos in endpoint URL

#### 500 Internal Server Error

**Symptoms:**
- Generic server error
- Operation fails unexpectedly

**Solutions:**
1. Check application logs:
   ```bash
   tail -f logs/app.log
   ```
2. Review error details in log entries
3. Check database connectivity
4. Verify device manager is initialized
5. Review stack traces in logs

---

### WebSocket Issues

#### WebSocket Connection Fails

**Symptoms:**
- Cannot connect to WebSocket endpoint
- Connection closes immediately

**Solutions:**
1. Verify WebSocket URL: `ws://localhost:8080/ws`
2. Check CORS configuration if accessing from different origin
3. Verify WebSocket hub is initialized
4. Check firewall rules
5. Review WebSocket connection logs

#### WebSocket Events Not Received

**Symptoms:**
- Connected but no events received
- Events missing

**Solutions:**
1. Verify WebSocket connection is active
2. Check event broadcasting in code
3. Review WebSocket hub logs
4. Verify device manager is broadcasting events
5. Test with multiple clients

---

### Performance Issues

#### High CPU Usage

**Symptoms:**
- Application uses excessive CPU
- System becomes slow

**Solutions:**
1. Check polling intervals:
   ```yaml
   devices:
     melag:
       status_poll_interval: 2  # Increase if too frequent
   ```
2. Review number of active cycles
3. Check for tight polling loops
4. Monitor with metrics endpoint
5. Review device adapter implementations

#### High Memory Usage

**Symptoms:**
- Memory usage grows over time
- Application becomes unresponsive

**Solutions:**
1. Check log buffer size:
   ```go
   logging.InitLogBuffer(1000)  // Reduce if too large
   ```
2. Review cycle data retention
3. Check for memory leaks in adapters
4. Monitor with metrics endpoint
5. Restart application periodically if needed

#### Slow API Responses

**Symptoms:**
- API requests take long time
- Timeout errors

**Solutions:**
1. Check database query performance
2. Review indexes on database tables
3. Optimize database queries
4. Check network latency to devices
5. Review polling operations blocking API

---

### Log Issues

#### Logs Not Appearing

**Symptoms:**
- No log output
- Log file empty

**Solutions:**
1. Check log configuration:
   ```yaml
   logging:
     level: "INFO"  # Ensure not set to ERROR only
     output: "stdout"  # Or file path
   ```
2. Verify log file permissions
3. Check disk space
4. Verify logging is initialized in code

#### Log File Too Large

**Symptoms:**
- Log file grows very large
- Disk space issues

**Solutions:**
1. Enable log rotation:
   ```yaml
   logging:
     max_file_size_mb: 50
     max_backups: 10
     compress: true
   ```
2. Reduce log level in production:
   ```yaml
   logging:
     level: "WARN"  # Only warnings and errors
   ```
3. Archive old logs manually
4. Review log verbosity

---

## Diagnostic Tools

### Health Check

```bash
curl http://localhost:8080/api/health | jq
```

Check for:
- Database connectivity
- Device status summary
- Overall system status

### Metrics

```bash
curl http://localhost:8080/api/metrics | jq
```

Monitor:
- Uptime
- Request counts
- Cycle statistics
- Database size

### Device Diagnostics

```bash
curl http://localhost:8080/api/diagnostics/{deviceId} | jq
```

Review:
- Connection test results
- Protocol information
- Recent errors
- Connection history

### Database Inspection

```bash
sqlite3 data/steri-connect.db
.tables
.schema devices
SELECT * FROM devices;
```

---

## Getting Help

### Logs

Always check logs first:
- Application logs: `logs/app.log` or stdout
- System logs: `journalctl -u stericonnect` (Linux)
- Event Viewer (Windows)

### Diagnostic Information

Collect before reporting issues:
1. Health check output
2. Metrics output
3. Device diagnostics for affected devices
4. Relevant log entries
5. Configuration file (sanitized)

### Support Resources

- API Reference: `docs/API-Reference.md`
- Developer Guide: `docs/Developer-Guide.md`
- Deployment Guide: `docs/Deployment-Guide.md`

