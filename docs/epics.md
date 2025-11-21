# Steri-Connect-Melag-Getinge-GO - Epic Breakdown

**Author:** BMad
**Date:** 2025-11-21T09:33:54.802Z
**Project Level:** Level 2-3 (Medium complexity)
**Target Scale:** Single-instance deployment

---

## Overview

This document provides the complete epic and story breakdown for Steri-Connect-Melag-Getinge-GO, decomposing the requirements from the [PRD](./PRD-Steri-Connect-Melag-Getinge-GO.md) into implementable stories.

**Epic Summary:**
- **7 Epics** covering MVP scope
- **34 Stories** total (MVP)
- **Foundation → Device Management → Device Integration → Cycle Management → Test Tools → Monitoring**

---

## Functional Requirements Inventory

**Total Functional Requirements:** 37
- **MVP Requirements:** 34 (FR-001 to FR-010, FR-012 to FR-023, FR-025 to FR-035, FR-037)
- **Growth Requirements:** 3 (FR-011, FR-021, FR-024, FR-036)
- **Vision Requirements:** Future Getinge full integration

**FR Categories:**
- Steri-Suite Integration: FR-001 to FR-010
- GO-Schnittstelle Core: FR-012 to FR-015
- Melag Integration: FR-016 to FR-019
- Getinge Integration: FR-020 to FR-021
- Security: FR-022 to FR-024, FR-035 to FR-036
- Test UI: FR-025 to FR-031
- Health & Diagnostics: FR-032 to FR-034
- WebSocket: FR-037

---

## FR Coverage Map

| FR | Epic | Story | Status |
|----|------|-------|--------|
| FR-001 | Epic 2 | Story 2.1 | MVP |
| FR-002 | Epic 2 | Story 2.2 | MVP |
| FR-003 | Epic 2 | Story 2.3 | MVP |
| FR-004 | Epic 2 | Story 2.4 | MVP |
| FR-005 | Epic 5 | Story 5.1 | MVP |
| FR-006 | Epic 5 | Story 5.2 | MVP |
| FR-007 | Epic 5 | Story 5.3 | MVP |
| FR-008 | Epic 5 | Story 5.4 | MVP |
| FR-009 | Epic 5 | Story 5.5 | MVP |
| FR-010 | Epic 5 | Story 5.6 | MVP |
| FR-011 | Epic 5 | Story 5.7 | Growth |
| FR-012 | Epic 1 | Story 1.2 | MVP |
| FR-013 | Epic 1 | Story 1.3 | MVP |
| FR-014 | Epic 1 | Story 1.4 | MVP |
| FR-015 | Epic 1 | Story 1.5 | MVP |
| FR-016 | Epic 3 | Story 3.1 | MVP |
| FR-017 | Epic 3 | Story 3.2 | MVP |
| FR-018 | Epic 3 | Story 3.3 | MVP |
| FR-019 | Epic 3 | Story 3.4 | MVP |
| FR-020 | Epic 4 | Story 4.1 | MVP |
| FR-021 | Epic 4 | Story 4.2 | Growth |
| FR-022 | Epic 1 | Story 1.6 | MVP |
| FR-023 | Epic 1 | Story 1.7 | MVP |
| FR-024 | Epic 1 | Story 1.8 | Growth |
| FR-025 | Epic 6 | Story 6.1 | MVP |
| FR-026 | Epic 6 | Story 6.2 | MVP |
| FR-027 | Epic 6 | Story 6.3 | MVP |
| FR-028 | Epic 6 | Story 6.4 | MVP |
| FR-029 | Epic 6 | Story 6.5 | MVP |
| FR-030 | Epic 6 | Story 6.6 | MVP |
| FR-031 | Epic 6 | Story 6.7 | MVP |
| FR-032 | Epic 7 | Story 7.1 | MVP |
| FR-033 | Epic 7 | Story 7.2 | MVP |
| FR-034 | Epic 7 | Story 7.3 | MVP |
| FR-035 | Epic 1 | Story 1.6 | MVP |
| FR-036 | Epic 1 | Story 1.8 | Growth |
| FR-037 | Epic 1 | Story 1.3 | MVP |

---

## Epic 1: Foundation & Core Infrastructure

**Goal:** Establish the foundational infrastructure for the GO-App middleware, including project setup, database, API framework, and core services.

**Value Proposition:** Provides the base infrastructure that enables all device communication and API functionality.

**MVP Scope:** Core infrastructure required for all device operations.

---

### Story 1.1: Project Setup and Initialization

As a **developer**,
I want **a properly structured Go project with dependencies and configuration**,
So that **I can start implementing features immediately**.

**Acceptance Criteria:**

**Given** a new project directory
**When** I initialize the Go project
**Then** the project has:
- `go.mod` file with project name
- Standard project structure (cmd/, internal/, pkg/)
- Configuration file structure (`config/config.yaml`)
- README with setup instructions

**And** dependencies are defined:
- SQLite driver (`github.com/mattn/go-sqlite3`)
- WebSocket library (`github.com/gorilla/websocket`)
- YAML parser (`gopkg.in/yaml.v3`)
- HTTP router (standard library or Gin if needed)

**Prerequisites:** None (foundation story)

**Technical Notes:**
- Use `go mod init steri-connect-go`
- Create directory structure per Architecture document
- Set up basic configuration file with defaults
- Reference: Architecture Section "Project Structure"

**FR Coverage:** Foundation (enables all FRs)

