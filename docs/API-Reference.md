# API Reference

**Version:** 1.0.0  
**Base URL:** `http://localhost:8080/api`  
**Last Updated:** 2025-11-22

## Overview

The Steri-Connect API provides a RESTful interface for managing medical sterilization devices, monitoring cycles, and accessing system information. The API uses JSON for request and response bodies and follows standard HTTP status codes.

## Authentication

By default, the API runs on localhost and does not require authentication. For network access, API key authentication can be enabled via configuration.

### API Key Authentication (Optional)

When `api_key_required: true` is set in `config/config.yaml`:

- **Header:** `X-API-Key: your-api-key-here`
- **Status Code:** `401 Unauthorized` if missing or invalid

## Endpoints

### System Endpoints

#### Health Check

```http
GET /api/health
```

Returns system health status including database connectivity, device status, and uptime.

**Authentication:** Not required

**Response:**

```json
{
  "status": "ok",
  "timestamp": "2025-11-22T10:00:00Z",
  "version": "1.0.0",
  "uptime": "2h30m15s",
  "database": {
    "connected": true,
    "status": "connected"
  },
  "devices": {
    "total": 2,
    "online": 1,
    "offline": 1,
    "error": 0
  },
  "websocket": {
    "connections": 3
  },
  "memory": {
    "alloc_mb": 12.5,
    "total_alloc_mb": 45.2,
    "sys_mb": 128.0,
    "num_gc": 5
  }
}
```

**Status Codes:**
- `200 OK` - System is operational
- `503 Service Unavailable` - System degraded or error state

---

#### System Metrics

```http
GET /api/metrics
```

Returns system performance metrics including request counts, cycle statistics, and database size.

**Authentication:** Not required

**Response:**

```json
{
  "uptime_seconds": 9000,
  "active_device_connections": 1,
  "total_cycles": 42,
  "cycles_today": 5,
  "total_api_requests": 1234,
  "requests_per_minute": 12,
  "database_size_mb": 2.5,
  "timestamp": "2025-11-22T10:00:00Z"
}
```

---

#### Device Diagnostics

```http
GET /api/diagnostics/{deviceId}
```

Returns comprehensive diagnostic information for a specific device, including connection tests, protocol details, and error history.

**Authentication:** Not required

**Path Parameters:**
- `deviceId` (integer, required) - Device ID

**Response:**

```json
{
  "device_id": 1,
  "device_name": "Melag Cliniclave 45",
  "manufacturer": "Melag",
  "connection_test": {
    "success": true,
    "timestamp": "2025-11-22T10:00:00Z",
    "duration": "250ms"
  },
  "last_successful_communication": "2025-11-22T09:58:30Z",
  "protocol_info": {
    "type": "Melag",
    "melag_info": {
      "ftp_connected": true,
      "protocol_file_accessible": true
    }
  },
  "recent_errors": [],
  "connection_history": [
    {
      "timestamp": "2025-11-22T09:58:30Z",
      "success": true,
      "action": "device_added",
      "details": "{\"ip\":\"192.168.1.100\"}"
    }
  ],
  "timestamp": "2025-11-22T10:00:00Z"
}
```

**Status Codes:**
- `200 OK` - Diagnostics retrieved successfully
- `404 Not Found` - Device not found
- `400 Bad Request` - Invalid device ID

---

### Device Management

#### List All Devices

```http
GET /api/devices
```

Returns a list of all configured devices.

**Response:**

```json
[
  {
    "id": 1,
    "name": "Melag Cliniclave 45",
    "model": "Cliniclave 45",
    "manufacturer": "Melag",
    "ip": "192.168.1.100",
    "serial": "MC45-12345",
    "type": "Steri",
    "location": "Room 101",
    "connected": true,
    "created": "2025-11-20T08:00:00Z",
    "updated": "2025-11-22T09:58:30Z"
  }
]
```

---

#### Get Device by ID

```http
GET /api/devices/{id}
```

Returns detailed information for a specific device.

**Path Parameters:**
- `id` (integer, required) - Device ID

**Response:**

```json
{
  "id": 1,
  "name": "Melag Cliniclave 45",
  "model": "Cliniclave 45",
  "manufacturer": "Melag",
  "ip": "192.168.1.100",
  "serial": "MC45-12345",
  "type": "Steri",
  "location": "Room 101",
  "connected": true,
  "created": "2025-11-20T08:00:00Z",
  "updated": "2025-11-22T09:58:30Z"
}
```

