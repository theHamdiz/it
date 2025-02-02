package math

import (
	"fmt"
	"math"

	"golang.org/x/exp/constraints"
)

// Sum calculates the sum of numbers from 1 to n (or n to 1 if n is negative) in O(1) time
// because who needs loops when you have math from 300 BC?
// Gauss would be proud, or maybe just mildly amused.
func Sum[T constraints.Integer](n T) T {
	// For n = 0, because sometimes nothing plus nothing is... nothing
	// Thanks to to u/Skeeve-on-git for highlighting negative sum bug.
	if n == 0 {
		return 0
	}

	// If you're wondering why this works:
	// 1. Gauss figured this out when he was 8
	// 2. You're probably older than 8
	// 3. Don't feel bad, he was Gauss
	return (n * (n + 1)) / 2
}

// SumSlow calculates the sum of numbers from 1 to n using a loop
// because some of us like burning CPU cycles for no good reason
// obviously this should not be used, it's just here for the memes.
func SumSlow[T constraints.Integer](n T) T {
	var total T
	// Explicitly cast 1 to type T
	one := T(1)
	// Thanks to to u/Skeeve-on-git for highlighting this.
	// Compute -1 as type T
	negOne := -one

	if n > 0 {
		for i := one; i <= n; i++ {
			total += i
		}
	} else {
		for i := n; i <= negOne; i++ {
			total += i
		}
	}
	return total
}

// SumRange calculates the sum of numbers from start to end inclusive
// because sometimes you don't want to start at 1 like a peasant
func SumRange[T constraints.Integer](start, end T) T {
	// If you're trying to sum backwards, we'll fix that for you
	// because we're nice, not because you're smart
	if start > end {
		start, end = end, start
	}

	if start >= 1 {
		return Sum(end) - Sum(start-1)
	}

	return Sum(end) - Sum(start-1)
}

// SumRangeSlow calculates the sum of numbers from start to end inclusive
// because some of us like burning CPU cycles for no good reason
// obviously this should not be used, it's just here for the memes.
func SumRangeSlow[T constraints.Integer](start, end T) T {
	// If you're trying to sum backwards, we'll fix that for you
	// because we're nice, not because you're smart
	if start > end {
		start, end = end, start
	}

	// The formula is: (end * (end + 1) / 2) - (start * (start - 1) / 2)
	// Don't worry if you don't understand it, neither do most people
	return SumSlow(end) - SumSlow(start-1)
}

// SumWithOverflowCheck does the same as Sum but checks for overflow
// because some people actually care about correctness
func SumWithOverflowCheck[T constraints.Signed](n T) (T, error) {
	// Cast to int64 so we can reuse the same overflow logic
	// (this obviously won't hold for types outside int64 range).
	n64 := int64(n)

	// Check if n is negative, because negativity is bad for mental health
	if n64 < 0 {
		return 0, fmt.Errorf("received %d, but my therapist says to stay positive", n64)
	}

	// Check for overflow because numbers have limits
	// unlike human stupidity
	if n64 > math.MaxInt64/2 {
		return 0, fmt.Errorf(
			"number %d is too big, try something smaller, like your expectations",
			n64,
		)
	}

	// Check if (n * (n + 1)) would overflow
	// because multiplication is sneaky like that
	if n64 > 0 && n64 > math.MaxInt64/(n64+1) {
		return 0, fmt.Errorf(
			"would overflow calculating sum to %d, "+
				"try breaking it into smaller problems, like your life goals",
			n64,
		)
	}

	sum64 := (n64 * (n64 + 1)) / 2
	return T(sum64), nil
}

// MustSum is like SumWithOverflowCheck but panics on error
// because some days you just want to watch the world burn
func MustSum[T constraints.Signed](n T) T {
	sum, err := SumWithOverflowCheck(n)
	if err != nil {
		panic(fmt.Sprintf("Sum failed: %v (this is your fault, not mine)", err))
	}
	return sum
}

// SumOfSquaresSlow calculates the sum of squares from 1 to n using a loop
// because some of us like burning CPU cycles for no good reason
// obviously this should not be used, it's just here for the memes.
func SumOfSquaresSlow[T constraints.Integer](n T) T {
	var total T
	for i := T(1); i <= n; i++ {
		total += i * i
	}
	return total
}

// SumOfSquares calculates the sum of squares from 1 to n using the formula
// n(n+1)(2n+1)/6 in O(1) time
// because there's always a formula if you know where to look
func SumOfSquares[T constraints.Integer](n T) T {
	return (n * (n + 1) * (2*n + 1)) / 6
}

