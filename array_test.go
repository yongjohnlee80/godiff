
package godiff

import (
	"reflect"
	"testing"
)

type ConvertedPerson struct {
	FirstName string  `json:"Name"`
	HowOld    int     `json:"Age"`
	Job       string  `json:"Kind"`
	Address   Address `json:"Address"`
	Tags      []string
}

func TestConvertArray(t *testing.T) {
	address := &Address{
		Street: "ABC St.",
		City:   "Vancouver",
	}

	tests := []struct {
		name        string
		source      interface{}
		expected    interface{}
		expectError bool
	}{
		{
			name:        "convert int to float",
			source:      []int{1, 2, 3},
			expected:    []float64{1.0, 2.0, 3.0},
			expectError: false,
		},
		{
			name:        "empty source array",
			source:      []int{},
			expected:    []float64{},
			expectError: false,
		},
		{
			name:        "convert string to string",
			source:      []string{"a", "b", "c"},
			expected:    []string{"a", "b", "c"},
			expectError: false,
		},
		{
			name: "convert Person to Person2",
			source: []Person{
				{
					Name:    "John",
					Age:     20,
					Kind:    "Employee",
					Tags:    []string{"a"},
					Address: address,
				},
				{
					Name: "Jane",
					Age:  21,
					Kind: "Employee",
					Tags: []string{"b"},
				},
			},
			expected: []ConvertedPerson{
				{
					FirstName: "John",
					HowOld:    20,
					Job:       "Employee",
					Tags:      []string{"a"},
					Address:   *address,
				},
				{
					FirstName: "Jane",
					HowOld:    21,
					Job:       "Employee",
					Tags:      []string{"b"},
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result interface{}
			var err error

			switch src := tt.source.(type) {
			case []int:
				result, err = ConvertArray[int, float64](src)
			case []string:
				result, err = ConvertArray[string, string](src)
			case []Person:
				result, err = ConvertArray[Person, ConvertedPerson](src)
			default:
				t.Errorf("unexpected type: %T", src)
				return
			}

			if (err != nil) != tt.expectError {
				t.Errorf("expected error: %v, got error: %v", tt.expectError, err)
			}
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("expected: %+v, got: %+v", tt.expected, result)
			}
		})
	}
}
