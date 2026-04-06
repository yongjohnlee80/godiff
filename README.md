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

The package provides two interfaces for nullable/optional value types:

```go
// Nullable — minimal interface for nil-safe comparison.
// Two values where both IsNil() return true are considered equal.
type Nullable interface {
    IsNil() bool
}

// NullableWithData[T] — extends Nullable with typed data access.
// Used when the package needs to unwrap the underlying value (e.g. time comparison).
type NullableWithData[T any] interface {
    Nullable
    Data() T
}
```

The split design avoids interface conflicts with generic types. A type like `nilable.Value[T]` only needs `IsNil() bool` to satisfy `Nullable`, and can optionally add `Data() T` for `NullableWithData[T]` — no `Data() any` signature clash.

```go
type Value[T any] struct {
    data  T
    isSet bool
}

// Satisfies Nullable — enough for nil-safe comparison
func (v Value[T]) IsNil() bool { return !v.isSet }

// Satisfies NullableWithData[T] — enables value unwrapping
func (v Value[T]) Data() T { return v.data }
```

During comparison, two unset values are equal regardless of their underlying zero values:

```go
a := nilable.New[string]()           // IsNil() == true
b := nilable.New[string]()           // IsNil() == true
// godiff treats a and b as equal
```

The package also retains legacy support for types whose type string starts with `"nilable.Value"` via string-based detection.

## Configurable Struct Tag

The struct tag key defaults to `"audit_log"` and can be changed at runtime:

```go
// Check the current tag key
fmt.Println(godiff.GetStructTag()) // "audit_log"

// Change it before any comparison calls (e.g. in init or main)
godiff.SetStructTag("diff")

// Now the package reads `diff:"ignore"` instead of `audit_log:"ignore"`
type Document struct {
    ID       string
    Internal string `diff:"ignore"`
    Secret   string `diff:"redact"`
}
```

| Function | Description |
|----------|-------------|
| `GetStructTag()` | Returns the currently active tag key |
| `SetStructTag(tag)` | Overrides the tag key (call once at startup) |
| `DefaultStructTag` | Constant `"audit_log"` for reference |

## License

MIT