---

### Story 1.2: SQLite Database Setup and Schema

As a **developer**,
I want **SQLite database with proper schema for devices, cycles, and audit logs**,
So that **all application data can be stored locally**.

**Acceptance Criteria:**

**Given** the application starts for the first time
**When** database initialization runs
**Then** SQLite database is created at `./data/steri-connect.db`

**And** the following tables exist:
- `devices` table with all required fields (id, name, model, manufacturer, ip, serial, type, location, created, updated)
- `cycles` table with all required fields (id, device_id, program, start_ts, end_ts, result, error_code, error_description, phase, temperature, pressure, progress_percent)
- `rdg_status` table for Getinge status (id, device_id, timestamp, reachable)
- `audit_log` table for audit trails (id, timestamp, action, entity_type, entity_id, user, details, hash)

**And** indexes are created:
- `idx_cycles_device_id` on cycles(device_id)
- `idx_cycles_start_ts` on cycles(start_ts)
- `idx_rdg_status_device_id` on rdg_status(device_id)
- `idx_rdg_status_timestamp` on rdg_status(timestamp)
- `idx_audit_log_timestamp` on audit_log(timestamp)
- `idx_audit_log_entity` on audit_log(entity_type, entity_id)

**And** foreign key constraints are enforced

**Prerequisites:** Story 1.1

**Technical Notes:**
- Use migrations or schema initialization script
- Enable WAL mode for SQLite (Write-Ahead-Logging)
- Reference: Architecture Section "Data Architecture"

**FR Coverage:** FR-014 (SQLite-Integration)

---

### Story 1.3: REST API Server and WebSocket Setup

As a **developer**,
I want **HTTP server with REST endpoints and WebSocket support**,
So that **Steri-Suite can communicate with the GO-App**.

**Acceptance Criteria:**

**Given** the application starts
**When** the server initializes
**Then** HTTP server listens on configurable port (default: 8080)

**And** REST API endpoints are available:
- Base path: `/api/`
- Health endpoint: `GET /api/health`
- Devices endpoints: `GET /api/devices`, `POST /api/devices`, etc.

**And** WebSocket endpoint is available:
- `ws://localhost:{port}/ws`
- Supports multiple concurrent connections
- Broadcasts events to all connected clients

**And** CORS is configured (if needed for Steri-Suite)

**Prerequisites:** Story 1.1

**Technical Notes:**
- Use Go standard library `net/http` or Gin framework
- WebSocket using Gorilla WebSocket library
- Reference: Architecture Section "API Contracts"

**FR Coverage:** FR-012 (HTTP-API), FR-013 (WebSocket), FR-037 (WebSocket-Verbindung)

---

### Story 1.4: Structured Logging and Audit Trail

As a **developer**,
I want **structured logging with audit trail support**,
So that **all operations are logged for debugging and compliance**.

**Acceptance Criteria:**

**Given** the application runs
**When** any operation occurs (device add, cycle start, etc.)
**Then** structured logs are written:
- JSON format for production
- Log levels: DEBUG, INFO, WARN, ERROR
- Context fields: device_id, cycle_id, user, action

**And** audit logs are written to `audit_log` table:
- All device operations logged
- All cycle operations logged
- Hash calculated for integrity verification
- Immutable (append-only)

**And** log rotation is configured:
- Max file size: 10MB
- Keep last 10 files
- Compress old logs

**Prerequisites:** Story 1.2 (database)

**Technical Notes:**
- Use structured logging library (`log/slog` or `logrus`)
- Audit log hash calculation for integrity
- Reference: Architecture Section "Logging Strategy"

**FR Coverage:** FR-015 (Logging)

---

### Story 1.5: Configuration Management

As a **developer**,
I want **configuration file support with environment overrides**,
So that **the application can be configured without code changes**.

**Acceptance Criteria:**

**Given** a configuration file exists
**When** the application starts
**Then** configuration is loaded from `config.yaml`

**And** configuration includes:
- Server port (default: 8080)
- Database path (default: `./data/steri-connect.db`)
- Log level (default: INFO)
- API authentication settings
- Device polling intervals
- Test UI enabled/disabled

**And** environment variables can override config values

**And** default values are used if config file missing

**Prerequisites:** Story 1.1

**Technical Notes:**
- YAML configuration file
- Configuration struct with defaults
- Reference: Architecture Section "Deployment Architecture"

**FR Coverage:** Foundation (enables configuration)

---

### Story 1.6: API Authentication (Optional)

As a **developer**,
I want **API key authentication support**,
So that **the API can be secured when needed**.

**Acceptance Criteria:**

**Given** API authentication is enabled in config
**When** a request is made to the API
**Then** API key is validated from `X-API-Key` header

**And** if API key is missing or invalid:
- Request is rejected with 401 Unauthorized
- Error message is returned

**And** if API key is valid:
- Request proceeds normally

**And** authentication is optional for localhost (configurable)

**Prerequisites:** Story 1.3 (API server)

**Technical Notes:**
- Middleware for authentication
- API key stored in configuration
- Reference: Architecture Section "Security Architecture"

**FR Coverage:** FR-022 (API-Authentifizierung), FR-035 (API-Key Authentication)

---

### Story 1.7: Localhost-Only Access Control

As a **developer**,
I want **localhost-only access by default**,
So that **the API is secure by default**.

**Acceptance Criteria:**

