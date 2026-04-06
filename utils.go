package godiff

import (
	"errors"
	"reflect"
	"strings"
	"sync"
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

var (
	// structTag holds the active struct tag key. Defaults to DefaultStructTag.
	structTag   = DefaultStructTag
	structTagMu sync.RWMutex
)

// SetStructTag overrides the struct tag key used to read audit directives
// (ignore, required, immutable, redact, reduce_*) from struct fields.
// Call this once at program startup before any comparison calls.
// It is safe for concurrent use.
func SetStructTag(tag string) {
	structTagMu.Lock()
	defer structTagMu.Unlock()
	structTag = tag
}

// GetStructTag returns the currently active struct tag key.
// It is safe for concurrent use.
func GetStructTag() string {
	structTagMu.RLock()
	defer structTagMu.RUnlock()
	return structTag
}

// getStructTag is the internal non-exported accessor used by hot paths
// (struct.go, types.go) to avoid repeated function-call overhead while
// still being race-safe.
func getStructTag() string {
	structTagMu.RLock()
	defer structTagMu.RUnlock()
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
	if v == nil {
		return true
	}

	rv := reflect.ValueOf(v)
	if !rv.IsValid() {
		return true
	}

	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return true
		}
		rv = rv.Elem()
	}

	switch rv.Kind() {
	case reflect.Float64, reflect.Float32:
		return rv.Float() == 0
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return rv.Uint() == 0
	case reflect.Bool:
		return !rv.Bool()
	case reflect.String:
		return rv.Len() == 0
	case reflect.Slice, reflect.Array:
		return rv.Len() == 0
	case reflect.Map:
		return rv.Len() == 0
	case reflect.Struct:
		zero := reflect.Zero(rv.Type())
		return reflect.DeepEqual(rv.Interface(), zero.Interface())
	default:
		return false
	}
}

// IsEmptySlice checks if a value is an empty slice.
// Returns false for nil values (nil is not an empty slice).
func IsEmptySlice(v any) bool {
	if v == nil {
		return false
	}
	r := reflect.ValueOf(v)
	return r.IsValid() && r.Kind() == reflect.Slice && r.Len() == 0
}

// IsEmptyMap checks if a value is an empty map.
// Returns false for nil values (nil is not an empty map).
func IsEmptyMap(v any) bool {
	if v == nil {
		return false
	}
	r := reflect.ValueOf(v)
	return r.IsValid() && r.Kind() == reflect.Map && r.Len() == 0
}

// IsMap checks if the provided value is of the reflect.Map kind and returns
// true if it is, otherwise false.
func IsMap(v any) bool {
	if v == nil {
		return false
	}
	r := reflect.ValueOf(v)
	return r.IsValid() && r.Kind() == reflect.Map
}

// createPath constructs a hierarchical path by combining a parent path and a
// key using a forward slash separator.
func createPath(parent, key string) string {
	parent = strings.TrimSuffix(parent, PathSep)
	key = strings.TrimPrefix(key, PathSep)
	return parent + PathSep + key
}
