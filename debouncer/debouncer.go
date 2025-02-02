// Package debouncer - For functions that need to chill out and stop being so eager
package debouncer

import (
	"sync"
	"time"
)

// Debouncer is like a bouncer for your function calls
// Keeps the eager ones waiting outside until the VIPs have left
type Debouncer struct {
	mu    sync.Mutex    // The velvet rope
	timer *time.Timer   // The "maybe later" timer
	delay time.Duration // How long we make them wait
}

// NewDebouncer creates a new function cooldown manager
// delay: how long until we're ready to party again
func NewDebouncer(delay time.Duration) *Debouncer {
	return &Debouncer{
		delay: delay, // The mandatory cool-off period
	}
}

// Debounce wraps your hyperactive function in a calm, collected exterior
// Returns a function that's learned some patience
func (d *Debouncer) Debounce(fn func()) func() {
	return func() {
		d.mu.Lock()
		defer d.mu.Unlock()

		if d.timer != nil {
			// Sorry, we're resetting the queue
			d.timer.Stop()
		}
		// Come back later
		d.timer = time.AfterFunc(d.delay, fn)
	}
}

// Cancel tells everyone to go home, party's over
func (d *Debouncer) Cancel() {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.timer != nil {
		d.timer.Stop()
		// Clean up after ourselves
		d.timer = nil
	}
}

// Reset is like telling everyone "new plan, different waiting time"
func (d *Debouncer) Reset(delay time.Duration) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.delay = delay
	if d.timer != nil {
		// Surprise! Wait longer
		d.timer.Reset(delay)
	}
}

// SetDelay changes how long we make functions wait
// For when the current timeout isn't painful enough
func (d *Debouncer) SetDelay(delay time.Duration) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.delay = delay
}

// Delay returns how long we're making functions wait
// Spoiler: It's probably longer than they want
func (d *Debouncer) Delay() time.Duration {
	d.mu.Lock()
	defer d.mu.Unlock()

	return d.delay
}

// IsRunning checks if we're currently making something wait
func (d *Debouncer) IsRunning() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	// Are we ghosting any functions?
	return d.timer != nil
}

// IsStopped is the opposite of IsRunning
// Because sometimes you need to know if nothing's happening
func (d *Debouncer) IsStopped() bool {
	return !d.IsRunning()
}

// Timer returns the actual timer
// But seriously, you probably shouldn't mess with this
func (d *Debouncer) Timer() *time.Timer {
	d.mu.Lock()
	defer d.mu.Unlock()

	return d.timer
}

// Stop is like Cancel but sounds more professional
func (d *Debouncer) Stop() {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.timer != nil {
		d.timer.Stop()
		// Goodbye timer, we hardly knew ye
		d.timer = nil
	}
}
