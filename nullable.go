package godiff

import "reflect"

// Nullable defines the minimal interface for nullable/optional value types.
// Any type satisfying this interface will be treated with nullable semantics
// during comparison: two values where both IsNil() return true are considered
// equal, regardless of their underlying zero values.
//
// This is intentionally non-generic so that generic types like nilable.Value[T]
// can satisfy it without interface matching issues.
//
// Example:
//
//	func (v Value[T]) IsNil() bool { return !v.isSet }
type Nullable interface {
	// IsNil reports whether the value is in its unset/nil state.
	IsNil() bool
}

// NullableWithData extends Nullable with typed data access. Use this when
// you need to unwrap the underlying value (e.g. for time comparison).
//
// Example:
//
//	func (v Value[T]) IsNil() bool { return !v.isSet }
//	func (v Value[T]) Data() T     { return v.data }
type NullableWithData[T any] interface {
	Nullable

	// Data returns the underlying value. The caller should check IsNil
	// before using the returned value.
	Data() T
}

// isNullable checks if the value implements the Nullable interface.
func isNullable(v any) (Nullable, bool) {
	if v == nil {
		return nil, false
	}
	n, ok := v.(Nullable)
	return n, ok
}

// nullableData attempts to extract the underlying value from a value that has
// a Data() method returning any type. Because Go generics are invariant,
// NullableWithData[time.Time] does not satisfy NullableWithData[any], so we
// use reflection to call the Data method dynamically.
func nullableData(v any) (any, bool) {
	if v == nil {
		return nil, false
	}
	rv := reflect.ValueOf(v)
	method := rv.MethodByName("Data")
	if !method.IsValid() {
		return nil, false
	}
	// Data() must take zero args and return exactly one value
	mt := method.Type()
	if mt.NumIn() != 0 || mt.NumOut() != 1 {
		return nil, false
	}
	results := method.Call(nil)
	return results[0].Interface(), true
}
