package math_test

import (
	"fmt"
	"math"
	"testing"
	"time"

	math2 "github.com/theHamdiz/it/math"
)

// Test that formula-based methods produce the same result as the slow methods.
// We won't test SumRange here again since it has its own specialized test.
func TestCorrectness(t *testing.T) {
	tests := []struct {
		name       string
		slowFn     func(int64) int64
		fastFn     func(int64) int64
		testValues []int64
	}{
		{
			name:       "Sum vs SumSlow",
			slowFn:     math2.SumSlow[int64],
			fastFn:     math2.Sum[int64],
			testValues: []int64{0, 1, 5, 10, 100, 999},
		},
		{
			name:       "SumOfSquares vs SumOfSquaresSlow",
			slowFn:     math2.SumOfSquaresSlow[int64],
			fastFn:     math2.SumOfSquares[int64],
			testValues: []int64{0, 1, 5, 10, 100, 999},
		},
		{
			name:       "SumOfCubes vs SumOfCubesSlow",
			slowFn:     math2.SumOfCubesSlow[int64],
			fastFn:     math2.SumOfCubes[int64],
			testValues: []int64{0, 1, 5, 10, 50, 100},
		},
		{
			name:       "SumOfFourthPowers vs SumOfFourthPowersSlow",
			slowFn:     math2.SumOfFourthPowersSlow[int64],
			fastFn:     math2.SumOfFourthPowers[int64],
			testValues: []int64{0, 1, 5, 10, 20},
		},
	}

	for _, tc := range tests {
		tc := tc // capture range variable
		t.Run(tc.name, func(t *testing.T) {
			for _, val := range tc.testValues {
				gotSlow := tc.slowFn(val)
				gotFast := tc.fastFn(val)
				if gotSlow != gotFast {
					t.Errorf("[%s] mismatch for n=%d: slow=%d fast=%d",
						tc.name, val, gotSlow, gotFast)
				}
			}
		})
	}
}

// TestSumRangeCorrectness checks that SumRange and SumRangeSlow produce
// the same results, even for negative values.
func TestSumRangeCorrectness(t *testing.T) {
	// Thanks to to u/Skeeve-on-git for highlighting negative sum bug.
	type testCase struct {
		start int64
		end   int64
		want  int64
	}

	cases := []testCase{
		// Zero cases
		{start: 0, end: 0, want: 0},
		{start: -0, end: 0, want: 0},
		{start: 0, end: -0, want: 0},

		// Single number cases
		{start: 1, end: 1, want: 1},
		{start: -1, end: -1, want: -1},
		{start: 42, end: 42, want: 42},

		// Small positive ranges
		{start: 1, end: 5, want: 15},
		{start: 3, end: 7, want: 25},
		{start: 1, end: 10, want: 55},

		// Small negative ranges
		{start: -5, end: -1, want: -15},
		{start: -7, end: -3, want: -25},
		{start: -10, end: -1, want: -55},

		// Mixed ranges crossing zero
		{start: -5, end: 5, want: 0},
		{start: -3, end: 3, want: 0},
		{start: -2, end: 2, want: 0},
		{start: -10, end: 10, want: 0},

		// Asymmetric mixed ranges
		{start: -3, end: 5, want: 9},
		{start: -7, end: 4, want: -18},
		{start: -2, end: 8, want: 33},

		// Reversed ranges (testing swap logic)
		{start: 5, end: 1, want: 15},
		{start: -1, end: -5, want: -15},
		{start: 5, end: -5, want: 0},
		{start: 10, end: -10, want: 0},

		// Larger ranges
		{start: 1, end: 100, want: 5050},
		{start: -100, end: -1, want: -5050},
		{start: -50, end: 50, want: 0},
		{start: 999, end: 1000, want: 1999},

		// Adjacent numbers
		{start: -2, end: -1, want: -3},
		{start: -1, end: 0, want: -1},
		{start: 0, end: 1, want: 1},
		{start: 1, end: 2, want: 3},

		// Gaps
		{start: -10, end: -5, want: -45},
		{start: 5, end: 10, want: 45},
		{start: -3, end: 2, want: -3},

		{
			// int32 overflow check
			start: 65535,
			end:   65536,
			want:  131071,
		},
		{
			// int32 similar case with negative numbers
			start: -65536,
			end:   -65535,
			want:  -131071,
		},
		{
			// Large range where intermediate would overflow on int32
			start: 46340,
			end:   46341,
			want:  92681,
		},
	}

	for _, c := range cases {
		gotSlow := sumRangeSlow(c.start, c.end)
		gotFast := math2.SumRange[int64](c.start, c.end)

		if gotSlow != c.want {
			t.Errorf("sumRangeSlow(%d, %d) = %d; want %d",
				c.start, c.end, gotSlow, c.want)
		}
		if gotSlow != gotFast {
			t.Errorf("SumRange mismatch for start=%d end=%d: slow=%d fast=%d",
				c.start, c.end, gotSlow, gotFast)
		}
	}
}

