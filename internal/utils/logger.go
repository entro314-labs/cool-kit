package utils

import (
	"fmt"
	"os"
	"time"

	"github.com/fatih/color"
)

// Logger provides structured logging functionality
type Logger struct {
	logFile *os.File
	verbose bool
}

// NewLogger creates a new logger instance
func NewLogger(logPath string, verbose bool) (*Logger, error) {
	var logFile *os.File
	var err error

	if logPath != "" {
		logFile, err = os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}
	}

	return &Logger{
		logFile: logFile,
		verbose: verbose,
	}, nil
}

// Close closes the log file
func (l *Logger) Close() error {
	if l.logFile != nil {
		return l.logFile.Close()
	}
	return nil
}

// Info logs an info message
func (l *Logger) Info(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	color.Blue("ℹ %s", msg)
	l.writeToFile("INFO", msg)
}

// Success logs a success message
func (l *Logger) Success(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	color.Green("✓ %s", msg)
	l.writeToFile("SUCCESS", msg)
}

// Warning logs a warning message
func (l *Logger) Warning(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	color.Yellow("⚠ %s", msg)
	l.writeToFile("WARNING", msg)
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	color.Red("✗ %s", msg)
	l.writeToFile("ERROR", msg)
}

// Debug logs a debug message (only if verbose)
func (l *Logger) Debug(format string, args ...interface{}) {
	if !l.verbose {
		return
	}
	msg := fmt.Sprintf(format, args...)
	color.Cyan("• %s", msg)
	l.writeToFile("DEBUG", msg)
}

// Section logs a section header
func (l *Logger) Section(title string) {
	fmt.Println()
	fmt.Println("============================================================")
	fmt.Println(title)
	fmt.Println("============================================================")
	l.writeToFile("SECTION", title)
}

// Step logs a step message
func (l *Logger) Step(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	color.Magenta("▶ %s", msg)
	l.writeToFile("STEP", msg)
}

// writeToFile writes a log entry to the log file
func (l *Logger) writeToFile(level, message string) {
	if l.logFile == nil {
		return
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logLine := fmt.Sprintf("[%s] [%s] %s\n", timestamp, level, message)
	l.logFile.WriteString(logLine)
}

// Banner displays a banner
func (l *Logger) Banner(title, subtitle string) {
	color.Blue("╔════════════════════════════════════════════════════════════════╗")
	color.Blue("║ %-62s ║", title)
	if subtitle != "" {
		color.Blue("║ %-62s ║", subtitle)
	}
	color.Blue("╚════════════════════════════════════════════════════════════════╝")
	fmt.Println()
}
