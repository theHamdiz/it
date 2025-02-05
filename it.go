/*
Package it - Because writing good Go code shouldn't be harder than it already is.

This package provides utilities that you probably should have written yourself but
didn't have time for. It handles all those annoying little details that make
production code actually work in the real world, like logging (because print statements
aren't cutting it anymore), error handling (because panic/recover is so 2009), and
graceful shutdowns (because Ctrl+C isn't a valid shutdown strategy).

If you're looking for the underlying implementations that aren't wrapped in cotton wool,
check out the following sub-packages:

  - logger: For when fmt.Println isn't providing enough insight into your problems
  - retry: Because the first try is rarely the charm
  - rl: Rate limiting, because your APIs have feelings too
  - sm: Shutdown management, for when you actually care about cleanup
  - tk: Time keeping, because time.Now().Sub(start) gets old
  - cfg: Configuration, because hardcoding values is for amateurs

For those who prefer to live dangerously and have complete control, feel free to
use these packages directly. For everyone else, the simplified interfaces in this
package should keep you out of trouble.

Example usage:
    it.Must(ExpensiveOperation())  // Because YOLO
    it.Should(RiskyOperation())    // For the risk-averse
    it.Info("It worked!")         // Shocking, I know
*/

package it

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"os"
	"reflect"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/theHamdiz/it/cfg"
	"github.com/theHamdiz/it/logger"
	"github.com/theHamdiz/it/retry"
	"github.com/theHamdiz/it/rl"
	"github.com/theHamdiz/it/sm"
	"github.com/theHamdiz/it/tk"
)

// ===================================================
// Types & Variables - The Usual Suspects
// ===================================================

var (
	currentConfig *cfg.Config
)

// ===================================================
// Error Handling - Because Errors Are Not Optional
// ===================================================

// RecoverPanicAndContinue wraps your questionable code in a safety blanket
func RecoverPanicAndContinue() func() {
	return func() {
		if r := recover(); r != nil {
			logger.DefaultLogger().Errorf("Recovered from panic: %v\n%s", r, debug.Stack())
		}
	}
}

// Must is for when failure is not an option
// (or when you're feeling particularly optimistic)
func Must[T any](operation func() (T, error)) T {
	cfg_ := retry.DefaultRetryConfig()
	result, err := retry.WithBackoff(context.Background(), cfg_,
		func(ctx context.Context) (T, error) {
			return operation()
		})
	if err != nil {
		panic(fmt.Sprintf("all retries failed: %v", err))
	}
	return result
}

// Should is for when you care about errors, but not enough to handle them properly,
// because not every error is worth dying for!!
func Should[T any](operation func() (T, error)) T {
	var lastResult T
	cfg_ := retry.DefaultRetryConfig()
	result, err := retry.WithBackoff(context.Background(), cfg_,
		func(ctx context.Context) (T, error) {
			res, err := operation()
			if err != nil {
				// Save the last result even if there's an error
				lastResult = res
			}
			return res, err
		})
	if err != nil {
		logger.DefaultLogger().Error(fmt.Sprintf("operation failed after retries: %v", err))
		// Return the last result if all retries failed
		return lastResult
	}
	return result
}

// Could is for when you're not sure if you want to deal with this right now
func Could[T any](operation func() (T, error)) func() T {
	var (
		result T
		err    error
		done   bool
		mu     sync.Mutex
	)

	return func() T {
		mu.Lock()
		defer mu.Unlock()

		if !done {
			result, err = operation()
			if err != nil {
				logger.DefaultLogger().Warnf("well, we tried %s", err)
			}
			done = true
		}
		return result
	}
}

// Might is for when success is optional but you'd like to know about it
// Returns (value, true) if it worked, (zero, false) if it didn't
func Might[T any](operation func() (T, error)) (T, bool) {
	result, err := operation()
	if err != nil {
		logger.DefaultLogger().Debugf("it didn't work out: %v", err)
		var zero T
		return zero, false
	}
	return result, true
}