// SumOfCubesSlow calculates the sum of cubes from 1 to n using a loop
// let's waste even more CPU cycles just because we can
func SumOfCubesSlow[T constraints.Integer](n T) T {
	var total T
	for i := T(1); i <= n; i++ {
		total += i * i * i
	}
	return total
}

// SumOfCubes calculates the sum of cubes from 1 to n using the formula
// [n(n+1)/2]^2 in O(1) time
// because squares weren't enough to show off our math prowess
func SumOfCubes[T constraints.Integer](n T) T {
	halfSum := (n * (n + 1)) / 2
	return halfSum * halfSum
}

// SumOfFourthPowersSlow calculates the sum of the fourth powers from 1 to n using a loop
// because apparently squares and cubes are just too basic for some people
func SumOfFourthPowersSlow[T constraints.Integer](n T) T {
	var total T
	for i := T(1); i <= n; i++ {
		total += i * i * i * i
	}
	return total
}

// SumOfFourthPowers calculates the sum of the fourth powers from 1 to n using the formula
// n(n+1)(2n+1)(3n^2+3n-1)/30 in O(1) time
// because there's always a bigger power to inflate your ego
func SumOfFourthPowers[T constraints.Integer](n T) T {
	// Yes, there's a pattern for sums of kth powers too,
	// but let's not get carried away and write them all... yet.
	return (n * (n + 1) * (2*n + 1) * ((3 * n * n) + (3 * n) - 1)) / 30
}

// ArithmeticSeriesSlow calculates the sum of an arithmetic sequence with a loop
// for those who really enjoy repetitive tasks
func ArithmeticSeriesSlow[T constraints.Integer](start, diff, terms T) T {
	var total T
	current := start
	for i := T(0); i < terms; i++ {
		total += current
		current += diff
	}
	return total
}

// ArithmeticSeries calculates the sum of an arithmetic sequence in O(1) time
// because the formula n/2 * (2a + (n-1)d) has existed since forever
func ArithmeticSeries[T constraints.Integer](start, diff, terms T) T {
	// S = n/2 * [2*a + (n-1)*d]
	return (terms * (2*start + (terms-1)*diff)) / 2
}

// GeometricSeriesSlow calculates the sum of a geometric series with a loop
// because sometimes you just want to multiply the same thing over and over again
func GeometricSeriesSlow[T constraints.Integer](start, ratio, terms T) T {
	var total T
	current := start
	for i := T(0); i < terms; i++ {
		total += current
		current *= ratio
	}
	return total
}

// GeometricSeries calculates the sum of a geometric series in O(log n) time
// ignoring the cost of exponentiation just for your mental comfort
// sum = a * (r^n - 1) / (r - 1), if r != 1
func GeometricSeries[T constraints.Integer](start, ratio, terms T) T {
	if ratio == 1 {
		// if ratio == 1, the series is just `start` added `terms` times
		return start * terms
	}
	power := Pow(ratio, terms)
	return start * (power - 1) / (ratio - 1)
}

// Pow computes base^exp using fast exponentiation in O(log exp) time
// because naive exponentiation is so last century
func Pow[T constraints.Integer](base, exp T) T {
	// Negative exponents in integer math? Let's not go there.
	if exp < 0 {
		panic("Pow does not support negative exponents for integer types. Please consult floating-point or your local wizard.")
	}

	result := T(1)
	temp := base
	e := exp

	for e > 0 {
		if e&1 == 1 {
			result *= temp
		}
		temp *= temp
		e >>= 1
	}
	return result
}

// FactorialSlow calculates n! in the most straightforward way
// by multiplying from 1 up to n in O(n) time
// because why do in O(1) what you can do in O(n)?
func FactorialSlow[T constraints.Integer](n T) T {
	n64 := int64(n)
	if n64 < 0 {
		panic(fmt.Sprintf("invalid input %d, factorial of negative numbers is about as real as unicorns", n64))
	}
	var result int64 = 1
	for i := int64(1); i <= n64; i++ {
		result *= i
	}
	return T(result)
}

// Factorial tries to do factorial "faster", but let's be honest,
// there's no real O(1) direct formula for factorial that gives exact integers.
// We'll just do a loop with an overflow check, so you can
// pretend you're being safe with large numbers.
func Factorial[T constraints.Signed](n T) (T, error) {
	n64 := int64(n)
	if n64 < 0 {
		return 0, fmt.Errorf("received %d, but factorial of a negative is about as helpful as negative emotions", n64)
	}

	var result int64 = 1
	for i := int64(1); i <= n64; i++ {
		// Overflow check: if result > math.MaxInt64 / i, multiplication would overflow
		if i != 0 && result > math.MaxInt64/i {
			return 0, fmt.Errorf("overflow alert! factorial of %d is too big for your puny type", n64)
		}
		result *= i
	}
	return T(result), nil
}

