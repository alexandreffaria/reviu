// Package logger provides a simple logging system with levels
package logger

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

// LogLevel represents the severity of a log message
type LogLevel int

// Log level constants
const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

// Logger defines the interface for logging
type Logger interface {
	// Debug logs a debug message
	Debug(format string, args ...interface{})

	// Info logs an informational message
	Info(format string, args ...interface{})

	// Warn logs a warning message
	Warn(format string, args ...interface{})

	// Error logs an error message
	Error(format string, args ...interface{})

	// WithPrefix returns a new logger with the given prefix
	WithPrefix(prefix string) Logger

	// SetLevel sets the minimum log level to display
	SetLevel(level LogLevel)
}

// SimpleLogger implements the Logger interface
type SimpleLogger struct {
	writer     io.Writer
	level      LogLevel
	prefix     string
	showTime   bool
	timeFormat string
}

// LoggerOption defines functional options for configuring the logger
type LoggerOption func(*SimpleLogger)

// NewLogger creates a new logger with the given options
func NewLogger(options ...LoggerOption) Logger {
	// Default configuration
	logger := &SimpleLogger{
		writer:     os.Stdout,
		level:      INFO,
		prefix:     "",
		showTime:   true,
		timeFormat: "2006-01-02 15:04:05",
	}

	// Apply options
	for _, option := range options {
		option(logger)
	}

	return logger
}

// WithWriter sets the writer for the logger
func WithWriter(writer io.Writer) LoggerOption {
	return func(l *SimpleLogger) {
		l.writer = writer
	}
}

// WithLevel sets the minimum log level
func WithLevel(level LogLevel) LoggerOption {
	return func(l *SimpleLogger) {
		l.level = level
	}
}

// WithPrefix sets the prefix for log messages
func WithPrefix(prefix string) LoggerOption {
	return func(l *SimpleLogger) {
		l.prefix = prefix
	}
}

// WithTimeFormat sets the time format string
func WithTimeFormat(format string) LoggerOption {
	return func(l *SimpleLogger) {
		l.timeFormat = format
	}
}

// WithoutTime disables timestamp in log messages
func WithoutTime() LoggerOption {
	return func(l *SimpleLogger) {
		l.showTime = false
	}
}

// levelString returns a string representation of the log level
func levelString(level LogLevel) string {
	switch level {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO "
	case WARN:
		return "WARN "
	case ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// log logs a message at the specified level
func (l *SimpleLogger) log(level LogLevel, format string, args ...interface{}) {
	// Skip if level is below current minimum
	if level < l.level {
		return
	}

	// Build log message
	var message strings.Builder

	// Add timestamp if enabled
	if l.showTime {
		message.WriteString(time.Now().Format(l.timeFormat))
		message.WriteString(" ")
	}

	// Add level and prefix
	message.WriteString("[")
	message.WriteString(levelString(level))
	message.WriteString("] ")
	
	if l.prefix != "" {
		message.WriteString(l.prefix)
		message.WriteString(": ")
	}

	// Add formatted message
	message.WriteString(fmt.Sprintf(format, args...))

	// Write to output
	fmt.Fprintln(l.writer, message.String())
}

// Debug logs a debug message
func (l *SimpleLogger) Debug(format string, args ...interface{}) {
	l.log(DEBUG, format, args...)
}

// Info logs an informational message
func (l *SimpleLogger) Info(format string, args ...interface{}) {
	l.log(INFO, format, args...)
}

// Warn logs a warning message
func (l *SimpleLogger) Warn(format string, args ...interface{}) {
	l.log(WARN, format, args...)
}

// Error logs an error message
func (l *SimpleLogger) Error(format string, args ...interface{}) {
	l.log(ERROR, format, args...)
}

// WithPrefix returns a new logger with the given prefix
func (l *SimpleLogger) WithPrefix(prefix string) Logger {
	// Create a new logger with the same configuration
	newLogger := &SimpleLogger{
		writer:     l.writer,
		level:      l.level,
		prefix:     prefix,
		showTime:   l.showTime,
		timeFormat: l.timeFormat,
	}

	return newLogger
}

// SetLevel sets the minimum log level
func (l *SimpleLogger) SetLevel(level LogLevel) {
	l.level = level
}

// FileLogger creates a new logger that writes to a file
func FileLogger(filename string, options ...LoggerOption) (Logger, error) {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file %s: %w", filename, err)
	}

	// Add the file writer option to the provided options
	allOptions := append([]LoggerOption{WithWriter(file)}, options...)

	return NewLogger(allOptions...), nil
}

// MultiLogger creates a logger that writes to multiple outputs
func MultiLogger(loggers ...Logger) Logger {
	return &multiLogger{loggers: loggers}
}

// multiLogger implements a logger that writes to multiple loggers
type multiLogger struct {
	loggers []Logger
}

// Debug logs a debug message to all loggers
func (m *multiLogger) Debug(format string, args ...interface{}) {
	for _, logger := range m.loggers {
		logger.Debug(format, args...)
	}
}

// Info logs an informational message to all loggers
func (m *multiLogger) Info(format string, args ...interface{}) {
	for _, logger := range m.loggers {
		logger.Info(format, args...)
	}
}

// Warn logs a warning message to all loggers
func (m *multiLogger) Warn(format string, args ...interface{}) {
	for _, logger := range m.loggers {
		logger.Warn(format, args...)
	}
}

// Error logs an error message to all loggers
func (m *multiLogger) Error(format string, args ...interface{}) {
	for _, logger := range m.loggers {
		logger.Error(format, args...)
	}
}

// WithPrefix returns a new logger with the given prefix
func (m *multiLogger) WithPrefix(prefix string) Logger {
	newLoggers := make([]Logger, len(m.loggers))
	for i, logger := range m.loggers {
		newLoggers[i] = logger.WithPrefix(prefix)
	}
	return &multiLogger{loggers: newLoggers}
}

// SetLevel sets the minimum log level for all loggers
func (m *multiLogger) SetLevel(level LogLevel) {
	for _, logger := range m.loggers {
		logger.SetLevel(level)
	}
}