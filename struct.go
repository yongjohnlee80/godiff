package godiff

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

var (
	ErrUnexportedType = errors.New("unexported type")
)

// StructToMap converts a struct into a map and captures the type information
// of each field. It returns a DataMap, a DataTypes map containing the field
// types, and an error if the input is not a struct.
func StructToMap(s any, parentPaths ...string) (DataMap, DataTypes, error) {
	path := ""
	if len(parentPaths) > 0 {
		path = strings.Join(parentPaths, PathSep)
	}

	v := reflect.ValueOf(s)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil, nil, ErrNotStruct
	}

	m := make(DataMap)
	t := make(DataTypes)

	// if currently on the root level, populate the root type.
	if path == "" {
		t.Add("/", reflect.TypeOf(s).String())
	}

	for i := 0; i < v.NumField(); i++ {
		field := v.Type().Field(i)
		fieldValue := v.Field(i)
		fieldName := field.Name
		fieldType := field.Type

		// handle custom auditTags
		auditTag := field.Tag.Get(getStructTag())
		if auditTag == AuditLogIgnore {
			continue
		}

		if fieldType.Kind() == reflect.Struct {
			// handle nested struct recursively
			sub := fieldValue.Interface()
			subMap, subTypes, err := StructToMap(sub, createPath(path, fieldName))
			if err == nil {
				m[fieldName] = subMap
				t.Add(createPath(path, fieldName), fieldType.String())
				t.Merge(&subTypes)
				continue
			}
			if !errors.Is(err, ErrUnexportedType) {
				return nil, nil, err
			}
		}

		// handle default
		if !fieldValue.IsValid() || !fieldValue.CanInterface() {
			// If the field isn't valid or accessible, log an error and continue
			return nil, nil, fmt.Errorf("%w: field at %s", ErrUnexportedType, createPath(path, fieldName))
		}

		m[fieldName] = fieldValue.Interface()
		t.Add(createPath(path, fieldName), fieldType.String())
	}
	return m, t, nil
}
