// Steri-Connect Test UI - Device Management JavaScript

const API_BASE = '/api';
const WS_URL = `ws://${window.location.host}/ws`;

let ws = null; // WebSocket for device management
let wsTest = null; // WebSocket for testing
let devices = [];

// Initialize on page load
document.addEventListener('DOMContentLoaded', () => {
    loadDevices();
    setupEventListeners();
    connectWebSocket();
    setupApiTesting();
    setupWebSocketTesting();
    setupCycleControl();
});

// Setup event listeners
function setupEventListeners() {
    const addBtn = document.getElementById('addDeviceBtn');
    const modal = document.getElementById('deviceModal');
    const closeBtn = document.querySelector('.close');
    const cancelBtn = document.getElementById('cancelBtn');
    const deviceForm = document.getElementById('deviceForm');

    addBtn.addEventListener('click', () => openModal('Add Device'));
    closeBtn.addEventListener('click', closeModal);
    cancelBtn.addEventListener('click', closeModal);
    
    deviceForm.addEventListener('submit', handleFormSubmit);

    // Close modal when clicking outside
    window.addEventListener('click', (e) => {
        if (e.target === modal) {
            closeModal();
        }
    });

    // API endpoint selector change handler
    const endpointSelect = document.getElementById('apiEndpoint');
    if (endpointSelect) {
        endpointSelect.addEventListener('change', handleEndpointChange);
    }
}

// Tab switching
function showTab(tabName) {
    // Hide all tabs
    document.querySelectorAll('.tab-content').forEach(tab => {
        tab.classList.remove('active');
    });
    document.querySelectorAll('.tab-btn').forEach(btn => {
        btn.classList.remove('active');
    });

    // Show selected tab
    document.getElementById(`${tabName}-tab`).classList.add('active');
    event.target.classList.add('active');

    // Load Melag devices when cycles tab is shown
    if (tabName === 'cycles') {
        loadMelagDevices();
    }
    
    // Load logs when logs tab is shown
    if (tabName === 'logs') {
        loadLogs();
    }
    
    // Load system status when status tab is shown
    if (tabName === 'status') {
        loadSystemStatus();
        if (document.getElementById('statusAutoRefresh').checked) {
            toggleStatusAutoRefresh();
        }
    }
}

// API Testing Functions
function setupApiTesting() {
    const methodSelect = document.getElementById('apiMethod');
    if (methodSelect) {
        methodSelect.addEventListener('change', handleMethodChange);
    }
}

function handleMethodChange() {
    const method = document.getElementById('apiMethod').value;
    const bodyGroup = document.getElementById('requestBodyGroup');
    if (method === 'GET' || method === 'DELETE') {
        bodyGroup.style.display = 'none';
    } else {
        bodyGroup.style.display = 'block';
    }
}

function handleEndpointChange() {
    const endpointSelect = document.getElementById('apiEndpoint');
    const customGroup = document.getElementById('customEndpointGroup');
    const methodSelect = document.getElementById('apiMethod');
    
    if (endpointSelect.value === 'custom') {
        customGroup.style.display = 'block';
    } else {
        customGroup.style.display = 'none';
        // Auto-set method based on endpoint
        const endpoint = endpointSelect.value;
        if (endpoint.startsWith('POST')) {
            methodSelect.value = 'POST';
        } else if (endpoint.startsWith('PUT')) {
            methodSelect.value = 'PUT';
        } else if (endpoint.startsWith('DELETE')) {
            methodSelect.value = 'DELETE';
        } else {
            methodSelect.value = 'GET';
        }
        handleMethodChange();
    }
}

