package godiff

import (
	"reflect"
	"testing"
)

func TestUndo(t *testing.T) {
	_, resp, err := Compare(nil, EmployeeJohn)
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.GetDiff()) == 0 {
		t.Fatal("expected diffs")
	}

	_, resp2, err := Compare(EmployeeJohn, EmployeeJohn2)
	if err != nil {
		t.Fatal(err)
	}

	employee, err := NewDataMap(EmployeeJohn2)
	if err != nil {
		t.Fatal(err)
	}

	err = employee.Undo(resp2.GetDiff()...)
	if err != nil {
		t.Fatal(err)
	}
	_, resp3, err := Compare(employee, EmployeeJohn)
	if err != nil {
		t.Fatal(err)
	}
	if len(resp3.GetDiff()) != 0 {
		t.Fatalf("expected 0 diffs after undo, got %d", len(resp3.GetDiff()))
	}

	err = employee.Undo(resp.GetDiff()...)
	if err != nil {
		t.Fatal(err)
	}
	if employee == nil {
		t.Fatal("expected non-nil map")
	}
	if employee["Age"] != nil {
		t.Error("expected Age to be nil")
	}
	if employee["Kind"] != nil {
		t.Error("expected Kind to be nil")
	}
	if employee["Name"] != nil {
		t.Error("expected Name to be nil")
	}
	if employee["Roles"] != nil {
		t.Error("expected Roles to be nil")
	}
	if employee["Tags"] != nil {
		t.Error("expected Tags to be nil")
	}

	address, ok := employee["Address"].(map[string]interface{})
	if !ok {
		t.Fatal("expected Address to be a map")
	}
	if address["City"] != nil {
		t.Error("expected City to be nil")
	}
	if address["Street"] != nil {
		t.Error("expected Street to be nil")
	}
}

func TestRemoveFieldByPath(t *testing.T) {
	type testStruct1 struct {
		Name  string
		Age   int
		Notes string `audit_log:"ignore"`
	}

	type testStruct2 struct {
		Info     testStruct1
		JobTitle string
		Manager  testStruct1
	}

	tests := &testStruct2{
		Info: testStruct1{
			Name:  "Bob",
			Age:   25,
			Notes: "Some notes",
		},
		JobTitle: "Software Engineer",
		Manager: testStruct1{
			Name: "John",
			Age:  30,
		},
	}

	testMap, err := NewDataMap(tests)
	if err != nil {
		t.Fatal(err)
	}

	err = RemoveFieldByPath(&testMap, "/Manager/Name")
	if err != nil {
		t.Fatal(err)
	}

	err = UpdateMapByPath(&testMap, "/Manager/Age", 35)
	if err != nil {
		t.Fatal(err)
	}

	result := MustCast[testStruct2](testMap)
	if result.Manager.Age != 35 {
		t.Errorf("expected Manager.Age=35, got %d", result.Manager.Age)
	}
	if result.Manager.Name == "John" {
		t.Error("expected Manager.Name to be empty after removal")
	}
}