// sumRangeSlow is a corrected slow version that literally loops from start to end
// (including negatives) so it matches the actual sum of all integers in [start, end].
func sumRangeSlow(start, end int64) int64 {
	if start > end {
		start, end = end, start
	}
	var total int64
	for i := start; i <= end; i++ {
		total += i
	}
	return total
}

// TestArithmeticSeriesCorrectness checks that the arithmetic series formula
// matches the slow loop approach for a variety of inputs.
func TestArithmeticSeriesCorrectness(t *testing.T) {
	type testCase struct {
		start, diff, terms int64
	}
	cases := []testCase{
		{start: 0, diff: 0, terms: 10},
		{start: 1, diff: 1, terms: 1},
		{start: 2, diff: 3, terms: 5},
		{start: -5, diff: 2, terms: 5},
		{start: 100, diff: -10, terms: 20},
	}

	for _, c := range cases {
		want := math2.ArithmeticSeriesSlow[int64](c.start, c.diff, c.terms)
		got := math2.ArithmeticSeries[int64](c.start, c.diff, c.terms)
		if want != got {
			t.Errorf("ArithmeticSeries mismatch for start=%d diff=%d terms=%d: slow=%d fast=%d",
				c.start, c.diff, c.terms, want, got)
		}
	}
}

func TestGeometricSeriesCorrectness(t *testing.T) {
	type testCase struct {
		start, ratio, terms int64
	}
	cases := []testCase{
		{start: 1, ratio: 1, terms: 10},   // sum should be 10
		{start: 2, ratio: 2, terms: 5},    // 2 +4 +8 +16 +32=62
		{start: 3, ratio: 3, terms: 3},    // 3 +9 +27=39
		{start: 10, ratio: 1, terms: 100}, // 1000
		{start: 5, ratio: -1, terms: 4},   // 5 + -5 + 5 + -5=0
	}

	for _, c := range cases {
		want := math2.GeometricSeriesSlow[int64](c.start, c.ratio, c.terms)
		got := math2.GeometricSeries[int64](c.start, c.ratio, c.terms)
		if want != got {
			t.Errorf("GeometricSeries mismatch for start=%d ratio=%d terms=%d: slow=%d fast=%d",
				c.start, c.ratio, c.terms, want, got)
		}
	}
}

func TestMustSumAndOverflowCheck(t *testing.T) {
	// We test a few boundary cases. We'll use subtests
	// so we don't rely on defer within a loop.
	tests := []int64{-1, 0, 1, 10, math.MaxInt16, math.MaxInt32}
	for _, n := range tests {
		n := n // capture
		t.Run(
			// name for subtest
			func() string {
				return "SumWithOverflowCheck=" + itSum(n)
			}(),
			func(st *testing.T) {
				got, err := math2.SumWithOverflowCheck[int64](n)

				if n < 0 {
					if err == nil {
						st.Errorf("SumWithOverflowCheck(%d) expected error, got no error", n)
					}

					// Now let's check MustSum in a separate subtest
					st.Run("MustSum-negative", func(sst *testing.T) {
						defer func() {
							if r := recover(); r == nil {
								sst.Errorf("MustSum(%d) expected panic, got none", n)
							}
						}()
						_ = math2.MustSum[int64](n)
					})

				} else {
					if err != nil {
						st.Errorf("SumWithOverflowCheck(%d) returned error: %v", n, err)
					}
					want := math2.SumSlow[int64](n) // baseline check
					if got != want {
						st.Errorf("SumWithOverflowCheck(%d) = %d; want %d", n, got, want)
					}
					// Check MustSum doesn't panic
					st.Run("MustSum-non-negative", func(sst *testing.T) {
						defer func() {
							if r := recover(); r != nil {
								sst.Errorf("MustSum(%d) panicked unexpectedly: %v", n, r)
							}
						}()
						mustVal := math2.MustSum[int64](n)
						if mustVal != want {
							sst.Errorf("MustSum(%d) = %d; want %d", n, mustVal, want)
						}
					})
				}
			},
		)
	}
}

func TestFactorial(t *testing.T) {
	// Basic factorial correctness checks
	cases := []int64{0, 1, 2, 3, 4, 5, 10}
	for _, n := range cases {
		n := n
		t.Run("Factorial", func(st *testing.T) {
			gotSlow := math2.FactorialSlow[int64](n)
			gotFast, err := math2.Factorial[int64](n)
			if err != nil {
				st.Errorf("Factorial(%d) gave error: %v", n, err)
				return
			}
			if gotSlow != gotFast {
				st.Errorf("Factorial mismatch for n=%d: slow=%d fast=%d", n, gotSlow, gotFast)
			}
		})
	}
}