async function sendApiRequest() {
    const method = document.getElementById('apiMethod').value;
    const endpointSelect = document.getElementById('apiEndpoint');
    const customEndpoint = document.getElementById('customEndpoint').value;
    const endpoint = endpointSelect.value === 'custom' ? customEndpoint : endpointSelect.value;
    const bodyText = document.getElementById('requestBody').value;

    if (!endpoint) {
        alert('Please select or enter an endpoint');
        return;
    }

    // Replace placeholders in endpoint
    let finalEndpoint = endpoint;
    if (endpoint.includes('{id}')) {
        const deviceId = prompt('Enter device ID:');
        if (!deviceId) return;
        finalEndpoint = endpoint.replace('{id}', deviceId);
    }
    if (endpoint.includes('{cycle_id}')) {
        const cycleId = prompt('Enter cycle ID:');
        if (!cycleId) return;
        finalEndpoint = finalEndpoint.replace('{cycle_id}', cycleId);
    }

    const startTime = performance.now();
    const responseDiv = document.getElementById('apiResponse');
    const responseStatus = document.getElementById('responseStatus');
    const responseHeaders = document.getElementById('responseHeaders');
    const responseBody = document.getElementById('responseBody');
    const responseTime = document.getElementById('responseTime');

    try {
        const options = {
            method: method,
            headers: {
                'Content-Type': 'application/json',
            },
        };

        if ((method === 'POST' || method === 'PUT') && bodyText) {
            try {
                options.body = JSON.stringify(JSON.parse(bodyText));
            } catch (e) {
                alert('Invalid JSON in request body');
                return;
            }
        }

        const response = await fetch(finalEndpoint, options);
        const endTime = performance.now();
        const duration = (endTime - startTime).toFixed(2);

        // Display response
        responseDiv.style.display = 'block';
        responseTime.textContent = `Response time: ${duration}ms`;

        // Status
        const statusClass = response.ok ? 'status-online' : 'status-error';
        responseStatus.className = `status-badge ${statusClass}`;
        responseStatus.textContent = `${response.status} ${response.statusText}`;

        // Headers
        const headers = {};
        response.headers.forEach((value, key) => {
            headers[key] = value;
        });
        responseHeaders.textContent = JSON.stringify(headers, null, 2);

        // Body
        const contentType = response.headers.get('content-type');
        if (contentType && contentType.includes('application/json')) {
            const json = await response.json();
            responseBody.textContent = JSON.stringify(json, null, 2);
        } else {
            const text = await response.text();
            responseBody.textContent = text;
        }
    } catch (error) {
        const endTime = performance.now();
        const duration = (endTime - startTime).toFixed(2);
        responseDiv.style.display = 'block';
        responseTime.textContent = `Response time: ${duration}ms (Error)`;
        responseStatus.className = 'status-badge status-error';
        responseStatus.textContent = 'Error';
        responseHeaders.textContent = '{}';
        responseBody.textContent = error.message;
    }
}

function clearApiResponse() {
    document.getElementById('apiResponse').style.display = 'none';
}

// WebSocket Testing Functions
function setupWebSocketTesting() {
    // WebSocket testing uses separate connection from device management
}

function toggleWebSocket() {
    if (wsTest && wsTest.readyState === WebSocket.OPEN) {
        disconnectWebSocket();
    } else {
        connectWebSocketTest();
    }
}

function connectWebSocketTest() {
    const statusBadge = document.getElementById('wsStatus');
    const connectBtn = document.getElementById('wsConnectBtn');
    const sendBtn = document.getElementById('wsSendBtn');

    try {
        wsTest = new WebSocket(WS_URL);
        
        wsTest.onopen = () => {
            statusBadge.className = 'status-badge status-online';
            statusBadge.textContent = 'Connected';
            connectBtn.textContent = 'Disconnect';
            sendBtn.disabled = false;
            addWebSocketMessage('System', 'Connected to WebSocket', 'info');
        };
        
        wsTest.onmessage = (event) => {
            try {
                const data = JSON.parse(event.data);
                addWebSocketMessage('Server', JSON.stringify(data, null, 2), 'received');
            } catch (e) {
                addWebSocketMessage('Server', event.data, 'received');
            }
        };
        
        wsTest.onerror = (error) => {
            addWebSocketMessage('System', 'WebSocket error occurred', 'error');
        };
        
        wsTest.onclose = () => {
            statusBadge.className = 'status-badge status-offline';
            statusBadge.textContent = 'Disconnected';
            connectBtn.textContent = 'Connect';
            sendBtn.disabled = true;
            addWebSocketMessage('System', 'Disconnected from WebSocket', 'info');
        };
    } catch (error) {
        addWebSocketMessage('System', 'Failed to connect: ' + error.message, 'error');
    }
}

function disconnectWebSocket() {
    if (wsTest) {
        wsTest.close();
        wsTest = null;
    }
}

function sendWebSocketMessage() {
    const messageInput = document.getElementById('wsMessage');
    const message = messageInput.value.trim();

    if (!message) {
        alert('Please enter a message');
        return;
    }

    if (!wsTest || wsTest.readyState !== WebSocket.OPEN) {
        alert('WebSocket is not connected');
        return;
    }

    try {
        // Try to parse as JSON, if fails send as-is
        let parsed;
        try {
            parsed = JSON.parse(message);
        } catch (e) {
            parsed = message;
        }
        
        wsTest.send(typeof parsed === 'string' ? parsed : JSON.stringify(parsed));
        addWebSocketMessage('You', message, 'sent');
        messageInput.value = '';
    } catch (error) {
        addWebSocketMessage('System', 'Failed to send: ' + error.message, 'error');
    }
}

