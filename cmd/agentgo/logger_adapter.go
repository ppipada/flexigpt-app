package main

import (
	"context"
	"log/slog"
	"os"
)

// Define custom log levels for Trace and Fatal.
const (
	LevelTrace slog.Level = slog.LevelDebug - 4 // Trace level (more verbose than Debug)
	LevelFatal slog.Level = slog.LevelError + 4 // Fatal level (more severe than Error)
)

// SlogLoggerAdapter adapts slog.Logger to satisfy the Logger interface.
type SlogLoggerAdapter struct {
	logger *slog.Logger
}

func NewSlogLoggerAdapter(logger *slog.Logger) *SlogLoggerAdapter {
	return &SlogLoggerAdapter{logger: logger}
}

// Print logs a message at the Info level.
func (l *SlogLoggerAdapter) Print(message string) {
	l.logger.Info(message)
}

// Trace logs a message at the custom Trace level.
func (l *SlogLoggerAdapter) Trace(message string) {
	l.logger.Log(context.Background(), LevelTrace, message)
}

// Debug logs a message at the Debug level.
func (l *SlogLoggerAdapter) Debug(message string) {
	l.logger.Debug(message)
}

// Info logs a message at the Info level.
func (l *SlogLoggerAdapter) Info(message string) {
	l.logger.Info(message)
}

// Warning logs a message at the Warn level.
func (l *SlogLoggerAdapter) Warning(message string) {
	l.logger.Warn(message)
}

// Error logs a message at the Error level.
func (l *SlogLoggerAdapter) Error(message string) {
	l.logger.Error(message)
}

// Fatal logs a message at the custom Fatal level and exits the program.
func (l *SlogLoggerAdapter) Fatal(message string) {
	l.logger.Log(context.Background(), LevelFatal, message)
	os.Exit(1)
}