// WrapError wraps an error with a custom message and additional metadata.
// If the original error is nil, it simply returns nil.
func WrapError(err error, message string, metadata map[string]any) error {
	if err == nil {
		return nil
	}

	wrappedMessage := fmt.Sprintf("%s: %v", message, err)
	if metadata != nil && len(metadata) > 0 {
		wrappedMessage += fmt.Sprintf(" | Metadata: %+v", metadata)
	}

	return errors.New(wrappedMessage)
}

// ===================================================
// Goroutine Management - Threading The Needle
// ===================================================

// SafeGo -> runs a function in a goroutine with panic recovery
// because letting goroutines crash and burn is so 2010
func SafeGo(fn func()) {
	go func() {
		defer RecoverPanicAndContinue()()
		fn()
	}()
}

// SafeGoWithContext -> runs a function in a goroutine with context
// and panic recovery because context is king
func SafeGoWithContext(ctx context.Context, fn func(context.Context)) {
	go func() {
		defer RecoverPanicAndContinue()()
		fn(ctx)
	}()
}

// ===================================================
// Logging - Because println() Is Not A Logging Strategy
// ===================================================

// Info Basic Information Logging
func Info(msg string) {
	logger.DefaultLogger().Info(msg)
}

func Warn(msg string) {
	logger.DefaultLogger().Warn(msg)
}

func Error(msg string) {
	logger.DefaultLogger().Error(msg)
}

// Infof Formatted Logging of information
func Infof(format string, args ...interface{}) {
	logger.DefaultLogger().Infof(format, args...)
}

func Warnf(format string, args ...interface{}) {
	logger.DefaultLogger().Warnf(format, args...)
}

func Errorf(format string, args ...interface{}) {
	logger.DefaultLogger().Errorf(format, args...)
}

// SetLogLevel -> Change the log level of the default logger
func SetLogLevel(level logger.LogLevel) {
	logger.SetLogLevel(level)
}

func Debug(msg string) {
	logger.DefaultLogger().Debug(msg)
}

func Trace(msg string) {
	logger.DefaultLogger().Trace(msg)
}

// LogStackTrace Stack Trace Logging
func LogStackTrace() {
	buf := make([]byte, 1024)
	n := runtime.Stack(buf, false)
	logger.DefaultLogger().Info(string(buf[:n]))
}

func LogErrorWithStack(err error) {
	if err == nil {
		return
	}
	buf := make([]byte, 1024)
	n := runtime.Stack(buf, false)
	logger.DefaultLogger().Errorf("Error: %v\nStack Trace:\n%s", err, string(buf[:n]))
}

func LogOnce(msg string) {
	logger.DefaultLogger().LogOnce(msg)
}

// Audit logging
func Audit(msg string) {
	logger.DefaultLogger().StructuredLog(logger.LevelAudit, msg, nil)
}

// StructuredInfo Structured Logging of information.
func StructuredInfo(message string, data map[string]any) {
	logger.DefaultLogger().StructuredInfo(message, data)
}

func StructuredDebug(message string, data any) {
	if m, ok := data.(map[string]any); ok {
		logger.DefaultLogger().StructuredDebug(message, m)
	} else {
		logger.DefaultLogger().StructuredDebug(message, map[string]any{"data": data})
	}
}

func StructuredWarning(message string, data any) {
	if m, ok := data.(map[string]any); ok {
		logger.DefaultLogger().StructuredLog(logger.LevelWarning, message, m)
	} else {
		logger.DefaultLogger().StructuredLog(logger.LevelWarning, message, map[string]any{"data": data})
	}
}

func StructuredError(message string, data any) {
	if m, ok := data.(map[string]any); ok {
		logger.DefaultLogger().StructuredError(message, m)
	} else {
		logger.DefaultLogger().StructuredError(message, map[string]any{"data": data})
	}
}

// DeferWithLog returns a function that logs a message when executed
func DeferWithLog(message string) func() {
	return func() {
		logger.DefaultLogger().Info(message)
	}
}

// SetLogOutput redirects logs to the specified output
func SetLogOutput(writer *os.File) {
	logger.SetLogOutput(writer)
}

// ===================================================
// Retrials - It's all about trying again!
// ===================================================