**Given** the application starts
**When** server binds to network interface
**Then** by default, server only accepts localhost connections

**And** network access can be enabled via configuration:
- `bind_address: "0.0.0.0"` for network access
- `bind_address: "127.0.0.1"` for localhost only (default)

**And** Test UI is only accessible via localhost (always)

**Prerequisites:** Story 1.3 (API server)

**Technical Notes:**
- Network binding configuration
- Reference: Architecture Section "Security Architecture"

**FR Coverage:** FR-023 (Localhost-Zugriff)

---

### Story 1.8: Role-Based Access Control (Growth)

As a **developer**,
I want **role-based access control**,
So that **different user roles have appropriate permissions**.

**Acceptance Criteria:**

**Given** role-based access control is enabled
**When** a request is made
**Then** user role is determined from authentication

**And** permissions are checked:
- Operator: Can start cycles, view status
- Techniker: Can manage devices, view diagnostics
- QA: Can view all cycles, export protocols
- Administrator: Full access including Test UI

**And** unauthorized requests return 403 Forbidden

**Prerequisites:** Story 1.6 (Authentication)

**Technical Notes:**
- RBAC middleware
- Permission matrix
- Reference: PRD Section 8 "Benutzerrollen"

**FR Coverage:** FR-024 (Rollenbasierte Zugriffskontrolle), FR-036 (Token-based Authentication)

---

## Epic 2: Device Management

**Goal:** Enable users to add, configure, and monitor medical devices through the GO-App.

**Value Proposition:** Users can manage all their sterilization devices from a single interface, with real-time health monitoring.

**MVP Scope:** Complete device CRUD operations and health monitoring.

---

### Story 2.1: Add Device via API

As a **technician**,
I want **to add a new device to the system**,
So that **the device can be monitored and controlled**.

**Acceptance Criteria:**

**Given** I have device information (name, type, IP, manufacturer, serial)
**When** I send `POST /api/devices` with device data
**Then** device is created in database

**And** device record includes:
- Name, model, manufacturer (Melag/Getinge)
- IP address, serial number
- Type (Steri/RDG)
- Location (optional)
- Created timestamp

**And** response returns device ID and success message

**And** device is immediately available for monitoring

**Prerequisites:** Story 1.2 (database), Story 1.3 (API)

**Technical Notes:**
- Device validation (IP format, required fields)
- Duplicate check (IP + manufacturer)
- Reference: Architecture Section "Data Architecture"

**FR Coverage:** FR-001 (Geräte hinzufügen)

---

### Story 2.2: Update Device Configuration

As a **technician**,
I want **to update device configuration**,
So that **I can correct information or change settings**.

**Acceptance Criteria:**

**Given** a device exists in the system
**When** I send `PUT /api/devices/{id}` with updated data
**Then** device record is updated in database

**And** updated timestamp is set

**And** response returns updated device data

**And** if device is currently connected, connection is re-established with new settings

**Prerequisites:** Story 2.1

**Technical Notes:**
- Update only provided fields
- Validate IP changes
- Reference: Architecture API Contracts

**FR Coverage:** FR-002 (Geräte bearbeiten)

---

### Story 2.3: Delete Device

As a **technician**,
I want **to remove a device from the system**,
So that **I can clean up devices that are no longer in use**.

**Acceptance Criteria:**

**Given** a device exists in the system
**When** I send `DELETE /api/devices/{id}`
**Then** device is removed from database

**And** if device has associated cycles:
- Option 1: Prevent deletion (return error)
- Option 2: Soft delete (mark as deleted, keep cycles)

**And** response confirms deletion

**And** device connection is closed if active

**Prerequisites:** Story 2.1

**Technical Notes:**
- Foreign key constraints consideration
- Soft delete vs hard delete decision
- Reference: Architecture Data Architecture

**FR Coverage:** FR-003 (Geräte löschen)

---

### Story 2.4: Device Health Monitoring and Status Display

As a **technician**,
I want **to see the health status of all devices**,
So that **I can identify devices that need attention**.

**Acceptance Criteria:**

**Given** devices are configured in the system
**When** I request `GET /api/devices`
**Then** response includes health status for each device:
- Connection status (online/offline/error)
- Last seen timestamp
- Current state (for Melag: ready/running/completed)

**And** health status is updated in real-time:
- Melag: Connection status from adapter
- Getinge: ICMP ping result (online/offline)

**And** status changes are broadcast via WebSocket:
- `device_status_change` event sent to all connected clients

**And** status updates occur:
- Melag: Every 2 seconds (or on change)
- Getinge: Every 10-30 seconds (configurable)

**Prerequisites:** Story 2.1, Story 1.3 (WebSocket)

**Technical Notes:**
- Status polling in background goroutines
- Status caching in memory
- WebSocket event broadcasting
- Reference: Architecture "Implementation Patterns"

**FR Coverage:** FR-004 (Gesundheitsmonitoring)

---

## Epic 3: Melag Device Integration

**Goal:** Enable full integration with Melag Cliniclave 45 devices, including cycle control, status monitoring, and result retrieval.

**Value Proposition:** Users can start sterilization cycles, monitor progress in real-time, and automatically retrieve cycle results from Melag devices.

**MVP Scope:** Complete Melag integration via MELAnet Box (FTP-based).

---

### Story 3.1: Melag Device Connection via MELAnet Box

As a **developer**,
I want **to establish connection to Melag device via MELAnet Box**,
So that **I can communicate with the device**.

