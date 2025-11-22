package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"steri-connect-go/internal/logging"
)

// LogEntriesResponse represents log entries with pagination
type LogEntriesResponse struct {
	Entries    []logging.LogEntry `json:"entries"`
	TotalCount int                `json:"total_count"`
	Limit      int                `json:"limit"`
	Offset     int                `json:"offset"`
}

// GetLogsHandler handles GET /api/test-ui/logs
func GetLogsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "method_not_allowed",
			Message: "Only GET method is allowed",
		})
		return
	}

	// Parse query parameters
	limit := 100
	offset := 0
	levelFilter := r.URL.Query().Get("level")
	searchKeyword := r.URL.Query().Get("search")

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 500 {
			limit = l
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Get log entries
	entries, totalCount := logging.GetEntries(levelFilter, searchKeyword, limit, offset)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(LogEntriesResponse{
		Entries:    entries,
		TotalCount: totalCount,
		Limit:      limit,
		Offset:     offset,
	})
}

// ClearLogsHandler handles DELETE /api/test-ui/logs
func ClearLogsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "method_not_allowed",
			Message: "Only DELETE method is allowed",
		})
		return
	}

	logging.ClearBuffer()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Log buffer cleared",
	})
}

