package godiff

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
	n, ok := v.(Nullable)
	return n, ok
}

// nullableData attempts to extract the underlying value from a Nullable.
// It tries NullableWithData[any] first, then falls back to reflection-free
// approach. Returns the value and true if extraction succeeded.
func nullableData(v any) (any, bool) {
	if n, ok := v.(NullableWithData[any]); ok {
		return n.Data(), true
	}
	return nil, false
}
