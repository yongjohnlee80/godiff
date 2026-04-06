package godiff

import (
	"testing"
	"time"
)

// testNullable is a minimal Nullable implementation for testing.
type testNullable[T any] struct {
	data  T
	isSet bool
}

func newNullable[T any]() *testNullable[T] {
	return &testNullable[T]{}
}

func newNullableWith[T any](v T) *testNullable[T] {
	return &testNullable[T]{data: v, isSet: true}
}

func (n *testNullable[T]) IsNil() bool { return !n.isSet }
func (n *testNullable[T]) Data() T     { return n.data }

func TestNullableInterface(t *testing.T) {
	t.Run("both nil nullable are equal", func(t *testing.T) {
		a := newNullable[string]()
		b := newNullable[string]()
		if !DefaultCompareFunc("testNullable", a, b) {
			t.Error("expected two nil nullables to be equal")
		}
	})

	t.Run("nil nullable and plain nil are equal", func(t *testing.T) {
		a := newNullable[int]()
		if !DefaultCompareFunc("testNullable", a, nil) {
			t.Error("expected nil nullable and nil to be equal")
		}
	})

	t.Run("set nullable vs nil nullable are not equal", func(t *testing.T) {
		a := newNullableWith("hello")
		b := newNullable[string]()
		if DefaultCompareFunc("testNullable", a, b) {
			t.Error("expected set nullable and nil nullable to not be equal")
		}
	})

	t.Run("nullable time both nil", func(t *testing.T) {
		a := newNullable[time.Time]()
		b := newNullable[time.Time]()
		if !DefaultCompareFunc("time.Time", a, b) {
			t.Error("expected two nil nullable times to be equal")
		}
	})

	t.Run("nullable time set vs nil", func(t *testing.T) {
		now := time.Now()
		a := newNullableWith(now)
		b := newNullable[time.Time]()
		if DefaultCompareFunc("time.Time", a, b) {
			t.Error("expected set nullable time and nil nullable time to differ")
		}
	})
}

func TestNullableSatisfiesInterface(t *testing.T) {
	// Compile-time: testNullable satisfies Nullable (non-generic)
	var _ Nullable = newNullable[string]()
	var _ Nullable = newNullableWith(42)
	var _ Nullable = newNullable[time.Time]()

	// Compile-time: testNullable satisfies NullableWithData[T]
	var _ NullableWithData[string] = newNullable[string]()
	var _ NullableWithData[int] = newNullableWith(42)
	var _ NullableWithData[time.Time] = newNullable[time.Time]()
}
