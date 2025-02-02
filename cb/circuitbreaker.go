// Package cb - For when your dependencies are as reliable as a chocolate teapot
package cb

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

const (
	ErrCircuitOpen = "circuit breaker is open"
)

// CircuitBreaker implements the "nope, not gonna try that again" pattern
// It's like a bouncer for your function calls
type CircuitBreaker struct {
	failures    atomic.Int64  // Counter of disappointments
	lastFailure atomic.Int64  // Timestamp of our most recent disaster
	threshold   int64         // How many failures until we give up
	timeout     time.Duration // How long we sulk before trying again
	mu          sync.RWMutex  // Protects our delicate state
}

// NewCircuitBreaker creates a new failure detection system
// threshold: how many times you're willing to get hurt
// timeout: how long you need to recover from trust issues
func NewCircuitBreaker(threshold int64, timeout time.Duration) *CircuitBreaker {
	if threshold < 1 {
		threshold = 1 // Because zero tolerance is too harsh
	}
	return &CircuitBreaker{
		threshold: threshold,
		timeout:   timeout,
	}
}

// Execute attempts to run your probably-going-to-fail function
// Returns error when it inevitably breaks
func (cb *CircuitBreaker) Execute(fn func() error) error {
	cb.mu.RLock()
	fails := cb.failures.Load()
	lastFail := time.Unix(0, cb.lastFailure.Load())
	cb.mu.RUnlock()

	// Check if circuit is open
	if fails >= cb.threshold {
		// For zero timeout, circuit stays open indefinitely until Reset()
		if cb.timeout == 0 {
			return errors.New(ErrCircuitOpen)
		}
		// For non-zero timeout, check if enough time has passed
		if time.Since(lastFail) <= cb.timeout {
			return errors.New(ErrCircuitOpen)
		}
		// Reset circuit after timeout
		cb.mu.Lock()
		cb.reset()
		cb.mu.Unlock()
	}

	// Execute function
	if err := fn(); err != nil {
		cb.mu.Lock()
		cb.failures.Add(1)
		cb.lastFailure.Store(time.Now().UnixNano())
		currentFails := cb.failures.Load()
		cb.mu.Unlock()

		if currentFails >= cb.threshold {
			return errors.New(ErrCircuitOpen)
		}
		return err
	}

	return nil
}

// canTry checks if we're emotionally ready to try again
func (cb *CircuitBreaker) canTry() bool {
	fails := cb.failures.Load()
	if fails >= cb.threshold {
		// For zero timeout, circuit stays open indefinitely
		if cb.timeout == 0 {
			return false
		}
		lastFail := time.Unix(0, cb.lastFailure.Load())
		if time.Since(lastFail) <= cb.timeout {
			return false // Still in therapy
		}
		cb.reset()
	}
	return true
}

// recordFailure adds another tally to our wall of shame
func (cb *CircuitBreaker) recordFailure() {
	cb.failures.Add(1)
	cb.lastFailure.Store(time.Now().UnixNano())
}

// reset wipes the slate clean (but not your memory)
func (cb *CircuitBreaker) reset() {
	cb.failures.Store(0)
	cb.lastFailure.Store(0)
}

// Various getters because encapsulation is important (or something)

func (cb *CircuitBreaker) Timeout() time.Duration {
	return cb.timeout
}

func (cb *CircuitBreaker) Threshold() int64 {
	return cb.threshold
}

func (cb *CircuitBreaker) Failures() int64 {
	return cb.failures.Load() // Count of our collective disappointments
}

func (cb *CircuitBreaker) LastFailure() time.Time {
	return time.Unix(cb.lastFailure.Load(), 0)
}

// State checking functions, for those who care about such things

func (cb *CircuitBreaker) IsOpen() bool {
	return cb.failures.Load() >= cb.threshold // Are we currently in timeout?
}

func (cb *CircuitBreaker) IsClosed() bool {
	return !cb.IsOpen() // Everything is fine (for now)
}

func (cb *CircuitBreaker) IsHalfOpen() bool {
	return cb.failures.Load() < cb.threshold // Cautiously optimistic
}

// Setters for the masochists who want to adjust mid-flight

func (cb *CircuitBreaker) SetTimeout(timeout time.Duration) {
	cb.timeout = timeout
}

func (cb *CircuitBreaker) SetThreshold(threshold int64) {
	cb.threshold = threshold
}

func (cb *CircuitBreaker) Reset() {
	// Fresh start, same problems
	cb.reset()
}

func (cb *CircuitBreaker) String() string {
	return fmt.Sprintf("CircuitBreaker{threshold=%d, timeout=%s, failures=%d, lastFailure=%s}",
		cb.threshold, cb.timeout, cb.failures.Load(), cb.LastFailure())
}