**Status Codes:**
- `200 OK` - Device found
- `404 Not Found` - Device not found

---

#### Create Device

```http
POST /api/devices
```

Creates a new device configuration.

**Request Body:**

```json
{
  "name": "Melag Cliniclave 45",
  "model": "Cliniclave 45",
  "manufacturer": "Melag",
  "ip": "192.168.1.100",
  "serial": "MC45-12345",
  "type": "Steri",
  "location": "Room 101"
}
```

**Required Fields:**
- `name` (string) - Device name
- `manufacturer` (string) - "Melag" or "Getinge"
- `ip` (string) - IP address
- `type` (string) - "Steri" or "RDG"

**Response:**

```json
{
  "id": 1,
  "name": "Melag Cliniclave 45",
  "model": "Cliniclave 45",
  "manufacturer": "Melag",
  "ip": "192.168.1.100",
  "serial": "MC45-12345",
  "type": "Steri",
  "location": "Room 101",
  "connected": false,
  "created": "2025-11-22T10:00:00Z",
  "updated": "2025-11-22T10:00:00Z"
}
```

**Status Codes:**
- `201 Created` - Device created successfully
- `400 Bad Request` - Invalid request data
- `409 Conflict` - Device with same IP and manufacturer already exists

---

#### Update Device

```http
PUT /api/devices/{id}
```

Updates an existing device configuration. Only provided fields are updated.

**Path Parameters:**
- `id` (integer, required) - Device ID

**Request Body:**

```json
{
  "name": "Melag Cliniclave 45 - Updated",
  "location": "Room 102"
}
```

**Response:**

```json
{
  "id": 1,
  "name": "Melag Cliniclave 45 - Updated",
  "model": "Cliniclave 45",
  "manufacturer": "Melag",
  "ip": "192.168.1.100",
  "serial": "MC45-12345",
  "type": "Steri",
  "location": "Room 102",
  "connected": true,
  "created": "2025-11-20T08:00:00Z",
  "updated": "2025-11-22T10:05:00Z"
}
```

**Status Codes:**
- `200 OK` - Device updated successfully
- `404 Not Found` - Device not found
- `400 Bad Request` - Invalid request data

---

#### Delete Device

```http
DELETE /api/devices/{id}
```

Deletes a device configuration. Associated cycles and status records are also deleted (cascade).

**Path Parameters:**
- `id` (integer, required) - Device ID

**Response:**

```json
{
  "message": "Device deleted successfully"
}
```

**Status Codes:**
- `200 OK` - Device deleted successfully
- `404 Not Found` - Device not found

---

#### Get Device Status

```http
GET /api/devices/{id}/status
```

Returns current health and connection status for a device.

**Path Parameters:**
- `id` (integer, required) - Device ID

**Response:**

```json
{
  "device_id": 1,
  "manufacturer": "Melag",
  "ip": "192.168.1.100",
  "connected": true,
  "health_status": "healthy",
  "last_seen": "2025-11-22T09:58:30Z",
  "connection_type": "MELAnet"
}
```

**Status Codes:**
- `200 OK` - Status retrieved successfully
- `404 Not Found` - Device not found

---

### Melag Device Operations

#### Start Cycle

```http
POST /api/melag/{id}/start
```

Starts a sterilization cycle on a Melag device.

**Path Parameters:**
- `id` (integer, required) - Device ID

**Request Body:**

```json
{
  "program": "Standard",
  "temperature": 134.0,
  "pressure": 2.1,
  "duration": 15
}
```

**Fields:**
- `program` (string, optional) - Program name
- `temperature` (float, optional) - Target temperature in Celsius
- `pressure` (float, optional) - Target pressure in bar
- `duration` (integer, optional) - Duration in minutes

**Response:**

```json
{
  "cycle_id": 42,
  "device_id": 1,
  "program": "Standard",
  "start_ts": "2025-11-22T10:00:00Z",
  "status": "started"
}
```

**Status Codes:**
- `201 Created` - Cycle started successfully
- `400 Bad Request` - Invalid request or device not ready
- `404 Not Found` - Device not found
- `500 Internal Server Error` - Failed to start cycle

