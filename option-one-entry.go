package godiff

import (
	"slices"
	"strings"
)

// OneEntryIfNew creates a CompareOption that adds a single Diff entry if the
// old data is nil during the PostAuditStage.
func OneEntryIfNew(key string) CompareOption {
	return func(c CompareContext) error {
		switch c.GetStage() {
		case PostAuditStage:
			if *(c.Old()) == nil {
				diff := NewDiff(
					key,
					c.Types().Get("."),
					nil,
					VettedDataMap(c.New(), c.Types()),
					"/",
				)
				c.SetDiff([]*Diff{diff})
			}
		}
		return nil
	}
}

func OneEntryIfNewWithPath(key, path string) CompareOption {
	return func(c CompareContext) error {
		switch c.GetStage() {
		case PostAuditStage:
			if *(c.Old()) == nil {
				diff := NewDiff(
					key,
					c.Types().Get("."),
					nil,
					VettedDataMap(c.New(), c.Types()),
					path,
				)
				c.SetDiff([]*Diff{diff})
			}
		}
		return nil
	}
}

func VettedDataMap(data *DataMap, types *DataTypes) *DataMap {
	vetted := make(map[string]bool)
	for path, typ := range *types {
		if slices.Contains(typ.Tags, "redact") {
			p := strings.TrimPrefix(path, "/")
			vetted[p] = true
		}
	}
	temp := make(DataMap)
	for k, v := range *data {
		if vetted[k] {
			temp[k] = RedactedValue
		} else {
			temp[k] = v
		}
	}
	return &temp
}
