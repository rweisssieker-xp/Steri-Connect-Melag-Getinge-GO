# Developer Guide

**Version:** 1.0.0  
**Last Updated:** 2025-11-22

## Overview

This guide helps developers understand the codebase structure, development workflow, and how to contribute to the Steri-Connect project.

## Prerequisites

- **Go 1.21+** - [Download Go](https://go.dev/dl/)
- **Git** - Version control
- **Code Editor** - VS Code, GoLand, or similar
- **SQLite Tools** (optional) - For database inspection

## Project Structure

```
steri-connect-go/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── api/
│   │   ├── handlers/            # HTTP request handlers
│   │   │   ├── devices.go      # Device management endpoints
│   │   │   ├── melag.go        # Melag-specific endpoints
│   │   │   ├── cycles.go       # Cycle management endpoints
│   │   │   ├── health.go       # Health check endpoint
│   │   │   ├── metrics.go      # Metrics endpoint
│   │   │   ├── diagnostics.go  # Diagnostics endpoint
│   │   │   ├── database.go     # Test UI database endpoints
│   │   │   └── logs.go         # Test UI log endpoints
│   │   ├── middleware/
│   │   │   ├── auth.go         # API key authentication
│   │   │   ├── cors.go         # CORS handling
│   │   │   └── metrics.go      # Request metrics tracking
│   │   ├── websocket/
│   │   │   └── events.go       # WebSocket hub and events
│   │   ├── router.go           # Route configuration
│   │   └── server.go           # HTTP server setup
│   ├── adapters/
│   │   ├── device.go           # Device adapter interface
│   │   ├── melag/
│   │   │   └── melag.go        # Melag device adapter
│   │   └── getinge/
│   │       └── getinge.go      # Getinge device adapter
│   ├── database/
│   │   ├── sqlite.go           # Database connection and setup
│   │   ├── models.go           # Data models
│   │   ├── devices.go          # Device CRUD operations
│   │   ├── cycles.go           # Cycle CRUD operations
│   │   ├── rdg_status.go       # Getinge status operations
│   │   ├── audit.go            # Audit log operations
│   │   └── migrations/
│   │       └── 001_initial_schema.sql
│   ├── config/
│   │   └── config.go           # Configuration management
│   ├── logging/
│   │   ├── logger.go           # Structured logging
│   │   ├── buffer.go           # Log buffer for Test UI
│   │   └── buffered_handler.go # Buffered log handler
│   ├── devices/
│   │   ├── manager.go          # Device manager
│   │   └── manager_ping.go     # Ping monitoring
│   └── testui/
│       ├── handlers.go         # Test UI HTTP handler
│       └── templates/
│           └── index.html      # Test UI HTML template
├── web/
│   └── testui/
│       ├── css/
│       │   └── style.css       # Test UI styles
│       └── js/
│           └── app.js          # Test UI JavaScript
├── config/
│   └── config.yaml             # Configuration file
├── docs/                       # Documentation
├── go.mod                      # Go module definition
└── README.md                   # Project README
```

## Development Setup

### 1. Clone Repository

```bash
git clone <repository-url>
cd Steri-Connect-Melag-Getinge-GO
```

### 2. Install Dependencies

```bash
go mod download
```

### 3. Build Application

```bash
# Development build
go build ./cmd/server

# Production build
go build -ldflags="-s -w" -o steri-connect-go ./cmd/server
```

### 4. Run Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests for specific package
go test ./internal/config
```

### 5. Run Application

```bash
# Run from source
go run ./cmd/server

# Or run built executable
./steri-connect-go
```

## Code Organization

### Package Structure

- **`cmd/server`** - Application entry point, initialization
- **`internal/api`** - HTTP API layer (handlers, middleware, router)
- **`internal/adapters`** - Device abstraction layer
- **`internal/database`** - Database operations and models
- **`internal/config`** - Configuration management
- **`internal/logging`** - Structured logging
- **`internal/devices`** - Device manager and lifecycle
- **`internal/testui`** - Test UI components

### Key Design Patterns

#### Device Adapter Pattern

Devices are abstracted through the `DeviceAdapter` interface:

```go
type DeviceAdapter interface {
    Connect() error
    Disconnect() error
    IsConnected() bool
    GetDeviceID() int
    StartCycle(params CycleStartParams) error
    GetCycleStatus() (CycleStatus, error)
}
```

Each device manufacturer (Melag, Getinge) implements this interface.

#### Manager Pattern

The `DeviceManager` coordinates device adapters:
- Loads devices from database
- Manages adapter lifecycle
- Handles connection retries
- Coordinates cycle polling
- Broadcasts WebSocket events

#### Handler Pattern

HTTP handlers follow a consistent pattern:
1. Validate HTTP method
2. Parse request (path parameters, query params, body)
3. Validate input
4. Call business logic (database, device manager)
5. Return JSON response
6. Handle errors consistently

## Adding New Features

### Adding a New API Endpoint

1. **Create Handler Function**

```go
// internal/api/handlers/myfeature.go
package handlers

func MyFeatureHandler(w http.ResponseWriter, r *http.Request) {
    // Implementation
}
```

2. **Add Route**

```go
// internal/api/router.go
apiHandler.HandleFunc("/myfeature", handlers.MyFeatureHandler)
```

3. **Add Tests**

```go
// internal/api/handlers/myfeature_test.go
func TestMyFeatureHandler(t *testing.T) {
    // Test implementation
}
```

### Adding a New Device Type

1. **Create Adapter**

```go
// internal/adapters/mydevice/mydevice.go
package mydevice

type MyDeviceAdapter struct {
    // Implementation
}

func (a *MyDeviceAdapter) Connect() error {
    // Implementation
}
```

2. **Register in Device Manager**

```go
// internal/devices/manager.go
switch device.Manufacturer {
case "MyDevice":
    adapter, err = mydevice.NewMyDeviceAdapter(device)
}
```

3. **Add Device-Specific Endpoints** (if needed)

```go
// internal/api/handlers/mydevice.go
func MyDeviceOperationHandler(w http.ResponseWriter, r *http.Request) {
    // Implementation
}
```

## Testing

### Unit Tests

Place test files next to source files with `_test.go` suffix:

```go
// internal/config/config_test.go
package config

import "testing"

func TestLoad(t *testing.T) {
    // Test implementation
}
```

### Integration Tests

Create integration tests in `internal/*/integration_test.go`:

```go
// internal/database/integration_test.go
// +build integration

package database

import "testing"

func TestDatabaseIntegration(t *testing.T) {
    // Integration test implementation
}
```

Run with: `go test -tags=integration ./...`

### Test Coverage

```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...

# View coverage report
go tool cover -html=coverage.out
```

## Database Migrations

### Adding a Migration

1. Create new migration file:

```sql
-- internal/database/migrations/002_add_new_table.sql
CREATE TABLE IF NOT EXISTS new_table (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    -- columns
);
```

2. Update migration runner in `internal/database/sqlite.go`:

```go
func runMigrations() error {
    migrations := []string{
        readMigration("001_initial_schema.sql"),
        readMigration("002_add_new_table.sql"),
    }
    // Execute migrations
}
```

## Logging

### Structured Logging

Use the structured logger:

```go
logger := logging.Get()
logger.Info("Operation completed",
    "device_id", deviceID,
    "cycle_id", cycleID,
    "result", "success")
```

### Log Levels

- **DEBUG** - Detailed debugging information
- **INFO** - General informational messages
- **WARN** - Warning messages
- **ERROR** - Error messages

### Context Logging

```go
logger := logging.Get().WithDevice(deviceID).WithCycle(cycleID)
logger.Info("Cycle started")
```

## Configuration Management

### Adding Configuration

1. **Update Config Struct**

```go
// internal/config/config.go
type Config struct {
    Server ServerConfig
    MyFeature MyFeatureConfig  // New config section
}

type MyFeatureConfig struct {
    Enabled bool
    Timeout int
}
```

2. **Add to config.yaml**

```yaml
my_feature:
  enabled: true
  timeout: 30
```

3. **Use in Code**

```go
cfg := config.Get()
if cfg.MyFeature.Enabled {
    // Use feature
}
```

## WebSocket Events

### Broadcasting Events

```go
import "steri-connect-go/internal/api/websocket"

websocket.BroadcastEvent(websocket.Event{
    Event: "my_event",
    Data: map[string]interface{}{
        "field": "value",
    },
})
```

### Event Types

- `device_status_change` - Device connection status changed
- `cycle_started` - New cycle started
- `cycle_status_update` - Cycle progress update
- `cycle_completed` - Cycle finished successfully
- `cycle_failed` - Cycle failed with error

## Code Style

### Formatting

```bash
# Format code
go fmt ./...

# Run linter
golangci-lint run
```

### Naming Conventions

- **Packages:** lowercase, single word
- **Exported functions:** PascalCase
- **Unexported functions:** camelCase
- **Constants:** PascalCase or UPPER_CASE
- **Variables:** camelCase

### Error Handling

Always handle errors explicitly:

```go
result, err := operation()
if err != nil {
    logger.Error("Operation failed", "error", err)
    return err
}
```

## Debugging

### Enable Debug Logging

```yaml
# config/config.yaml
logging:
  level: "DEBUG"
```

### Database Inspection

Use Test UI Database tab or SQLite CLI:

```bash
sqlite3 data/steri-connect.db
.tables
SELECT * FROM devices;
```

### Profiling

```bash
# CPU profiling
go tool pprof http://localhost:8080/debug/pprof/profile

# Memory profiling
go tool pprof http://localhost:8080/debug/pprof/heap
```

## Contributing

### Workflow

1. Create feature branch: `git checkout -b feature/my-feature`
2. Make changes and test
3. Run tests: `go test ./...`
4. Format code: `go fmt ./...`
5. Commit changes: `git commit -m "Add my feature"`
6. Push branch: `git push origin feature/my-feature`
7. Create pull request

### Code Review Checklist

- [ ] Code follows project style guidelines
- [ ] Tests added/updated
- [ ] Documentation updated
- [ ] No linter errors
- [ ] Error handling implemented
- [ ] Logging added where appropriate

## Resources

- **API Reference:** `docs/API-Reference.md`
- **Architecture:** `docs/architecture.md`
- **PRD:** `docs/PRD-Steri-Connect-Melag-Getinge-GO.md`
- **Go Documentation:** https://go.dev/doc/

