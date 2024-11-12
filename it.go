// Package it provides utility functions for error handling and logging,
// simplifying common patterns while adhering to Go's best practices.
//
// This package includes functions that help manage errors by panicking on unrecoverable errors,
// logging errors while continuing execution, and other utilities for robust error handling.
//
// Example usage:
//
//	import "github.com/theHamdiz/it"
//
//	func main() {
//	    hardResult := it.Must(SomeFunction())
//	    softResult := it.Should(SomeFunction())
//	    fmt.Println(hardResult)
//	    fmt.Println(softResult)
//	}
package it

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"sync"
	"time"

	"github.com/fatih/color"
)

// ===================================================
// Declarations Area
// ===================================================

// Define log levels
const (
	LevelTrace = iota
	LevelDebug
	LevelInfo
	LevelWarning
	LevelError
	LevelFatal
	LevelAudit
)

var (
	logOutput    = os.Stdout                 // Default to standard output
	logger       = &Logger{level: LevelInfo} // Default log level
	magentaColor = color.New(color.FgMagenta)
	blueColor    = color.New(color.FgBlue)
	redColor     = color.New(color.FgRed)
	cyanColor    = color.New(color.FgCyan)
	yellowColor  = color.New(color.FgYellow)
	once         sync.Once
)

// ===================================================
// Definitions Area
// ===================================================

// Logger struct to manage log level
type Logger struct {
	level int
	mu    sync.Mutex
}

