# Story 7.2: System Metrics Endpoint

Status: done

## Story

As an **administrator**,
I want **to view system metrics**,
so that **I can monitor system performance and usage**.

## Acceptance Criteria

1. **Given** the application is running
   **When** I request `GET /api/metrics`
   **Then** response includes:
   - Service uptime (seconds)
   - Active device connections count
   - Total cycles processed (all time)
   - Cycles processed today
   - Total API requests (all time)
   - API requests per minute (current rate)
   - Database size (if available)

2. **And** metrics are:
   - Updated in real-time
   - Formatted as JSON
   - Suitable for monitoring dashboards

## Tasks / Subtasks

- [x] Create metrics tracking system (AC: 1)
  - [x] Add metrics counters for API requests
  - [x] Add cycle counting
  - [x] Track API request rate
  - [x] Track database size

- [x] Create metrics endpoint handler (AC: 1, 2)
  - [x] Create `internal/api/handlers/metrics.go`
  - [x] Implement GET /api/metrics endpoint
  - [x] Return metrics in JSON format
  - [x] Calculate request rate per minute

- [x] Add metrics route (AC: 1)
  - [x] Add GET /api/metrics route to router
  - [x] Integrate metrics middleware

## Dev Notes

### Relevant Architecture Patterns and Constraints

- **Metrics Collection:** Track metrics in memory with counters
- **API Requests:** Use middleware to track request counts
- **Cycles:** Query database for cycle counts
- **Database Size:** Get database file size if available

### Source Tree Components to Touch

- `internal/api/handlers/metrics.go` - Create new file for metrics handler
- `internal/api/middleware/metrics.go` - Create middleware for request tracking
- `internal/api/router.go` - Add metrics route

### Testing Standards Summary

- Manual testing via API
- Verify metrics accuracy
- Test request rate calculation

### Learnings from Previous Story

**From Story 7.1 (Status: done)**
- Health endpoint pattern established
- Metrics can be added to health endpoint or separate endpoint

[Source: docs/sprint-artifacts/7-1-health-check-endpoint.md]

**From Story 1.3 (Status: done)**
- API server structure established
- Router pattern established

[Source: docs/sprint-artifacts/1-3-rest-api-server-and-websocket-setup.md]

### Project Structure Notes

- **Alignment:** Follows existing API handler pattern
- **Handler Location:** `internal/api/handlers/metrics.go` (new file)
- **Middleware:** `internal/api/middleware/metrics.go` (new file)

### References

- **Epic:** Epic 7 - Health Monitoring and Diagnostics [Source: docs/epics.md#Epic-7]
- **PRD:** FR-033 (System Metrics) [Source: docs/PRD-Steri-Connect-Melag-Getinge-GO.md]

## Dev Agent Record

### Context Reference

<!-- Path(s) to story context XML will be added here by context workflow -->

### Agent Model Used

Auto (Claude Sonnet 4.5 via Cursor)

### Debug Log References

### Completion Notes List

✅ **Story 7.2 Complete - System Metrics Endpoint**

**Implementation Summary:**
- `internal/api/middleware/metrics.go`: Created metrics tracking middleware
  - `Metrics` struct: Tracks total requests and requests per minute
  - `InitMetrics()`: Initializes global metrics instance
  - `GetMetrics()`: Returns global metrics instance
  - `IncrementRequest()`: Increments request counter and tracks per-minute rate
  - `GetTotalRequests()`: Returns total API requests count
  - `GetRequestsPerMinute()`: Returns current request rate (requests in last minute)
  - `MetricsMiddleware()`: HTTP middleware to track API requests (skips health and WebSocket)
- `internal/api/handlers/metrics.go`: Created metrics endpoint handler
  - `MetricsResponse` struct: Response format with all required metrics
  - `MetricsHandler`: GET /api/metrics endpoint handler
  - Calculates uptime in seconds
  - Gets active device connections count
  - Gets total cycles from database
  - Gets cycles processed today (filtered by date range)
  - Gets total API requests from metrics middleware
  - Gets requests per minute from metrics middleware
  - Gets database file size in MB
- `internal/api/router.go`: Added metrics route and middleware
  - Added GET /api/metrics route (no auth required)
  - Applied `MetricsMiddleware` to API handler to track requests
- `cmd/server/main.go`: Initialize metrics on startup
  - Calls `middleware.InitMetrics()` before router setup

**Features Implemented:**
- Service uptime in seconds
- Active device connections count
- Total cycles processed (all time)
- Cycles processed today (filtered by date)
- Total API requests (all time)
- API requests per minute (current rate, sliding window)
- Database size in MB
- Real-time metrics updates
- JSON format suitable for monitoring dashboards

**Verification:**
- Code structure verified: All components properly implemented
- No linter errors: All code passes linting
- Build successful: Application compiles without errors
- Metrics tracking: Request counting and rate calculation working
- All acceptance criteria met

**Files Created/Modified:**
- `internal/api/middleware/metrics.go` (New)
- `internal/api/handlers/metrics.go` (New)
- `internal/api/router.go` (Modified - added metrics route and middleware)
- `cmd/server/main.go` (Modified - initialize metrics)

