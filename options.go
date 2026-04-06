package godiff

// AuditStage defines specific stages within an auditing or comparison process,
// represented as a string type.
type AuditStage string

const (
	PreAuditStage  AuditStage = "pre_audit"
	PostAuditStage AuditStage = "post_audit"
)

// EntityOption represents a functional option for modifying a DataMap or DataTypes,
// returning an error if it fails.
type EntityOption func(m *DataMap, t *DataTypes) error

// CompareOption applies optional behaviour during audit compare processes.
// They are executed at each stage of the audit
type CompareOption func(c CompareContext) error

// CompareResult defines an interface for obtaining and analyzing differences
// between two versions of data after func Compare call.
type CompareResult interface {
	// Old returns the "old" state of the data as a DataMap, representing the
	// state prior to the recent update event.
	Old() *DataMap

	// New represents the recent update event to the data performed by the user.
	New() *DataMap

	// Types return a pointer to the DataTypes map, which associates field paths
	// with their corresponding data types.
	Types() *DataTypes

	// GetDiff returns a slice of Diff objects representing changes between two
	// data states stored in the CompareContext.
	GetDiff() []*Diff

	// Warnings returns a slice of errors collected during the auditing process.
	Warnings() []error
}

type CompareContext interface {
	// CompareResult defines an interface for comparing two data versions and
	// providing differences, types, and warnings.
	CompareResult

	// GetStage returns the current AuditStage of the audit compare process.
	GetStage() AuditStage

	// SetStage sets the current stage of the audit process using the provided
	// AuditStage.
	SetStage(AuditStage)

	// AddDiff appends one or more Diff objects to the current comparison context,
	// representing changes between data states.
	AddDiff(...*Diff)

	// SetDiff sets the list of differences (Diff) within the context, replacing
	// any previously stored data.
	SetDiff([]*Diff)

	// AddWarnings adds an error to the list of warnings within the context for
	// tracking audit-related issues.
	// NOTE: Critical error should be returned directly back to the main process
	// to handle. It is assumed that the audit log is placed in the transactional
	// process of the data update, returning critical errors will presumably
	// rollback the transaction.
	AddWarnings(...error)
}

// ContextAuditCompare encapsulates the context for comparing two versions of
// data during an auditing process.
type ContextAuditCompare struct {
	old, new *DataMap
	types    *DataTypes
	diff     []*Diff
	stage    AuditStage
	reports  []error
}

// NewContextAuditCompare initializes a CompareContext for auditing differences
// between two DataMap instances.
func NewContextAuditCompare(old, new *DataMap, types *DataTypes) CompareContext {
	return &ContextAuditCompare{
		old:   old,
		new:   new,
		types: types,
		stage: PreAuditStage,
	}
}

// Old retrieves the "old" DataMap representing the prior state in the context
// comparison.
func (c *ContextAuditCompare) Old() *DataMap {
	return c.old
}

// New represents the recent update event to the data performed by the user.
func (c *ContextAuditCompare) New() *DataMap {
	return c.new
}

// Types returns the DataTypes instance associated with the ContextAuditCompare.
func (c *ContextAuditCompare) Types() *DataTypes {
	return c.types
}

// GetStage retrieves the current stage of the auditing or comparison process
// stored in the ContextAuditCompare.
func (c *ContextAuditCompare) GetStage() AuditStage {
	return c.stage
}

// GetDiff retrieves the list of Diff objects representing differences in the
// context audit comparison.
func (c *ContextAuditCompare) GetDiff() []*Diff {
	return c.diff
}

// SetStage updates the current stage of the auditing process in the
// ContextAuditCompare instance.
func (c *ContextAuditCompare) SetStage(stage AuditStage) {
	c.stage = stage
}

// AddDiff appends one or more Diff objects to the diff slice in the
// ContextAuditCompare instance.
func (c *ContextAuditCompare) AddDiff(diff ...*Diff) {
	c.diff = append(c.diff, diff...)
}

// SetDiff replaces the existing list of differences with the provided slice in
// the comparison context.
func (c *ContextAuditCompare) SetDiff(diff []*Diff) {
	c.diff = diff
}

// AddWarnings appends the provided error to the list of warnings in the context.
func (c *ContextAuditCompare) AddWarnings(err ...error) {
	c.reports = append(c.reports, err...)
}

// Warnings return the list of collected warning errors associated with the
// audit comparison process.
func (c *ContextAuditCompare) Warnings() []error {
	return c.reports
}