function addWebSocketMessage(sender, message, type) {
    const messagesDiv = document.getElementById('wsMessages');
    const emptyState = messagesDiv.querySelector('.empty-state');
    if (emptyState) {
        emptyState.remove();
    }

    const messageDiv = document.createElement('div');
    messageDiv.className = `ws-message ws-message-${type}`;
    
    const timestamp = new Date().toLocaleTimeString();
    messageDiv.innerHTML = `
        <div class="ws-message-header">
            <span class="ws-message-sender">${escapeHtml(sender)}</span>
            <span class="ws-message-time">${timestamp}</span>
        </div>
        <pre class="ws-message-content">${escapeHtml(message)}</pre>
    `;
    
    messagesDiv.appendChild(messageDiv);
    messagesDiv.scrollTop = messagesDiv.scrollHeight;
}

function clearWebSocketMessages() {
    const messagesDiv = document.getElementById('wsMessages');
    messagesDiv.innerHTML = '<p class="empty-state">No messages received yet</p>';
}

// Cycle Control Functions
let currentCycleDeviceId = null;
let currentCycleId = null;
let cycleStatusInterval = null;

function setupCycleControl() {
    // Listen to WebSocket events for cycle updates
    if (ws) {
        ws.addEventListener('message', handleCycleWebSocketMessage);
    }
}

function loadMelagDevices() {
    const deviceSelect = document.getElementById('cycleDevice');
    if (!deviceSelect) return;

    deviceSelect.innerHTML = '<option value="">Select Melag device...</option>';

    devices.forEach(device => {
        if (device.manufacturer === 'Melag') {
            const option = document.createElement('option');
            option.value = device.id;
            option.textContent = `${device.name} (${device.ip})`;
            deviceSelect.appendChild(option);
        }
    });
}

async function startCycle() {
    const deviceId = document.getElementById('cycleDevice').value;
    const program = document.getElementById('cycleProgram').value;
    const temperature = document.getElementById('cycleTemperature').value;
    const pressure = document.getElementById('cyclePressure').value;
    const duration = document.getElementById('cycleDuration').value;

    if (!deviceId) {
        alert('Please select a Melag device');
        return;
    }

    const startBtn = document.getElementById('startCycleBtn');
    startBtn.disabled = true;
    startBtn.textContent = 'Starting...';

    try {
        const body = {};
        if (program) body.program = program;
        if (temperature) body.temperature = parseFloat(temperature);
        if (pressure) body.pressure = parseFloat(pressure);
        if (duration) body.duration = parseInt(duration);

        const response = await fetch(`${API_BASE}/melag/${deviceId}/start`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(body),
        });

        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.message || 'Failed to start cycle');
        }

        const result = await response.json();
        currentCycleDeviceId = parseInt(deviceId);
        currentCycleId = result.cycle_id;

        // Show cycle status
        document.getElementById('cycleStatus').style.display = 'block';
        document.getElementById('cycleId').textContent = currentCycleId;
        document.getElementById('cycleStatusBadge').className = 'status-badge status-online';
        document.getElementById('cycleStatusBadge').textContent = 'Running';

        // Start polling for cycle status
        startCycleStatusPolling();

        alert('Cycle started successfully!');
    } catch (error) {
        console.error('Error starting cycle:', error);
        alert('Failed to start cycle: ' + error.message);
    } finally {
        startBtn.disabled = false;
        startBtn.textContent = 'Start Cycle';
    }
}

function startCycleStatusPolling() {
    if (cycleStatusInterval) {
        clearInterval(cycleStatusInterval);
    }

    // Poll immediately
    refreshCycleStatus();

    // Then poll every 2 seconds
    cycleStatusInterval = setInterval(() => {
        refreshCycleStatus();
    }, 2000);
}

function stopCycleStatusPolling() {
    if (cycleStatusInterval) {
        clearInterval(cycleStatusInterval);
        cycleStatusInterval = null;
    }
}

async function refreshCycleStatus() {
    if (!currentCycleDeviceId) return;

    try {
        const response = await fetch(`${API_BASE}/melag/${currentCycleDeviceId}/status`);
        if (!response.ok) {
            throw new Error('Failed to fetch cycle status');
        }

        const status = await response.json();
        updateCycleStatusDisplay(status);
    } catch (error) {
        console.error('Error refreshing cycle status:', error);
    }
}

