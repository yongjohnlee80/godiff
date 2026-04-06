package godiff

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrEmptyDataMap = errors.New("cannot perform action on empty map")
	ErrEmptyPath    = errors.New("cannot perform action with empty path")
	ErrInvalidPath  = errors.New("invalid path")
)

// Undo reverts changes in the DataMap using the provided Diff list,
// restoring the old values at specified paths.
func (m *DataMap) Undo(diff ...*Diff) error {
	for _, d := range diff {
		err := UpdateMapByPath(m, d.Path, d.Old)
		if err != nil {
			return err
		}
	}
	return nil
}

// Redo re-applies the changes in the DataMap using the provided Diff list,
// updating the map with new values at specified paths.
func (m *DataMap) Redo(diff ...*Diff) error {
	for _, d := range diff {
		err := UpdateMapByPath(m, d.Path, d.New)
		if err != nil {
			return err
		}
	}
	return nil
}

// UpdateMapByPath modifies a DataMap by updating a specific path with the
// provided value or creating nested maps as needed. m is the map to update,
// path is the "/"-separated key, and value is the new value to set at the
// specified path. Returns an error if the path conflicts with a non-map value.
func UpdateMapByPath(m *DataMap, path string, value any) error {
	if *m == nil {
		*m = make(DataMap)
	}

	// if path is empty, directly update the root of the map
	if path == "" || path == PathSep {
		// Explicitly delete the map by deleting all keys.
		for k := range *m {
			delete(*m, k)
		}

		// create a new map with the value.
		newMap, err := NewDataMap(value)
		if err != nil {
			return fmt.Errorf("failed to create new map from value: %w", err)
		}
		*m = newMap
		return nil
	}

	// Split the path into parts
	parts := strings.Split(strings.TrimPrefix(path, PathSep), PathSep)

	// Iterate to navigate the map
	current := *m
	for i, key := range parts {
		// If we’re at the last part, update the value
		if i == len(parts)-1 {
			current[key] = value
			return nil
		}

		// If the key does not exist, create a nested map
		if _, exists := current[key]; !exists {
			current[key] = make(DataMap)
		}

		// Move deeper into the nested map
		next, ok := current[key].(map[string]any)
		if !ok {
			next, ok = current[key].(DataMap)
			if !ok {
				return fmt.Errorf("%w: path segment '%s' is not a nested object", ErrInvalidPath, key)
			}
		}
		current = next
	}

	return nil
}

// GetValueByPath retrieves a value from the DataMap based on the hierarchical
// path, returning an error if not found.
func GetValueByPath(m *DataMap, path string) (any, error) {
	if m == nil {
		return nil, ErrEmptyDataMap
	}
	if path == "" {
		return nil, ErrEmptyPath
	}

	path = strings.TrimPrefix(path, PathSep)
	parts := strings.Split(path, PathSep)

	// Handle single-level paths (base case)
	if len(parts) == 1 {
		if v, exists := (*m)[parts[0]]; exists {
			return v, nil // Field successfully removed
		}
		return nil, fmt.Errorf("%w: field '%s' does not exist", ErrInvalidPath, parts[0])
	}

	// Navigate to the nested DataMap at parts[0]
	next, ok := (*m)[parts[0]].(map[string]any)
	if !ok {
		next, ok = (*m)[parts[0]].(DataMap)
		if !ok {
			return nil, fmt.Errorf("%w: path segment '%s' is not a nested object", ErrInvalidPath, parts[0])
		}
	}

	// Recursively get the field from the nested map
	return GetValueByPath((*DataMap)(&next), strings.Join(parts[1:], PathSep))
}

// RemoveFieldByPath removes a field from a DataMap at the specified
// hierarchical path. It traverses nested DataMap structures using PathSep to
// locate and delete the target field. Returns an error if the path is invalid
// or if traversal fails at any intermediate segment.
func RemoveFieldByPath(m *DataMap, path string) error {
	if m == nil {
		return ErrEmptyDataMap
	}
	if path == "" {
		return ErrEmptyPath
	}

	// Trim the leading separator and split the path into components
	path = strings.TrimPrefix(path, PathSep)
	parts := strings.Split(path, PathSep)

	// Handle single-level paths (base case)
	if len(parts) == 1 {
		if _, exists := (*m)[parts[0]]; exists {
			delete(*m, parts[0]) // Delete the key
			return nil           // Field successfully removed
		}
		return fmt.Errorf("%w: field '%s' does not exist", ErrInvalidPath, parts[0])
	}

	// Navigate to the nested DataMap at parts[0]
	next, ok := (*m)[parts[0]].(map[string]any)
	if !ok {
		next, ok = (*m)[parts[0]].(DataMap)
		if !ok {
			return fmt.Errorf("%w: path segment '%s' is not a nested object", ErrInvalidPath, parts[0])
		}
	}

	// Recursively remove the field from the nested map
	err := RemoveFieldByPath((*DataMap)(&next), strings.Join(parts[1:], PathSep))
	if err != nil {
		return err
	}

	// Remove the empty map if it exists (clean up)
	if len(next) == 0 {
		delete(*m, parts[0])
	}

	return nil
}

// ReduceFieldByKeys filters the nested DataMap at the specified path by
// retaining only the provided keys in the output map.
// It updates the original map with the filtered keys or returns an error
// if the path or keys are invalid.
func ReduceFieldByKeys(m *DataMap, path string, keys []string) error {
	obj, err := GetValueByPath(m, path)
	if err != nil {
		return err
	}

	value, err := NewDataMap(obj)
	if err != nil {
		return err
	}

	newValue := make(map[string]any)
	for _, key := range keys {
		v, ok := value[key]
		if ok {
			newValue[key] = v
		}
	}

	return UpdateMapByPath(m, path, newValue)
}
