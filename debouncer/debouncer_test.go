package debouncer_test

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/theHamdiz/it/debouncer"
)

func TestNewDebouncer(t *testing.T) {
	delay := 100 * time.Millisecond
	d := debouncer.NewDebouncer(delay)

	if d == nil {
		t.Fatal("Expected debouncer instance, got nil")
	}

	if d.Delay() != delay {
		t.Errorf("Expected delay to be %v, got %v", delay, d.Delay())
	}

	if d.Timer() != nil {
		t.Errorf("Expected timer to be nil initially, but it is not")
	}
}

func TestDebouncer_Debounce(t *testing.T) {
	var executed int32
	delay := 50 * time.Millisecond
	d := debouncer.NewDebouncer(delay)
	debouncedFn := d.Debounce(func() {
		atomic.AddInt32(&executed, 1)
	})

	debouncedFn()
	debouncedFn()
	debouncedFn()

	time.Sleep(2 * delay)

	if atomic.LoadInt32(&executed) != 1 {
		t.Errorf("Expected function to execute once, but executed %d times", executed)
	}
}

func TestDebouncer_Debounce_ResetTimer(t *testing.T) {
	var executed int32
	delay := 100 * time.Millisecond
	d := debouncer.NewDebouncer(delay)
	debouncedFn := d.Debounce(func() {
		atomic.AddInt32(&executed, 1)
	})

	debouncedFn()
	time.Sleep(50 * time.Millisecond)
	debouncedFn()

	time.Sleep(2 * delay)

	if atomic.LoadInt32(&executed) != 1 {
		t.Errorf("Expected function to execute once, but executed %d times", executed)
	}
}

func TestDebouncer_Debounce_MultipleExecutions(t *testing.T) {
	var executed int32
	delay := 30 * time.Millisecond
	d := debouncer.NewDebouncer(delay)
	debouncedFn := d.Debounce(func() {
		atomic.AddInt32(&executed, 1)
	})

	debouncedFn()
	debouncedFn()
	time.Sleep(2 * delay)

	debouncedFn()
	debouncedFn()
	time.Sleep(2 * delay)

	if atomic.LoadInt32(&executed) != 2 {
		t.Errorf("Expected function to execute twice, but executed %d times", executed)
	}
}
