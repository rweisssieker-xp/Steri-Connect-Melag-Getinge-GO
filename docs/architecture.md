# Architecture

## Executive Summary

**Project:** Steri-Connect-Melag-Getinge-GO  
**Architecture Type:** Middleware/Backend Service  
**Language:** Go (golang)  
**Deployment:** Portable Executable (Windows/Linux)  
**Date:** 2025-11-21T09:33:54.802Z

This architecture document defines the system design for the GO-App middleware that bridges communication between the existing Steri-Suite frontend and medical sterilization devices (Melag Cliniclave 45 and Getinge Aquadis 56). The architecture emphasizes device abstraction, local-first data storage, and portable deployment.

**Key Architectural Principles:**
- Device Abstraction Layer: Unified interface for different device types
- Local-First: SQLite for all data storage, no external dependencies
- Portable: Single executable, no installation required
- Protocol Encapsulation: Manufacturer-specific protocols hidden from frontend
- Extensibility: Easy addition of future device types

---

## Decision Summary

| Category | Decision | Version | Affects Epics | Rationale |
| -------- | -------- | ------- | ------------- | --------- |
| Programming Language | Go (golang) | Latest stable | All | Portable, concurrent, suitable for device communication |
| Data Storage | SQLite | Latest | All data epics | Local-first, no server required, portable |
| API Pattern | REST + WebSocket | - | API epics | Standard HTTP for CRUD, WebSocket for real-time |
| Device Integration | Device Abstraction Layer | - | Device adapter epics | Enables multiple device types, protocol encapsulation |
| Melag Integration | MELAnet Box (FTP) | - | Melag epic | Documented integration path, manufacturer support |
| Getinge Integration | ICMP Ping (Phase 1) | - | Getinge epic | Only available option, full API pending approval |
| Authentication | API Key (optional) | - | Security epic | Simple, sufficient for local deployment |
| Test Interface | Web UI (localhost) | - | Test UI epic | Development/debugging without Steri-Suite |
| Logging | Structured Logging | - | All epics | Audit trails, debugging, compliance |
| Error Handling | Graceful Degradation | - | All epics | Network interruptions, device failures |

---

## Project Structure

```
steri-connect-go/
├── cmd/
│   └── server/
│       └── main.go                 # Application entry point
├── internal/
│   ├── api/
│   │   ├── handlers/              # HTTP request handlers
│   │   │   ├── devices.go
│   │   │   ├── cycles.go
│   │   │   ├── melag.go
│   │   │   ├── getinge.go
│   │   │   ├── health.go
│   │   │   └── testui.go
│   │   ├── middleware/            # HTTP middleware
│   │   │   ├── auth.go
│   │   │   └── logging.go
│   │   └── websocket/             # WebSocket handlers
│   │       └── events.go
│   ├── adapters/
│   │   ├── device/                # Device abstraction interface
│   │   │   └── interface.go
│   │   ├── melag/                 # Melag device adapter
│   │   │   ├── adapter.go
│   │   │   ├── ftp_client.go
│   │   │   └── protocol_parser.go
│   │   └── getinge/               # Getinge device adapter
│   │       ├── adapter.go
│   │       └── icmp_monitor.go
│   ├── database/
│   │   ├── sqlite.go              # SQLite connection and setup
│   │   ├── migrations/            # Database migrations
│   │   └── models.go               # Data models
│   ├── config/
│   │   └── config.go               # Configuration management
│   ├── logging/
│   │   └── logger.go               # Structured logging
│   └── testui/
│       ├── handlers.go             # Test UI HTTP handlers
│       └── templates/              # HTML templates
├── pkg/
│   └── utils/                     # Shared utilities
├── web/
│   └── testui/                    # Test UI static assets
│       ├── index.html
│       ├── css/
│       └── js/
├── config/
│   └── config.yaml                 # Configuration file
├── go.mod
├── go.sum
└── README.md
```

---

## Epic to Architecture Mapping

