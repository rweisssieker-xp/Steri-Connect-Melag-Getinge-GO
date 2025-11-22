package logging

import (
	"encoding/json"
	"sync"
	"time"
)

// LogEntry represents a single log entry
type LogEntry struct {
	Time    time.Time              `json:"time"`
	Level   string                 `json:"level"`
	Message string                 `json:"message"`
	Fields  map[string]interface{} `json:"fields,omitempty"`
}

// LogBuffer stores recent log entries in memory
type LogBuffer struct {
	entries []LogEntry
	maxSize int
	mu      sync.RWMutex
}

var globalBuffer *LogBuffer

// InitBuffer initializes the global log buffer
func InitBuffer(maxSize int) {
	globalBuffer = &LogBuffer{
		entries: make([]LogEntry, 0, maxSize),
		maxSize: maxSize,
	}
}

// AddEntry adds a log entry to the buffer
func AddEntry(level, message string, fields map[string]interface{}) {
	if globalBuffer == nil {
		return
	}

	entry := LogEntry{
		Time:    time.Now(),
		Level:   level,
		Message: message,
		Fields:  fields,
	}

	globalBuffer.mu.Lock()
	defer globalBuffer.mu.Unlock()

	globalBuffer.entries = append(globalBuffer.entries, entry)

	// Keep only the most recent entries
	if len(globalBuffer.entries) > globalBuffer.maxSize {
		globalBuffer.entries = globalBuffer.entries[len(globalBuffer.entries)-globalBuffer.maxSize:]
	}
}

// GetEntries returns log entries matching the filter criteria
func GetEntries(levelFilter string, searchKeyword string, limit, offset int) ([]LogEntry, int) {
	if globalBuffer == nil {
		return []LogEntry{}, 0
	}

	globalBuffer.mu.RLock()
	defer globalBuffer.mu.RUnlock()

	var filtered []LogEntry

	for _, entry := range globalBuffer.entries {
		// Filter by level
		if levelFilter != "" && entry.Level != levelFilter {
			continue
		}

		// Search by keyword
		if searchKeyword != "" {
			keyword := searchKeyword
			matched := false

			// Check message
			if contains(entry.Message, keyword) {
				matched = true
			}

			// Check fields
			if !matched {
				for _, v := range entry.Fields {
					if containsString(v, keyword) {
						matched = true
						break
					}
				}
			}

			if !matched {
				continue
			}
		}

		filtered = append(filtered, entry)
	}

	totalCount := len(filtered)

	// Apply pagination
	start := offset
	end := offset + limit
	if start > len(filtered) {
		start = len(filtered)
	}
	if end > len(filtered) {
		end = len(filtered)
	}

	if start < end {
		return filtered[start:end], totalCount
	}

	return []LogEntry{}, totalCount
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		containsIgnoreCase(s, substr))
}

func containsIgnoreCase(s, substr string) bool {
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if toLower(s[i+j]) != toLower(substr[j]) {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

func toLower(c byte) byte {
	if c >= 'A' && c <= 'Z' {
		return c + ('a' - 'A')
	}
	return c
}

func containsString(v interface{}, keyword string) bool {
	switch val := v.(type) {
	case string:
		return contains(val, keyword)
	case []byte:
		return contains(string(val), keyword)
	default:
		// Try to convert to string
		if jsonBytes, err := json.Marshal(v); err == nil {
			return contains(string(jsonBytes), keyword)
		}
		return false
	}
}

// Clear clears all log entries
func ClearBuffer() {
	if globalBuffer == nil {
		return
	}

	globalBuffer.mu.Lock()
	defer globalBuffer.mu.Unlock()

	globalBuffer.entries = make([]LogEntry, 0, globalBuffer.maxSize)
}

