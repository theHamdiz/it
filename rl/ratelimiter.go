// Package rl - Because sometimes you need to tell your code "slow down there, buddy"
package rl

import (
	"context"
	"time"
)

// RateLimiter is like a bouncer for your function calls
// Keeps them in line and makes sure they don't cause a scene
type RateLimiter struct {
	tokens    chan struct{}      // VIP passes
	interval  time.Duration      // How often we let the next batch in
	batchSize int                // How many get in at once
	ctx       context.Context    // The party's context
	cancel    context.CancelFunc // The panic button
}

// NewRateLimiter creates a new function traffic controller
// interval: how often we hand out passes
// batchSize: how many passes we give out at once
func NewRateLimiter(interval time.Duration, batchSize int) *RateLimiter {
	ctx, cancel := context.WithCancel(context.Background())
	rl := &RateLimiter{
		tokens:    make(chan struct{}, batchSize), // The VIP list
		interval:  interval,
		batchSize: batchSize,
		ctx:       ctx,
		cancel:    cancel,
	}

	go rl.replenishTokens() // Start the token fairy
	return rl
}

// NewRateLimiterWithContext is like NewRateLimiter but with a bedtime
func NewRateLimiterWithContext(ctx context.Context, interval time.Duration, batchSize int) *RateLimiter {
	rl := &RateLimiter{
		tokens:    make(chan struct{}, batchSize),
		interval:  interval,
		batchSize: batchSize,
		ctx:       ctx,
		cancel:    func() {}, // Fake cancel because we're using someone else's context
	}

	go rl.replenishTokens()
	return rl
}

// Execute runs your function when it's allowed to
// Return an error when your function misbehaves
func (rl *RateLimiter) Execute(ctx context.Context, operation func() error) error {
	select {
	case <-ctx.Done():
		return ctx.Err() // Sorry, party's over
	case <-rl.tokens:
		return operation() // Your turn to shine
	}
}

// ExecuteRateLimited is like Execute but for functions that actually return something
func ExecuteRateLimited[T any](rl *RateLimiter, ctx context.Context, operation func() (T, error)) (T, error) {
	var zero T // In case we need to leave empty-handed
	select {
	case <-ctx.Done():
		return zero, ctx.Err()
	case <-rl.tokens:
		return operation()
	}
}

// replenishTokens is the backstage worker keeping the party supplied
func (rl *RateLimiter) replenishTokens() {
	ticker := time.NewTicker(rl.interval)
	defer ticker.Stop()

	for {
		select {
		case <-rl.ctx.Done():
			return // Time to go home
		case <-ticker.C:
			for i := 0; i < rl.batchSize; i++ {
				select {
				case rl.tokens <- struct{}{}:
				default:
					// Club's full, try again later
				}
			}
		}
	}
}

// Close tells everyone to go home
func (rl *RateLimiter) Close() {
	rl.cancel()
}

// DefaultRateLimiter creates a rate limiter for the indecisive
// 1-second interval, 10 passes because why not
func DefaultRateLimiter() *RateLimiter {
	return NewRateLimiter(1*time.Second, 10)
}

// DefaultRateLimiterWithContext is like DefaultRateLimiter but with a curfew
func DefaultRateLimiterWithContext(ctx context.Context) *RateLimiter {
	return NewRateLimiterWithContext(ctx, 1*time.Second, 10)
}

// Tokens returns the channel controlling access
// But seriously, don't mess with this directly
func (rl *RateLimiter) Tokens() chan struct{} {
	return rl.tokens
}

// Interval tells you how long you have to wait
func (rl *RateLimiter) Interval() time.Duration {
	return rl.interval
}

// BatchSize tells you how many get in at once
func (rl *RateLimiter) BatchSize() int {
	return rl.batchSize
}

// Ctx returns the rate limiter's context
// In case you need more ways to shut things down
func (rl *RateLimiter) Ctx() context.Context {
	return rl.ctx
}
