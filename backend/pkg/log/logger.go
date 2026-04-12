// Package log provides a shared slog wrapper with context propagation
// for use across both HTTP (Echo) and gRPC services.
package log

import (
	"context"
	"io"
	"log/slog"
	"os"
	"sync/atomic"
)

type ctxKey struct{}

// defaultLogger uses atomic.Pointer for thread-safe access.
var defaultLogger atomic.Pointer[slog.Logger]

func init() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	defaultLogger.Store(logger)
}

// Config holds logger configuration.
type Config struct {
	Level  string `envconfig:"LEVEL" split_words:"true" default:"info"`
	Format string `envconfig:"FORMAT" split_words:"true" default:"json"` // json or text
}

// NewLogger creates a new slog.Logger instance.
func NewLogger(cfg Config, w io.Writer) *slog.Logger {
	if w == nil {
		w = os.Stdout
	}

	level := parseLevel(cfg.Level)

	opts := &slog.HandlerOptions{
		Level: level,
	}

	var handler slog.Handler
	if cfg.Format == "text" {
		handler = slog.NewTextHandler(w, opts)
	} else {
		handler = slog.NewJSONHandler(w, opts)
	}

	return slog.New(handler)
}

// parseLevel parses a log level string into slog.Level.
func parseLevel(s string) slog.Level {
	var level slog.Level
	if err := level.UnmarshalText([]byte(s)); err != nil {
		return slog.LevelInfo
	}
	return level
}

// SetDefault sets the default logger (thread-safe).
func SetDefault(l *slog.Logger) {
	defaultLogger.Store(l)
}

// Default returns the default logger (thread-safe).
func Default() *slog.Logger {
	return defaultLogger.Load()
}

// WithLogger returns a new context with the logger attached.
func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, ctxKey{}, logger)
}

// FromContext extracts the logger from the context.
// If no logger is found, it returns the default logger.
func FromContext(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(ctxKey{}).(*slog.Logger); ok {
		return logger
	}
	return Default()
}

// With returns a logger with additional attributes from context.
// If no logger is in context, uses the default logger.
func With(ctx context.Context, args ...any) *slog.Logger {
	return FromContext(ctx).With(args...)
}
