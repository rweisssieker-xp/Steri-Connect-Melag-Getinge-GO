# Steri-Connect-Melag-Getinge-GO

Middleware/Backend service for integrating medical sterilization devices (Melag Cliniclave 45 and Getinge Aquadis 56) with the Steri-Suite frontend application.

## Overview

This GO-App provides a local middleware layer that bridges communication between the existing Steri-Suite frontend and medical sterilization devices. It offers device abstraction, local-first data storage (SQLite), and a portable executable deployment model.

**Key Features:**
- Device abstraction layer for multiple device types (Melag, Getinge)
- REST API and WebSocket support for real-time communication
- Local SQLite database for all data storage
- Portable executable (no installation required)
- Web-based test interface for development and debugging

## Prerequisites

- **Go 1.21+** - [Download Go](https://go.dev/dl/)
- **Git** - For cloning the repository
- **Network access** - To reach medical devices on local network

## Setup Instructions

### 1. Clone Repository

```bash
git clone <repository-url>
cd Steri-Connect-Melag-Getinge-GO
```

### 2. Install Dependencies

```bash
go mod download
```

This will download all required dependencies:
- `modernc.org/sqlite` - Pure Go SQLite driver (no CGO required)
- `github.com/gorilla/websocket` - WebSocket support
- `gopkg.in/yaml.v3` - YAML configuration parser

### 3. Configuration Setup

The application uses a YAML configuration file at `config/config.yaml` with default values:

- **Server Port:** 8080 (default)
- **Bind Address:** `127.0.0.1` (localhost only by default for security)
  - Use `0.0.0.0` for network access (requires API key authentication)
- **Database Path:** `./data/steri-connect.db` (created automatically)
- **Log Level:** INFO
- **Test UI:** Enabled (for development)

**Security Note:** By default, the server binds to `127.0.0.1` (localhost only) for security. If you need network access, set `bind_address: "0.0.0.0"` in the config file and **enable API key authentication** (`api_key_required: true`).

You can modify `config/config.yaml` to change these settings.

### 4. Build Application

```bash
# Development build
go build ./cmd/server

# Production build (strip debug info, no CGO)
CGO_ENABLED=0 go build -ldflags="-s -w" -o steri-connect-go ./cmd/server

# Cross-compile for Windows (from Linux/Mac)
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o steri-connect-go.exe ./cmd/server

# Cross-compile for Linux (from Windows/Mac)
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o steri-connect-go ./cmd/server
```

### 5. Run Application

```bash
# Run from source
go run ./cmd/server

# Or run built executable
./steri-connect-go  # Linux/Mac
# or
steri-connect-go.exe  # Windows
```

The application will:
- Create SQLite database at `./data/steri-connect.db` on first run
- Start HTTP server on `http://localhost:8080`
- Make API available at `http://localhost:8080/api/`
- Provide WebSocket endpoint at `ws://localhost:8080/ws`
- Serve test UI at `http://localhost:8080/test-ui` (if enabled)

## Project Structure

```
steri-connect-go/
├── cmd/
│   └── server/
│       └── main.go                 # Application entry point
├── internal/
│   ├── api/
│   │   ├── handlers/              # HTTP request handlers
│   │   ├── middleware/            # HTTP middleware
│   │   └── websocket/             # WebSocket handlers
│   ├── adapters/
│   │   ├── device/                # Device abstraction interface
│   │   ├── melag/                 # Melag device adapter
│   │   └── getinge/               # Getinge device adapter
│   ├── database/
│   │   ├── migrations/            # Database migrations
│   │   ├── sqlite.go              # SQLite connection and setup
│   │   └── models.go              # Data models
│   ├── config/
│   │   └── config.go              # Configuration management
│   ├── logging/
│   │   └── logger.go              # Structured logging
│   └── testui/
│       ├── handlers.go            # Test UI HTTP handlers
│       └── templates/             # HTML templates
├── pkg/
│   └── utils/                     # Shared utilities
├── web/
│   └── testui/                    # Test UI static assets
│       ├── css/
│       ├── js/
│       └── index.html
├── config/
│   └── config.yaml                # Configuration file
├── data/                          # Database files (created at runtime)
├── docs/                          # Project documentation
├── go.mod                         # Go module definition
├── go.sum                         # Go dependency checksums
└── README.md                      # This file
```

## Development

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests for specific package
go test ./internal/config
```

### Development Workflow

1. Make changes to source code
2. Run tests: `go test ./...`
3. Build: `go build ./cmd/server`
4. Test locally: `go run ./cmd/server`

## API Endpoints

### Base URL
`http://localhost:8080/api/`

### Key Endpoints
- `GET /api/health` - System health check
- `GET /api/devices` - List all devices
- `POST /api/devices` - Add new device
- `GET /api/melag/{id}/status` - Get Melag device status
- `POST /api/melag/{id}/start` - Start Melag cycle
- `GET /api/cycles` - List all cycles

See `docs/PRD-Steri-Connect-Melag-Getinge-GO.md` Section 7 for complete API documentation.

## WebSocket Events

Connect to `ws://localhost:8080/ws` for real-time events:
- `cycle_status_update` - Cycle progress updates
- `device_status_change` - Device connection status changes

## Test UI

Access the test interface at `http://localhost:8080/test-ui` (if enabled in config):
- Device management
- API endpoint testing
- Database inspection
- Log viewing
- System status

## Configuration

See `config/config.yaml` for configuration options:
- Server settings (port, bind address)
- Database path
- Logging configuration
- Device polling intervals
- Test UI settings

## Documentation

### Getting Started
- **User Guide:** `docs/User-Guide.md` - How to use the application
- **API Reference:** `docs/API-Reference.md` - Complete API documentation
- **Troubleshooting Guide:** `docs/Troubleshooting-Guide.md` - Common issues and solutions

### Development
- **Developer Guide:** `docs/Developer-Guide.md` - Development setup and workflow
- **Architecture:** `docs/architecture.md` - System architecture and design decisions

### Deployment
- **Deployment Guide:** `docs/Deployment-Guide.md` - Production deployment instructions

### Planning & Design
- **PRD:** `docs/PRD-Steri-Connect-Melag-Getinge-GO.md` - Product requirements
- **Epics & Stories:** `docs/epics.md` - Feature breakdown
- **Test Design:** `docs/test-design-system.md` - Testing strategy

## Quick Links

- **User Guide:** See `docs/User-Guide.md` for usage instructions
- **API Documentation:** See `docs/API-Reference.md` for complete API reference
- **Troubleshooting:** See `docs/Troubleshooting-Guide.md` for common issues
- **Deployment:** See `docs/Deployment-Guide.md` for production deployment

## License

[License information to be added]

## Support

For support and questions:
- Review the Troubleshooting Guide: `docs/Troubleshooting-Guide.md`
- Check the API Reference: `docs/API-Reference.md`
- Review application logs for detailed error information

