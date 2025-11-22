# Story 6.5: Test UI - Log Viewing

Status: done

## Story

As a **developer**,
I want **to view application logs from Test UI**,
so that **I can debug issues without accessing log files**.

## Acceptance Criteria

1. **Given** Test UI is accessible
   **When** I navigate to logs section
   **Then** I see log viewer displaying:
   - Recent log entries
   - Log level indicators (INFO, ERROR, DEBUG, WARN)
   - Timestamps
   - Log messages with context

2. **And** I can:
   - Filter by log level
   - Search logs by keyword
   - Auto-refresh (toggle)
   - Export logs
   - Clear log display

3. **And** interface is integrated into existing Test UI

## Tasks / Subtasks

- [x] Create log viewing API handler (AC: 1, 2)
  - [x] Create GET /api/test-ui/logs endpoint
  - [x] Support filtering by log level
  - [x] Support search by keyword
  - [x] Support pagination
  - [x] Return log entries in JSON format

- [x] Add log viewing section to Test UI (AC: 1, 3)
  - [x] Update `internal/testui/templates/index.html` with log viewing tab
  - [x] Add log level filter dropdown
  - [x] Add search input
  - [x] Add auto-refresh toggle
  - [x] Add export and clear buttons
  - [x] Add log display area

- [x] Implement log viewing JavaScript (AC: 1, 2)
  - [x] Update `web/testui/js/app.js` with log viewing functions
  - [x] Load logs from API
  - [x] Display logs with level indicators
  - [x] Implement filtering
  - [x] Implement search
  - [x] Implement auto-refresh
  - [x] Implement export functionality

- [x] Update CSS for log viewing section (AC: 3)
  - [x] Add styles for log viewer
  - [x] Add log level color coding
  - [x] Ensure consistent design

## Dev Notes

### Relevant Architecture Patterns and Constraints

- **Test UI Structure:** Extend existing Test UI with new tab [Source: Story 6.1-6.4]
- **Logging:** Use existing logging system [Source: Story 1.4]
- **Log Format:** Structured logging with JSON format [Source: Story 1.4]
- **Log Storage:** Logs may be in file or stdout [Source: Story 1.4]

### Source Tree Components to Touch

- `internal/api/handlers/logs.go` - Create new file for log viewing handler
- `internal/api/router.go` - Add log viewing route
- `internal/testui/templates/index.html` - Add log viewing tab
- `web/testui/js/app.js` - Add log viewing functions
- `web/testui/css/style.css` - Add log viewing styles

### Testing Standards Summary

- Manual testing via browser
- Test log filtering
- Test search functionality
- Test auto-refresh
- Test export functionality

### Learnings from Previous Story

**From Story 6.1-6.4 (Status: done)**
- Tab navigation pattern established
- Test UI structure and styling patterns

[Source: docs/sprint-artifacts/6-1-test-ui-device-management-interface.md through 6-4-test-ui-database-inspection.md]

**From Story 1.4 (Status: done)**
- Structured logging with log/slog
- Log levels: DEBUG, INFO, WARN, ERROR
- Log format: JSON or text
- Log output: stdout or file

[Source: docs/sprint-artifacts/1-4-structured-logging-and-audit-trail.md]

### Project Structure Notes

- **Alignment:** Extends existing Test UI structure [Source: Story 6.1-6.4]
- **Handler Location:** `internal/api/handlers/logs.go` (new file)
- **Template:** Update existing `index.html` with log viewing tab
- **JavaScript:** Extend existing `app.js` with log viewing functions
- **CSS:** Extend existing `style.css` with log viewing styles

### References