// StructuredLogEntry represents a structured log message.
type StructuredLogEntry struct {
	Level   string `json:"level"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// ===================================================
// Testing Area
// ===================================================

// Must returns the value x if err is nil. If err is not nil, Must panics with the error.
// Use Must when an error is unrecoverable and should halt the program execution.
//
// Example usage:
//
//	result := it.Must(SomeFunction())
func Must[T any](x T, err error) T {
	CheckError(err)
	return x
}

// Ensure panics if err is not nil. Use Ensure for critical operations
// where an error is unacceptable, much like Must but
// for methods that only return error.
//
// Example usage:
//
//	it.Ensure(SomeErrorOnlyFunction())
func Ensure(err error) {
	if err != nil {
		panic(err)
	}
}

// Should returns the value x regardless of whether err is nil. If err is not nil,
// Should logs the error with the filename and line number.
// Use Should when you want to log an error but continue execution.
//
// Note: Be cautious when using Should, as x may be in an invalid state if err is not nil.
// Ensure that the returned value is safe to use in such cases.
//
// Example usage:
//
//	result := it.Should(SomeFunction())
func Should[T any](x T, err error) T {
	LogError(err)
	return x
}

// Attempt logs the error with filename and line number if err is not nil,
// but allows the program to continue, much like Should but for methods
// that only return an error.
//
// Example usage:
//
//	it.Attempt(SomeErrorOnlyFunction())
func Attempt(err error) {
	if err != nil {
		LogError(err)
	}
}

// ===================================================
// Logging Area
// ===================================================

// Trace logs a trace-level message.
// Use Trace to log very detailed information for tracing program execution.
//
// Example usage:
//
//	it.Trace("Entering function X")
func Trace(message string) {
	if logger.level <= LevelTrace {
		_, err := magentaColor.Fprintf(logOutput, "> Trace: %s\n", message)
		if err != nil {
			return
		}
	}
}

// Tracef logs a formatted trace-level message.
// Use Tracef to log formatted detailed tracing information.
//
// Example usage:
//
//	it.Tracef("Processing item %d of %d", currentItem, totalItems)
func Tracef(format string, args ...interface{}) {
	if logger.level <= LevelTrace {
		_, err := magentaColor.Fprintf(logOutput, "> Trace: "+format+"\n", args...)
		if err != nil {
			return
		}
	}
}

// Debug logs a debug-level message.
// Use Debug to log detailed information useful for debugging.
//
// Example usage:
//
//	it.Debug("Loaded configuration successfully")
func Debug(message string) {
	if logger.level <= LevelDebug {
		_, err := blueColor.Fprintf(logOutput, "> Debug: %s\n", message)
		if err != nil {
			return
		}
	}
}

// Debugf logs a formatted debug-level message.
// Use Debugf to log detailed formatted information useful for debugging.
//
// Example usage:
//
//	it.Debugf("User %s has %d pending messages", username, messageCount)
func Debugf(format string, args ...interface{}) {
	if logger.level <= LevelDebug {
		_, err := blueColor.Printf("> Debug: "+format+"\n", args...)
		if err != nil {
			return
		}
	}
}

// LogError logs the error with the filename and line number if err is not nil.
// Use LogError when you want to log an error but handle it manually.
//
// Example usage:
//
//	if err != nil {
//	    it.LogError(err)
//	    // Additional error handling...
//	}
func LogError(err error) {
	if err != nil {
		Errorf("%v", err)
	}
}

// CheckError logs the error and exits the program if err is not nil.
// Use CheckError when an error is unrecoverable and the program should terminate.
//
// Example usage:
//
//	it.CheckError(err)
func CheckError(err error) {
	if err != nil {
		panic(err)
	}
}

// Info logs an informational message.
// Use Info to log general information about the program's execution.
//
// Example usage:
//
//	it.Info("Starting the application")
func Info(message string) {
	if logger.level <= LevelInfo {
		_, err := cyanColor.Fprintf(logOutput, "> Info: %s\n", message)
		if err != nil {
			return
		}
	}
}

// Infof logs a formatted informational message.
// Use Infof to log formatted information about the program's execution.
//
// Example usage:
//
//	it.Infof("Starting the application version %s", version)
func Infof(format string, args ...interface{}) {
	if logger.level <= LevelInfo {
		_, err := cyanColor.Fprintf(logOutput, "> Info: "+format+"\n", args...)
		if err != nil {
			return
		}

	}
}

// Warn logs a warning message.
// Use Warn to log non-critical issues that should be addressed.
//
// Example usage:
//
//	it.Warn("Configuration file not found, using defaults")
func Warn(message string) {
	if logger.level <= LevelWarning {
		_, err := yellowColor.Fprintf(logOutput, "> Warning: %s\n", message)
		if err != nil {
			return
		}
	}
}

// Warnf logs a formatted warning message.
// Use Warnf to log formatted non-critical issues.
//
// Example usage:
//
//	it.Warnf("Configuration file %s not found, using defaults", configFile)
func Warnf(format string, args ...interface{}) {
	if logger.level <= LevelWarning {
		_, err := yellowColor.Fprintf(logOutput, "> Warning: "+format+"\n", args...)
		if err != nil {
			return
		}
	}
}

// Error logs an error message.
// Use Error to log errors that occurred during execution.
//
// Example usage:
//
//	it.Error("Failed to connect to the database")
func Error(message string) {
	if logger.level <= LevelError {
		_, err := redColor.Fprintf(logOutput, "> Error: %s\n", message)
		if err != nil {
			return
		}
	}
}

// Errorf logs a formatted error message.
// Use Errorf to log formatted errors that occurred during execution.
//
// Example usage:
//
//	it.Errorf("Failed to connect to the database: %v", err)
func Errorf(format string, args ...interface{}) {
	if logger.level <= LevelError {
		_, err := redColor.Fprintf(logOutput, "> Error: "+format+"\n", args...)
		if err != nil {
			return
		}
	}
}

// CheckErrorf logs a formatted error message and exits the program if err is not nil.
// Use CheckErrorf when an error is unrecoverable and the program should terminate.
//
// Example usage:
//
//	it.CheckErrorf(err, "Failed to start server on port %d", port)
func CheckErrorf(err error, format string, args ...interface{}) {
	if err != nil {
		Errorf(format+": %v", append(args, err)...)
		os.Exit(1)
	}
}

// StructuredLog logs a message with a specified level in a structured format.
// Use StructuredLog to log messages with additional data for any log level.
//
// Example usage:
//
//	it.StructuredLog("ERROR", "Failed to process request", map[string]interface{}{"requestID": "abc123", "error": err})
func StructuredLog(level string, message string, data any) {
	if shouldLogLevel(level) {
		entry := StructuredLogEntry{
			Level:   level,
			Message: message,
			Data:    data,
		}
		jsonData, err := json.Marshal(entry)
		if err != nil {
			_, err := fmt.Fprintf(logOutput, "> Error: Failed to marshal log entry: %v\n", err)
			if err != nil {
				return
			}
			return
		}
		_, err = fmt.Fprintln(logOutput, string(jsonData))
		if err != nil {
			return
		}
	}
}

// StructuredInfo logs an informational message in a structured (e.g., JSON) format.
// Use StructuredInfo to log messages with additional data in a structured format for easier parsing and analysis.
//
// Example usage:
//
//	it.StructuredInfo("User logged in", map[string]string{"username": "johndoe", "ip": "192.168.1.1"})
func StructuredInfo(message string, data any) {
	StructuredLog("INFO", message, data)
}

// StructuredDebug logs a debug-level message in a structured format.
// Use StructuredDebug to log detailed debug information with additional data.
//
// Example usage:
//
//	it.StructuredDebug("Cache hit", map[string]string{"key": "user:1234"})

func StructuredDebug(message string, data interface{}) {
	StructuredLog("DEBUG", message, data)
}

// StructuredWarning logs a warning message in a structured format.
// Use StructuredWarning to log warnings with additional contextual data.
//
// Example usage:
//
//	it.StructuredWarning("High memory usage detected", map[string]interface{}{"usage": 95})

func StructuredWarning(message string, data interface{}) {
	StructuredLog("WARNING", message, data)
}

// StructuredError logs an error message in a structured format.
// Use StructuredError to log errors with additional data for better analysis.
//
// Example usage:
//
//	it.StructuredError("File not found", map[string]string{"filename": "config.yaml"})
func StructuredError(message string, data interface{}) {
	StructuredLog("ERROR", message, data)
}

// WrapError wraps an error with a message.
// Use WrapError to add context to an error.
//
// Example usage:
//
//	return it.WrapError(err, "failed to open file")
func WrapError(err error, message string) error {
	if err != nil {
		return fmt.Errorf("%s: %w", message, err)
	}
	return nil
}

// WrapErrorf wraps an error with a formatted message.
// Use WrapErrorf to add formatted context to an error.
//
// Example usage:
//
//	return it.WrapErrorf(err, "failed to open file %s", filename)
func WrapErrorf(err error, format string, args ...interface{}) error {
	if err != nil {
		return fmt.Errorf(format+": %w", append(args, err)...)
	}
	return nil
}

// WrapErrorWithContext wraps an error with additional contextual information.
// Useful for adding more details to errors without losing the original error.
//
// Example usage:
//
//	err := it.WrapErrorWithContext(err, "processing file", map[string]string{"file": filename})
func WrapErrorWithContext(err error, message string, context map[string]string) error {
	if err != nil {
		return fmt.Errorf("%s: %w [context: %v]", message, err, context)
	}
	return nil
}

// LogStackTrace logs the current stack trace.
// Use LogStackTrace to debug complex issues.
//
// Example usage:
//
//	it.LogStackTrace()
func LogStackTrace() {
	stack := debug.Stack()
	Error("Stack trace:")
	Error(string(stack))
}

// LogErrorWithStack logs an error along with the current stack trace.
// Use LogErrorWithStack to provide detailed error information.
//
// Example usage:
//
//	if err != nil {
//	    it.LogErrorWithStack(err)
//	}
func LogErrorWithStack(err error) {
	if err != nil {
		Errorf("Error: %v", err)
		LogStackTrace()
	}
}

// Audit logs a message with the custom Audit level.
// Use Audit for logging audit-related information.
//
// Example usage:
//
//	it.Audit("User login attempt recorded")
func Audit(message string) {
	if logger.level <= LevelAudit {
		_, err := fmt.Fprintf(logOutput, "> Audit: %s\n", message)
		if err != nil {
			return
		}
	}
}

// DeferWithLog defers a function with an optional log message.
// Useful for managing complex defer chains with logs.
//
// Example usage:
//
//	defer it.DeferWithLog("Cleanup complete")()
func DeferWithLog(message string) func() {
	return func() {
		Info(message)
	}
}

// LogOnce logs a message only once.
// Useful in scenarios where a repeated log would clutter output.
//
// Example usage:
//
//	it.LogOnce("This message will only be logged once")
func LogOnce(message string) {
	once.Do(func() {
		Info(message)
	})
}

// WaitFor waits until the provided function returns true or times out after `timeout` duration.
// Useful for waiting on certain conditions in concurrent environments.
//
// Example usage:
//
//	it.WaitFor(time.Second * 10, func() bool { return someCondition() })
func WaitFor(timeout time.Duration, condition func() bool) bool {
	end := time.Now().Add(timeout)
	for time.Now().Before(end) {
		if condition() {
			return true
		}
		time.Sleep(100 * time.Millisecond)
	}
	return false
}

// ===================================================
// Panic Recovery Area
// ===================================================

// RecoverPanicAndExit recovers from a panic, logs the error and stack trace, and exits.
// Use RecoverPanic with defer to handle panics gracefully.
//
// Example usage:
//
//	defer it.RecoverPanic()
func RecoverPanicAndExit() {
	if r := recover(); r != nil {
		Errorf("Panic recovered: %v", r)
		Error(string(debug.Stack()))
		os.Exit(1)
	}
}

// RecoverPanicAndContinue returns a function suitable for use with defer.
// It recovers from a panic, logs the error and stack trace.
func RecoverPanicAndContinue() {
	if r := recover(); r != nil {
		Errorf("Panic recovered: %v", r)
		Error(string(debug.Stack()))
	}
}

// ===================================================
// Timing Area
// ===================================================

// TimeFunction measures and logs the execution time of a function.
// Use TimeFunction to profile the performance of specific functions.
//
// Example usage:
//
//	it.TimeFunction("compute", compute)
func TimeFunction(name string, f func()) {
	start := time.Now()
	f()
	duration := time.Since(start)
	Infof("Function %s took %v", name, duration)
}

// TimeBlock starts a timer and returns a function to stop the timer and log the duration.
// Use TimeBlock with defer to measure the execution time of a code block.
//
// Example usage:
//
//	defer it.TimeBlock("main")()
func TimeBlock(name string) func() {
	start := time.Now()
	return func() {
		duration := time.Since(start)
		Infof("Block %s took %v", name, duration)
	}
}

// ===================================================
// Retrial Area
// ===================================================

// Retry retries the given function `fn` up to `attempts` times with a delay between attempts.
// Returns the error from the last attempt if all attempts fail.
//
// Example usage:
//
//	err := it.Retry(3, time.Second, SomeErrorOnlyFunction)
func Retry(attempts int, delay time.Duration, fn func() error) error {
	for i := 0; i < attempts; i++ {
		if err := fn(); err != nil {
			if i == attempts-1 {
				return err
			}
			time.Sleep(delay)
			continue
		}
		return nil
	}
	return nil
}

// RetryExponential retries the given function `fn` up to `attempts` times with an exponential backoff delay between attempts.
// The delay starts at `initialDelay` and doubles with each attempt. Returns the error from the last attempt if all attempts fail.
//
// Example usage:
//
//	err := it.RetryExponential(5, time.Second, SomeErrorOnlyFunction)
func RetryExponential(attempts int, initialDelay time.Duration, fn func() error) error {
	delay := initialDelay
	for i := 0; i < attempts; i++ {
		if err := fn(); err != nil {
			if i == attempts-1 {
				return err // Return the last error if all attempts fail
			}
			time.Sleep(delay)
			delay *= 2 // Double the delay for exponential backoff
			continue
		}
		return nil // Success, no need to retry
	}
	return nil
}

// ===================================================
// Configuration Area
// ===================================================

// InitFromEnv initializes the logger settings from environment variables.
// Supported environment variables:
// - LOG_LEVEL: TRACE, DEBUG, INFO, WARN, ERROR, FATAL
// - LOG_FILE: Path to a file to write logs to
//
// Example usage:
//
//	it.InitFromEnv()
func InitFromEnv() {
	levelStr := os.Getenv("LOG_LEVEL")
	switch levelStr {
	case "TRACE":
		SetLogLevel(LevelTrace)
	case "DEBUG":
		SetLogLevel(LevelDebug)
	case "INFO":
		SetLogLevel(LevelInfo)
	case "WARN":
		SetLogLevel(LevelWarning)
	case "ERROR":
		SetLogLevel(LevelError)
	case "FATAL":
		SetLogLevel(LevelFatal)
	case "AUDIT":
		SetLogLevel(LevelAudit)
	default:
		SetLogLevel(LevelInfo)
	}

	logFile := os.Getenv("LOG_FILE")
	if logFile != "" {
		file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err == nil {
			SetLogOutput(file)
		} else {
			Errorf("Failed to open log file %s: %v", logFile, err)
		}
	}
}

// init sets the logger to include standard flags along with the filename and line number.
func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

// SetLogOutput sets the output destination for logs.
// It can be a file, os.Stdout, os.Stderr, or any io.Writer.
//
// Example usage:
//
//	file, err := os.OpenFile("app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
//	if err != nil {
//	    it.Fatalf("Failed to open log file: %v", err)
//	}
//	defer file.Close()
//	it.SetLogOutput(file)
func SetLogOutput(output *os.File) {
	logger.mu.Lock()
	defer logger.mu.Unlock()
	logOutput = output
}

// SetLogLevel sets the minimum log level for logging.
// Messages below this level will not be logged.
//
// Available log levels:
//
//		it.LevelTrace
//		it.LevelDebug
//		it.LevelInfo
//		it.LevelWarn
//		it.LevelError
//		it.LevelFatal
//	    it.LevelAudit
//
// Example usage:
//
//	it.SetLogLevel(it.LevelDebug)
func SetLogLevel(level int) {
	logger.mu.Lock()
	defer logger.mu.Unlock()
	logger.level = level
}

// ===================================================
// Helper Functions Area
// ===================================================

// Helper function to check log level
func shouldLogLevel(level string) bool {
	levelMap := map[string]int{
		"TRACE": LevelTrace,
		"DEBUG": LevelDebug,
		"INFO":  LevelInfo,
		"WARN":  LevelWarning,
		"ERROR": LevelError,
		"FATAL": LevelFatal,
	}
	return logger.level <= levelMap[level]
}
