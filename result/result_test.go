package result_test

import (
	"errors"
	"fmt"
	"strings"
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

func TestMatch(t *testing.T) {
	tests := []struct {
		name    string
		r       result.Result[string]
		wantOk  string
		wantErr string
		matched bool
	}{
		{
			name:    "matches success case",
			r:       result.Ok("success"),
			wantOk:  "SUCCESS",
			matched: true,
		},
		{
			name:    "matches error case",
			r:       result.Err[string](errors.New("failed")),
			wantErr: "FAILED",
			matched: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var matched bool
			tt.r.Match(
				func(v string) {
					matched = true
					if strings.ToUpper(v) != tt.wantOk {
						t.Errorf("Match ok got = %v, want %v", v, tt.wantOk)
					}
				},
				func(err error) {
					matched = true
					if strings.ToUpper(err.Error()) != tt.wantErr {
						t.Errorf("Match err got = %v, want %v", err, tt.wantErr)
					}
				},
			)
			if matched != tt.matched {
				t.Errorf("Match called = %v, want %v", matched, tt.matched)
			}
		})
	}
}

func TestFilter(t *testing.T) {
	tests := []struct {
		name      string
		r         result.Result[int]
		predicate func(int) bool
		want      bool
	}{
		{
			name:      "passes filter",
			r:         result.Ok(42),
			predicate: func(i int) bool { return i > 40 },
			want:      true,
		},
		{
			name:      "fails filter",
			r:         result.Ok(38),
			predicate: func(i int) bool { return i > 40 },
			want:      false,
		},
		{
			name:      "error stays error",
			r:         result.Err[int](errors.New("failed")),
			predicate: func(i int) bool { return true },
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filtered := tt.r.Filter(tt.predicate)
			if got := filtered.IsOk(); got != tt.want {
				t.Errorf("Filter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInspect(t *testing.T) {
	var inspected int
	r := result.Ok(42)

	// Test Inspect
	res := r.Inspect(func(v int) {
		inspected = v
	})

	if inspected != 42 {
		t.Errorf("Inspect() didn't call function with correct value, got %v", inspected)
	}
	if !res.IsOk() {
		t.Error("Inspect() changed Result status")
	}

	// Test InspectErr
	var inspectedErr error
	errResult := result.Err[int](errors.New("test error"))

	res = errResult.InspectErr(func(err error) {
		inspectedErr = err
	})

	if inspectedErr == nil || inspectedErr.Error() != "test error" {
		t.Errorf("InspectErr() didn't call function with correct error")
	}
	if !res.IsErr() {
		t.Error("InspectErr() changed Result status")
	}
}

func TestCollect(t *testing.T) {
	tests := []struct {
		name    string
		results []result.Result[int]
		want    []int
		wantErr bool
	}{
		{
			name: "all success",
			results: []result.Result[int]{
				result.Ok(1),
				result.Ok(2),
				result.Ok(3),
			},
			want:    []int{1, 2, 3},
			wantErr: false,
		},
		{
			name: "contains error",
			results: []result.Result[int]{
				result.Ok(1),
				result.Err[int](errors.New("failed")),
				result.Ok(3),
			},
			wantErr: true,
		},
		{
			name:    "empty slice",
			results: []result.Result[int]{},
			want:    []int{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collected := result.Collect(tt.results)
			if tt.wantErr {
				if collected.IsOk() {
					t.Error("Collect() expected error, got success")
				}
				return
			}

			if collected.IsErr() {
				t.Errorf("Collect() unexpected error: %v", collected.UnwrapErr())
				return
			}

			got, _ := collected.Unwrap()
			if len(got) != len(tt.want) {
				t.Errorf("Collect() got len = %v, want %v", len(got), len(tt.want))
				return
			}

			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("Collect() got[%d] = %v, want %v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestZip(t *testing.T) {
	tests := []struct {
		name    string
		r1      result.Result[int]
		r2      result.Result[string]
		want    result.Pair[int, string]
		wantErr bool
	}{
		{
			name:    "both success",
			r1:      result.Ok(42),
			r2:      result.Ok("success"),
			want:    result.Pair[int, string]{First: 42, Second: "success"},
			wantErr: false,
		},
		{
			name:    "first error",
			r1:      result.Err[int](errors.New("first failed")),
			r2:      result.Ok("success"),
			wantErr: true,
		},
		{
			name:    "second error",
			r1:      result.Ok(42),
			r2:      result.Err[string](errors.New("second failed")),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zipped := result.Zip(tt.r1, tt.r2)
			if tt.wantErr {
				if zipped.IsOk() {
					t.Error("Zip() expected error, got success")
				}
				return
			}

			if zipped.IsErr() {
				t.Errorf("Zip() unexpected error: %v", zipped.UnwrapErr())
				return
			}

			got, _ := zipped.Unwrap()
			if got.First != tt.want.First || got.Second != tt.want.Second {
				t.Errorf("Zip() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTry(t *testing.T) {
	tests := []struct {
		name    string
		fn      func() string
		want    string
		wantErr string
	}{
		{
			name: "success",
			fn: func() string {
				return "success"
			},
			want: "success",
		},
		{
			name: "panic with error",
			fn: func() string {
				panic(errors.New("planned error"))
			},
			wantErr: "planned error",
		},
		{
			name: "panic with string",
			fn: func() string {
				panic("something went wrong")
			},
			wantErr: "something went wrong",
		},
		{
			name: "panic with other type",
			fn: func() string {
				panic(42)
			},
			wantErr: "panic: 42",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := result.Try(tt.fn)

			if tt.wantErr != "" {
				if r.IsOk() {
					t.Error("Try() expected error, got success")
					return
				}
				if err := r.UnwrapErr(); err.Error() != tt.wantErr {
					t.Errorf("Try() error = %v, want %v", err, tt.wantErr)
				}
				return
			}

			if r.IsErr() {
				t.Errorf("Try() unexpected error: %v", r.UnwrapErr())
				return
			}

			got, _ := r.Unwrap()
			if got != tt.want {
				t.Errorf("Try() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOption(t *testing.T) {
	t.Run("Some", func(t *testing.T) {
		opt := result.Some(42)

		if opt.IsNone() {
			t.Error("Some value reported as None")
		}
		if !opt.IsSome() {
			t.Error("Some value not reported as Some")
		}
		if v := opt.UnwrapOr(0); v != 42 {
			t.Errorf("UnwrapOr returned %v, want 42", v)
		}
		if v := opt.UnwrapOrPanic(); v != 42 {
			t.Errorf("UnwrapOrPanic returned %v, want 42", v)
		}
	})

	t.Run("None", func(t *testing.T) {
		opt := result.None[int]()

		if !opt.IsNone() {
			t.Error("None value not reported as None")
		}
		if opt.IsSome() {
			t.Error("None value reported as Some")
		}
		if v := opt.UnwrapOr(42); v != 42 {
			t.Errorf("UnwrapOr returned %v, want 42", v)
		}
	})

	t.Run("None_UnwrapOrPanic", func(t *testing.T) {
		opt := result.None[int]()

		defer func() {
			if r := recover(); r == nil {
				t.Error("UnwrapOrPanic did not panic on None")
			}
		}()

		_ = opt.UnwrapOrPanic()
	})
}

func TestTranspose(t *testing.T) {
	t.Run("Some_Ok", func(t *testing.T) {
		r := result.Ok(result.Some(42))
		opt := result.Transpose(r)

		if opt.IsNone() {
			t.Fatal("Expected Some, got None")
		}

		res := opt.UnwrapOrPanic()
		if res.IsErr() {
			t.Fatal("Expected Ok, got Err")
		}

		val, _ := res.Unwrap()
		if val != 42 {
			t.Errorf("Got %v, want 42", val)
		}
	})

	t.Run("Some_Err", func(t *testing.T) {
		err := errors.New("test error")
		r := result.Err[result.Option[int]](err)
		opt := result.Transpose(r)

		if opt.IsNone() {
			t.Fatal("Expected Some, got None")
		}

		res := opt.UnwrapOrPanic()
		if !res.IsErr() {
			t.Fatal("Expected Err, got Ok")
		}

		if e := res.UnwrapErr(); e != err {
			t.Errorf("Got error %v, want %v", e, err)
		}
	})

	t.Run("None", func(t *testing.T) {
		r := result.Ok(result.None[int]())
		opt := result.Transpose(r)

		if !opt.IsNone() {
			t.Error("Expected None, got Some")
		}
	})
}

func TestFromOption(t *testing.T) {
	err := errors.New("none error")

	t.Run("Some", func(t *testing.T) {
		opt := result.Some(42)
		r := result.FromOption(opt, err)

		if r.IsErr() {
			t.Fatal("Expected Ok, got Err")
		}

		val, _ := r.Unwrap()
		if val != 42 {
			t.Errorf("Got %v, want 42", val)
		}
	})

	t.Run("None", func(t *testing.T) {
		opt := result.None[int]()
		r := result.FromOption(opt, err)

		if !r.IsErr() {
			t.Fatal("Expected Err, got Ok")
		}

		if e := r.UnwrapErr(); e != err {
			t.Errorf("Got error %v, want %v", e, err)
		}
	})
}
