package retry_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/theHamdiz/it/retry"
)

// TestDefaultRetryConfig ensures the default configuration values are correct.
func TestDefaultRetryConfig(t *testing.T) {
	config := retry.DefaultRetryConfig()

	if config.Attempts != 3 {
		t.Errorf("Expected Attempts to be 3, got %d", config.Attempts)
	}
	if config.InitialDelay != 100*time.Millisecond {
		t.Errorf("Expected InitialDelay to be 100ms, got %v", config.InitialDelay)
	}
	if config.MaxDelay != 10*time.Second {
		t.Errorf("Expected MaxDelay to be 10s, got %v", config.MaxDelay)
	}
	if config.Multiplier != 2.0 {
		t.Errorf("Expected Multiplier to be 2.0, got %f", config.Multiplier)
	}
	if config.RandomFactor != 0.1 {
		t.Errorf("Expected RandomFactor to be 0.1, got %f", config.RandomFactor)
	}
}

// TestRetryWithBackoff_Success ensures that an operation that succeeds on the first attempt returns immediately.
func TestRetryWithBackoff_Success(t *testing.T) {
	config := retry.DefaultRetryConfig()
	ctx := context.Background()

	operation := func(ctx context.Context) (string, error) {
		return "success", nil
	}

	result, err := retry.WithBackoff(ctx, config, operation)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result != "success" {
		t.Errorf("Expected 'success', got '%s'", result)
	}
}

// TestRetryWithBackoff_SuccessAfterRetry ensures that an operation succeeds after a few failed attempts.
func TestRetryWithBackoff_SuccessAfterRetry(t *testing.T) {
	config := retry.DefaultRetryConfig()
	ctx := context.Background()

	attempts := 0
	operation := func(ctx context.Context) (string, error) {
		attempts++
		if attempts < 3 {
			return "", errors.New("temporary error")
		}
		return "success", nil
	}

	result, err := retry.WithBackoff(ctx, config, operation)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result != "success" {
		t.Errorf("Expected 'success', got '%s'", result)
	}
	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}
}

// TestRetryWithBackoff_Fail ensures that an operation fails after exhausting all attempts.
func TestRetryWithBackoff_Fail(t *testing.T) {
	config := retry.DefaultRetryConfig()
	ctx := context.Background()

	operation := func(ctx context.Context) (string, error) {
		return "", errors.New("permanent error")
	}

	result, err := retry.WithBackoff(ctx, config, operation)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
	if result != "" {
		t.Errorf("Expected empty result, got '%s'", result)
	}
}

// TestRetryWithBackoff_ContextCancelled ensures that the function respects context cancellation.
func TestRetryWithBackoff_ContextCancelled(t *testing.T) {
	config := retry.DefaultRetryConfig()
	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()

	operation := func(ctx context.Context) (string, error) {
		time.Sleep(100 * time.Millisecond) // Simulate slow operation
		return "", errors.New("temporary error")
	}

	result, err := retry.WithBackoff(ctx, config, operation)
	if !errors.Is(err, context.DeadlineExceeded) && err != nil {
		t.Errorf("Expected context.DeadlineExceeded, got %v", err)
	}
	if result != "" {
		t.Errorf("Expected empty result, got '%s'", result)
	}
}

// TestRetryWithBackoff_RespectsMaxDelay ensures that delays do not exceed MaxDelay.
func TestRetryWithBackoff_RespectsMaxDelay(t *testing.T) {
	config := retry.Config{
		Attempts:     5,
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     250 * time.Millisecond,
		Multiplier:   2.0,
		RandomFactor: 0.0,
	}
	ctx := context.Background()

	attempts := 0
	operation := func(ctx context.Context) (string, error) {
		attempts++
		return "", errors.New("temporary error")
	}

	start := time.Now()
	_, _ = retry.WithBackoff(ctx, config, operation)
	duration := time.Since(start)

	expectedMaxDuration := 100*time.Millisecond + 200*time.Millisecond + 250*time.Millisecond + 250*time.Millisecond + 250*time.Millisecond
	if duration > expectedMaxDuration {
		t.Errorf("Expected duration <= %v, got %v", expectedMaxDuration, duration)
	}
}

// TestRetryWithBackoff_NoRetries ensures that when Attempts = 1, no retries occur.
func TestRetryWithBackoff_NoRetries(t *testing.T) {
	config := retry.Config{
		Attempts:     1,
		InitialDelay: 500 * time.Millisecond,
		MaxDelay:     5 * time.Second,
		Multiplier:   2.0,
		RandomFactor: 0.1,
	}
	ctx := context.Background()

	attempts := 0
	operation := func(ctx context.Context) (string, error) {
		attempts++
		return "", errors.New("immediate failure")
	}

	result, err := retry.WithBackoff(ctx, config, operation)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
	if result != "" {
		t.Errorf("Expected empty result, got '%s'", result)
	}
	if attempts != 1 {
		t.Errorf("Expected 1 attempt, got %d", attempts)
	}
}

// TestRetryWithBackoff_Jitter ensures that randomness is applied to the delay.
func TestRetryWithBackoff_Jitter(t *testing.T) {
	config := retry.Config{
		Attempts:     3,
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     500 * time.Millisecond,
		Multiplier:   2.0,
		RandomFactor: 0.5,
	}
	ctx := context.Background()

	attempts := 0
	operation := func(ctx context.Context) (string, error) {
		attempts++
		return "", errors.New("jitter test")
	}

	start := time.Now()
	_, _ = retry.WithBackoff(ctx, config, operation)
	duration := time.Since(start)

	// We only get two sleeps for three attempts:
	//   1st sleep: ~100ms (+ jitter)
	//   2nd sleep: ~200ms (+ jitter)
	baseDelay := 100*time.Millisecond + 200*time.Millisecond
	minExpectedDuration := baseDelay
	maxExpectedDuration := baseDelay + time.Duration(float64(baseDelay)*config.RandomFactor)

	// Allow some margin on both sides
	if duration < minExpectedDuration-50*time.Millisecond || duration > maxExpectedDuration+200*time.Millisecond {
		t.Errorf(
			"Expected duration between %v and %v, got %v",
			minExpectedDuration, maxExpectedDuration, duration,
		)
	}
}
