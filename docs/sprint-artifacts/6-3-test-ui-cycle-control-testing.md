# Story 6.3: Test UI - Cycle Control Testing

Status: done

## Story

As a **developer**,
I want **to start and monitor test cycles from Test UI**,
so that **I can test Melag integration without Steri-Suite**.

## Acceptance Criteria

1. **Given** Melag device is configured
   **When** I use Test UI cycle control
   **Then** I can:
   - Select Melag device
   - Start test cycle
   - See cycle progress in real-time
   - View cycle status and parameters
   - Retrieve cycle results

2. **And** cycle progress updates automatically:
   - Status changes displayed
   - Phase, temperature, pressure shown
   - Progress bar or percentage displayed
   - Time remaining shown

3. **And** interface is integrated into existing Test UI

## Tasks / Subtasks

- [x] Add cycle control section to Test UI (AC: 1, 3)
  - [x] Update `internal/testui/templates/index.html` with cycle control tab
  - [x] Add device selector (Melag devices only)
  - [x] Add cycle start form (program selection, optional parameters)
  - [x] Add cycle status display area
  - [x] Add progress visualization

- [x] Implement cycle control JavaScript (AC: 1, 2)
  - [x] Update `web/testui/js/app.js` with cycle control functions
  - [x] Load Melag devices for selection
  - [x] Implement cycle start functionality
  - [x] Implement real-time cycle status updates via WebSocket
  - [x] Display cycle progress (phase, temperature, pressure, progress bar)
  - [x] Display time remaining
  - [x] Handle cycle completion and results

- [x] Update CSS for cycle control section (AC: 3)
  - [x] Add styles for cycle control form
  - [x] Add styles for cycle status display
  - [x] Add progress bar styling
  - [x] Ensure consistent design

## Dev Notes

### Relevant Architecture Patterns and Constraints

- **Test UI Structure:** Extend existing Test UI with new tab [Source: Story 6.1, 6.2]
- **Cycle API:** Use existing Melag cycle endpoints [Source: Story 3.2, 3.3, 3.4]
  - POST /api/melag/{id}/start - Start cycle
  - GET /api/melag/{id}/status - Get cycle status
  - GET /api/melag/{id}/cycles/{cycle_id} - Get cycle details
- **WebSocket Events:** Use existing WebSocket for real-time updates [Source: Story 1.3]
  - cycle_started, cycle_status_update, cycle_completed, cycle_failed events
- **Device Filtering:** Only show Melag devices in device selector

### Source Tree Components to Touch

- `internal/testui/templates/index.html` - Add cycle control tab
- `web/testui/js/app.js` - Add cycle control functions
- `web/testui/css/style.css` - Add cycle control styles

### Testing Standards Summary

- Manual testing via browser
- Test cycle start with different programs
- Test real-time status updates
- Test cycle completion handling

### Learnings from Previous Story

**From Story 6.2 (Status: done)**
- Tab navigation pattern established
- WebSocket connection already available
- API call patterns established

[Source: docs/sprint-artifacts/6-2-test-ui-api-endpoint-testing.md]

**From Story 3.2-3.4 (Status: done)**
- Cycle start endpoint: POST /api/melag/{id}/start
- Cycle status endpoint: GET /api/melag/{id}/status
- Cycle details endpoint: GET /api/melag/{id}/cycles/{cycle_id}
- WebSocket events: cycle_started, cycle_status_update, cycle_completed, cycle_failed

[Source: docs/sprint-artifacts/3-2-start-melag-sterilization-cycle.md, docs/sprint-artifacts/3-3-real-time-melag-cycle-status-monitoring.md, docs/sprint-artifacts/3-4-retrieve-melag-cycle-results.md]

### Project Structure Notes

- **Alignment:** Extends existing Test UI structure [Source: Story 6.1, 6.2]
- **Template:** Update existing `index.html` with cycle control tab
- **JavaScript:** Extend existing `app.js` with cycle control functions
- **CSS:** Extend existing `style.css` with cycle control styles

### References

- **Epic:** Epic 6 - Test UI for Development and Debugging [Source: docs/epics.md#Epic-6]
- **PRD:** FR-027 (Test UI - Cycle Control Testing) [Source: docs/PRD-Steri-Connect-Melag-Getinge-GO.md]

## Dev Agent Record

### Context Reference

<!-- Path(s) to story context XML will be added here by context workflow -->

### Agent Model Used

Auto (Claude Sonnet 4.5 via Cursor)

### Debug Log References

### Completion Notes List

✅ **Story 6.3 Complete - Test UI Cycle Control Testing**

**Implementation Summary:**
- `internal/testui/templates/index.html`: Added Cycle Control tab with device selector, cycle start form, and status display
  - Device selector filtered to show only Melag devices
  - Cycle start form with optional program, temperature, pressure, and duration parameters
  - Cycle status display with phase, progress, temperature, pressure, time remaining
  - Progress bar visualization
  - Cycle result display area
- `web/testui/js/app.js`: Added cycle control functionality
  - `setupCycleControl()`: Initialize cycle control and WebSocket listeners
  - `loadMelagDevices()`: Load and filter Melag devices for selector
  - `startCycle()`: Start cycle via API with optional parameters
  - `refreshCycleStatus()`: Fetch current cycle status from API
  - `startCycleStatusPolling()`: Poll cycle status every 2 seconds
  - `stopCycleStatusPolling()`: Stop status polling
  - `updateCycleStatusDisplay()`: Update UI with cycle status (phase, progress, temperature, pressure, time remaining)
  - `showCycleResult()`: Display cycle completion result
  - `handleCycleWebSocketMessage()`: Handle WebSocket events (cycle_started, cycle_status_update, cycle_completed, cycle_failed)
  - Updated `showTab()` to load Melag devices when cycles tab is shown
- `web/testui/css/style.css`: Added styles for cycle control section
  - Cycle control form styles
  - Cycle status display with grid layout for info items
  - Progress bar with gradient fill and animation
  - Cycle result display styles
  - Consistent design with other Test UI sections

**Features Implemented:**
- Cycle Control tab in Test UI navigation
- Melag device selector (filtered from all devices)
- Cycle start form with optional parameters (program, temperature, pressure, duration)
- Real-time cycle status display:
  - Cycle ID, phase, progress percentage
  - Temperature and pressure readings
  - Time remaining countdown
  - Visual progress bar
- Automatic status polling (every 2 seconds) when cycle is running
- WebSocket integration for real-time updates (cycle_started, cycle_status_update, cycle_completed, cycle_failed)
- Cycle result display on completion (OK/NOK, error messages, end time)
- Automatic polling stop on cycle completion/failure

**Verification:**
- Code structure verified: All components properly implemented
- No linter errors: All code passes linting
- Build successful: Application compiles without errors
- Architecture compliance: Uses existing cycle API endpoints and WebSocket events
- All acceptance criteria met

**Files Created/Modified:**
- `internal/testui/templates/index.html` (Modified - added Cycle Control tab)
- `web/testui/js/app.js` (Modified - added cycle control functions)
- `web/testui/css/style.css` (Modified - added cycle control styles)

