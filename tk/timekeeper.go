package tk

import (
	"sync"
	"time"

	logger2 "github.com/theHamdiz/it/logger"
)

// TimeKeeper tracks execution time because time is money,
// and we're all about that ROI
type TimeKeeper struct {
	start    time.Time
	name     string
	logger   *logger2.Logger
	callback func(duration time.Duration)
}

// NewTimeKeeper creates a new timekeeper because someone has to
// watch the clock
func NewTimeKeeper(name string, opts ...TimeKeeperOption) *TimeKeeper {
	tk := &TimeKeeper{
		name:   name,
		logger: logger2.DefaultLogger(),
	}
	for _, opt := range opts {
		opt(tk)
	}
	return tk
}

type TimeKeeperOption func(*TimeKeeper)

// WithCallback adds a callback function because sometimes you want
// to do more than just log
func WithCallback(cb func(duration time.Duration)) TimeKeeperOption {
	return func(tk *TimeKeeper) {
		tk.callback = cb
	}
}

// Start begins timing because every journey begins with a single step
func (tk *TimeKeeper) Start() *TimeKeeper {
	tk.start = time.Now()
	return tk
}

// Stop ends timing and logs the duration because all good things
// must come to an end
func (tk *TimeKeeper) Stop() time.Duration {
	duration := time.Since(tk.start)
	tk.logger.Infof("⏱️ %s took %v", tk.name, duration)
	if tk.callback != nil {
		tk.callback(duration)
	}
	return duration
}

// TimeFn wraps a function with timing because knowing how long
// things take is occasionally useful
func TimeFn[T any](name string, fn func() T) T {
	tk := NewTimeKeeper(name).Start()
	defer tk.Stop()
	return fn()
}

// AsyncTimeKeeper tracks concurrent operations because parallel
// timing is twice the fun
type AsyncTimeKeeper struct {
	timekeeper *TimeKeeper
	wg         sync.WaitGroup
	durations  []time.Duration
	mu         sync.Mutex
}

// NewAsyncTimeKeeper creates a new async timekeeper because
// concurrent timing needs special handling
func NewAsyncTimeKeeper(name string) *AsyncTimeKeeper {
	return &AsyncTimeKeeper{
		timekeeper: NewTimeKeeper(name),
	}
}

// Track adds a new operation to track because keeping track of
// parallel operations is like herding cats
func (atk *AsyncTimeKeeper) Track(fn func()) {
	atk.wg.Add(1)
	start := time.Now()

	go func() {
		defer func() {
			duration := time.Since(start)
			atk.mu.Lock()
			atk.durations = append(atk.durations, duration)
			atk.mu.Unlock()
			atk.wg.Done()
		}()
		fn()
	}()
}

// Wait waits for all operations to complete because patience
// is still a virtue
func (atk *AsyncTimeKeeper) Wait() []time.Duration {
	atk.wg.Wait()
	return atk.durations
}
