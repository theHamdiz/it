package lb_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/theHamdiz/it/lb"
)

func TestNewLoadBalancer(t *testing.T) {
	lb_ := lb.NewLoadBalancer(5)
	if lb_ == nil {
		t.Fatal("Expected LoadBalancer instance, got nil")
	}
	if cap(lb_.Workers()) != 5 {
		t.Errorf("Expected worker capacity to be 5, got %d", cap(lb_.Workers()))
	}
}

func TestNewLoadBalancerWithContext(t *testing.T) {
	ctx := context.Background()
	lb_ := lb.NewLoadBalancerWithContext(ctx, 3)
	if lb_ == nil {
		t.Fatal("Expected LoadBalancer instance, got nil")
	}
	if cap(lb_.Workers()) != 3 {
		t.Errorf("Expected worker capacity to be 3, got %d", cap(lb_.Workers()))
	}
}

func TestLoadBalancer_Execute_Success(t *testing.T) {
	lb_ := lb.NewLoadBalancer(2)
	defer lb_.Close()

	ctx := context.Background()
	err := lb_.Execute(ctx, func() error {
		return nil
	})

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestLoadBalancer_Execute_ContextCancelled(t *testing.T) {
	lb_ := lb.NewLoadBalancer(1)
	defer lb_.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel before execution

	err := lb_.Execute(ctx, func() error {
		return nil
	})

	if !errors.Is(err, context.Canceled) {
		t.Errorf("Expected context.Canceled error, got %v", err)
	}
}

func TestLoadBalancer_Execute_LimitExceeded(t *testing.T) {
	lb_ := lb.NewLoadBalancer(1)
	defer lb_.Close()

	ctx := context.Background()

	// Run the first task concurrently so it occupies the worker slot
	go func() {
		_ = lb_.Execute(ctx, func() error {
			time.Sleep(100 * time.Millisecond)
			return nil
		})
	}()

	// Give the goroutine a moment to start and occupy the worker
	time.Sleep(10 * time.Millisecond)

	// Now the worker slot should still be in use. Let's do a second call with a short timeout.
	ctxTimeout, cancel := context.WithTimeout(ctx, 50*time.Millisecond)
	defer cancel()

	err := lb_.Execute(ctxTimeout, func() error {
		return nil
	})

	// Now we expect context.DeadlineExceeded, since the slot never freed in time.
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Expected context.DeadlineExceeded, got %v", err)
	}
}

func TestLoadBalancer_Execute_LoadBalancerClosed(t *testing.T) {
	lb_ := lb.NewLoadBalancer(1)
	lb_.Close()

	ctx := context.Background()
	err := lb_.Execute(ctx, func() error {
		return nil
	})

	if err == nil || err.Error() != "load balancer is closed" {
		t.Errorf("Expected 'load balancer is closed' error, got %v", err)
	}
}

func TestExecuteBalanced_Success(t *testing.T) {
	lb_ := lb.NewLoadBalancer(2)
	defer lb_.Close()

	ctx := context.Background()
	result, err := lb.ExecuteBalanced(lb_, ctx, func() (string, error) {
		return "success", nil
	})

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if result != "success" {
		t.Errorf("Expected result 'success', got '%s'", result)
	}
}

func TestExecuteBalanced_ContextCancelled(t *testing.T) {
	lb_ := lb.NewLoadBalancer(2)
	defer lb_.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	result, err := lb.ExecuteBalanced(lb_, ctx, func() (string, error) {
		return "should not execute", nil
	})

	if !errors.Is(err, context.Canceled) {
		t.Errorf("Expected context.Canceled error, got %v", err)
	}

	if result != "" {
		t.Errorf("Expected empty result due to cancellation, got '%s'", result)
	}
}

func TestExecuteBalanced_LoadBalancerClosed(t *testing.T) {
	lb_ := lb.NewLoadBalancer(1)
	lb_.Close()

	ctx := context.Background()
	result, err := lb.ExecuteBalanced(lb_, ctx, func() (string, error) {
		return "should not execute", nil
	})

	if err == nil || err.Error() != "load balancer is closed" {
		t.Errorf("Expected 'load balancer is closed' error, got %v", err)
	}

	if result != "" {
		t.Errorf("Expected empty result due to closure, got '%s'", result)
	}
}

func TestLoadBalancer_Close(t *testing.T) {
	lb_ := lb.NewLoadBalancer(1)
	lb_.Close()

	// Check if the internal context is cancelled
	select {
	case <-lb_.Ctx().Done():
		// Expected behavior
	default:
		t.Errorf("Expected LoadBalancer context to be cancelled after Close()")
	}
}

func TestDefaultLoadBalancer(t *testing.T) {
	lb_ := lb.DefaultLoadBalancer()
	if lb_ == nil {
		t.Fatal("Expected DefaultLoadBalancer instance, got nil")
	}
	if cap(lb_.Workers()) != 10 {
		t.Errorf("Expected worker capacity to be 10, got %d", cap(lb_.Workers()))
	}
}

func TestDefaultLoadBalancerWithContext(t *testing.T) {
	ctx := context.Background()
	lb_ := lb.DefaultLoadBalancerWithContext(ctx)
	if lb_ == nil {
		t.Fatal("Expected DefaultLoadBalancer instance, got nil")
	}
	if cap(lb_.Workers()) != 10 {
		t.Errorf("Expected worker capacity to be 10, got %d", cap(lb_.Workers()))
	}
}