**Acceptance Criteria:**

**Given** a Melag device is configured with IP address
**When** connection is initiated
**Then** system performs:
- ICMP ping to verify device reachability
- FTP connection to MELAnet Box (if applicable)
- Connection status stored in memory

**And** if connection fails:
- Error is logged
- Status set to "error"
- Retry logic attempts reconnection (exponential backoff)

**And** connection state is tracked:
- `DISCONNECTED` → `CONNECTING` → `CONNECTED` or `ERROR`

**Prerequisites:** Story 2.1 (device management), Story 1.2 (database)

**Technical Notes:**
- Melag adapter implementation
- FTP client for MELAnet Box
- Connection state machine
- Reference: Architecture "Device Abstraction Pattern", Research "MELAnet Box Integration"

**FR Coverage:** FR-016 (Melag-Verbindung initialisieren)

---

### Story 3.2: Start Melag Sterilization Cycle

As an **operator**,
I want **to start a sterilization cycle on a Melag device**,
So that **sterilization can begin without manual device operation**.

**Acceptance Criteria:**

**Given** a Melag device is connected and ready
**When** I send `POST /api/melag/{id}/start` with optional program parameters
**Then** system:
- Validates device is available and ready
- Sends start command to Melag device (via MELAnet Box FTP or direct)
- Creates cycle record in database with status "STARTING"
- Returns cycle ID and confirmation

**And** if start fails:
- Error is logged
- Cycle record marked as "FAILED"
- Error response returned to caller

**And** cycle state transitions: `READY` → `STARTING` → `RUNNING`

**Prerequisites:** Story 3.1 (connection)

**Technical Notes:**
- Melag adapter `StartCycle()` method
- Protocol-specific start command (format TBD from manufacturer docs)
- Cycle state management
- Reference: Architecture "State Management Pattern"

**FR Coverage:** FR-005 (Zyklus starten), FR-017 (Melag-Zyklus starten)

---

### Story 3.3: Real-Time Melag Cycle Status Monitoring

As an **operator**,
I want **to see real-time status of running Melag cycles**,
So that **I can monitor sterilization progress**.

**Acceptance Criteria:**

**Given** a Melag cycle is running
**When** status is polled (every 2 seconds)
**Then** system retrieves status from device:
- Current phase (Aufheizen, Sterilisation, Trocknung)
- Temperature and pressure (if available)
- Time remaining
- Progress percentage

**And** status is:
- Updated in database (cycle record)
- Broadcast via WebSocket (`cycle_status_update` event)
- Available via `GET /api/melag/{id}/status`

**And** status updates continue until cycle completes

**Prerequisites:** Story 3.2 (start cycle)

**Technical Notes:**
- Status polling in goroutine
- WebSocket event broadcasting
- Status parsing from device protocol
- Reference: Architecture "Concurrency Pattern", PRD Section 7.4

**FR Coverage:** FR-006 (Laufende Zyklen anzeigen), FR-007 (Echtzeitstatus), FR-018 (Melag-Status abrufen)

---

### Story 3.4: Retrieve Melag Cycle Results

As an **operator**,
I want **to retrieve completed cycle results automatically**,
So that **cycle documentation is complete without manual entry**.

**Acceptance Criteria:**

**Given** a Melag cycle completes
**When** cycle ends
**Then** system:
- Retrieves final cycle result from device
- Extracts: OK/NOK status, program type, timestamp, error codes (if any)
- Updates cycle record in database:
  - Sets `end_ts`
  - Sets `result` (OK/NOK)
  - Sets error information if failed
- Broadcasts completion event via WebSocket

**And** cycle result is available via:
- `GET /api/melag/{id}/last-cycle`
- `GET /api/cycles/{cycle_id}`

**And** cycle state transitions: `RUNNING` → `COMPLETED` or `FAILED`

**Prerequisites:** Story 3.3 (status monitoring)

**Technical Notes:**
- Protocol file parsing (format TBD from manufacturer)
- Result extraction and validation
- Error handling for incomplete results
- Reference: Research "MELAnet Box Integration"

**FR Coverage:** FR-019 (Melag-Zyklusergebnis abrufen)

---

## Epic 4: Getinge Device Monitoring

**Goal:** Enable basic monitoring of Getinge Aquadis 56 devices via ICMP ping, providing online/offline status visibility.

**Value Proposition:** Users can see which Getinge devices are online and available, even without full API integration.

**MVP Scope:** ICMP ping monitoring with status tracking.

---

### Story 4.1: Getinge ICMP Ping Monitoring

As a **technician**,
I want **to monitor Getinge device online/offline status**,
So that **I know when devices are available**.

**Acceptance Criteria:**

**Given** a Getinge device is configured
**When** monitoring starts
**Then** system performs ICMP ping every 10-30 seconds (configurable)

**And** ping results are stored in `rdg_status` table:
- Timestamp
- Reachable (0 or 1)

**And** device status is updated:
- `online` if ping successful
- `offline` if ping fails
- Status change broadcast via WebSocket

**And** status is available via:
- `GET /api/getinge/{id}/ping` (current status)
- `GET /api/devices/{id}` (includes Getinge status)

**Prerequisites:** Story 2.1 (device management), Story 1.2 (database)

**Technical Notes:**
- ICMP ping using Go `net` package
- Background goroutine for periodic pings
- Status persistence in database
- Reference: Architecture "Getinge Adapter", PRD FR-020

