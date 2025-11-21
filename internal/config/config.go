package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Logging  LoggingConfig  `yaml:"logging"`
	Auth     AuthConfig     `yaml:"auth"`
	Devices  DevicesConfig  `yaml:"devices"`
	TestUI   TestUIConfig   `yaml:"test_ui"`
}

// ServerConfig represents server configuration
type ServerConfig struct {
	Port       int    `yaml:"port"`
	BindAddress string `yaml:"bind_address"`
}

// DatabaseConfig represents database configuration
type DatabaseConfig struct {
	Path string `yaml:"path"`
}

// LoggingConfig represents logging configuration
type LoggingConfig struct {
	Level         string `yaml:"level"`
	Format        string `yaml:"format"`
	Output        string `yaml:"output"`
	MaxFileSizeMB int    `yaml:"max_file_size_mb"`
	MaxBackups    int    `yaml:"max_backups"`
	Compress      bool   `yaml:"compress"`
}

// AuthConfig represents authentication configuration
type AuthConfig struct {
	APIKeyRequired bool   `yaml:"api_key_required"`
	APIKey         string `yaml:"api_key"`
}

// DevicesConfig represents device configuration
type DevicesConfig struct {
	Melag  MelagConfig  `yaml:"melag"`
	Getinge GetingeConfig `yaml:"getinge"`
}

// MelagConfig represents Melag device configuration
type MelagConfig struct {
	StatusPollInterval int `yaml:"status_poll_interval"`
}

// GetingeConfig represents Getinge device configuration
type GetingeConfig struct {
	PingInterval int `yaml:"ping_interval"`
	PingTimeout  int `yaml:"ping_timeout"`
}

// TestUIConfig represents Test UI configuration
type TestUIConfig struct {
	Enabled    bool `yaml:"enabled"`
	RequireAuth bool `yaml:"require_auth"`
}

var globalConfig *Config

// Load loads configuration from file and environment variables
func Load(configPath string) (*Config, error) {
	cfg := getDefaults()

	// Load from YAML file if it exists
	if configPath != "" {
		if _, err := os.Stat(configPath); err == nil {
			data, err := os.ReadFile(configPath)
			if err != nil {
				return nil, fmt.Errorf("failed to read config file: %w", err)
			}

			if err := yaml.Unmarshal(data, cfg); err != nil {
				return nil, fmt.Errorf("failed to parse config file: %w", err)
			}
		}
	}

	// Apply environment variable overrides
	applyEnvOverrides(cfg)

	// Validate configuration
	if err := validate(cfg); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	globalConfig = cfg
	return cfg, nil
}

// Get returns the global configuration
func Get() *Config {
	if globalConfig == nil {
		return getDefaults()
	}
	return globalConfig
}

// getDefaults returns default configuration values
func getDefaults() *Config {
	return &Config{
		Server: ServerConfig{
			Port:        8080,
			BindAddress: "127.0.0.1",
		},
		Database: DatabaseConfig{
			Path: "./data/steri-connect.db",
		},
		Logging: LoggingConfig{
			Level:         "INFO",
			Format:        "json",
			Output:        "stdout",
			MaxFileSizeMB: 10,
			MaxBackups:    10,
			Compress:      true,
		},
		Auth: AuthConfig{
			APIKeyRequired: false,
		},
		Devices: DevicesConfig{
			Melag: MelagConfig{
				StatusPollInterval: 2,
			},
			Getinge: GetingeConfig{
				PingInterval: 15,
				PingTimeout:  5,
			},
		},
		TestUI: TestUIConfig{
			Enabled:     true,
			RequireAuth: false,
		},
	}
}

// applyEnvOverrides applies environment variable overrides
func applyEnvOverrides(cfg *Config) {
	// Server port
	if portStr := os.Getenv("SERVER_PORT"); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil {
			cfg.Server.Port = port
		}
	}

	// Server bind address
	if addr := os.Getenv("SERVER_BIND_ADDRESS"); addr != "" {
		cfg.Server.BindAddress = addr
	}

	// Database path
	if dbPath := os.Getenv("DATABASE_PATH"); dbPath != "" {
		cfg.Database.Path = dbPath
	}

	// Log level
	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		cfg.Logging.Level = strings.ToUpper(logLevel)
	}

	// Log format
	if logFormat := os.Getenv("LOG_FORMAT"); logFormat != "" {
		cfg.Logging.Format = strings.ToLower(logFormat)
	}

	// Log output
	if logOutput := os.Getenv("LOG_OUTPUT"); logOutput != "" {
		cfg.Logging.Output = logOutput
	}

	// API key required
	if apiKeyRequired := os.Getenv("API_KEY_REQUIRED"); apiKeyRequired != "" {
		cfg.Auth.APIKeyRequired = apiKeyRequired == "true" || apiKeyRequired == "1"
	}

	// API key
	if apiKey := os.Getenv("API_KEY"); apiKey != "" {
		cfg.Auth.APIKey = apiKey
	}
}

// validate validates the configuration
func validate(cfg *Config) error {
	// Validate server port
	if cfg.Server.Port < 1 || cfg.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d (must be between 1 and 65535)", cfg.Server.Port)
	}

	// Validate bind address
	if cfg.Server.BindAddress == "" {
		return fmt.Errorf("bind address cannot be empty")
	}

	// Validate log level
	validLogLevels := map[string]bool{
		"DEBUG": true,
		"INFO":  true,
		"WARN":  true,
		"ERROR": true,
	}
	if !validLogLevels[strings.ToUpper(cfg.Logging.Level)] {
		return fmt.Errorf("invalid log level: %s (must be DEBUG, INFO, WARN, or ERROR)", cfg.Logging.Level)
	}

	// Validate log format
	validLogFormats := map[string]bool{
		"json": true,
		"text": true,
	}
	if !validLogFormats[strings.ToLower(cfg.Logging.Format)] {
		return fmt.Errorf("invalid log format: %s (must be json or text)", cfg.Logging.Format)
	}

	// Validate database path
	if cfg.Database.Path == "" {
		return fmt.Errorf("database path cannot be empty")
	}

	// Validate Melag polling interval
	if cfg.Devices.Melag.StatusPollInterval < 1 {
		return fmt.Errorf("invalid Melag polling interval: %d (must be >= 1)", cfg.Devices.Melag.StatusPollInterval)
	}

	// Validate Getinge ping interval
	if cfg.Devices.Getinge.PingInterval < 1 {
		return fmt.Errorf("invalid Getinge ping interval: %d (must be >= 1)", cfg.Devices.Getinge.PingInterval)
	}

	// Validate Getinge ping timeout
	if cfg.Devices.Getinge.PingTimeout < 1 {
		return fmt.Errorf("invalid Getinge ping timeout: %d (must be >= 1)", cfg.Devices.Getinge.PingTimeout)
	}

	return nil
}

