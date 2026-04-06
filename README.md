# godiff

A zero-dependency Go library for struct-level diffing, comparison, and audit trail generation.

## Features

- **Struct Comparison** — Compare two Go structs and get a list of field-level differences
- **Diff Generation** — Produces `Diff` objects with path, type, old value, and new value
- **Undo/Redo** — Revert or replay changes using diff objects
- **Audit Tags** — Control comparison behavior via struct tags (`ignore`, `required`, `immutable`, `redact`, `reduce`)
- **Nullable Interface** — Any type implementing `Nullable` (IsNil/Data) gets nil-safe comparison semantics
- **Configurable Struct Tag** — Tag key defined as `StructTag` constant (default `"audit_log"`) for easy customization
- **DataMap** — Flexible `map[string]any` with hierarchical path operations (get, set, remove, update-by-path)
- **Type Casting** — Convert between structs and DataMaps with `Cast`/`MustCast` generics
- **Array Conversion** — Transform typed slices between compatible struct shapes
- **Zero Dependencies** — Uses only the Go standard library

## Install

```bash
go get github.com/yongjohnlee80/godiff
```

## Quick Start

### Compare Two Structs

```go
package main

import (
    "fmt"
    "github.com/yongjohnlee80/godiff"
)

type User struct {
    Name string
    Age  int
    Tags []string
}

func main() {
    old := User{Name: "Alice", Age: 30, Tags: []string{"admin"}}
    new := User{Name: "Alice", Age: 31, Tags: []string{"admin", "editor"}}

    keys, result, err := godiff.Compare(old, new)
    if err != nil {
        panic(err)
    }

    fmt.Println("Changed fields:", keys)
    // Changed fields: [Age Tags]

    for _, d := range result.GetDiff() {
        fmt.Println(d.String())
        // /Age(int), 30 -> 31
        // /Tags([]string), ["admin"] -> ["admin","editor"]
    }
}
```

### Undo Changes

```go
// Create a DataMap from the new state
userMap, _ := godiff.NewDataMap(new)

// Get diffs
_, result, _ := godiff.Compare(old, new)

// Undo — reverts userMap back to old state
userMap.Undo(result.GetDiff()...)

// Cast back to struct
reverted := godiff.MustCast[User](userMap)
fmt.Println(reverted.Age) // 30
```

### Detect First-Time Creation

```go
// When old is nil, all fields appear as diffs (nil → value)
keys, result, _ := godiff.Compare(nil, new)
for _, d := range result.GetDiff() {
    fmt.Printf("%s: %v\n", d.Path, d.New)
}
```

## Audit Tags

Control comparison behavior with the `audit_log` struct tag:

```go
type Document struct {
    ID        string   `audit_log:"required"`   // Must be non-empty in new data
    Title     string                              // Normal comparison
    Secret    string   `audit_log:"redact"`      // Values masked as "=== REDACTED ===" in diffs
    CreatedBy string   `audit_log:"immutable"`   // Cannot change after first set
    Internal  string   `audit_log:"ignore"`      // Excluded from comparison entirely
    Metadata  Metadata `audit_log:"reduce_Id_Name"` // Only compare Id and Name fields
}

// Enable audit tags during comparison
keys, result, err := godiff.Compare(old, new, godiff.WithAuditTags)
```

| Tag | Behavior |
|-----|----------|
| `ignore` | Field is removed from both old and new before comparison |
| `required` | Returns error if field is nil/empty in new data |
| `immutable` | Returns error if field changes after being initially set |
| `redact` | Replaces old and new values with `"=== REDACTED ==="` in diffs |
| `reduce_Key1_Key2` | Only keeps specified keys when comparing map/struct fields |

## DataMap Operations

```go
// Convert struct to flexible map
m, _ := godiff.NewDataMap(myStruct)

// Path-based operations
godiff.UpdateMapByPath(&m, "/nested/field", "new value")
val, _ := godiff.GetValueByPath(&m, "/nested/field")
godiff.RemoveFieldByPath(&m, "/nested/field")

// Cast back to typed struct
result := godiff.MustCast[MyStruct](m)
```

## Compare Options

```go
// Combine multiple options
keys, result, err := godiff.Compare(old, new,
    godiff.WithAuditTags,                        // Process audit tags
    godiff.WithDiffPathPrefix("users"),           // Prefix all diff paths
    godiff.OneEntryIfNew("user_created"),         // Single diff for new entities
)
```

## OmitEmpty

```go
// Remove zero-valued fields from DataMap (useful before storage)
m, _ := godiff.NewDataMap(myStruct, godiff.OmitEmpty("Active")) // keep "Active" even if false
```

## Array Conversion

```go
type PersonV1 struct {
    Name string `json:"Name"`
    Age  int    `json:"Age"`
}

type PersonV2 struct {
    FullName string `json:"Name"`  // Maps via JSON tag
    Years    int    `json:"Age"`
}

people := []PersonV1{{Name: "Alice", Age: 30}}
converted, err := godiff.ConvertArray[PersonV1, PersonV2](people)
// converted[0].FullName == "Alice", converted[0].Years == 30
```

## Nullable Interface

Types implementing the `Nullable` interface get special nil-safe comparison semantics. Two values where both `IsNil()` return true are considered equal, regardless of their underlying zero values.

```go
// Nullable is satisfied by any type with these two methods:
type Nullable interface {
    IsNil() bool
    Data() any
}
```

This is useful for optional/nullable wrapper types (like `nilable.Value[T]`):

```go
type OptionalString struct {
    value string
    isSet bool
}

func (o OptionalString) IsNil() bool { return !o.isSet }
func (o OptionalString) Data() any   { return o.value }

// During comparison, two unset OptionalString values are equal
// even if their underlying string is different ("" vs "something")
```

The package also retains legacy support for types whose type string starts with `"nilable.Value"` via string-based detection.

## Configurable Struct Tag

The struct tag key is defined as the `StructTag` constant (default `"audit_log"`). To use a different tag namespace across the package, update this constant:

```go
// In godiff/utils.go
const StructTag = "audit_log"  // change to "diff", "audit", etc.
```

## License

MIT
