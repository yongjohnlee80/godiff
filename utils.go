package godiff

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

const (
	// AuditLogIgnore defines a constant used to mark struct fields that should
	// be excluded from audit logging.
	AuditLogIgnore = "ignore"

	// DefaultStructTag is the default struct tag key used to read audit
	// directives from struct fields.
	DefaultStructTag = "audit_log"

	// PathSep defines the delimiter used to separate elements in a hierarchical
	// path in Entity objects
	PathSep = "/"
)

// structTag holds the active struct tag key. Defaults to DefaultStructTag.
var structTag = DefaultStructTag

// SetStructTag overrides the struct tag key used to read audit directives
// (ignore, required, immutable, redact, reduce_*) from struct fields.
// Call this once at program startup before any comparison calls.
func SetStructTag(tag string) {
	structTag = tag
}

// GetStructTag returns the currently active struct tag key.
func GetStructTag() string {
	return structTag
}

var (
	// ErrNotStruct is returned when the provided value is not of kind struct
	// during reflection operations.
	ErrNotStruct = errors.New("not a struct")
)

// IsEmpty checks if the provided value is "empty" by comparing it to its type's
// zero value or nil for pointers.
func IsEmpty(v any) bool {
	rv := reflect.ValueOf(v)

	if rv.Kind() == reflect.Ptr {
		// If it's a pointer, dereference it if non-nil
		if rv.IsNil() {
			return true
		}
		rv = rv.Elem()
	}

	switch rv.Kind() {
	case reflect.Float64, reflect.Float32:
		return rv.Float() == 0
	case reflect.String, reflect.Slice, reflect.Array:
		return rv.Len() == 0
	default:
		// checking for untyped nil
		if !rv.IsValid() {
			return true
		}
	}

	// Compare against the zero value of the type
	zero := reflect.Zero(rv.Type())
	return reflect.DeepEqual(rv.Interface(), zero.Interface())
}

// IsEmptySlice checks if a value is an empty slice.
func IsEmptySlice(v any) bool {
	r := reflect.ValueOf(v)
	return r.Kind() == reflect.Slice && r.Len() == 0
}

// IsEmptyMap checks if a value is an empty map.
func IsEmptyMap(v any) bool {
	r := reflect.ValueOf(v)
	return r.Kind() == reflect.Map && len(r.MapKeys()) == 0
}

// IsMap checks if the provided value is of the reflect.Map kind and returns
// true if it is, otherwise false.
func IsMap(v any) bool {
	r := reflect.ValueOf(v)
	return r.Kind() == reflect.Map
}

// createPath constructs a hierarchical path by combining a parent path and a
// key using a forward slash separator.
func createPath(parent, key string) string {
	parent = strings.TrimSuffix(parent, PathSep)
	key = strings.TrimPrefix(key, PathSep)
	return fmt.Sprintf("%s%s%s", parent, PathSep, key)
}
