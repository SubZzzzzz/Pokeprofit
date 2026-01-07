package logger

import (
	"io"
	"log/slog"
	"os"
	"strings"
)

// Logger wraps slog.Logger with additional context support.
type Logger struct {
	*slog.Logger
}

// Config holds logger configuration.
type Config struct {
	Level  string
	Output io.Writer
}

// DefaultConfig returns default logger configuration.
func DefaultConfig() Config {
	return Config{
		Level:  "info",
		Output: os.Stdout,
	}
}

// New creates a new Logger with the given configuration.
func New(cfg Config) *Logger {
	level := parseLevel(cfg.Level)

	output := cfg.Output
	if output == nil {
		output = os.Stdout
	}

	handler := slog.NewJSONHandler(output, &slog.HandlerOptions{
		Level: level,
	})

	return &Logger{
		Logger: slog.New(handler),
	}
}

// NewText creates a new Logger with text output (for development).
func NewText(cfg Config) *Logger {
	level := parseLevel(cfg.Level)

	output := cfg.Output
	if output == nil {
		output = os.Stdout
	}

	handler := slog.NewTextHandler(output, &slog.HandlerOptions{
		Level: level,
	})

	return &Logger{
		Logger: slog.New(handler),
	}
}

// With returns a new Logger with additional attributes.
func (l *Logger) With(args ...any) *Logger {
	return &Logger{
		Logger: l.Logger.With(args...),
	}
}

// WithComponent returns a new Logger with a component name.
func (l *Logger) WithComponent(name string) *Logger {
	return l.With("component", name)
}

// WithRequestID returns a new Logger with a request ID.
func (l *Logger) WithRequestID(id string) *Logger {
	return l.With("request_id", id)
}

// WithUserID returns a new Logger with a user ID.
func (l *Logger) WithUserID(id string) *Logger {
	return l.With("user_id", id)
}

// parseLevel parses a level string into slog.Level.
func parseLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// Global logger instance
var defaultLogger *Logger

// Init initializes the default logger.
func Init(cfg Config) {
	defaultLogger = New(cfg)
}

// InitText initializes the default logger with text output.
func InitText(cfg Config) {
	defaultLogger = NewText(cfg)
}

// Default returns the default logger instance.
func Default() *Logger {
	if defaultLogger == nil {
		defaultLogger = New(DefaultConfig())
	}
	return defaultLogger
}

// Debug logs at debug level using the default logger.
func Debug(msg string, args ...any) {
	Default().Debug(msg, args...)
}

// Info logs at info level using the default logger.
func Info(msg string, args ...any) {
	Default().Info(msg, args...)
}

// Warn logs at warn level using the default logger.
func Warn(msg string, args ...any) {
	Default().Warn(msg, args...)
}

// Error logs at error level using the default logger.
func Error(msg string, args ...any) {
	Default().Error(msg, args...)
}

// With returns a new Logger with additional attributes using the default logger.
func With(args ...any) *Logger {
	return Default().With(args...)
}
