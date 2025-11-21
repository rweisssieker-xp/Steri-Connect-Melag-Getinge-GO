package middleware

import (
	"net/http"
)

// CORSMiddleware handles CORS headers
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow localhost origins by default
		origin := r.Header.Get("Origin")
		
		// Allow localhost and 127.0.0.1 origins
		if origin == "" || 
		   origin == "http://localhost:3000" || 
		   origin == "http://127.0.0.1:3000" ||
		   origin == "http://localhost:8080" ||
		   origin == "http://127.0.0.1:8080" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
		
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "3600")

		// Handle preflight requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

