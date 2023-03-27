package tasks

type Scripts []Script

type Script struct {
	ID    string
	Tasks []CoreTask
}

type FieldNameMapper interface {
	BuildFieldMap(t CoreTask)
	GetFieldName(fk string) (fieldName string)
	SetFieldName(fk string, fieldName string)
}

type FieldStatus struct {
	HasNewValue   bool
	ChangeApplied bool
	Clear         bool
}

type FieldStatusTracker interface {
	GetFieldStatus(fieldName string) (status FieldStatus, found bool)
	SetFieldStatus(fieldName string, status FieldStatus)
	SetHasNewValue(fieldName string) (err error)
	HasNewValue(fieldName string) (hasNew bool)
	SetClear(fieldName string) (err error)
	ShouldClear(fieldName string) (should bool)
	SetChangeApplied(fieldName string) (err error)
	WithNewValues(applyFn func(fieldName string, fs FieldStatus) (err error)) (err error)
}

type CoreTask interface {
	GetTypeName() string
	Validate(goos string) error
	GetPath() string
	GetRequirements() []string
	GetCreatesFilesList() []string
	GetOnlyIfCmds() []string
	GetUnlessCmds() []string
}

// TaskWithFieldTracker allows the task access to both field mapper and tracker info.
// New interfaces will be required if there's a requirement for allowing access to only one or
// the other.
type TaskWithFieldTracker interface {
	SetMapper(mapper FieldNameMapper)
	SetTracker(tracker FieldStatusTracker)
	IsChangeField(fieldName string) (excluded bool)
}