| Epic | Architecture Components | Key Decisions |
| ---- | ----------------------- | ------------- |
| Foundation Setup | Project structure, Go modules, SQLite setup | Go project layout, database initialization |
| Device Management | Device abstraction layer, SQLite models | Device interface design, data model |
| Melag Integration | Melag adapter, FTP client, protocol parser | MELAnet Box integration, FTP handling |
| Getinge Monitoring | Getinge adapter, ICMP monitor | ICMP ping implementation, status tracking |
| REST API | HTTP handlers, middleware, routing | REST endpoints, request/response handling |
| WebSocket Events | WebSocket handler, event broadcasting | Real-time event distribution |
| Test UI | Test UI handlers, HTML templates | Simple web interface for testing |
| Health & Diagnostics | Health endpoints, metrics collection | System monitoring, diagnostics |

---

## Technology Stack Details

### Core Technologies

**Programming Language:**
- **Go (golang)** - Latest stable version
  - Rationale: Excellent concurrency support for device communication, portable binaries, strong standard library
  - Key packages:
    - `net/http` - HTTP server and client
    - `database/sql` - Database interface
    - `github.com/mattn/go-sqlite3` - SQLite driver
    - `github.com/gorilla/websocket` - WebSocket support

**Database:**
- **SQLite** - Latest version
  - Rationale: Local-first storage, no server required, portable, sufficient for single-instance deployment
  - Schema: Devices, cycles, audit logs, Getinge status

**HTTP Framework:**
- **Standard library `net/http`** or **Gin** (if needed)
  - Rationale: Go standard library sufficient for REST API, lightweight
  - Alternative: Gin framework if more features needed

**WebSocket:**
- **Gorilla WebSocket** (`github.com/gorilla/websocket`)
  - Rationale: Standard WebSocket library for Go, well-maintained

**Configuration:**
- **YAML** configuration file (`gopkg.in/yaml.v3`)
  - Rationale: Human-readable, supports nested configuration

**Logging:**
- **Structured logging** (`log/slog` or `github.com/sirupsen/logrus`)
  - Rationale: Structured logs for audit trails, JSON format for parsing

### Integration Points

**Melag Integration:**
- **Protocol:** FTP (via MELAnet Box)
- **Library:** Standard `net/ftp` or `github.com/jlaffaye/ftp`
- **Data Format:** Protocol files (format TBD from manufacturer docs)

**Getinge Integration:**
- **Protocol:** ICMP Ping
- **Library:** Standard `net` package (`net.IP`, `net.Dial`)
- **Data:** Online/offline status only

**Steri-Suite Integration:**
- **Protocol:** HTTP REST + WebSocket
- **Interface:** RESTful API endpoints, WebSocket events
- **Authentication:** API Key (optional for localhost)

---

## Implementation Patterns

### Device Abstraction Pattern

**Pattern:** Adapter Pattern with Strategy Pattern

**Interface Definition:**
```go
type Device interface {
    Connect() error
    Disconnect() error
    StartCycle(params CycleParams) error
    GetStatus() (Status, error)
    GetLastCycle() (Cycle, error)
    IsConnected() bool
}
```

**Adapter Implementation:**
- Each device type (Melag, Getinge) implements the Device interface
- Protocol-specific logic encapsulated in adapters
- Unified interface exposed to API layer

**Benefits:**
- Easy to add new device types
- Protocol changes isolated to adapters
- Consistent API regardless of device type

### State Management Pattern

**Pattern:** State Machine for Device and Cycle States

**Device States:**
- `DISCONNECTED` - Not connected
- `CONNECTING` - Connection in progress
- `CONNECTED` - Connected and ready
- `ERROR` - Connection error

**Cycle States:**
- `READY` - Device ready for cycle
- `STARTING` - Cycle start initiated
- `RUNNING` - Cycle in progress
- `COMPLETED` - Cycle finished successfully
- `FAILED` - Cycle failed
- `CANCELLED` - Cycle cancelled

**Implementation:**
- State transitions validated
- State persisted in database
- State changes broadcast via WebSocket

### Error Handling Pattern

**Pattern:** Graceful Degradation with Retry Logic

