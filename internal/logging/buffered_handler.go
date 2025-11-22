package logging

import (
	"context"
	"log/slog"
)

// BufferedHandler wraps a slog handler and also writes to the log buffer
type BufferedHandler struct {
	handler slog.Handler
}

// NewBufferedHandler creates a new buffered handler
func NewBufferedHandler(handler slog.Handler) *BufferedHandler {
	return &BufferedHandler{handler: handler}
}

// Enabled reports whether the handler handles records at the given level
func (h *BufferedHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

// Handle handles the record
func (h *BufferedHandler) Handle(ctx context.Context, record slog.Record) error {
	// Write to buffer
	if globalBuffer != nil {
		fields := make(map[string]interface{})
		record.Attrs(func(a slog.Attr) bool {
			fields[a.Key] = a.Value.Any()
			return true
		})

		levelStr := record.Level.String()
		AddEntry(levelStr, record.Message, fields)
	}

	// Write to original handler
	return h.handler.Handle(ctx, record)
}

// WithAttrs returns a new handler with the given attributes
func (h *BufferedHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return NewBufferedHandler(h.handler.WithAttrs(attrs))
}

// WithGroup returns a new handler with the given group
func (h *BufferedHandler) WithGroup(name string) slog.Handler {
	return NewBufferedHandler(h.handler.WithGroup(name))
}

