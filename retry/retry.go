package retry

import (
	"context"
	"math/rand"
	"time"
)

// Config holds configuration for retry operations because sometimes
// you need more than just "try again and hope for the best"
type Config struct {
	Attempts     int
	InitialDelay time.Duration
	MaxDelay     time.Duration
	Multiplier   float64
	RandomFactor float64
}

// DefaultRetryConfig returns a configuration that's probably better than
// whatever you were going to come up with
func DefaultRetryConfig() Config {
	return Config{
		Attempts:     3,
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     10 * time.Second,
		Multiplier:   2.0,
		RandomFactor: 0.1,
	}
}

// WithBackoff retries an operation with exponential backoff because
// hammering a service repeatedly is so last decade
func WithBackoff[T any](ctx context.Context, config Config, operation func(context.Context) (T, error)) (T, error) {
	var result T
	var lastError error
	delay := config.InitialDelay

	for attempt := 0; attempt < config.Attempts; attempt++ {
		select {
		case <-ctx.Done():
			return result, ctx.Err()
		default:
			if attempt > 0 {
				// Ensure jitter only adds to the delay (never reduces it)
				jitter := time.Duration(rand.Float64() * float64(delay) * config.RandomFactor)
				actualDelay := delay + jitter

				time.Sleep(actualDelay)

				// Ensure delay accumulates correctly with max cap
				delay = time.Duration(float64(delay) * config.Multiplier)
				if delay > config.MaxDelay {
					delay = config.MaxDelay
				}
			}

			result, err := operation(ctx)
			if err == nil {
				return result, nil
			}
			lastError = err
		}
	}

	return result, lastError
}
