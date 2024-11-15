// Package it provides utility functions for error handling, logging, load balancing,
// testing, timing & graceful actions.
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
//	    it.Info(hardResult)
//	    it.Info(softResult)
//	}
package it

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"reflect"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"syscall"
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
	logOutput    atomic.Value
	output       io.Writer
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
	level int32
}

type LoggerOptions struct {
	level     int
	timestamp bool
}

// BufferedLogger provides a buffered logging mechanism, allowing logs to be written to any io.Writer.
type BufferedLogger struct {
	writer *bufio.Writer
}

// StructuredLogEntry represents a structured log message.
type StructuredLogEntry struct {
	Level   string `json:"level"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// LoggerOption is a functional option type for configuring Logger.
type LoggerOption func(*LoggerOptions)

// ===================================================
// Builders & Constructors Area
// ===================================================

// WithLogLevel sets the log level for Logger.
func WithLogLevel(level int) LoggerOption {
	return func(opts *LoggerOptions) {
		opts.level = level
	}
}

// WithTimestamp enables timestamp logging.
func WithTimestamp(enabled bool) LoggerOption {
	return func(opts *LoggerOptions) {
		opts.timestamp = enabled
	}
}

// NewLogger creates a logger with configurable options.
//
// Example usage:
//
//	logger := it.NewLogger(it.WithLogLevel(it.LevelDebug), it.WithTimestamp(true))
func NewLogger(options ...LoggerOption) *Logger {
	opts := &LoggerOptions{
		level:     LevelInfo,
		timestamp: true,
	}
	for _, option := range options {
		option(opts)
	}
	// Create and configure logger with opts
	return &Logger{level: int32(opts.level)}
}

// NewBufferedLogger creates a new BufferedLogger that writes to the specified io.Writer.
//
// Example usage:
//
//	file, _ := os.OpenFile("app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
//	defer file.Close()
//	logger := it.NewBufferedLogger(file)
//	logger.Log("Logging to a file")
func NewBufferedLogger(output io.Writer) *BufferedLogger {
	return &BufferedLogger{
		writer: bufio.NewWriter(output),
	}
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

// Log writes a message to the buffered writer and flushes the buffer.
//
// Example usage:
//
//	logger := it.NewBufferedLogger(os.Stdout)
//	logger.Log("Logging to standard output")
func (b *BufferedLogger) Log(message string) {
	_, err := b.writer.WriteString(fmt.Sprintf("%s\n", message))
	if err != nil {
		fmt.Printf("Buffered log error: %v\n", err)
		return
	}
	_ = b.writer.Flush()
}

// Flush forces any buffered data to be written out to the underlying writer.
// Useful for ensuring all logs are written at program exit.
//
// Example usage:
//
//	logger.Flush()
func (b *BufferedLogger) Flush() error {
	return b.writer.Flush()
}

// Trace logs a trace-level message.
// Use Trace to log very detailed information for tracing program execution.
//
// Example usage:
//
//	it.Trace("Entering function X")
func Trace(message string) {
	if logger.level <= LevelTrace {
		_, err := magentaColor.Fprintf(output, "🛤️ Trace: %s\n", message)
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
		_, err := magentaColor.Fprintf(output, "🛤️ Trace: "+format+"\n", args...)
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
		_, err := blueColor.Fprintf(output, "🐛 Debug: %s\n", message)
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
		_, err := blueColor.Printf("🐛 Debug: "+format+"\n", args...)
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
		_, err := cyanColor.Fprintf(output, "✅️ Info: %s\n", message)
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
		_, err := cyanColor.Fprintf(output, "✅️ Info: "+format+"\n", args...)
		if err != nil {
			return
		}

	}
}

// Tip logs a tip message.
// Use Tip to log general information about the program's execution.
//
// Example usage:
//
//	it.Tip("Use select {} to create an infinite blocking loop without consuming CPU, useful for keeping a program running or holding a goroutine until a signal")
func Tip(message string) {
	if logger.level <= LevelInfo {
		_, err := cyanColor.Fprintf(output, "🌟 Tip: %s\n", message)
		if err != nil {
			return
		}
	}
}

// Tipf logs an informational message.
// Use Tipf to log general information about the program's execution.
//
// Example usage:
//
//	it.Tipf("Use select {} to create an infinite blocking loop without consuming CPU, %s", "useful for keeping a program running or holding a goroutine until a signal")
func Tipf(format string, args ...interface{}) {
	if logger.level <= LevelInfo {
		_, err := cyanColor.Fprintf(output, "🌟 Tip: "+format+"\n", args...)
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
		_, err := yellowColor.Fprintf(output, "🚧 Warn: %s\n", message)
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
		_, err := yellowColor.Fprintf(output, "🚧 Warn: "+format+"\n", args...)
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
		_, err := redColor.Fprintf(output, "❌ Error: %s\n", message)
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
		_, err := redColor.Fprintf(output, "❌ Error: "+format+"\n", args...)
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

var jsonBufferPool = sync.Pool{
	New: func() any { return new(json.Encoder) },
}

// StructuredLog logs a message with a specified level in a structured format.
// Use StructuredLog to log messages with additional data for any log level.
//
// Example usage:
//
//	it.StructuredLog("ERROR", "Failed to process request", map[string]interface{}{"requestID": "abc123", "error": err})
func StructuredLog(level string, message string, data any) {
	if shouldLogLevel(getLevelInt(level)) {
		entry := StructuredLogEntry{Level: level, Message: message, Data: data}
		buf := jsonBufferPool.Get().(*json.Encoder)
		defer jsonBufferPool.Put(buf)

		err := buf.Encode(entry)
		if err != nil {
			Errorf("Error encoding structured log: %v", err)
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
//	return it.WrapError(err, "failed to open file", ctx)
func WrapError(err error, message string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	if len(args) > 0 {
		message = fmt.Sprintf(message, args...)
	}
	return fmt.Errorf("%s: %w", message, err)
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
		_, err := fmt.Fprintf(output, "🏋️‍♂️ Audit: %s\n", message)
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
// Graceful Actions Area
// ===================================================

// GracefulShutdown listens for an interrupt signal (e.g., SIGINT or SIGTERM) and attempts to
// gracefully shut down the given server within the specified timeout. If an `action` function
// is provided, it will be executed after shutdown completes. If a `done` channel is provided,
// it will signal completion on the channel after shutdown and executing the action.
//
// Example usage:
//
//	it.GracefulShutdown(context.Background(), server, 5*time.Second, nil, nil)
//
// Or with a done channel and an action function:
//
//	done := make(chan bool)
//	cleanupAction := func() { log.Println("Performing post-shutdown cleanup...") }
//	go it.GracefulShutdown(context.Background(), server, 5*time.Second, done, cleanupAction)
//	<-done
func GracefulShutdown(ctx context.Context, server interface{ Shutdown(context.Context) error }, timeout time.Duration, done chan<- bool, action func()) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
		<-stop
		cancel()
	}()

	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, timeout)
	defer shutdownCancel()

	// Attempt shutdown
	if err := server.Shutdown(shutdownCtx); err != nil {
		Errorf("Error during shutdown: %v", err)
	} else {
		Info("Server shut down gracefully.")
	}

	if action != nil {
		action()
	} else {
		Warn("No post-shutdown action provided.")
	}

	if done != nil {
		done <- true
	}
}

// GracefulRestart listens for a signal to restart the server gracefully.
// It can be used with any server that has a Shutdown method.
//
// Example usage:
//
//	go it.GracefulRestart(context.Background(), server, 5*time.Second)
//
// GracefulRestart listens for a signal to restart the server gracefully. It attempts to
// shut down the given server within the specified timeout and then optionally performs an
// action before signaling completion on the `done` channel, if provided.
//
// Example usage:
//
//	it.GracefulRestart(context.Background(), server, 5*time.Second, nil, nil)
//
// Or with a done channel and an action function:
//
//	done := make(chan bool)
//	restartAction := func() { log.Println("Performing custom restart actions...") }
//	go it.GracefulRestart(context.Background(), server, 5*time.Second, done, restartAction)
//	<-done
func GracefulRestart(ctx context.Context, server interface{ Shutdown(context.Context) error }, timeout time.Duration, done chan<- bool, action func()) {
	restart := make(chan os.Signal, 1)
	signal.Notify(restart, os.Interrupt, syscall.SIGHUP) // Listen for interrupt and SIGHUP signals for restart

	go func() {
		<-restart
		Warn("Gracefully restarting...")

		// Create a context with the specified timeout for graceful shutdown
		shutdownCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		// Attempt to shut down the server gracefully
		if err := server.Shutdown(shutdownCtx); err != nil {
			Errorf("Error during shutdown for restart: %v", err)
		} else {
			Info("Server shut down gracefully.")
		}

		// Execute the optional action if provided
		if action != nil {
			action()
		} else {
			Warn("No post-shutdown action provided.")
		}

		// Notify the done channel, if provided
		if done != nil {
			done <- true
		}
	}()
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

// RetryWithContext retries a function while respecting context cancellation.
// It tries up to `attempts` times with a delay between each try and cancels if the context is done.
//
// Example usage:
//
//	err := it.RetryWithContext(ctx, 3, time.Second, SomeErrorOnlyFunction)
func RetryWithContext(ctx context.Context, attempts int, delay time.Duration, fn func() error) error {
	for i := 0; i < attempts; i++ {
		if err := fn(); err != nil {
			if i == attempts-1 {
				return err
			}
			select {
			case <-ctx.Done():
				return ctx.Err() // Return context error if it’s done
			case <-time.After(delay):
				// Retry after delay
			}
		} else {
			return nil
		}
	}
	return nil
}

// RetryExponentialWithContext retries the given function `fn` up to `attempts` times with an exponential
// backoff delay between attempts. The delay starts at `initialDelay` and doubles with each attempt.
// If the context is canceled, it stops retrying and returns the context error.
//
// Example usage:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
//	defer cancel()
//	err := it.RetryExponentialWithContext(ctx, 5, time.Second, SomeErrorOnlyFunction)
func RetryExponentialWithContext(ctx context.Context, attempts int, initialDelay time.Duration, fn func() error) error {
	delay := initialDelay
	for i := 0; i < attempts; i++ {
		// Check if the context is done before each attempt
		select {
		case <-ctx.Done():
			return ctx.Err() // Return the context error if it was canceled
		default:
		}

		// Try executing the function
		if err := fn(); err != nil {
			// If it's the last attempt, return the error
			if i == attempts-1 {
				return err
			}

			// Wait for the exponential delay or until the context is canceled
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
				// Double the delay for the next attempt
				delay *= 2
			}
		} else {
			// If successful, return nil
			return nil
		}
	}
	return nil
}

// ===================================================
// Rate Limiting Area
// ===================================================

// RateLimiter returns a rate-limited version of any handler function. The handler will only be allowed to
// execute once per specified `rate` interval. It supports handlers with any number of input arguments and return values.
//
// Example usage:
//
//	limitedHandler := it.RateLimiter(1*time.Second, handler)
//	result := limitedHandler(arg1, arg2)
func RateLimiter(rate time.Duration, fn interface{}) interface{} {
	ticker := time.NewTicker(rate)
	fnVal := reflect.ValueOf(fn)

	return reflect.MakeFunc(fnVal.Type(), func(args []reflect.Value) []reflect.Value {
		// Wait for the rate limit interval
		<-ticker.C
		// Call the original function and return its result
		return fnVal.Call(args)
	}).Interface()
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
	logOutput.Store(io.Writer(os.Stdout)) // Default to stdout
	output = logOutput.Load().(io.Writer)
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
func SetLogOutput(newOutput *os.File) {
	logOutput.Store(newOutput)
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
	atomic.StoreInt32(&logger.level, int32(level))
}

// ===================================================
// Utility Functions Area
// ===================================================

// GenerateSecret -> generates a random 32-bit secret key.
//
// Example usage:
//
//	key := it.GenerateSecret()
func GenerateSecret() string {
	bucket := make([]byte, 32)
	if _, err := rand.Read(bucket); err != nil {
		Errorf(err.Error())
	}
	return hex.EncodeToString(bucket)
}

// ===================================================
// Helper Functions Area
// ===================================================

// Helper function to check log level
func shouldLogLevel(level int32) bool {
	return atomic.LoadInt32(&logger.level) <= level
}

func getLevelInt(level string) int32 {
	switch level {
	case "TRACE":
		return LevelTrace
	case "DEBUG":
		return LevelDebug
	case "INFO":
		return LevelInfo
	case "WARN":
		return LevelWarning
	case "ERROR":
		return LevelError
	case "FATAL":
		return LevelFatal
	case "AUDIT":
		return LevelAudit
	default:
		return LevelInfo // Default to INFO if the level is unrecognized
	}
}
