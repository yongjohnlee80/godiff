package godiff

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"
)

// Diff represents the differences between two data states, capturing old and
// new values as well as their metadata.
type Diff struct {
	Key  string
	New  interface{}
	Old  interface{}
	Path string
	Type string
}

// NewDiff creates a new Diff object representing a change with specified
// metadata, old and new values, and hierarchical path.
func NewDiff(key, typ string, old, new interface{}, path string) *Diff {
	return &Diff{
		Key:  key,
		Type: typ,
		New:  new,
		Old:  old,
		Path: createPath(path, key),
	}
}

// OldBytes returns the serialized JSON representation of the Old field.
// Falls back to string conversion if serialization fails.
func (d *Diff) OldBytes() []byte {
	data, err := json.Marshal(d.Old)
	if err != nil {
		str := fmt.Sprintf("%+v", d.Old)
		return []byte(str)
	}
	return data
}

// NewBytes serializes the New field into a JSON byte slice or returns the
// string representation if serialization fails.
func (d *Diff) NewBytes() []byte {
	data, err := json.Marshal(d.New)
	if err != nil {
		str := fmt.Sprintf("%+v", d.New)
		return []byte(str)
	}
	return data
}

// String returns a string representation of the Diff, showing its path, type,
// and old and new values.
func (d *Diff) String() string {
	return fmt.Sprintf("%s(%s), %s -> %s", d.Path, d.Type, d.OldBytes(), d.NewBytes())
}

// Compare analyzes the differences between two data structures 'old' and 'new'
// and returns changed keys, differences, and errors. It converts the inputs
// into DataMaps for comparison, detecting additions, deletions, and changes in
// fields.
// It uses fields from the new DataMap to compare against the old DataMap.
// The values from the old DataMap are not compared against the values of the
// new DataMap by design to accommodate partial updates to the entity.
func Compare(old, new any, opts ...CompareOption) ([]string, CompareResult, error) {
	oldMap, err := NewDataMap(old)
	if err != nil {
		return nil, nil, err
	}

	types, err := NewDataTypes(new)
	if err != nil {
		types = make(DataTypes)
	}
	types.Add(".", reflect.TypeOf(new).String())

	newMap, err := NewDataMap(new)
	if err != nil {
		return nil, nil, err
	}

	updatedKeys, context, err := CompareMaps(oldMap, newMap, types, opts...)
	if err != nil {
		return nil, context, err
	}

	return updatedKeys, context, nil
}

// CompareMaps analyzes the differences between two DataMaps, detecting changes,
// additions, and deletions across keys. It returns a list of changed keys,
// a slice of Diff objects representing the changes, and an optional error.
func CompareMaps(
	oldMap, newMap DataMap,
	types DataTypes,
	opts ...CompareOption,
) ([]string, CompareContext, error) {
	context := NewContextAuditCompare(&oldMap, &newMap, &types)

	// apply pre-audit stage options
	for _, op := range opts {
		if err := op(context); err != nil {
			return nil, context, err
		}
	}

	keys, diffs := generateDiff("", oldMap, newMap, types)
	context.AddDiff(diffs...)

	// apply post-audit stage options
	context.SetStage(PostAuditStage)
	for _, op := range opts {
		if err := op(context); err != nil {
			return nil, context, err
		}
	}

	return keys, context, nil
}

// generateDiff compares any changes from the old DataMap with the new DataMap.
// It uses fields from the new DataMap to compare against the old DataMap.
// The values from the old DataMap are not compared against the values of the
// new DataMap by design to accommodate partial updates to the entity.
func generateDiff(
	parentPath string,
	old, new DataMap,
	types DataTypes,
) ([]string, []*Diff) {
	var diffs []*Diff
	var keys []string

	compare := func(key string, a, b any) {
		r := reflect.ValueOf(b)
		switch r.Kind() {
		case reflect.Slice, reflect.Array:
			// Treat nil and empty slices as equivalent
			if (a == nil || IsEmptySlice(a)) && (b == nil || IsEmptySlice(b)) {
				return // Both are nil or empty, so they are considered the same
			}
			if !reflect.DeepEqual(a, b) {
				typ := types.Get(createPath(parentPath, key))
				if typ == "" {
					typ = r.Kind().String()
				}
				diffs = append(diffs, NewDiff(key, typ, a, b, parentPath))
				keys = append(keys, key)
			}
		case reflect.Map, reflect.Struct:
			aMap, _ := NewDataMap(a)
			bMap, _ := NewDataMap(b)
			// Treat nil and empty maps as equivalent
			if (a == nil || IsEmptyMap(aMap)) && (b == nil || IsEmptyMap(bMap)) {
				return // Both are nil or empty, so they are considered the same
			}

			_key, _diffs := generateDiff(createPath(parentPath, key), aMap, bMap, types)
			keys = append(keys, _key...)
			diffs = append(diffs, _diffs...)
		default:
			typ := types.Get(createPath(parentPath, key))
			if typ == "" {
				typ = r.Kind().String()
			}
			if !DefaultCompareFunc(typ, a, b) {
				diffs = append(diffs, NewDiff(key, typ, a, b, parentPath))
				keys = append(keys, key)
			}
		}
	}

	for k, v := range new {
		compare(k, old[k], v)
	}
	return keys, diffs
}

// DefaultCompareFunc compares two values of the specified type and returns true
// if they are considered equivalent.
//
// Nullable handling: if either value implements the Nullable interface, both
// values being nil/unset are treated as equal. This also applies to types whose
// type string starts with "nilable.Value" (legacy string-based detection).
//
// For time types, it checks if the values are within 1 minute of each other.
// For scalar types, it performs direct equality comparison.
func DefaultCompareFunc(typ string, a, b any) bool {
	isEmptyOrNil := func(v any) bool {
		if v == nil {
			return true
		}
		if n, ok := isNullable(v); ok {
			return n.IsNil()
		}
		return IsEmpty(v)
	}

	parseTime := func(v any) (time.Time, error) {
		// Unwrap Nullable to get the underlying value
		if n, ok := isNullable(v); ok {
			if n.IsNil() {
				return time.Time{}, nil
			}
			if data, ok := nullableData(v); ok {
				v = data
			}
		}
		switch v := v.(type) {
		case string:
			t, err := time.Parse(time.RFC3339, v)
			if err != nil {
				return time.Time{}, err
			}
			return t, nil
		case time.Time:
			return v, nil
		default:
			return time.Time{}, fmt.Errorf("unsupported time format")
		}
	}

	// Check Nullable interface on either side
	aNullable, aIsNullable := isNullable(a)
	bNullable, bIsNullable := isNullable(b)
	if aIsNullable || bIsNullable {
		aNil := (a == nil) || (aIsNullable && aNullable.IsNil())
		bNil := (b == nil) || (bIsNullable && bNullable.IsNil())
		if aNil && bNil {
			return true
		}
	}

	// Legacy string-based detection for types serialized as "nilable.Value[T]"
	if strings.HasPrefix(typ, "nilable.Value") {
		if isEmptyOrNil(a) && isEmptyOrNil(b) {
			return true
		}
	}

	switch typ {
	case "time.Time", "nilable.Value[time.Time]":
		t1, err1 := parseTime(a)
		t2, err2 := parseTime(b)

		if err1 != nil || err2 != nil {
			return false
		}

		diff := t1.Sub(t2)
		if diff < time.Minute && diff > -time.Minute {
			return true
		}
		return false
	}

	// for scalars (int, string, etc), just compare values
	return a == b
}
