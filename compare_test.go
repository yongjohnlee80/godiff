package godiff

import (
	"fmt"
	"testing"
	"time"
)

type Role struct {
	Title   string
	Manager Person
}

type Address struct {
	Street string
	City   string
}
type Person struct {
	Name    string
	Age     int
	Address *Address `audit_log:"ignore"`
	Tags    []string
	Kind    string `audit_log:"immutable"`
}

type Employee struct {
	Person
	Roles []*Role
}

var (
	John = Person{
		Name: "John",
		Age:  30,
		Address: &Address{
			Street: "123 Hangul Street",
			City:   "Seoul",
		},
		Tags: []string{"engineer", "male"},
	}

	Bryan = Person{
		Name: "Bryan",
		Age:  35,
		Address: &Address{
			Street: "123 Main St",
			City:   "Vancouver",
		},
		Tags: []string{"senior engineer", "male"},
	}

	Sarah = Person{
		Name: "Sarah",
		Age:  33,
		Address: &Address{
			Street: "123 Main St",
			City:   "Vancouver",
		},
		Tags: []string{"director", "female"},
	}

	EmployeeJohn = Employee{
		Person: John,
		Roles: []*Role{
			{
				Title:   "Full Stack Developer",
				Manager: Sarah,
			},
		},
	}

	EmployeeJohn2 = Employee{
		Person: John,
		Roles: []*Role{
			{
				Title:   "Software Engineer",
				Manager: Bryan,
			},
		},
	}
)

func TestCompare(t *testing.T) {
	t.Parallel()

	t.Run("TestComparePersons", func(t *testing.T) {
		keys, resp, err := Compare(John, Bryan)
		if err != nil {
			t.Fatal(err)
		}
		assertContains(t, keys, "Tags")
		assertContains(t, keys, "Name")
		assertContains(t, keys, "Age")
		assertContains(t, keys, "City")
		assertContains(t, keys, "Street")
		for _, d := range resp.GetDiff() {
			fmt.Println(d.String())
		}
	})

	t.Run("TestCompareEmployees", func(t *testing.T) {
		keys, resp, err := Compare(EmployeeJohn, EmployeeJohn2)
		if err != nil {
			t.Fatal(err)
		}
		assertContains(t, keys, "Roles")
		for _, d := range resp.GetDiff() {
			fmt.Println(d.String())
		}
	})

	t.Run("TestCompareNilInitially", func(t *testing.T) {
		_, resp, err := Compare(nil, EmployeeJohn)
		if err != nil {
			t.Fatal(err)
		}

		for _, d := range resp.GetDiff() {
			fmt.Println(d.String())
		}
	})

	t.Run("TestCompareAuditTags", func(t *testing.T) {
		keys, _, err := Compare(John, Sarah, WithAuditTags)
		if err != nil {
			t.Fatal(err)
		}
		assertContains(t, keys, "Tags")
		assertContains(t, keys, "Name")
		assertContains(t, keys, "Age")
		assertNotContains(t, keys, "City")
		assertNotContains(t, keys, "Street")
	})

	t.Run("TestCompareAuditTags2", func(t *testing.T) {
		john := John
		john.Kind = "human"
		keys, _, err := Compare(John, john, WithAuditTags)
		if err != nil {
			t.Fatal(err)
		}
		assertContains(t, keys, "Kind")

		john2 := John
		john2.Kind = "angel"
		_, _, err = Compare(john, john2, WithAuditTags)
		if err == nil {
			t.Fatal("expected ErrImmutableFieldUpdated")
		}
	})
}

func TestDataMapCast(t *testing.T) {
	m, err := NewDataMap(EmployeeJohn)
	if err != nil {
		t.Fatal(err)
	}

	employee := MustCast[Employee](m)
	if employee.Name != EmployeeJohn.Name || employee.Age != EmployeeJohn.Age {
		t.Fatalf("expected %+v, got %+v", EmployeeJohn, employee)
	}
}

func TestDefaultCompFunc(t *testing.T) {
	t.Parallel()

	t.Run("Compare Time", func(t *testing.T) {
		t1 := time.Now()
		t2 := t1.Add(time.Second * 59)
		if !DefaultCompareFunc("time.Time", t1, t2) {
			t.Error("expected times within 59s to be equal")
		}

		t2 = t1.Add(time.Minute * 1)
		if DefaultCompareFunc("time.Time", t1, t2) {
			t.Error("expected times 1m apart to be not equal")
		}
	})

	t.Run("Compare nilable empty", func(t *testing.T) {
		// Simulate nilable.Value behavior: empty string vs nil
		if !DefaultCompareFunc("nilable.Value[string]", "", nil) {
			t.Error("expected empty nilable values to be equal")
		}
	})

	t.Run("Compare nilable.Value[time.Time]", func(t *testing.T) {
		// Both nil → equal (nilable empty values)
		if !DefaultCompareFunc("nilable.Value[time.Time]", nil, nil) {
			t.Error("expected two nil nilable times to be equal")
		}
	})

	t.Run("Compare nilable number", func(t *testing.T) {
		if !DefaultCompareFunc("nilable.Value[int]", nil, 0) {
			t.Error("expected nil nilable int and 0 to be equal")
		}
	})
}

// Test helpers — stdlib only, no testify

func assertContains(t *testing.T, slice []string, val string) {
	t.Helper()
	for _, s := range slice {
		if s == val {
			return
		}
	}
	t.Errorf("expected slice to contain %q, got %v", val, slice)
}

func assertNotContains(t *testing.T, slice []string, val string) {
	t.Helper()
	for _, s := range slice {
		if s == val {
			t.Errorf("expected slice NOT to contain %q", val)
			return
		}
	}
}
