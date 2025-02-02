package cb_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/theHamdiz/it/cb"
)

var (
	errTest = errors.New("test error")
)

func TestNewCircuitBreaker(t *testing.T) {
	tests := []struct {
		name           string
		inputThreshold int64
		inputTimeout   time.Duration
		wantThreshold  int64
		wantTimeout    time.Duration
	}{
		{
			name:           "Default values",
			inputThreshold: 5,
			inputTimeout:   time.Second,
			wantThreshold:  5,
			wantTimeout:    time.Second,
		},
		{
			name:           "Zero threshold becomes one",
			inputThreshold: 0,
			inputTimeout:   time.Second,
			wantThreshold:  1,
			wantTimeout:    time.Second,
		},
		{
			name:           "Negative threshold becomes one",
			inputThreshold: -1,
			inputTimeout:   time.Second,
			wantThreshold:  1,
			wantTimeout:    time.Second,
		},
		{
			name:           "Zero timeout",
			inputThreshold: 5,
			inputTimeout:   0,
			wantThreshold:  5,
			wantTimeout:    0,
		},
		{
			name:           "Large values",
			inputThreshold: 1000,
			inputTimeout:   time.Hour,
			wantThreshold:  1000,
			wantTimeout:    time.Hour,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cb_ := cb.NewCircuitBreaker(tt.inputThreshold, tt.inputTimeout)
			if cb_ == nil {
				t.Fatal("Expected non-nil CircuitBreaker")
			}
			if cb_.Threshold() != tt.wantThreshold {
				t.Errorf("Expected threshold %d, got %d", tt.wantThreshold, cb_.Threshold())
			}
			if cb_.Timeout() != tt.wantTimeout {
				t.Errorf("Expected timeout %v, got %v", tt.wantTimeout, cb_.Timeout())
			}
		})
	}
}

func TestCircuitBreaker_Execute(t *testing.T) {
	tests := []struct {
		name       string
		threshold  int64
		timeout    time.Duration
		operations []error
		expectOpen bool
		setupPause time.Duration
	}{
		{
			name:       "Success case",
			threshold:  3,
			timeout:    time.Second,
			operations: []error{nil, nil, nil},
			expectOpen: false,
		},
		{
			name:       "Failure threshold reached",
			threshold:  3,
			timeout:    time.Second,
			operations: []error{errTest, errTest, errTest},
			expectOpen: true,
		},
		{
			name:       "Mixed success and failure",
			threshold:  3,
			timeout:    time.Second,
			operations: []error{errTest, nil, errTest},
			expectOpen: false,
		},
		{
			name:       "Recovery after timeout",
			threshold:  2,
			timeout:    100 * time.Millisecond,
			operations: []error{errTest, errTest},
			expectOpen: false,
			setupPause: 150 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cb_ := cb.NewCircuitBreaker(tt.threshold, tt.timeout)

			for _, err := range tt.operations {
				_ = cb_.Execute(func() error { return err })
			}

			if tt.setupPause > 0 {
				time.Sleep(tt.setupPause)
			}

			err := cb_.Execute(func() error { return nil })
			if tt.expectOpen && err == nil {
				t.Error("Expected circuit to be open")
			} else if !tt.expectOpen && err != nil {
				t.Errorf("Expected circuit to be closed, got error: %v", err)
			}
		})
	}
}

func TestCircuitBreaker_Concurrency(t *testing.T) {
	cb_ := cb.NewCircuitBreaker(100, time.Second)
	const goroutines = 10
	const iterations = 100

	done := make(chan bool, goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			for j := 0; j < iterations; j++ {
				_ = cb_.Execute(func() error {
					if j%2 == 0 {
						return errTest
					}
					return nil
				})
			}
			done <- true
		}()
	}

	for i := 0; i < goroutines; i++ {
		<-done
	}
}

func TestCircuitBreaker_Reset(t *testing.T) {
	cb_ := cb.NewCircuitBreaker(2, time.Second)

	_ = cb_.Execute(func() error { return errTest })
	_ = cb_.Execute(func() error { return errTest })

	if err := cb_.Execute(func() error { return nil }); err == nil {
		t.Error("Expected circuit to be open")
	}

	cb_.Reset()

	if err := cb_.Execute(func() error { return nil }); err != nil {
		t.Errorf("Expected circuit to be closed after reset, got error: %v", err)
	}
}