**Principles:**
- Network interruptions: Retry with exponential backoff
- Device failures: Log error, notify via WebSocket, continue operation
- Database errors: Log, return error response, don't crash
- Protocol errors: Log, return structured error, allow recovery

**Error Types:**
- `DeviceUnavailableError` - Device not reachable
- `ProtocolError` - Communication protocol error
- `DatabaseError` - Database operation failed
- `ValidationError` - Input validation failed

### Concurrency Pattern

**Pattern:** Goroutines for Concurrent Device Communication

**Implementation:**
- Each device connection managed by goroutine
- Status polling in separate goroutines
- WebSocket connections handled concurrently
- Channel-based communication between components

**Benefits:**
- Non-blocking device communication
- Real-time status updates
- Efficient resource usage

---

## Consistency Rules

### Naming Conventions

**Packages:**
- Lowercase, single word when possible
- `internal/` for private code
- `pkg/` for reusable packages

**Files:**
- Snake_case for multi-word files: `device_manager.go`
- Match package name when single file: `logger.go` in `logging` package

**Types:**
- PascalCase: `Device`, `CycleStatus`, `MelagAdapter`
- Interfaces end with `-er` when appropriate: `DeviceConnector`

**Functions:**
- PascalCase for exported: `StartCycle()`
- camelCase for unexported: `parseProtocolFile()`

**Variables:**
- camelCase: `deviceID`, `cycleStatus`
- Constants: PascalCase or UPPER_CASE: `DefaultPort`, `MAX_RETRIES`

### Code Organization

**Layered Architecture:**
1. **API Layer** (`internal/api/`) - HTTP handlers, WebSocket
2. **Service Layer** (`internal/services/`) - Business logic (if needed)
3. **Adapter Layer** (`internal/adapters/`) - Device-specific implementations
4. **Data Layer** (`internal/database/`) - Database operations

**Dependency Direction:**
- API → Adapters → Database
- No circular dependencies
- Adapters independent of each other

### Error Handling

**Error Return Pattern:**
```go
result, err := operation()
if err != nil {
    log.Error("operation failed", "error", err)
    return nil, fmt.Errorf("operation failed: %w", err)
}
```

**Error Wrapping:**
- Use `fmt.Errorf` with `%w` for error wrapping
- Preserve original error context
- Add domain-specific context

**Error Logging:**
- Log errors with context (device ID, cycle ID, etc.)
- Use structured logging fields
- Different log levels: DEBUG, INFO, WARN, ERROR

### Logging Strategy

**Structured Logging:**
- JSON format for production
- Include context fields: device_id, cycle_id, user_id
- Log levels: DEBUG, INFO, WARN, ERROR

**Audit Logging:**
- Separate audit log table in SQLite
- Immutable audit records (Write-Ahead-Log)
- Hash-based integrity verification
- All device operations logged

**Log Rotation:**
- File-based logging with rotation
- Max file size: 10MB
- Keep last 10 files
- Compress old logs

---

## Data Architecture

### Database Schema

**devices Table:**
```sql
CREATE TABLE devices (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    model TEXT,
    manufacturer TEXT NOT NULL,  -- 'Melag' or 'Getinge'
    ip TEXT NOT NULL,
    serial TEXT,
    type TEXT NOT NULL,  -- 'Steri' or 'RDG'
    location TEXT,
    created DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(ip, manufacturer)
);
```

**cycles Table:**
```sql
CREATE TABLE cycles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    device_id INTEGER NOT NULL,
    program TEXT,
    start_ts DATETIME NOT NULL,
    end_ts DATETIME,
    result TEXT,  -- 'OK' or 'NOK'
    error_code TEXT,
    error_description TEXT,
    phase TEXT,
    temperature REAL,
    pressure REAL,
    progress_percent INTEGER,
    FOREIGN KEY (device_id) REFERENCES devices(id)
);
CREATE INDEX idx_cycles_device_id ON cycles(device_id);
CREATE INDEX idx_cycles_start_ts ON cycles(start_ts);
```

