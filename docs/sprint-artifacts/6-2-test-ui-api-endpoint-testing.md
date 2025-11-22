# Story 6.2: Test UI - API Endpoint Testing

Status: done

## Story

As a **developer**,
I want **to test API endpoints from the Test UI**,
so that **I can verify API functionality without external tools**.

## Acceptance Criteria

1. **Given** Test UI is accessible
   **When** I navigate to API testing section
   **Then** I see interface for testing endpoints:
   - Endpoint selector (dropdown of available endpoints)
   - Request method selector (GET, POST, PUT, DELETE)
   - Request body editor (for POST/PUT)
   - Send button

2. **And** when I send a request:
   - Request details are displayed
   - Response is shown (status code, headers, body)
   - Response time is displayed
   - Errors are shown clearly

3. **And** WebSocket testing is available:
   - Connect/disconnect button
   - Message send interface
   - Received messages display

4. **And** interface is integrated into existing Test UI

## Tasks / Subtasks

- [x] Add API testing section to Test UI (AC: 1, 2, 4)
  - [x] Update `internal/testui/templates/index.html` with API testing section
  - [x] Add endpoint selector dropdown
  - [x] Add request method selector
  - [x] Add request body editor (textarea)
  - [x] Add send button and response display area

- [x] Implement API testing JavaScript (AC: 1, 2)
  - [x] Update `web/testui/js/app.js` with API testing functions
  - [x] Implement endpoint list generation
  - [x] Implement request sending with fetch API
  - [x] Display response (status, headers, body)
  - [x] Display response time
  - [x] Handle errors gracefully

- [x] Add WebSocket testing section (AC: 3, 4)
  - [x] Update HTML template with WebSocket testing UI
  - [x] Add connect/disconnect button
  - [x] Add message input and send button
  - [x] Add received messages display area

- [x] Implement WebSocket testing JavaScript (AC: 3)
  - [x] Update JavaScript with WebSocket client functions
  - [x] Implement connect/disconnect functionality
  - [x] Implement message sending
  - [x] Display received messages with timestamps
  - [x] Handle connection errors

- [x] Update CSS for new sections (AC: 4)
  - [x] Add styles for API testing section
  - [x] Add styles for WebSocket testing section
  - [x] Ensure consistent design with device management section

## Dev Notes

### Relevant Architecture Patterns and Constraints

- **Test UI Structure:** Extend existing Test UI with new sections [Source: Story 6.1]
- **API Endpoints:** Use existing API endpoints from router [Source: internal/api/router.go]
- **WebSocket:** Use existing WebSocket endpoint `/ws` [Source: Story 1.3]
- **JavaScript:** Use fetch API for HTTP requests, WebSocket API for WebSocket [Source: Standard web APIs]

### Source Tree Components to Touch

- `internal/testui/templates/index.html` - Add API and WebSocket testing sections
- `web/testui/js/app.js` - Add API and WebSocket testing functions
- `web/testui/css/style.css` - Add styles for new sections

### Testing Standards Summary

- Manual testing via browser
- Test various API endpoints
- Test WebSocket connection and messaging
- Verify error handling

### Learnings from Previous Story

**From Story 6.1 (Status: done)**
- Test UI structure: Handler, templates, static assets
- WebSocket connection already implemented in device management
- API calls pattern established with fetch API
- CSS styling patterns established

[Source: docs/sprint-artifacts/6-1-test-ui-device-management-interface.md]

**From Story 1.3 (Status: done)**
- WebSocket endpoint: `/ws`
- WebSocket events: device_status_change, cycle_status_update, etc.

[Source: docs/sprint-artifacts/1-3-rest-api-server-and-websocket-setup.md]

### Project Structure Notes

- **Alignment:** Extends existing Test UI structure [Source: Story 6.1]
- **Template:** Update existing `index.html` with new sections
- **JavaScript:** Extend existing `app.js` with new functions
- **CSS:** Extend existing `style.css` with new styles

### References

- **Epic:** Epic 6 - Test UI for Development and Debugging [Source: docs/epics.md#Epic-6]
- **PRD:** FR-026 (Test UI - API Endpoint Testing) [Source: docs/PRD-Steri-Connect-Melag-Getinge-GO.md]

## Dev Agent Record

### Context Reference

<!-- Path(s) to story context XML will be added here by context workflow -->

### Agent Model Used

Auto (Claude Sonnet 4.5 via Cursor)

### Debug Log References

### Completion Notes List

✅ **Story 6.2 Complete - Test UI API Endpoint Testing**

**Implementation Summary:**
- `internal/testui/templates/index.html`: Extended with tab navigation and API/WebSocket testing sections
  - Added tab navigation (Devices, API Testing, WebSocket)
  - API Testing section with endpoint selector, method selector, request body editor, and response display
  - WebSocket Testing section with connect/disconnect, message input, and received messages display
- `web/testui/js/app.js`: Added API and WebSocket testing functionality
  - `showTab()`: Tab switching functionality
  - `setupApiTesting()`: Initialize API testing handlers
  - `handleMethodChange()`: Show/hide request body based on method
  - `handleEndpointChange()`: Handle endpoint selection and auto-set method
  - `sendApiRequest()`: Send API requests with fetch, display response (status, headers, body, timing)
  - `clearApiResponse()`: Clear response display
  - `setupWebSocketTesting()`: Initialize WebSocket testing
  - `toggleWebSocket()`: Connect/disconnect WebSocket for testing
  - `connectWebSocketTest()`: Establish WebSocket connection (separate from device management)
  - `disconnectWebSocket()`: Close WebSocket connection
  - `sendWebSocketMessage()`: Send messages via WebSocket
  - `addWebSocketMessage()`: Display received/sent messages with timestamps
  - `clearWebSocketMessages()`: Clear message history
- `web/testui/css/style.css`: Added styles for tabs, API testing, and WebSocket testing
  - Tab navigation styles
  - API testing form and response display styles
  - WebSocket message display styles with color coding (sent/received/info/error)
  - Consistent design with device management section

**Features Implemented:**
- Tab-based navigation between Devices, API Testing, and WebSocket sections
- API endpoint selector with predefined endpoints and custom endpoint option
- Request method selector (GET, POST, PUT, DELETE)
- Request body editor (shown for POST/PUT, hidden for GET/DELETE)
- Response display with status code, headers, body (JSON formatted), and response time
- Error handling with clear error messages
- WebSocket connection management (separate from device management WebSocket)
- WebSocket message sending and receiving
- Message history with timestamps and color-coded message types
- Placeholder replacement for {id} and {cycle_id} in endpoints

**Verification:**
- Code structure verified: All components properly implemented
- No linter errors: All code passes linting
- Build successful: Application compiles without errors
- Architecture compliance: Extends existing Test UI structure
- All acceptance criteria met

**Files Created/Modified:**
- `internal/testui/templates/index.html` (Modified - added tabs and API/WebSocket sections)
- `web/testui/js/app.js` (Modified - added API and WebSocket testing functions)
- `web/testui/css/style.css` (Modified - added styles for new sections)

