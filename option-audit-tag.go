package godiff

import (
	"errors"
	"fmt"
	"slices"
	"strings"
)

var (
	ErrImmutableFieldUpdated = errors.New("immutable field updated")
	ErrRequiredFieldMissing  = errors.New("missing required field")
)

const (
	RedactedValue = "=== REDACTED ==="
)

// WithAuditTags processes audit-related tags during comparison in defined
// stages PreAuditStage and PostAuditStage.
// In PreAuditStage:
//   - It removes fields tagged with "ignore" from the old and new data, adding
//     warnings if removal fails.
//   - It checks the required field if a value is provided, if not, returns an error
//   - It reduces the size of the field by the specified keys if the field is a
//     map or slice, adding warnings if reduction fails.
//
// In PostAuditStage:
//   - It checks for updates to fields tagged as "immutable" and returns an error
//     if any are modified.
//   - It replaces the values of fields tagged as "redact" with "=== REDACTED ===".
func WithAuditTags(c CompareContext) error {
	isNotNil := func(v any) bool {
		return !(v == nil || IsEmpty(v))
	}

	switch c.GetStage() {
	case PreAuditStage:
		var requiredFields []string
		for key, typ := range *(c.Types()) {
			if slices.Contains(typ.Tags, "ignore") {
				if err := RemoveFieldByPath(c.Old(), key); err != nil {
					c.AddWarnings(err)
				}
				if err := RemoveFieldByPath(c.New(), key); err != nil {
					c.AddWarnings(err)
				}
			}
			if slices.Contains(typ.Tags, "required") {
				if v, err := GetValueByPath(c.New(), key); err == nil {
					if !isNotNil(v) {
						requiredFields = append(requiredFields, key)
					}
				}
			}
			for _, tag := range typ.Tags {
				if strings.HasPrefix(tag, "reduce_") {
					reduceBy := strings.Split(tag, "_")
					if err := ReduceFieldByKeys(c.Old(), key, reduceBy[1:]); err != nil {
						c.AddWarnings(err)
					}
					if err := ReduceFieldByKeys(c.New(), key, reduceBy[1:]); err != nil {
						c.AddWarnings(err)
					}
				}
			}
		}
		if len(requiredFields) > 0 {
			return fmt.Errorf("%w: %s", ErrRequiredFieldMissing, strings.Join(requiredFields, ";;"))
		}
	case PostAuditStage:
		immutables := make(map[string]bool)
		vetted := make(map[string]bool)
		for path, typ := range *(c.Types()) {
			if slices.Contains(typ.Tags, "immutable") {
				immutables[path] = true
			}
			if slices.Contains(typ.Tags, "redact") {
				vetted[path] = true
			}
		}
		for _, updated := range c.GetDiff() {
			if immutables[updated.Path] && isNotNil(updated.Old) {
				// This is a security concern should return this error directly
				// back to the main process.
				return fmt.Errorf("%w: %s", ErrImmutableFieldUpdated, updated.Path)
			}
			if vetted[updated.Path] {
				updated.Old = RedactedValue
				updated.New = RedactedValue
			}
		}
	}
	return nil
}