**rdg_status Table (Getinge):**
```sql
CREATE TABLE rdg_status (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    device_id INTEGER NOT NULL,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    reachable INTEGER NOT NULL,  -- 0 or 1
    FOREIGN KEY (device_id) REFERENCES devices(id)
);
CREATE INDEX idx_rdg_status_device_id ON rdg_status(device_id);
CREATE INDEX idx_rdg_status_timestamp ON rdg_status(timestamp);
```

**audit_log Table:**
```sql
CREATE TABLE audit_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    action TEXT NOT NULL,  -- 'device_added', 'cycle_started', etc.
    entity_type TEXT,  -- 'device', 'cycle'
    entity_id INTEGER,
    user TEXT,
    details TEXT,  -- JSON details
    hash TEXT  -- Integrity hash
);
CREATE INDEX idx_audit_log_timestamp ON audit_log(timestamp);
CREATE INDEX idx_audit_log_entity ON audit_log(entity_type, entity_id);
```

### Data Relationships

```
devices (1) ──< (many) cycles
devices (1) ──< (many) rdg_status
cycles (many) ──> (1) devices
rdg_status (many) ──> (1) devices
audit_log (independent, references entities)
```

### Data Access Patterns

**Device Operations:**
- CRUD operations via repository pattern
- Connection state cached in memory, persisted on changes
- Status polling updates database periodically

**Cycle Operations:**
- Create cycle on start
- Update cycle status during execution
- Finalize cycle on completion
- Read-only access for history

**Audit Log:**
- Append-only (immutable)
- Hash verification for integrity
- Query by entity type/ID for traceability

---

## API Contracts

### REST API Endpoints

**Base URL:** `http://localhost:8080` (configurable)

**Device Management:**
- `GET /api/devices` - List all devices
- `POST /api/devices` - Add new device
- `GET /api/devices/{id}` - Get device details
- `PUT /api/devices/{id}` - Update device
- `DELETE /api/devices/{id}` - Delete device

**Melag Operations:**
- `POST /api/melag/{id}/start` - Start cycle
- `GET /api/melag/{id}/status` - Get current status
- `GET /api/melag/{id}/last-cycle` - Get last cycle result

**Getinge Operations:**
- `GET /api/getinge/{id}/ping` - Check reachability

**Cycles:**
- `GET /api/cycles` - List all cycles
- `GET /api/cycles/{id}` - Get cycle details

**Health & Diagnostics:**
- `GET /api/health` - System health check
- `GET /api/metrics` - System metrics
- `GET /api/diagnostics/{deviceId}` - Device diagnostics

**Test UI:**
- `GET /test-ui` - Test interface (HTML)
- `GET /test-ui/api/test` - API testing endpoint
- `GET /test-ui/logs` - Log viewing
- `GET /test-ui/db` - Database inspection

### WebSocket API

**Connection:** `ws://localhost:8080/ws`

**Event Types:**

**cycle_status_update:**
```json
{
  "event": "cycle_status_update",
  "device_id": 1,
  "cycle_id": 123,
  "status": "running",
  "phase": "Sterilisation",
  "progress_percent": 45,
  "temperature": 134.5,
  "pressure": 2.1,
  "time_remaining": 300,
  "timestamp": "2025-11-21T09:30:00Z"
}
```

**device_status_change:**
```json
{
  "event": "device_status_change",
  "device_id": 2,
  "status": "offline",
  "timestamp": "2025-11-21T09:30:00Z"
}
```

**error_event:**
```json
{
  "event": "error",
  "device_id": 1,
  "error_code": "DEVICE_UNAVAILABLE",
  "message": "Device not reachable",
  "timestamp": "2025-11-21T09:30:00Z"
}
```

### Request/Response Examples

See PRD Section 7.2 for detailed request/response schemas.

---

## Security Architecture

### Authentication

**API Key Authentication:**
- API key in HTTP header: `X-API-Key: {key}`
- API key stored in configuration file
- Optional for localhost, required for network access
- Configuration: `auth.required = false` (localhost) or `true` (network)

**Future: Token-based Authentication (Growth):**
- JWT tokens for stateless authentication
- Token expiration and renewal
- Role-based access control

### Network Security

**Localhost-First:**
- Default: Only localhost access
- Network access configurable
- Firewall-friendly (no inbound ports by default)