// Retry retries a function with a fixed delay
func Retry(attempts int, delay time.Duration, operation func() error) error {
	// Create a context that will be used for retries
	ctx := context.Background()

	// Convert the simple operation to one that accepts context
	contextOperation := func(ctx context.Context) (interface{}, error) {
		return nil, operation()
	}

	// Create config for fixed delay retries
	config := retry.Config{
		Attempts:     attempts,
		InitialDelay: delay,
		MaxDelay:     delay, // Keep delay fixed
		Multiplier:   1.0,   // No multiplication
		RandomFactor: 0.0,   // No jitter
	}

	_, err := retry.WithBackoff(ctx, config, contextOperation)
	return err
}

// RetryExponential retries a function with exponential backoff
func RetryExponential(attempts int, initialDelay time.Duration, operation func() error) error {
	ctx := context.Background()

	contextOperation := func(ctx context.Context) (interface{}, error) {
		return nil, operation()
	}

	config := retry.Config{
		Attempts:     attempts,
		InitialDelay: initialDelay,
		MaxDelay:     initialDelay * time.Duration(1<<uint(attempts)), // Max delay based on attempts
		Multiplier:   2.0,
		RandomFactor: 0.1,
	}

	_, err := retry.WithBackoff(ctx, config, contextOperation)
	return err
}

// RetryWithContext retries a function with a fixed delay, respecting context cancellation
func RetryWithContext(ctx context.Context, attempts int, delay time.Duration, operation func() error) error {
	contextOperation := func(ctx context.Context) (interface{}, error) {
		return nil, operation()
	}

	config := retry.Config{
		Attempts:     attempts,
		InitialDelay: delay,
		MaxDelay:     delay, // Keep delay fixed
		Multiplier:   1.0,   // No multiplication
		RandomFactor: 0.0,   // No jitter
	}

	_, err := retry.WithBackoff(ctx, config, contextOperation)
	return err
}

// RetryExponentialWithContext retries a function with exponential backoff, respecting context cancellation
func RetryExponentialWithContext(ctx context.Context, attempts int, initialDelay time.Duration, operation func() error) error {
	if attempts <= 0 {
		return errors.New("attempts must be greater than 0")
	}
	if initialDelay <= 0 {
		return errors.New("initial delay must be greater than 0")
	}

	var lastErr error
	for i := 0; i < attempts; i++ {
		// Create a done channel for the operation
		done := make(chan error, 1)

		// Run the operation in a goroutine
		go func() {
			done <- operation()
		}()

		// Wait for either the operation to complete or context to be cancelled
		select {
		case err := <-done:
			if err == nil {
				return nil // Success
			}
			lastErr = err
		case <-ctx.Done():
			return ctx.Err()
		}

		// Don't sleep after the last attempt
		if i == attempts-1 {
			break
		}

		// Calculate delay: initialDelay * (2^i)
		delay := initialDelay * time.Duration(1<<uint(i))

		// Wait for either the delay or context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
			continue
		}
	}

	return lastErr
}

// ===================================================
// Graceful Operations - Because SIGKILL Is Not The Answer
// ===================================================

// GracefulShutdown performs a graceful shutdown on the provided server.
// The server parameter can implement Shutdown with either signature:
//   - Shutdown(context.Context) error
//   - Shutdown() error
func GracefulShutdown(
	ctx context.Context,
	server interface{},
	timeout time.Duration,
	done chan<- bool,
	action func(),
) {
	// Create shutdown manager with SIGINT and SIGTERM.
	manager := sm.NewShutdownManager(syscall.SIGINT, syscall.SIGTERM)

	// Create a context with timeout for shutdown operations.
	shutdownCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Add server shutdown as a critical action using the reflection helper.
	manager.AddAction(
		"server-shutdown",
		func(actionCtx context.Context) error {
			return callShutdown(server, actionCtx)
		},
		timeout,
		true, // Critical action
	)

	// If a post-shutdown action is provided, add it as non-critical.
	if action != nil {
		manager.AddAction(
			"post-shutdown-action",
			func(actionCtx context.Context) error {
				action()
				return nil
			},
			timeout,
			false, // Non-critical action
		)
	}

	// Start the shutdown manager.
	manager.Start()

	// Wait for shutdown to complete or for the timeout.
	errChan := make(chan error, 1)
	go func() {
		errChan <- manager.Wait()
	}()

	var err error
	select {
	case err = <-errChan:
		// Shutdown completed.
	case <-shutdownCtx.Done():
		err = shutdownCtx.Err()
	}

	// Signal completion if a done channel was provided.
	if done != nil {
		done <- err == nil
		close(done)
	}
}

