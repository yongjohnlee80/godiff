package godiff

// WithDiffPathPrefix appends a given prefix to the key and path of each Diff
// in the CompareContext during PostAuditStage.
func WithDiffPathPrefix(prefix string) CompareOption {
	return func(c CompareContext) error {
		switch c.GetStage() {
		case PostAuditStage:
			diffs := c.GetDiff()
			for _, diff := range diffs {
				diff.Key = createPath(prefix, diff.Key)
				diff.Path = createPath(prefix, diff.Path)
			}
			c.SetDiff(diffs)
		}
		return nil
	}
}

func WithDiffKeyAndPath(key, path string) CompareOption {
	return func(c CompareContext) error {
		switch c.GetStage() {
		case PostAuditStage:
			diffs := c.GetDiff()
			for _, diff := range diffs {
				diff.Key = createPath(key, diff.Key)
				diff.Path = createPath(path, diff.Path)
			}
			c.SetDiff(diffs)
		}
		return nil
	}
}
