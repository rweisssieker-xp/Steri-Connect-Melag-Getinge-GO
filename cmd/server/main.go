package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"steri-connect-go/internal/api"
	"steri-connect-go/internal/api/middleware"
	"steri-connect-go/internal/config"
	"steri-connect-go/internal/database"
	"steri-connect-go/internal/devices"
	"steri-connect-go/internal/logging"
)

func main() {
	// Load configuration
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger with config
	logConfig := logging.Config{
		Level:         cfg.Logging.Level,
		Format:        cfg.Logging.Format,
		Output:        cfg.Logging.Output,
		MaxFileSizeMB: cfg.Logging.MaxFileSizeMB,
		MaxBackups:    cfg.Logging.MaxBackups,
		Compress:      cfg.Logging.Compress,
	}

	if err := logging.Init(logConfig); err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logging.Get().Close()

	// Initialize log buffer for Test UI (keep last 1000 entries)
	logging.InitBuffer(1000)

	logger := logging.Get()
	logger.Info("Starting application", "version", "1.0.0")

	// Initialize database with config
	logger.Info("Initializing database", "path", cfg.Database.Path)

	if err := database.InitializeDatabase(cfg.Database.Path); err != nil {
		logger.Error("Failed to initialize database", "error", err)
		os.Exit(1)
	}
	defer database.Close()

	logger.Info("Database initialized successfully")

	// Initialize device manager
	deviceManager := devices.NewManager()
	devices.SetManager(deviceManager) // Set as global manager for API handlers
	if err := deviceManager.LoadDevices(); err != nil {
		logger.Warn("Failed to load devices", "error", err)
		// Continue anyway - devices can be added later
	}
	defer deviceManager.Shutdown()

	logger.Info("Device manager initialized")

	// Initialize metrics middleware
	middleware.InitMetrics()

	// Setup router
	router := api.SetupRouter()

	// Create and start HTTP server with config
	server := api.NewServer(cfg.Server.Port, cfg.Server.BindAddress, router)

	// Security warning if binding to network without authentication
	if cfg.Server.BindAddress == "0.0.0.0" && !cfg.Auth.APIKeyRequired {
		logger.Warn("Security warning: Server is binding to 0.0.0.0 (all interfaces) without API key authentication",
			"bind_address", cfg.Server.BindAddress,
			"api_key_required", cfg.Auth.APIKeyRequired,
			"recommendation", "Enable API key authentication for network access")
	}

	logger.Info("Starting HTTP server",
		"port", cfg.Server.Port,
		"bind_address", cfg.Server.BindAddress,
		"api_key_required", cfg.Auth.APIKeyRequired,
		"api_path", fmt.Sprintf("http://%s:%d/api/", cfg.Server.BindAddress, cfg.Server.Port),
		"websocket_path", fmt.Sprintf("ws://%s:%d/ws", cfg.Server.BindAddress, cfg.Server.Port))

	// Start server in goroutine
	go func() {
		if err := server.Start(); err != nil {
			logger.Error("Server failed", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan

	logger.Info("Shutdown signal received")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Gracefully shutdown server
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Error shutting down server", "error", err)
	} else {
		logger.Info("Server shut down gracefully")
	}
}
