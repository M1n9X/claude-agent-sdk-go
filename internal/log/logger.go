package log

import (
	"fmt"
	"os"
)

// Logger provides simple logging for the SDK.
// It writes to stderr with an [SDK] prefix to distinguish from CLI output.
type Logger struct {
	verbose bool
}

// NewLogger creates a new logger instance.
func NewLogger(verbose bool) *Logger {
	return &Logger{
		verbose: verbose,
	}
}

// Debug logs a debug message (only when verbose mode is enabled).
func (l *Logger) Debug(format string, args ...interface{}) {
	if l.verbose {
		fmt.Fprintf(os.Stderr, "[SDK DEBUG] "+format+"\n", args...)
	}
}

// Info logs an informational message (only when verbose mode is enabled).
func (l *Logger) Info(format string, args ...interface{}) {
	if l.verbose {
		fmt.Fprintf(os.Stderr, "[SDK INFO] "+format+"\n", args...)
	}
}

// Warning logs a warning message (always displayed).
func (l *Logger) Warning(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "[SDK WARNING] "+format+"\n", args...)
}

// Error logs an error message (always displayed).
func (l *Logger) Error(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "[SDK ERROR] "+format+"\n", args...)
}
