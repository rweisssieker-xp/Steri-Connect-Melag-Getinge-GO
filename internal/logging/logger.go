package logging

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
)

// Logger wraps the structured logger
type Logger struct {
	*slog.Logger
	file   *os.File
	writer io.Writer
	mu     sync.Mutex
}

var globalLogger *Logger

// Config represents logging configuration
type Config struct {
	Level          string // DEBUG, INFO, WARN, ERROR
	Format         string // json or text
	Output         string // stdout or file path
	MaxFileSizeMB  int    // Max file size in MB before rotation
	MaxBackups     int    // Number of backup files to keep
	Compress       bool   // Whether to compress old logs
}

// Init initializes the global logger
func Init(config Config) error {
	var level slog.Level
	switch config.Level {
	case "DEBUG":
		level = slog.LevelDebug
	case "INFO":
		level = slog.LevelInfo
	case "WARN":
		level = slog.LevelWarn
	case "ERROR":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	var writer io.Writer
	var file *os.File

	if config.Output == "stdout" {
		writer = os.Stdout
	} else {
		// Create log directory if needed
		logDir := filepath.Dir(config.Output)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return fmt.Errorf("failed to create log directory: %w", err)
		}

		// Open log file
		var err error
		file, err = os.OpenFile(config.Output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return fmt.Errorf("failed to open log file: %w", err)
		}
		writer = file
	}

	var baseHandler slog.Handler
	if config.Format == "json" {
		baseHandler = slog.NewJSONHandler(writer, opts)
	} else {
		baseHandler = slog.NewTextHandler(writer, opts)
	}

	// Wrap handler to also write to buffer
	handler := NewBufferedHandler(baseHandler)

	globalLogger = &Logger{
		Logger: slog.New(handler),
		file:   file,
		writer: writer,
	}

	return nil
}

// Get returns the global logger
func Get() *Logger {
	if globalLogger == nil {
		// Default logger if not initialized
		handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
		globalLogger = &Logger{
			Logger: slog.New(handler),
			writer: os.Stdout,
		}
	}
	return globalLogger
}

// WithContext adds context fields to the logger
func (l *Logger) WithContext(fields map[string]interface{}) *Logger {
	args := make([]interface{}, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}
	return &Logger{
		Logger: l.Logger.With(args...),
		file:   l.file,
		writer: l.writer,
	}
}

// WithDevice adds device_id context
func (l *Logger) WithDevice(deviceID int) *Logger {
	return &Logger{
		Logger: l.Logger.With("device_id", deviceID),
		file:   l.file,
		writer: l.writer,
	}
}

// WithCycle adds cycle_id context
func (l *Logger) WithCycle(cycleID int) *Logger {
	return &Logger{
		Logger: l.Logger.With("cycle_id", cycleID),
		file:   l.file,
		writer: l.writer,
	}
}

// WithUser adds user context
func (l *Logger) WithUser(user string) *Logger {
	return &Logger{
		Logger: l.Logger.With("user", user),
		file:   l.file,
		writer: l.writer,
	}
}

// WithAction adds action context
func (l *Logger) WithAction(action string) *Logger {
	return &Logger{
		Logger: l.Logger.With("action", action),
		file:   l.file,
		writer: l.writer,
	}
}

// Close closes the log file if it was opened
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

// rotateLog rotates the log file if it exceeds max size
func RotateLog(logPath string, maxSizeMB int, maxBackups int, compress bool) error {
	info, err := os.Stat(logPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // File doesn't exist yet, nothing to rotate
		}
		return err
	}

	maxSizeBytes := int64(maxSizeMB) * 1024 * 1024
	if info.Size() < maxSizeBytes {
		return nil // File hasn't reached max size
	}

	// Rotate existing files
	for i := maxBackups - 1; i >= 1; i-- {
		oldPath := fmt.Sprintf("%s.%d", logPath, i)
		if compress && i > 0 {
			oldPath += ".gz"
		}
		newPath := fmt.Sprintf("%s.%d", logPath, i+1)
		if compress && i > 0 {
			newPath += ".gz"
		}

		if _, err := os.Stat(oldPath); err == nil {
			os.Rename(oldPath, newPath)
		}
	}

	// Rotate current log file
	backupPath := fmt.Sprintf("%s.1", logPath)
	if compress {
		backupPath += ".gz"
		// Note: Actual compression would require gzip package
		// For now, just rename the file
	}
	return os.Rename(logPath, backupPath)
}

