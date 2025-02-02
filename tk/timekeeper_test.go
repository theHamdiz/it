package tk_test

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/theHamdiz/it/tk"
)

// TestTimeKeeperBasic checks that TimeKeeper measures time and invokes optional callback.
func TestTimeKeeperBasic(t *testing.T) {
	var callbackCalled int32

	var callbackDuration time.Duration

	tk_ := tk.NewTimeKeeper("basic-test",
		tk.WithCallback(func(d time.Duration) {
			atomic.AddInt32(&callbackCalled, 1)
			callbackDuration = d
		}),
	)

	tk_.Start()
	time.Sleep(50 * time.Millisecond)
	measured := tk_.Stop()

	if measured < 50*time.Millisecond {
		t.Errorf("expected measured >= 50ms, got %v", measured)
	}

	if atomic.LoadInt32(&callbackCalled) != 1 {
		t.Errorf("expected callback called once, got %d", callbackCalled)
	}

	const tolerance = 5 * time.Millisecond
	diff := measured - callbackDuration
	if diff < 0 {
		diff = -diff
	}
	if diff > tolerance {
		t.Errorf("callback duration %v too far from measured %v (diff=%v)", callbackDuration, measured, diff)
	}
}

// TestTimeFn checks that TimeFn times the execution of a function and returns its result.
func TestTimeFn(t *testing.T) {
	result := tk.TimeFn("test-fn", func() string {
		time.Sleep(20 * time.Millisecond)
		return "hello"
	})

	if result != "hello" {
		t.Errorf("expected TimeFn to return 'hello', got %q", result)
	}
}

// TestAsyncTimeKeeper checks that AsyncTimeKeeper tracks multiple concurrent tasks.
func TestAsyncTimeKeeper(t *testing.T) {
	atk := tk.NewAsyncTimeKeeper("async-test")

	sleeps := []time.Duration{10 * time.Millisecond, 30 * time.Millisecond, 50 * time.Millisecond}

	for _, d := range sleeps {
		d := d
		atk.Track(func() {
			time.Sleep(d)
		})
	}

	durations := atk.Wait()

	if len(durations) != 3 {
		t.Errorf("expected 3 durations, got %d", len(durations))
	}

	// Each duration should be at least the sleep time for that goroutine,
	// though we can't guarantee the order concurrency-wise.
	for i, d := range durations {
		if d <= 0 {
			t.Errorf("duration %d is nonpositive: %v", i, d)
		}
	}
}

// TestAsyncTimeKeeperParallel checks concurrency in a slightly more complicated scenario
func TestAsyncTimeKeeperParallel(t *testing.T) {
	atk := tk.NewAsyncTimeKeeper("async-parallel")

	// We'll track 5 parallel tasks
	const numTasks = 5
	for i := 0; i < numTasks; i++ {
		atk.Track(func() {
			// do some "work"
			time.Sleep(10 * time.Millisecond)
		})
	}

	// The durations slice should have exactly 5 entries after Wait()
	durations := atk.Wait()
	if len(durations) != numTasks {
		t.Errorf("expected %d durations, got %d", numTasks, len(durations))
	}
	for i, d := range durations {
		if d < 10*time.Millisecond {
			t.Errorf("expected each task to take at least 10ms, got durations[%d] = %v", i, d)
		}
	}
}