function updateCycleStatusDisplay(status) {
    const cycleInfo = status.cycle;
    if (!cycleInfo) return;

    // Update phase
    document.getElementById('cyclePhase').textContent = cycleInfo.phase || '-';

    // Update progress
    const progress = cycleInfo.progress_percent || 0;
    document.getElementById('cycleProgress').textContent = `${progress}%`;
    document.getElementById('progressBarFill').style.width = `${progress}%`;
    document.getElementById('progressText').textContent = `${progress}%`;

    // Update temperature
    document.getElementById('cycleTemp').textContent = cycleInfo.temperature 
        ? `${cycleInfo.temperature.toFixed(1)} °C` 
        : '-';

    // Update pressure
    document.getElementById('cyclePressureDisplay').textContent = cycleInfo.pressure 
        ? `${cycleInfo.pressure.toFixed(2)} bar` 
        : '-';

    // Update time remaining
    if (cycleInfo.time_remaining) {
        const minutes = Math.floor(cycleInfo.time_remaining / 60);
        const seconds = Math.floor(cycleInfo.time_remaining % 60);
        document.getElementById('cycleTimeRemaining').textContent = 
            `${minutes}:${seconds.toString().padStart(2, '0')}`;
    } else {
        document.getElementById('cycleTimeRemaining').textContent = '-';
    }

    // Update status badge
    const statusBadge = document.getElementById('cycleStatusBadge');
    if (cycleInfo.phase === 'COMPLETED') {
        statusBadge.className = 'status-badge status-online';
        statusBadge.textContent = 'Completed';
        stopCycleStatusPolling();
        showCycleResult(cycleInfo);
    } else if (cycleInfo.phase === 'FAILED') {
        statusBadge.className = 'status-badge status-error';
        statusBadge.textContent = 'Failed';
        stopCycleStatusPolling();
        showCycleResult(cycleInfo);
    } else {
        statusBadge.className = 'status-badge status-online';
        statusBadge.textContent = 'Running';
    }
}

function showCycleResult(cycleInfo) {
    const resultDiv = document.getElementById('cycleResult');
    const resultContent = document.getElementById('cycleResultContent');
    
    resultDiv.style.display = 'block';
    
    let html = `<p><strong>Result:</strong> ${cycleInfo.result || 'N/A'}</p>`;
    if (cycleInfo.error_description) {
        html += `<p><strong>Error:</strong> ${escapeHtml(cycleInfo.error_description)}</p>`;
    }
    if (cycleInfo.end_ts) {
        html += `<p><strong>End Time:</strong> ${new Date(cycleInfo.end_ts).toLocaleString()}</p>`;
    }
    
    resultContent.innerHTML = html;
}

function handleCycleWebSocketMessage(event) {
    try {
        const data = JSON.parse(event.data);
        
        if (data.type === 'cycle_started' && data.device_id === currentCycleDeviceId) {
            currentCycleId = data.cycle_id;
            document.getElementById('cycleStatus').style.display = 'block';
            document.getElementById('cycleId').textContent = currentCycleId;
            startCycleStatusPolling();
        } else if (data.type === 'cycle_status_update' && data.device_id === currentCycleDeviceId) {
            if (data.cycle) {
                updateCycleStatusDisplay({ cycle: data.cycle });
            }
        } else if (data.type === 'cycle_completed' && data.device_id === currentCycleDeviceId) {
            stopCycleStatusPolling();
            refreshCycleStatus();
        } else if (data.type === 'cycle_failed' && data.device_id === currentCycleDeviceId) {
            stopCycleStatusPolling();
            refreshCycleStatus();
        }
    } catch (e) {
        // Not a cycle-related message, ignore
    }
}

// Database Inspection Functions
let currentTableData = null;
let currentTableColumns = null;
let dbCurrentPage = 0;
let dbPageSize = 50;

async function loadTableData() {
    const tableName = document.getElementById('dbTable').value;
    if (!tableName) {
        alert('Please select a table');
        return;
    }

    const deviceId = document.getElementById('dbDeviceId').value;
    const startDate = document.getElementById('dbStartDate').value;
    const endDate = document.getElementById('dbEndDate').value;

    const params = new URLSearchParams({
        limit: dbPageSize.toString(),
        offset: (dbCurrentPage * dbPageSize).toString(),
    });

    if (deviceId) params.append('device_id', deviceId);
    if (startDate) params.append('start_date', startDate);
    if (endDate) params.append('end_date', endDate);

    try {
        const response = await fetch(`${API_BASE}/test-ui/db/tables/${tableName}?${params}`);
        if (!response.ok) {
            throw new Error('Failed to load table data');
        }

        const data = await response.json();
        currentTableData = data.rows;
        currentTableColumns = data.columns;

        displayTableData(data);
        updateDbPagination(data.total_count);
    } catch (error) {
        console.error('Error loading table data:', error);
        alert('Failed to load table data: ' + error.message);
    }
}

