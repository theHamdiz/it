package it_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/theHamdiz/it"
)

// TestRecoverPanicAndContinue tests panic recovery
func TestRecoverPanicAndContinue(t *testing.T) {
	defer it.RecoverPanicAndContinue()()

	recovered := false
	func() {
		defer func() {
			if r := recover(); r != nil {
				recovered = true
			}
		}()
		panic("test panic")
	}()

	if !recovered {
		t.Error("Expected panic to be recovered")
	}
}

// TestMust tests the Must function
func TestMust(t *testing.T) {
	// Test successful case
	result := it.Must(func() (string, error) {
		return "success", nil
	})
	if result != "success" {
		t.Errorf("Expected 'success', got %s", result)
	}

	// Test panic case
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected Must to panic on error")
		}
	}()

	it.Must(func() (string, error) {
		return "", errors.New("test error")
	})
}

// TestShould tests the Should function
func TestShould(t *testing.T) {
	// Test successful case
	result := it.Should(func() (string, error) {
		return "success", nil
	})
	if result != "success" {
		t.Errorf("Expected 'success', got %s", result)
	}

	// Test error case (should not panic)
	result = it.Should(func() (string, error) {
		return "default", errors.New("test error")
	})
	if result != "default" {
		t.Errorf("Expected 'default', got %s", result)
	}
}

// TestSafeGo tests goroutine safety
func TestSafeGo(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(1)

	executed := false
	it.SafeGo(func() {
		defer wg.Done()
		executed = true
	})

	wg.Wait()
	if !executed {
		t.Error("SafeGo didn't execute the function")
	}
}

// TestSafeGoWithContext tests context-aware goroutine safety
func TestSafeGoWithContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg := sync.WaitGroup{}
	wg.Add(1)

	executed := false
	it.SafeGoWithContext(ctx, func(ctx context.Context) {
		defer wg.Done()
		executed = true
	})

	wg.Wait()
	if !executed {
		t.Error("SafeGoWithContext didn't execute the function")
	}
}

// TestLogging tests various logging functions
func TestLogging(t *testing.T) {
	// Redirect logs to a temporary file for testing
	tmpFile, err := os.CreateTemp("", "log_test")
	if err != nil {
		t.Fatal(err)
	}
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
			t.Errorf("Failed to remove temp file %s: %v", name, err)
		}
	}(tmpFile.Name())

	it.SetLogOutput(tmpFile)

	tests := []struct {
		name    string
		logFunc func(string)
		message string
	}{
		{"Info", it.Info, "info message"},
		{"Warn", it.Warn, "warning message"},
		{"Error", it.Error, "error message"},
		{"Debug", it.Debug, "debug message"},
		{"Trace", it.Trace, "trace message"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.logFunc(tt.message)
			// Here you could read the temp file and verify the log message
		})
	}
}

// TestRetry tests retry functionality
func TestRetry(t *testing.T) {
	attempts := 0
	err := it.Retry(3, time.Millisecond, func() error {
		attempts++
		if attempts < 3 {
			return errors.New("not yet")
		}
		return nil
	})

	if err != nil {
		t.Error("Expected retry to succeed")
	}
	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}
}

type mockServer struct {
	shutdownCalled bool
	shutdownError  error
	mu             sync.Mutex
}

func (m *mockServer) Shutdown(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shutdownCalled = true
	return m.shutdownError
}

func (m *mockServer) WasShutdownCalled() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.shutdownCalled
}