// GracefulRestart performs a graceful restart on the provided server.
// The server parameter can implement Shutdown with either signature:
//   - Shutdown(context.Context) error
//   - Shutdown() error
func GracefulRestart(
	ctx context.Context,
	server interface{},
	timeout time.Duration,
	done chan<- bool,
	action func(),
) {
	// Create shutdown manager with SIGHUP (for restart).
	manager := sm.NewShutdownManager(syscall.SIGHUP)

	// Add server shutdown as a critical action.
	manager.AddAction(
		"server-shutdown",
		func(actionCtx context.Context) error {
			return callShutdown(server, actionCtx)
		},
		timeout,
		true, // Critical action
	)

	// If a restart action is provided, add it as critical.
	if action != nil {
		manager.AddAction(
			"restart-action",
			func(actionCtx context.Context) error {
				action()
				return nil
			},
			timeout,
			true, // Critical for restart
		)
	}

	// Start the shutdown manager.
	manager.Start()

	// Wait for restart to complete.
	err := manager.Wait()

	// Signal completion if a done channel was provided.
	if done != nil {
		done <- err == nil
		close(done)
	}
}

// ===================================================
// Rate Limiting - Your Infrastructure Will Thank You
// ===================================================

// RateLimiter wraps any function with rate limiting capability
func RateLimiter(rate time.Duration, fn interface{}) interface{} {
	// Create a rate limiter with batch size 1 for simple function rate limiting
	rateLimiter := rl.NewRateLimiter(rate, 1)

	// Get the type of the function
	fnValue := reflect.ValueOf(fn)
	fnType := fnValue.Type()

	// Create a new function with the same signature
	wrappedFn := reflect.MakeFunc(fnType, func(args []reflect.Value) []reflect.Value {
		// Create a context for this execution
		ctx := context.Background()

		// Create a function that will execute the original function with its arguments
		operation := func() ([]reflect.Value, error) {
			return fnValue.Call(args), nil
		}

		// Execute the operation with rate limiting
		result, err := rl.ExecuteRateLimited(rateLimiter, ctx, func() ([]reflect.Value, error) {
			return operation()
		})

		if err != nil {
			// In case of error, return zero values for all return types
			zeroValues := make([]reflect.Value, fnType.NumOut())
			for i := range zeroValues {
				zeroValues[i] = reflect.Zero(fnType.Out(i))
			}
			return zeroValues
		}

		return result
	})

	return wrappedFn.Interface()
}

// WaitFor waits for a condition to be met or times out
func WaitFor(timeout time.Duration, condition func() bool) bool {
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	// Check every 100ms
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		if condition() {
			return true
		}

		select {
		case <-timer.C:
			return false
		case <-ticker.C:
			continue
		}
	}
}

// ===================================================
// Timing & Measurement - Time Is Money, Friend
// ===================================================

// TimeFunction measures and logs the execution time of a function
func TimeFunction[T any](name string, fn func() T) T {
	return tk.TimeFn(name, fn)
}

// TimeBlock starts a timer and returns a function that logs the execution time when called
func TimeBlock(name string) func() {
	timekeeper := tk.NewTimeKeeper(name).Start()
	return func() {
		timekeeper.Stop()
	}
}

// TimeFunctionWithCallback measures execution time and calls a callback with the duration
func TimeFunctionWithCallback[T any](
	name string,
	fn func() T,
	callback func(duration time.Duration),
) T {
	timekeeper := tk.NewTimeKeeper(name, tk.WithCallback(callback)).Start()
	defer timekeeper.Stop()
	return fn()
}

// TimeParallel measures execution time of parallel operations
func TimeParallel(name string, fns ...func()) []time.Duration {
	asyncTimer := tk.NewAsyncTimeKeeper(name)

	for _, fn := range fns {
		asyncTimer.Track(fn)
	}

	return asyncTimer.Wait()
}

