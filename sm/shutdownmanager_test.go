package sm_test

import (
	"context"
	"errors"
	"log"
	"os"
	"syscall"
	"testing"
	"time"

	. "github.com/theHamdiz/it/sm"
)

// TestNewShutdownManager ensures that creating a new manager initializes correctly.
func TestNewShutdownManager(t *testing.T) {
	sm_ := NewShutdownManager()
	if sm_ == nil {
		t.Fatal("Expected non-nil ShutdownManager")
	}
}

// TestShutdownManager_NoSignal tests that if no signal is sent, the manager
// does not execute shutdown actions automatically. We then manually call Close().
func TestShutdownManager_NoSignal(t *testing.T) {
	sm_ := NewShutdownManager(syscall.SIGUSR2)
	defer sm_.Close()

	actionRan := false
	sm_.AddAction("never-run", func(ctx context.Context) error {
		actionRan = true
		return nil
	}, 500*time.Millisecond, false)

	sm_.Start()

	select {
	case <-time.After(200 * time.Millisecond):
	}

	sm_.Close()

	if err := sm_.Wait(); err != nil {
		t.Errorf("Did not expect an error but got: %v", err)
	}

	if actionRan {
		t.Error("Action ran without a signal being sent; expected it NOT to run")
	}
}

// TestShutdownManager_Signal tests that sending a signal triggers shutdown actions.
func TestShutdownManager_Signal(t *testing.T) {
	sm_ := NewShutdownManager(syscall.SIGUSR1)
	defer sm_.Close()

	actionRan := make(chan bool, 1)

	sm_.AddAction("test-action", func(ctx context.Context) error {
		actionRan <- true
		return nil
	}, time.Second, false)

	sm_.Start()

	p, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatalf("failed to find current process: %v", err)
	}
	if err := p.Signal(syscall.SIGUSR1); err != nil {
		t.Fatalf("failed to send signal: %v", err)
	}

	if err := sm_.Wait(); err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	select {
	case ran := <-actionRan:
		if !ran {
			t.Error("Expected action to run, but it didn't")
		}
	default:
		t.Error("Action channel was never triggered")
	}
}

// TestShutdownManager_CriticalFailure checks that a critical action error
// stops subsequent actions and returns an error.
func TestShutdownManager_CriticalFailure(t *testing.T) {
	sm_ := NewShutdownManager(syscall.SIGUSR2)
	defer sm_.Close()

	action1Ran := false
	action2Ran := false

	sm_.AddAction("critical-fail", func(ctx context.Context) error {
		action1Ran = true
		return errors.New("simulated critical failure")
	}, 50*time.Millisecond, true)

	sm_.AddAction("non-critical-after", func(ctx context.Context) error {
		action2Ran = true
		return nil
	}, 50*time.Millisecond, false)

	sm_.Start()

	p, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatalf("failed to find process: %v", err)
	}
	_ = p.Signal(syscall.SIGUSR2)

	err = sm_.Wait()
	if err == nil {
		t.Fatal("Expected error due to critical failure, but got nil")
	}
	expected := "critical shutdown action critical-fail failed:"
	if err.Error()[:len(expected)] != expected {
		t.Errorf("Expected error to start with %q, got %q", expected, err.Error())
	}

	if !action1Ran {
		t.Error("Expected action1 to run")
	}
	if action2Ran {
		t.Error("Action2 should not run after critical failure, but it did")
	}
}

// TestShutdownManager_NonCriticalFailure checks that a non-critical failure
// does not prevent subsequent actions from running, and does not propagate
// as the final error.
func TestShutdownManager_NonCriticalFailure(t *testing.T) {
	sm_ := NewShutdownManager(syscall.SIGUSR2)
	defer sm_.Close()

	action1Ran := false
	action2Ran := false

	sm_.AddAction("non-critical-fail", func(ctx context.Context) error {
		action1Ran = true
		return errors.New("simulated non-critical failure")
	}, 50*time.Millisecond, false)

	sm_.AddAction("second-action", func(ctx context.Context) error {
		action2Ran = true
		return nil
	}, 50*time.Millisecond, false)

	sm_.Start()

	p, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatalf("failed to find process: %v", err)
	}
	_ = p.Signal(syscall.SIGUSR2)

	err = sm_.Wait()
	if err != nil {
		t.Fatalf("Did not expect error from non-critical failure, got: %v", err)
	}

	if !action1Ran {
		t.Error("Expected non-critical action1 to run")
	}
	if !action2Ran {
		t.Error("Expected action2 to run after non-critical failure, but it didn't")
	}
}

// TestShutdownManager_Timeout checks that each action respects its individual timeout.
func TestShutdownManager_Timeout(t *testing.T) {
	sm_ := NewShutdownManager(syscall.SIGUSR2)
	defer sm_.Close()

	longRunning := func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			// Timed out or canceled
			return ctx.Err()
		case <-time.After(100 * time.Millisecond):
			return nil
		}
	}

	sm_.AddAction("timeout-action", longRunning, 10*time.Millisecond, false)

	sm_.Start()

	p, _ := os.FindProcess(os.Getpid())
	_ = p.Signal(syscall.SIGUSR2)

	err := sm_.Wait()
	if err != nil {
		t.Errorf("Timeout is non-critical; expected no final error, got: %v", err)
	}
}

// TestShutdownManager_Close ensures Close cancels the context so actions cannot proceed.
func TestShutdownManager_Close(t *testing.T) {
	sm_ := NewShutdownManager(syscall.SIGUSR1)

	actionRan := false
	sm_.AddAction("test-action", func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(100 * time.Millisecond):
			actionRan = true
			return nil
		}
	}, time.Second, false)

	sm_.Start()

	sm_.Close()

	time.Sleep(200 * time.Millisecond)

	// Because we closed early (before sending a signal), no actions should be executed
	// The manager won't exit on its own if never signaled, so forcibly wait in a goroutine:
	go func() {
		if err := sm_.Wait(); err != nil {
			log.Printf("Wait returned error after Close(): %v", err)
		}
	}()

	// We expect the action to have been canceled.
	if actionRan {
		t.Error("Action should not have completed; expected context cancellation")
	}
}
