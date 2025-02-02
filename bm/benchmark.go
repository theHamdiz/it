// Package bm - Because measuring performance makes you feel better about your terrible code
package bm

import (
	"math"
	"sort"
	"strconv"
	"time"
)

// BenchmarkResult contains all the stats you'll use to lie to your boss about performance
type BenchmarkResult struct {
	Name       string        // What you're trying to prove is "fast enough"
	Min        time.Duration // That one lucky run
	Max        time.Duration // The run where GC decided to party
	Average    time.Duration // The number you'll actually show people
	Median     time.Duration // For when averages make you look bad
	StdDev     time.Duration // Proof that your benchmark is totally stable*
	Iterations int           // How many times you tried to prove yourself right
}

// Benchmark runs your function multiple times until the numbers look good
// or until you give up and decide it's "fast enough for now"
func Benchmark(name string, iterations int, fn func()) BenchmarkResult {
	var durations []time.Duration

	for i := 0; i < iterations; i++ {
		start := time.Now()
		fn()
		durations = append(durations, time.Since(start))
	}

	return AnalyzeBenchmark(name, durations)
}

// AnalyzeBenchmark does math you probably learned in school but forgot
// returns stats that will make your function look better than it is
func AnalyzeBenchmark(name string, durations []time.Duration) BenchmarkResult {
	if len(durations) == 0 {
		// Empty stats are the best stats
		return BenchmarkResult{Name: name}
	}

	sort.Slice(durations, func(i, j int) bool {
		return durations[i] < durations[j]
	})

	var sum time.Duration
	for _, d := range durations {
		sum += d
	}

	avg := sum / time.Duration(len(durations))
	median := durations[len(durations)/2]

	var variance float64
	for _, d := range durations {
		diff := float64(d - avg)
		variance += diff * diff
	}
	variance /= float64(len(durations))
	stdDev := time.Duration(math.Sqrt(variance))

	return BenchmarkResult{
		Name:       name,
		Min:        durations[0],                // The number you'll quote
		Max:        durations[len(durations)-1], // The number you'll ignore
		Average:    avg,                         // For the PowerPoint slides
		Median:     median,                      // For when the average looks bad
		StdDev:     stdDev,                      // Nobody understands this anyway
		Iterations: len(durations),              // Bigger = more legitimate, right?
	}
}

// String converts your benchmark results into something you can paste in Slack
func (b BenchmarkResult) String() string {
	return b.Name + ": " +
		"min=" + b.Min.String() + ", " +
		"max=" + b.Max.String() + ", " +
		"avg=" + b.Average.String() + ", " +
		"median=" + b.Median.String() + ", " +
		"stddev=" + b.StdDev.String() + ", " +
		"iterations=" + strconv.Itoa(b.Iterations)
}

// MarshalJSON because apparently everything needs to be JSON these days
func (b BenchmarkResult) MarshalJSON() ([]byte, error) {
	return []byte(`"` + b.String() + `"`), nil
}
