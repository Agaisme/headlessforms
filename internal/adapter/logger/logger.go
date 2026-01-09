// Package logger provides structured logging using Go's slog package.
// It provides a centralized, configured logger for the entire application.
package logger

import (
	"context"
	"log/slog"
	"os"
)

var defaultLogger *slog.Logger

func init() {
	// Default to JSON handler for production-friendly structured logs
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	defaultLogger = slog.New(handler)
	slog.SetDefault(defaultLogger)
}

// Init initializes the logger with custom options
func Init(level slog.Level, jsonOutput bool) {
	var handler slog.Handler
	opts := &slog.HandlerOptions{Level: level}

	if jsonOutput {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	defaultLogger = slog.New(handler)
	slog.SetDefault(defaultLogger)
}

// Debug logs a debug message
func Debug(msg string, args ...any) {
	defaultLogger.Debug(msg, args...)
}

// Info logs an info message
func Info(msg string, args ...any) {
	defaultLogger.Info(msg, args...)
}

// Warn logs a warning message
func Warn(msg string, args ...any) {
	defaultLogger.Warn(msg, args...)
}

// Error logs an error message
func Error(msg string, args ...any) {
	defaultLogger.Error(msg, args...)
}

// With returns a logger with the given attributes
func With(args ...any) *slog.Logger {
	return defaultLogger.With(args...)
}

// WithContext returns a logger with context
func WithContext(ctx context.Context) *slog.Logger {
	return defaultLogger
}

// API logs an API request
func API(method, path string) {
	defaultLogger.Info("API request", "method", method, "path", path)
}

// Webhook logs webhook events
func Webhook(event string, url string, args ...any) {
	attrs := append([]any{"event", event, "url", url}, args...)
	defaultLogger.Info("Webhook", attrs...)
}

// WebhookError logs webhook errors
func WebhookError(event string, url string, err error, args ...any) {
	attrs := append([]any{"event", event, "url", url, "error", err}, args...)
	defaultLogger.Error("Webhook failed", attrs...)
}
