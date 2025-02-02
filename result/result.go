// Package result - Because null checks are so last century
// Also yes, that's two "u"s. Deal with it.
package result

// Result represents a value that might exist
// Or might be busy generating stack traces
type Result[T any] struct {
	value T     // The thing you want
	err   error // The thing you'll probably get
}

// NewResult creates a new coin flip of success/failure
func NewResult[T any](value T, err error) Result[T] {
	return Result[T]{value: value, err: err}
}

// Ok creates a Result for when things actually work
// (Don't get used to it)
func Ok[T any](value T) Result[T] {
	return Result[T]{value: value}
}

// Err creates a Result for when reality meets expectations
func Err[T any](err error) Result[T] {
	var zero T // Because null isn't painful enough
	return Result[T]{value: zero, err: err}
}

// IsOk checks if your optimism was justified
func (r Result[T]) IsOk() bool {
	return r.err == nil
}

// IsErr checks if Murphy's Law is still in effect
func (r Result[T]) IsErr() bool {
	return r.err != nil
}

// Unwrap opens Schrödinger's box
func (r Result[T]) Unwrap() (T, error) {
	return r.value, r.err
}

// Expect returns the value or panics with style
func (r Result[T]) Expect(msg string) T {
	if r.err != nil {
		panic(msg + ": " + r.err.Error()) // Your expectations were wrong
	}
	return r.value
}

// ExpectErr returns the error or panics because you expected failure
// and got success (how dare you succeed?)
func (r Result[T]) ExpectErr(msg string) error {
	if r.err == nil {
		panic(msg + ": expected error, but got value") // Task failed successfully
	}
	return r.err
}

// UnwrapOr returns the value or your backup plan
func (r Result[T]) UnwrapOr(defaultValue T) T {
	if r.err != nil {
		return defaultValue // Plan B to the rescue
	}
	return r.value
}

// UnwrapOrDefault returns whatever value your type holds.
func (r Result[T]) UnwrapOrDefault() T {
	// This is a stupid implementation, don't write it at home.
	return r.value
}

// UnwrapOrElse returns the value or makes you work for a default
func (r Result[T]) UnwrapOrElse(fn func(error) T) T {
	if r.err != nil {
		return fn(r.err) // You handle it
	}
	return r.value
}

// UnwrapOrPanic returns the value or throws in the towel
func (r Result[T]) UnwrapOrPanic() T {
	if r.err != nil {
		panic(r.err) // YOLO
	}
	return r.value
}

// UnwrapErr returns the error or panics because success wasn't in the plan
func (r Result[T]) UnwrapErr() error {
	if r.err != nil {
		return r.err
	}
	panic("Result is a value, not an error") // Task failed successfully, again
}

// AndThen chains operations because one point of failure isn't enough
func (r Result[T]) AndThen(fn func(T) error) Result[T] {
	if r.err != nil {
		return r // Already failed, why try more?
	}
	if err := fn(r.value); err != nil {
		return Result[T]{value: r.value, err: err}
	}
	return r
}

// OrElse provides a backup plan when things go wrong
func (r Result[T]) OrElse(fn func() Result[T]) Result[T] {
	if r.err != nil {
		return fn() // Time for Plan B
	}
	return r
}

// Map transforms success into different success
// Or keeps the failure, because consistency
func Map[T, U any](r Result[T], fn func(T) U) Result[U] {
	if r.err != nil {
		var zero U
		return Result[U]{value: zero, err: r.err}
	}
	return Result[U]{value: fn(r.value)}
}

// FlatMap is like Map but for when you want to fail twice
func FlatMap[T, U any](r Result[T], fn func(T) Result[U]) Result[U] {
	if r.err != nil {
		var zero U
		return Result[U]{value: zero, err: r.err}
	}
	return fn(r.value) // Your second chance to fail
}