// Improved test with multiple scenarios
func TestGracefulShutdown(t *testing.T) {
	testCases := []struct {
		name          string
		timeout       time.Duration
		shutdownError error
		action        func()
		setupTest     func()
		validateTest  func(*testing.T, *mockServer, bool)
	}{
		{
			name:    "successful shutdown",
			timeout: time.Second,
			setupTest: func() {
				// Small delay to ensure shutdown manager is ready
				time.Sleep(100 * time.Millisecond)
				_ = syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
			},
			validateTest: func(t *testing.T, server *mockServer, success bool) {
				if !success {
					t.Error("Graceful shutdown reported failure")
				}
				if !server.WasShutdownCalled() {
					t.Error("Server shutdown was not called")
				}
			},
		},
		{
			name:          "shutdown with error",
			timeout:       time.Second,
			shutdownError: errors.New("planned shutdown error"),
			setupTest: func() {
				time.Sleep(100 * time.Millisecond)
				_ = syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
			},
			validateTest: func(t *testing.T, server *mockServer, success bool) {
				if success {
					t.Error("Graceful shutdown should have reported failure")
				}
				if !server.WasShutdownCalled() {
					t.Error("Server shutdown was not called")
				}
			},
		},
		{
			name:    "shutdown with action",
			timeout: time.Second,
			action: func() {
			},
			setupTest: func() {
				time.Sleep(100 * time.Millisecond)
				_ = syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
			},
			validateTest: func(t *testing.T, server *mockServer, success bool) {
				if !success {
					t.Error("Graceful shutdown reported failure")
				}
				if !server.WasShutdownCalled() {
					t.Error("Server shutdown was not called")
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a new server for each test
			server := &mockServer{shutdownError: tc.shutdownError}
			done := make(chan bool)
			ctx, cancel := context.WithTimeout(context.Background(), tc.timeout)
			defer cancel()

			// Start graceful shutdown in a goroutine
			go func() {
				it.GracefulShutdown(ctx, server, tc.timeout, done, tc.action)
			}()

			// Setup test conditions
			if tc.setupTest != nil {
				tc.setupTest()
			}

			// Wait for completion with timeout
			var success bool
			select {
			case success = <-done:
				// Shutdown completed
			case <-time.After(tc.timeout * 2):
				t.Fatal("Test timed out waiting for shutdown")
			}

			// Validate test results
			if tc.validateTest != nil {
				tc.validateTest(t, server, success)
			}
		})
	}
}

// TestWaitFor tests the WaitFor function
func TestWaitFor(t *testing.T) {
	// Test successful wait
	success := it.WaitFor(time.Millisecond*100, func() bool {
		return true
	})
	if !success {
		t.Error("WaitFor should return true for immediate condition")
	}

	// Test timeout
	start := time.Now()
	success = it.WaitFor(time.Millisecond*100, func() bool {
		return false
	})
	duration := time.Since(start)

	if success {
		t.Error("WaitFor should return false on timeout")
	}
	if duration < time.Millisecond*100 {
		t.Error("WaitFor didn't wait for the full timeout duration")
	}
}

// TestGenerateSecret tests secret generation
func TestGenerateSecret(t *testing.T) {
	secret1 := it.GenerateSecret()
	secret2 := it.GenerateSecret()

	if secret1 == "" {
		t.Error("GenerateSecret returned empty string")
	}
	if secret1 == secret2 {
		t.Error("Generated secrets should be different")
	}
}

// TestStructuredLogging tests structured logging functionality
func TestStructuredLogging(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "structured_log_test")
	if err != nil {
		t.Fatal(err)
	}
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
			t.Errorf("Failed to remove temp file %s: %v", name, err)
		}
	}(tmpFile.Name())

	it.SetLogOutput(tmpFile)

	testCases := []struct {
		name    string
		logFunc func(string, map[string]any)
		message string
		data    map[string]any
	}{
		{
			name:    "StructuredInfo",
			logFunc: it.StructuredInfo,
			message: "test info",
			data: map[string]any{
				"key": "value",
				"num": 123,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.logFunc(tc.message, tc.data)
			// Add verification of log output if needed
		})
	}
}

