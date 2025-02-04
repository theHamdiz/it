// Package result - Because nil checks are so last century
package result

import (
	"errors"
	"fmt"
)

// Result represents a value that might exist
// Or might be busy generating stack traces
type Result[T any] struct {
	value T     // The thing you want
	err   error // The thing you'll probably get
}

// Pair holds two values of potentially different types
type Pair[T, U any] struct {
	First  T
	Second U
}

// Option represents a value that may or may not exist
type Option[T any] struct {
	value T
	valid bool
}

// Some creates an Option containing a value
func Some[T any](value T) Option[T] {
	return Option[T]{
		value: value,
		valid: true,
	}
}

// None creates an empty Option
func None[T any]() Option[T] {
	return Option[T]{
		valid: false,
	}
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
	var zero T // Because nil isn't painful enough
	return Result[T]{value: zero, err: err}
}

// Err returns the error if present, nil otherwise
func (r Result[T]) Err() error {
	return r.err
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

// Match is for when you want to handle both cases elegantly
func (r Result[T]) Match(ok func(T), err func(error)) {
	if r.IsErr() {
		err(r.err)
	} else {
		ok(r.value)
	}
}

// Filter turns success into failure if a condition isn't met
func (r Result[T]) Filter(predicate func(T) bool) Result[T] {
	if !r.IsOk() || !predicate(r.value) {
		return Err[T](errors.New("predicate failed"))
	}
	return r
}

// Inspect lets you peek at success without changing it
func (r Result[T]) Inspect(fn func(T)) Result[T] {
	if r.IsOk() {
		fn(r.value)
	}
	return r
}

// InspectErr lets you peek at failure without changing it
func (r Result[T]) InspectErr(fn func(error)) Result[T] {
	if r.IsErr() {
		fn(r.err)
	}
	return r
}

// Transpose converts Result[Option[T]] to Option[Result[T]]
func Transpose[T any](r Result[Option[T]]) Option[Result[T]] {
	if r.IsErr() {
		return Some(Err[T](r.err))
	}
	opt := r.UnwrapOrPanic()
	if opt.IsNone() {
		return None[Result[T]]()
	}
	return Some(Ok(opt.UnwrapOrPanic()))
}

// Collect turns a slice of Results into a Result of slice
func Collect[T any](results []Result[T]) Result[[]T] {
	values := make([]T, 0, len(results))
	for _, r := range results {
		if r.IsErr() {
			return Err[[]T](r.err)
		}
		values = append(values, r.UnwrapOrPanic())
	}
	return Ok(values)
}

// Try converts a panic into a Result
func Try[T any](fn func() T) Result[T] {
	var result T
	var err error

	func() {
		defer func() {
			if r := recover(); r != nil {
				switch v := r.(type) {
				case error:
					err = v
				case string:
					err = errors.New(v)
				default:
					err = fmt.Errorf("panic: %v", r)
				}
			}
		}()
		result = fn()
	}()

	if err != nil {
		return Err[T](err)
	}
	return Ok(result)
}

// FromOption converts an Option to a Result with custom error
func FromOption[T any](opt Option[T], err error) Result[T] {
	if opt.IsNone() {
		return Err[T](err)
	}
	return Ok(opt.UnwrapOrPanic())
}

// Zip combines two Results into one
func Zip[T, U any](r1 Result[T], r2 Result[U]) Result[Pair[T, U]] {
	if r1.IsErr() {
		return Err[Pair[T, U]](r1.err)
	}
	if r2.IsErr() {
		return Err[Pair[T, U]](r2.err)
	}
	return Ok(Pair[T, U]{First: r1.value, Second: r2.value})
}

// IsNone returns true if the Option contains no value
func (o Option[T]) IsNone() bool {
	return !o.valid
}

// IsSome returns true if the Option contains a value
func (o Option[T]) IsSome() bool {
	return o.valid
}

// UnwrapOr returns the contained value or a default
func (o Option[T]) UnwrapOr(defaultValue T) T {
	if o.valid {
		return o.value
	}
	return defaultValue
}

// UnwrapOrPanic returns the contained value or panics
func (o Option[T]) UnwrapOrPanic() T {
	if !o.valid {
		panic("called UnwrapOrPanic on None value")
	}
	return o.value
}
