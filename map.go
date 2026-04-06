package godiff

import (
	"encoding/json"
	"fmt"
)

// DataMap is a type alias for a map with string keys and values of any type,
// enabling flexible data storage or manipulation, as well as auditing
type DataMap map[string]any

// toDataMap converts a struct or any compatible type into a DataMap and returns it.
func toDataMap(data any) (DataMap, error) {
	switch v := data.(type) {
	case DataMap:
		return v, nil
	case map[string]any:
		return v, nil
	}

	// Marshal the struct to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("error marshaling struct: %w", err)
	}

	// Unmarshal JSON into a map
	result := make(DataMap)
	err = json.Unmarshal(jsonData, &result)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling JSON to map: %w", err)
	}

	return result, nil
}

// NewDataMap converts a struct or any compatible type into a DataMap and
// returns it along with any encountered error.
func NewDataMap(data any, opts ...EntityOption) (DataMap, error) {
	result, err := toDataMap(data)
	if err != nil {
		return nil, err
	}

	err = result.Options(opts...)
	if err != nil {
		return nil, fmt.Errorf("error applying options: %w", err)
	}

	return result, nil
}

// Set updates the DataMap by assigning the given value to the specified path.
func (m *DataMap) Set(path string, value any) {
	(*m)[path] = value
}

// Get retrieves the value associated with the specified path from the DataMap.
func (m *DataMap) Get(path string) (any, bool) {
	v, ok := (*m)[path]
	return v, ok
}

// GetStringValue retrieves the value at the specified path as a string.
// Returns an empty string if the path is not found.
func (m *DataMap) GetStringValue(path string) string {
	v, ok := (*m)[path]
	if !ok {
		return ""
	}
	return fmt.Sprintf("%v", v)
}

// Bytes convert the DataMap into a JSON-encoded byte slice.
// It falls back to a formatted string representation if marshaling fails.
func (m *DataMap) Bytes() []byte {
	data, err := json.Marshal(m)
	if err != nil {
		str := fmt.Sprintf("%+v", m)
		return []byte(str)
	}
	return data
}

// Options apply a series of EntityOption operations to the DataMap and
// return an error if any operation fails.
func (m *DataMap) Options(ops ...EntityOption) error {
	for _, op := range ops {
		err := op(m, nil)
		if err != nil {
			return err
		}
	}
	return nil
}

// RemoveByPath removes one or more fields from the DataMap based on the
// specified hierarchical paths.
func (m *DataMap) RemoveByPath(paths ...string) {
	for _, path := range paths {
		RemoveFieldByPath(m, path)
	}
}

// Cast converts a DataMap into a value of the specified generic type T by
// encoding it as JSON and then decoding it.
// Returns the converted value of type T and an error if the conversion fails.
func Cast[T any](v DataMap) (T, error) {
	return CastBytes[T](v.Bytes())
}

// MustCast converts a DataMap into a specified type by unmarshalling its JSON
// representation into the target type. Panics if conversion fails.
func MustCast[T any](v DataMap) T {
	return MustCastBytes[T](v.Bytes())
}

// CastBytes decodes a JSON-encoded byte slice into a value of the specified
// generic type T.
// Returns the value of type T and an error if the decoding fails.
func CastBytes[T any](b []byte) (T, error) {
	var t T
	err := json.Unmarshal(b, &t)
	return t, err
}

// MustCastBytes decodes a JSON-encoded byte slice into a value of the specified
// generic type T. Panics if decoding fails.
func MustCastBytes[T any](b []byte) T {
	t, err := CastBytes[T](b)
	if err != nil {
		panic("godiff: MustCastBytes failed: " + err.Error())
	}
	return t
}