func TestCircuitBreaker_Timeout(t *testing.T) {
	tests := []struct {
		name     string
		timeout  time.Duration
		sleep    time.Duration
		wantOpen bool
	}{
		{
			name:     "Should remain open before timeout",
			timeout:  200 * time.Millisecond,
			sleep:    50 * time.Millisecond,
			wantOpen: true,
		},
		{
			name:     "Should close after timeout",
			timeout:  100 * time.Millisecond,
			sleep:    200 * time.Millisecond,
			wantOpen: false,
		},
	}

	errTest := errors.New("test error")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cb_ := cb.NewCircuitBreaker(2, tt.timeout)

			if err := cb_.Execute(func() error { return errTest }); !errors.Is(err, errTest) {
				t.Fatalf("First execution: expected %v, got %v", errTest, err)
			}
			if err := cb_.Execute(func() error { return errTest }); err == nil || !strings.Contains(err.Error(), cb.ErrCircuitOpen) {
				t.Fatalf("Second execution: expected circuit to open, got %v", err)
			}

			err := cb_.Execute(func() error { return nil })
			if err == nil || !strings.Contains(err.Error(), cb.ErrCircuitOpen) {
				t.Fatal("Circuit should be open initially")
			}

			time.Sleep(tt.sleep)

			err = cb_.Execute(func() error { return nil })
			isOpen := err != nil && strings.Contains(err.Error(), cb.ErrCircuitOpen)

			if tt.wantOpen != isOpen {
				t.Errorf("After waiting %v, expected open=%v, got open=%v (err=%v)",
					tt.sleep, tt.wantOpen, isOpen, err)
			}
		})
	}
}

func TestCircuitBreaker_EdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(*cb.CircuitBreaker)
		operation func() error
		wantErr   string
	}{
		{
			name: "zero threshold",
			setup: func(cb_ *cb.CircuitBreaker) {
			},
			operation: func() error { return errors.New("test") },
			wantErr:   "test",
		},
		{
			name: "reset after success",
			setup: func(cb_ *cb.CircuitBreaker) {
				_ = cb_.Execute(func() error { return errors.New("test") })
			},
			operation: func() error { return nil },
			wantErr:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cb_ := cb.NewCircuitBreaker(2, 100*time.Millisecond)
			if tt.setup != nil {
				tt.setup(cb_)
			}
			err := tt.operation()
			if (err == nil && tt.wantErr != "") || (err != nil && err.Error() != tt.wantErr) {
				t.Errorf("got error %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestCircuitBreaker_ZeroValues(t *testing.T) {
	t.Run("Zero threshold", func(t *testing.T) {
		cb_ := cb.NewCircuitBreaker(0, time.Second)
		if cb_.Threshold() != 1 {
			t.Errorf("Expected threshold to be set to 1, got %d", cb_.Threshold())
		}
		if err := cb_.Execute(func() error { return nil }); err != nil {
			t.Error("Expected success with minimum threshold")
		}
	})

	t.Run("Zero timeout", func(t *testing.T) {
		cb_ := cb.NewCircuitBreaker(1, 0)

		if !cb_.IsClosed() {
			t.Error("Circuit should start closed")
		}

		err := cb_.Execute(func() error { return errors.New("test error") })
		if err == nil || err.Error() != cb.ErrCircuitOpen {
			t.Fatalf("Expected test error, got: %v", err)
		}

		if !cb_.IsOpen() {
			t.Error("Circuit should be open after failure")
		}

		err = cb_.Execute(func() error { return nil })
		if err == nil || err.Error() != cb.ErrCircuitOpen {
			t.Fatalf("Expected circuit open error, got: %v", err)
		}

		time.Sleep(100 * time.Millisecond)
		err = cb_.Execute(func() error { return nil })
		if err == nil || err.Error() != cb.ErrCircuitOpen {
			t.Errorf("Expected circuit to remain open with zero timeout, got: %v", err)
		}

		cb_.Reset()
		if !cb_.IsClosed() {
			t.Error("Circuit should be closed after reset")
		}
		if err := cb_.Execute(func() error { return nil }); err != nil {
			t.Errorf("Expected success after reset, got: %v", err)
		}
	})

	t.Run("Zero timeout state transitions", func(t *testing.T) {
		cb_ := cb.NewCircuitBreaker(2, 0)

		_ = cb_.Execute(func() error { return errors.New("error 1") })
		if !cb_.IsHalfOpen() {
			t.Error("Circuit should be half-open after first failure")
		}

		_ = cb_.Execute(func() error { return errors.New("error 2") })
		if !cb_.IsOpen() {
			t.Error("Circuit should be open after second failure")
		}

		time.Sleep(100 * time.Millisecond)
		if !cb_.IsOpen() {
			t.Error("Circuit should remain open with zero timeout")
		}

		cb_.Reset()
		if !cb_.IsClosed() {
			t.Error("Circuit should be closed after manual reset")
		}
	})
}

func BenchmarkCircuitBreaker(b *testing.B) {
	cb_ := cb.NewCircuitBreaker(1000, time.Second)

	b.Run("Successful executions", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = cb_.Execute(func() error { return nil })
		}
	})

	b.Run("Failed executions", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = cb_.Execute(func() error { return errTest })
		}
	})
}