---

#### Get Melag Status

```http
GET /api/melag/{id}/status
```

Returns current device and cycle status for a Melag device.

**Path Parameters:**
- `id` (integer, required) - Device ID

**Response:**

```json
{
  "device_id": 1,
  "device_name": "Melag Cliniclave 45",
  "connected": true,
  "current_cycle": {
    "cycle_id": 42,
    "program": "Standard",
    "phase": "Sterilisation",
    "progress_percent": 65,
    "temperature": 134.5,
    "pressure": 2.1,
    "time_remaining": "5m30s"
  },
  "last_cycle": {
    "cycle_id": 41,
    "program": "Standard",
    "result": "OK",
    "end_ts": "2025-11-22T09:45:00Z"
  }
}
```

**Status Codes:**
- `200 OK` - Status retrieved successfully
- `404 Not Found` - Device not found

---

#### Get Cycle Details

```http
GET /api/melag/{id}/cycles/{cycle_id}
```

Returns detailed information for a specific cycle.

**Path Parameters:**
- `id` (integer, required) - Device ID
- `cycle_id` (integer, required) - Cycle ID

**Response:**

```json
{
  "id": 42,
  "device_id": 1,
  "program": "Standard",
  "start_ts": "2025-11-22T10:00:00Z",
  "end_ts": "2025-11-22T10:15:00Z",
  "phase": "COMPLETED",
  "progress_percent": 100,
  "temperature": 134.5,
  "pressure": 2.1,
  "result": "OK",
  "error_code": null,
  "error_description": null
}
```

**Status Codes:**
- `200 OK` - Cycle found
- `404 Not Found` - Cycle not found
- `400 Bad Request` - Cycle does not belong to device

---

### Cycle Management

#### List All Cycles

```http
GET /api/cycles
```

Returns a paginated list of all cycles with optional filtering and sorting.

**Query Parameters:**
- `limit` (integer, optional) - Number of results per page (default: 20)
- `offset` (integer, optional) - Number of results to skip (default: 0)
- `sort_by` (string, optional) - Sort field: "start_ts", "device_id", "result" (default: "start_ts")
- `sort_order` (string, optional) - Sort order: "asc" or "desc" (default: "desc")
- `device_id` (integer, optional) - Filter by device ID
- `start_date` (string, optional) - Filter by start date (RFC3339 or YYYY-MM-DD)
- `end_date` (string, optional) - Filter by end date (RFC3339 or YYYY-MM-DD)
- `result` (string, optional) - Filter by result: "OK" or "NOK"

**Example:**

```http
GET /api/cycles?limit=10&offset=0&device_id=1&result=OK&sort_by=start_ts&sort_order=desc
```

**Response:**

```json
{
  "cycles": [
    {
      "id": 42,
      "device_id": 1,
      "program": "Standard",
      "start_ts": "2025-11-22T10:00:00Z",
      "end_ts": "2025-11-22T10:15:00Z",
      "phase": "COMPLETED",
      "result": "OK",
      "device_name": "Melag Cliniclave 45",
      "device_manufacturer": "Melag",
      "device_ip": "192.168.1.100"
    }
  ],
  "total_count": 42,
  "limit": 10,
  "offset": 0
}
```

---

#### Get Cycle by ID

```http
GET /api/cycles/{id}
```

Returns detailed information for a specific cycle.

**Path Parameters:**
- `id` (integer, required) - Cycle ID

**Response:**

```json
{
  "id": 42,
  "device_id": 1,
  "program": "Standard",
  "start_ts": "2025-11-22T10:00:00Z",
  "end_ts": "2025-11-22T10:15:00Z",
  "phase": "COMPLETED",
  "progress_percent": 100,
  "temperature": 134.5,
  "pressure": 2.1,
  "result": "OK",
  "error_code": null,
  "error_description": null
}
```

**Status Codes:**
- `200 OK` - Cycle found
- `404 Not Found` - Cycle not found

---

#### Get Running Cycles

```http
GET /api/cycles/running
```

Returns all currently running cycles.

**Response:**