**FR Coverage:** FR-020 (Getinge-ICMP-Monitoring)

---

### Story 4.2: Getinge Failure Notification (Growth)

As a **technician**,
I want **to receive notifications when Getinge devices go offline**,
So that **I can address connectivity issues promptly**.

**Acceptance Criteria:**

**Given** Getinge device monitoring is active
**When** device becomes unreachable for configured duration (e.g., 5 minutes)
**Then** notification is sent:
- Logged as ERROR
- Optional: WebSocket event to connected clients
- Optional: Email/alert (if configured)

**And** notification includes:
- Device name and IP
- Duration of outage
- Last successful ping timestamp

**Prerequisites:** Story 4.1 (ICMP monitoring)

**Technical Notes:**
- Failure threshold configuration
- Notification mechanism (log, WebSocket, email)
- Reference: PRD FR-021

**FR Coverage:** FR-021 (Getinge-Ausfallbenachrichtigung)

---

## Epic 5: Cycle Management and Protocols

**Goal:** Enable users to view cycle history, export protocols, and manage cycle data through the API.

**Value Proposition:** Users have complete visibility into all sterilization cycles with export capabilities for compliance and reporting.

**MVP Scope:** Cycle listing, details, and export functionality.

---

### Story 5.1: List All Cycles

As a **quality assurance officer**,
I want **to see a list of all sterilization cycles**,
So that **I can review cycle history**.

**Acceptance Criteria:**

**Given** cycles exist in the database
**When** I request `GET /api/cycles`
**Then** response includes list of cycles with:
- Cycle ID, device name, start time, end time
- Result (OK/NOK)
- Program type

**And** list supports:
- Pagination (limit/offset)
- Sorting (by start time, device, result)
- Filtering (by device, date range, result)

**And** response includes total count for pagination

**Prerequisites:** Story 3.4 (cycle results), Story 1.2 (database)

**Technical Notes:**
- Database query with pagination
- Filter and sort parameters
- Reference: Architecture "API Contracts"

**FR Coverage:** FR-008 (Zyklus-Listenansicht)

---

### Story 5.2: View Cycle Details

As a **quality assurance officer**,
I want **to see detailed information about a specific cycle**,
So that **I can review complete cycle parameters**.

**Acceptance Criteria:**

**Given** a cycle exists
**When** I request `GET /api/cycles/{id}`
**Then** response includes complete cycle details:
- Device information
- Start and end timestamps
- Program type
- Result (OK/NOK)
- All process parameters (temperature, pressure, phases)
- Error information (if failed)
- Progress history (if available)

**And** response format matches PRD API schema

**Prerequisites:** Story 5.1 (list cycles)

**Technical Notes:**
- Database query with joins to device table
- JSON response formatting
- Reference: PRD Section 7.2

**FR Coverage:** FR-009 (Zyklus-Detailansicht)

---

### Story 5.3: Export Cycle Protocol as PDF

As a **quality assurance officer**,
I want **to export cycle protocols as PDF**,
So that **I can generate compliance documentation**.

**Acceptance Criteria:**

**Given** a cycle exists
**When** I request cycle export (via API or Test UI)
**Then** PDF document is generated containing:
- Cycle header (ID, device, dates)
- Process parameters (temperature, pressure, phases)
- Result and error information
- Audit information (timestamps, user)

**And** PDF format is:
- Professional and readable
- Suitable for compliance documentation
- Includes all required audit information

**Prerequisites:** Story 5.2 (cycle details)

**Technical Notes:**
- PDF generation library (e.g., `github.com/jung-kurt/gofpdf`)
- Template for PDF format
- Reference: PRD Section 3.5

**FR Coverage:** FR-010 (Protokoll-Export - PDF)

---

### Story 5.4: Export Cycle Protocol as CSV

As a **quality assurance officer**,
I want **to export cycle protocols as CSV**,
So that **I can analyze cycle data in spreadsheet applications**.

**Acceptance Criteria:**

**Given** cycles exist
**When** I request CSV export (via API or Test UI)
**Then** CSV file is generated containing:
- One row per cycle
- All cycle fields (ID, device, dates, parameters, result)
- Headers in first row

**And** CSV format is:
- Standard CSV (comma-separated)
- UTF-8 encoded
- Suitable for Excel/Google Sheets import

**Prerequisites:** Story 5.1 (list cycles)

**Technical Notes:**
- CSV generation (standard library `encoding/csv`)
- Field selection and formatting
- Reference: PRD Section 3.5

**FR Coverage:** FR-010 (Protokoll-Export - CSV)

---

### Story 5.5: Export Cycle Protocol as JSON

As a **developer**,
I want **to export cycle protocols as JSON**,
So that **I can integrate cycle data with other systems**.

**Acceptance Criteria:**

**Given** cycles exist
**When** I request JSON export (via API)
**Then** JSON response includes:
- Array of cycle objects
- Complete cycle data (same as detail view)
- Proper JSON formatting

**And** JSON format matches API response schema

**Prerequisites:** Story 5.1 (list cycles)

**Technical Notes:**
- JSON encoding (standard library `encoding/json`)
- Response formatting
- Reference: PRD Section 7.2

**FR Coverage:** FR-010 (Protokoll-Export - JSON)

---

### Story 5.6: View Running Cycles

As an **operator**,
I want **to see all currently running cycles**,
So that **I can monitor active sterilization processes**.

**Acceptance Criteria:**

