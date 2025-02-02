package result_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/theHamdiz/it/result"
)

// TestNewResult ensures that NewResult correctly initializes a Result.
func TestNewResult(t *testing.T) {
	res := result.NewResult(42, nil)
	if res.IsErr() {
		t.Errorf("Expected Result to be Ok, got error: %v", res.UnwrapErr())
	}
	value, err := res.Unwrap()
	if err != nil || value != 42 {
		t.Errorf("Expected (42, nil), got (%v, %v)", value, err)
	}

	errRes := result.NewResult(0, errors.New("test error"))
	if errRes.IsOk() {
		t.Errorf("Expected Result to be Err, got Ok")
	}
	if errRes.UnwrapErr() == nil {
		t.Errorf("Expected an error, got nil")
	}
}

// TestOk ensures Ok() correctly creates a success result.
func TestOk(t *testing.T) {
	res := result.Ok("hello")
	if res.IsErr() {
		t.Errorf("Expected Ok Result, got Err: %v", res.UnwrapErr())
	}
	if res.UnwrapOr("default") != "hello" {
		t.Errorf("Expected 'hello', got '%v'", res.UnwrapOr("default"))
	}
}

// TestErr ensures Err() correctly creates an error result.
func TestErr(t *testing.T) {
	err := errors.New("some error")
	res := result.Err[int](err)
	if res.IsOk() {
		t.Errorf("Expected Err Result, got Ok")
	}
	if !errors.Is(err, res.UnwrapErr()) {
		t.Errorf("Expected error '%v', got '%v'", err, res.UnwrapErr())
	}
}

// TestUnwrap ensures Unwrap returns the correct value or error.
func TestUnwrap(t *testing.T) {
	res := result.Ok(10)
	value, err := res.Unwrap()
	if err != nil || value != 10 {
		t.Errorf("Expected (10, nil), got (%v, %v)", value, err)
	}

	errRes := result.Err[int](errors.New("unwrap error"))
	_, err = errRes.Unwrap()
	if err == nil {
		t.Errorf("Expected an error, got nil")
	}
}

// TestUnwrapOr ensures UnwrapOr returns the default value when Result is an error.
func TestUnwrapOr(t *testing.T) {
	okRes := result.Ok(7)
	errRes := result.Err[int](errors.New("fallback test"))

	if okRes.UnwrapOr(100) != 7 {
		t.Errorf("Expected 7, got %v", okRes.UnwrapOr(100))
	}

	if errRes.UnwrapOr(100) != 100 {
		t.Errorf("Expected fallback value 100, got %v", errRes.UnwrapOr(100))
	}
}

// TestUnwrapOrElse ensures UnwrapOrElse computes a default value when Result is an error.
func TestUnwrapOrElse(t *testing.T) {
	errRes := result.Err[int](errors.New("fallback needed"))
	result_ := errRes.UnwrapOrElse(func(err error) int {
		return 999
	})
	if result_ != 999 {
		t.Errorf("Expected 999, got %v", result_)
	}
}

// TestUnwrapOrPanic ensures UnwrapOrPanic panics when Result is an error.
func TestUnwrapOrPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic, but it did not occur")
		}
	}()
	errRes := result.Err[int](errors.New("panic test"))
	// Should panic
	errRes.UnwrapOrPanic()
}

// TestUnwrapErr ensures UnwrapErr returns the error or panics if the Result is Ok.
func TestUnwrapErr(t *testing.T) {
	errRes := result.Err[int](errors.New("unwrap error test"))
	if errRes.UnwrapErr() == nil {
		t.Errorf("Expected an error, got nil")
	}

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic on UnwrapErr for Ok result, but it did not occur")
		}
	}()
	okRes := result.Ok(42)
	// Should panic
	_ = okRes.UnwrapErr()
}

// TestMap ensures Map correctly transforms an Ok Result.
func TestMap(t *testing.T) {
	res := result.Ok(5)
	mappedRes := result.Map(res, func(x int) string {
		return "value: " + string(rune(x+'0'))
	})
	if mappedRes.UnwrapOr("error") != "value: 5" {
		t.Errorf("Expected 'value: 5', got '%v'", mappedRes.UnwrapOr("error"))
	}

	errRes := result.Err[int](errors.New("map fail"))
	mappedErrRes := result.Map(errRes, func(x int) string {
		return "should not execute"
	})
	if mappedErrRes.IsOk() {
		t.Errorf("Expected error, but got Ok result")
	}
}

// TestFlatMap ensures FlatMap correctly transforms and chains Result computations.
func TestFlatMap(t *testing.T) {
	res := result.Ok(10)
	flatMappedRes := result.FlatMap(res, func(x int) result.Result[string] {
		return result.Ok("Number: " + fmt.Sprintf("%d", x))
	})
	if flatMappedRes.UnwrapOr("error") != "Number: 10" {
		t.Errorf("Expected 'Number: 10', got '%v'", flatMappedRes.UnwrapOr("error"))
	}

	errRes := result.Err[int](errors.New("flatmap fail"))
	flatMappedErrRes := result.FlatMap(errRes, func(x int) result.Result[string] {
		return result.Ok("should not execute")
	})
	if flatMappedErrRes.IsOk() {
		t.Errorf("Expected error, but got Ok result")
	}
}

// TestAndThen ensures AndThen executes a function that may fail.
func TestAndThen(t *testing.T) {
	res := result.Ok(42)
	newRes := res.AndThen(func(x int) error {
		if x > 40 {
			return errors.New("too large")
		}
		return nil
	})
	if newRes.IsOk() {
		t.Errorf("Expected error in AndThen, but got Ok result")
	}

	successRes := result.Ok(30)
	successRes = successRes.AndThen(func(x int) error {
		return nil
	})
	if successRes.IsErr() {
		t.Errorf("Expected Ok result in AndThen, but got Err")
	}
}

// TestOrElse ensures OrElse provides an alternative Result when there is an error.
func TestOrElse(t *testing.T) {
	errRes := result.Err[int](errors.New("original error"))
	fallbackRes := errRes.OrElse(func() result.Result[int] {
		return result.Ok(99)
	})
	if fallbackRes.UnwrapOr(0) != 99 {
		t.Errorf("Expected fallback value 99, got %v", fallbackRes.UnwrapOr(0))
	}

	okRes := result.Ok(123)
	noFallbackRes := okRes.OrElse(func() result.Result[int] {
		return result.Ok(99) // Should not execute
	})
	if noFallbackRes.UnwrapOr(0) != 123 {
		t.Errorf("Expected original Ok value 123, got %v", noFallbackRes.UnwrapOr(0))
	}
}