func TestUpdateMapByPath(t *testing.T) {
	tests := []struct {
		name      string
		inputMap  DataMap
		path      string
		value     interface{}
		expected  DataMap
		expectErr bool
	}{
		{
			name:      "simple key update",
			inputMap:  DataMap{"key1": "value1"},
			path:      "/key2",
			value:     "value2",
			expected:  DataMap{"key1": "value1", "key2": "value2"},
			expectErr: false,
		},
		{
			name:      "nested key update",
			inputMap:  DataMap{"key1": DataMap{"key2": "value2"}},
			path:      "/key1/key3",
			value:     "value3",
			expected:  DataMap{"key1": DataMap{"key2": "value2", "key3": "value3"}},
			expectErr: false,
		},
		{
			name:      "create nested structure",
			inputMap:  DataMap{},
			path:      "/key1/key2/key3",
			value:     "value3",
			expected:  DataMap{"key1": DataMap{"key2": DataMap{"key3": "value3"}}},
			expectErr: false,
		},
		{
			name:      "overwrite existing key",
			inputMap:  DataMap{"key1": "value1"},
			path:      "/key1",
			value:     "newValue",
			expected:  DataMap{"key1": "newValue"},
			expectErr: false,
		},
		{
			name:      "path conflict with non-map value",
			inputMap:  DataMap{"key1": "value1"},
			path:      "/key1/key2",
			value:     "value2",
			expected:  DataMap{"key1": "value1"},
			expectErr: true,
		},
		{
			name:      "deeply nested merge",
			inputMap:  DataMap{"key1": DataMap{"key2": DataMap{"key3": "value3"}}},
			path:      "/key1/key2/key4",
			value:     "value4",
			expected:  DataMap{"key1": DataMap{"key2": DataMap{"key3": "value3", "key4": "value4"}}},
			expectErr: false,
		},
		{
			name:      "empty path",
			inputMap:  DataMap{"key1": "value1"},
			path:      "/",
			value:     DataMap{"key1": "value1"},
			expected:  DataMap{"key1": "value1"},
			expectErr: false,
		},
		{
			name:      "nil input map",
			inputMap:  nil,
			path:      "/key1",
			value:     "value1",
			expected:  DataMap{"key1": "value1"},
			expectErr: false,
		},
		{
			name:      "complex object insertion",
			inputMap:  DataMap{"key1": "value1"},
			path:      "/key2",
			value:     DataMap{"nestedKey": "nestedValue"},
			expected:  DataMap{"key1": "value1", "key2": DataMap{"nestedKey": "nestedValue"}},
			expectErr: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := UpdateMapByPath(&test.inputMap, test.path, test.value)
			if (err != nil) != test.expectErr {
				t.Fatalf("expected error: %v, got: %v", test.expectErr, err)
			}
			if !test.expectErr && !reflect.DeepEqual(test.inputMap, test.expected) {
				t.Errorf("expected map: %v, got: %v", test.expected, test.inputMap)
			}
		})
	}
}

func TestReduceFieldByKeys(t *testing.T) {
	type testStruct struct {
		Id   string
		Name string
		Age  int
	}

	type testStruct2 struct {
		Info     testStruct `audit_log:"reduce_Id_Name"`
		JobTitle string
		Manager  string `audit_log:"ignore"`
	}

	testInput := testStruct2{
		Info: testStruct{
			Id:   "1",
			Name: "Bob",
			Age:  25,
		},
		JobTitle: "Software Engineer",
		Manager:  "John",
	}

	tests := []struct {
		name      string
		inputMap  testStruct2
		path      string
		keys      []string
		expected  testStruct2
		expectErr bool
	}{
		{
			name:     "Reduce By Id and Name",
			inputMap: testInput,
			path:     "/Info",
			keys:     []string{"Id", "Name"},
			expected: testStruct2{
				Info: testStruct{
					Id:   "1",
					Name: "Bob",
				},
				JobTitle: "Software Engineer",
				Manager:  "John",
			},
			expectErr: false,
		},
		{
			name:     "Invalid Field, marshalling error",
			inputMap: testInput,
			path:     "/JobTitle",
			keys:     []string{"JobTitle"},
			expected: testStruct2{
				Info: testStruct{
					Id:   "1",
					Name: "Bob",
					Age:  25,
				},
				JobTitle: "Software Engineer",
				Manager:  "John",
			},
			expectErr: true,
		},
		{
			name:     "Invalid Keys, should return empty field and no error",
			inputMap: testInput,
			path:     "/Info",
			keys:     []string{"Salary"},
			expected: testStruct2{
				Info:     testStruct{},
				JobTitle: "Software Engineer",
				Manager:  "John",
			},
			expectErr: false,
		},
		{
			name:      "Invalid Path",
			inputMap:  testInput,
			path:      "/InvalidPath",
			expected:  testInput,
			expectErr: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			input, err := NewDataMap(test.inputMap)
			if err != nil {
				t.Fatal(err)
			}

			err = ReduceFieldByKeys(&input, test.path, test.keys)
			if (err != nil) != test.expectErr {
				t.Fatalf("expected error: %v, got: %v", test.expectErr, err)
			}

			actual := MustCast[testStruct2](input)
			if !test.expectErr && !reflect.DeepEqual(test.expected, actual) {
				t.Errorf("expected map: %v, got: %v", test.expected, actual)
			}
		})
	}
}