**Encryption:**
- HTTPS/TLS for network access (if enabled)
- Device communication encryption (if supported by device)
- SQLite database encryption (optional, via SQLCipher)

### Audit Trail Security

**Immutable Logs:**
- Write-Ahead-Log (WAL) mode for SQLite
- Hash-based integrity verification
- Append-only audit log table
- No deletion or modification of audit records

**Data Integrity:**
- Hash calculation for audit entries
- Periodic integrity checks
- Backup and recovery procedures

---

## Performance Considerations

### Response Time Targets

- API response time: ≤ 500ms (p95)
- Status updates: ≤ 2 seconds (Melag)
- WebSocket event latency: ≤ 100ms
- Test UI page load: < 1 second

### Optimization Strategies

**Database:**
- Indexes on frequently queried columns (device_id, timestamp)
- Connection pooling (SQLite connection reuse)
- Prepared statements for repeated queries

**Device Communication:**
- Concurrent device status polling (goroutines)
- Connection pooling for device connections
- Caching of device status (in-memory cache)

**API:**
- Efficient JSON serialization
- Pagination for large result sets
- Response compression (gzip) for large payloads

**Memory:**
- Bounded buffers for WebSocket events
- Connection limits to prevent resource exhaustion
- Garbage collection tuning if needed

---

## Deployment Architecture

### Portable Executable

**Build:**
- Single binary executable
- Static linking (no external dependencies)
- Cross-compilation for Windows/Linux

**Distribution:**
- Executable file + configuration file
- SQLite database created on first run
- No installation required

**Configuration:**
- YAML configuration file (`config.yaml`)
- Environment variables override (optional)
- Default values for localhost deployment

**File Structure:**
```
steri-connect-go/
├── steri-connect-go.exe (or steri-connect-go)
├── config.yaml
└── data/
    └── steri-connect.db (created automatically)
```

### Runtime Requirements

**Minimum:**
- Windows 10+ or Linux (recent distribution)
- Network access to devices (local network)
- No admin rights required

**Optional:**
- MELAnet Box hardware (for Melag integration)
- Network access for remote Steri-Suite

---

## Development Environment

### Prerequisites

- Go 1.21+ installed
- Git for version control
- SQLite3 (for database inspection/testing)

### Setup Commands

```bash
# Clone repository (when created)
git clone <repository-url>
cd steri-connect-go

# Initialize Go module
go mod init steri-connect-go

# Install dependencies
go get github.com/mattn/go-sqlite3
go get github.com/gorilla/websocket
go get gopkg.in/yaml.v3

# Run development server
go run cmd/server/main.go

# Build executable
go build -o steri-connect-go cmd/server/main.go

# Cross-compile for Windows (from Linux)
GOOS=windows GOARCH=amd64 go build -o steri-connect-go.exe cmd/server/main.go

# Cross-compile for Linux (from Windows)
GOOS=linux GOARCH=amd64 go build -o steri-connect-go cmd/server/main.go
```

### Development Workflow

1. **Local Development:**
   - Run `go run cmd/server/main.go`
   - Access Test UI at `http://localhost:8080/test-ui`
   - SQLite database created in `./data/steri-connect.db`

2. **Testing:**
   - Unit tests: `go test ./...`
   - Integration tests: Test against mock devices
   - Manual testing via Test UI

3. **Building:**
   - Development build: `go build`
   - Release build: `go build -ldflags="-s -w"` (strip debug info)

---

## Architecture Decision Records (ADRs)

### ADR-001: Go Language Choice

**Status:** Accepted

**Context:**
Middleware service requiring concurrent device communication, portable deployment, and local data storage.

**Decision:**
Use Go (golang) as the programming language.

**Rationale:**
- Excellent concurrency support (goroutines) for device communication
- Single binary deployment (portable executable)
- Strong standard library (HTTP, database, networking)
- Good performance for I/O-bound operations
- Growing ecosystem for medical device integration

**Consequences:**
- Positive: Portable, concurrent, performant
- Negative: Team needs Go expertise
- Neutral: Different from Steri-Suite tech stack (acceptable for middleware)