func TestBinomial(t *testing.T) {
	// small checks
	type input struct{ n, k int64 }
	cases := []input{
		{0, 0}, {1, 0}, {5, 2}, {5, 3}, {6, 3}, {10, 5},
	}
	for _, c := range cases {
		c := c
		t.Run("C(n,k)", func(st *testing.T) {
			want, err1 := math2.BinomialSlow[int64](c.n, c.k)
			if err1 != nil {
				st.Errorf("BinomialSlow(%d,%d) error: %v", c.n, c.k, err1)
				return
			}
			got, err2 := math2.Binomial[int64](c.n, c.k)
			if err2 != nil {
				st.Errorf("Binomial(%d,%d) error: %v", c.n, c.k, err2)
				return
			}
			if want != got {
				st.Errorf("C(%d,%d) mismatch: slow=%d, fast=%d", c.n, c.k, want, got)
			}
		})
	}
}

func TestFibonacci(t *testing.T) {
	// Just small correctness checks
	for _, n := range []int64{0, 1, 2, 3, 5, 10, 15, 20} {
		n := n
		t.Run("Fibonacci", func(st *testing.T) {
			want := fibNaiveCheck(n)
			gotSlow := math2.FibonacciSlow[int64](n)
			gotFast := math2.Fibonacci[int64](n)
			if gotSlow != want {
				st.Errorf("FibonacciSlow(%d) = %d; want %d", n, gotSlow, want)
			}
			if gotFast != want {
				st.Errorf("Fibonacci(%d) = %d; want %d", n, gotFast, want)
			}
		})
	}
}

// fibNaiveCheck is a simple iterative approach for checking small Fibonacci numbers.
func fibNaiveCheck(n int64) int64 {
	if n < 2 {
		return n
	}
	a, b := int64(0), int64(1)
	for i := int64(2); i <= n; i++ {
		a, b = b, a+b
	}
	return b
}

// -------------------------------------------------------
//   Basic Performance Tests (naive approach)
//   to demonstrate formula-based methods are faster
//   than their loop-based counterparts for large n
// -------------------------------------------------------

// TestPerformanceSum checks that the formula-based Sum is faster than SumSlow.
func TestPerformanceSum(t *testing.T) {
	// Choose a large input that won't explode your CPU, but big enough to measure
	const large = 10_000_000

	t1 := time.Now()
	_ = math2.Sum[int64](large)
	formulaTime := time.Since(t1)

	t2 := time.Now()
	_ = math2.SumSlow[int64](large)
	loopTime := time.Since(t2)

	if loopTime <= formulaTime {
		t.Errorf("Expected SumSlow to be slower than Sum for n=%d\n  Sum time=%v\n  SumSlow time=%v", large, formulaTime, loopTime)
	}
}

// TestPerformanceSquares checks that the formula-based SumOfSquares is faster than SumOfSquaresSlow.
func TestPerformanceSquares(t *testing.T) {
	const large = 5_000_000

	t1 := time.Now()
	_ = math2.SumOfSquares[int64](large)
	formulaTime := time.Since(t1)

	t2 := time.Now()
	_ = math2.SumOfSquaresSlow[int64](large)
	slowTime := time.Since(t2)

	if slowTime <= formulaTime {
		t.Errorf("Expected SumOfSquaresSlow to be slower than SumOfSquares for n=%d\n  formula=%v\n  slow=%v", large, formulaTime, slowTime)
	}
}

// TestPerformanceCubes checks that SumOfCubes is faster than SumOfCubesSlow for large n.
func TestPerformanceCubes(t *testing.T) {
	const large = 2_000_000

	t1 := time.Now()
	_ = math2.SumOfCubes[int64](large)
	formulaTime := time.Since(t1)

	t2 := time.Now()
	_ = math2.SumOfCubesSlow[int64](large)
	slowTime := time.Since(t2)

	if slowTime <= formulaTime {
		t.Errorf("Expected SumOfCubesSlow to be slower than SumOfCubes for n=%d\n  formula=%v\n  slow=%v", large, formulaTime, slowTime)
	}
}

// TestPerformanceFourthPowers checks that SumOfFourthPowers is faster than SumOfFourthPowersSlow.
func TestPerformanceFourthPowers(t *testing.T) {
	// 1e5 is enough to see a difference
	const large = 100_000

	t1 := time.Now()
	_ = math2.SumOfFourthPowers[int64](large)
	formulaTime := time.Since(t1)

	t2 := time.Now()
	_ = math2.SumOfFourthPowersSlow[int64](large)
	slowTime := time.Since(t2)

	if slowTime <= formulaTime {
		t.Errorf("Expected SumOfFourthPowersSlow to be slower than SumOfFourthPowers for n=%d\n  formula=%v\n  slow=%v", large, formulaTime, slowTime)
	}
}

// Just a tiny helper so we can name subtests more conveniently
func itSum(n int64) string {
	return "n=" + formatInt(n)
}

func formatInt(n int64) string {
	return fmt.Sprintf("%d", n)
}