function displayTableData(data) {
    const tableDiv = document.getElementById('dbDataTable');
    
    if (!data.rows || data.rows.length === 0) {
        tableDiv.innerHTML = '<p class="empty-state">No data found</p>';
        return;
    }

    let html = '<table class="data-table-content"><thead><tr>';
    data.columns.forEach(col => {
        html += `<th>${escapeHtml(col)}</th>`;
    });
    html += '</tr></thead><tbody>';

    data.rows.forEach(row => {
        html += '<tr>';
        data.columns.forEach(col => {
            const val = row[col];
            html += `<td>${val === null ? '<em>null</em>' : escapeHtml(String(val))}</td>`;
        });
        html += '</tr>';
    });

    html += '</tbody></table>';
    tableDiv.innerHTML = html;
}

function updateDbPagination(totalCount) {
    const paginationDiv = document.getElementById('dbPagination');
    const totalPages = Math.ceil(totalCount / dbPageSize);
    
    let html = `<div class="pagination-info">Showing ${dbCurrentPage * dbPageSize + 1}-${Math.min((dbCurrentPage + 1) * dbPageSize, totalCount)} of ${totalCount}</div>`;
    html += '<div class="pagination-controls">';
    
    if (dbCurrentPage > 0) {
        html += `<button class="btn btn-secondary btn-small" onclick="dbPreviousPage()">Previous</button>`;
    }
    
    html += `<span>Page ${dbCurrentPage + 1} of ${totalPages}</span>`;
    
    if (dbCurrentPage < totalPages - 1) {
        html += `<button class="btn btn-secondary btn-small" onclick="dbNextPage()">Next</button>`;
    }
    
    html += '</div>';
    paginationDiv.innerHTML = html;
}

function dbPreviousPage() {
    if (dbCurrentPage > 0) {
        dbCurrentPage--;
        loadTableData();
    }
}

function dbNextPage() {
    dbCurrentPage++;
    loadTableData();
}

async function loadTableSchema() {
    const tableName = document.getElementById('dbTable').value;
    if (!tableName) {
        alert('Please select a table');
        return;
    }

    try {
        const response = await fetch(`${API_BASE}/test-ui/db/tables/${tableName}/schema`);
        if (!response.ok) {
            throw new Error('Failed to load schema');
        }

        const schema = await response.json();
        const schemaDiv = document.getElementById('dbSchema');
        const schemaContent = document.getElementById('dbSchemaContent');
        
        schemaDiv.style.display = 'block';
        schemaContent.textContent = JSON.stringify(schema, null, 2);
    } catch (error) {
        console.error('Error loading schema:', error);
        alert('Failed to load schema: ' + error.message);
    }
}

function exportTableData(format) {
    const tableName = document.getElementById('dbTable').value;
    if (!tableName) {
        alert('Please select a table');
        return;
    }

    const deviceId = document.getElementById('dbDeviceId').value;
    const startDate = document.getElementById('dbStartDate').value;
    const endDate = document.getElementById('dbEndDate').value;

    const params = new URLSearchParams({ format });
    if (deviceId) params.append('device_id', deviceId);
    if (startDate) params.append('start_date', startDate);
    if (endDate) params.append('end_date', endDate);

    window.open(`${API_BASE}/test-ui/db/tables/${tableName}/export?${params}`, '_blank');
}

// Log Viewing Functions
let logAutoRefreshInterval = null;
let logCurrentPage = 0;
let logPageSize = 50;
let logDebounceTimer = null;

function debounceLoadLogs() {
    if (logDebounceTimer) {
        clearTimeout(logDebounceTimer);
    }
    logDebounceTimer = setTimeout(() => {
        logCurrentPage = 0;
        loadLogs();
    }, 500);
}

async function loadLogs() {
    const levelFilter = document.getElementById('logLevelFilter').value;
    const searchKeyword = document.getElementById('logSearch').value;

    const params = new URLSearchParams({
        limit: logPageSize.toString(),
        offset: (logCurrentPage * logPageSize).toString(),
    });

    if (levelFilter) params.append('level', levelFilter);
    if (searchKeyword) params.append('search', searchKeyword);

    try {
        const response = await fetch(`${API_BASE}/test-ui/logs?${params}`);
        if (!response.ok) {
            throw new Error('Failed to load logs');
        }

        const data = await response.json();
        displayLogs(data.entries);
        updateLogsPagination(data.total_count);
    } catch (error) {
        console.error('Error loading logs:', error);
        document.getElementById('logsList').innerHTML = `<div class="error">Failed to load logs: ${escapeHtml(error.message)}</div>`;
    }
}

