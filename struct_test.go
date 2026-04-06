package godiff

import (
	"fmt"
	"testing"
)

func TestStructToMap(t *testing.T) {
	type testStruct struct {
		Name  string
		Age   int
		Notes string `audit_log:"ignore"`
	}

	s := &testStruct{
		Name:  "Bob",
		Age:   25,
		Notes: "Some notes",
	}

	m, _, err := StructToMap(s)
	if err != nil {
		t.Fatal(err)
	}
	if m["Name"] != "Bob" {
		t.Errorf("expected Name=Bob, got %v", m["Name"])
	}
	if m["Age"] != 25 {
		t.Errorf("expected Age=25, got %v", m["Age"])
	}
}

func TestStructToMap2(t *testing.T) {
	type testAddress struct {
		Street string
		City   string
		State  string `audit_log:"ignore"`
	}
	type testStruct struct {
		Name    string
		Age     int `audit_log:"ignore"`
		Address testAddress
	}

	s := &testStruct{
		Name: "Bob",
		Age:  25,
		Address: testAddress{
			Street: "123 Main St",
			City:   "Vancouver",
			State:  "BC",
		},
	}
	m, types, err := StructToMap(s)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(m)
	fmt.Println(types)
}