// TestRetryExponentialWithContext tests exponential backoff retry with context
func TestRetryExponentialWithContext(t *testing.T) {
	testCases := []struct {
		name             string
		attempts         int
		initialDelay     time.Duration
		operation        func(int *int) error
		expectedError    string
		minDuration      time.Duration
		maxDuration      time.Duration
		expectedAttempts int
		setupContext     func(context.Context, context.CancelFunc) // Added setup function
	}{
		{
			name:         "basic retry with failure",
			attempts:     3,
			initialDelay: time.Millisecond * 10,
			operation: func(attempts *int) error {
				*attempts++
				return errors.New("persistent error")
			},
			expectedError:    "persistent error",
			minDuration:      time.Millisecond * 30,
			maxDuration:      time.Millisecond * 50,
			expectedAttempts: 3,
			setupContext:     nil, // No special setup needed
		},
		{
			name:         "success on second attempt",
			attempts:     3,
			initialDelay: time.Millisecond * 10,
			operation: func(attempts *int) error {
				*attempts++
				if *attempts < 2 {
					return errors.New("temporary error")
				}
				return nil
			},
			expectedError:    "",
			minDuration:      time.Millisecond * 10,
			maxDuration:      time.Millisecond * 30,
			expectedAttempts: 2,
			setupContext:     nil,
		},
		{
			name:         "context cancellation",
			attempts:     3,
			initialDelay: time.Millisecond * 10,
			operation: func(attempts *int) error {
				*attempts++
				time.Sleep(time.Millisecond * 5) // Small delay to ensure context cancellation
				return errors.New("should be cancelled")
			},
			expectedError:    context.Canceled.Error(),
			minDuration:      time.Millisecond * 0,
			maxDuration:      time.Millisecond * 20,
			expectedAttempts: 1,
			setupContext: func(ctx context.Context, cancel context.CancelFunc) {
				// Cancel after a short delay to allow first attempt
				go func() {
					time.Sleep(time.Millisecond)
					cancel()
				}()
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			attempts := 0
			start := time.Now()

			ctx, cancel := context.WithTimeout(context.Background(), tc.maxDuration)
			defer cancel()

			if tc.setupContext != nil {
				tc.setupContext(ctx, cancel)
			}

			err := it.RetryExponentialWithContext(ctx, tc.attempts, tc.initialDelay, func() error {
				return tc.operation(&attempts)
			})

			duration := time.Since(start)

			// Check error
			if tc.expectedError == "" {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
			} else {
				if err == nil || !strings.Contains(err.Error(), tc.expectedError) {
					t.Errorf("Expected error containing '%s', got: %v", tc.expectedError, err)
				}
			}

			// Check attempts
			if attempts != tc.expectedAttempts {
				t.Errorf("Expected %d attempts, got %d", tc.expectedAttempts, attempts)
			}

			// Check duration with some tolerance
			if duration < tc.minDuration {
				t.Errorf("Retries happened too quickly. Expected minimum %v, got %v", tc.minDuration, duration)
			}
			if duration > tc.maxDuration {
				t.Errorf("Retries took too long. Expected maximum %v, got %v", tc.maxDuration, duration)
			}
		})
	}
}

// TestTimeParallel tests parallel execution timing
func TestTimeParallel(t *testing.T) {
	durations := it.TimeParallel("parallel_test",
		func() { time.Sleep(time.Millisecond * 50) },
		func() { time.Sleep(time.Millisecond * 100) },
		func() { time.Sleep(time.Millisecond * 75) },
	)

	if len(durations) != 3 {
		t.Errorf("Expected 3 durations, got %d", len(durations))
	}

	// Verify that durations are reasonable
	for i, d := range durations {
		if d < time.Millisecond*50 {
			t.Errorf("Duration %d too short: %v", i, d)
		}
	}
}

// TestRateLimiterWithHTTPRequests tests rate limiting with HTTP requests
func TestRateLimiterWithHTTPRequests(t *testing.T) {
	// Create a test server
	requestCount := int32(0)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create a rate-limited HTTP client
	makeRequest := func() error {
		resp, err := http.Get(server.URL)
		if err != nil {
			return err
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				t.Errorf("Failed to close response body: %v", err)
			}
		}(resp.Body)
		return nil
	}

	rateLimitedRequest := it.RateLimiter(time.Millisecond*100, makeRequest).(func() error)

	// Make several requests
	start := time.Now()
	for i := 0; i < 5; i++ {
		err := rateLimitedRequest()
		if err != nil {
			t.Errorf("Request %d failed: %v", i, err)
		}
	}

	duration := time.Since(start)
	if duration < time.Millisecond*400 {
		t.Error("Requests were not rate limited properly")
	}
}

// TestTimeFunctionWithCallback tests timing with callback
func TestTimeFunctionWithCallback(t *testing.T) {
	var measured time.Duration
	callback := func(duration time.Duration) {
		measured = duration
	}

	result := it.TimeFunctionWithCallback(
		"test_function",
		func() string {
			time.Sleep(time.Millisecond * 50)
			return "done"
		},
		callback,
	)

	if result != "done" {
		t.Error("Function didn't return expected result")
	}
	if measured < time.Millisecond*50 {
		t.Error("Measured duration too short")
	}
}