// ===================================================
// Utility Functions - The Kitchen Sink
// ===================================================

// GenerateSecret generates a random secret of the given byte length.
func GenerateSecret(numBytes int) string {
	bytes := make([]byte, numBytes)

	if _, err := rand.Read(bytes); err != nil {
		// When random generation fails, fallback to time-based generation.
		// The below approach has very low entropy and is completely insecure.
		// However, crypto/rand fails very rarely, so who cares.
		bytesWritten := 0
		for bytesWritten < numBytes {
			byteFromCurrentTime := byte(time.Now().UnixNano() & 0xFF)
			bytes[bytesWritten] = byteFromCurrentTime
			bytesWritten++
		}
	}

	// Convert to hex string
	return hex.EncodeToString(bytes)
}

// =======================================================
// Configuration - Making Things Configurable Since 2025
// =======================================================

// InitFromEnv initializes logger settings from environment variables
func InitFromEnv() {
	// Handle LOG_LEVEL
	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		level := parseLogLevel(logLevel)
		if level != logger.LevelInfo { // Only change if valid level found
			SetLogLevel(level)
		}
	}

	// Handle LOG_FILE
	if logFile := os.Getenv("LOG_FILE"); logFile != "" {
		file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err == nil {
			SetLogOutput(file)
		}
	}
}

// ConfigureLogger provides more detailed configuration options
func ConfigureLogger(options ...cfg.ConfigOption) {
	currentConfig = cfg.Configure(options...)
}

// GetCurrentConfig returns the current configuration
func GetCurrentConfig() *cfg.Config {
	return currentConfig
}

// EnableColoredOutput enables or disables colored output
func EnableColoredOutput(enabled bool) {
	currentConfig = cfg.Configure(cfg.WithColors(enabled))
}

// SetShutdownTimeout sets the global shutdown timeout
func SetShutdownTimeout(timeout time.Duration) {
	currentConfig = cfg.Configure(cfg.WithShutdownTimeout(timeout))
}

// ===================================================
// Internal Helpers - Nothing To See Here
// ===================================================

func parseLogLevel(level string) logger.LogLevel {
	switch strings.ToUpper(level) {
	case "TRACE":
		return logger.LevelTrace
	case "DEBUG":
		return logger.LevelDebug
	case "INFO":
		return logger.LevelInfo
	case "WARN", "WARNING":
		return logger.LevelWarning
	case "ERROR":
		return logger.LevelError
	case "FATAL":
		return logger.LevelFatal
	default:
		return logger.LevelInfo
	}
}

// callShutdown is a helper that calls the Shutdown method on a server instance.
// It supports both signatures:
//   - Shutdown(context.Context) error
//   - Shutdown() error
func callShutdown(server interface{}, ctx context.Context) error {
	v := reflect.ValueOf(server)
	method := v.MethodByName("Shutdown")
	if !method.IsValid() {
		return fmt.Errorf("server does not implement a Shutdown method")
	}

	methodType := method.Type()
	var args []reflect.Value
	switch methodType.NumIn() {
	case 0:
		// Signature: Shutdown() error
		args = []reflect.Value{}
	case 1:
		// Signature: Shutdown(context.Context) error (or similar)
		paramType := methodType.In(0)
		if !reflect.TypeOf(ctx).AssignableTo(paramType) {
			return fmt.Errorf("Shutdown method expects parameter of type %v; context.Context is not assignable", paramType)
		}
		args = []reflect.Value{reflect.ValueOf(ctx)}
	default:
		return fmt.Errorf("Shutdown method has unsupported number of parameters: %d", methodType.NumIn())
	}

	results := method.Call(args)
	// If no return values, assume success.
	if len(results) == 0 {
		return nil
	}
	// If one result and it's an error, return it.
	if len(results) == 1 {
		if err, ok := results[0].Interface().(error); ok {
			return err
		}
		return nil
	}
	// If multiple return values, assume the last one is error.
	last := results[len(results)-1]
	if err, ok := last.Interface().(error); ok {
		return err
	}
	return nil
}

func init() {
	currentConfig = cfg.Configure()
	// Because someone has to set sensible defaults
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}
