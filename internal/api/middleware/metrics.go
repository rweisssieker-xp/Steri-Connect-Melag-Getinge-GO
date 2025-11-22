package middleware

import (
	"net/http"
	"sync"
	"time"
)

// Metrics tracks API request metrics
type Metrics struct {
	totalRequests    int64
	requestsPerMinute []time.Time
	mu               sync.RWMutex
}

var globalMetrics *Metrics

// InitMetrics initializes the global metrics tracker
func InitMetrics() {
	globalMetrics = &Metrics{
		requestsPerMinute: make([]time.Time, 0, 1000),
	}
}

// GetMetrics returns the global metrics instance
func GetMetrics() *Metrics {
	if globalMetrics == nil {
		InitMetrics()
	}
	return globalMetrics
}

// IncrementRequest increments the request counter
func (m *Metrics) IncrementRequest() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.totalRequests++
	now := time.Now()
	m.requestsPerMinute = append(m.requestsPerMinute, now)
	
	// Keep only requests from the last minute
	oneMinuteAgo := now.Add(-1 * time.Minute)
	validRequests := make([]time.Time, 0)
	for _, t := range m.requestsPerMinute {
		if t.After(oneMinuteAgo) {
			validRequests = append(validRequests, t)
		}
	}
	m.requestsPerMinute = validRequests
}

// GetTotalRequests returns the total number of API requests
func (m *Metrics) GetTotalRequests() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.totalRequests
}

// GetRequestsPerMinute returns the current request rate per minute
func (m *Metrics) GetRequestsPerMinute() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.requestsPerMinute)
}

// MetricsMiddleware tracks API requests
func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip metrics for health check and WebSocket
		if r.URL.Path == "/api/health" || r.URL.Path == "/ws" {
			next.ServeHTTP(w, r)
			return
		}

		// Track API requests
		metrics := GetMetrics()
		metrics.IncrementRequest()

		next.ServeHTTP(w, r)
	})
}