**Given** cycles are running
**When** I request running cycles (via API or Test UI)
**Then** response includes:
- List of all cycles with status "RUNNING"
- Current phase and progress for each
- Estimated completion time
- Device information

**And** list updates in real-time (via WebSocket or polling)

**Prerequisites:** Story 3.3 (status monitoring), Story 5.1 (list cycles)

**Technical Notes:**
- Database query filtering by status
- Real-time updates via WebSocket
- Reference: PRD FR-006

**FR Coverage:** FR-006 (Laufende Zyklen anzeigen)

---

### Story 5.7: Load-to-Cycle Traceability (Growth)

As a **quality assurance officer**,
I want **to associate loads (trays/instruments) with sterilization cycles**,
So that **I can trace which instruments were sterilized in which cycle**.

**Acceptance Criteria:**

**Given** a cycle exists
**When** load information is provided
**Then** load is associated with cycle:
- Load ID/tray ID stored
- Instrument list stored (if available)
- Association recorded in database

**And** traceability queries are supported:
- Find all cycles for a specific load
- Find all loads for a specific cycle
- Export traceability report

**Prerequisites:** Story 5.2 (cycle details), Database schema extension

**Technical Notes:**
- New table: `loads` and `cycle_loads` (junction table)
- Traceability queries
- Reference: PRD FR-011, Section 3.2

**FR Coverage:** FR-011 (Beladungszuordnung)

---

## Epic 6: Test UI for Development and Debugging

**Goal:** Provide a simple web-based interface for testing API endpoints, viewing logs, and inspecting the system without requiring the full Steri-Suite application.

**Value Proposition:** Developers and administrators can test device communication, debug issues, and inspect system state without external tools.

**MVP Scope:** Complete test UI with all core testing capabilities.

---

### Story 6.1: Test UI - Device Management Interface

As a **developer**,
I want **a web interface to view and manage devices**,
So that **I can test device management without Steri-Suite**.

**Acceptance Criteria:**

**Given** Test UI is enabled
**When** I access `http://localhost:8080/test-ui`
**Then** I see device management section showing:
- List of all configured devices
- Device status (online/offline) with visual indicators
- Device details (name, IP, type, manufacturer)

**And** I can:
- Add new device (form with validation)
- Edit device configuration
- Delete device
- Manually trigger device connection/disconnection

**And** interface is simple HTML/CSS/JavaScript (no complex framework)

**Prerequisites:** Story 1.3 (API server), Story 2.1 (device management)

**Technical Notes:**
- Simple HTML template
- JavaScript for API calls
- Basic CSS for styling
- Reference: Architecture "Test UI", PRD FR-025

**FR Coverage:** FR-025 (Test UI - Device Management)

---

### Story 6.2: Test UI - API Endpoint Testing

As a **developer**,
I want **to test API endpoints from the Test UI**,
So that **I can verify API functionality without external tools**.

**Acceptance Criteria:**

**Given** Test UI is accessible
**When** I navigate to API testing section
**Then** I see interface for testing endpoints:
- Endpoint selector (dropdown of available endpoints)
- Request method selector (GET, POST, PUT, DELETE)
- Request body editor (for POST/PUT)
- Send button

**And** when I send a request:
- Request details are displayed
- Response is shown (status code, headers, body)
- Response time is displayed
- Errors are shown clearly

**And** WebSocket testing is available:
- Connect/disconnect button
- Message send interface
- Received messages display

**Prerequisites:** Story 6.1 (Test UI base), Story 1.3 (API)

**Technical Notes:**
- JavaScript fetch API for HTTP requests
- WebSocket client for WebSocket testing
- Response formatting (JSON pretty-print)
- Reference: PRD FR-026

**FR Coverage:** FR-026 (Test UI - API Endpoint Testing)

---

### Story 6.3: Test UI - Cycle Control Testing

As a **developer**,
I want **to start and monitor test cycles from Test UI**,
So that **I can test Melag integration without Steri-Suite**.

**Acceptance Criteria:**

**Given** Melag device is configured
**When** I use Test UI cycle control
**Then** I can:
- Select Melag device
- Start test cycle
- See cycle progress in real-time
- View cycle status and parameters
- Retrieve cycle results

**And** cycle progress updates automatically:
- Status changes displayed
- Phase, temperature, pressure shown
- Progress bar or percentage displayed
- Time remaining shown

**Prerequisites:** Story 6.1 (Test UI base), Story 3.2 (start cycle), Story 3.3 (status)

**Technical Notes:**
- WebSocket connection for real-time updates
- Cycle control API calls
- Status display updates
- Reference: PRD FR-027

**FR Coverage:** FR-027 (Test UI - Cycle Control Testing)

---

### Story 6.4: Test UI - Database Inspection

As a **developer**,
I want **to inspect database contents from Test UI**,
So that **I can verify data storage and debug issues**.

**Acceptance Criteria:**

**Given** Test UI is accessible
**When** I navigate to database inspection section
**Then** I can view:
- Devices table contents
- Cycles table contents
- RDG status table contents
- Audit log entries

**And** I can:
- Filter by device ID, date range, etc.
- Export data as CSV or JSON
- View raw SQL queries (read-only)
- See table schemas

**And** database inspection is read-only (no modifications)

**Prerequisites:** Story 6.1 (Test UI base), Story 1.2 (database)

**Technical Notes:**
- Database query endpoints (read-only)
- Data display in tables
- Export functionality
- Reference: PRD FR-028

