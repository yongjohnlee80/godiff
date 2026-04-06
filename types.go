package godiff

import (
	"reflect"
	"strings"
)

type DataType struct {
	Kind string
	Tags []string
}

// DataTypes is a type alias for a map used to a path of associate field with
// their data types as string representations.
type DataTypes map[string]DataType

// NewDataTypes converts a struct to a DataTypes map and applies EntityOptions,
// returning the DataTypes and any errors.
func NewDataTypes(data any, opts ...EntityOption) (DataTypes, error) {
	t, err := retrieveTypesFromStruct(data, "")
	if err != nil {
		return nil, err
	}
	err = t.Options(opts...)
	return t, err
}

// Add associates the given field name with the specified data type in the
// DataTypes map.
func (d *DataTypes) Add(field, typ string, tags ...string) {
	(*d)[field] = DataType{
		Kind: typ,
		Tags: tags,
	}
}

// Get retrieves the data type associated with the specified field name from
// the DataTypes map.
func (d *DataTypes) Get(field string) string {
	return (*d)[field].Kind
}

func (d *DataTypes) GetTags(field string) []string {
	return (*d)[field].Tags
}

// Options apply a series of EntityOption operations to the DataTypes instance
// and return an error if any operation fails.
func (d *DataTypes) Options(ops ...EntityOption) error {
	for _, op := range ops {
		err := op(nil, d)
		if err != nil {
			return err
		}
	}
	return nil
}

// Merge copies all key-value pairs from another DataTypes map into the current
// one, overwriting any existing keys.
func (d *DataTypes) Merge(other *DataTypes) {
	for k, v := range *other {
		(*d)[k] = v
	}
}

// retrieveTypesFromStruct recursively retrieves the types of fields from
// a struct or nested struct, returning them as DataTypes.
// Returns an error if the provided input is not a struct.
func retrieveTypesFromStruct(s any, path string) (DataTypes, error) {
	t := make(DataTypes)
	// if currently on the root level, populate the root type.
	if path == "" {
		t.Add("/", reflect.TypeOf(s).String())
	}

	v := reflect.ValueOf(s)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil, ErrNotStruct
	}

	for i := 0; i < v.NumField(); i++ {
		field := v.Type().Field(i)
		fieldValue := v.Field(i)
		fieldName := field.Name
		fieldType := field.Type

		if fieldType.Kind() == reflect.Struct && fieldValue.CanInterface() && fieldValue.IsValid() {
			// handle nested struct recursively
			sub := fieldValue.Interface()
			subTypes, err := retrieveTypesFromStruct(sub, createPath(path, fieldName))
			if err == nil {
				tags := strings.Split(field.Tag.Get(structTag), ",")
				t.Add(createPath(path, fieldName), fieldType.String(), tags...)
				t.Merge(&subTypes)
				continue
			}
		}
		tags := strings.Split(field.Tag.Get(structTag), ",")
		t.Add(createPath(path, fieldName), fieldType.String(), tags...)
	}
	return t, nil
}
