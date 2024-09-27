package composer

import (
	"errors"
	"log"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	t.Parallel()

	// Act
	composer := New[int, string]()

	// Assert
	assert.NotNil(t, composer)
}

func TestNew_WithDependencies(t *testing.T) {
	t.Parallel()
	// Arrange
	c := New[int, string]()

	// Act
	composer := New(WithFn(c, "a", double),
		WithFn(c, "b", double))

	// Assert
	assert.Len(t, composer.tasks, 2)
}

func TestComposer_Compose_ReturnsCorrectResult(t *testing.T) {
	t.Parallel()

	type testCase struct {
		c             *Composer[int, string]
		shouldCompose bool
	}

	// Arrange
	c := New[int, string]()

	testCases := map[string]testCase{
		"nil composer": {
			c: nil,
		},
		"empty composer": {
			c: New[int, string](),
		},
		"1 func, correct types": {
			c:             New(WithFn(c, "toString", toString)),
			shouldCompose: true,
		},
		"1 func, incorrect types": {
			c: New(WithFn(c, "addOne", addOne)),
		},
		"multiple funcs, correct types": {
			c: New(WithFn(c, "addOne", addOne),
				WithFn(c, "toString", toString)),
			shouldCompose: true,
		},
		"multiple funcs, incorrect types": {
			c: New(WithFn(c, "addOne", addOne),
				WithFn(c, "double", double)),
		},
		"multiple funcs, types not alligned": {
			c: New(WithFn(c, "toString", toString),
				WithFn(c, "double", double)),
		},
	}

	for name, test := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Act
			fn, err := test.c.Compose()

			// Assert
			if test.shouldCompose {
				assert.NoError(t, err)
				assert.NotNil(t, fn)
			} else {
				assert.Error(t, err)
				assert.Nil(t, fn)
			}
		})
	}
}

func TestComposer_MustCompose_PanicsOnError(t *testing.T) {
	t.Parallel()

	type testCase struct {
		c           *Composer[int, string]
		shouldPanic bool
	}

	// Arrange
	c := New[int, string]()

	testCases := map[string]testCase{
		"nil composer": {
			c:           nil,
			shouldPanic: true,
		},
		"empty composer": {
			c:           New[int, string](),
			shouldPanic: true,
		},
		"1 func, correct types": {
			c:           New(WithFn(c, "toString", toString)),
			shouldPanic: false,
		},
		"1 func, incorrect types": {
			c:           New(WithFn(c, "addOne", addOne)),
			shouldPanic: true,
		},
	}

	for name, test := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Act & Assert
			if test.shouldPanic {
				assert.Panics(t, func() {
					test.c.MustCompose()
				})
			} else {
				assert.NotPanics(t, func() {
					test.c.MustCompose()
				})
			}
		})
	}
}

func TestComposer_ComposeNoErr_ReturnsCorrectResult(t *testing.T) {
	t.Parallel()

	type testCase struct {
		c             *Composer[int, string]
		shouldCompose bool
	}

	// Arrange
	c := New[int, string]()

	testCases := map[string]testCase{
		"nil composer": {
			c: nil,
		},
		"empty composer": {
			c: New[int, string](),
		},
		"1 func, correct types": {
			c:             New(WithFn(c, "toString", toString)),
			shouldCompose: true,
		},
		"1 func, incorrect types": {
			c: New(WithFn(c, "addOne", addOne)),
		},
		"multiple funcs, correct types": {
			c: New(WithErrFn(c, "errFunc", errFunc[int]),
				WithFn(c, "toString", toString)),
			shouldCompose: true,
		},
	}

	for name, test := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Act
			fn, err := test.c.ComposeNoErr(false)

			// Assert
			if test.shouldCompose {
				assert.NoError(t, err)
				assert.NotNil(t, fn)
			} else {
				assert.Error(t, err)
				assert.Nil(t, fn)
			}
		})
	}
}

func TestComposer_MustComposeNoErr_PanicsOnError(t *testing.T) {
	t.Parallel()

	type testCase struct {
		c            *Composer[int, string]
		panicOnError bool
	}

	// Arrange
	c := New[int, string]()

	testCases := map[string]testCase{
		"nil composer": {
			c:            nil,
			panicOnError: true,
		},
		"empty composer": {
			c:            New[int, string](),
			panicOnError: true,
		},
		"1 func, correct types": {
			c:            New(WithFn(c, "toString", toString)),
			panicOnError: false,
		},
		"1 func, incorrect types": {
			c:            New(WithErrFn(c, "errFunc", errFunc[int])),
			panicOnError: true,
		},
	}

	for name, test := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Act & Assert
			if test.panicOnError {
				assert.Panics(t, func() {
					test.c.MustComposeNoErr(test.panicOnError)
				})
			} else {
				assert.NotPanics(t, func() {
					test.c.MustComposeNoErr(test.panicOnError)
				})
			}
		})
	}
}

func addOne(n int) int      { return n + 1 }
func double(n int) int      { return int(n) * 2 }
func toString(n int) string { return strconv.Itoa(n) }
func errFunc[T any](n T) (T, error) {
	return n, errors.New("an error occurred")
}

func logVal[T any](f func(T) T) func(T) T {
	return func(t T) T {
		log.Println("value before: ", t)
		defer func() { log.Println("value after: ", t) }()
		t = f(t)
		return t
	}
}
