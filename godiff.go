// Package godiff provides struct-level diffing, comparison, and audit trail
// generation for Go data structures.
//
// Core workflow:
//  1. Convert structs to DataMap via NewDataMap or StructToMap
//  2. Compare old and new states via Compare or CompareMaps
//  3. Inspect Diff objects for changes (path, old value, new value, type)
//  4. Apply Undo/Redo to revert or replay changes
//
// Struct tags control audit behavior using the tag key defined by [StructTag]
// (default: "audit_log"):
//   - `audit_log:"ignore"` — exclude field from comparison
//   - `audit_log:"required"` — fail if field is nil/empty in new data
//   - `audit_log:"immutable"` — fail if field is changed after initial set
//   - `audit_log:"redact"` — mask values in diff output
//   - `audit_log:"reduce_Key1_Key2"` — only keep specified keys in map/struct fields
//
// Nullable types: any value implementing the [Nullable] interface (IsNil() bool,
// Data() any) receives special treatment during comparison — two nil Nullable
// values are considered equal regardless of their underlying zero values.
package godiff