**FR Coverage:** FR-028 (Test UI - Database Inspection)

---

### Story 6.5: Test UI - Log Viewing

As a **developer**,
I want **to view application logs from Test UI**,
So that **I can debug issues without accessing log files**.

**Acceptance Criteria:**

**Given** Test UI is accessible
**When** I navigate to logs section
**Then** I see log viewer displaying:
- Recent log entries
- Log level indicators (INFO, ERROR, DEBUG, WARN)
- Timestamps
- Log messages with context

**And** I can:
- Filter by log level
- Search logs by keyword
- Auto-refresh (toggle)
- Export logs
- Clear log display

**Prerequisites:** Story 6.1 (Test UI base), Story 1.4 (logging)

**Technical Notes:**
- Log endpoint (`GET /test-ui/logs`)
- Log streaming or pagination
- Client-side filtering
- Reference: PRD FR-029

**FR Coverage:** FR-029 (Test UI - Logging and Diagnostics)

---

### Story 6.6: Test UI - System Status Display

As a **developer**,
I want **to see system health status in Test UI**,
So that **I can verify system is operating correctly**.

**Acceptance Criteria:**

**Given** Test UI is accessible
**When** I navigate to system status section
**Then** I see:
- System health status (OK/DEGRADED/ERROR)
- Database connection status
- Device connectivity summary (online/offline counts)
- Service uptime
- Memory usage (if available)
- Active WebSocket connections count

**And** status updates automatically (every few seconds)

**Prerequisites:** Story 6.1 (Test UI base), Story 7.1 (health check)

**Technical Notes:**
- Health endpoint integration
- Status display with visual indicators
- Auto-refresh
- Reference: PRD FR-030

**FR Coverage:** FR-030 (Test UI - System Status)

---

### Story 6.7: Test UI - Access Control and Configuration

As a **developer**,
I want **Test UI to be secure and configurable**,
So that **it can be safely used in development and disabled in production**.

**Acceptance Criteria:**

**Given** Test UI is implemented
**When** application starts
**Then** Test UI:
- Is only accessible via localhost (cannot be accessed from network)
- Can be disabled via configuration (`test_ui.enabled = false`)
- Supports optional simple authentication (if enabled)

**And** if Test UI is disabled:
- Routes return 404
- No Test UI resources loaded

**And** Test UI errors do not affect main application functionality

**Prerequisites:** Story 6.1 (Test UI base), Story 1.5 (configuration)

**Technical Notes:**
- Localhost-only binding
- Configuration flag
- Optional authentication
- Error isolation
- Reference: PRD FR-031

**FR Coverage:** FR-031 (Test UI - Access Control)

---

## Epic 7: Health Monitoring and Diagnostics

**Goal:** Provide system health monitoring, metrics, and diagnostic capabilities for operational visibility and troubleshooting.

**Value Proposition:** Administrators can monitor system health, identify issues, and diagnose problems without deep technical knowledge.

**MVP Scope:** Health checks, metrics, and basic diagnostics.

---

### Story 7.1: Health Check Endpoint

As an **administrator**,
I want **to check system health via API**,
So that **I can monitor system availability**.

**Acceptance Criteria:**

**Given** the application is running
**When** I request `GET /api/health`
**Then** response includes:
- Overall status: `OK`, `DEGRADED`, or `ERROR`
- Database connection status
- Device connectivity summary:
  - Total devices
  - Online devices count
  - Offline devices count
- Service uptime in seconds

**And** status determination:
- `OK`: Database connected, all critical devices online
- `DEGRADED`: Database connected, some devices offline
- `ERROR`: Database disconnected or critical failure

**Prerequisites:** Story 1.2 (database), Story 2.4 (device monitoring)

**Technical Notes:**
- Health check logic
- Status aggregation
- Reference: Architecture "API Contracts", PRD FR-032

**FR Coverage:** FR-032 (Health Check API)

---

### Story 7.2: System Metrics Endpoint

As an **administrator**,
I want **to view system metrics**,
So that **I can monitor system performance and usage**.

**Acceptance Criteria:**

**Given** the application is running
**When** I request `GET /api/metrics`
**Then** response includes:
- Service uptime (seconds)
- Active device connections count
- Total cycles processed (all time)
- Cycles processed today
- Total API requests (all time)
- API requests per minute (current rate)
- Database size (if available)

**And** metrics are:
- Updated in real-time
- Formatted as JSON
- Suitable for monitoring dashboards

**Prerequisites:** Story 7.1 (health check)

**Technical Notes:**
- Metrics collection and storage
- Counter management
- Reference: Architecture "API Contracts", PRD FR-033

**FR Coverage:** FR-033 (System Metrics)

---

### Story 7.3: Device Diagnostics Endpoint

As a **technician**,
I want **to diagnose device connection issues**,
So that **I can troubleshoot communication problems**.

**Acceptance Criteria:**

**Given** a device is configured
**When** I request `GET /api/diagnostics/{deviceId}`
**Then** response includes:
- Device connection test results
- Last successful communication timestamp
- Protocol-specific debugging information:
  - Melag: FTP connection status, protocol file access
  - Getinge: ICMP ping history, last successful ping
- Recent error logs for this device
- Connection attempt history

**And** diagnostics help identify:
- Network connectivity issues
- Protocol errors
- Device configuration problems