---

### ADR-002: SQLite for Data Storage

**Status:** Accepted

**Context:**
Local-first architecture, single-instance deployment, no external database server required.

**Decision:**
Use SQLite for all data storage.

**Rationale:**
- No server required (portable)
- Sufficient for single-instance deployment
- ACID compliance for audit trails
- Simple backup (single file)
- Good performance for expected data volume

**Consequences:**
- Positive: Simple deployment, no dependencies
- Negative: Limited scalability (acceptable for single-instance)
- Neutral: File-based (backup strategy needed)

---

### ADR-003: Device Abstraction Layer

**Status:** Accepted

**Context:**
Multiple device types (Melag, Getinge) with different protocols, future device additions expected.

**Decision:**
Implement Device Abstraction Layer pattern with protocol adapters.

**Rationale:**
- Unified interface for different device types
- Protocol changes isolated to adapters
- Easy to add new device types
- Consistent API regardless of device

**Consequences:**
- Positive: Extensible, maintainable, consistent API
- Negative: Additional abstraction layer (acceptable complexity)
- Neutral: Adapter pattern well-understood

---

### ADR-004: MELAnet Box Integration (Melag)

**Status:** Accepted

**Context:**
Melag devices require integration, MELAnet Box provides documented network integration path.

**Decision:**
Use MELAnet Box for Melag device integration via FTP protocol.

**Rationale:**
- Documented integration path
- Manufacturer support available
- Network-based (flexible deployment)
- Lower risk than direct device communication

**Consequences:**
- Positive: Documented approach, manufacturer support
- Negative: Additional hardware component (MELAnet Box)
- Neutral: FTP-based (may have latency, acceptable)

---

### ADR-005: ICMP Monitoring (Getinge Phase 1)

**Status:** Accepted

**Context:**
Getinge devices have no public API, full integration requires manufacturer approval.

**Decision:**
Implement ICMP ping monitoring for Getinge devices in Phase 1, pursue full integration in Phase 2.

**Rationale:**
- Only available option without manufacturer approval
- Provides immediate value (online/offline status)
- Foundation for future full integration
- Aligns with project roadmap

**Consequences:**
- Positive: Immediate implementation possible, no approval delays
- Negative: Limited functionality (online/offline only)
- Neutral: Foundation for future expansion

---

### ADR-006: REST + WebSocket API

**Status:** Accepted

**Context:**
Steri-Suite requires HTTP API for operations and real-time status updates.

**Decision:**
Provide REST API for CRUD operations and WebSocket for real-time events.

**Rationale:**
- REST standard for CRUD operations
- WebSocket efficient for real-time updates
- Both well-supported in Go standard library
- Familiar pattern for frontend integration

**Consequences:**
- Positive: Standard patterns, efficient real-time updates
- Negative: Two protocols to maintain (acceptable)
- Neutral: WebSocket connection management needed

---

### ADR-007: Test UI for Development

**Status:** Accepted

**Context:**
Development and debugging require testing without full Steri-Suite application.

**Decision:**
Provide simple web-based test interface accessible via localhost.

**Rationale:**
- Enables development without Steri-Suite
- Simplifies debugging and testing
- Provides database inspection capabilities
- Deactivatable for production

**Consequences:**
- Positive: Development efficiency, debugging capabilities
- Negative: Additional code to maintain (acceptable)
- Neutral: Simple HTML interface sufficient

---

## References

- **PRD:** `docs/PRD-Steri-Connect-Melag-Getinge-GO.md`
- **Research:** `docs/research-technical-device-interfaces-2025-11-21.md`
- **Product Brief:** `docs/product-brief-Steri-Connect-Melag-Getinge-GO-2025-11-21.md`
- **Go Documentation:** https://go.dev/doc/
- **SQLite Documentation:** https://www.sqlite.org/docs.html
- **Gorilla WebSocket:** https://github.com/gorilla/websocket

---

**Document Status:** Complete architecture document ready for implementation  
**Next Steps:** Create epics.md with detailed epic and story breakdown based on this architecture