function displayLogs(entries) {
    const logsList = document.getElementById('logsList');
    
    if (!entries || entries.length === 0) {
        logsList.innerHTML = '<p class="empty-state">No logs found</p>';
        return;
    }

    let html = '';
    entries.forEach(entry => {
        const levelClass = `log-level-${entry.Level.toLowerCase()}`;
        const timeStr = new Date(entry.Time).toLocaleString();
        
        html += `<div class="log-entry ${levelClass}">`;
        html += `<div class="log-header">`;
        html += `<span class="log-level">${escapeHtml(entry.Level)}</span>`;
        html += `<span class="log-time">${escapeHtml(timeStr)}</span>`;
        html += `</div>`;
        html += `<div class="log-message">${escapeHtml(entry.Message)}</div>`;
        
        if (entry.Fields && Object.keys(entry.Fields).length > 0) {
            html += `<div class="log-fields">`;
            for (const [key, value] of Object.entries(entry.Fields)) {
                html += `<span class="log-field"><strong>${escapeHtml(key)}:</strong> ${escapeHtml(String(value))}</span>`;
            }
            html += `</div>`;
        }
        
        html += `</div>`;
    });

    logsList.innerHTML = html;
    logsList.scrollTop = 0;
}

function updateLogsPagination(totalCount) {
    const paginationDiv = document.getElementById('logsPagination');
    const totalPages = Math.ceil(totalCount / logPageSize);
    
    let html = `<div class="pagination-info">Showing ${logCurrentPage * logPageSize + 1}-${Math.min((logCurrentPage + 1) * logPageSize, totalCount)} of ${totalCount}</div>`;
    html += '<div class="pagination-controls">';
    
    if (logCurrentPage > 0) {
        html += `<button class="btn btn-secondary btn-small" onclick="logsPreviousPage()">Previous</button>`;
    }
    
    html += `<span>Page ${logCurrentPage + 1} of ${totalPages}</span>`;
    
    if (logCurrentPage < totalPages - 1) {
        html += `<button class="btn btn-secondary btn-small" onclick="logsNextPage()">Next</button>`;
    }
    
    html += '</div>';
    paginationDiv.innerHTML = html;
}

function logsPreviousPage() {
    if (logCurrentPage > 0) {
        logCurrentPage--;
        loadLogs();
    }
}

function logsNextPage() {
    logCurrentPage++;
    loadLogs();
}

function toggleAutoRefresh() {
    const autoRefresh = document.getElementById('logAutoRefresh').checked;
    
    if (autoRefresh) {
        logAutoRefreshInterval = setInterval(() => {
            loadLogs();
        }, 2000); // Refresh every 2 seconds
    } else {
        if (logAutoRefreshInterval) {
            clearInterval(logAutoRefreshInterval);
            logAutoRefreshInterval = null;
        }
    }
}

async function clearLogs() {
    if (!confirm('Are you sure you want to clear the log buffer?')) {
        return;
    }

    try {
        const response = await fetch(`${API_BASE}/test-ui/logs`, {
            method: 'DELETE',
        });

        if (!response.ok) {
            throw new Error('Failed to clear logs');
        }

        loadLogs();
    } catch (error) {
        console.error('Error clearing logs:', error);
        alert('Failed to clear logs: ' + error.message);
    }
}

function exportLogs() {
    const levelFilter = document.getElementById('logLevelFilter').value;
    const searchKeyword = document.getElementById('logSearch').value;

    const params = new URLSearchParams({
        limit: '10000', // Export all logs
        offset: '0',
    });

    if (levelFilter) params.append('level', levelFilter);
    if (searchKeyword) params.append('search', searchKeyword);

    window.open(`${API_BASE}/test-ui/logs?${params}`, '_blank');
}

// System Status Functions
let statusAutoRefreshInterval = null;

async function loadSystemStatus() {
    try {
        const response = await fetch(`${API_BASE}/health`);
        if (!response.ok) {
            throw new Error('Failed to load system status');
        }

        const status = await response.json();
        displaySystemStatus(status);
    } catch (error) {
        console.error('Error loading system status:', error);
        document.getElementById('overallStatusBadge').textContent = 'ERROR';
        document.getElementById('overallStatusBadge').className = 'status-badge status-error';
    }
}