**Prerequisites:** Story 3.1 (Melag connection), Story 4.1 (Getinge monitoring), Story 1.4 (logging)

**Technical Notes:**
- Diagnostic data collection
- Error log filtering by device
- Connection test execution
- Reference: Architecture "API Contracts", PRD FR-034

**FR Coverage:** FR-034 (Diagnostic Endpoints)

---

## FR Coverage Matrix

| FR | Epic | Story | Status |
|----|------|-------|--------|
| FR-001 | Epic 2 | Story 2.1 | ✅ MVP |
| FR-002 | Epic 2 | Story 2.2 | ✅ MVP |
| FR-003 | Epic 2 | Story 2.3 | ✅ MVP |
| FR-004 | Epic 2 | Story 2.4 | ✅ MVP |
| FR-005 | Epic 5 | Story 5.1 (via Epic 3) | ✅ MVP |
| FR-006 | Epic 5 | Story 5.6 | ✅ MVP |
| FR-007 | Epic 5 | Story 5.1 (via Epic 3) | ✅ MVP |
| FR-008 | Epic 5 | Story 5.1 | ✅ MVP |
| FR-009 | Epic 5 | Story 5.2 | ✅ MVP |
| FR-010 | Epic 5 | Stories 5.3, 5.4, 5.5 | ✅ MVP |
| FR-011 | Epic 5 | Story 5.7 | ⏭️ Growth |
| FR-012 | Epic 1 | Story 1.3 | ✅ MVP |
| FR-013 | Epic 1 | Story 1.3 | ✅ MVP |
| FR-014 | Epic 1 | Story 1.2 | ✅ MVP |
| FR-015 | Epic 1 | Story 1.4 | ✅ MVP |
| FR-016 | Epic 3 | Story 3.1 | ✅ MVP |
| FR-017 | Epic 3 | Story 3.2 | ✅ MVP |
| FR-018 | Epic 3 | Story 3.3 | ✅ MVP |
| FR-019 | Epic 3 | Story 3.4 | ✅ MVP |
| FR-020 | Epic 4 | Story 4.1 | ✅ MVP |
| FR-021 | Epic 4 | Story 4.2 | ⏭️ Growth |
| FR-022 | Epic 1 | Story 1.6 | ✅ MVP |
| FR-023 | Epic 1 | Story 1.7 | ✅ MVP |
| FR-024 | Epic 1 | Story 1.8 | ⏭️ Growth |
| FR-025 | Epic 6 | Story 6.1 | ✅ MVP |
| FR-026 | Epic 6 | Story 6.2 | ✅ MVP |
| FR-027 | Epic 6 | Story 6.3 | ✅ MVP |
| FR-028 | Epic 6 | Story 6.4 | ✅ MVP |
| FR-029 | Epic 6 | Story 6.5 | ✅ MVP |
| FR-030 | Epic 6 | Story 6.6 | ✅ MVP |
| FR-031 | Epic 6 | Story 6.7 | ✅ MVP |
| FR-032 | Epic 7 | Story 7.1 | ✅ MVP |
| FR-033 | Epic 7 | Story 7.2 | ✅ MVP |
| FR-034 | Epic 7 | Story 7.3 | ✅ MVP |
| FR-035 | Epic 1 | Story 1.6 | ✅ MVP |
| FR-036 | Epic 1 | Story 1.8 | ⏭️ Growth |
| FR-037 | Epic 1 | Story 1.3 | ✅ MVP |

**Coverage Summary:**
- ✅ **34 MVP Requirements** covered by 28 MVP Stories
- ⏭️ **3 Growth Requirements** covered by 3 Growth Stories
- **Total:** 37 Requirements → 31 Stories across 7 Epics

---

## Summary

**Epic Breakdown:**

1. **Epic 1: Foundation & Core Infrastructure** (8 stories)
   - Project setup, database, API, logging, authentication
   - Foundation for all other epics

2. **Epic 2: Device Management** (4 stories)
   - Device CRUD operations, health monitoring
   - Enables device configuration and status visibility

3. **Epic 3: Melag Device Integration** (4 stories)
   - Complete Melag integration via MELAnet Box
   - Cycle control, status monitoring, result retrieval

4. **Epic 4: Getinge Device Monitoring** (2 stories)
   - ICMP ping monitoring, failure notifications
   - Basic monitoring until full API available

5. **Epic 5: Cycle Management and Protocols** (7 stories)
   - Cycle listing, details, export (PDF/CSV/JSON)
   - Traceability (Growth)

6. **Epic 6: Test UI for Development** (7 stories)
   - Complete test interface for development/debugging
   - Device management, API testing, cycle control, database inspection, logs, status

7. **Epic 7: Health Monitoring and Diagnostics** (3 stories)
   - Health checks, metrics, diagnostics
   - Operational visibility

**Story Sequencing:**
- Epic 1 establishes foundation (must be first)
- Epic 2 enables device management (prerequisite for device integration)
- Epic 3 and 4 can be parallel (different device types)
- Epic 5 depends on Epic 3 (cycles from Melag)
- Epic 6 can be developed in parallel (independent testing tool)
- Epic 7 can be developed anytime (monitoring)

**Vertical Slicing:**
- Each story delivers complete, testable functionality
- Stories integrate across layers (database + API + device communication)
- Each story leaves system in working state

---

_For implementation: Use the `create-story` workflow to generate individual story implementation plans from this epic breakdown._

_This document incorporates architecture decisions and technical context from the Architecture document._

