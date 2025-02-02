package rl_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/theHamdiz/it/rl"
)

// TestNewRateLimiter verifies that a new rate limiter initializes correctly
func TestNewRateLimiter(t *testing.T) {
	rl_ := rl.NewRateLimiter(500*time.Millisecond, 5)
	if rl_ == nil {
		t.Fatal("Expected RateLimiter instance, got nil")
	}
	if rl_.BatchSize() != 5 {
		t.Errorf("Expected batch size 5, got %d", rl_.BatchSize())
	}
	if rl_.Interval() != 500*time.Millisecond {
		t.Errorf("Expected interval 500ms, got %v", rl_.Interval())
	}
}

// TestNewRateLimiterWithContext ensures rate limiter respects external context
func TestNewRateLimiterWithContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rl_ := rl.NewRateLimiterWithContext(ctx, 200*time.Millisecond, 3)
	if rl_.Ctx() != ctx {
		t.Errorf("Expected rate limiter to use provided context")
	}
}

// TestRateLimiter_Execute_Success ensures operations execute correctly when tokens are available
func TestRateLimiter_Execute_Success(t *testing.T) {
	rl_ := rl.NewRateLimiter(100*time.Millisecond, 2)
	defer rl_.Close()

	ctx := context.Background()

	err := rl_.Execute(ctx, func() error {
		return nil
	})

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

// TestRateLimiter_Execute_NoTokens ensures execution blocks when no tokens are available
func TestRateLimiter_Execute_NoTokens(t *testing.T) {
	rl_ := rl.NewRateLimiter(500*time.Millisecond, 1)
	defer rl_.Close()

	ctx := context.Background()

	_ = rl_.Execute(ctx, func() error {
		return nil
	})

	ctxTimeout, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
	defer cancel()

	err := rl_.Execute(ctxTimeout, func() error {
		return nil
	})

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Expected context.DeadlineExceeded, got %v", err)
	}
}

// TestRateLimiter_TokenReplenishment ensures tokens replenish at the correct interval
func TestRateLimiter_TokenReplenishment(t *testing.T) {
	rl_ := rl.NewRateLimiter(100*time.Millisecond, 2)
	defer rl_.Close()

	ctx := context.Background()

	_ = rl_.Execute(ctx, func() error { return nil })
	_ = rl_.Execute(ctx, func() error { return nil })

	ctxTimeout, cancel := context.WithTimeout(ctx, 50*time.Millisecond)
	defer cancel()

	err := rl_.Execute(ctxTimeout, func() error {
		return nil
	})

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Expected context.DeadlineExceeded before replenishment, got %v", err)
	}

	time.Sleep(150 * time.Millisecond)

	err = rl_.Execute(ctx, func() error { return nil })
	if err != nil {
		t.Errorf("Expected execution to succeed after token replenishment, got %v", err)
	}
}

// TestExecuteRateLimited_Success verifies that ExecuteRateLimited executes properly when tokens are available
func TestExecuteRateLimited_Success(t *testing.T) {
	rl_ := rl.NewRateLimiter(100*time.Millisecond, 1)
	defer rl_.Close()

	ctx := context.Background()

	result, err := rl.ExecuteRateLimited(rl_, ctx, func() (string, error) {
		return "success", nil
	})

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if result != "success" {
		t.Errorf("Expected 'success', got '%s'", result)
	}
}

// TestExecuteRateLimited_NoTokens ensures ExecuteRateLimited respects rate limiting
func TestExecuteRateLimited_NoTokens(t *testing.T) {
	rl_ := rl.NewRateLimiter(500*time.Millisecond, 1)
	defer rl_.Close()

	ctx := context.Background()

	_, _ = rl.ExecuteRateLimited(rl_, ctx, func() (string, error) {
		return "done", nil
	})

	ctxTimeout, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
	defer cancel()

	result, err := rl.ExecuteRateLimited(rl_, ctxTimeout, func() (string, error) {
		return "blocked", nil
	})

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Expected context.DeadlineExceeded, got %v", err)
	}

	if result != "" {
		t.Errorf("Expected empty result due to timeout, got '%s'", result)
	}
}

// TestRateLimiter_Close ensures that closing the rate limiter stops replenishment
func TestRateLimiter_Close(t *testing.T) {
	rl_ := rl.NewRateLimiter(100*time.Millisecond, 1)
	rl_.Close()

	select {
	case <-rl_.Ctx().Done():
	default:
		t.Errorf("Expected RateLimiter context to be canceled after Close()")
	}
}

// TestDefaultRateLimiter verifies the default rate limiter initialization
func TestDefaultRateLimiter(t *testing.T) {
	rl_ := rl.DefaultRateLimiter()
	if rl_ == nil {
		t.Fatal("Expected DefaultRateLimiter instance, got nil")
	}
	if rl_.BatchSize() != 10 {
		t.Errorf("Expected batch size 10, got %d", rl_.BatchSize())
	}
	if rl_.Interval() != 1*time.Second {
		t.Errorf("Expected interval 1s, got %v", rl_.Interval())
	}
}

// TestDefaultRateLimiterWithContext verifies that the default rate limiter with context initializes properly
func TestDefaultRateLimiterWithContext(t *testing.T) {
	ctx := context.Background()
	rl_ := rl.DefaultRateLimiterWithContext(ctx)
	if rl_ == nil {
		t.Fatal("Expected DefaultRateLimiter instance, got nil")
	}
	if rl_.BatchSize() != 10 {
		t.Errorf("Expected batch size 10, got %d", rl_.BatchSize())
	}
	if rl_.Interval() != 1*time.Second {
		t.Errorf("Expected interval 1s, got %v", rl_.Interval())
	}
	if rl_.Ctx() != ctx {
		t.Errorf("Expected rate limiter to use provided context")
	}
}
