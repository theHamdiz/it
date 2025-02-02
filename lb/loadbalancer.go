// Package lb - Because your CPU cores need a union representative
package lb

import (
	"context"
	"errors"
	"time"
)

// LoadBalancer is like a bouncer for your goroutines
// Keeps them in line and makes sure nobody gets trampled
type LoadBalancer struct {
	workers chan struct{}      // The VIP list
	ctx     context.Context    // The party's context
	cancel  context.CancelFunc // The "everybody out" button
}

// NewLoadBalancer creates a new work distribution committee
// maxWorkers: how many goroutines we trust at once
func NewLoadBalancer(maxWorkers int) *LoadBalancer {
	ctx, cancel := context.WithCancel(context.Background())
	return &LoadBalancer{
		workers: make(chan struct{}, maxWorkers), // Our exclusive guest list
		ctx:     ctx,
		cancel:  cancel,
	}
}

// NewLoadBalancerWithContext is like NewLoadBalancer but with a bedtime
func NewLoadBalancerWithContext(ctx context.Context, maxWorkers int) *LoadBalancer {
	ctx, cancel := context.WithCancel(ctx)
	return &LoadBalancer{
		workers: make(chan struct{}, maxWorkers),
		ctx:     ctx,
		cancel:  cancel,
	}
}

// Execute runs your function through security
// Returns error when things inevitably go wrong
func (lb *LoadBalancer) Execute(ctx context.Context, operation func() error) error {
	select {
	// Sorry, your party got canceled
	case <-ctx.Done():
		return ctx.Err()
	// We're closed for renovation
	case <-lb.ctx.Done():
		return errors.New("load balancer is closed")
	default:
		// The party's still going
	}

	deadline, hasDeadline := ctx.Deadline()
	if hasDeadline {
		timer := time.NewTimer(time.Until(deadline))
		defer timer.Stop()

		select {
		// Someone pulled the fire alarm
		case <-ctx.Done():
			return ctx.Err()
		// Management called it a night
		case <-lb.ctx.Done():
			return errors.New("load balancer is closed")
		// Time's up, go home
		case <-timer.C:
			return context.DeadlineExceeded
		// Don't forget to return your VIP pass
		case lb.workers <- struct{}{}:
			defer func() { <-lb.workers }()
		}
	} else {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-lb.ctx.Done():
			return errors.New("load balancer is closed")
		case lb.workers <- struct{}{}:
			defer func() { <-lb.workers }()
		}
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-lb.ctx.Done():
		return errors.New("load balancer is closed")
	// Finally, do the actual work
	default:
		return operation()
	}
}

// ExecuteBalanced is like Execute but for functions that actually return something
// Because sometimes void isn't good enough
func ExecuteBalanced[T any](lb *LoadBalancer, ctx context.Context, operation func() (T, error)) (T, error) {
	// In case everything goes wrong
	var zero T
	select {
	case lb.workers <- struct{}{}:
		defer func() { <-lb.workers }()
		select {
		case <-ctx.Done():
			return zero, ctx.Err()
		case <-lb.ctx.Done():
			return zero, errors.New("load balancer is closed")
		default:
			return operation()
		}
	case <-ctx.Done():
		return zero, ctx.Err()
	case <-lb.ctx.Done():
		return zero, errors.New("load balancer is closed")
	}
}

// Close tells everyone to go home
func (lb *LoadBalancer) Close() {
	lb.cancel()
}

// DefaultLoadBalancer creates a load balancer for the indecisive
// 10 workers because why not
func DefaultLoadBalancer() *LoadBalancer {
	return NewLoadBalancer(10)
}

// DefaultLoadBalancerWithContext is like DefaultLoadBalancer but with a curfew
func DefaultLoadBalancerWithContext(ctx context.Context) *LoadBalancer {
	return NewLoadBalancerWithContext(ctx, 10)
}

// Workers returns the channel controlling access
// But seriously, don't mess with this directly
func (lb *LoadBalancer) Workers() chan struct{} {
	return lb.workers
}

// Ctx returns the load balancer's context
// In case you need more ways to cancel things
func (lb *LoadBalancer) Ctx() context.Context {
	return lb.ctx
}
