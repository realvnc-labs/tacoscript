package tasks

import (
	"fmt"
	"time"
)

type Scripts []Script

type Script struct {
	ID    string
	Tasks []Task
}

type FieldNameMapper interface {
	BuildFieldMap(t Task)
	GetFieldName(fk string) (fieldName string)
	SetFieldName(fk string, fieldName string)
}

type FieldStatusTracker interface {
	GetFieldStatus(fieldName string) (status FieldStatus, found bool)
	SetFieldStatus(fieldName string, status FieldStatus)
	SetHasNewValue(fieldName string) (err error)
	HasNewValue(fieldName string) (hasNew bool)
	SetClear(fieldName string) (err error)
	ShouldClear(fieldKey string) (should bool)
	SetChangeApplied(fieldName string) (err error)
	WithNewValues(applyFn func(fieldName string, fs FieldStatus) (err error)) (err error)
}

type Task interface {
	GetTypeName() string
	Validate(goos string) error
	GetPath() string
	GetRequirements() []string
	GetCreatesFilesList() []string
	GetOnlyIfCmds() []string
	GetUnlessCmds() []string

	GetMapper() (mapper FieldNameMapper)
}

type TaskWithTracker interface {
	GetTracker() (tracker FieldStatusTracker)
	IsChangeField(fieldName string) (excluded bool)
}

type ExecutionResult struct {
	Err        error
	Duration   time.Duration
	StdErr     string
	StdOut     string
	IsSkipped  bool
	SkipReason string
	Pid        int
	Name       string
	Comment    string
	Changes    map[string]string
}

func (tr *ExecutionResult) String() string {
	if tr.Err != nil {
		return fmt.Sprintf(`Execution failed: %v, StdErr: %s, Took: %v, StdOut: %s`, tr.Err, tr.StdErr, tr.Duration, tr.StdOut)
	}

	if tr.IsSkipped {
		return fmt.Sprintf(`Execution is Skipped: %s, StdOut: %s, StdErr: %s, Took: %v`, tr.SkipReason, tr.StdOut, tr.StdErr, tr.Duration)
	}

	return fmt.Sprintf(`Execution success, StdOut: %s, StdErr: %s, Took: %s`, tr.StdOut, tr.StdErr, tr.Duration)
}

// Succeeded returns true if task succeeded or was skipped
func (tr *ExecutionResult) Succeeded() bool {
	return tr.Err == nil
}
