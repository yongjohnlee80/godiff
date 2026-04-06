package godiff

import "slices"

// OmitEmpty removes empty and zero fields from the DataMap unless they are
// specified in the exceptions list. It is important to be mindful that false
// value for bool or zero for int types are also treated as empty, so you must
// declare them explicitly in the exceptions list if you want to preserve them.
// It is an EntityOption for DataMap type.
func OmitEmpty(exceptions ...string) EntityOption {
	return func(m *DataMap, t *DataTypes) error {
		if m == nil {
			return ErrEmptyDataMap
		}
		for n, v := range *m {
			if !slices.Contains(exceptions, n) && IsEmpty(v) {
				delete(*m, n)
			}
		}
		return nil
	}
}