```json
[
  {
    "id": 42,
    "device_id": 1,
    "program": "Standard",
    "start_ts": "2025-11-22T10:00:00Z",
    "phase": "Sterilisation",
    "progress_percent": 65,
    "temperature": 134.5,
    "pressure": 2.1,
    "device_name": "Melag Cliniclave 45"
  }
]
```

---

#### Export Cycle as PDF

```http
GET /api/cycles/{id}/export/pdf
```

Exports a cycle protocol as a PDF document.

**Path Parameters:**
- `id` (integer, required) - Cycle ID

**Response:**
- Content-Type: `application/pdf`
- File download with cycle protocol

**Status Codes:**
- `200 OK` - PDF generated successfully
- `404 Not Found` - Cycle not found

---

#### Export Cycles as CSV

```http
GET /api/cycles/export/csv
```

Exports cycles as CSV data.

**Query Parameters:**
- `device_id` (integer, optional) - Filter by device ID
- `start_date` (string, optional) - Filter by start date
- `end_date` (string, optional) - Filter by end date
- `result` (string, optional) - Filter by result

**Response:**
- Content-Type: `text/csv`
- CSV file download

---

#### Export Cycles as JSON

```http
GET /api/cycles/export/json
```

Exports cycles as JSON data.

**Query Parameters:**
- `device_id` (integer, optional) - Filter by device ID
- `start_date` (string, optional) - Filter by start date
- `end_date` (string, optional) - Filter by end date
- `result` (string, optional) - Filter by result

**Response:**

```json
[
  {
    "id": 42,
    "device_id": 1,
    "program": "Standard",
    "start_ts": "2025-11-22T10:00:00Z",
    "end_ts": "2025-11-22T10:15:00Z",
    "result": "OK"
  }
]
```

---

## WebSocket Events

Connect to `ws://localhost:8080/ws` for real-time events.

### Event Types

#### Device Status Change

```json
{
  "event": "device_status_change",
  "timestamp": "2025-11-22T10:00:00Z",
  "data": {
    "device_id": 1,
    "connected": true,
    "status": "connected"
  }
}
```

#### Cycle Started

```json
{
  "event": "cycle_started",
  "timestamp": "2025-11-22T10:00:00Z",
  "data": {
    "cycle_id": 42,
    "device_id": 1,
    "program": "Standard"
  }
}
```

#### Cycle Status Update

```json
{
  "event": "cycle_status_update",
  "timestamp": "2025-11-22T10:05:00Z",
  "data": {
    "cycle_id": 42,
    "device_id": 1,
    "phase": "Sterilisation",
    "progress_percent": 65,
    "temperature": 134.5,
    "pressure": 2.1,
    "time_remaining": "5m30s"
  }
}
```

#### Cycle Completed

```json
{
  "event": "cycle_completed",
  "timestamp": "2025-11-22T10:15:00Z",
  "data": {
    "cycle_id": 42,
    "device_id": 1,
    "result": "OK"
  }
}
```

#### Cycle Failed

```json
{
  "event": "cycle_failed",
  "timestamp": "2025-11-22T10:15:00Z",
  "data": {
    "cycle_id": 42,
    "device_id": 1,
    "error_code": "E001",
    "error_description": "Temperature too low"
  }
}
```

---

## Error Responses

All error responses follow this format:

```json
{
  "error": "error_code",
  "message": "Human-readable error message"
}
```

### Common Error Codes

- `method_not_allowed` - HTTP method not supported for this endpoint
- `invalid_device_id` - Invalid or missing device ID
- `device_not_found` - Device does not exist
- `invalid_request` - Request body validation failed
- `internal_error` - Server-side error occurred
- `unauthorized` - Authentication required or failed

### HTTP Status Codes

- `200 OK` - Request successful
- `201 Created` - Resource created successfully
- `400 Bad Request` - Invalid request data
- `401 Unauthorized` - Authentication required
- `404 Not Found` - Resource not found
- `409 Conflict` - Resource conflict (e.g., duplicate device)
- `500 Internal Server Error` - Server error
- `503 Service Unavailable` - Service degraded

---

## Rate Limiting

Currently, no rate limiting is enforced. For production deployments, consider implementing rate limiting based on your requirements.

---

## Versioning

The API version is included in the health check response. Future API versions may be introduced with versioned endpoints (e.g., `/api/v2/...`).

---

## Support

For API support and questions, refer to the project documentation or contact the development team.