// FactorialStirlingApprox returns a float64 approximation of n!
// using Stirling's approximation: sqrt(2πn) * (n/e)^n
// because sometimes "close enough" is good enough for government work
func FactorialStirlingApprox[T constraints.Integer](n T) float64 {
	n64 := int64(n)
	if n64 < 0 {
		panic(fmt.Sprintf("received %d, but negative factorial approximations are not in this reality", n64))
	}
	if n64 == 0 || n64 == 1 {
		return 1
	}
	nFloat := float64(n64)
	return math.Sqrt(2*math.Pi*nFloat) * math.Pow(nFloat/math.E, nFloat)
}

// BinomialSlow calculates C(n, k) = n! / (k!(n-k)!) in the most naive way
// guaranteed to overflow for large n, just like your inbox
func BinomialSlow[T constraints.Signed](n, k T) (T, error) {
	n64 := int64(n)
	k64 := int64(k)

	if k64 < 0 || k64 > n64 {
		return 0, fmt.Errorf("c(%d, %d) is zero or not defined in normal combinatorics, kind of like your questionable assumptions", n64, k64)
	}

	fn, err := Factorial(n)
	if err != nil {
		return 0, err
	}
	fk, err := Factorial(k)
	if err != nil {
		return 0, err
	}
	fnk, err := Factorial(n - k)
	if err != nil {
		return 0, err
	}

	// We'll cast them to int64 for multiplication
	fn64 := int64(fn)
	fk64 := int64(fk)
	fnk64 := int64(fnk)

	if fk64 == 0 || fnk64 == 0 {
		return 0, fmt.Errorf("invalid factorial result, because apparently the math gods hate you")
	}
	divisor := fk64 * fnk64
	if divisor == 0 {
		return 0, fmt.Errorf("division by zero, your math teacher would be so proud")
	}
	if fn64 > math.MaxInt64 || divisor > math.MaxInt64 {
		return 0, fmt.Errorf("c(%d, %d) overflowed. Maybe find a bigger type or smaller dreams", n64, k64)
	}

	val := fn64 / divisor
	return T(val), nil
}

// Binomial calculates C(n, k) in O(k) without computing factorials directly.
// It's "faster" and less prone to immediate overflow than the naive approach
// but let's not pretend it won't blow up eventually for big n.
func Binomial[T constraints.Signed](n, k T) (T, error) {
	n64 := int64(n)
	k64 := int64(k)

	if k64 < 0 || k64 > n64 {
		return 0, fmt.Errorf("c(%d, %d) is about as valid as chasing unicorns", n64, k64)
	}
	if k64 == 0 || k64 == n64 {
		return 1, nil
	}
	// exploit symmetry: C(n, k) = C(n, n-k)
	if k64 > n64-k64 {
		k64 = n64 - k64
	}
	var result int64 = 1
	for i := int64(1); i <= k64; i++ {
		top := n64 - (k64 - i)
		result *= top
		result /= i
		// Not a perfect overflow check, but let's be dramatic
		if result > math.MaxInt64 {
			return 0, fmt.Errorf("overflow in the middle of combinatorial nirvana for C(%d, %d)", n64, k64)
		}
	}
	return T(result), nil
}

// FibonacciSlow calculates the nth Fibonacci number recursively
// in O(1.618^n) time, or something equally terrifying
// because we all love the idea of a stack overflow
func FibonacciSlow[T constraints.Integer](n T) T {
	n64 := int64(n)
	if n64 < 0 {
		panic(fmt.Sprintf("fibonacci of a negative (%d)? Do you also believe in negative time?", n64))
	}
	return T(fibonacciSlowSigned(n64))
}

// private helper so we can do safe recursion in int64
func fibonacciSlowSigned(n int64) int64 {
	if n < 2 {
		return n
	}
	return fibonacciSlowSigned(n-1) + fibonacciSlowSigned(n-2)
}

// fibFastDoublingSigned is a helper that returns (F(n), F(n+1)) using fast doubling
func fibFastDoublingSigned(n int64) (int64, int64) {
	if n == 0 {
		return 0, 1
	}
	a, b := fibFastDoublingSigned(n >> 1) // shift in int64 land
	c := a * (2*b - a)
	d := a*a + b*b
	if n&1 == 1 {
		return d, c + d
	}
	return c, d
}

// Fibonacci calculates the nth Fibonacci number using fast doubling
// in O(log n) time because life is too short for naive recursion
func Fibonacci[T constraints.Integer](n T) T {
	n64 := int64(n)
	if n64 < 0 {
		panic(fmt.Sprintf("fibonacci of a negative (%d)? I'd love to see that proof", n64))
	}
	f, _ := fibFastDoublingSigned(n64)
	return T(f)
}
