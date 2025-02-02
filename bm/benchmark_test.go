package bm_test

import (
	"testing"
	"time"

	"github.com/theHamdiz/it/bm"
)

func TestBenchmark(t *testing.T) {
	// Test case 1: Simple function with consistent timing
	t.Run("Consistent timing", func(t *testing.T) {
		result := bm.Benchmark("consistent", 100, func() {
			time.Sleep(time.Millisecond)
		})

		if result.Name != "consistent" {
			t.Errorf("Expected name 'consistent', got %s", result.Name)
		}
		if result.Iterations != 100 {
			t.Errorf("Expected 100 iterations, got %d", result.Iterations)
		}
		if result.Min < time.Millisecond {
			t.Errorf("Expected min >= 1ms, got %v", result.Min)
		}
	})

	// Test case 2: Zero iterations
	t.Run("Zero iterations", func(t *testing.T) {
		result := bm.Benchmark("zero", 0, func() {})

		if result.Name != "zero" {
			t.Errorf("Expected name 'zero', got %s", result.Name)
		}
		if result.Iterations != 0 {
			t.Errorf("Expected 0 iterations, got %d", result.Iterations)
		}
		if result.Min != 0 || result.Max != 0 || result.Average != 0 {
			t.Error("Expected all durations to be 0 for zero iterations")
		}
	})

	// Test case 3: Variable timing
	t.Run("Variable timing", func(t *testing.T) {
		count := 0
		result := bm.Benchmark("variable", 100, func() {
			if count%2 == 0 {
				time.Sleep(time.Millisecond)
			} else {
				time.Sleep(2 * time.Millisecond)
			}
			count++
		})

		if result.Min >= result.Max {
			t.Error("Expected min to be less than max")
		}
		if result.Average <= result.Min || result.Average >= result.Max {
			t.Error("Expected average to be between min and max")
		}
	})
}

func TestAnalyzeBenchmark(t *testing.T) {
	// Test case 1: Empty durations
	t.Run("Empty durations", func(t *testing.T) {
		result := bm.AnalyzeBenchmark("empty", []time.Duration{})

		if result.Name != "empty" {
			t.Errorf("Expected name 'empty', got %s", result.Name)
		}
		if result.Iterations != 0 {
			t.Errorf("Expected 0 iterations, got %d", result.Iterations)
		}
	})

	// Test case 2: Single duration
	t.Run("Single duration", func(t *testing.T) {
		duration := time.Millisecond
		result := bm.AnalyzeBenchmark("single", []time.Duration{duration})

		if result.Min != duration || result.Max != duration || result.Average != duration {
			t.Error("Expected all metrics to equal the single duration")
		}
		if result.Iterations != 1 {
			t.Errorf("Expected 1 iteration, got %d", result.Iterations)
		}
	})

	// Test case 3: Multiple identical durations
	t.Run("Multiple identical durations", func(t *testing.T) {
		duration := time.Millisecond
		durations := []time.Duration{duration, duration, duration}
		result := bm.AnalyzeBenchmark("identical", durations)

		if result.StdDev != 0 {
			t.Errorf("Expected 0 standard deviation for identical durations, got %v", result.StdDev)
		}
	})

	// Test case 4: Known distribution
	t.Run("Known distribution", func(t *testing.T) {
		durations := []time.Duration{
			1 * time.Millisecond,
			2 * time.Millisecond,
			3 * time.Millisecond,
			4 * time.Millisecond,
			5 * time.Millisecond,
		}
		result := bm.AnalyzeBenchmark("known", durations)

		expectedAvg := 3 * time.Millisecond
		if result.Average != expectedAvg {
			t.Errorf("Expected average of %v, got %v", expectedAvg, result.Average)
		}

		expectedMedian := 3 * time.Millisecond
		if result.Median != expectedMedian {
			t.Errorf("Expected median of %v, got %v", expectedMedian, result.Median)
		}
	})
}

// Benchmark the Benchmark function itself
func BenchmarkBenchmark(b *testing.B) {
	b.Run("Simple function", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			bm.Benchmark("test", 10, func() {
				time.Sleep(time.Microsecond)
			})
		}
	})
}

func TestBenchmarkResultString(t *testing.T) {
	result := bm.BenchmarkResult{
		Name:       "test",
		Min:        time.Millisecond,
		Max:        2 * time.Millisecond,
		Average:    time.Millisecond + (time.Millisecond / 2), // 1.5ms
		Median:     time.Millisecond + (time.Millisecond / 2), // 1.5ms
		StdDev:     time.Millisecond / 2,                      // 0.5ms
		Iterations: 100,
	}

	_ = result
}

// Table-driven tests for various scenarios
func TestBenchmarkScenarios(t *testing.T) {
	tests := []struct {
		name       string
		iterations int
		fn         func()
		wantMin    time.Duration
		wantErr    bool
	}{
		{
			name:       "Fast function",
			iterations: 100,
			fn:         func() {},
			wantMin:    0,
			wantErr:    false,
		},
		{
			name:       "Slow function",
			iterations: 10,
			fn: func() {
				time.Sleep(time.Millisecond)
			},
			wantMin: time.Millisecond,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := bm.Benchmark(tt.name, tt.iterations, tt.fn)
			if result.Min < tt.wantMin {
				t.Errorf("Benchmark() min = %v, want >= %v", result.Min, tt.wantMin)
			}
		})
	}
}

func TestBenchmarkEdgeCases(t *testing.T) {
	t.Run("Nil function", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic for nil function")
			}
		}()
		bm.Benchmark("nil", 1, nil)
	})

	t.Run("Negative iterations", func(t *testing.T) {
		result := bm.Benchmark("negative", -1, func() {})
		if result.Iterations != 0 {
			t.Error("Expected 0 iterations for negative input")
		}
	})

	t.Run("Large iterations", func(t *testing.T) {
		result := bm.Benchmark("large", 1000000, func() {})
		if result.Iterations != 1000000 {
			t.Error("Failed to handle large number of iterations")
		}
	})
}
