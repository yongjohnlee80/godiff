package godiff

import "slices"

// DataArray represents a generic container type that wraps a slice of elements
// of any given type R.
type DataArray[R any] struct {
	Data []R
}

// Len returns the number of elements in the Data slice of the DataArray.
func (a *DataArray[R]) Len() int {
	return len(a.Data)
}

// ConvertArray transforms a slice of type S into a slice of type T using generic
// type casting and mapping operations.
func ConvertArray[S any, T any](source []S) ([]T, error) {
	initial := DataArray[S]{
		Data: source,
	}
	dataMap, err := NewDataMap(initial)
	if err != nil {
		return nil, err
	}
	res := MustCast[DataArray[T]](dataMap)
	return res.Data, nil
}

func IsValueIn[R comparable](val R, arr ...R) bool {
	return slices.Contains(arr, val)
}
