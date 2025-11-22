# User Guide

**Version:** 1.0.0  
**Last Updated:** 2025-11-22

## Introduction

This guide helps you get started with Steri-Connect, the middleware service that connects your Steri-Suite frontend application with medical sterilization devices (Melag Cliniclave 45 and Getinge Aquadis 56).

## Quick Start

### 1. Download and Run

Download the `steri-connect-go.exe` executable and place it in your desired directory.

### 2. Configure

Edit `config/config.yaml` to match your environment:

```yaml
server:
  port: 8080
  bind_address: "127.0.0.1"  # localhost only (default)

database:
  path: "./data/steri-connect.db"

logging:
  level: "INFO"
  format: "json"
  output: "stdout"
```

### 3. Start the Service

```bash
./steri-connect-go.exe
```

The service will:
- Create the database automatically on first run
- Start the HTTP server on `http://localhost:8080`
- Initialize device connections

### 4. Access Test UI

Open your browser and navigate to:
```
http://localhost:8080/test-ui
```

## Adding Devices

### Via Test UI

1. Open the Test UI at `http://localhost:8080/test-ui`
2. Click "Add Device" button
3. Fill in device information:
   - **Name:** Descriptive name (e.g., "Melag Cliniclave 45 - Room 101")
   - **Manufacturer:** Select "Melag" or "Getinge"
   - **IP Address:** Device IP address on your network
   - **Type:** "Steri" for sterilization devices, "RDG" for cleaning devices
   - **Location:** Optional location identifier
4. Click "Save"

### Via API

```bash
curl -X POST http://localhost:8080/api/devices \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Melag Cliniclave 45",
    "manufacturer": "Melag",
    "ip": "192.168.1.100",
    "type": "Steri",
    "location": "Room 101"
  }'
```

## Starting Sterilization Cycles

### Via Test UI

1. Navigate to "Cycle Control" tab in Test UI
2. Select a Melag device from the dropdown
3. Enter cycle parameters:
   - **Program:** Program name (optional)
   - **Temperature:** Target temperature in Celsius (optional)
   - **Pressure:** Target pressure in bar (optional)
   - **Duration:** Duration in minutes (optional)
4. Click "Start Cycle"
5. Monitor progress in real-time

### Via API

```bash
curl -X POST http://localhost:8080/api/melag/1/start \
  -H "Content-Type: application/json" \
  -d '{
    "program": "Standard",
    "temperature": 134.0,
    "pressure": 2.1,
    "duration": 15
  }'
```

## Monitoring Cycles

### Real-Time Status

The Test UI displays real-time cycle status including:
- Current phase (e.g., "Aufheizen", "Sterilisation", "Trocknung")
- Temperature and pressure readings
- Progress percentage
- Time remaining

### WebSocket Events

Connect to `ws://localhost:8080/ws` to receive real-time events:
- `cycle_status_update` - Progress updates every 2 seconds
- `cycle_completed` - Cycle finished successfully
- `cycle_failed` - Cycle failed with error details

### Viewing Cycle History

1. Navigate to "Database" tab in Test UI
2. Select "cycles" table
3. Filter by device ID or date range
4. Export data as JSON or CSV if needed

## Viewing System Status

### Health Check

```bash
curl http://localhost:8080/api/health
```

Returns system health including:
- Database connection status
- Device connectivity summary
- Service uptime
- Memory usage

### Metrics

```bash
curl http://localhost:8080/api/metrics
```

Returns performance metrics:
- Total cycles processed
- API request counts
- Active device connections
- Database size

## Device Diagnostics

### Via Test UI

1. Navigate to "System Status" tab
2. View device connectivity summary
3. Check individual device status

### Via API

```bash
curl http://localhost:8080/api/diagnostics/1
```

Returns comprehensive diagnostic information:
- Connection test results
- Last successful communication
- Protocol-specific debugging info
- Recent error logs
- Connection history

## Exporting Data

### Cycle Protocols

#### PDF Export

```bash
curl http://localhost:8080/api/cycles/42/export/pdf -o cycle_42.pdf
```

#### CSV Export

```bash
curl http://localhost:8080/api/cycles/export/csv?device_id=1 -o cycles.csv
```

#### JSON Export

```bash
curl http://localhost:8080/api/cycles/export/json?device_id=1 -o cycles.json
```

### Via Test UI

1. Navigate to "Database" tab
2. Select table (e.g., "cycles")
3. Apply filters if needed
4. Click "Export JSON" or "Export CSV"

## Configuration

### Server Settings

Edit `config/config.yaml`:

```yaml
server:
  port: 8080                    # HTTP server port
  bind_address: "127.0.0.1"     # "127.0.0.1" for localhost only, "0.0.0.0" for network access
```

**Security Note:** If you set `bind_address: "0.0.0.0"` for network access, enable API key authentication:

```yaml
auth:
  api_key_required: true
  api_key: "your-secure-api-key-here"
```

### Database Settings

```yaml
database:
  path: "./data/steri-connect.db"  # Database file path
```

### Logging Settings

```yaml
logging:
  level: "INFO"              # DEBUG, INFO, WARN, ERROR
  format: "json"             # json or text
  output: "stdout"           # stdout or file path
  max_file_size_mb: 10      # Log rotation size
  max_backups: 10           # Number of backup files
  compress: true            # Compress old logs
```

### Device Polling Intervals

```yaml
devices:
  melag:
    status_poll_interval: 2    # Seconds between status polls
  getinge:
    ping_interval: 15          # Seconds between ping checks
    ping_timeout: 5             # Ping timeout in seconds
```

### Test UI Settings

```yaml
test_ui:
  enabled: true              # Enable/disable Test UI
  require_auth: false        # Optional authentication for Test UI
```

## Troubleshooting

### Device Not Connecting

1. Check device IP address is correct
2. Verify network connectivity: `ping <device-ip>`
3. Check device diagnostics: `GET /api/diagnostics/{deviceId}`
4. Review logs for connection errors

### Cycle Not Starting

1. Verify device is connected (check status endpoint)
2. Ensure device is not already running a cycle
3. Check device diagnostics for protocol errors
4. Review audit logs for error details

### Database Errors

1. Check database file permissions
2. Verify disk space is available
3. Review database path in configuration
4. Check logs for database errors

### Service Won't Start

1. Check if port 8080 is already in use
2. Verify configuration file syntax (YAML)
3. Check file permissions for database directory
4. Review startup logs for errors

## Best Practices

### Security

- Keep Test UI disabled in production (`test_ui.enabled: false`)
- Use API key authentication for network access
- Restrict bind address to localhost when possible
- Regularly review audit logs

### Performance

- Monitor system metrics regularly
- Adjust polling intervals based on network conditions
- Archive old cycle data periodically
- Monitor database size and disk space

### Maintenance

- Regularly backup database file
- Review and rotate log files
- Monitor device connection health
- Keep configuration files version-controlled

## Support

For additional help:
- Review API Reference: `docs/API-Reference.md`
- Check Troubleshooting Guide: `docs/Troubleshooting-Guide.md`
- Review system logs for detailed error information

