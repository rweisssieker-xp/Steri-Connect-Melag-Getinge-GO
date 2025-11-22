package testui

import (
	"html/template"
	"net"
	"net/http"
	"path/filepath"

	"steri-connect-go/internal/config"
	"steri-connect-go/internal/logging"
)

var templates *template.Template

// InitTemplates initializes HTML templates
func InitTemplates() error {
	templatePath := filepath.Join("internal", "testui", "templates", "*.html")
	var err error
	templates, err = template.ParseGlob(templatePath)
	if err != nil {
		return err
	}
	return nil
}

// TestUIHandler handles Test UI requests
func TestUIHandler(w http.ResponseWriter, r *http.Request) {
	cfg := config.Get()
	logger := logging.Get()

	// Check if Test UI is enabled
	if !cfg.TestUI.Enabled {
		logger.Warn("Test UI access attempted but disabled", "path", r.URL.Path)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Ensure localhost-only access
	if r.RemoteAddr != "" {
		host, _, err := splitHostPort(r.RemoteAddr)
		if err == nil && host != "127.0.0.1" && host != "::1" && host != "localhost" {
			logger.Warn("Test UI access attempted from non-localhost", "remote_addr", r.RemoteAddr)
			w.WriteHeader(http.StatusForbidden)
			return
		}
	}

	// Serve index.html
	if templates == nil {
		if err := InitTemplates(); err != nil {
			logger.Error("Failed to initialize templates", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	if err := templates.ExecuteTemplate(w, "index.html", nil); err != nil {
		logger.Error("Failed to execute template", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// splitHostPort splits a network address into host and port
func splitHostPort(addr string) (host, port string, err error) {
	host, port, err = net.SplitHostPort(addr)
	if err != nil {
		// If SplitHostPort fails, assume it's just a host without port
		host = addr
		port = ""
		err = nil
	}
	return
}