function displaySystemStatus(status) {
    // Overall status
    const overallBadge = document.getElementById('overallStatusBadge');
    overallBadge.textContent = status.status.toUpperCase();
    overallBadge.className = `status-badge status-${status.status.toLowerCase()}`;
    
    document.getElementById('systemUptime').textContent = status.uptime || '-';
    document.getElementById('systemVersion').textContent = status.version || '-';

    // Database status
    const dbBadge = document.getElementById('dbStatusBadge');
    if (status.database && status.database.connected) {
        dbBadge.textContent = 'CONNECTED';
        dbBadge.className = 'status-badge status-online';
        document.getElementById('dbStatus').textContent = 'Connected';
    } else {
        dbBadge.textContent = 'DISCONNECTED';
        dbBadge.className = 'status-badge status-error';
        document.getElementById('dbStatus').textContent = 'Disconnected';
    }

    // Device status
    if (status.devices) {
        document.getElementById('deviceTotal').textContent = status.devices.total || 0;
        document.getElementById('deviceOnline').textContent = status.devices.online || 0;
        document.getElementById('deviceOffline').textContent = status.devices.offline || 0;
    }

    // WebSocket status
    if (status.websocket) {
        document.getElementById('wsConnections').textContent = status.websocket.connections || 0;
    }

    // Memory status
    if (status.memory) {
        document.getElementById('memoryAlloc').textContent = `${status.memory.alloc_mb.toFixed(2)} MB`;
        document.getElementById('memoryTotalAlloc').textContent = `${status.memory.total_alloc_mb.toFixed(2)} MB`;
        document.getElementById('memorySys').textContent = `${status.memory.sys_mb.toFixed(2)} MB`;
        document.getElementById('memoryGC').textContent = status.memory.num_gc || 0;
    }
}

function toggleStatusAutoRefresh() {
    const autoRefresh = document.getElementById('statusAutoRefresh').checked;
    
    if (autoRefresh) {
        statusAutoRefreshInterval = setInterval(() => {
            loadSystemStatus();
        }, 3000); // Refresh every 3 seconds
    } else {
        if (statusAutoRefreshInterval) {
            clearInterval(statusAutoRefreshInterval);
            statusAutoRefreshInterval = null;
        }
    }
}

// Load devices from API
async function loadDevices() {
    try {
        const response = await fetch(`${API_BASE}/devices`);
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        devices = await response.json();
        renderDevices();
    } catch (error) {
        console.error('Error loading devices:', error);
        showError('Failed to load devices: ' + error.message);
    }
}

// Render devices list
function renderDevices() {
    const deviceList = document.getElementById('deviceList');
    
    if (devices.length === 0) {
        deviceList.innerHTML = `
            <div class="empty-state">
                <p>No devices configured</p>
                <p>Click "Add Device" to get started</p>
            </div>
        `;
        return;
    }

    deviceList.innerHTML = devices.map(device => `
        <div class="device-card" data-device-id="${device.id}">
            <div class="device-info">
                <div class="device-name">${escapeHtml(device.name)}</div>
                <div class="device-details">
                    <span class="device-detail-item">
                        <strong>IP:</strong> ${escapeHtml(device.ip)}
                    </span>
                    <span class="device-detail-item">
                        <strong>Manufacturer:</strong> ${escapeHtml(device.manufacturer)}
                    </span>
                    <span class="device-detail-item">
                        <strong>Type:</strong> ${escapeHtml(device.type)}
                    </span>
                    ${device.model ? `<span class="device-detail-item"><strong>Model:</strong> ${escapeHtml(device.model)}</span>` : ''}
                    ${device.location ? `<span class="device-detail-item"><strong>Location:</strong> ${escapeHtml(device.location)}</span>` : ''}
                    <span class="device-detail-item">
                        <span class="status-badge status-${getStatusClass(device)}">${getStatusText(device)}</span>
                    </span>
                </div>
            </div>
            <div class="device-actions">
                <button class="btn btn-primary btn-small" onclick="editDevice(${device.id})">Edit</button>
                <button class="btn btn-secondary btn-small" onclick="toggleConnection(${device.id})">${device.connected ? 'Disconnect' : 'Connect'}</button>
                <button class="btn btn-danger btn-small" onclick="deleteDevice(${device.id})">Delete</button>
            </div>
        </div>
    `).join('');
}

// Get status class for device
function getStatusClass(device) {
    if (device.connected) {
        return 'online';
    }
    if (device.connection_state === 'CONNECTING') {
        return 'connecting';
    }
    if (device.connection_state === 'ERROR') {
        return 'error';
    }
    return 'offline';
}

// Get status text for device
function getStatusText(device) {
    if (device.connected) {
        return 'Online';
    }
    if (device.connection_state) {
        return device.connection_state.toLowerCase();
    }
    return 'Offline';
}

