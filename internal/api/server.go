package api

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"steri-connect-go/internal/logging"
)

// Server represents the HTTP server
type Server struct {
	httpServer *http.Server
	port       int
	bindAddr   string
}

// NewServer creates a new HTTP server instance
func NewServer(port int, bindAddr string, handler http.Handler) *Server {
	return &Server{
		port:     port,
		bindAddr: bindAddr,
		httpServer: &http.Server{
			Addr:         fmt.Sprintf("%s:%d", bindAddr, port),
			Handler:      handler,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	addr := s.httpServer.Addr
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	logger := logging.Get()
	
	// Log security information based on bind address
	if bindAddr == "0.0.0.0" {
		logger.Info("HTTP server starting on all network interfaces",
			"address", fmt.Sprintf("http://%s", addr),
			"api_path", fmt.Sprintf("http://%s/api/", addr),
			"websocket_path", fmt.Sprintf("ws://%s/ws", addr),
			"security_note", "Server accessible from network - ensure API key authentication is enabled")
	} else {
		logger.Info("HTTP server starting on localhost only",
			"address", fmt.Sprintf("http://%s", addr),
			"api_path", fmt.Sprintf("http://%s/api/", addr),
			"websocket_path", fmt.Sprintf("ws://%s/ws", addr),
			"security_note", "Server accessible only from localhost")
	}

	if err := s.httpServer.Serve(listener); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server failed: %w", err)
	}

	return nil
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	logger := logging.Get()
	logger.Info("Shutting down HTTP server")
	return s.httpServer.Shutdown(ctx)
}

// Address returns the server address
func (s *Server) Address() string {
	return s.httpServer.Addr
}

