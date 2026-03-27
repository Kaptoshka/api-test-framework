package logger

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"autotests/internal/config"

	"github.com/lmittmann/tint"
)

// New creates a logger with dual output: colored console and timestamped file.
// Log level is controlled by LOG_LEVEL env var (default: info).
// File is written to artifacts/logs/session_[timestamp].log.
// Falls back to console-only if file creation fails.
func New() *slog.Logger {
	level := os.Getenv("LOG_LEVEL")

	var lvl slog.Level
	if err := lvl.UnmarshalText([]byte(level)); err != nil {
		lvl = slog.LevelInfo
	}

	logDir := config.DefaultLogDir
	_ = os.MkdirAll(logDir, 0o750)

	consoleHandler := tint.NewHandler(os.Stdout, &tint.Options{
		Level:      lvl,
		TimeFormat: "2006-01-02T15:04:05",
	})

	logFile, err := os.OpenFile(
		logDir+"/session_"+time.Now().Format("2006-01-02_15-04-05")+".log",
		os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600,
	)

	var handler slog.Handler
	if err != nil {
		handler = consoleHandler
	} else {
		fileHandler := slog.NewTextHandler(logFile, &slog.HandlerOptions{
			Level: lvl,
		})
		handler = &multiHandler{
			handlers: []slog.Handler{consoleHandler, fileHandler},
		}
	}

	return slog.New(handler)
}

// multiHandler combines multiple slog handlers for dual output.
// It delegates all operations to each underlying handler sequentially.
type multiHandler struct {
	// handlers are the underlying slog handlers to delegate to.
	handlers []slog.Handler
}

// Enabled checks if the given log level is enabled by delegating to the first handler.
func (h *multiHandler) Enabled(
	ctx context.Context,
	l slog.Level,
) bool {
	return h.handlers[0].Enabled(ctx, l)
}

// Handle passes the log record to all handlers sequentially.
// Returns the first error encountered.
func (h *multiHandler) Handle(
	ctx context.Context,
	r slog.Record,
) error {
	for _, handler := range h.handlers {
		if err := handler.Handle(ctx, r); err != nil {
			return err
		}
	}

	return nil
}

// WithAttrs returns a new handler with attrs applied to all underlying handlers.
func (h *multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newHandlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		newHandlers[i] = handler.WithAttrs(attrs)
	}
	return &multiHandler{handlers: newHandlers}
}

// WithGroup returns a new handler with group applied to all underlying handlers.
func (h *multiHandler) WithGroup(name string) slog.Handler {
	newHandlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		newHandlers[i] = handler.WithGroup(name)
	}
	return &multiHandler{handlers: newHandlers}
}

// ForTest creates a test-scoped logger with test name in context.
// Use this for test-specific logging with automatic test name tagging.
func ForTest(t *testing.T) *slog.Logger {
	return New().With("test", t.Name())
}