- **Epic:** Epic 6 - Test UI for Development and Debugging [Source: docs/epics.md#Epic-6]
- **PRD:** FR-029 (Test UI - Logging and Diagnostics) [Source: docs/PRD-Steri-Connect-Melag-Getinge-GO.md]

## Dev Agent Record

### Context Reference

<!-- Path(s) to story context XML will be added here by context workflow -->

### Agent Model Used

Auto (Claude Sonnet 4.5 via Cursor)

### Debug Log References

### Completion Notes List

✅ **Story 6.5 Complete - Test UI Log Viewing**

**Implementation Summary:**
- `internal/logging/buffer.go`: Created in-memory log buffer for storing recent log entries
  - `LogEntry` struct: Represents a single log entry with time, level, message, and fields
  - `LogBuffer` struct: Stores log entries with max size limit
  - `InitBuffer()`: Initializes global log buffer
  - `AddEntry()`: Adds log entry to buffer
  - `GetEntries()`: Retrieves log entries with filtering (level, search keyword) and pagination
  - `ClearBuffer()`: Clears all log entries
- `internal/logging/buffered_handler.go`: Created wrapper handler for slog
  - `BufferedHandler`: Wraps slog handler and writes to both original handler and buffer
  - Ensures all log entries are captured in buffer for Test UI viewing
- `internal/logging/logger.go`: Updated to use buffered handler
  - Wrapped base handler with `BufferedHandler` to capture logs
- `cmd/server/main.go`: Initialize log buffer on startup
  - Calls `logging.InitBuffer(1000)` to keep last 1000 log entries
- `internal/api/handlers/logs.go`: Created log viewing API handlers
  - `GetLogsHandler`: GET /api/test-ui/logs with filtering and pagination
  - `ClearLogsHandler`: DELETE /api/test-ui/logs to clear buffer
  - Supports level filter, search keyword, limit, and offset parameters
- `internal/api/router.go`: Added log viewing routes
  - GET /api/test-ui/logs (get logs)
  - DELETE /api/test-ui/logs (clear logs)
- `internal/testui/templates/index.html`: Added Log Viewer tab
  - Log level filter dropdown
  - Search input with debouncing
  - Auto-refresh toggle checkbox
  - Export and Clear buttons
  - Log display area
- `web/testui/js/app.js`: Added log viewing functions
  - `loadLogs()`: Load logs from API with filters
  - `displayLogs()`: Display logs with level indicators and color coding
  - `updateLogsPagination()`: Update pagination controls
  - `logsPreviousPage()` / `logsNextPage()`: Pagination navigation
  - `toggleAutoRefresh()`: Enable/disable auto-refresh (every 2 seconds)
  - `clearLogs()`: Clear log buffer
  - `exportLogs()`: Export logs (opens in new window)
  - `debounceLoadLogs()`: Debounce search input
  - Updated `showTab()` to load logs when logs tab is shown
- `web/testui/css/style.css`: Added styles for log viewer
  - Dark theme for log display (terminal-like appearance)
  - Color-coded log levels (DEBUG: blue, INFO: cyan, WARN: yellow, ERROR: red)
  - Log entry styling with left border color coding
  - Log header with level badge and timestamp
  - Log fields display
  - Consistent design with other sections

**Features Implemented:**
- Log Viewer tab in Test UI navigation
- In-memory log buffer storing last 1000 entries
- Log level filtering (DEBUG, INFO, WARN, ERROR, or All)
- Search functionality with keyword matching (case-insensitive)
- Auto-refresh toggle (refreshes every 2 seconds when enabled)
- Pagination with previous/next page navigation
- Export functionality (opens logs in new window)
- Clear log buffer functionality
- Color-coded log levels for easy visual identification
- Dark theme terminal-like display
- Debounced search input for better performance

**Verification:**
- Code structure verified: All components properly implemented
- No linter errors: All code passes linting
- Build successful: Application compiles without errors
- Log buffer: Captures all log entries via buffered handler
- All acceptance criteria met

**Files Created/Modified:**
- `internal/logging/buffer.go` (New)
- `internal/logging/buffered_handler.go` (New)
- `internal/logging/logger.go` (Modified - use buffered handler)
- `cmd/server/main.go` (Modified - initialize log buffer)
- `internal/api/handlers/logs.go` (New)
- `internal/api/router.go` (Modified - added log routes)
- `internal/testui/templates/index.html` (Modified - added Logs tab)
- `web/testui/js/app.js` (Modified - added log functions)
- `web/testui/css/style.css` (Modified - added log styles)