// Open modal for add/edit
function openModal(title, device = null) {
    const modal = document.getElementById('deviceModal');
    const modalTitle = document.getElementById('modalTitle');
    const form = document.getElementById('deviceForm');
    
    modalTitle.textContent = title;
    form.reset();
    
    if (device) {
        document.getElementById('deviceId').value = device.id;
        document.getElementById('deviceName').value = device.name || '';
        document.getElementById('deviceIP').value = device.ip || '';
        document.getElementById('deviceManufacturer').value = device.manufacturer || '';
        document.getElementById('deviceType').value = device.type || '';
        document.getElementById('deviceModel').value = device.model || '';
        document.getElementById('deviceSerial').value = device.serial || '';
        document.getElementById('deviceLocation').value = device.location || '';
    } else {
        document.getElementById('deviceId').value = '';
    }
    
    modal.style.display = 'block';
}

// Close modal
function closeModal() {
    const modal = document.getElementById('deviceModal');
    modal.style.display = 'none';
}

// Handle form submission
async function handleFormSubmit(e) {
    e.preventDefault();
    
    const formData = {
        name: document.getElementById('deviceName').value,
        ip: document.getElementById('deviceIP').value,
        manufacturer: document.getElementById('deviceManufacturer').value,
        type: document.getElementById('deviceType').value,
        model: document.getElementById('deviceModel').value,
        serial: document.getElementById('deviceSerial').value,
        location: document.getElementById('deviceLocation').value,
    };
    
    const deviceId = document.getElementById('deviceId').value;
    const url = deviceId ? `${API_BASE}/devices/${deviceId}` : `${API_BASE}/devices`;
    const method = deviceId ? 'PUT' : 'POST';
    
    try {
        const response = await fetch(url, {
            method: method,
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(formData),
        });
        
        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.message || 'Failed to save device');
        }
        
        const result = await response.json();
        showSuccess(deviceId ? 'Device updated successfully' : 'Device added successfully');
        closeModal();
        loadDevices();
    } catch (error) {
        console.error('Error saving device:', error);
        showError('Failed to save device: ' + error.message);
    }
}

// Edit device
async function editDevice(id) {
    const device = devices.find(d => d.id === id);
    if (!device) {
        showError('Device not found');
        return;
    }
    
    // Fetch full device details including status
    try {
        const response = await fetch(`${API_BASE}/devices/${id}`);
        if (!response.ok) {
            throw new Error('Failed to fetch device details');
        }
        const deviceDetails = await response.json();
        openModal('Edit Device', deviceDetails);
    } catch (error) {
        console.error('Error fetching device:', error);
        showError('Failed to load device details');
    }
}

// Delete device
async function deleteDevice(id) {
    if (!confirm('Are you sure you want to delete this device?')) {
        return;
    }
    
    try {
        const response = await fetch(`${API_BASE}/devices/${id}`, {
            method: 'DELETE',
        });
        
        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.message || 'Failed to delete device');
        }
        
        showSuccess('Device deleted successfully');
        loadDevices();
    } catch (error) {
        console.error('Error deleting device:', error);
        showError('Failed to delete device: ' + error.message);
    }
}

// Toggle device connection (placeholder - requires API endpoint)
async function toggleConnection(id) {
    // This would require a new API endpoint for connect/disconnect
    // For now, just show a message
    alert('Connect/Disconnect functionality requires API endpoint implementation');
}

// Connect to WebSocket for real-time updates
function connectWebSocket() {
    try {
        ws = new WebSocket(WS_URL);
        
        ws.onopen = () => {
            console.log('WebSocket connected');
        };
        
        ws.onmessage = (event) => {
            const data = JSON.parse(event.data);
            if (data.type === 'device_status_change') {
                // Reload devices when status changes
                loadDevices();
            }
        };
        
        ws.onerror = (error) => {
            console.error('WebSocket error:', error);
        };
        
        ws.onclose = () => {
            console.log('WebSocket disconnected, reconnecting...');
            setTimeout(connectWebSocket, 3000);
        };
    } catch (error) {
        console.error('Failed to connect WebSocket:', error);
    }
}

// Utility functions
function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

function showError(message) {
    const deviceList = document.getElementById('deviceList');
    deviceList.innerHTML = `<div class="error">${escapeHtml(message)}</div>`;
    setTimeout(() => loadDevices(), 3000);
}

function showSuccess(message) {
    const deviceList = document.getElementById('deviceList');
    const successDiv = document.createElement('div');
    successDiv.className = 'success';
    successDiv.textContent = message;
    deviceList.insertBefore(successDiv, deviceList.firstChild);
    setTimeout(() => successDiv.remove(), 3000);
}