// TestConfigurationChanges tests configuration modifications
func TestConfigurationChanges(t *testing.T) {
	// Test color output configuration
	it.EnableColoredOutput(true)
	config := it.GetCurrentConfig()
	if !config.ColorsEnabled() {
		t.Error("Color output not enabled")
	}

	// Test shutdown timeout configuration
	timeout := time.Second * 5
	it.SetShutdownTimeout(timeout)
	config = it.GetCurrentConfig()
	if config.ShutdownTimeout != timeout {
		t.Error("Shutdown timeout not set correctly")
	}
}

// TestEnvironmentVariables tests environment variable handling
func TestEnvironmentVariables(t *testing.T) {
	// Save original env vars
	originalLogLevel := os.Getenv("LOG_LEVEL")
	originalLogFile := os.Getenv("LOG_FILE")
	defer func() {
		_ = os.Setenv("LOG_LEVEL", originalLogLevel)
		_ = os.Setenv("LOG_FILE", originalLogFile)
	}()

	// Test LOG_LEVEL
	_ = os.Setenv("LOG_LEVEL", "DEBUG")
	it.InitFromEnv()
	// Add verification of log level change

	// Test LOG_FILE
	tmpFile, err := os.CreateTemp("", "log_test")
	if err != nil {
		t.Fatal(err)
	}
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
			t.Errorf("Failed to remove temp file %s: %v", name, err)
		}
	}(tmpFile.Name())

	_ = os.Setenv("LOG_FILE", tmpFile.Name())
	it.InitFromEnv()
	// Add verification of log file change
}

// TestLogOnce ensures messages are logged only once
func TestLogOnce(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "log_once_test")
	if err != nil {
		t.Fatal(err)
	}
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
			t.Errorf("Failed to remove temp file %s: %v", name, err)
		}
	}(tmpFile.Name())

	it.SetLogOutput(tmpFile)

	message := "this should appear once"
	for i := 0; i < 3; i++ {
		it.LogOnce(message)
	}

	// Read log file and verify message appears only once
	content, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}

	occurrences := 0
	for _, line := range strings.Split(string(content), "\n") {
		if strings.Contains(line, message) {
			occurrences++
		}
	}

	if occurrences != 1 {
		t.Errorf("Expected message to appear once, got %d occurrences", occurrences)
	}
}

// TestAuditLogging tests audit log functionality
func TestAuditLogging(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "audit_log_test")
	if err != nil {
		t.Fatal(err)
	}
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
			t.Errorf("Failed to remove temp file %s: %v", name, err)
		}
	}(tmpFile.Name())

	it.SetLogOutput(tmpFile)

	auditMessage := "sensitive operation performed"
	it.Audit(auditMessage)

	// Verify audit log format and content
	content, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(content), "AUDIT") || !strings.Contains(string(content), auditMessage) {
		t.Error("Audit log entry not formatted correctly")
	}
}

// TestTimeBlock measures block execution time
func TestTimeBlock(t *testing.T) {
	done := it.TimeBlock("test_block")
	time.Sleep(time.Millisecond * 50)
	done()

	// Note: This test might need adjustment based on how TimeBlock logs its output
}

// TestConcurrentLogging tests thread-safety of logging
func TestConcurrentLogging(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "concurrent_log_test")
	if err != nil {
		t.Fatal(err)
	}
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
			t.Errorf("Failed to remove temp file %s: %v", name, err)
		}
	}(tmpFile.Name())

	it.SetLogOutput(tmpFile)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			it.Infof("Concurrent log %d", id)
		}(i)
	}

	wg.Wait()
}

// BenchmarkRateLimiter benchmarks rate limiting performance
func BenchmarkRateLimiter(b *testing.B) {
	operation := func() error {
		return nil
	}
	rateLimitedOp := it.RateLimiter(time.Microsecond, operation).(func() error)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rateLimitedOp()
	}
}

// BenchmarkStructuredLogging benchmarks structured logging performance
func BenchmarkStructuredLogging(b *testing.B) {
	data := map[string]any{
		"key1": "value1",
		"key2": 123,
		"key3": true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		it.StructuredInfo("benchmark message", data)
	}
}
